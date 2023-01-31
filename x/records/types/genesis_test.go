package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/x/records/types"

	sdkmath "cosmossdk.io/math"
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
				DepositRecordList: []types.DepositRecord{
					{
						Id: sdkmath.ZeroInt(),
					},
					{
						Id: sdkmath.NewInt(1),
					},
				},
				DepositRecordCount: sdkmath.NewInt(2),
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated depositRecord",
			genState: &types.GenesisState{
				DepositRecordList: []types.DepositRecord{
					{
						Id: sdkmath.ZeroInt(),
					},
					{
						Id: sdkmath.ZeroInt(),
					},
				},
			},
			valid: false,
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
