package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v16/utils"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
)

// TradeRewardBalanceCallback is a callback handler for TradeRewardBalance queries.
// The query response will return the trade ICA account balance for a specific (foreign ibc) denom
// If the balance is non-zero, ICA MsgSends are submitted to initiate a swap on the tradeZone
//
// Note: for now, to get proofs in your ICQs, you need to query the entire store on the host zone! e.g. "store/bank/key"
func PoolSpotPriceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_PoolSpotPrice,
		"Starting pool spot price callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	// chainId := query.ChainId // should be the tradeZoneId, used in logging

	// TODO: [DYDX] Fix in separate PR
	// Unmarshal the query response args, should be a SpotPriceResponse type
	// var reponse types.QuerySpotPriceResponse
	// err := reponse.Unmarshal(args)
	// if err != nil {
	// 	return errorsmod.Wrap(err, "unable to unmarshal the query response")
	// }

	// response.SpotPrice should be a string representation of the denom ratios in pool like "10.203"

	// Unmarshal the callback data containing the tradeRoute we are on
	// var tradeRoute types.TradeRoute
	// if err := proto.Unmarshal(query.CallbackData, &tradeRoute); err != nil {
	// 	return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	// }
	// k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_PoolSpotPrice,
	// 	"Query response - spot price ratio of %s to %s is %s",
	// 	tradeRoute.RewardDenomOnTradeZone, tradeRoute.TargetDenomOnTradeZone, reponse.SpotPrice))

	// // Update the spot price stored on the trade route data in the keeper
	// tradeRoute.SpotPrice = reponse.SpotPrice
	// k.SetTradeRoute(ctx, tradeRoute)

	return nil
}
