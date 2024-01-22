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

// Reads a unbonding record from the active store
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
// To preserve history, we write it to the archive store
func (k Keeper) ArchiveUnbondingRecord(ctx sdk.Context, recordId uint64) {
	activeStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsKeyPrefix)
	archiveStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsArchiveKeyPrefix)

	recordKey := types.IntKey(recordId)
	recordBz := activeStore.Get(recordKey)

	// No action necessary if the record doesn't exist
	if len(recordBz) == 0 {
		return
	}

	// Update the status to ARCHIVE
	var unbondingRecord types.UnbondingRecord
	k.cdc.MustUnmarshal(recordBz, &unbondingRecord)
	unbondingRecord.Status = types.UNBONDING_ARCHIVE
	recordBz = k.cdc.MustMarshal(&unbondingRecord)

	// Write the archived record to the store
	archiveStore.Set(recordKey, recordBz)

	// Then remove the original record from the active store
	activeStore.Delete(recordKey)
}

// Returns all active unbonding records
func (k Keeper) GetAllActiveUnbondingRecords(ctx sdk.Context) (unbondingRecords []types.UnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsKeyPrefix)
	return k.getAllUnbondingRecords(store)
}

// Returns all unbonding records that have been archived
func (k Keeper) GetAllArchivedUnbondingRecords(ctx sdk.Context) (unbondingRecords []types.UnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.UnbondingRecordsArchiveKeyPrefix)
	return k.getAllUnbondingRecords(store)
}

// Returns all unbonding records for a specified store (either active or archive)
func (k Keeper) getAllUnbondingRecords(store prefix.Store) (unbondingRecords []types.UnbondingRecord) {
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
// Searches only active records
func (k Keeper) GetAllUnbondingRecordsByStatus(ctx sdk.Context, status types.UnbondingRecordStatus) (unbondingRecords []types.UnbondingRecord) {
	for _, unbondingRecord := range k.GetAllActiveUnbondingRecords(ctx) {
		if unbondingRecord.Status == status {
			unbondingRecords = append(unbondingRecords, unbondingRecord)
		}
	}
	return unbondingRecords
}

// Gets the ACCUMULATING unbonding record (there should only be one)
func (k Keeper) GetAccumulatingUnbondingRecord(ctx sdk.Context) (unbondingRecord types.UnbondingRecord, err error) {
	accumulatingRecord := k.GetAllUnbondingRecordsByStatus(ctx, types.ACCUMULATING_REDEMPTIONS)
	if len(accumulatingRecord) == 0 {
		return unbondingRecord, types.ErrBrokenUnbondingRecordInvariant.Wrap("no unbonding record in status ACCUMULATING")
	}
	if len(accumulatingRecord) != 1 {
		return unbondingRecord, types.ErrBrokenUnbondingRecordInvariant.Wrap("more than one record in status ACCUMULATING")
	}
	return accumulatingRecord[0], nil
}

// Sets the unbonding record only if a record does not already exist for that ID
func (k Keeper) SafelySetUnbondingRecord(ctx sdk.Context, unbondingRecord types.UnbondingRecord) error {
	if _, found := k.GetUnbondingRecord(ctx, unbondingRecord.Id); found {
		return types.ErrUnbondingRecordAlreadyExists.Wrapf("unbonding record already exists for ID %d", unbondingRecord.Id)
	}
	k.SetUnbondingRecord(ctx, unbondingRecord)
	return nil
}
