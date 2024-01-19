package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v17/testutil/sample"
)

// ----------------------------------------------
//               MsgLiquidStake
// ----------------------------------------------

func TestMsgLiquidStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgLiquidStake
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgLiquidStake{
				Staker:       "invalid_address",
				NativeAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: MsgLiquidStake{
				Staker:       "celestia1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				NativeAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid inputs",
			msg: MsgLiquidStake{
				Staker:       sample.AccAddress(),
				NativeAmount: sdkmath.NewInt(1200000),
			},
		},
		{
			name: "amount below threshold",
			msg: MsgLiquidStake{
				Staker:       sample.AccAddress(),
				NativeAmount: sdkmath.NewInt(20000),
			},
			err: ErrInvalidAmountBelowMinimum,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// check validatebasic()
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMsgLiquidStake_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := NewMsgLiquidStake(addr, sdkmath.NewInt(1000))
	res := msg.GetSignBytes()

	expected := `{"type":"staketia/MsgLiquidStake","value":{"native_amount":"1000","staker":"stride1v9jxgu33kfsgr5"}}`
	require.Equal(t, expected, string(res))
}

// ----------------------------------------------
//               MsgRedeemStake
// ----------------------------------------------

func TestMsgRedeemStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRedeemStake
		err  error
	}{
		{
			name: "success",
			msg: MsgRedeemStake{
				Redeemer:      sample.AccAddress(),
				StTokenAmount: sdkmath.NewInt(1000000),
			},
		},
		{
			name: "invalid creator",
			msg: MsgRedeemStake{
				Redeemer:      "invalid_address",
				StTokenAmount: sdkmath.NewInt(1000000),
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "amount below threshold",
			msg: MsgRedeemStake{
				Redeemer:      sample.AccAddress(),
				StTokenAmount: sdkmath.NewInt(20000),
			},
			err: ErrInvalidAmountBelowMinimum,
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

func TestMsgRedeemStake_GetSignBytes(t *testing.T) {
	addr := "stride1v9jxgu33kfsgr5"
	msg := NewMsgRedeemStake(addr, sdkmath.NewInt(1000000))
	res := msg.GetSignBytes()

	expected := `{"type":"staketia/MsgRedeemStake","value":{"redeemer":"stride1v9jxgu33kfsgr5","st_token_amount":"1000000"}}`
	require.Equal(t, expected, string(res))
}
