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

var (
	OracleKeyPrefix      = KeyPrefix("oracle")
	MetricKeyPrefix      = KeyPrefix("metric")
	MetricQueueKeyPrefix = KeyPrefix("queue")
)
