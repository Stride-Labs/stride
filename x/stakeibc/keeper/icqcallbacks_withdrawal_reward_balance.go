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

	chainId := query.ChainId

	// Unmarshal the query response args to determine the balance
	withdrawalRewardBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	// Confirm the balance is greater than zero, or else exit early without further action
	if withdrawalRewardBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalBalance,
			"Not enough reward tokens yet found in withdrawalICA, balance: %v", withdrawalRewardBalanceAmount))
		return nil
	}

	// Unmarshal the callback data which is just the trade route
	var tradeRoute types.TradeRoute
	if err := proto.Unmarshal(query.CallbackData, &tradeRoute); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal withdrawal reward balance callback data")
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance,
		"Query response - Withdrawal Reward Balance: %v %s", withdrawalRewardBalanceAmount, tradeRoute.RewardDenomOnHostZone))
	
	// Using ICA commands on the withdrawal address, transfer the found reward tokens from the host zone to the trade zone
	k.TransferRewardTokensHostToTrade(ctx, withdrawalRewardBalanceAmount, tradeRoute)
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_WithdrawalRewardBalance, 
		"Sending discovered reward tokens %v %s from hostZone to tradeZone", 
		withdrawalRewardBalanceAmount, tradeRoute.RewardDenomOnHostZone))

	return nil
}
