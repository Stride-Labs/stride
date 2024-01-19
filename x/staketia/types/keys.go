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
	TransferInProgressRecordIdKeyPrefix = []byte("transfer-in-progress")
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
