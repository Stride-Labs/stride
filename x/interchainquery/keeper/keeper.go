package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/ingenuity-build/quicksilver/x/interchainquery/types"
)

// Keeper of this module maintains collections of registered zones.
type Keeper struct {
	cdc       codec.Codec
	storeKey  sdk.StoreKey
	callbacks map[string]types.QueryCallbacks
	IBCKeeper *ibckeeper.Keeper
}

// NewKeeper returns a new instance of zones Keeper
func NewKeeper(cdc codec.Codec, storeKey sdk.StoreKey, ibckeeper *ibckeeper.Keeper) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		callbacks: make(map[string]types.QueryCallbacks),
		IBCKeeper: ibckeeper,
	}
}

func (k *Keeper) SetCallbackHandler(module string, handler types.QueryCallbacks) error {
	_, found := k.callbacks[module]
	if found {
		return fmt.Errorf("callback handler already set for %s", module)
	}
	k.callbacks[module] = handler.RegisterCallbacks()
	return nil
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) SetDatapointForId(ctx sdk.Context, id string, result []byte, height sdk.Int) error {
	mapping := types.DataPoint{Id: id, RemoteHeight: height, LocalHeight: sdk.NewInt(ctx.BlockHeight()), Value: result}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixData)
	bz := k.cdc.MustMarshal(&mapping)
	store.Set([]byte(id), bz)
	return nil
}

func (k *Keeper) GetDatapointForId(ctx sdk.Context, id string) (types.DataPoint, error) {
	mapping := types.DataPoint{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixData)
	bz := store.Get([]byte(id))
	if len(bz) == 0 {
		return types.DataPoint{}, fmt.Errorf("unable to find data for id %s", id)
	}

	k.cdc.MustUnmarshal(bz, &mapping)
	return mapping, nil
}

// IterateDatapoints iterate through datapoints
func (k Keeper) IterateDatapoints(ctx sdk.Context, fn func(index int64, dp types.DataPoint) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixData)
	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		datapoint := types.DataPoint{}
		k.cdc.MustUnmarshal(iterator.Value(), &datapoint)
		stop := fn(i, datapoint)

		if stop {
			break
		}
		i++
	}
}

// DeleteQuery delete datapoint
func (k Keeper) DeleteDatapoint(ctx sdk.Context, id string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixData)
	store.Delete([]byte(id))
}

func (k *Keeper) GetDatapoint(ctx sdk.Context, module string, connection_id string, chain_id string, query_type string, request []byte) (types.DataPoint, error) {
	id := GenerateQueryHash(connection_id, chain_id, query_type, request, module)
	return k.GetDatapointForId(ctx, id)
}

func (k *Keeper) GetDatapointOrRequest(ctx sdk.Context, module string, connection_id string, chain_id string, query_type string, request []byte, max_age uint64) (types.DataPoint, error) {
	val, err := k.GetDatapoint(ctx, module, connection_id, chain_id, query_type, request)
	if err != nil {
		// no datapoint
		k.MakeRequest(ctx, connection_id, chain_id, query_type, request, sdk.NewInt(-1), "", "", max_age)
		return types.DataPoint{}, fmt.Errorf("no data; query submitted")
	}

	if val.LocalHeight.LT(sdk.NewInt(ctx.BlockHeight() - int64(max_age))) { // this is somewhat arbitrary; TODO: make this better
		k.MakeRequest(ctx, connection_id, chain_id, query_type, request, sdk.NewInt(-1), "", "", max_age)
		return types.DataPoint{}, fmt.Errorf("stale data; query submitted")
	}
	// check ttl
	return val, nil
}

func (k *Keeper) MakeRequest(ctx sdk.Context, connection_id string, chain_id string, query_type string, request []byte, period sdk.Int, module string, callback_id string, ttl uint64) {
	k.Logger(ctx).Info(
		"MakeRequest",
		"connection_id", connection_id,
		"chain_id", chain_id,
		"query_type", query_type,
		"request", request,
		"period", period,
		"module", module,
		"callback", callback_id,
		"ttl", ttl,
	)
	key := GenerateQueryHash(connection_id, chain_id, query_type, request, module)
	existingQuery, found := k.GetQuery(ctx, key)
	if !found {
		if module != "" {
			if _, exists := k.callbacks[module]; !exists {
				err := fmt.Errorf("no callback handler registered for module %s", module)
				k.Logger(ctx).Error(err.Error())
				panic(err)
			}
			if exists := k.callbacks[module].Has(callback_id); !exists {
				err := fmt.Errorf("no callback %s registered for module %s", callback_id, module)
				k.Logger(ctx).Error(err.Error())
				panic(err)
			}
		}
		newQuery := k.NewQuery(ctx, module, connection_id, chain_id, query_type, request, period, callback_id, ttl)
		k.SetQuery(ctx, *newQuery)

	} else {
		// a re-request of an existing query triggers resetting of height to trigger immediately.
		existingQuery.LastHeight = sdk.ZeroInt()
		k.SetQuery(ctx, existingQuery)
	}
}
