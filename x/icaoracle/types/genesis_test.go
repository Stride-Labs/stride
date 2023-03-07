package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestValidateGenesis(t *testing.T) {
	apptesting.SetupConfig()

	validChainId := "chain"
	validMetricKey := "key"
	validUpdateTime := uint64(1)
	validMetric := types.Metric{Key: validMetricKey, UpdateTime: validUpdateTime}

	tests := []struct {
		name         string
		genesisState types.GenesisState
		valid        bool
	}{
		{
			name: "valid genesis",
			genesisState: types.GenesisState{
				Oracles: []types.Oracle{
					{ChainId: validChainId},
				},
				QueuedMetrics: []types.Metric{
					validMetric,
				},
				PendingMetrics: []types.PendingMetricUpdate{
					{Metric: &validMetric, OracleChainId: validChainId},
				},
			},
			valid: true,
		},
		{
			name: "invalid oracle",
			genesisState: types.GenesisState{
				Oracles: []types.Oracle{
					{ChainId: ""},
				},
				QueuedMetrics: []types.Metric{
					validMetric,
				},
				PendingMetrics: []types.PendingMetricUpdate{
					{Metric: &validMetric, OracleChainId: validChainId},
				},
			},
			valid: false,
		},
		{
			name: "invalid queued metric",
			genesisState: types.GenesisState{
				Oracles: []types.Oracle{
					{ChainId: validChainId},
				},
				QueuedMetrics: []types.Metric{
					{Key: "", UpdateTime: validUpdateTime},
				},
				PendingMetrics: []types.PendingMetricUpdate{
					{Metric: &validMetric, OracleChainId: validChainId},
				},
			},
			valid: false,
		},
		{
			name: "invalid pending metric",
			genesisState: types.GenesisState{
				Oracles: []types.Oracle{
					{ChainId: validChainId},
				},
				QueuedMetrics: []types.Metric{
					validMetric,
				},
				PendingMetrics: []types.PendingMetricUpdate{
					{Metric: &validMetric, OracleChainId: ""},
				},
			},
			valid: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.valid {
				require.NoError(t, test.genesisState.Validate(), "test: %v", test.name)
			} else {
				require.ErrorContains(t, test.genesisState.Validate(), types.ErrInvalidGenesisState.Error(), "test: %v", test.name)
			}
		})
	}
}
