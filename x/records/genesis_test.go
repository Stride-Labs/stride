package records_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v14/testutil/keeper"
	"github.com/Stride-Labs/stride/v14/testutil/nullify"
	"github.com/Stride-Labs/stride/v14/x/records"
	"github.com/Stride-Labs/stride/v14/x/records/types"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		UserRedemptionRecordList: []types.UserRedemptionRecord{
			{
				Id: "0",
			},
			{
				Id: "1",
			},
		},
		UserRedemptionRecordCount: 2,
		EpochUnbondingRecordList: []types.EpochUnbondingRecord{
			{
				EpochNumber: 0,
			},
			{
				EpochNumber: 1,
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
		DepositRecordList: []types.DepositRecord{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		DepositRecordCount: 2,
		LsmTokenDepositList: []types.LSMTokenDeposit{
			{
				DepositId: "ID1",
				ChainId:   "chain-1",
				Denom:     "denom1",
			},
			{
				DepositId: "ID2",
				ChainId:   "chain-2",
				Denom:     "denom2",
			},
		},
	}
	k, ctx := keepertest.RecordsKeeper(t)
	records.InitGenesis(ctx, *k, genesisState)
	got := records.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	require.ElementsMatch(t, genesisState.DepositRecordList, got.DepositRecordList)
	require.Equal(t, genesisState.DepositRecordCount, got.DepositRecordCount)
	require.ElementsMatch(t, genesisState.LsmTokenDepositList, got.LsmTokenDepositList)
	// this line is used by starport scaffolding # genesis/test/assert
}
