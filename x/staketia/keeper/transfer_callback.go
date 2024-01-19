package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// Stores the record ID for a pending outbound transfer of native tokens
func (k Keeper) SetTransferInProgressRecordId(ctx sdk.Context, channelId string, sequence uint64, recordId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferInProgressRecordIdKeyPrefix)

	recordIdKey := types.TransferInProgressRecordKey(channelId, sequence)
	recordIdBz := types.IntKey(recordId) 

	store.Set(recordIdKey, recordIdBz)
}

// Gets the record ID for a pending outbound transfer of native tokens
func (k Keeper) GetTransferInProgressRecordId(ctx sdk.Context, channelId string, sequence uint64) (recordId uint64, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferInProgressRecordIdKeyPrefix)

	recordIdKey := types.TransferInProgressRecordKey(channelId, sequence)
	recordIdBz := store.Get(recordIdKey)

	if len(recordIdBz) == 0 {
		return 0, false
	}

	recordId = binary.BigEndian.Uint64(recordIdBz)
	return recordId, true
}

// Remove the record ID for a pending outbound transfer of native tokens
// Happens after the packet acknowledement comes back to stride
func (k Keeper) RemoveTransferInProgressRecordId(ctx sdk.Context, channelId string, sequence uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferInProgressRecordIdKeyPrefix)
	recordIdKey := types.TransferInProgressRecordKey(channelId, sequence)
	store.Delete(recordIdKey)	
}
