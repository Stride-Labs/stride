package ratelimit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/ratelimit/keeper"
	"github.com/Stride-Labs/stride/v5/x/ratelimit/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	for _, rateLimit := range genState.RateLimits {
		k.SetRateLimit(ctx, rateLimit)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	rateLimits := k.GetAllRateLimits(ctx)

	genesis.Params = k.GetParams(ctx)
	genesis.RateLimits = rateLimits

	return genesis
}
