package interchainquery

import (
	"github.com/Stride-Labs/stride/v10/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/v10/x/interchainquery/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// set registered zones info from genesis
	for _, query := range genState.Queries {
		// Initialize empty epoch values via Cosmos SDK
		k.SetQuery(ctx, query)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	return &types.GenesisState{
		Queries: k.AllQueries(ctx),
	}
}
