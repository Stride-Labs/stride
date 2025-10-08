package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v29/x/icqoracle/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, tokenPrice := range genState.TokenPrices {
		k.SetTokenPrice(ctx, tokenPrice)
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params := k.GetParams(ctx)
	genesis := types.DefaultGenesis()
	genesis.Params = params
	genesis.TokenPrices = k.GetAllTokenPrices(ctx)
	return genesis
}
