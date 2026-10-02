package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v35/app/apptesting"
	"github.com/Stride-Labs/stride/v35/x/stakeibc/types"
)

func TestMsgCalibrateDelegation_ValidateBasic(t *testing.T) {
	apptesting.SetupConfig()
	validNonAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	adminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"
	validValoper := "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p"

	tests := []struct {
		name string
		msg  types.MsgCalibrateDelegation
		err  string
	}{
		{
			name: "valid admin message",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgCalibrateDelegation{
				Creator: invalidAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
			err: "invalid creator address",
		},
		{
			name: "non-admin creator",
			msg: types.MsgCalibrateDelegation{
				Creator: validNonAdminAddress,
				ChainId: validChainId,
				Valoper: validValoper,
			},
			err: "is not an admin",
		},
		{
			name: "missing chain id",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: "",
				Valoper: validValoper,
			},
			err: "chainid is required",
		},
		{
			name: "missing valoper",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: "",
			},
			err: "valoper is required",
		},
		{
			name: "valoper without the valoper prefix",
			msg: types.MsgCalibrateDelegation{
				Creator: adminAddress,
				ChainId: validChainId,
				Valoper: "cosmos1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgl2scj",
			},
			err: "must contrain 'valoper'",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.err != "" {
				require.ErrorContains(t, err, tc.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
