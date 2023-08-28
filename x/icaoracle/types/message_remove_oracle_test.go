package types_test

import (
	"testing"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func TestMsgRemoveOracle(t *testing.T) {
	apptesting.SetupConfig()

	validChainId := "chain-id"

	tests := []struct {
		name string
		msg  types.MsgRemoveOracle
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgRemoveOracle{
				Authority:     authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				OracleChainId: validChainId,
			},
		},
		{
			name: "empty chain-id",
			msg: types.MsgRemoveOracle{
				Authority:     authtypes.NewModuleAddress(govtypes.ModuleName).String(),
				OracleChainId: "",
			},
			err: "oracle-chain-id is required",
		},
		{
			name: "invalid authority",
			msg: types.MsgRemoveOracle{
				Authority:     "invalid",
				OracleChainId: validChainId,
			},
			err: "invalid authority address",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.OracleChainId, validChainId, "oracle chain-id")
				require.Equal(t, test.msg.Type(), "remove_oracle", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
