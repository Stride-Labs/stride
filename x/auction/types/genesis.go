package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// ZeroAddress is a valid 20 byte canonical address (all zeros) used as a placeholder address
// during genesis validation. We need this because MsgCreateAuction requires an admin address
// parameter, but during genesis validation we only care about validating the auction parameters
// themselves, not the admin who created them. Using a zero address allows us to validate auction
// properties without needing a real cryptographic address, while still maintaining compatibility
// with the MsgCreateAuction validation logic.
var ZeroAddress = sdk.AccAddress{
	0, 0, 0, 0, 0,
	0, 0, 0, 0, 0,
	0, 0, 0, 0, 0,
	0, 0, 0, 0, 0,
}

// Performs basic genesis state validation by iterating through all auctions and validating
// using ValidateBasic() since it already implements thorough validation of all auction fields
func (gs GenesisState) Validate() error {
	for i, auction := range gs.Auctions {

		msg := NewMsgCreateAuction(
			ZeroAddress.String(),
			auction.Type,
			auction.Denom,
			auction.Enabled,
			auction.PriceMultiplier.String(),
			auction.MinBidAmount.Uint64(),
			auction.Beneficiary,
		)

		if err := msg.ValidateBasic(); err != nil {
			return fmt.Errorf("invalid genesis auction at index %d: %s", i, err.Error())
		}
	}
	return nil
}
