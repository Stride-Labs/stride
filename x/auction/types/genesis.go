package types

import (
	"fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Performs basic genesis state validation by iterating through all auctions and validating
// using ValidateCreateAuctionParams()
func (gs GenesisState) Validate() error {
	for i, auction := range gs.Auctions {
		err := ValidateCreateAuctionParams(
			auction.Name,
			auction.Type,
			auction.SellingDenom,
			auction.PaymentDenom,
			auction.MinPriceMultiplier,
			auction.MinBidAmount,
			auction.Beneficiary,
		)
		if err != nil {
			return fmt.Errorf("invalid genesis auction at index %d: %w", i, err)
		}
	}
	return nil
}
