package types

import fmt "fmt"

const (
	// ModuleName defines the module name
	ModuleName = "icacallbacks"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey is the message route for slashing
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_icacallbacks"

	// Version defines the current version the IBC module supports
	Version = "icacallbacks-1"

	// PortID is the default port id that module binds to
	PortID = "icacallbacks"
)

// PortKey defines the key to store the port ID in store
var PortKey = KeyPrefix("icacallbacks-port-")

func KeyPrefix(p string) []byte {
	return []byte(p)
}

func PacketID(portID string, channelID string, sequence uint64) string {
	return fmt.Sprintf("%s.%s.%d", portID, channelID, sequence)
}
