package epochs

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/epochs/keeper"
	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// set epoch info from genesis
	for _, epoch := range genState.Epochs {
		// Initialize empty epoch values via Cosmos SDK
		if epoch.StartTime.Equal(time.Time{}) {
			epoch.StartTime = ctx.BlockTime()
		}

		if epoch.CurrentEpoch < 0 {
			panic("current epoch should NOT be negative")
		}

		if epoch.CurrentEpochStartHeight < 0 {
			panic("current epoch start height should NOT be negative")
		}

		epoch.CurrentEpochStartHeight = ctx.BlockHeight()

		k.SetEpochInfo(ctx, epoch)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Epochs: k.AllEpochInfos(ctx),
	}
}
