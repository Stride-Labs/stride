package types

import (
	"math"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v16/testutil/sample"
)

func TestMsgClaimUndelegatedTokens_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgClaimUndelegatedTokens
		err  error
	}{
		{
			name: "success",
			msg: MsgClaimUndelegatedTokens{
				Creator:    sample.AccAddress(),
				Receiver:   sample.HostAddress(),
				HostZoneId: "GAIA",
				Epoch:      uint64(1),
			},
		},
		{
			name: "invalid address",
			msg: MsgClaimUndelegatedTokens{
				Creator:    "invalid_address",
				Receiver:   sample.HostAddress(),
				HostZoneId: "GAIA",
				Epoch:      uint64(1),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "no host zone",
			msg: MsgClaimUndelegatedTokens{
				Creator:  sample.AccAddress(),
				Receiver: sample.HostAddress(),
				Epoch:    uint64(1),
			},
			err: ErrRequiredFieldEmpty,
		},
		{
			name: "epoch max int",
			msg: MsgClaimUndelegatedTokens{
				Creator:    sample.AccAddress(),
				Receiver:   sample.HostAddress(),
				HostZoneId: "GAIA",
				Epoch:      math.MaxUint64,
			},
			err: ErrInvalidAmount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
