package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		panic(err)
	}

	for _, tokenPrice := range genState.TokenPrices {
		if err := k.SetTokenPrice(ctx, tokenPrice); err != nil {
			panic(err)
		}
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	params, err := k.GetParams(ctx)
	if err != nil {
		panic(err)
	}

	genesis := types.DefaultGenesis()
	genesis.Params = params
	genesis.TokenPrices = k.GetAllTokenPrices(ctx)
	return genesis
}
