package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/stakedym/types"
)

// Writes a delegation record to the active store
func (k Keeper) SetDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)

	recordKey := types.IntKey(delegationRecord.Id)
	recordBz := k.cdc.MustMarshal(&delegationRecord)

	store.Set(recordKey, recordBz)
}

// Writes a delegation record to the archive store
func (k Keeper) SetArchivedDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsArchiveKeyPrefix)

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

// Removes a delegation record from the active store
func (k Keeper) RemoveDelegationRecord(ctx sdk.Context, recordId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)
	recordKey := types.IntKey(recordId)
	store.Delete(recordKey)
}

// Removes a delegation record from the active store and writes it to the archive store,
// to preserve history
func (k Keeper) ArchiveDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) {
	k.RemoveDelegationRecord(ctx, delegationRecord.Id)
	k.SetArchivedDelegationRecord(ctx, delegationRecord)
}

// Returns all active delegation records
func (k Keeper) GetAllActiveDelegationRecords(ctx sdk.Context) (delegationRecords []types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)
	delegationRecordsInActiveStore := k.getAllDelegationRecords(store)

	// There should only be TRANSFER_IN_PROGRESS or DELEGATION_QUEUE records in this store
	// up we'll add the check here to be safe
	for _, delegationRecord := range delegationRecordsInActiveStore {
		if delegationRecord.Status == types.TRANSFER_IN_PROGRESS || delegationRecord.Status == types.DELEGATION_QUEUE {
			delegationRecords = append(delegationRecords, delegationRecord)
		}
	}
	return delegationRecords
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
