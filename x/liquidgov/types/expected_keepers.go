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
	GetHostZoneFromHostDenom(ctx sdk.Context, denom string) (*stakeibctypes.HostZone, error)
	GetConnectionId(ctx sdk.Context, portId string) (string, error)
	SubmitTxsStrideEpoch(ctx sdk.Context, connectionId string, msgs []sdk.Msg, account stakeibctypes.ICAAccount, callbackId string, callbackArgs []byte) (uint64, error)
	GetICATimeoutNanos(ctx sdk.Context, epochType string) (uint64, error)
	GetChainID(ctx sdk.Context, connectionID string) (string, error)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(
		ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
	) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	// Methods imported from bank should be defined here
}
