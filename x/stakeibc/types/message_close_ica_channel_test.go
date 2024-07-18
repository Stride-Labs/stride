package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

func TestMsgCloseICAChannel(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChannelId := "channel-10"
	validPortId := "port"

	tests := []struct {
		name string
		msg  types.MsgCloseICAChannel
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgCloseICAChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    validPortId,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgCloseICAChannel{
				Creator:   invalidAddress,
				ChannelId: validChannelId,
				PortId:    validPortId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgCloseICAChannel{
				Creator:   validNotAdminAddress,
				ChannelId: validChannelId,
				PortId:    validPortId,
			},
			err: "is not an admin",
		},
		{
			name: "invalid channel prefix",
			msg: types.MsgCloseICAChannel{
				Creator:   validAdminAddress,
				ChannelId: "chann-1",
				PortId:    validPortId,
			},
			err: "invalid channel-id",
		},
		{
			name: "invalid connection suffix",
			msg: types.MsgCloseICAChannel{
				Creator:   validAdminAddress,
				ChannelId: "channel-X",
				PortId:    validPortId,
			},
			err: "invalid channel-id",
		},
		{
			name: "invalid port ID",
			msg: types.MsgCloseICAChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    "",
			},
			err: "port ID must be specified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.ChannelId, validChannelId, "channel-id")
				require.Equal(t, test.msg.Type(), "close_ica_channel", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
