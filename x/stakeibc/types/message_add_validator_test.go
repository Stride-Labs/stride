package types

import (
	"testing"

	"github.com/Stride-Labs/stride/v2/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgAddValidator_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgAddValidator
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgAddValidator{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid but not whitelisted address",
			msg: MsgAddValidator{
				Creator: sample.AccAddress(),
			},
			err: sdkerrors.ErrInvalidAddress,
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
