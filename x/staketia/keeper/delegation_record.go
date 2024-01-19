package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a delegation record to the store
func (k Keeper) SetDelegationRecord(ctx sdk.Context, delegationRecord types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)

	recordKey := types.IntKey(delegationRecord.Id)
	recordBz := k.cdc.MustMarshal(&delegationRecord)

	store.Set(recordKey, recordBz)
}

// Reads a delegation record from the store
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

// Removes a delegation record from the store
func (k Keeper) RemoveDelegationRecord(ctx sdk.Context, recordId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)
	recordKey := types.IntKey(recordId)
	store.Delete(recordKey)
}

// Returns all delegation records
func (k Keeper) GetAllDelegationRecords(ctx sdk.Context) (delegationRecords []types.DelegationRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelegationRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		delegationRecord := types.DelegationRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &delegationRecord)
		delegationRecords = append(delegationRecords, delegationRecord)
	}

	return delegationRecords
}

// Updates the status on a delegation record
func (k Keeper) UpdateDelegationRecordStatus(ctx sdk.Context, recordId uint64, status types.DelegationRecordStatus) error {
	delegationRecord, found := k.GetDelegationRecord(ctx, recordId)
	if !found {
		return types.ErrDelegationRecordNotFound.Wrapf("delegation record not found for %d", recordId)
	}
	delegationRecord.Status = status
	k.SetDelegationRecord(ctx, delegationRecord)
	return nil
}
