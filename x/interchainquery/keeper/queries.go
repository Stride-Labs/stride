package keeper

import (
	"encoding/binary"
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	connectiontypes "github.com/cosmos/ibc-go/v5/modules/core/03-connection/types"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/tendermint/tendermint/crypto"

	"github.com/Stride-Labs/stride/v8/x/interchainquery/types"
)

// Generates a query ID based on the request information
// If forceUnique is false, queries of the same request type will have the same query ID
//  (e.g. "query the ATOM balance of address X", will always have the same query ID)
// If forceUnique is true, a unique suffix will be appended to the query ID
//  so that the ID is different for queries of the same type
func (k Keeper) GetQueryId(ctx sdk.Context, query types.Query, forceUnique bool) string {
	queryKey := append([]byte(query.CallbackModule+query.ConnectionId+query.ChainId+query.QueryType+query.CallbackId), query.RequestData...)

	// If forceUnique is true, grab and append the unique query UID
	if forceUnique {
		queryUID := k.GetQueryUID(ctx)
		queryKey = append(queryKey, queryUID...)
	}
	return fmt.Sprintf("%x", crypto.Sha256(queryKey))
}

// ValidateQuery validates that all the required attributes of a query are supplied when submitting an ICQ
func (k Keeper) ValidateQuery(ctx sdk.Context, query types.Query) error {
	if query.ConnectionId == "" {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "connection-id cannot be empty")
	}
	if !strings.HasPrefix(query.ConnectionId, connectiontypes.ConnectionPrefix) {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "invalid connection ID (%s)", query.ConnectionId)
	}
	if query.ChainId == "" {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "chain ID cannot be empty")
	}
	if query.CallbackModule == "" {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "module must be specified")
	}
	if query.CallbackId == "" {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "callback ID cannot be empty")
	}
	if query.QueryType == "" {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "query type cannot be empty")
	}
	if query.Timeout == 0 {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "TTL must be specified and non-zero")
	}
	if _, exists := k.callbacks[query.CallbackModule]; !exists {
		return errorsmod.Wrapf(types.ErrInvalidICQRequest, "no callback handler registered for module %s", query.CallbackModule)
	}
	if exists := k.callbacks[query.CallbackModule].HasICQCallback(query.CallbackId); !exists {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "callback %s is not registered for module %s", query.CallbackId, query.CallbackModule)
	}

	return nil
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

// To optionally force queries to be unique, a UID suffix can be supplied to the query Id
// This is implemented by a counter that increments every time a UID is retrieved
// Return the uid as a byte array since it is used in the serialized query key
func (k Keeper) GetQueryUID(ctx sdk.Context) []byte {
	store := ctx.KVStore(k.storeKey)
	uidBz := store.Get(types.KeyQueryUID)

	// Initialize the UID if there is nothing in the store yet, otherwise deserialize it
	uid := uint64(1)
	if len(uidBz) > 0 {
		uid = binary.BigEndian.Uint64(uidBz)
	}

	// Reset the uid after 1M
	// In practice, this is not necessary, but in theory, we could have int overflow
	if uid > 1_000_000 {
		uid = 1
	}

	// Increment the UID
	nextUidBz := make([]byte, 8)
	binary.BigEndian.PutUint64(nextUidBz, uid+1)

	return uidBz
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
