package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/utils"
	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v25/x/interchainquery/types"
)

const (
	ICQCallbackID_OsmosisPool = "osmosispool"
)

// ICQCallbacks wrapper struct for stakeibc keeper
type ICQCallback func(Keeper, sdk.Context, []byte, icqtypes.Query) error

type ICQCallbacks struct {
	k         Keeper
	callbacks map[string]ICQCallback
}

var _ icqtypes.QueryCallbacks = ICQCallbacks{}

func (k Keeper) ICQCallbackHandler() ICQCallbacks {
	return ICQCallbacks{k, make(map[string]ICQCallback)}
}

func (c ICQCallbacks) CallICQCallback(ctx sdk.Context, id string, args []byte, query icqtypes.Query) error {
	return c.callbacks[id](c.k, ctx, args, query)
}

func (c ICQCallbacks) HasICQCallback(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c ICQCallbacks) AddICQCallback(id string, fn interface{}) icqtypes.QueryCallbacks {
	c.callbacks[id] = fn.(ICQCallback)
	return c
}

func (c ICQCallbacks) RegisterICQCallbacks() icqtypes.QueryCallbacks {
	return c.
		AddICQCallback(ICQCallbackID_OsmosisPool, ICQCallback(OsmosisPoolCallback))
}

// Submits an ICQ to get a concentrated liquidity pool from Osmosis' store
func (k Keeper) SubmitOsmosisPoolICQ(
	ctx sdk.Context,
	tokenPrice types.TokenPrice,
) error {
	k.Logger(ctx).Info(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Submitting OsmosisPool ICQ"))

	params := k.GetParams(ctx)

	tokenPriceBz, err := k.cdc.Marshal(&tokenPrice)
	if err != nil {
		return errorsmod.Wrapf(err, "Error serializing tokenPrice '%+v' to bytes", tokenPrice)
	}

	queryData := icqtypes.FormatOsmosisMostRecentTWAPKey(
		tokenPrice.OsmosisPoolId,
		tokenPrice.OsmosisBaseDenom,
		tokenPrice.OsmosisQuoteDenom,
	)

	query := icqtypes.Query{
		ChainId:         params.OsmosisChainId,
		ConnectionId:    params.OsmosisConnectionId,
		QueryType:       icqtypes.OSMOSIS_TWAP_STORE_QUERY_WITH_PROOF,
		RequestData:     queryData,
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_OsmosisPool,
		CallbackData:    tokenPriceBz,
		TimeoutDuration: time.Duration(params.UpdateIntervalSec) * time.Second,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}

	if err := k.IcqKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrap(err, "Error submitting OsmosisPool ICQ")
	}

	if err := k.SetQueryInProgress(ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId); err != nil {
		return errorsmod.Wrap(err, "Error updating token price query to in progress")
	}

	return nil
}

// Callback handler for the Omsosis pool spot price query.
func OsmosisPoolCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	var tokenPrice types.TokenPrice
	if err := k.cdc.Unmarshal(query.CallbackData, &tokenPrice); err != nil {
		return fmt.Errorf("Error deserializing query.CallbackData '%s' as TokenPrice", hex.EncodeToString(query.CallbackData))
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "OsmosisPool",
		"Starting OsmosisPool ICQ callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	tokenPrice, err := k.GetTokenPrice(ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	if err != nil {
		return errorsmod.Wrap(err, "Error getting current spot price")
	}

	if !tokenPrice.QueryInProgress {
		return nil
	}

	newSpotPrice, err := UnmarshalSpotPriceFromOsmosis(k, tokenPrice, args)
	if err != nil {
		return errorsmod.Wrap(err, "Error determining spot price from query response")
	}

	k.SetQueryComplete(ctx, tokenPrice, newSpotPrice)

	return nil
}

// Unmarshals the Osmosis pool query response and extracts the actual spot price
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
// When storing down the price, we want to store down the ratio of QuoteDenom to BaseDenom
// Meaning, if Asset0Denom is the QuoteDenom, we want to store P0LastSpotPrice, and vice versa
func UnmarshalSpotPriceFromOsmosis(k Keeper, tokenPrice types.TokenPrice, queryResponseBz []byte) (price math.LegacyDec, err error) {
	var twapRecord types.OsmosisTwapRecord

	if err := twapRecord.Unmarshal(queryResponseBz); err != nil {
		return price, errorsmod.Wrap(err, "unable to unmarshal the query response")
	}

	if err := AssertTwapAssetsMatchTokenPrice(twapRecord, tokenPrice); err != nil {
		return price, err
	}

	// Get the associate "SpotPrice" from the twap record, based on the asset ordering
	// The "SpotPrice" is actually a ratio of the assets in the pool
	if twapRecord.Asset0Denom == tokenPrice.QuoteDenom {
		price = twapRecord.P0LastSpotPrice
	} else {
		price = twapRecord.P1LastSpotPrice
	}

	return AdjustSpotPriceForDecimals(
		price,
		tokenPrice.BaseDenomDecimals,
		tokenPrice.QuoteDenomDecimals,
	), nil
}

// Helper function to confirm that the two assets in the twap record match the assets in the token price
// The assets in the twap record are sorted alphabetically, so we have to check both orderings
func AssertTwapAssetsMatchTokenPrice(twapRecord types.OsmosisTwapRecord, tokenPrice types.TokenPrice) error {
	baseDenomFirst := twapRecord.Asset0Denom == tokenPrice.OsmosisBaseDenom
	quoteDenomSecond := twapRecord.Asset1Denom == tokenPrice.OsmosisQuoteDenom

	quoteDenomFirst := twapRecord.Asset0Denom == tokenPrice.OsmosisQuoteDenom
	baseDenomSecond := twapRecord.Asset1Denom == tokenPrice.OsmosisBaseDenom

	if (baseDenomFirst && quoteDenomSecond) || (quoteDenomFirst && baseDenomSecond) {
		return nil
	}

	return fmt.Errorf("Assets in query response (%s, %s) do not match denom's from token price (%s, %s)",
		twapRecord.Asset0Denom, twapRecord.Asset1Denom, tokenPrice.OsmosisBaseDenom, tokenPrice.OsmosisQuoteDenom)
}

// AdjustSpotPriceForDecimals corrects the spot price to account for different decimal places between tokens
// Example: For BTC (8 decimals) / USDC (6 decimals):
// - If raw price is 1,000 USDC/BTC, we multiply by 10^(8-6) to get 100,000 USDC/BTC
func AdjustSpotPriceForDecimals(rawPrice math.LegacyDec, baseDecimals, quoteDecimals int64) math.LegacyDec {
	decimalsDiff := baseDecimals - quoteDecimals
	if decimalsDiff == 0 {
		return rawPrice
	}

	decimalAdjustmentExp := abs(decimalsDiff)
	decimalAdjustment := math.LegacyNewDec(10).Power(decimalAdjustmentExp)

	if decimalsDiff > 0 {
		return rawPrice.Mul(decimalAdjustment)
	} else {
		return rawPrice.Quo(decimalAdjustment)
	}
}

func abs(num int64) uint64 {
	if num < 0 {
		return uint64(-num)
	}
	return uint64(num)
}
