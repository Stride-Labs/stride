package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func (k Keeper) GetAllocationRecords(ctx sdk.Context) []types.AirdropRecord {
	// TODO add pagination?
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AllocationRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allAllocations := []types.AirdropRecord{}
	for ; iterator.Valid(); iterator.Next() {

		airdrop := types.AirdropRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &airdrop)
		allAllocations = append(allAllocations, airdrop)
	}

	return allAllocations
}

func (k Keeper) SetAllocationRecords(ctx sdk.Context, allocationRecord []types.AllocationRecord) {
	for _, allocation := range allocationRecord {
		store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AllocationRecordsKeyPrefix)

		key := types.AllocationRecordKeyPrefix(allocation.AirdropId, allocation.UserAddress)
		value := k.cdc.MustMarshal(&allocation)

		store.Set(key, value)
	}
}
