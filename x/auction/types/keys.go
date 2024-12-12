package types

import fmt "fmt"

const (
	ModuleName = "auction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	ParamsPrefix     = "params"
	KeyAuctionPrefix = "auction"
)

func AuctionKey(denom string) []byte {
	return []byte(fmt.Sprintf("%s%s", KeyAuctionPrefix, denom))
}
