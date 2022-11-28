package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v3/testutil/sample"
)

func TestMsgRebalanceValidators_ValidateBasic(t *testing.T) {
	creator := sample.AccAddress()
	errorString := fmt.Sprintf("invalid creator address (%s): invalid address", creator)

	tests := []struct {
		name string
		msg  MsgRebalanceValidators
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgRebalanceValidators{
				Creator: "invalid_address",
			},
			err: fmt.Errorf("%s", &Error{errorCode: "invalid creator address (decoding bech32 failed: invalid separator index -1)"}),
		}, {
			name: "valid address but not whitelisted",
			msg: MsgRebalanceValidators{
				Creator: creator,
			},
			err: fmt.Errorf("%s", &Error{errorCode: errorString}),
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
