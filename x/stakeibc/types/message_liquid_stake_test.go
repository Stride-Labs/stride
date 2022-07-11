package types

import (
	"testing"

	"github.com/Stride-Labs/stride/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
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
				Creator: "invalid_address",
				Amount: 1,
				HostDenom: "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, 
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: MsgLiquidStake{
				Creator: "osmo1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				Amount: 1,
				HostDenom: "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, 
		{
			name: "valid inputs",
			msg: MsgLiquidStake{
				Creator: sample.AccAddress(),
				Amount: 1,
				HostDenom: "uatom",
			},
		},
		{
			name: "zero amount",
			msg: MsgLiquidStake{
				Creator: sample.AccAddress(),
				Amount: 0,
				HostDenom: "uatom",
			},
			err: ErrInvalidAmount,
		},
		{
			name: "negative amount",
			msg: MsgLiquidStake{
				Creator: sample.AccAddress(),
				Amount: -1,
				HostDenom: "uatom",
			},
			err: ErrInvalidAmount,
		},
		{
			name: "empty host denom",
			msg: MsgLiquidStake{
				Creator: sample.AccAddress(),
				Amount: 1,
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
