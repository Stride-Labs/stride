package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func TestMsgAddOracle(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validConnectionId := "connection-10"

	tests := []struct {
		name string
		msg  types.MsgAddOracle
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgAddOracle{
				Creator:      validAdminAddress,
				ConnectionId: validConnectionId,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgAddOracle{
				Creator:      invalidAddress,
				ConnectionId: validConnectionId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgAddOracle{
				Creator:      validNotAdminAddress,
				ConnectionId: validConnectionId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid connection prefix",
			msg: types.MsgAddOracle{
				Creator:      validAdminAddress,
				ConnectionId: "connect-1",
			},
			err: "invalid connection-id",
		},
		{
			name: "invalid connection suffix",
			msg: types.MsgAddOracle{
				Creator:      validAdminAddress,
				ConnectionId: "connection-X",
			},
			err: "invalid connection-id",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.ConnectionId, validConnectionId, "connnectionId")
				require.Equal(t, test.msg.Type(), "add_oracle", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
