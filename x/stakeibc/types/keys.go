package types

const (
	// ModuleName defines the module name
	ModuleName = "stakeibc"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_stakeibc"

	// Version defines the current version the IBC module supports
	Version = "stakeibc-1"

	// PortID is the default port id that module binds to
	PortID = "stakeibc"

	// fee account - F1
	FeeAccount = "stride1czvrk3jkvtj8m27kqsqu2yrkhw3h3ykwj3rxh6"

	RewardCollectorName = "reward_collector"
)

// PortKey defines the key to store the port ID in store
var PortKey = KeyPrefix("stakeibc-port-")

// Generates a key byte prefix from a string
func KeyPrefix(p string) []byte {
	return []byte(p)
}

// EpochTrackerKey returns the store key to retrieve a EpochTracker from the index fields
func EpochTrackerKey(epochIdentifier string) []byte {
	var key []byte

	epochIdentifierBytes := []byte(epochIdentifier)
	key = append(key, epochIdentifierBytes...)
	key = append(key, []byte("/")...)

	return key
}

// Definition for the store key format based on tradeRoute start and end denoms
func TradeRouteKeyFromDenoms(rewardDenom, hostDenom string) (key []byte) {
	return []byte(rewardDenom + "-" + hostDenom)
}

const (
	// Host zone keys prefix the HostZone structs
	HostZoneKey = "HostZone-value-"

	// EpochTrackerKeyPrefix is the prefix to retrieve all EpochTracker
	EpochTrackerKeyPrefix = "EpochTracker/value/"

	// TradeRoute keys prefix to retrieve all TradeZones
	TradeRouteKeyPrefix = "TradeRoute-value-"
)
