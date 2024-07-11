package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

// Writes a user allocation record to the store
func (k Keeper) SetUserAllocation(ctx sdk.Context, userAllocation types.UserAllocation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)

	key := types.UserAllocationKey(userAllocation.AirdropId, userAllocation.Address)
	value := k.cdc.MustMarshal(&userAllocation)

	store.Set(key, value)
}

// Retrieves a user allocation record from the store
func (k Keeper) GetUserAllocation(ctx sdk.Context, airdropId, address string) (allocation types.UserAllocation, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)

	key := types.UserAllocationKey(airdropId, address)
	allocationBz := store.Get(key)

	if len(allocationBz) == 0 {
		return allocation, false
	}

	k.cdc.MustUnmarshal(allocationBz, &allocation)
	return allocation, true
}

// Removes a user allocation record from the store
func (k Keeper) RemoveUserAllocation(ctx sdk.Context, airdropId, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)
	key := types.UserAllocationKey(airdropId, address)
	store.Delete(key)
}

// Retrieves all user allocations across all airdrops
func (k Keeper) GetAllUserAllocations(ctx sdk.Context) []types.UserAllocation {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allUserAllocations := []types.UserAllocation{}
	for ; iterator.Valid(); iterator.Next() {

		allocation := types.UserAllocation{}
		k.cdc.MustUnmarshal(iterator.Value(), &allocation)
		allUserAllocations = append(allUserAllocations, allocation)
	}

	return allUserAllocations
}
