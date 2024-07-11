package types

const (
	// ModuleName defines the module name
	ModuleName = "airdrop"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_airdrop"

    
)

var (
	ParamsKey = []byte("p_airdrop")
)



func KeyPrefix(p string) []byte {
    return []byte(p)
}
