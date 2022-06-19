package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // ControllerBalancesKeyPrefix is the prefix to retrieve all ControllerBalances
	ControllerBalancesKeyPrefix = "ControllerBalances/value/"
)

// ControllerBalancesKey returns the store key to retrieve a ControllerBalances from the index fields
func ControllerBalancesKey(
index string,
) []byte {
	var key []byte
    
    indexBytes := []byte(index)
    key = append(key, indexBytes...)
    key = append(key, []byte("/")...)
    
	return key
}