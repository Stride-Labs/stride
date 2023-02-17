package types

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

// Generates a byte key for a pending metric update
// (i.e. a metric update ICA that has been sent to an specific oracle)
func GetPendingMetricKey(metricKey string, oracleMoniker string) []byte {
	return append([]byte(metricKey), []byte(oracleMoniker)...)
}

var (
	OracleKeyPrefix        = KeyPrefix("oracle")
	MetricQueueKeyPrefix   = KeyPrefix("metric-queue")
	MetricPendingKeyPrefix = KeyPrefix("metric-pending")
)
