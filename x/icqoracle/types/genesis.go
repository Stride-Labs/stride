package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Performs basic genesis state validation
func (gs GenesisState) Validate() error {
	// TODO: ???
	return nil
}
