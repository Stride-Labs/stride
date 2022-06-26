package types_test

import (
	"testing"

	"github.com/Stride-Labs/stride/x/records/types"
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
				UserRedemptionRecordList: []types.UserRedemptionRecord{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				UserRedemptionRecordCount: 2,
				EpochUnbondingRecordList: []types.EpochUnbondingRecord{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				EpochUnbondingRecordCount: 2,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated userRedemptionRecord",
			genState: &types.GenesisState{
				UserRedemptionRecordList: []types.UserRedemptionRecord{
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
			desc: "invalid userRedemptionRecord count",
			genState: &types.GenesisState{
				UserRedemptionRecordList: []types.UserRedemptionRecord{
					{
						Id: 1,
					},
				},
				UserRedemptionRecordCount: 0,
			},
			valid: false,
		},
		{
			desc: "duplicated epochUnbondingRecord",
			genState: &types.GenesisState{
				EpochUnbondingRecordList: []types.EpochUnbondingRecord{
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
			desc: "invalid epochUnbondingRecord count",
			genState: &types.GenesisState{
				EpochUnbondingRecordList: []types.EpochUnbondingRecord{
					{
						Id: 1,
					},
				},
				EpochUnbondingRecordCount: 0,
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
