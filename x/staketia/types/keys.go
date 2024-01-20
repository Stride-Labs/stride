package types

import "encoding/binary"

const (
	ModuleName = "staketia"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	// Prefix store keys
	HostZoneKey                         = []byte("host-zone")
	DelegationRecordsKeyPrefix          = []byte("delegation-records-active")
	DelegationRecordsArchiveKeyPrefix   = []byte("delegation-records-archive")
	UnbondingRecordsKeyPrefix           = []byte("unbonding-records-active")
	UnbondingRecordsArchiveKeyPrefix    = []byte("unbonding-records-archive")
	RedemptionRecordsKeyPrefix          = []byte("redemption-records")
	SlashRecordsKeyPrefix               = []byte("slash-records")
	SlashRecordStoreKeyPrefix           = []byte("slash-record-id")
	TransferInProgressRecordIdKeyPrefix = []byte("transfer-in-progress")

	ChannelIdBufferFixedLength int = 16
)

// Serializes an string to use as a prefix when needed
func StringKey(p string) []byte {
	return []byte(p)
}

// Serializes an int to use as a prefix when needed
func IntKey(i uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, i)
	return bz
}

// Builds the redemption record key from an unbonding record ID and address
func RedemptionRecordKey(unbondingRecordId uint64, redeemerAddress string) []byte {
	return append(IntKey(unbondingRecordId), StringKey(redeemerAddress)...)
}

// Builds the transfer-in-progress record key from the channelId and sequence number
func TransferInProgressRecordKey(channelId string, sequence uint64) []byte {
	channelIdBz := make([]byte, ChannelIdBufferFixedLength)
	copy(channelIdBz[:], channelId)
	return append(channelIdBz, IntKey(sequence)...)
}
