package types

import (
	"math"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v4/testutil/sample"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
				Amount:   sdk.NewInt(1),
			},
		},
		{
			name: "invalid creator",
			msg: MsgRedeemStake{
				Creator:  "invalid_address",
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   sdk.NewInt(1),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "no host zone",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				Receiver: sample.AccAddress(),
				Amount:   sdk.NewInt(1),
			},
			err: ErrRequiredFieldEmpty,
		},
		{
			name: "invalid receiver",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Amount:   sdk.NewInt(1),
			},
			err: ErrRequiredFieldEmpty,
		},
		{
			name: "amount max int",
			msg: MsgRedeemStake{
				Creator:  sample.AccAddress(),
				HostZone: "GAIA",
				Receiver: sample.AccAddress(),
				Amount:   sdk.NewIntFromUint64(math.MaxUint64),
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
