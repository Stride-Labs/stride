package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v29/app/apptesting"
	"github.com/Stride-Labs/stride/v29/x/stakeibc/types"
)

func TestMsgCloseDelegationChannel(t *testing.T) {
	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"

	tests := []struct {
		name string
		msg  types.MsgCloseDelegationChannel
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgCloseDelegationChannel{
				Creator: validAdminAddress,
				ChainId: validChainId,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgCloseDelegationChannel{
				Creator: invalidAddress,
				ChainId: validChainId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgCloseDelegationChannel{
				Creator: validNotAdminAddress,
				ChainId: validChainId,
			},
			err: "is not an admin",
		},
		{
			name: "invalid chain-id",
			msg: types.MsgCloseDelegationChannel{
				Creator: validAdminAddress,
				ChainId: "",
			},
			err: "chain ID must be specified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)
				require.Equal(t, test.msg.Type(), "close_delegation_channel", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
