package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// SetICAAccount set iCAAccount in the store
func (k Keeper) SetICAAccount(ctx sdk.Context, iCAAccount types.ICAAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ICAAccountKey))
	b := k.cdc.MustMarshal(&iCAAccount)
	store.Set([]byte{0}, b)
}

// GetICAAccount returns iCAAccount
func (k Keeper) GetICAAccount(ctx sdk.Context) (val types.ICAAccount, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ICAAccountKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveICAAccount removes iCAAccount from the store
func (k Keeper) RemoveICAAccount(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ICAAccountKey))
	store.Delete([]byte{0})
}
