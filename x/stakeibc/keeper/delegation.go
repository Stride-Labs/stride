package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// SetDelegation set delegation in the store
func (k Keeper) SetDelegation(ctx sdk.Context, delegation types.Delegation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DelegationKey))
	b := k.cdc.MustMarshal(&delegation)
	store.Set([]byte{0}, b)
}

// GetDelegation returns delegation
func (k Keeper) GetDelegation(ctx sdk.Context) (val types.Delegation, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DelegationKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveDelegation removes delegation from the store
func (k Keeper) RemoveDelegation(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DelegationKey))
	store.Delete([]byte{0})
}
