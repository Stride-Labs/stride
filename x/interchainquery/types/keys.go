package types

import (
	fmt "fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	// ModuleName defines the module name
	ModuleName = "interchainquery"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

// prefix bytes for the interchainquery persistent store
const (
	prefixData         = iota + 1
	prefixQuery        = iota + 1
	prefixQueryCounter = iota + 1
)

// keys for proof queries to various stores, note: there's an implicit assumption here that
// the stores on the counterparty chain are prefixed with the standard cosmos-sdk module names
// this might not be true for all IBC chains, and is something we should verify before onboarding a
// new chain

const (
	// The staking store is key'd by the validator's address
	STAKING_STORE_QUERY_WITH_PROOF = "store/staking/key"
	// The bank store is key'd by the account address
	BANK_STORE_QUERY_WITH_PROOF = "store/bank/key"
	// The Osmosis twap store - key'd by the pool ID and denom's
	OSMOSIS_TWAP_STORE_QUERY_WITH_PROOF = "store/twap/key"
	// The Osmosis concentrated liquidity store
	CONCENTRATEDLIQUIDITY_STORE_QUERY_WITH_PROOF = "store/concentratedliquidity/key"
	// The Osmosis gamm store
	GAMM_STORE_QUERY_WITH_PROOF = "store/gamm/key"
)

var (
	// Osmosis TWAP query info
	OsmosisKeySeparator          = "|"
	OsmosisMostRecentTWAPsPrefix = "recent_twap" + OsmosisKeySeparator
)

var (
	KeyPrefixData   = []byte{prefixData}
	KeyPrefixQuery  = []byte{prefixQuery}
	KeyQueryCounter = []byte{prefixQueryCounter}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func FormatOsmosisMostRecentTWAPKey(poolId uint64, denom1, denom2 string) []byte {
	// Sort denoms
	if denom1 > denom2 {
		denom1, denom2 = denom2, denom1
	}

	poolIdBz := fmt.Sprintf("%0.20d", poolId)
	return []byte(fmt.Sprintf("%s%s%s%s%s%s", OsmosisMostRecentTWAPsPrefix, poolIdBz, OsmosisKeySeparator, denom1, OsmosisKeySeparator, denom2))
}

// Source: https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/concentrated-liquidity/types/keys.go#L227-L235
// Used by: https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/concentrated-liquidity/pool.go#L117-L130
// Which is used by: https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/concentrated-liquidity/pool.go#L179-L209
//
// CL Pool Prefix Keys
// Maps a pool id to a pool struct
func FormatOsmosisCLKeyPool(poolId uint64) []byte {
	// Start with PoolPrefix initialized
	poolPrefix := []byte{0x03}
	// Directly append the string representation of poolId as bytes
	return strconv.AppendUint(poolPrefix, poolId, 10)
}

// Source: https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/gamm/types/key.go#L60-L62
// Used by: https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/gamm/keeper/pool.go#L53-L72
// Which is used by: https://github.com/osmosis-labs/osmosis/blob/v27.0.0/x/gamm/keeper/pool.go#L30-L33
//
// Gamm Pool Prefix Keys
// Maps a pool ID to a pool struct
func FormatOsmosisGammKeyPool(poolId uint64) []byte {
	keyPrefixPools := []byte{0x02}
	return append(keyPrefixPools, sdk.Uint64ToBigEndian(poolId)...)
}
