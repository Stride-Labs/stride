package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
)

// SetPendingClaims set a specific pendingClaims in the store from its index
func (k Keeper) SetPendingClaims(ctx sdk.Context, pendingClaims types.PendingClaims) {
	store :=  prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingClaimsKeyPrefix))
	b := k.cdc.MustMarshal(&pendingClaims)
	store.Set(types.PendingClaimsKey(
        pendingClaims.Sequence,
    ), b)
}

// GetPendingClaims returns a pendingClaims from its index
func (k Keeper) GetPendingClaims(
    ctx sdk.Context,
    sequence string,
    
) (val types.PendingClaims, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingClaimsKeyPrefix))

	b := store.Get(types.PendingClaimsKey(
        sequence,
    ))
    if b == nil {
        return val, false
    }

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePendingClaims removes a pendingClaims from the store
func (k Keeper) RemovePendingClaims(
    ctx sdk.Context,
    sequence string,
    
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingClaimsKeyPrefix))
	store.Delete(types.PendingClaimsKey(
	    sequence,
    ))
}

// GetAllPendingClaims returns all pendingClaims
func (k Keeper) GetAllPendingClaims(ctx sdk.Context) (list []types.PendingClaims) {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PendingClaimsKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PendingClaims
		k.cdc.MustUnmarshal(iterator.Value(), &val)
        list = append(list, val)
	}

    return
}
