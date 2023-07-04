package types_test

import (
	"testing"
	"time"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v11/x/icaoracle/types"
)

func TestValidateICATx(t *testing.T) {
	validConnectionId := "connection-0"
	validChannelId := "channel-0"
	validPortId := "port-0"
	validMessages := []proto.Message{&banktypes.MsgSend{}}
	validTimeout := time.Second
	validCallbackId := "callback-id"

	tests := []struct {
		name string
		tx   types.ICATx
		err  string
	}{
		{
			name: "successful ICA Tx",
			tx: types.ICATx{
				ConnectionId:    validConnectionId,
				ChannelId:       validChannelId,
				PortId:          validPortId,
				Messages:        validMessages,
				RelativeTimeout: validTimeout,
				CallbackId:      validCallbackId,
			},
		},
		{
			name: "invalid connection-id",
			tx: types.ICATx{
				ConnectionId:    "",
				ChannelId:       validChannelId,
				PortId:          validPortId,
				Messages:        validMessages,
				RelativeTimeout: validTimeout,
				CallbackId:      validCallbackId,
			},
			err: "connection-id is empty",
		},
		{
			name: "invalid channel-id",
			tx: types.ICATx{
				ConnectionId:    validConnectionId,
				ChannelId:       "",
				PortId:          validPortId,
				Messages:        validMessages,
				RelativeTimeout: validTimeout,
				CallbackId:      validCallbackId,
			},
			err: "channel-id is empty",
		},
		{
			name: "invalid port-id",
			tx: types.ICATx{
				ConnectionId:    validConnectionId,
				ChannelId:       validChannelId,
				PortId:          "",
				Messages:        validMessages,
				RelativeTimeout: validTimeout,
				CallbackId:      validCallbackId,
			},
			err: "port-id is empty",
		},
		{
			name: "invalid messages",
			tx: types.ICATx{
				ConnectionId:    validConnectionId,
				ChannelId:       validChannelId,
				PortId:          validPortId,
				Messages:        []proto.Message{},
				RelativeTimeout: validTimeout,
				CallbackId:      validCallbackId,
			},
			err: "messages are empty",
		},
		{
			name: "invalid timeout",
			tx: types.ICATx{
				ConnectionId:    validConnectionId,
				ChannelId:       validChannelId,
				PortId:          validPortId,
				Messages:        validMessages,
				RelativeTimeout: -1 * time.Second,
				CallbackId:      validCallbackId,
			},
			err: "timeout is not in the future",
		},
		{
			name: "invalid callback-id",
			tx: types.ICATx{
				ConnectionId:    validConnectionId,
				ChannelId:       validChannelId,
				PortId:          validPortId,
				Messages:        validMessages,
				RelativeTimeout: validTimeout,
				CallbackId:      "",
			},
			err: "callback-id is empty",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.tx.ValidateICATx(), "test: %v", test.name)
			} else {
				require.ErrorContains(t, test.tx.ValidateICATx(), test.err, "test: %v", test.name)
			}
		})
	}
}
