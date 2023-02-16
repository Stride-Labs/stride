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

	// Hard coding this initialization of an auction pool now to be SealedBid
	// Generating pools should eventually be called from the stakeibc module

	properties := types.AuctionPoolProperties{}
	properties.PoolAddress = "stride1u20df3trc2c2zdhm8qvh2hdjx9ewh00sv6eyy8"
	properties.MaxAllowedSupply = ^uint64(0) // max value from 2 complement
	properties.AllowedAlgorithms = [](types.AuctionType){
		types.AuctionType_SEALEDBID,
	}

	properties.DefaultSealedBidAuctionProperties = &types.SealedBidAuctionProperties{}
	properties.DefaultSealedBidAuctionProperties.AuctionDuration = 20
	properties.DefaultSealedBidAuctionProperties.RevealDuration = 10
	properties.DefaultSealedBidAuctionProperties.MaxAllowedBid = ^uint64(0)
	properties.DefaultSealedBidAuctionProperties.RedemptionRate = 1.0
	properties.DefaultSealedBidAuctionProperties.Collateral = 0

	k.StartNewAuctionPool(ctx, properties)

	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params, _ = k.GetParams(ctx)

	genesis.AuctionPoolList = k.GetAllAuctionPool(ctx)
	genesis.AuctionPoolCount = k.GetAuctionPoolCount(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
