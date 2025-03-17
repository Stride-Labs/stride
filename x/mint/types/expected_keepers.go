package types // noalias

import (
	context "context"

	epochstypes "github.com/Stride-Labs/stride/v26/x/epochs/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountKeeper defines the contract required for account APIs.
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool

	SetAccount(ctx context.Context, acc sdk.AccountI)
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI

	// TODO remove with genesis 2-phases refactor https://github.com/cosmos/cosmos-sdk/issues/2862
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
}

// BankKeeper defines the contract needed to be fulfilled for banking and supply
// dependencies.
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, name string, amt sdk.Coins) error
	BurnCoins(ctx context.Context, name string, amt sdk.Coins) error

	// AddSupplyOffset(ctx sdk.Context, denom string, offsetAmount math.Int)
}

// DistrKeeper defines the contract needed to be fulfilled for distribution keeper.
type DistrKeeper interface {
	FundCommunityPool(ctx context.Context, amount sdk.Coins, sender sdk.AccAddress) error
}

// EpochKeeper defines the contract needed to be fulfilled for epochs keeper.
type EpochKeeper interface {
	GetEpochInfo(ctx sdk.Context, identifier string) (epochstypes.EpochInfo, bool)
}
