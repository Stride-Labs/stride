package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/Stride-Labs/stride/v8/x/autopilot/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
)

type (
	Keeper struct {
		Cdc            codec.BinaryCodec
		storeKey       storetypes.StoreKey
		paramstore     paramtypes.Subspace
		stakeibcKeeper stakeibckeeper.Keeper
	}
)

func NewKeeper(
	Cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	stakeibcKeeper stakeibckeeper.Keeper,
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
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
