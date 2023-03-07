package icaoracle_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestGenesis(t *testing.T) {
	oracle := types.Oracle{
		ChainId: "chain",
	}
	metric := types.Metric{
		Key:        "key",
		UpdateTime: uint64(1),
	}
	pendingMetricUpdate := types.PendingMetricUpdate{
		Metric:        &metric,
		OracleChainId: "chain",
	}

	genesisState := types.GenesisState{
		Params:         types.Params{},
		Oracles:        []types.Oracle{oracle},
		QueuedMetrics:  []types.Metric{metric},
		PendingMetrics: []types.PendingMetricUpdate{pendingMetricUpdate},
	}

	s := apptesting.SetupSuitelessTestHelper()

	icaoracle.InitGenesis(s.Ctx, s.App.ICAOracleKeeper, genesisState)
	exported := icaoracle.ExportGenesis(s.Ctx, s.App.ICAOracleKeeper)

	require.Equal(t, genesisState, *exported)
}
