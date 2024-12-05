package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
)

// Initializes the genesis state in the store
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	// TODO: InitGenesis
}

// Exports the current state
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	// TODO: InitGenesis
	return genesis
}
