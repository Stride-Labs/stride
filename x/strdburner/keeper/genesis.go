package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/strdburner/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// Initialize module account in account keeper if not already initialized
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)

	// Set Total STRD Burned
	k.SetTotalStrdBurned(ctx, genState.TotalUstrdBurned)
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.TotalUstrdBurned = k.GetTotalStrdBurned(ctx)
	return genesis
}
