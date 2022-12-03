package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/testutil/sample"
)

func TestMsgLiquidStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgLiquidStake
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgLiquidStake{
				Creator:   "invalid_address",
				Amount:    1,
				HostDenom: "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: MsgLiquidStake{
				Creator:   "osmo1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				Amount:    1,
				HostDenom: "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid inputs",
			msg: MsgLiquidStake{
				Creator:   sample.AccAddress(),
				Amount:    1,
				HostDenom: "uatom",
			},
		},
		{
			name: "zero amount",
			msg: MsgLiquidStake{
				Creator:   sample.AccAddress(),
				Amount:    0,
				HostDenom: "uatom",
			},
			err: ErrInvalidAmount,
		},
		{
			name: "empty host denom",
			msg: MsgLiquidStake{
				Creator:   sample.AccAddress(),
				Amount:    1,
				HostDenom: "",
			},
			err: ErrRequiredFieldEmpty,
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
	msg := NewMsgLiquidStake(addr, 1000, "ustrd")
	res := msg.GetSignBytes()

	expected := `{"type":"stakeibc/LiquidStake","value":{"amount":"1000","creator":"cosmos1v9jxgu33kfsgr5","host_denom":"ustrd"}}`
	require.Equal(t, expected, string(res))
}
