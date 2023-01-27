package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

type StakeibcKeeper interface {
	GetHostZone(ctx sdk.Context, chain_id string) (val stakeibctypes.HostZone, found bool)
	GetStartTimeNextEpoch(ctx sdk.Context, epochType string) (uint64, error)
	IsWithinBufferWindow(ctx sdk.Context) (bool, error)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}
