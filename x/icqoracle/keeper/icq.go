package keeper

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v25/utils"
	deps "github.com/Stride-Labs/stride/v25/x/icqoracle/deps/types"
	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v25/x/interchainquery/types"
)

const (
	ICQCallbackID_OsmosisClPool = "osmosisclpool"
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
		AddICQCallback(ICQCallbackID_OsmosisClPool, ICQCallback(OsmosisClPoolCallback))
}

// Submits an ICQ to get a concentrated liquidity pool from Osmosis' store
func (k Keeper) SubmitOsmosisClPoolICQ(
	ctx sdk.Context,
	tokenPrice types.TokenPrice,
) error {
	k.Logger(ctx).Info(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Submitting OsmosisClPool ICQ"))

	params, err := k.GetParams(ctx)
	if err != nil {
		k.Logger(ctx).Error(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Error getting module params: %s", err.Error()))
		return errorsmod.Wrapf(err, "Error getting module params")
	}

	osmosisPoolId, err := strconv.ParseUint(tokenPrice.OsmosisPoolId, 10, 64)
	if err != nil {
		k.Logger(ctx).Error(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Error converting osmosis pool id '%s' to uint64, error '%s'", tokenPrice.OsmosisPoolId, err.Error()))
		return errorsmod.Wrapf(err, "Error converting osmosis pool id '%s' to uint64", tokenPrice.OsmosisPoolId)
	}

	tokenPriceBz, err := k.cdc.Marshal(&tokenPrice)
	if err != nil {
		k.Logger(ctx).Error(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Error serializing tokenPrice '%+v' to bytes, error '%s'", tokenPrice, err.Error()))
		return errorsmod.Wrapf(err, "Error serializing tokenPrice '%+v' to bytes", tokenPrice)
	}

	query := icqtypes.Query{
		ChainId:         params.OsmosisChainId,
		ConnectionId:    params.OsmosisConnectionId,
		QueryType:       icqtypes.CONCENTRATEDLIQUIDITY_STORE_QUERY_WITH_PROOF,
		RequestData:     icqtypes.FormatOsmosisKeyPool(osmosisPoolId),
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_OsmosisClPool,
		CallbackData:    tokenPriceBz,
		TimeoutDuration: time.Duration(params.IcqTimeoutSec) * time.Second,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_REJECT_QUERY_RESPONSE,
	}
	if err := k.IcqKeeper.SubmitICQRequest(ctx, query, true); err != nil {
		k.Logger(ctx).Error(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Error submitting OsmosisClPool ICQ, error '%s'", err.Error()))
		return errorsmod.Wrapf(err, "Error submitting OsmosisClPool ICQ")
	}

	if err := k.SetTokenPriceQueryInProgress(ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, true); err != nil {
		k.Logger(ctx).Error(utils.LogWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "Error updating queryInProgress=true, error '%s'", err.Error()))
		return errorsmod.Wrapf(err, "Error updating queryInProgress=true")
	}

	return nil
}

func OsmosisClPoolCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	var tokenPrice types.TokenPrice
	if err := k.cdc.Unmarshal(query.CallbackData, &tokenPrice); err != nil {
		return fmt.Errorf("Error deserializing query.CallbackData '%s' as TokenPrice", hex.EncodeToString(query.CallbackData))
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithTokenPriceQuery(tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, "OsmosisClPool",
		"Starting OsmosisClPool ICQ callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	tokenPrice, err := k.GetTokenPrice(ctx, tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId)
	if err != nil {
		return errorsmod.Wrap(err, "Error getting current spot price")
	}

	// TODO review this
	// this should never happen
	if !tokenPrice.QueryInProgress {
		return nil
	}

	// Unmarshal the query response args to determine the balance
	newSpotPrice, err := UnmarshalSpotPriceFromOsmosisClPool(tokenPrice, args)
	if err != nil {
		return errorsmod.Wrap(err, "Error determining spot price from query response")
	}

	tokenPrice.SpotPrice = newSpotPrice
	tokenPrice.QueryInProgress = false
	tokenPrice.UpdatedAt = ctx.BlockTime()

	if err := k.SetTokenPrice(ctx, tokenPrice); err != nil {
		return errorsmod.Wrap(err, "Error updating spot price from query response")
	}

	return nil
}

func UnmarshalSpotPriceFromOsmosisClPool(tokenPrice types.TokenPrice, queryResponseBz []byte) (price math.LegacyDec, err error) {
	var pool deps.OsmosisConcentratedLiquidityPool
	if err := proto.Unmarshal(queryResponseBz, &pool); err != nil {
		return math.LegacyZeroDec(), err
	}

	rawSpotPrice, err := pool.SpotPrice(tokenPrice.OsmosisQuoteDenom, tokenPrice.OsmosisBaseDenom)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	return AdjustSpotPriceForDecimals(
		rawSpotPrice,
		tokenPrice.BaseDenomDecimals,
		tokenPrice.QuoteDenomDecimals,
	), nil
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
