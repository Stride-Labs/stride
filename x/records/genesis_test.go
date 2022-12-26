package records_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/Stride-Labs/stride/v4/testutil/keeper"
	"github.com/Stride-Labs/stride/v4/testutil/nullify"
	"github.com/Stride-Labs/stride/v4/x/records"
	"github.com/Stride-Labs/stride/v4/x/records/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
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
		UserRedemptionRecordCount: sdk.NewInt(2),
		EpochUnbondingRecordList: []types.EpochUnbondingRecord{
			{
				EpochNumber: sdk.ZeroInt(),
			},
			{
				EpochNumber: sdk.NewInt(1),
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
		DepositRecordList: []types.DepositRecord{
			{
				Id: sdk.ZeroInt(),
			},
			{
				Id: sdk.NewInt(1),
			},
		},
		DepositRecordCount: sdk.NewInt(2),
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
	// this line is used by starport scaffolding # genesis/test/assert
}
