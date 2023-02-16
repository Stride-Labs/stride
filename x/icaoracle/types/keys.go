package types

const (
	ModuleName = "icaoracle"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// PortID is the default port id that module binds to
	PortID = ModuleName
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

var PortKey = KeyPrefix("icaoracle-port")
