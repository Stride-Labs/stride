package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/ingenuity-build/quicksilver/x/interchainquery/types"
)

func GenerateQueryHash(connection_id string, chain_id string, query_type string, query_params map[string]string) string {
	param_bytes, _ := json.Marshal(query_params)
	return fmt.Sprintf("%x", crypto.Sha256(append([]byte(connection_id+chain_id+query_type), param_bytes...)))
}

// ----------------------------------------------------------------

func (k Keeper) NewQuery(ctx sdk.Context, connection_id string, chain_id string, query_type string, query_params map[string]string, period sdk.Int) *types.Query {
	return &types.Query{Id: GenerateQueryHash(connection_id, chain_id, query_type, query_params), ConnectionId: connection_id, ChainId: chain_id, QueryType: query_type, QueryParameters: query_params, Period: period, LastHeight: sdk.ZeroInt()}
}

// GetQuery returns query
func (k Keeper) GetQuery(ctx sdk.Context, id string) (types.Query, bool) {
	query := types.Query{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixQuery)
	bz := store.Get([]byte(id))
	if len(bz) == 0 {
		return query, false
	}
	k.cdc.MustUnmarshal(bz, &query)
	return query, true
}

// SetQuery set query info
func (k Keeper) SetQuery(ctx sdk.Context, query types.Query) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixQuery)
	bz := k.cdc.MustMarshal(&query)
	k.Logger(ctx).Info("Created/updated query", "ID", query.Id)
	store.Set([]byte(query.Id), bz)
}

// DeleteQuery delete query info
func (k Keeper) DeleteQuery(ctx sdk.Context, id string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixQuery)
	store.Delete([]byte(id))
}

// IterateQueries iterate through querys
func (k Keeper) IterateQueries(ctx sdk.Context, fn func(index int64, queryInfo types.Query) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixQuery)
	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)
	for ; iterator.Valid(); iterator.Next() {
		query := types.Query{}
		k.cdc.MustUnmarshal(iterator.Value(), &query)
		stop := fn(i, query)

		if stop {
			break
		}
		i++
	}
}

// AllQueries returns every queryInfo in the store
func (k Keeper) AllQueries(ctx sdk.Context) []types.Query {
	querys := []types.Query{}
	k.IterateQueries(ctx, func(_ int64, queryInfo types.Query) (stop bool) {
		querys = append(querys, queryInfo)
		return false
	})
	return querys
}
