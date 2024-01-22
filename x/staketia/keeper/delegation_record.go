package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a delegation record to the store based on the status
// If the status is archive, it writes to the archive store, otherwise it writes to the active store
func (k Keeper) SetDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) {
	activeStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)
	archiveStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsArchiveKeyPrefix)

	if delegationRecord.Status == types.DELEGATION_ARCHIVE {
		k.setDelegationRecord(archiveStore, delegationRecord)
	} else {
		k.setDelegationRecord(activeStore, delegationRecord)
	}
}

// Writes a delegation record to a specific store (either active or archive)
func (k Keeper) setDelegationRecord(store prefix.Store, delegationRecord types.DelegationRecord) {
	recordKey := types.IntKey(delegationRecord.Id)
	recordBz := k.cdc.MustMarshal(&delegationRecord)
	store.Set(recordKey, recordBz)
}

// Writes a delegation record to the store only if a record does not already exist for that ID
func (k Keeper) SafelySetDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) error {
	if _, found := k.GetDelegationRecord(ctx, delegationRecord.Id); found {
		return types.ErrDelegationRecordAlreadyExists.Wrapf("delegation record already exists for ID %d", delegationRecord.Id)
	}
	k.SetDelegationRecord(ctx, delegationRecord)
	return nil
}

// Reads a delegation record from the active store
func (k Keeper) GetDelegationRecord(ctx sdk.Context, recordId uint64) (delegationRecord types.DelegationRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)

	recordKey := types.IntKey(recordId)
	recordBz := store.Get(recordKey)

	if len(recordBz) == 0 {
		return delegationRecord, false
	}

	k.cdc.MustUnmarshal(recordBz, &delegationRecord)
	return delegationRecord, true
}

// Reads a delegation record from the archive store
func (k Keeper) GetArchivedDelegationRecord(ctx sdk.Context, recordId uint64) (delegationRecord types.DelegationRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsArchiveKeyPrefix)

	recordKey := types.IntKey(recordId)
	recordBz := store.Get(recordKey)

	if len(recordBz) == 0 {
		return delegationRecord, false
	}

	k.cdc.MustUnmarshal(recordBz, &delegationRecord)
	return delegationRecord, true
}

// Removes a delegation record from the store
// To preserve history, we write it to the archive store
func (k Keeper) ArchiveDelegationRecord(ctx sdk.Context, recordId uint64) {
	activeStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)
	archiveStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsArchiveKeyPrefix)

	recordKey := types.IntKey(recordId)
	recordBz := activeStore.Get(recordKey)

	// No action necessary if the record doesn't exist
	if len(recordBz) == 0 {
		return
	}

	// Update the status to ARCHIVE
	var delegationRecord types.DelegationRecord
	k.cdc.MustUnmarshal(recordBz, &delegationRecord)
	delegationRecord.Status = types.DELEGATION_ARCHIVE
	recordBz = k.cdc.MustMarshal(&delegationRecord)

	// Write the archived record to the store
	archiveStore.Set(recordKey, recordBz)

	// Then remove from active store
	activeStore.Delete(recordKey)
}

// Returns all active delegation records
func (k Keeper) GetAllActiveDelegationRecords(ctx sdk.Context) (delegationRecords []types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)
	return k.getAllDelegationRecords(store)
}

// Returns all active delegation records
func (k Keeper) GetAllArchivedDelegationRecords(ctx sdk.Context) (delegationRecords []types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsArchiveKeyPrefix)
	return k.getAllDelegationRecords(store)
}

// Returns all delegation records for a specified store (either active or archive)
func (k Keeper) getAllDelegationRecords(store prefix.Store) (delegationRecords []types.DelegationRecord) {
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegationRecord := types.DelegationRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &delegationRecord)
		delegationRecords = append(delegationRecords, delegationRecord)
	}

	return delegationRecords
}
