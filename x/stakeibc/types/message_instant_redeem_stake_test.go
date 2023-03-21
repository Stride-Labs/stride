package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v7/testutil/sample"
)

func TestMsgInstantRedeemStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgInstantRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: MsgInstantRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Amount:   sdkmath.NewInt(1),
			},
		},
		{
			name: "invalid creator",
			msg: MsgInstantRedeemStake{
				Creator:  "invalid_address",
				HostZone: "GAIA",
				Amount:   sdkmath.NewInt(1),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "no host zone",
			msg: MsgInstantRedeemStake{
				Creator: sample.AccAddress(),
				Amount:  sdkmath.NewInt(1),
			},
			err: ErrRequiredFieldEmpty,
		},
		{
			name: "zero amount",
			msg: MsgInstantRedeemStake{
				Creator:  sample.AccAddress(),
				Amount:   sdkmath.ZeroInt(),
				HostZone: "GAIA",
			},
			err: ErrInvalidAmount,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
