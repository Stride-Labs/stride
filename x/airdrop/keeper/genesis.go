package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	for _, airdrop := range genState.AirdropRecords {
		k.SetAirdropRecords(ctx, airdrop)
	}
	for _, allocation := range genState.AllocationRecords {
		k.SetAllocationRecords(ctx, allocation)
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.AirdropRecords = k.GetAirdropRecords(ctx)
	genesis.AllocationRecords = k.GetAllocationRecords(ctx)
	return genesis
}
