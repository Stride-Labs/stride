package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestMsgResetRateLimit(t *testing.T) {
	apptesting.SetupConfig()
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()
	validPathId := "denom/channel-0"

	tests := []struct {
		name string
		msg  types.MsgResetRateLimit
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgResetRateLimit{
				Creator: validAddr,
				PathId:  validPathId,
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgResetRateLimit{
				Creator: invalidAddr,
				PathId:  validPathId,
			},
			err: "invalid creator address",
		},
		{
			name: "empty path",
			msg: types.MsgResetRateLimit{
				Creator: validAddr,
				PathId:  "",
			},
			err: "empty pathId",
		},
		{
			name: "invalid path",
			msg: types.MsgResetRateLimit{
				Creator: validAddr,
				PathId:  "denom_channel-0",
			},
			err: "invalid pathId",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "reset_rate_limit")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)

				require.Equal(t, test.msg.PathId, validPathId)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
