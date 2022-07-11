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
					Address: "79",
				},
				HostZoneList: []types.HostZone{
					{
						ChainId: "0",
					},
					{
						ChainId: "1",
					},
				},
				HostZoneCount: 2,
			},
			valid: true,
		},
		{
	desc:     "duplicated pendingClaims",
	genState: &types.GenesisState{
		PendingClaimsList: []types.PendingClaims{
			{
				Sequence: "0",
},
			{
				Sequence: "0",
},
		},
	},
	valid:    false,
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
