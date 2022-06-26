package records_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/records"
	"github.com/Stride-Labs/stride/x/records/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		UserRedemptionRecordList: []types.UserRedemptionRecord{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		UserRedemptionRecordCount: 2,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.RecordsKeeper(t)
	records.InitGenesis(ctx, *k, genesisState)
	got := records.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	require.ElementsMatch(t, genesisState.UserRedemptionRecordList, got.UserRedemptionRecordList)
	require.Equal(t, genesisState.UserRedemptionRecordCount, got.UserRedemptionRecordCount)
	// this line is used by starport scaffolding # genesis/test/assert
}
