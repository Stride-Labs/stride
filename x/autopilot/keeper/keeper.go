package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/Stride-Labs/stride/v9/x/autopilot/types"
	claimkeeper "github.com/Stride-Labs/stride/v9/x/claim/keeper"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
)

type (
	Keeper struct {
		Cdc            codec.BinaryCodec
		storeKey       storetypes.StoreKey
		paramstore     paramtypes.Subspace
		stakeibcKeeper stakeibckeeper.Keeper
		claimKeeper    claimkeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	stakeibcKeeper stakeibckeeper.Keeper,
	claimKeeper claimkeeper.Keeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		Cdc:            Cdc,
		storeKey:       storeKey,
		paramstore:     ps,
		stakeibcKeeper: stakeibcKeeper,
		claimKeeper:    claimKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
