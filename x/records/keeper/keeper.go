package keeper

import (
	"fmt"

	ibckeeper "github.com/cosmos/ibc-go/v5/modules/core/keeper"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"

	icacallbackskeeper "github.com/Stride-Labs/stride/v9/x/icacallbacks/keeper"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/Stride-Labs/stride/v9/x/records/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		Cdc                codec.BinaryCodec
		storeKey           storetypes.StoreKey
		memKey             storetypes.StoreKey
		paramstore         paramtypes.Subspace
		scopedKeeper       capabilitykeeper.ScopedKeeper
		AccountKeeper      types.AccountKeeper
		TransferKeeper     ibctransferkeeper.Keeper
		IBCKeeper          ibckeeper.Keeper
		ICACallbacksKeeper icacallbackskeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	AccountKeeper types.AccountKeeper,
	TransferKeeper ibctransferkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	ICACallbacksKeeper icacallbackskeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Cdc:                Cdc,
		storeKey:           storeKey,
		memKey:             memKey,
		paramstore:         ps,
		scopedKeeper:       scopedKeeper,
		AccountKeeper:      AccountKeeper,
		TransferKeeper:     TransferKeeper,
		IBCKeeper:          ibcKeeper,
		ICACallbacksKeeper: ICACallbacksKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}
