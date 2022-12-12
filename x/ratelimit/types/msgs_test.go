package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	cmdcfg "github.com/Stride-Labs/stride/v4/cmd/strided/config"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestMsgAddQuota(t *testing.T) {
	cmdcfg.SetupConfig()
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()
	tests := []struct {
		name       string
		msg        types.MsgAddQuota
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgAddQuota{
				Creator:         validAddr,
				Name:            "quota",
				MaxPercentSend:  10,
				MaxPercentRecv:  10,
				DurationMinutes: 60,
			},
			expectPass: true,
		},
		{
			name: "invalid creator",
			msg: types.MsgAddQuota{
				Creator:         invalidAddr,
				Name:            "quota",
				MaxPercentSend:  10,
				MaxPercentRecv:  10,
				DurationMinutes: 60,
			},
		},
		{
			name: "invalid name",
			msg: types.MsgAddQuota{
				Creator:         validAddr,
				Name:            "",
				MaxPercentSend:  10,
				MaxPercentRecv:  10,
				DurationMinutes: 60,
			},
		},
		{
			name: "invalid percent",
			msg: types.MsgAddQuota{
				Creator:         validAddr,
				Name:            "quota",
				MaxPercentSend:  101,
				MaxPercentRecv:  101,
				DurationMinutes: 60,
			},
		},
		{
			name: "invalid duration",
			msg: types.MsgAddQuota{
				Creator:         validAddr,
				Name:            "quota",
				MaxPercentSend:  10,
				MaxPercentRecv:  10,
				DurationMinutes: 0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "add_quota")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)
				require.Equal(t, test.msg.Name, "quota")
				require.Equal(t, test.msg.MaxPercentSend, uint64(10))
				require.Equal(t, test.msg.MaxPercentRecv, uint64(10))
				require.Equal(t, test.msg.DurationMinutes, uint64(60))
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

func TestMsgRemoveQuota(t *testing.T) {
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()
	tests := []struct {
		name       string
		msg        types.MsgRemoveQuota
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgRemoveQuota{
				Creator: validAddr,
				Name:    "quota",
			},
			expectPass: true,
		},
		{
			name: "invalid creator",
			msg: types.MsgRemoveQuota{
				Creator: invalidAddr,
				Name:    "quota",
			},
		},
		{
			name: "invalid name",
			msg: types.MsgRemoveQuota{
				Creator: validAddr,
				Name:    "",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "remove_quota")
				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)
				require.Equal(t, test.msg.Name, "quota")
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}
