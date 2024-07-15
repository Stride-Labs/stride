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
	AirdropKeyPrefix        = KeyPrefix("airdrops")
	UserAllocationKeyPrefix = KeyPrefix("user-allocations")
)

// Generates a key byte prefix from a string
func KeyPrefix(p string) []byte {
	return []byte(p)
}

func UserAllocationKey(airdropId string, userAddress string) []byte {
	return KeyPrefix(fmt.Sprintf("%s/%s", airdropId, userAddress))
}
