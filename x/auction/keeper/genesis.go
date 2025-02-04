package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/auction/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, auction := range genState.Auctions {
		if err := k.SetAuction(ctx, &auction); err != nil {
			panic(err)
		}
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
