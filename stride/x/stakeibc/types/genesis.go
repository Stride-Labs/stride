package types

import (
	"fmt"
	host "github.com/cosmos/ibc-go/v2/modules/core/24-host"
)

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		ICAAccount:   nil,
		HostZoneList: []HostZone{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
		PortId: PortID,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := host.PortIdentifierValidator(gs.PortId); err != nil {
		return err
	}
	// Check for duplicated ID in hostZone
	hostZoneIdMap := make(map[uint64]bool)
	hostZoneCount := gs.GetHostZoneCount()
	for _, elem := range gs.HostZoneList {
		if _, ok := hostZoneIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for hostZone")
		}
		if elem.Id >= hostZoneCount {
			return fmt.Errorf("hostZone id should be lower or equal than the last id")
		}
		hostZoneIdMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
