package auction

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/Stride-Labs/stride/v5/x/auction/keeper"
	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
    // Set all the auctionPool
for _, elem := range genState.AuctionPoolList {
	k.SetAuctionPool(ctx, elem)
}

// Set auctionPool count
k.SetAuctionPoolCount(ctx, genState.AuctionPoolCount)
// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

    genesis.AuctionPoolList = k.GetAllAuctionPool(ctx)
genesis.AuctionPoolCount = k.GetAuctionPoolCount(ctx)
// this line is used by starport scaffolding # genesis/module/export

    return genesis
}
