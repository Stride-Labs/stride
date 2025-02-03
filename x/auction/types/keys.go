package types

const (
	ModuleName = "auction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	ParamsKey     = []byte("params")
	AuctionPrefix = []byte("auction")
)
