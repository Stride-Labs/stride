package stakeibc_test

import (
	"testing"

	keepertest "github.com/Stride-labs/stride/testutil/keeper"
	"github.com/Stride-labs/stride/testutil/nullify"
	"github.com/Stride-labs/stride/x/stakeibc"
	"github.com/Stride-labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		PortId: types.PortID,
		Validator: &types.Validator{
			Name:           "80",
			Address:        "96",
			CommissionRate: 65,
			DelegationAmt:  27,
		},
		Delegation: &types.Delegation{
			DelegateAcctAddress: "1",
			ValidatorAddr:       "66",
			Amt:                 14,
		},
		MinValidatorRequirements: &types.MinValidatorRequirements{
			CommissionRate: 46,
			Uptime:         71,
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

	require.Equal(t, genesisState.Validator, got.Validator)
	require.Equal(t, genesisState.Delegation, got.Delegation)
	require.Equal(t, genesisState.MinValidatorRequirements, got.MinValidatorRequirements)
	// this line is used by starport scaffolding # genesis/test/assert
}
