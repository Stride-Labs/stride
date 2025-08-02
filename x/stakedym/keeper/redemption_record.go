package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/stakedym/types"
)

// Writes a redemption record to the store
func (k Keeper) SetRedemptionRecord(ctx sdk.Context, redemptionRecord types.RedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)

	recordKey := types.RedemptionRecordKey(redemptionRecord.UnbondingRecordId, redemptionRecord.Redeemer)
	recordBz := k.cdc.MustMarshal(&redemptionRecord)

	store.Set(recordKey, recordBz)
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
func (k Keeper) GetRedemptionRecordsFromUnbondingId(ctx sdk.Context, unbondingRecordId uint64) (redemptionRecords []types.RedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.RedemptionRecordsKeyPrefix)

	// Iterate though just the records that match the unbonding record ID prefix
	unbondingRecordPrefix := types.IntKey(unbondingRecordId)
	iterator := storetypes.KVStorePrefixIterator(store, unbondingRecordPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		redemptionRecord := types.RedemptionRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &redemptionRecord)
		redemptionRecords = append(redemptionRecords, redemptionRecord)
	}

	return redemptionRecords
}

// Returns all redemption records for a given address
func (k Keeper) GetRedemptionRecordsFromAddress(ctx sdk.Context, address string) (redemptionRecords []types.RedemptionRecord) {
	for _, redemptionRecord := range k.GetAllRedemptionRecords(ctx) {
		if redemptionRecord.Redeemer == address {
			redemptionRecords = append(redemptionRecords, redemptionRecord)
		}
	}
	return redemptionRecords
}
