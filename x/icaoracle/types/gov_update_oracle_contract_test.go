package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestGovUpdateOracleContract(t *testing.T) {
	apptesting.SetupConfig()

	validTitle := "UpdateOracleContract"
	validDescription := "Update oracle contract"
	validMoniker := "moniker"
	validContractAddress := "juno14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9skjuwg8"

	tests := []struct {
		name     string
		proposal types.UpdateOracleContractProposal
		err      string
	}{
		{
			name: "successful proposal",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     validDescription,
				OracleMoniker:   validMoniker,
				ContractAddress: validContractAddress,
			},
		},
		{
			name: "invalid title",
			proposal: types.UpdateOracleContractProposal{
				Title:           "",
				Description:     validDescription,
				OracleMoniker:   validMoniker,
				ContractAddress: validContractAddress,
			},
			err: "title cannot be blank",
		},
		{
			name: "invalid description",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     "",
				OracleMoniker:   validMoniker,
				ContractAddress: validContractAddress,
			},
			err: "description cannot be blank",
		},
		{
			name: "empty moniker",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     validDescription,
				OracleMoniker:   "",
				ContractAddress: validContractAddress,
			},
			err: "oracle-moniker is required",
		},
		{
			name: "invalid moniker",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     validDescription,
				OracleMoniker:   "moniker 1",
				ContractAddress: validContractAddress,
			},
			err: "oracle-moniker cannot contain any spaces",
		},
		{
			name: "invalid contract address",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     validDescription,
				OracleMoniker:   validMoniker,
				ContractAddress: "",
			},
			err: "contract address is required",
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
