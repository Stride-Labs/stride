package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	stakeibctypes "github.com/Stride-Labs/stride/v5/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v5/testutil/sample"
)

func TestMsgLiquidStake_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgLockupTokens
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgLockupTokens{
				Creator: "invalid_address",
				Amount:  sdkmath.NewInt(1),
				Denom:   "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: MsgLockupTokens{
				Creator: "osmo1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				Amount:  sdkmath.NewInt(1),
				Denom:   "uatom",
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid inputs",
			msg: MsgLockupTokens{
				Creator: sample.AccAddress(),
				Amount:  sdkmath.NewInt(1),
				Denom:   "uatom",
			},
		},
		{
			name: "zero amount",
			msg: MsgLockupTokens{
				Creator: sample.AccAddress(),
				Amount:  sdkmath.ZeroInt(),
				Denom:   "uatom",
			},
			err: stakeibctypes.ErrInvalidAmount,
		},
		{
			name: "empty host denom",
			msg: MsgLockupTokens{
				Creator: sample.AccAddress(),
				Amount:  sdkmath.NewInt(1),
				Denom:   "",
			},
			err: stakeibctypes.ErrRequiredFieldEmpty,
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

			// check msg_server
		})
	}
}

func TestMsgLiquidStake_GetSignBytes(t *testing.T) {
	addr := "cosmos1v9jxgu33kfsgr5"
	msg := NewMsgLockupTokens(addr, sdkmath.NewInt(1000), "ustrd")
	res := msg.GetSignBytes()

	expected := `{"type":"liquidgov/LockupTokens","value":{"amount":"1000","creator":"cosmos1v9jxgu33kfsgr5","denom":"ustrd"}}`
	require.Equal(t, expected, string(res))
}
