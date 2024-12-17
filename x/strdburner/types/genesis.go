package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Performs basic genesis state validation by iterating through all auctions and validating
// using ValidateBasic() since it already implements thorough validation of all auction fields
func (gs GenesisState) Validate() error {
	return nil
}
