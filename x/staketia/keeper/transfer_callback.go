package keeper

import (
	"encoding/binary"
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/staketia/types"
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

// Get all pending transfers
func (k Keeper) GetAllTransferInProgressId(ctx sdk.Context) (transferInProgressRecordIds []types.TransferInProgressRecordIds) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.TransferInProgressRecordIdKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		channelId := string(key[:types.ChannelIdBufferFixedLength])
		channelId = strings.TrimRight(channelId, "\x00") // removes null bytes from suffix
		sequence := binary.BigEndian.Uint64(key[types.ChannelIdBufferFixedLength:])
		recordId := binary.BigEndian.Uint64(iterator.Value())

		transferInProgressRecordIds = append(transferInProgressRecordIds, types.TransferInProgressRecordIds{
			ChannelId: channelId,
			Sequence:  sequence,
			RecordId:  recordId,
		})
	}

	return transferInProgressRecordIds
}
