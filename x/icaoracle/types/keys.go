package types

import (
	"encoding/binary"
	fmt "fmt"
)

const (
	ModuleName = "icaoracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// Generates a key byte prefix from a string
func KeyPrefix(p string) []byte {
	return []byte(p)
}

// Generates a byte key for the pending metric update substore
// (i.e. stores all ICAs have been submitted for a given metric + oracle combo)
func GetPendingMetricSubstoreKey(metricKey string, oracleChainId string) []byte {
	return KeyPrefix(fmt.Sprintf("%s-%s", metricKey, oracleChainId))
}

// Returns the key to the pending metric update substore
// The key is build from the metric's timestamp
func GetPendingMetricKey(metricTime uint64) []byte {
	key := make([]byte, 8)
	binary.BigEndian.PutUint64(key, metricTime)
	return key
}

var (
	OracleKeyPrefix        = KeyPrefix("oracle")
	MetricQueueKeyPrefix   = KeyPrefix("metric-queue")
	MetricPendingKeyPrefix = KeyPrefix("metric-pending")
)
