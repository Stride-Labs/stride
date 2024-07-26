package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v23/x/airdrop/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	for _, airdrop := range genState.Airdrops {
		k.SetAirdrop(ctx, airdrop)
	}
	for _, allocation := range genState.UserAllocations {
		k.SetUserAllocation(ctx, allocation)
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Airdrops = k.GetAllAirdrops(ctx)
	genesis.UserAllocations = k.GetAllUserAllocations(ctx)
	genesis.Params = k.GetParams(ctx)
	return genesis
}
