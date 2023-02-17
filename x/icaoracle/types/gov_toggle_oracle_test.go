package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestGovToggleOracle(t *testing.T) {
	apptesting.SetupConfig()

	validTitle := "ToggleOracle"
	validDescription := "Toggle oracle"
	validChainId := "chain-1"

	tests := []struct {
		name     string
		proposal types.ToggleOracleProposal
		err      string
	}{
		{
			name: "successful proposal",
			proposal: types.ToggleOracleProposal{
				Title:         validTitle,
				Description:   validDescription,
				OracleChainId: validChainId,
			},
		},
		{
			name: "invalid title",
			proposal: types.ToggleOracleProposal{
				Title:         "",
				Description:   validDescription,
				OracleChainId: validChainId,
			},
			err: "title cannot be blank",
		},
		{
			name: "invalid description",
			proposal: types.ToggleOracleProposal{
				Title:         validTitle,
				Description:   "",
				OracleChainId: validChainId,
			},
			err: "description cannot be blank",
		},
		{
			name: "empty chain-id",
			proposal: types.ToggleOracleProposal{
				Title:         validTitle,
				Description:   validDescription,
				OracleChainId: "",
			},
			err: "oracle-chain-id is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.proposal.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.proposal.OracleChainId, validChainId, "oracle chain-id")
			} else {
				require.ErrorContains(t, test.proposal.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
