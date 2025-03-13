package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

// Required AccountKeeper functions
type AccountKeeper interface {
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	GetModuleAddress(name string) sdk.AccAddress
}

// Required BankKeeper functions
type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error
	BurnCoins(ctx sdk.Context, name string, amt sdk.Coins) error
	GetSupply(ctx sdk.Context, denom string) sdk.Coin
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddress sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule string, recipientModule string, amt sdk.Coins) error
}

// Required TransferKeeper functions
type TransferKeeper interface {
	Transfer(goCtx context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error)
}

// Required RatelimitKeeper functions
type RatelimitKeeper interface {
	AddDenomToBlacklist(ctx sdk.Context, denom string)
	RemoveDenomFromBlacklist(ctx sdk.Context, denom string)
}

// Required ICAOracleKeeper functions
type ICAOracleKeeper interface {
	QueueMetricUpdate(ctx sdk.Context, key, value, metricType, attributes string)
}

// Required StakeibcKeeper functions
type StakeibcKeeper interface {
	GetHostZone(ctx sdk.Context, chainId string) (val stakeibctypes.HostZone, found bool)
	GetActiveHostZone(ctx sdk.Context, chainId string) (hostZone stakeibctypes.HostZone, err error)
	SetHostZone(ctx sdk.Context, hostZone stakeibctypes.HostZone)
	RedeemStake(ctx sdk.Context, msg *stakeibctypes.MsgRedeemStake) (*stakeibctypes.MsgRedeemStakeResponse, error)
	EnableRedemptions(ctx sdk.Context, chainId string) error
	RegisterHostZone(ctx sdk.Context, msg *stakeibctypes.MsgRegisterHostZone) (*stakeibctypes.MsgRegisterHostZoneResponse, error)
}

// Required RecordsKeeper functions
type RecordsKeeper interface {
	GetAllDepositRecord(ctx sdk.Context) (list []recordtypes.DepositRecord)
	SetDepositRecord(ctx sdk.Context, depositRecord recordtypes.DepositRecord)
}
