package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v30/utils"
	"github.com/Stride-Labs/stride/v30/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v30/x/interchainquery/types"
)

const (
	ICQCallbackID_OsmosisPrice = "osmosisprice"
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
		AddICQCallback(ICQCallbackID_OsmosisPrice, ICQCallback(OsmosisPriceCallback))
}

// Submits an ICQ to get a concentrated liquidity pool from Osmosis' store
func (k Keeper) SubmitOsmosisPriceICQ(
	ctx sdk.Context,
	tokenPrice types.TokenPrice,
) error {
	k.Logger(ctx).Info(fmt.Sprintf("Submitting OsmosisPrice ICQ - Base: %s / Quote: %s / Pool: %d",
		tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId))

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
		CallbackId:      ICQCallbackID_OsmosisPrice,
		CallbackData:    tokenPriceBz,
		TimeoutDuration: time.Duration(utils.UintToInt(params.UpdateIntervalSec)) * time.Second,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}

	if err := k.IcqKeeper.SubmitICQRequest(ctx, query, false); err != nil {
		return errorsmod.Wrap(err, "Error submitting OsmosisPrice ICQ")
	}

	if err := k.SetQueryInProgress(ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId); err != nil {
		return errorsmod.Wrap(err, "Error updating token price query to in progress")
	}

	return nil
}

// Callback handler for the Omsosis pool spot price query.
func OsmosisPriceCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	var tokenPrice types.TokenPrice
	if err := k.cdc.Unmarshal(query.CallbackData, &tokenPrice); err != nil {
		return fmt.Errorf("Error deserializing query.CallbackData '%s' as TokenPrice", hex.EncodeToString(query.CallbackData))
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone("osmosis", query.CallbackId,
		"Starting OsmosisPrice ICQ callback, QueryId: %vs, QueryType: %s, Connection: %s, Base Denom: %s, Quote Denom: %s, PoolId: %d",
		query.Id, query.QueryType, query.ConnectionId, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId))

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

	k.Logger(ctx).Info(utils.LogICQCallbackWithHostZone("osmosis", query.CallbackId,
		"Price of %s in terms of %s: %vs", tokenPrice.BaseDenom, tokenPrice.QuoteDenom, newSpotPrice))

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
// ** When storing down the price, we want to store down the ratio of QuoteDenom to BaseDenom **
// This will give us the human readable price
//
// Ex: Let's say the price of OSMO is $2 and it's in a pool with USDC
// This means 1 OSMO is equal to 2 USDC, and there should be twice as much USDC in the pool
// The ratio of OSMO:USDC is 0.5 and the ratio of USDC:OSMO is 2.0
// Since we want to store the quote denom in terms of base denom, we want to store USDC:OSMO = 2
//
// In this example, if Asset0 was USDC and Asset1 was OSMO, we would want to store (Asset0Denom / Asset1Denom),
// since we want USDC / OSMO, so we would store P0LastSpotPrice
//
// However, if Asset0 was OSMO and Asset1 was USDC, we would want (Asset1Denom / Asset0Denom),
// since we want USDC / OSMO, so we would store P1LastSpotPrice
//
// To summarize, we check if Asset0 is equal to our quote denom (USDC in the example), and if
// it is, we store P0; otherwise, we store P1
func UnmarshalSpotPriceFromOsmosis(k Keeper, tokenPrice types.TokenPrice, queryResponseBz []byte) (price sdkmath.LegacyDec, err error) {
	var twapRecord types.OsmosisTwapRecord

	if err := twapRecord.Unmarshal(queryResponseBz); err != nil {
		return price, errorsmod.Wrap(err, "unable to unmarshal the query response")
	}

	if err := AssertTwapAssetsMatchTokenPrice(twapRecord, tokenPrice); err != nil {
		return price, err
	}

	// Get the associate "SpotPrice" from the twap record, based on the asset ordering
	// The "SpotPrice" is actually a ratio of the assets in the pool
	if twapRecord.Asset0Denom == tokenPrice.OsmosisQuoteDenom {
		price = twapRecord.P0LastSpotPrice
	} else {
		price = twapRecord.P1LastSpotPrice
	}

	return price, nil
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
