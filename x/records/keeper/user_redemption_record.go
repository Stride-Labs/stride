package keeper

import (
	"encoding/binary"

	"github.com/Stride-Labs/stride/x/records/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetUserRedemptionRecordCount get the total number of userRedemptionRecord
func (k Keeper) GetUserRedemptionRecordCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.UserRedemptionRecordCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetUserRedemptionRecordCount set the total number of userRedemptionRecord
func (k Keeper) SetUserRedemptionRecordCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.UserRedemptionRecordCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendUserRedemptionRecord appends a userRedemptionRecord in the store with a new id and update the count
func (k Keeper) AppendUserRedemptionRecord(
	ctx sdk.Context,
	userRedemptionRecord types.UserRedemptionRecord,
) uint64 {
	// Create the userRedemptionRecord
	count := k.GetUserRedemptionRecordCount(ctx)

	// Set the ID of the appended value
	userRedemptionRecord.Id = count

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	appendedValue := k.cdc.MustMarshal(&userRedemptionRecord)
	store.Set(GetUserRedemptionRecordIDBytes(userRedemptionRecord.Id), appendedValue)

	// Update userRedemptionRecord count
	k.SetUserRedemptionRecordCount(ctx, count+1)

	return count
}

// SetUserRedemptionRecord set a specific userRedemptionRecord in the store
func (k Keeper) SetUserRedemptionRecord(ctx sdk.Context, userRedemptionRecord types.UserRedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	b := k.cdc.MustMarshal(&userRedemptionRecord)
	store.Set(GetUserRedemptionRecordIDBytes(userRedemptionRecord.Id), b)
}

// GetUserRedemptionRecord returns a userRedemptionRecord from its id
func (k Keeper) GetUserRedemptionRecord(ctx sdk.Context, id uint64) (val types.UserRedemptionRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	b := store.Get(GetUserRedemptionRecordIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveUserRedemptionRecord removes a userRedemptionRecord from the store
func (k Keeper) RemoveUserRedemptionRecord(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	store.Delete(GetUserRedemptionRecordIDBytes(id))
}

// GetAllUserRedemptionRecord returns all userRedemptionRecord
func (k Keeper) GetAllUserRedemptionRecord(ctx sdk.Context) (list []types.UserRedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.UserRedemptionRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetUserRedemptionRecordIDBytes returns the byte representation of the ID
func GetUserRedemptionRecordIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetUserRedemptionRecordIDFromBytes returns ID in uint64 format from a byte array
func GetUserRedemptionRecordIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
