package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Required AccountKeeper functions
type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// Required BankKeeper functions
type BankKeeper interface {
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// Required IcqOracleKeeper functions
type IcqOracleKeeper interface {
	GetTokenPriceForQuoteDenom(ctx sdk.Context, baseDenom string, quoteDenom string) (price math.LegacyDec, err error)
}
