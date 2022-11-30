package types

const (
	// ModuleName defines the module name
	ModuleName = "approuter"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// Version defines the current version the IBC module supports
	Version = "app_router-1"

	// PortID is the default port id that module binds to
	PortID = "app_router"
)

var (
	// PortKey defines the key to store the port ID in store
	PortKey = KeyPrefix("app_router-port-")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
