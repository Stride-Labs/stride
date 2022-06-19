package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
)

// SetControllerBalances set a specific controllerBalances in the store from its index
func (k Keeper) SetControllerBalances(ctx sdk.Context, controllerBalances types.ControllerBalances) {
	store :=  prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ControllerBalancesKeyPrefix))
	b := k.cdc.MustMarshal(&controllerBalances)
	store.Set(types.ControllerBalancesKey(
        controllerBalances.Index,
    ), b)
}

// GetControllerBalances returns a controllerBalances from its index
func (k Keeper) GetControllerBalances(
    ctx sdk.Context,
    index string,
    
) (val types.ControllerBalances, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ControllerBalancesKeyPrefix))

	b := store.Get(types.ControllerBalancesKey(
        index,
    ))
    if b == nil {
        return val, false
    }

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveControllerBalances removes a controllerBalances from the store
func (k Keeper) RemoveControllerBalances(
    ctx sdk.Context,
    index string,
    
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ControllerBalancesKeyPrefix))
	store.Delete(types.ControllerBalancesKey(
	    index,
    ))
}

// GetAllControllerBalances returns all controllerBalances
func (k Keeper) GetAllControllerBalances(ctx sdk.Context) (list []types.ControllerBalances) {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ControllerBalancesKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.ControllerBalances
		k.cdc.MustUnmarshal(iterator.Value(), &val)
        list = append(list, val)
	}

    return
}
