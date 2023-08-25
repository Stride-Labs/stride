package types

const (
	// ModuleName defines the module name
	ModuleName = "records"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_records"

	// Version defines the current version the IBC module supports
	Version = "records-1"

	// PortID is the default port id that module binds to
	PortID = "records"
)

// PortKey defines the key to store the port ID in store
var PortKey = KeyPrefix("records-port-")

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// Create the LSMTokenDeposit prefix as chainId + denom
func GetLSMTokenDepositKey(chainId, denom string) []byte {
	return append([]byte(chainId), []byte(denom)...)
}

const (
	UserRedemptionRecordKey      = "UserRedemptionRecord-value-"
	UserRedemptionRecordCountKey = "UserRedemptionRecord-count-"
)

const (
	EpochUnbondingRecordKey      = "EpochUnbondingRecord-value-"
	EpochUnbondingRecordCountKey = "EpochUnbondingRecord-count-"
	DepositRecordKey             = "DepositRecord-value-"
	DepositRecordCountKey        = "DepositRecord-count-"
	LSMTokenDepositKey           = "LSMTokenDeposit"
)
