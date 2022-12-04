package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	epochstypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
)

// BankKeeper defines the banking contract that must be fulfilled when
// creating a x/claim keeper.
type BankKeeper interface {
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	SetModuleAccount(ctx sdk.Context, macc authtypes.ModuleAccountI)
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
	SetAccount(sdk.Context, authtypes.AccountI)
	NewAccountWithAddress(sdk.Context, sdk.AccAddress) authtypes.AccountI
	// Fetch the sequence of an account at a specified address.
	GetSequence(sdk.Context, sdk.AccAddress) (uint64, error)
}

// DistrKeeper is the keeper of the distribution store
type DistrKeeper interface {
	FundCommunityPool(ctx sdk.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// StakingKeeper expected staking keeper (noalias)
type StakingKeeper interface {
	// BondDenom - Bondable coin denomination
	BondDenom(sdk.Context) string
}

// EpochsKeeper expected epoch keeper
type EpochsKeeper interface {
	SetEpochInfo(ctx sdk.Context, epoch epochstypes.EpochInfo)
	DeleteEpochInfo(ctx sdk.Context, identifier string)
	AllEpochInfos(ctx sdk.Context) []epochstypes.EpochInfo
}
