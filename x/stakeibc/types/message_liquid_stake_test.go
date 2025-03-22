package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

func TestMsgLiquidStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgLiquidStake
		err  error
	}{
		{
			name: "invalid address",
			msg: types.MsgLiquidStake{
				Creator:   "invalid_address",
				Amount:    sdkmath.NewInt(1),
				HostDenom: "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: types.MsgLiquidStake{
				Creator:   "osmo1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				Amount:    sdkmath.NewInt(1),
				HostDenom: "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid inputs",
			msg: types.MsgLiquidStake{
				Creator:   apptesting.SampleStrideAddress(),
				Amount:    sdkmath.NewInt(1),
				HostDenom: "uatom",
			},
		},
		{
			name: "zero amount",
			msg: types.MsgLiquidStake{
				Creator:   apptesting.SampleStrideAddress(),
				Amount:    sdkmath.ZeroInt(),
				HostDenom: "uatom",
			},
			err: types.ErrInvalidAmount,
		},
		{
			name: "empty host denom",
			msg: types.MsgLiquidStake{
				Creator:   apptesting.SampleStrideAddress(),
				Amount:    sdkmath.NewInt(1),
				HostDenom: "",
			},
			err: types.ErrRequiredFieldEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// check validatebasic()
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)

			// check msg_server
		})
	}
}

func TestMsgLiquidStake_GetSignBytes(t *testing.T) {
	addr := "cosmos1v9jxgu33kfsgr5"
	msg := types.NewMsgLiquidStake(addr, sdkmath.NewInt(1000), "ustrd")
	res := msg.GetSignBytes()

	expected := `{"type":"stakeibc/MsgLiquidStake","value":{"amount":"1000","creator":"cosmos1v9jxgu33kfsgr5","host_denom":"ustrd"}}`
	require.Equal(t, expected, string(res))
}
