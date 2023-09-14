package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/x/records/types"
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
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				DepositRecordCount: 2,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated depositRecord",
			genState: &types.GenesisState{
				DepositRecordList: []types.DepositRecord{
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
			desc: "duplicated lsm token deposit id",
			genState: &types.GenesisState{
				LsmTokenDepositList: []types.LSMTokenDeposit{
					{
						DepositId: "1",
					},
					{
						DepositId: "1",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated lsm token deposit id",
			genState: &types.GenesisState{
				LsmTokenDepositList: []types.LSMTokenDeposit{
					{
						ChainId: "chain-1",
						Denom:   "denom1",
					},
					{
						ChainId: "chain-1",
						Denom:   "denom1",
					},
				},
			},
			valid: false,
		},
		{
			desc: "valid lsm token deposits",
			genState: &types.GenesisState{
				PortId: "port-1",
				LsmTokenDepositList: []types.LSMTokenDeposit{
					{
						DepositId: "1",
						ChainId:   "chain-1",
						Denom:     "denom1",
					},
					{
						DepositId: "2",
						ChainId:   "chain-2",
						Denom:     "denom2",
					},
				},
			},
			valid: true,
		},
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
