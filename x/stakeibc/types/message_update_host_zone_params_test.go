package types_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

func TestMsgUpdateHostZoneParams(t *testing.T) {
	apptesting.SetupConfig()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	validChainId := "chain-0"

	tests := []struct {
		name string
		msg  types.MsgUpdateHostZoneParams
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgUpdateHostZoneParams{
				Authority:           authority,
				ChainId:             validChainId,
				MaxMessagesPerIcaTx: 30,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgUpdateHostZoneParams{
				Authority:           "",
				ChainId:             validChainId,
				MaxMessagesPerIcaTx: 30,
			},
			err: "invalid authority address",
		},
		{
			name: "missing chain ID",
			msg: types.MsgUpdateHostZoneParams{
				Authority:           authority,
				ChainId:             "",
				MaxMessagesPerIcaTx: 30,
			},
			err: "chain ID must be specified",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "update_host_zone_params")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), authority)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
