package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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
				HostZoneList: []types.HostZone{
					{
						ChainId: "0",
					},
					{
						ChainId: "1",
					},
				},
				Params: types.DefaultParams(),
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
