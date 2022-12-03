package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/records/types"
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

// SetDepositRecordCount set the total number of depositRecord
func (k Keeper) SetDepositRecordCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.DepositRecordCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendDepositRecord appends a depositRecord in the store with a new id and update the count
func (k Keeper) AppendDepositRecord(
	ctx sdk.Context,
	depositRecord types.DepositRecord,
) uint64 {
	// Create the depositRecord
	count := k.GetDepositRecordCount(ctx)

	// Set the ID of the appended value
	depositRecord.Id = count

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	appendedValue := k.Cdc.MustMarshal(&depositRecord)
	store.Set(GetDepositRecordIDBytes(depositRecord.Id), appendedValue)

	// Update depositRecord count
	k.SetDepositRecordCount(ctx, count+1)

	return count
}

// SetDepositRecord set a specific depositRecord in the store
func (k Keeper) SetDepositRecord(ctx sdk.Context, depositRecord types.DepositRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	b := k.Cdc.MustMarshal(&depositRecord)
	store.Set(GetDepositRecordIDBytes(depositRecord.Id), b)
}

// GetDepositRecord returns a depositRecord from its id
func (k Keeper) GetDepositRecord(ctx sdk.Context, id uint64) (val types.DepositRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	b := store.Get(GetDepositRecordIDBytes(id))
	if b == nil {
		return val, false
	}
	k.Cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveDepositRecord removes a depositRecord from the store
func (k Keeper) RemoveDepositRecord(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	store.Delete(GetDepositRecordIDBytes(id))
}

// GetAllDepositRecord returns all depositRecord
func (k Keeper) GetAllDepositRecord(ctx sdk.Context) (list []types.DepositRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.DepositRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.DepositRecord
		k.Cdc.MustUnmarshal(iterator.Value(), &val)
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

func (k Keeper) GetDepositRecordByEpochAndChain(ctx sdk.Context, epochNumber uint64, chainId string) (val *types.DepositRecord, found bool) {
	records := k.GetAllDepositRecord(ctx)
	for _, depositRecord := range records {
		if depositRecord.DepositEpochNumber == epochNumber && depositRecord.HostZoneId == chainId {
			return &depositRecord, true
		}
	}
	return nil, false
}
