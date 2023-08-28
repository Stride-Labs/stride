package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func TestValidateGenesis(t *testing.T) {
	apptesting.SetupConfig()

	validChainId := "chain"
	validMetricKey := "key"
	validUpdateTime := int64(1)
	validMetric := types.Metric{
		Key:               validMetricKey,
		UpdateTime:        validUpdateTime,
		DestinationOracle: validChainId,
	}

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
				Metrics: []types.Metric{
					validMetric,
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
				Metrics: []types.Metric{
					validMetric,
				},
			},
			valid: false,
		},
		{
			name: "invalid metric",
			genesisState: types.GenesisState{
				Oracles: []types.Oracle{
					{ChainId: validChainId},
				},
				Metrics: []types.Metric{
					{Key: "", UpdateTime: validUpdateTime, DestinationOracle: validChainId},
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
