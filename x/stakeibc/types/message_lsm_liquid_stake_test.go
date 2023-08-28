package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func TestMsgLSMLiquidStake(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()

	tests := []struct {
		name string
		msg  types.MsgLSMLiquidStake
		err  string
	}{
		{
			name: "invalid address",
			msg: types.MsgLSMLiquidStake{
				Creator:          invalidAddress,
				Amount:           sdkmath.NewInt(1),
				LsmTokenIbcDenom: "validator0032vj2y9sea9d9jfstpxn",
			},
			err: "invalid creator address",
		},
		{
			name: "valid inputs",
			msg: types.MsgLSMLiquidStake{
				Creator:          validNotAdminAddress,
				Amount:           sdkmath.NewInt(1),
				LsmTokenIbcDenom: "validator0032vj2y9sea9d9jfstpxn",
			},
		},
		{
			name: "zero amount",
			msg: types.MsgLSMLiquidStake{
				Creator:          validNotAdminAddress,
				Amount:           sdkmath.ZeroInt(),
				LsmTokenIbcDenom: "validator0032vj2y9sea9d9jfstpxn",
			},
			err: "invalid amount",
		},
		{
			name: "empty lsm token denom",
			msg: types.MsgLSMLiquidStake{
				Creator:          validNotAdminAddress,
				Amount:           sdkmath.NewInt(1),
				LsmTokenIbcDenom: "",
			},
			err: "LSM token denom cannot be empty",
		},
		{
			name: "bad format lsm token denom",
			msg: types.MsgLSMLiquidStake{
				Creator:          validNotAdminAddress,
				Amount:           sdkmath.NewInt(1),
				LsmTokenIbcDenom: "38",
			},
			err: "invalid LSM token denom",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "lsm_liquid_stake")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validNotAdminAddress)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
