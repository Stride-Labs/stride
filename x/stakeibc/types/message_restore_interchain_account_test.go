package types

import (
	fmt "fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v3/testutil/sample"
)

func TestMsgRestoreInterchainAccount_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRestoreInterchainAccount
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgRestoreInterchainAccount{
				Creator: "invalid_address",
			},
			err: fmt.Errorf("%s", &Error{errorCode: "invalid creator address (decoding bech32 failed: invalid separator index -1)"}),
		}, {
			name: "not admin address",
			msg: MsgRestoreInterchainAccount{
				Creator: sample.AccAddress(),
			},
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
