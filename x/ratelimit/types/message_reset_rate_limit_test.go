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

	validDenom := "denom"
	validChannelId := "channel-0"

	tests := []struct {
		name string
		msg  types.MsgResetRateLimit
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgResetRateLimit{
				Creator:   validAddr,
				Denom:     validDenom,
				ChannelId: validChannelId,
			},
			err: "",
		},
		{
			name: "invalid creator",
			msg: types.MsgResetRateLimit{
				Creator:   invalidAddr,
				Denom:     validDenom,
				ChannelId: validChannelId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid denom",
			msg: types.MsgResetRateLimit{
				Creator:   validAddr,
				Denom:     "",
				ChannelId: validChannelId,
			},
			err: "invalid denom",
		},
		{
			name: "invalid channel-id",
			msg: types.MsgResetRateLimit{
				Creator:   validAddr,
				Denom:     validDenom,
				ChannelId: "chan-1",
			},
			err: "invalid channel-id",
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

				require.Equal(t, test.msg.Denom, validDenom, "denom")
				require.Equal(t, test.msg.ChannelId, validChannelId, "channelId")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
