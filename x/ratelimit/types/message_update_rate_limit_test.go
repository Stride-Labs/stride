package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/app/apptesting"
	cmdcfg "github.com/Stride-Labs/stride/v4/cmd/strided/config"
	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

func TestMsgUpdateRateLimit(t *testing.T) {
	cmdcfg.SetupConfig()
	validAddr, invalidAddr := apptesting.GenerateTestAddrs()

	validPathId := "denom/channel-0"
	validMaxPercentSend := uint64(10)
	validMaxPercentRecv := uint64(10)
	validDurationHours := uint64(60)

	tests := []struct {
		name       string
		msg        types.MsgUpdateRateLimit
		expectPass bool
	}{
		{
			name: "successful msg",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         validPathId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
			expectPass: true,
		},
		{
			name: "invalid creator",
			msg: types.MsgUpdateRateLimit{
				Creator:        invalidAddr,
				PathId:         validPathId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "empty pathId",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         "",
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid pathId",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         "denom_channel-0",
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid send percent",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         validPathId,
				MaxPercentSend: 101,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid receive percent",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         validPathId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: 101,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid send and receive percent",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         validPathId,
				MaxPercentSend: 0,
				MaxPercentRecv: 0,
				DurationHours:  validDurationHours,
			},
		},
		{
			name: "invalid duration",
			msg: types.MsgUpdateRateLimit{
				Creator:        validAddr,
				PathId:         validPathId,
				MaxPercentSend: validMaxPercentSend,
				MaxPercentRecv: validMaxPercentRecv,
				DurationHours:  0,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "add_rate_limit")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAddr)

				require.Equal(t, test.msg.PathId, validPathId)
				require.Equal(t, test.msg.MaxPercentSend, validMaxPercentSend)
				require.Equal(t, test.msg.MaxPercentRecv, validMaxPercentRecv)
				require.Equal(t, test.msg.DurationHours, validMaxPercentRecv)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}
