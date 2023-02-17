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
	validChainId := "chain-1"
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
				OracleChainId:   validChainId,
				ContractAddress: validContractAddress,
			},
		},
		{
			name: "invalid title",
			proposal: types.UpdateOracleContractProposal{
				Title:           "",
				Description:     validDescription,
				OracleChainId:   validChainId,
				ContractAddress: validContractAddress,
			},
			err: "title cannot be blank",
		},
		{
			name: "invalid description",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     "",
				OracleChainId:   validChainId,
				ContractAddress: validContractAddress,
			},
			err: "description cannot be blank",
		},
		{
			name: "empty chain-id",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     validDescription,
				OracleChainId:   "",
				ContractAddress: validContractAddress,
			},
			err: "oracle-chain-id is required",
		},
		{
			name: "invalid contract address",
			proposal: types.UpdateOracleContractProposal{
				Title:           validTitle,
				Description:     validDescription,
				OracleChainId:   validChainId,
				ContractAddress: "",
			},
			err: "contract address is required",
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
