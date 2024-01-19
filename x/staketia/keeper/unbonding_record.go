package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a unbonding record to the store
func (k Keeper) SetUnbondingRecord(ctx sdk.Context, unbondingRecord types.UnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsKeyPrefix)

	recordKey := types.IntKey(unbondingRecord.Id)
	recordValue := k.cdc.MustMarshal(&unbondingRecord)

	store.Set(recordKey, recordValue)
}

// Reads a unbonding record from the store
func (k Keeper) GetUnbondingRecord(ctx sdk.Context, recordId uint64) (unbondingRecord types.UnbondingRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsKeyPrefix)

	recordKey := types.IntKey(recordId)
	recordBz := store.Get(recordKey)

	if len(recordBz) == 0 {
		return unbondingRecord, false
	}

	k.cdc.MustUnmarshal(recordBz, &unbondingRecord)
	return unbondingRecord, true
}

// Removes a unbonding record from the store
func (k Keeper) RemoveUnbondingRecord(ctx sdk.Context, recordId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsKeyPrefix)
	recordKey := types.IntKey(recordId)
	store.Delete(recordKey)
}

// Returns all unbonding records
func (k Keeper) GetAllUnbondingRecords(ctx sdk.Context) (unbondingRecords []types.UnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		unbondingRecord := types.UnbondingRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &unbondingRecord)
		unbondingRecords = append(unbondingRecords, unbondingRecord)
	}

	return unbondingRecords
}

// Returns all unbonding records with a specific status
func (k Keeper) GetAllUnbondingRecordsByStatus(ctx sdk.Context, status types.UnbondingRecordStatus) (unbondingRecords []types.UnbondingRecord) {
	for _, unbondingRecord := range k.GetAllUnbondingRecords(ctx) {
		if unbondingRecord.Status == status {
			unbondingRecords = append(unbondingRecords, unbondingRecord)
		}
	}
	return unbondingRecords
}

// Gets the ACCUMULATING unbonding record (there should only be one)
func (k Keeper) GetAccumulatingUnbondingRecord(ctx sdk.Context) (unbondingRecord types.UnbondingRecord, err error) {
	// QUESTION: This is kind of inefficient - do you think it's worth indexing instead of looping each time?
	accumulatingRecord := k.GetAllUnbondingRecordsByStatus(ctx, types.ACCUMULATING_REDEMPTIONS)
	if len(accumulatingRecord) == 0 {
		return unbondingRecord, types.ErrBrokenUnbondingRecordInvariant.Wrap("no unbonding record in status ACCUMULATING")
	}
	if len(accumulatingRecord) != 1 {
		return unbondingRecord, types.ErrBrokenUnbondingRecordInvariant.Wrap("more than one record in status ACCUMULATING")
	}
	return accumulatingRecord[0], nil
}
