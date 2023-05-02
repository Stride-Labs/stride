package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v9/app/apptesting"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func TestMsgAddValidators_ValidateBasic(t *testing.T) {
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	valName1 := "val1"
	valName2 := "val2"
	valAddress1 := "cosmosvaloper1"
	valAddress2 := "cosmosvaloper2"

	tests := []struct {
		name string
		msg  types.MsgAddValidators
		err  string
	}{
		{
			name: "valid message",
			msg: types.MsgAddValidators{
				Creator: adminAddress,
				Validators: []*types.Validator{
					{Name: valName1, Address: valAddress1},
					{Name: valName2, Address: valAddress2},
				},
			},
		},
		{
			name: "invalid address",
			msg: types.MsgAddValidators{
				Creator: invalidAddress,
			},
			err: "invalid address",
		},
		{
			name: "non-admin address",
			msg: types.MsgAddValidators{
				Creator: validNonAdminAddress,
			},
			err: "invalid address",
		},
		{
			name: "no validators",
			msg: types.MsgAddValidators{
				Creator:    adminAddress,
				Validators: []*types.Validator{},
			},
			err: "at least one validator must be provided",
		},
		{
			name: "invalid validator name",
			msg: types.MsgAddValidators{
				Creator: adminAddress,
				Validators: []*types.Validator{
					{Name: valName1, Address: valAddress1},
					{Name: "", Address: valAddress2},
				},
			},
			err: "validator name is required (index 1)",
		},
		{
			name: "invalid validator address",
			msg: types.MsgAddValidators{
				Creator: adminAddress,
				Validators: []*types.Validator{
					{Name: valName1, Address: valAddress1},
					{Name: valName2, Address: ""},
				},
			},
			err: "validator address is required (index 1)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualError := tc.msg.ValidateBasic()
			if tc.err == "" {
				require.NoError(t, actualError)
			} else {
				require.ErrorContains(t, actualError, tc.err)
			}
		})
	}
}
