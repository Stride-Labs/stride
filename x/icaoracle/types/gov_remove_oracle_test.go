package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestGovRemoveOracle(t *testing.T) {
	apptesting.SetupConfig()

	validTitle := "RemoveOracle"
	validDescription := "Remove oracle"
	validChainId := "chain-id"

	tests := []struct {
		name     string
		proposal types.RemoveOracleProposal
		err      string
	}{
		{
			name: "successful proposal",
			proposal: types.RemoveOracleProposal{
				Title:         validTitle,
				Description:   validDescription,
				OracleChainId: validChainId,
			},
		},
		{
			name: "invalid title",
			proposal: types.RemoveOracleProposal{
				Title:         "",
				Description:   validDescription,
				OracleChainId: validChainId,
			},
			err: "title cannot be blank",
		},
		{
			name: "invalid description",
			proposal: types.RemoveOracleProposal{
				Title:         validTitle,
				Description:   "",
				OracleChainId: validChainId,
			},
			err: "description cannot be blank",
		},
		{
			name: "empty chain-id",
			proposal: types.RemoveOracleProposal{
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
