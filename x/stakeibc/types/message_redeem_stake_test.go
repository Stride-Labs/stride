package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v5/testutil/sample"
)

func TestMsgRedeemStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   sdkmath.NewInt(1),
			},
		},
		{
			name: "invalid creator",
			msg: MsgRedeemStake{
				Creator:  "invalid_address",
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   sdkmath.NewInt(1),
			},
			err: ErrInvalidAddress,
		},
		{
			name: "no host zone",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Amount:   sdkmath.NewInt(1),
			},
			err: ErrRequiredFieldEmpty,
		},
		{
			name: "invalid receiver",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Amount:   sdkmath.NewInt(1),
			},
			err: ErrRequiredFieldEmpty,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorAs(t, err, &tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
