package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"

	icacallbackskeeper "github.com/Stride-Labs/stride/v12/x/icacallbacks/keeper"
	"github.com/Stride-Labs/stride/v12/x/icaoracle/types"
)

type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   storetypes.StoreKey
	paramstore paramtypes.Subspace
	authority  string

	ICS4Wrapper         types.ICS4Wrapper
	IBCKeeper           ibckeeper.Keeper
	ICAControllerKeeper icacontrollerkeeper.Keeper
	ICACallbacksKeeper  icacallbackskeeper.Keeper
}

func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramstore paramtypes.Subspace,
	authority string,

	ics4Wrapper types.ICS4Wrapper,
	ibcKeeper ibckeeper.Keeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	icaCallbacksKeeper icacallbackskeeper.Keeper,
) *Keeper {
	return &Keeper{
		cdc:        cdc,
		storeKey:   key,
		paramstore: paramstore,
		authority:  authority,

		ICS4Wrapper:         ics4Wrapper,
		IBCKeeper:           ibcKeeper,
		ICAControllerKeeper: icaControllerKeeper,
		ICACallbacksKeeper:  icaCallbacksKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetAuthority returns the x/staking module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
