package auction_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/v5/testutil/keeper"
	"github.com/Stride-Labs/stride/v5/testutil/nullify"
	"github.com/Stride-Labs/stride/v5/x/auction"
	"github.com/Stride-Labs/stride/v5/x/auction/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params:	types.DefaultParams(),
		
		AuctionPoolList: []types.AuctionPool{
		{
			Id: 0,
		},
		{
			Id: 1,
		},
	},
	AuctionPoolCount: 2,
	// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.AuctionKeeper(t)
	auction.InitGenesis(ctx, *k, genesisState)
	got := auction.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	

	require.ElementsMatch(t, genesisState.AuctionPoolList, got.AuctionPoolList)
require.Equal(t, genesisState.AuctionPoolCount, got.AuctionPoolCount)
// this line is used by starport scaffolding # genesis/test/assert
}
