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

	ValidatorKey  = "Validator-value-"
	DelegationKey = "Delegation-value-"

	MinValidatorRequirementsKey = "MinValidatorRequirements-value-"
	ICAAccountKey               = "ICAAccount-value-"
)

// prefix bytes for the epoch persistent store
const (
	HostZoneKey            = "HostZone-value-"
	HostZoneCountKey       = "HostZone-count-"
	prefixZone             = iota + 1
	prefixIntent           = iota + 1
	prefixPortMapping      = iota + 1
	prefixReceipt          = iota + 1
	prefixWithdrawalRecord = iota + 1
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey                   = KeyPrefix("stakeibc-port-")
	KeyPrefixZone             = []byte{prefixZone}
	KeyPrefixIntent           = []byte{prefixIntent}
	KeyPrefixPortMapping      = []byte{prefixPortMapping}
	KeyPrefixReceipt          = []byte{prefixReceipt}
	KeyPrefixWithdrawalRecord = []byte{prefixWithdrawalRecord}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
