package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestMsgInstantiateOracle(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-1"
	validCodeId := uint64(1)

	tests := []struct {
		name string
		msg  types.MsgInstantiateOracle
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgInstantiateOracle{
				Creator:        validAdminAddress,
				OracleChainId:  validChainId,
				ContractCodeId: validCodeId,
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgInstantiateOracle{
				Creator:        validNotAdminAddress,
				OracleChainId:  validChainId,
				ContractCodeId: validCodeId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgInstantiateOracle{
				Creator:        invalidAddress,
				OracleChainId:  validChainId,
				ContractCodeId: validCodeId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid chain-id",
			msg: types.MsgInstantiateOracle{
				Creator:        validAdminAddress,
				OracleChainId:  "",
				ContractCodeId: validCodeId,
			},
			err: "oracle-chain-id is required",
		},
		{
			name: "invalid code ID",
			msg: types.MsgInstantiateOracle{
				Creator:        validAdminAddress,
				OracleChainId:  validChainId,
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
				require.Equal(t, test.msg.Type(), "instantiate_oracle")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.OracleChainId, validChainId, "chainId")
				require.Equal(t, test.msg.ContractCodeId, validCodeId, "codeId")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
