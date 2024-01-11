package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "autopilot"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	TransferFallbackAddressPrefix = []byte("fallback")

	FallbackAddressChannelPrefixLength int = 16
)

// Builds the store key for a fallback address, key'd by channel ID and sequence number
// The serialized channelId is set to a fixed array size to assist deserialization
func GetTransferFallbackAddressKey(channelId string, sequenceNumber uint64) []byte {
	channelIdBz := make([]byte, FallbackAddressChannelPrefixLength)
	copy(channelIdBz[:], channelId)

	sequenceNumberBz := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceNumberBz, sequenceNumber)

	return append(channelIdBz, sequenceNumberBz...)
}
