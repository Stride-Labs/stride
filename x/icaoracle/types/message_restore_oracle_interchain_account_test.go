package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v5/app/apptesting"
	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

func TestMsgRestoreOracleICA(t *testing.T) {
	apptesting.SetupConfig()
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()

	validMoniker := "moniker1"

	tests := []struct {
		name string
		msg  types.MsgRestoreOracleICA
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgRestoreOracleICA{
				Creator:       validAddr,
				OracleMoniker: validMoniker,
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgRestoreOracleICA{
				Creator:       invalidAddr,
				OracleMoniker: validMoniker,
			},
			err: "invalid creator address",
		},
		{
			name: "empty moniker",
			msg: types.MsgRestoreOracleICA{
				Creator:       validAddr,
				OracleMoniker: "",
			},
			err: "oracle-moniker is required",
		},
		{
			name: "invalid moniker",
			msg: types.MsgRestoreOracleICA{
				Creator:       validAddr,
				OracleMoniker: "moniker 1",
			},
			err: "oracle-moniker cannot contain any spaces",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "restore_oracle_ica")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)

				require.Equal(t, test.msg.OracleMoniker, validMoniker, "oracle-moniker")
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
