package types_test

import (
	"testing"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
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
					Address:            "79",
					UndelegatedBalance: 2,
					DelegatedBalance:   8,
				},
				HostZoneList: []types.HostZone{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				HostZoneCount: 2,
			genState: &types.GenesisState{
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
