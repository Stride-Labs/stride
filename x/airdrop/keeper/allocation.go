package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/airdrop/types"
)

// Writes a user allocation record to the store
func (k Keeper) SetUserAllocation(ctx sdk.Context, userAllocation types.UserAllocation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)

	key := types.UserAllocationKey(userAllocation.AirdropId, userAllocation.Address)
	allocationBz := k.cdc.MustMarshal(&userAllocation)

	store.Set(key, allocationBz)
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
func (k Keeper) GetAllUserAllocations(ctx sdk.Context) (userAllocations []types.UserAllocation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		allocation := types.UserAllocation{}
		k.cdc.MustUnmarshal(iterator.Value(), &allocation)
		userAllocations = append(userAllocations, allocation)
	}

	return userAllocations
}

// Retrieves all the user allocations for a given airdrop
func (k Keeper) GetUserAllocationsForAirdrop(ctx sdk.Context, airdropId string) (userAllocations []types.UserAllocation) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UserAllocationKeyPrefix)

	iterator := storetypes.KVStorePrefixIterator(store, types.KeyPrefix(airdropId))
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		allocation := types.UserAllocation{}
		k.cdc.MustUnmarshal(iterator.Value(), &allocation)
		userAllocations = append(userAllocations, allocation)
	}

	return userAllocations
}

// Retreives all user allocations for a given address
func (k Keeper) GetUserAllocationsForAddress(ctx sdk.Context, address string) (userAllocations []types.UserAllocation) {
	for _, airdrop := range k.GetAllAirdrops(ctx) {
		allocation, found := k.GetUserAllocation(ctx, airdrop.Id, address)
		if found {
			userAllocations = append(userAllocations, allocation)
		}
	}
	return userAllocations
}
