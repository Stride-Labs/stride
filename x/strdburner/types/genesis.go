package types

import (
	"fmt"

	"cosmossdk.io/math"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		TotalUstrdBurned: math.ZeroInt(),
	}
}

// Performs basic genesis state validation by
func (gs GenesisState) Validate() error {
	if gs.TotalUstrdBurned.IsNil() {
		return fmt.Errorf("GenesisState.TotalUstrdBurned cannot be nil")
	} else {
		return nil
	}
}
