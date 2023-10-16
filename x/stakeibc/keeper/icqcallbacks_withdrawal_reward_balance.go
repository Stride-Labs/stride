package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"

	icqkeeper "github.com/Stride-Labs/stride/v14/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v14/utils"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// WithdrawalRewardBalanceCallback is a callback handler for WithdrawalRewardBalance queries.
// The query response will return the withdrawal account balance for a specific (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to transfer the discovered balance to the tradeZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func WithdrawalRewardBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_WithdrawalRewardBalance,
		"Starting withdrawal reward balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm host exists
	chainId := query.ChainId
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance
	withdrawalRewardBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	// Confirm the balance is greater than zero
	if withdrawalRewardBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"No balance of reward tokens yet found in address: %s, balance: %v", hostZone.WithdrawalIcaAddress, withdrawalRewardBalanceAmount))
		return nil
	}

	// Unmarshal the callback data containing the hostChain and rewardDenom on that host zone
	var callbackData types.WithdrawalRewardBalanceQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal withdrawal reward balance callback data")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance,
		"Query response - Withdrawal Reward Balance: %v %s", withdrawalRewardBalanceAmount, callbackData.RewardDenomOnHostZone))

	// Transfer all found reward tokens from the hostZone to the tradeZone to be converted
	tradeZoneID := k.GetTradeZoneID(ctx, callbackData.RewardDenomOnHostZone, callbackData.HostZoneId)
	tradeZone, found := k.GetHostZone(ctx, tradeZoneID)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for the needed trade zone ID (%s)", tradeZoneID)
	}
	
	// Using ICA commands on the withdrawal address, transfer the found reward tokens from the host zone to the trade zone
	k.TransferHostRewardTokensToTrade(ctx, withdrawalRewardBalanceAmount, callbackData.RewardDenomOnHostZone, hostZone, tradeZone)
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance, 
		"Sending discovered reward tokens %v %s from hostZone %s to tradeZone %s", 
		withdrawalRewardBalanceAmount, callbackData.RewardDenomOnHostZone, hostZone.ChainId, tradeZone.ChainId))

	return nil
}
