package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Required AccountKeeper functions
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// Required BankKeeper functions
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	BurnCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
}
