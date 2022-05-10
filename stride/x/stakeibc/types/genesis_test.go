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
				ICAAccount: &types.ICAAccount{
					Address:          "79",
					Balance:          2,
					DelegatedBalance: 8,
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated hostZone",
			genState: &types.GenesisState{
				HostZoneList: []types.HostZone{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid hostZone count",
			genState: &types.GenesisState{
				HostZoneList: []types.HostZone{
					{
						Id: 1,
					},
				},
				HostZoneCount: 0,
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
