package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"

	"github.com/Stride-Labs/stride/v9/x/interchainquery/types"
)

func GenerateQueryHash(connectionId string, chainId string, queryType string, request []byte, module string, callbackId string) string {
	return fmt.Sprintf("%x", crypto.Sha256(append([]byte(module+connectionId+chainId+queryType+callbackId), request...)))
}

func (k Keeper) NewQuery(ctx sdk.Context, module string, callbackId string, chainId string, connectionId string, queryType string, request []byte, ttl uint64) *types.Query {
	return &types.Query{
		Id:           GenerateQueryHash(connectionId, chainId, queryType, request, module, callbackId),
		ConnectionId: connectionId,
		ChainId:      chainId,
		QueryType:    queryType,
		Request:      request,
		CallbackId:   callbackId,
		Ttl:          ttl,
		RequestSent:  false,
	}
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
	store.Set([]byte(query.Id), bz)
}

// DeleteQuery delete query info
func (k Keeper) DeleteQuery(ctx sdk.Context, id string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixQuery)
	store.Delete([]byte(id))
}

// IterateQueries iterate through queries
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
	queries := []types.Query{}
	k.IterateQueries(ctx, func(_ int64, queryInfo types.Query) (stop bool) {
		queries = append(queries, queryInfo)
		return false
	})
	return queries
}

// Helper function to unmarshal a Balance query response across SDK versions
// Before SDK v46, the query response returned a sdk.Coin type. SDK v46 returns an int type
// https://github.com/cosmos/cosmos-sdk/pull/9832
func UnmarshalAmountFromBalanceQuery(cdc codec.BinaryCodec, queryResponseBz []byte) (amount sdkmath.Int, err error) {
	// An nil should not be possible, exit immediately if it occurs
	if queryResponseBz == nil {
		return sdkmath.Int{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "query response is nil")
	}

	// If the query response is empty, that means the account was never registed (and thus has a 0 balance)
	if len(queryResponseBz) == 0 {
		return sdkmath.ZeroInt(), nil
	}

	// First attempt to unmarshal as an Int (for SDK v46+)
	// If the result was serialized as a `Coin` type, it should contain a string (representing the denom)
	// which will cause the unmarshalling to throw an error
	intError := amount.Unmarshal(queryResponseBz)
	if intError == nil {
		return amount, nil
	}

	// If the Int unmarshaling was unsuccessful, attempt again using a Coin type (for SDK v45 and below)
	// If successful, return the amount field from the coin (if the coin is not nil)
	var coin sdk.Coin
	coinError := cdc.Unmarshal(queryResponseBz, &coin)
	if coinError == nil {
		return coin.Amount, nil
	}

	// If it failed unmarshaling with either data structure, return an error with the failure messages combined
	return sdkmath.Int{}, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
		"unable to unmarshal balance query response %v as sdkmath.Int (err: %s) or sdk.Coin (err: %s)", queryResponseBz, intError.Error(), coinError.Error())
}
