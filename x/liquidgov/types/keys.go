package types

const (
	// ModuleName defines the module name
	ModuleName = "liquidgov"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute is the querier route for the liquidgov store.
	QuerierRoute = StoreKey

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_liquidgov"
)
