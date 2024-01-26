package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"

	icqkeeper "github.com/Stride-Labs/stride/v18/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v18/utils"
	icqtypes "github.com/Stride-Labs/stride/v18/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

// TradeRewardBalanceCallback is a callback handler for TradeRewardBalance queries.
// The query response will return the trade ICA account balance for a specific (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to initiate a swap on the tradeZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func TradeRewardBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_TradeRewardBalance,
		"Starting withdrawal reward balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	chainId := query.ChainId // should be the tradeZoneId, used in logging

	// Unmarshal the query response args to determine the balance
	tradeRewardBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
	}

	// Confirm the balance is greater than zero, or else exit without further action for now
	if tradeRewardBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeRewardBalance,
			"Not enough reward tokens found in trade ICA, balance: %+v", tradeRewardBalanceAmount))
		return nil
	}

	// Unmarshal the callback data containing the tradeRoute we are on
	var tradeRouteCallback types.TradeRouteCallback
	if err := proto.Unmarshal(query.CallbackData, &tradeRouteCallback); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	}

	// Lookup the trade route from the keys in the callback
	tradeRoute, found := k.GetTradeRoute(ctx, tradeRouteCallback.RewardDenom, tradeRouteCallback.HostDenom)
	if !found {
		return types.ErrTradeRouteNotFound.Wrapf("trade route from %s to %s not found",
			tradeRouteCallback.RewardDenom, tradeRouteCallback.HostDenom)
	}
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeRewardBalance,
		"Query response - Withdrawal Reward Balance: %v %s", tradeRewardBalanceAmount, tradeRoute.RewardDenomOnTradeZone))

	// Trade all found reward tokens in the trade ICA to the output denom of their trade pool
	if err := k.SwapRewardTokens(ctx, tradeRewardBalanceAmount, tradeRoute); err != nil {
		return errorsmod.Wrapf(err, "unable to swap reward tokens")
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeRewardBalance,
		"Swapping discovered reward tokens %v %s for %s",
		tradeRewardBalanceAmount, tradeRoute.RewardDenomOnTradeZone, tradeRoute.HostDenomOnTradeZone))

	return nil
}
