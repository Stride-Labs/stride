package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func TestMsgInstantiateOracle(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-1"
	validCodeId := uint64(1)
	validChannelId := "channel-0"

	tests := []struct {
		name string
		msg  types.MsgInstantiateOracle
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgInstantiateOracle{
				Creator:                 validAdminAddress,
				OracleChainId:           validChainId,
				ContractCodeId:          validCodeId,
				TransferChannelOnOracle: validChannelId,
			},
		},
		{
			name: "invalid creator address",
			msg: types.MsgInstantiateOracle{
				Creator:                 invalidAddress,
				OracleChainId:           validChainId,
				ContractCodeId:          validCodeId,
				TransferChannelOnOracle: validChannelId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid admin address",
			msg: types.MsgInstantiateOracle{
				Creator:                 validNotAdminAddress,
				OracleChainId:           validChainId,
				ContractCodeId:          validCodeId,
				TransferChannelOnOracle: validChannelId,
			},
			err: "invalid creator address",
		},
		{
			name: "invalid chain-id",
			msg: types.MsgInstantiateOracle{
				Creator:                 validAdminAddress,
				OracleChainId:           "",
				ContractCodeId:          validCodeId,
				TransferChannelOnOracle: validChannelId,
			},
			err: "oracle-chain-id is required",
		},
		{
			name: "invalid code ID",
			msg: types.MsgInstantiateOracle{
				Creator:                 validAdminAddress,
				OracleChainId:           validChainId,
				ContractCodeId:          0,
				TransferChannelOnOracle: validChannelId,
			},
			err: "contract code-id cannot be 0",
		},
		{
			name: "invalid channel ID",
			msg: types.MsgInstantiateOracle{
				Creator:                 validAdminAddress,
				OracleChainId:           validChainId,
				ContractCodeId:          validCodeId,
				TransferChannelOnOracle: "chan-0",
			},
			err: "invalid channel-id (chan-0)",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)

				require.Equal(t, test.msg.OracleChainId, validChainId, "chainId")
				require.Equal(t, test.msg.ContractCodeId, validCodeId, "codeId")
				require.Equal(t, test.msg.Type(), "instantiate_oracle", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
