package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v3/testutil/sample"
)

func TestMsgDeleteValidator_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteValidator
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeleteValidator{
				Creator: "invalid_address",
			},
			err: fmt.Errorf("%s", &Error{errorCode: "invalid creator address (decoding bech32 failed: invalid separator index -1)"}),
		}, {
			name: "valid address but not whitelisted",
			msg: MsgDeleteValidator{
				Creator: sample.AccAddress(),
			},
			err: fmt.Errorf("%s", &Error{errorCode: "invalid creator address (decoding bech32 failed: invalid separator index -1)"}),
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
