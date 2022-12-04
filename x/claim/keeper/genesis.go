package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/claim/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// If its the chain genesis, set the airdrop start time to be now, and setup the needed module accounts.
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}
	if err := k.SetClaimRecordsWithWeights(ctx, genState.ClaimRecords); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the capability module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}
	genesis := types.DefaultGenesis()
	genesis.Params = params
	genesis.ClaimRecords = k.GetClaimRecords(ctx, "")
	return genesis
}
