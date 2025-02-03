package types

import fmt "fmt"

const (
	ModuleName = "icqoracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	ParamsKey        = []byte("params")
	PriceQueryPrefix = []byte("pricequery")
)

func TokenPriceQueryKey(baseDenom, quoteDenom, poolId string) []byte {
	return []byte(fmt.Sprintf("%s|%s|%s", baseDenom, quoteDenom, poolId))
}

func TokenPriceByDenomKey(baseDenom string) []byte {
	return []byte(baseDenom)
}
