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

	// fee account
	// TODO(TEST-174): this is a random testing address, update this before launch
	FeeAccount = "stride19uvw0azm9u0k6vqe4e22cga6kteskdqq3ulj6q"
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("stakeibc-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	HostZoneKey      = "HostZone-value-"
	HostZoneCountKey = "HostZone-count-"
)
