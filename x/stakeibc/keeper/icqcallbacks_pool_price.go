package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v16/utils"
	icqtypes "github.com/Stride-Labs/stride/v16/x/interchainquery/types"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

// PoolPriceCallback is a callback handler for PoolPrice query.
// The query response returns an Osmosis TwapRecord for the associated pool denom's
//
// The assets in the response are identified by indicies and are sorted alphabetically
// (e.g. if the two denom's are ibc/AXXX, and ibc/BXXX,
// then Asset0Denom is ibc/AXXX and Asset1Denom is ibc/BXXX)
//
// The price fields (P0LastSpotPrice and P1LastSpotPrice) represent the relative
// ratios of tokens in the pool
//
//	P0LastSpotPrice gives the ratio of Asset0Denom / Asset1Denom
//	P1LastSpotPrice gives the ratio of Asset1Denom / Asset0Denom
//
// When storing down the price, we want to store down the ratio of HostDenom.
// Meaning, if Asset0Denom is the host denom, we want to store P0LastSpotPrice
func PoolPriceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(query.ChainId, ICQCallbackID_PoolPrice,
		"Starting pool spot price callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	chainId := query.ChainId // should be the tradeZoneId, used in logging

	// Unmarshal the query response args, should be a TwapRecord type
	var twapRecord types.OsmosisTwapRecord
	err := twapRecord.Unmarshal(args)
	if err != nil {
		return errorsmod.Wrap(err, "unable to unmarshal the query response")
	}

	// Unmarshal the callback data containing the tradeRoute we are on
	var tradeRoute types.TradeRoute
	if err := proto.Unmarshal(query.CallbackData, &tradeRoute); err != nil {
		return errorsmod.Wrapf(err, "unable to unmarshal trade reward balance callback data")
	}

	// Confirm the denom's from the query response match the denom's in the route
	if err := AssertTwapAssetsMatchTradeRoute(twapRecord, tradeRoute); err != nil {
		return err
	}

	// Get the associate "SpotPrice" from the twap record, based on the asset ordering
	// The "SpotPrice" is actually a ratio of the assets in the pool
	var price sdk.Dec
	if twapRecord.Asset0Denom == tradeRoute.HostDenomOnTradeZone {
		price = twapRecord.P0LastSpotPrice
	} else {
		price = twapRecord.P1LastSpotPrice
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone(chainId, ICQCallbackID_PoolPrice,
		"Query response - price ratio of %s to %s is %s",
		tradeRoute.RewardDenomOnTradeZone, tradeRoute.HostDenomOnTradeZone, price))

	// Update the price and time on the trade route data
	tradeRoute.SwapPrice = price
	tradeRoute.PriceUpdateTimestamp = uint64(ctx.BlockTime().UnixNano())
	k.SetTradeRoute(ctx, tradeRoute)

	return nil
}

// Helper function to confirm that the two assets in the twap record match the assets in the trade route
// The assets in the twap record are sorted alphabetically, so we have to check both orderings
func AssertTwapAssetsMatchTradeRoute(twapRecord types.OsmosisTwapRecord, tradeRoute types.TradeRoute) error {
	hostDenomMatchFirst := twapRecord.Asset0Denom == tradeRoute.HostDenomOnTradeZone
	rewardDenomMatchSecond := twapRecord.Asset1Denom == tradeRoute.RewardDenomOnTradeZone

	rewardDenomMatchFirst := twapRecord.Asset0Denom == tradeRoute.RewardDenomOnTradeZone
	hostDenomMatchSecond := twapRecord.Asset1Denom == tradeRoute.HostDenomOnTradeZone

	if (hostDenomMatchFirst && rewardDenomMatchSecond) || (rewardDenomMatchFirst && hostDenomMatchSecond) {
		return nil
	}

	return fmt.Errorf("Assets in query response (%s, %s) do not match denom's from trade route (%s, %s)",
		twapRecord.Asset0Denom, twapRecord.Asset1Denom, tradeRoute.HostDenomOnTradeZone, tradeRoute.RewardDenomOnTradeZone)
}
