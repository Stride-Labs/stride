package types

import fmt "fmt"

const (
	ModuleName = "icqoracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the routing key
	RouterKey = ModuleName

	ParamsPrefix      = "params"
	KeyPricePrefix    = "price"
	KeyLastUpdateTime = "last_update_time"
)

func TokenPriceKey(baseDenom, quoteDenom, poolId string) []byte {
	return []byte(fmt.Sprintf("%s%s%s%s", KeyPricePrefix, baseDenom, quoteDenom, poolId))
}

func TokenPriceByDenomKey(baseDenom string) []byte {
	return []byte(fmt.Sprintf("%s%s", KeyPricePrefix, baseDenom))
}
