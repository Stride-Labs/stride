package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	"github.com/Stride-Labs/stride/v14/x/icaoracle/types"
)

func TestMsgRestoreOracleICA(t *testing.T) {
	apptesting.SetupConfig()
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()

	validChainId := "chain-1"

	tests := []struct {
		name string
		msg  types.MsgRestoreOracleICA
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgRestoreOracleICA{
				Creator:       validAddr,
				OracleChainId: validChainId,
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgRestoreOracleICA{
				Creator:       invalidAddr,
				OracleChainId: validChainId,
			},
			err: "invalid creator address",
		},
		{
			name: "empty chain-id",
			msg: types.MsgRestoreOracleICA{
				Creator:       validAddr,
				OracleChainId: "",
			},
			err: "oracle-chain-id is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)

				require.Equal(t, test.msg.OracleChainId, validChainId, "oracle-chain-id")
				require.Equal(t, test.msg.Type(), "restore_oracle_ica", "type")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
