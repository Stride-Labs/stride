package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

func TestMsgCloseDelegationChannel(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChannelId := "channel-10"
	validPortId := "icacontroller-DELEGATION"

	tests := []struct {
		name string
		msg  types.MsgCloseDelegationChannel
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    validPortId,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgCloseDelegationChannel{
				Creator:   invalidAddress,
				ChannelId: validChannelId,
				PortId:    validPortId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validNotAdminAddress,
				ChannelId: validChannelId,
				PortId:    validPortId,
			},
			err: "is not an admin",
		},
		{
			name: "invalid channel prefix",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: "chann-1",
				PortId:    validPortId,
			},
			err: "invalid channel-id",
		},
		{
			name: "invalid connection suffix",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: "channel-X",
				PortId:    validPortId,
			},
			err: "invalid channel-id",
		},
		{
			name: "invalid port ID",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    "",
			},
			err: "port ID must be specified",
		},
		{
			name: "invalid port ID",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    "",
			},
			err: "port ID must be specified",
		},
		{
			name: "not ICA channel",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    "DELEGATION",
			},
			err: "must be an ICA channel",
		},
		{
			name: "not delegation channel",
			msg: types.MsgCloseDelegationChannel{
				Creator:   validAdminAddress,
				ChannelId: validChannelId,
				PortId:    "icacontroller-WITHDRAWAL",
			},
			err: "must be the delegation ICA channel",
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
				require.Equal(t, test.msg.Type(), "close_delegation_channel", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
