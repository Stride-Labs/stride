package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/staketia/types"
)

// Writes a slash record to the store
func (k Keeper) SetSlashRecord(ctx sdk.Context, slashRecord types.SlashRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SlashRecordsKeyPrefix)

	key := types.IntKey(slashRecord.Id)
	value := k.cdc.MustMarshal(&slashRecord)

	store.Set(key, value)
}

// Returns all slash records
func (k Keeper) GetAllSlashRecords(ctx sdk.Context) (slashRecords []types.SlashRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SlashRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allSlashRecords := []types.SlashRecord{}
	for ; iterator.Valid(); iterator.Next() {

		slashRecord := types.SlashRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &slashRecord)
		allSlashRecords = append(allSlashRecords, slashRecord)
	}

	return allSlashRecords
}

// Increments the current slash record ID and returns the new ID
func (k Keeper) IncrementSlashRecordId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	currentIdBz := store.Get(types.SlashRecordStoreKeyPrefix)

	// return 1 if there's nothing in the store yet
	currentId := uint64(1)
	if len(currentIdBz) != 0 {
		currentId = binary.BigEndian.Uint64(currentIdBz)
	}

	// Increment the ID
	nextId := currentId + 1
	store.Set(types.SlashRecordStoreKeyPrefix, types.IntKey(nextId))

	return nextId
}
