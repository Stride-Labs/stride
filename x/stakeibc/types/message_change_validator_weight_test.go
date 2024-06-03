package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

func TestMsgChangeValidatorWeight_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"
	validAddress := "val1"

	tests := []struct {
		name          string
		msg           types.MsgChangeValidatorWeights
		expectedError string
	}{
		{
			name: "valid message",
			msg: types.MsgChangeValidatorWeights{
				Creator:  adminAddress,
				HostZone: validChainId,
				ValidatorWeights: []*types.ValidatorWeight{
					{Address: validAddress, Weight: 1},
				},
			},
		},
		{
			name: "invalid address",
			msg: types.MsgChangeValidatorWeights{
				Creator:  invalidAddress,
				HostZone: validChainId,
				ValidatorWeights: []*types.ValidatorWeight{
					{Address: validAddress, Weight: 1},
				},
			},
			expectedError: "invalid creator address",
		},
		{
			name: "non-admin address",
			msg: types.MsgChangeValidatorWeights{
				Creator:  validNonAdminAddress,
				HostZone: validChainId,
				ValidatorWeights: []*types.ValidatorWeight{
					{Address: validAddress, Weight: 1},
				},
			},
			expectedError: "is not an admin",
		},
		{
			name: "missing chain id",
			msg: types.MsgChangeValidatorWeights{
				Creator:  adminAddress,
				HostZone: "",
				ValidatorWeights: []*types.ValidatorWeight{
					{Address: validAddress, Weight: 1},
				},
			},
			expectedError: "host zone must be specified",
		},
		{
			name: "no validators",
			msg: types.MsgChangeValidatorWeights{
				Creator:          adminAddress,
				HostZone:         validChainId,
				ValidatorWeights: []*types.ValidatorWeight{},
			},
			expectedError: "at least one validator must be specified",
		},
		{
			name: "missing validator address",
			msg: types.MsgChangeValidatorWeights{
				Creator:  adminAddress,
				HostZone: validChainId,
				ValidatorWeights: []*types.ValidatorWeight{
					{Address: "", Weight: 1},
				},
			},
			expectedError: "validator address must be specified",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.expectedError != "" {
				require.ErrorContains(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}
