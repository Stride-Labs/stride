package types

import (
	fmt "fmt"
	"strconv"
)

const (
	// ModuleName defines the module name
	ModuleName = "airdrop"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	// Module Account for Fee Collection
	FeeAddress = "staketia_fee_address"
)

var (
	AirdropRecordsKeyPrefix    = KeyPrefix("airdrop-records")
	AllocationRecordsKeyPrefix = KeyPrefix("allocation-records")
)

// Generates a key byte prefix from a string
func KeyPrefix(p string) []byte {
	return []byte(p)
}

func AirdropRecordKeyPrefix(airdropId uint64) []byte {
	return KeyPrefix(strconv.FormatUint(airdropId, 10))
}

func AllocationRecordKeyPrefix(airdropId uint64, userAddress string) []byte {
	return KeyPrefix(fmt.Sprintf("/%d/%s", airdropId, userAddress))
}
