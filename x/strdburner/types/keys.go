package types

const (
	ModuleName = "strdburner"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	TotalStrdBurnedKey       = "total_burned"
	ProtocolStrdBurnedKey    = "total_protocol_burned"
	TotalUserStrdBurnedKey   = "total_user_burned"
	BurnedByAddressKeyPrefix = []byte("burned_by_address")
)
