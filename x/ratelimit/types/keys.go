package types

import "encoding/binary"

const (
	ModuleName = "ratelimit"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	PathKeyPrefix             = KeyPrefix("path")
	RateLimitKeyPrefix        = KeyPrefix("rate-limit")
	PendingSendPacketPrefix   = KeyPrefix("pending-send-packet")
	DenomBlacklistKeyPrefix   = KeyPrefix("denom-blacklist")
	AddressWhitelistKeyPrefix = KeyPrefix("address-blacklist")

	PendingSendPacketChannelLength int = 16
)

func GetPendingSendPacketKey(channelId string, sequenceNumber uint64) []byte {
	channelIdBz := make([]byte, PendingSendPacketChannelLength)
	copy(channelIdBz[:], channelId)

	sequenceNumberBz := make([]byte, 8)
	binary.BigEndian.PutUint64(sequenceNumberBz, sequenceNumber)

	return append(channelIdBz, sequenceNumberBz...)
}

func GetAddressWhitelistKey(sender, receiver string) []byte {
	return append(KeyPrefix(sender), KeyPrefix(receiver)...)
}
