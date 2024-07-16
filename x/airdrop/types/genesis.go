package types

import "errors"

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Airdrops:        []Airdrop{},
		UserAllocations: []UserAllocation{},
		Params: Params{
			AllocationWindowSeconds: 24 * 60 * 60, // 1 day
		},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Params.AllocationWindowSeconds == 0 {
		return errors.New("allocation window seconds must be set as a module param")
	}
	return nil
}
