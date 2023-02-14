package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v5/modules/apps/27-interchain-accounts/controller/keeper"
	channelkeeper "github.com/cosmos/ibc-go/v5/modules/core/04-channel/keeper"

	icacallbackskeeper "github.com/Stride-Labs/stride/v5/x/icacallbacks/keeper"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramstore paramtypes.Subspace

	channelKeeper       channelkeeper.Keeper
	icaControllerKeeper icacontrollerkeeper.Keeper
	icaCallbacksKeeper  icacallbackskeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramstore paramtypes.Subspace,
	channelKeeper channelkeeper.Keeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	icaCallbacksKeeper icacallbackskeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:                 cdc,
		storeKey:            key,
		paramstore:          paramstore,
		channelKeeper:       channelKeeper,
		icaControllerKeeper: icaControllerKeeper,
		icaCallbacksKeeper:  icaCallbacksKeeper,
	}
}
