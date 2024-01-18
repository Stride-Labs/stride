package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Writes a redemption record to the store
func (k Keeper) SetRedemptionRecord(ctx sdk.Context, redemptionRecord types.RedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)

	recordKey := types.RedemptionRecordKey(redemptionRecord.UnbondingRecordId, redemptionRecord.Redeemer)
	recordValue := k.cdc.MustMarshal(&redemptionRecord)

	store.Set(recordKey, recordValue)
}

// Reads a redemption record from the store
func (k Keeper) GetRedemptionRecord(ctx sdk.Context, unbondingRecordId uint64, address string) (redemptionRecord types.RedemptionRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)

	recordKey := types.RedemptionRecordKey(unbondingRecordId, address)
	recordBz := store.Get(recordKey)

	if len(recordBz) == 0 {
		return redemptionRecord, false
	}

	k.cdc.MustUnmarshal(recordBz, &redemptionRecord)
	return redemptionRecord, true
}

// Removes a redemption record from the store
func (k Keeper) RemoveRedemptionRecord(ctx sdk.Context, unbondingRecordId uint64, address string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)
	recordKey := types.RedemptionRecordKey(unbondingRecordId, address)
	store.Delete(recordKey)
}

// Returns all redemption records
func (k Keeper) GetAllRedemptionRecords(ctx sdk.Context) (redemptionRecords []types.RedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		redemptionRecord := types.RedemptionRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &redemptionRecord)
		redemptionRecords = append(redemptionRecords, redemptionRecord)
	}

	return redemptionRecords
}

// Returns all redemption records for a given unbonding record
func (k Keeper) GetAllRedemptionRecordsFromUnbondingId(ctx sdk.Context, unbondingRecordId uint64) (redemptionRecords []types.RedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)

	// Iterate though just the records that match the unbonding record ID prefix
	unbondingRecordPrefix := types.IntKey(unbondingRecordId)
	iterator := store.Iterator(unbondingRecordPrefix, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		redemptionRecord := types.RedemptionRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &redemptionRecord)
		redemptionRecords = append(redemptionRecords, redemptionRecord)
	}

	return redemptionRecords
}
