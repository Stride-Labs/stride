package types

const (
	// ModuleName defines the module name
	ModuleName = "auction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_auction"

	// ParamsKey defines the store key for claim module parameters
	ParamsKey string = "params"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	AuctionPoolKey      = "AuctionPool/value/"
	AuctionPoolCountKey = "AuctionPool/count/"
)
