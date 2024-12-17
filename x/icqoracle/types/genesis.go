package types

import (
	fmt "fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Performs basic genesis state validation by iterating through all token prices and validating
// using ValidateBasic() since it already implements thorough validation of the important token
// price fields.
// We ignore the SpotPrice, UpdatedAt & QueryInProgress fields since they are reset in InitGenesis().
func (gs GenesisState) Validate() error {
	for i, tokenPrice := range gs.TokenPrices {

		msg := NewMsgRegisterTokenPriceQuery(
			"stride1palmssweatykneesweakarmsareheavy8ahm9u", // dummy address, not stored in token price
			tokenPrice.BaseDenom,
			tokenPrice.QuoteDenom,
			tokenPrice.OsmosisPoolId,
			tokenPrice.OsmosisBaseDenom,
			tokenPrice.OsmosisQuoteDenom,
		)

		if err := msg.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid genesis token price at index %d: %w", i, err)
		}
	}
	return nil
}
