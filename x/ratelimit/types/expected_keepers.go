package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/ratelimit keeper.
type BankKeeper interface {
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
}
