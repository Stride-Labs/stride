package types

import fmt "fmt"

const (
	ModuleName = "auction"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	ParamsKey        = "params"
	KeyAuctionPrefix = "auction"
)

func AuctionKey(auctionName string) []byte {
	return []byte(fmt.Sprintf("%s|%s", KeyAuctionPrefix, auctionName))
}
