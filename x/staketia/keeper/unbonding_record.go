package keeper

import (
	"errors"

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

// Updates the status on a unbonding record
func (k Keeper) UpdateUnbondingRecordStatus(ctx sdk.Context, recordId uint64, status types.UnbondingRecordStatus) error {
	unbondingRecord, found := k.GetUnbondingRecord(ctx, recordId)
	if !found {
		return types.ErrUnbondingRecordNotFound.Wrapf("unbonding record not found for %d", recordId)
	}
	unbondingRecord.Status = status
	k.SetUnbondingRecord(ctx, unbondingRecord)
	return nil
}

// Gets the TALLYING unbonding record (there should only be one)
func (k Keeper) GetTallyingUnbondingRecord(ctx sdk.Context) (unbondingRecord types.UnbondingRecord, err error) {
	// QUESTION: This is kind of inefficient - do you think it's worth indexing instead of looping each time?
	tallyRecords := k.GetAllUnbondingRecordsByStatus(ctx, types.TALLYING_REDEMPTIONS)
	if len(tallyRecords) == 0 {
		return unbondingRecord, errors.New("no unbonding record in status TALLYING")
	}
	if len(tallyRecords) != 1 {
		return unbondingRecord, errors.New("more than one record in status TALLYING")
	}
	return tallyRecords[0], nil
}
