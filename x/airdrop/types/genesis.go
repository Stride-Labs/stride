package types

import "errors"

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Airdrops:        []Airdrop{},
		UserAllocations: []UserAllocation{},
		Params: Params{
			PeriodLengthSeconds: 24 * 60 * 60, // 1 day
		},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if gs.Params.PeriodLengthSeconds == 0 {
		return errors.New("allocation window seconds must be set as a module param")
	}
	return nil
}
