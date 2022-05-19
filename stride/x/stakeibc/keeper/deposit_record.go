package keeper

import (
	"encoding/binary"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetDepositRecordCount get the total number of depositRecord
func (k Keeper) GetDepositRecordCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.DepositRecordCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// GetDepositRecord returns a depositRecord from its id
func (k Keeper) GetDepositRecord(ctx sdk.Context, id uint64) (val types.DepositRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	b := store.Get(GetDepositRecordIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllDepositRecord returns all depositRecord
func (k Keeper) GetAllDepositRecord(ctx sdk.Context) (list []types.DepositRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.DepositRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetDepositRecordIDBytes returns the byte representation of the ID
func GetDepositRecordIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetDepositRecordIDFromBytes returns ID in uint64 format from a byte array
func GetDepositRecordIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
