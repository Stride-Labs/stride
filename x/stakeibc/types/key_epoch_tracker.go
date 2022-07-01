package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// EpochTrackerKeyPrefix is the prefix to retrieve all EpochTracker
	EpochTrackerKeyPrefix = "EpochTracker/value/"
)

// EpochTrackerKey returns the store key to retrieve a EpochTracker from the index fields
func EpochTrackerKey(
	epochIdentifier string,
) []byte {
	var key []byte

	epochIdentifierBytes := []byte(epochIdentifier)
	key = append(key, epochIdentifierBytes...)
	key = append(key, []byte("/")...)

	return key
}
