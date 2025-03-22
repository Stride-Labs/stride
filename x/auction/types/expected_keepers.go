package types

import (
	context "context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Required AccountKeeper functions
type AccountKeeper interface {
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// Required BankKeeper functions
type BankKeeper interface {
	BlockedAddr(addr sdk.AccAddress) bool
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx context.Context, from, to sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// Required IcqOracleKeeper functions
type IcqOracleKeeper interface {
	GetTokenPriceForQuoteDenom(ctx sdk.Context, baseDenom string, quoteDenom string) (price math.LegacyDec, err error)
}
