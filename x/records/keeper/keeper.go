package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/Stride-Labs/stride/x/records/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		Cdc          codec.BinaryCodec
		storeKey     sdk.StoreKey
		memKey       sdk.StoreKey
		paramstore   paramtypes.Subspace
		scopedKeeper capabilitykeeper.ScopedKeeper
		AccountKeeper types.AccountKeeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	scopedKeeper capabilitykeeper.ScopedKeeper,
	AccountKeeper types.AccountKeeper,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Cdc:          Cdc,
		storeKey:     storeKey,
		memKey:       memKey,
		paramstore:   ps,
		scopedKeeper: scopedKeeper,
		AccountKeeper: AccountKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}