package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/testutil/sample"
)

func TestMsgChangeValidatorWeight_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgChangeValidatorWeight
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgChangeValidatorWeight{
				Creator: "invalid_address",
			},
			err: ErrInvalidAddress,
		}, {
			name: "valid address but not whitelisted",
			msg: MsgChangeValidatorWeight{
				Creator: sample.AccAddress(),
			},
			err: ErrInvalidAddress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorAs(t, err, &tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
