package stakeibc_test

import (
	"testing"

	keepertest "github.com/Stride-Labs/stride/testutil/keeper"
	"github.com/Stride-Labs/stride/testutil/nullify"
	"github.com/Stride-Labs/stride/x/stakeibc"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		ICAAccount: &types.ICAAccount{
			Address:            "78",
			UndelegatedBalance: 49,
			DelegatedBalance:   80,
		},
		HostZoneList: []types.HostZone{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		HostZoneCount: 2,
		DepositRecordList: []types.DepositRecord{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		DepositRecordCount: 2,
		ControllerBalancesList: []types.ControllerBalances{
			{
				Index: "0",
			},
			{
				Index: "1",
			},
		},
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.StakeibcKeeper(t)
	stakeibc.InitGenesis(ctx, *k, genesisState)
	got := stakeibc.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.Equal(t, genesisState.PortId, got.PortId)

	require.Equal(t, genesisState.ICAAccount, got.ICAAccount)
	require.ElementsMatch(t, genesisState.HostZoneList, got.HostZoneList)
	require.Equal(t, genesisState.HostZoneCount, got.HostZoneCount)
	require.ElementsMatch(t, genesisState.DepositRecordList, got.DepositRecordList)
	require.Equal(t, genesisState.DepositRecordCount, got.DepositRecordCount)
	require.ElementsMatch(t, genesisState.ControllerBalancesList, got.ControllerBalancesList)
	// this line is used by starport scaffolding # genesis/test/assert
}
