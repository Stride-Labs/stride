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
	validMoniker := "moniker"

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
				OracleMoniker: validMoniker,
			},
		},
		{
			name: "invalid title",
			proposal: types.ToggleOracleProposal{
				Title:         "",
				Description:   validDescription,
				OracleMoniker: validMoniker,
			},
			err: "title cannot be blank",
		},
		{
			name: "invalid description",
			proposal: types.ToggleOracleProposal{
				Title:         validTitle,
				Description:   "",
				OracleMoniker: validMoniker,
			},
			err: "description cannot be blank",
		},
		{
			name: "empty moniker",
			proposal: types.ToggleOracleProposal{
				Title:         validTitle,
				Description:   validDescription,
				OracleMoniker: "",
			},
			err: "oracle-moniker is required",
		},
		{
			name: "invalid moniker",
			proposal: types.ToggleOracleProposal{
				Title:         validTitle,
				Description:   validDescription,
				OracleMoniker: "moniker 1",
			},
			err: "oracle-moniker cannot contain any spaces",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.proposal.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.proposal.OracleMoniker, validMoniker, "oracle moniker")
			} else {
				require.ErrorContains(t, test.proposal.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
