package icacallbacks

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/icacallbacks/keeper"
	"github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the callbackData
	for _, elem := range genState.CallbackDataList {
		k.SetCallbackData(ctx, elem)
	}
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.CallbackDataList = k.GetAllCallbackData(ctx)

	return genesis
}
