package types

import (
	fmt "fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Performs basic genesis state validation by iterating through all token prices and validating
// using ValidateTokenPriceQueryParams().
// We ignore the SpotPrice, UpdatedAt & QueryInProgress fields since they are reset in InitGenesis().
func (gs GenesisState) Validate() error {
	for i, tokenPrice := range gs.TokenPrices {
		err := ValidateTokenPriceQueryParams(
			tokenPrice.BaseDenom,
			tokenPrice.QuoteDenom,
			tokenPrice.BaseDenomDecimals,
			tokenPrice.QuoteDenomDecimals,
			tokenPrice.OsmosisPoolId,
			tokenPrice.OsmosisBaseDenom,
			tokenPrice.OsmosisQuoteDenom,
		)
		if err != nil {
			return fmt.Errorf("invalid genesis token price query at index %d: %w", i, err)
		}
	}
	return nil
}
