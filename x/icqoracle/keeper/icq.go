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

	"github.com/Stride-Labs/stride/v24/utils"
	deps "github.com/Stride-Labs/stride/v24/x/icqoracle/deps/types"
	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
	icqtypes "github.com/Stride-Labs/stride/v24/x/interchainquery/types"
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
	k.Logger(ctx).Info(utils.LogWithPriceToken(tokenPrice, "Submitting OsmosisClPool ICQ"))

	params := k.GetParams(ctx)

	osmosisPoolId, err := strconv.ParseUint(tokenPrice.OsmosisPoolId, 10, 64)
	if err != nil {
		k.Logger(ctx).Error(utils.LogWithPriceToken(tokenPrice, "Error converting osmosis pool id '%s' to uint64, error '%s'", tokenPrice.OsmosisPoolId, err.Error()))
		return err
	}

	tokenPriceBz, err := k.cdc.Marshal(&tokenPrice)
	if err != nil {
		k.Logger(ctx).Error(utils.LogWithPriceToken(tokenPrice, "Error serializing tokenPrice '%+v' to bytes, error '%s'", tokenPrice, err.Error()))
		return err
	}

	queryId := fmt.Sprintf("%s|%s|%s|%d", tokenPrice.BaseDenom, tokenPrice.QuoteDenom, tokenPrice.OsmosisPoolId, ctx.BlockHeight())
	query := icqtypes.Query{
		Id:              queryId,
		ChainId:         params.OsmosisChainId,
		ConnectionId:    params.OsmosisConnectionId,
		QueryType:       icqtypes.CONCENTRATEDLIQUIDITY_STORE_QUERY_WITH_PROOF,
		RequestData:     icqtypes.FormatOsmosisKeyPool(osmosisPoolId),
		CallbackModule:  types.ModuleName,
		CallbackId:      ICQCallbackID_OsmosisClPool,
		CallbackData:    tokenPriceBz,
		TimeoutDuration: time.Duration(params.IcqTimeoutSec) * time.Second,
		TimeoutPolicy:   icqtypes.TimeoutPolicy_RETRY_QUERY_REQUEST,
	}
	if err := k.icqKeeper.SubmitICQRequest(ctx, query, true); err != nil {
		k.Logger(ctx).Error(utils.LogWithPriceToken(tokenPrice, "Error submitting OsmosisClPool ICQ, error '%s'", err.Error()))
		return err
	}

	if err := k.SetTokenPriceQueryInProgress(ctx, tokenPrice, true); err != nil {
		k.Logger(ctx).Error(utils.LogWithPriceToken(tokenPrice, "Error updating queryInProgress=true, error '%s'", err.Error()))
		return err
	}

	return nil
}

func OsmosisClPoolCallback(k Keeper, ctx sdk.Context, args []byte, query icqtypes.Query) error {
	var tokenPrice types.TokenPrice
	if err := k.cdc.Unmarshal(query.CallbackData, &tokenPrice); err != nil {
		return fmt.Errorf("Error deserializing query.CallbackData '%s' as TokenPrice", hex.EncodeToString(query.CallbackData))
	}

	k.Logger(ctx).Info(utils.LogICQCallbackWithPriceToken(tokenPrice, "OsmosisClPool",
		"Starting OsmosisClPool ICQ callback, QueryId: %vs, QueryType: %s, Connection: %s", query.Id, query.QueryType, query.ConnectionId))

	tokenPrice, err := k.GetTokenPrice(ctx, tokenPrice)
	if err != nil {
		return errorsmod.Wrap(err, "Error getting current spot price")
	}

	// TODO review this
	// this should never happen
	if !tokenPrice.QueryInProgress {
		return nil
	}

	// Unmarshal the query response args to determine the balance
	newSpotPrice, err := unmarshalSpotPriceFromOsmosisClPool(tokenPrice, args)
	if err != nil {
		return errorsmod.Wrap(err, "Error determining spot price from query response")
	}

	tokenPrice.SpotPrice = newSpotPrice
	tokenPrice.QueryInProgress = false

	if err := k.SetTokenPrice(ctx, tokenPrice); err != nil {
		return errorsmod.Wrap(err, "Error updating spot price from query response")
	}

	return nil
}

func unmarshalSpotPriceFromOsmosisClPool(tokenPrice types.TokenPrice, queryResponseBz []byte) (price math.LegacyDec, err error) {
	var pool deps.OsmosisConcentratedLiquidityPool
	if err := proto.Unmarshal(queryResponseBz, &pool); err != nil {
		return math.LegacyZeroDec(), err
	}

	spotPrice, err := pool.SpotPrice(tokenPrice.OsmosisQuoteDenom, tokenPrice.OsmosisBaseDenom)
	if err != nil {
		return math.LegacyZeroDec(), err
	}

	return spotPrice, nil
}
