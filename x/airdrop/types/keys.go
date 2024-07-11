package types

const (
	// ModuleName defines the module name
	ModuleName = "airdrop"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	// Module Account for Fee Collection
	FeeAddress = "staketia_fee_address"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
