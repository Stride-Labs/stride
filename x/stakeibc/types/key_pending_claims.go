package types

import "encoding/binary"

var _ binary.ByteOrder

const (
    // PendingClaimsKeyPrefix is the prefix to retrieve all PendingClaims
	PendingClaimsKeyPrefix = "PendingClaims/value/"
)

// PendingClaimsKey returns the store key to retrieve a PendingClaims from the index fields
func PendingClaimsKey(
sequence string,
) []byte {
	var key []byte
    
    sequenceBytes := []byte(sequence)
    key = append(key, sequenceBytes...)
    key = append(key, []byte("/")...)
    
	return key
}