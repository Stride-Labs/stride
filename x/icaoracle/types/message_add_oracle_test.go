package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestMsgAddOracle(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validConnectionId := "connection-10"
	validCodeId := uint64(1)

	tests := []struct {
		name string
		msg  types.MsgAddOracle
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgAddOracle{
				Creator:        validAdminAddress,
				ConnectionId:   validConnectionId,
				ContractCodeId: validCodeId,
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgAddOracle{
				Creator:        validNotAdminAddress,
				ConnectionId:   validConnectionId,
				ContractCodeId: validCodeId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgAddOracle{
				Creator:        invalidAddress,
				ConnectionId:   validConnectionId,
				ContractCodeId: validCodeId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid connection prefix",
			msg: types.MsgAddOracle{
				Creator:        validAdminAddress,
				ConnectionId:   "connect-1",
				ContractCodeId: validCodeId,
			},
			err: "invalid connection-id",
		},
		{
			name: "invalid connection suffix",
			msg: types.MsgAddOracle{
				Creator:        validAdminAddress,
				ConnectionId:   "connection-X",
				ContractCodeId: validCodeId,
			},
			err: "invalid connection-id",
		},
		{
			name: "invalid code ID",
			msg: types.MsgAddOracle{
				Creator:        validAdminAddress,
				ConnectionId:   validConnectionId,
				ContractCodeId: 0,
			},
			err: "contract code-id cannot be 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "add_oracle")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.ConnectionId, validConnectionId, "connnectionId")
				require.Equal(t, test.msg.ContractCodeId, validCodeId, "codeId")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
