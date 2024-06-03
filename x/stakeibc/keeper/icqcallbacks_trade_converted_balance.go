package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/gogoproto/proto"

	icqkeeper "github.com/Stride-Labs/stride/v22/x/interchainquery/keeper"

	"github.com/Stride-Labs/stride/v22/utils"
	icqtypes "github.com/Stride-Labs/stride/v22/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

// TradeConvertedBalanceCallback is a callback handler for TradeConvertedBalance queries.
// The query response will return the trade account balance for a converted (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to transfer the discovered balance back to hostZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func TradeConvertedBalanceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_TradeConvertedBalance,
		"Starting trade converted balance callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	chainId := query.ChainId // should be the tradeZoneId

	// Unmarshal the query response args to determine the balance
	tradeConvertedBalanceAmount, err := icqkeeper.UnmarshalAmountFromBalanceQuery(k.cdc, args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to determine balance from query response")
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

	// Confirm the balance is greater than zero, or else exit with no further action
	if tradeConvertedBalanceAmount.LTE(sdkmath.ZeroInt()) {
		k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeConvertedBalance,
			"Not enough balance of traded tokens yet, balance: %v", tradeConvertedBalanceAmount))
		return nil
	}

	// Using ICA commands on the trade address, transfer the found converted tokens from the trade zone to the host zone
	if err := k.TransferConvertedTokensTradeToHost(ctx, tradeConvertedBalanceAmount, tradeRoute); err != nil {
		return errorsmod.Wrapf(err, "initiating transfer of converted tokens to back to host zone failed")
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_TradeConvertedBalance,
		"Sending discovered converted tokens %v %s from tradeZone back to hostZone",
		tradeConvertedBalanceAmount, tradeRoute.HostDenomOnTradeZone))

	return nil
}
