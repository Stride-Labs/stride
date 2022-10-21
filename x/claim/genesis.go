package claim

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/claim/keeper"
	"github.com/Stride-Labs/stride/x/claim/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// If its the chain genesis, set the airdrop start time to be now, and setup the needed module accounts.
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	if err := k.SetClaimRecords(ctx, genState.ClaimRecords); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	genesis := types.DefaultGenesis()
	genesis.Params = params
	genesis.ClaimRecords = k.GetClaimRecords(ctx, "")
	return genesis
}
