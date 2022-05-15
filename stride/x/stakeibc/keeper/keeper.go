package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/Stride-labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
		icaControllerKeeper icacontrollerkeeper.Keeper
		ibcKeeper           ibckeeper.Keeper
		scopedKeeper        capabilitykeeper.ScopedKeeper

		bankKeeper types.BankKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	// channelKeeper cosmosibckeeper.ChannelKeeper,
	// portKeeper cosmosibckeeper.PortKeeper,
	// scopedKeeper cosmosibckeeper.ScopedKeeper,
	bankKeeper types.BankKeeper,
	icaControllerKeeper icacontrollerkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		// Scaffolding an ibc module using ignite creates a cosmosibckeeper.NewKeeper for the module,
		// but this is not compatible with ibc-v3
		// Keeper: cosmosibckeeper.NewKeeper(
		// 	types.PortKey,
		// 	storeKey,
		// 	channelKeeper,
		// 	portKeeper,
		// 	scopedKeeper,
		// ),
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		bankKeeper: bankKeeper,
		icaControllerKeeper: icaControllerKeeper,
		ibcKeeper:           ibcKeeper,
		scopedKeeper:        scopedKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

