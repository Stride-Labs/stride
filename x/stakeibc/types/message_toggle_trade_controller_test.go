package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v21/app/apptesting"
	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
)

func TestMsgToggleTradeController(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validAddress := "cosmosXXX"
	validChainId := "chain-0"
	validPermissionChange := types.AuthzPermissionChange_GRANT

	tests := []struct {
		name string
		msg  types.MsgToggleTradeController
		err  string
	}{
		{
			name: "valid message",
			msg: types.MsgToggleTradeController{
				Creator:          validAdminAddress,
				ChainId:          validChainId,
				PermissionChange: validPermissionChange,
				Address:          validAddress,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgToggleTradeController{
				Creator:          invalidAddress,
				ChainId:          validChainId,
				PermissionChange: validPermissionChange,
				Address:          validAddress,
			},
			err: "invalid creator address",
		},
		{
			name: "not admin address",
			msg: types.MsgToggleTradeController{
				Creator:          validNotAdminAddress,
				ChainId:          validChainId,
				PermissionChange: validPermissionChange,
				Address:          validAddress,
			},
			err: "not an admin",
		},
		{
			name: "invalid address",
			msg: types.MsgToggleTradeController{
				Creator:          validAdminAddress,
				ChainId:          validChainId,
				PermissionChange: validPermissionChange,
				Address:          "",
			},
			err: "address must be specified",
		},
		{
			name: "invalid chain ID",
			msg: types.MsgToggleTradeController{
				Creator:          validAdminAddress,
				ChainId:          "",
				PermissionChange: validPermissionChange,
				Address:          validAddress,
			},
			err: "chain ID must be specified",
		},
		{
			name: "invalid permission change",
			msg: types.MsgToggleTradeController{
				Creator:          validAdminAddress,
				ChainId:          validChainId,
				PermissionChange: 100,
				Address:          validAddress,
			},
			err: "invalid permission change enum value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "toggle_trade_controller")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
