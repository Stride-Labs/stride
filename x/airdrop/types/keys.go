package types

import (
	fmt "fmt"
)

const (
	// ModuleName defines the module name
	ModuleName = "airdrop"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	AirdropRecordsKeyPrefix    = KeyPrefix("airdrop-records")
	AllocationRecordsKeyPrefix = KeyPrefix("allocation-records")
)

// Generates a key byte prefix from a string
func KeyPrefix(p string) []byte {
	return []byte(p)
}

func AllocationRecordKeyPrefix(airdropId uint64, userAddress string) []byte {
	return KeyPrefix(fmt.Sprintf("%d/%s", airdropId, userAddress))
}
