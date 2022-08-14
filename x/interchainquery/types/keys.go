package types

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
	prefixData  = iota + 1
	prefixQuery = iota + 1
)

// keys for proof queries to various stores
const (
	STAKING_STORE_QUERY_WITH_PROOF = "store/staking/key"
	BANK_STORE_QUERY_WITH_PROOF    = "store/bank/key"
)

var (
	KeyPrefixData  = []byte{prefixData}
	KeyPrefixQuery = []byte{prefixQuery}
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
