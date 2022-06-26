package keeper

import (
	"encoding/binary"

	"github.com/Stride-Labs/stride/x/records/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetEpochUnbondingRecordCount get the total number of epochUnbondingRecord
func (k Keeper) GetEpochUnbondingRecordCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.EpochUnbondingRecordCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetEpochUnbondingRecordCount set the total number of epochUnbondingRecord
func (k Keeper) SetEpochUnbondingRecordCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.EpochUnbondingRecordCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendEpochUnbondingRecord appends a epochUnbondingRecord in the store with a new id and update the count
func (k Keeper) AppendEpochUnbondingRecord(
	ctx sdk.Context,
	epochUnbondingRecord types.EpochUnbondingRecord,
) uint64 {
	// Create the epochUnbondingRecord
	count := k.GetEpochUnbondingRecordCount(ctx)

	// Set the ID of the appended value
	epochUnbondingRecord.Id = count

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	appendedValue := k.cdc.MustMarshal(&epochUnbondingRecord)
	store.Set(GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.Id), appendedValue)

	// Update epochUnbondingRecord count
	k.SetEpochUnbondingRecordCount(ctx, count+1)

	return count
}

// SetEpochUnbondingRecord set a specific epochUnbondingRecord in the store
func (k Keeper) SetEpochUnbondingRecord(ctx sdk.Context, epochUnbondingRecord types.EpochUnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	b := k.cdc.MustMarshal(&epochUnbondingRecord)
	store.Set(GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.Id), b)
}

// GetEpochUnbondingRecord returns a epochUnbondingRecord from its id
func (k Keeper) GetEpochUnbondingRecord(ctx sdk.Context, id uint64) (val types.EpochUnbondingRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	b := store.Get(GetEpochUnbondingRecordIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveEpochUnbondingRecord removes a epochUnbondingRecord from the store
func (k Keeper) RemoveEpochUnbondingRecord(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	store.Delete(GetEpochUnbondingRecordIDBytes(id))
}

// GetAllEpochUnbondingRecord returns all epochUnbondingRecord
func (k Keeper) GetAllEpochUnbondingRecord(ctx sdk.Context) (list []types.EpochUnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.EpochUnbondingRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetEpochUnbondingRecordIDBytes returns the byte representation of the ID
func GetEpochUnbondingRecordIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetEpochUnbondingRecordIDFromBytes returns ID in uint64 format from a byte array
func GetEpochUnbondingRecordIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
