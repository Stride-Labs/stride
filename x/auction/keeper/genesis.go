package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/auction/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Initialize module account in account keeper if not already initialized
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)

	for _, auction := range genState.Auctions {
		k.SetAuction(ctx, &auction)
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	genesis := types.DefaultGenesis()
	genesis.Params = params
	genesis.Auctions = k.GetAllAuctions(ctx)
	return genesis
}
