package keeper

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
)

// Loads module state from genesis
func (k Keeper) InitGenesis(ctx sdk.Context, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)
	for _, tokenPrice := range genState.TokenPrices {
		tokenPrice.SpotPrice = math.LegacyZeroDec()
		tokenPrice.UpdatedAt = time.Time{}
		tokenPrice.QueryInProgress = false

		if err := k.SetTokenPrice(ctx, tokenPrice); err != nil {
			panic(err)
		}
	}
}

// Export's module state into genesis file
func (k Keeper) ExportGenesis(ctx sdk.Context) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)
	genesis.TokenPrices = k.GetAllTokenPrices(ctx)
	return genesis
}
