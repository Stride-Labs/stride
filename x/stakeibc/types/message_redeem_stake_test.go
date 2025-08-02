package types_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

func TestMsgRedeemStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  types.MsgRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: types.MsgRedeemStake{
				Creator:  apptesting.SampleStrideAddress(),
				HostZone: "GAIA",
				Receiver: apptesting.SampleHostAddress(),
				Amount:   sdkmath.NewInt(1),
			},
		},
		{
			name: "invalid creator",
			msg: types.MsgRedeemStake{
				Creator:  "invalid_address",
				HostZone: "GAIA",
				Receiver: apptesting.SampleHostAddress(),
				Amount:   sdkmath.NewInt(1),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "no host zone",
			msg: types.MsgRedeemStake{
				Creator:  apptesting.SampleStrideAddress(),
				Receiver: apptesting.SampleHostAddress(),
				Amount:   sdkmath.NewInt(1),
			},
			err: types.ErrRequiredFieldEmpty,
		},
		{
			name: "invalid receiver",
			msg: types.MsgRedeemStake{
				Creator:  apptesting.SampleStrideAddress(),
				HostZone: "GAIA",
				Amount:   sdkmath.NewInt(1),
			},
			err: types.ErrRequiredFieldEmpty,
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
