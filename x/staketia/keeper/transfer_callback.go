package keeper

// Stores the record ID for a pending outbound transfer of native tokens
func (k Keeper) SetTransferInProgressRecordId(channelId string, sequence uint64, recordId uint64) {

}

// Gets the record ID for a pending outbound transfer of native tokens
func (k Keeper) GetTransferInProgressRecordId(channelId string, sequence uint64) (recordId uint64) {
	return recordId
}

// Remove the record ID for a pending outbound transfer of native tokens
// Happens after the packet acknowledement comes back to stride
func (k Keeper) RemoveTransferInProgressRecordId(channelId string, sequence uint64) {
}
