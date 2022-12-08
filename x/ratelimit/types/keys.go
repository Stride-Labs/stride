package types

const (
	ModuleName = "ratelimit" // IBC at the end to avoid conflicts with the ibc prefix

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// QuotaKey defines the store key for quotas
	QuotaKeyPrefix = "quota"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var (
	PathKeyPrefix      = KeyPrefix("path")
	RateLimitKeyPrefix = KeyPrefix("rate-limit")
)
