package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestMsgResetRateLimit(t *testing.T) {
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()
	validPathId := "denom/channel-0"

	tests := []struct {
		name       string
		msg        types.MsgResetRateLimit
		expectPass bool
	}{
		{
			name: "successful message",
			msg: types.MsgResetRateLimit{
				Creator: validAddr,
				PathId:  validPathId,
			},
			expectPass: true,
		},
		{
			name: "invalid creator",
			msg: types.MsgResetRateLimit{
				Creator: invalidAddr,
				PathId:  validPathId,
			},
			expectPass: false,
		},
		{
			name: "empty path",
			msg: types.MsgResetRateLimit{
				Creator: validAddr,
				PathId:  "",
			},
			expectPass: false,
		},
		{
			name: "invalid path",
			msg: types.MsgResetRateLimit{
				Creator: validAddr,
				PathId:  "denom_channel-0",
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "reset_rate_limit")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)

				require.Equal(t, test.msg.PathId, validPathId)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}
