package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ccvtypes "github.com/cosmos/interchain-security/v6/x/ccv/types"

	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
// Methods imported from account should be defined here
type AccountKeeper interface {
	NewAccount(context.Context, sdk.AccountI) sdk.AccountI
	SetAccount(ctx context.Context, acc sdk.AccountI)
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	GetAllAccounts(ctx context.Context) []sdk.AccountI
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	RemoveAccount(ctx context.Context, acc sdk.AccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
// BankKeeper interface: https://github.com/cosmos/cosmos-sdk/blob/main/x/bank/keeper/keeper.go
// Methods imported from bank should be defined here
type BankKeeper interface {
	SpendableCoins(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	MintCoins(ctx context.Context, moduleName string, amt sdk.Coins) error
	GetAllBalances(ctx context.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToModule(ctx context.Context, senderModule string, recipientModule string, amt sdk.Coins) error
}

// Event Hooks
// These can be utilized to communicate between a stakeibc keeper and another
// keeper which must take particular actions when liquid staking happens

// StakeIBCHooks event hooks for stakeibc
type StakeIBCHooks interface {
	AfterLiquidStake(ctx context.Context, addr sdk.AccAddress) // Must be called after liquid stake is completed
}

type ICAOracleKeeper interface {
	QueueMetricUpdate(ctx sdk.Context, key, value, metricType, attributes string)
}

type RatelimitKeeper interface {
	AddDenomToBlacklist(ctx sdk.Context, denom string)
	RemoveDenomFromBlacklist(ctx sdk.Context, denom string)
	SetWhitelistedAddressPair(ctx sdk.Context, whitelist ratelimittypes.WhitelistedAddressPair)
	RemoveWhitelistedAddressPair(ctx sdk.Context, sender, receiver string)
}

type ConsumerKeeper interface {
	GetConsumerParams(ctx sdk.Context) ccvtypes.ConsumerParams
	SetParams(ctx sdk.Context, params ccvtypes.ConsumerParams)
}
