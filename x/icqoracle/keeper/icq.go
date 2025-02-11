package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v25/utils"
	deps "github.com/Stride-Labs/stride/v25/x/icqoracle/deps/types"
	cltypes "github.com/Stride-Labs/stride/v25/x/icqoracle/deps/types/concentratedliquidity"
	gammtypes "github.com/Stride-Labs/stride/v25/x/icqoracle/deps/types/gamm"
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

	var queryType string
	var requestData []byte
	switch tokenPrice.OsmosisPoolType {
	case types.GAMM:
		queryType = icqtypes.GAMM_STORE_QUERY_WITH_PROOF
		requestData = icqtypes.FormatOsmosisGammKeyPool(tokenPrice.OsmosisPoolId)
	case types.CONCENTRATED_LIQUIDITY:
		queryType = icqtypes.CONCENTRATEDLIQUIDITY_STORE_QUERY_WITH_PROOF
		requestData = icqtypes.FormatOsmosisCLKeyPool(tokenPrice.OsmosisPoolId)
	default:
		return errorsmod.Wrapf(err, "Unsupported pool type: %d", tokenPrice.OsmosisPoolType)
	}

	query := icqtypes.Query{
		ChainId:         params.OsmosisChainId,
		ConnectionId:    params.OsmosisConnectionId,
		QueryType:       queryType,
		RequestData:     requestData,
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

// Callback from Osmosis spot price query
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

	newSpotPrice, err := UnmarshalSpotPriceFromOsmosisPool(k, tokenPrice, args)
	if err != nil {
		return errorsmod.Wrap(err, "Error determining spot price from query response")
	}

	k.SetQueryComplete(ctx, tokenPrice, newSpotPrice)

	return nil
}

// Unmarshals the Osmosis pool query response and extracts the actual spot price
// Supports both CL and GAMM pools
func UnmarshalSpotPriceFromOsmosisPool(k Keeper, tokenPrice types.TokenPrice, queryResponseBz []byte) (price math.LegacyDec, err error) {
	var pool deps.Pool
	var gammPool gammtypes.OsmosisGammPool
	var clPool cltypes.OsmosisConcentratedLiquidityPool

	switch tokenPrice.OsmosisPoolType {
	case types.GAMM:
		if err := k.cdc.UnmarshalInterface(queryResponseBz, &gammPool); err != nil {
			return math.LegacyZeroDec(), err
		}
		pool = &gammPool
	case types.CONCENTRATED_LIQUIDITY:
		if err := proto.Unmarshal(queryResponseBz, &clPool); err != nil {
			return math.LegacyZeroDec(), err
		}
		pool = &clPool
	default:
		return price, fmt.Errorf("Unsupported pool type: %d", tokenPrice.OsmosisPoolType)
	}

	rawSpotPrice, err := pool.CalcSpotPrice(tokenPrice.OsmosisQuoteDenom, tokenPrice.OsmosisBaseDenom)
	if err != nil {
		return price, err
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
