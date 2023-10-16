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

// TradeRewardBalanceCallback is a callback handler for TradeRewardBalance queries.
// The query response will return the trade ICA account balance for a specific (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to initiate a swap on the tradeZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func TradeRewardBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_WithdrawalRewardBalance,
		"Starting withdrawal reward balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// Confirm queried zone exists
	chainId := query.ChainId
	tradeZone, found := k.GetHostZone(ctx, chainId) // this query went to the trade zone
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no registered zone for queried chain ID (%s)", chainId)
	}

	// Unmarshal the query response args to determine the balance
	tradeRewardBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	// Unmarshal the callback data containing the hostChain and rewardDenom on that host zone
	var callbackData types.TradeRewardBalanceQueryCallback
	if err := proto.Unmarshal(query.CallbackData, &callbackData); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	}
	hostZone, found := k.GetHostZone(ctx, callbackData.HostZoneId) // this query went to the trade zone, hostZone comes from callback data
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "no host zone could be loaded for callback chain ID (%s)", callbackData.HostZoneId)
	}	

	// Confirm the balance is greater than zero
	if tradeRewardBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeRewardBalance,
			"No balance of reward tokens yet found in address: %s, balance: %v", hostZone.RewardTradeIcaAddress, tradeRewardBalanceAmount))
		return nil
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeRewardBalance,
		"Query response - Withdrawal Reward Balance: %v %s", tradeRewardBalanceAmount, callbackData.RewardDenomOnTradeZone))


	// Trade all found reward tokens in the trade ICA to the output denom of their trade pool
	k.TradeRewardTokens(ctx, tradeRewardBalanceAmount, callbackData.RewardDenomOnTradeZone, hostZone, tradeZone)
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(callbackData.TradeZoneId, ICQCallbackID_TradeRewardBalance, 
		"Swapping discovered reward tokens %v %s on tradeZone %s", 
		tradeRewardBalanceAmount, callbackData.RewardDenomOnTradeZone, tradeZone.ChainId))

	return nil
}
