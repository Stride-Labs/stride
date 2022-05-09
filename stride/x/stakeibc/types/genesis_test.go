package types_test

import (
	"testing"

	"github.com/Stride-labs/stride/x/stakeibc/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				PortId: types.PortID,
				Validator: &types.Validator{
					Name:           "5",
					Address:        "1",
					CommissionRate: 9,
					DelegationAmt:  49,
				},
				Delegation: &types.Delegation{
					DelegateAcctAddress: "67",
					ValidatorAddr:       "17",
					Amt:                 45,
				},
				MinValidatorRequirements: &types.MinValidatorRequirements{
					CommissionRate: 49,
					Uptime:         88,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
