package types

import (
	"fmt"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Performs basic genesis state validation by iterating through all auctions and validating
// using ValidateBasic() since it already implements thorough validation of all auction fields
func (gs GenesisState) Validate() error {
	for i, auction := range gs.Auctions {

		msg := NewMsgCreateAuction(
			"stride16eenchewedupsplt0ut600ed0ffstageeeervs", // dummy address, not stored in auction
			auction.Type,
			auction.Denom,
			auction.Enabled,
			auction.PriceMultiplier.String(),
			auction.MinBidAmount.Uint64(),
			auction.Beneficiary,
		)

		if err := msg.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid genesis auction at index %d: %w", i, err)
		}
	}
	return nil
}
