package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
// Methods imported from account should be defined here
type AccountKeeper interface {
	NewAccount(sdk.Context, authtypes.AccountI) authtypes.AccountI
	SetAccount(ctx sdk.Context, acc authtypes.AccountI)
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	GetModuleAccount(ctx sdk.Context, moduleName string) types.ModuleAccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
// BankKeeper interface: https://github.com/cosmos/cosmos-sdk/blob/main/x/bank/keeper/keeper.go
// Methods imported from bank should be defined here
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
}

// Event Hooks
// These can be utilized to communicate between a stakeibc keeper and another
// keeper which must take particular actions when liquid staking happens

// StakeIBCHooks event hooks for stakeibc
type StakeIBCHooks interface {
	AfterLiquidStake(ctx sdk.Context, addr sdk.AccAddress) // Must be called after liquid stake is completed
}
