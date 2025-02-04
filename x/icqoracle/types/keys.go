package types

import (
	fmt "fmt"
)

const (
	ModuleName = "icqoracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName
)

var (
	ParamsKey        = []byte("params")
	TokenPricePrefix = []byte("tokenprice")
)

func TokenPriceKey(baseDenom, quoteDenom string, poolId uint64) []byte {
	return []byte(fmt.Sprintf("%s|%s|%d", baseDenom, quoteDenom, poolId))
}

func TokenPriceByDenomKey(baseDenom string) []byte {
	return []byte(baseDenom)
}
