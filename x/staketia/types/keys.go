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
	DelegationRecordsKeyPrefix          = []byte("delegation-records")
	UnbondingRecordsKeyPrefix           = []byte("unbonding-records")
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
