package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v28/app/apptesting"
	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

func TestMsgBurn_ValidateBasic(t *testing.T) {
	validAddress, _ := apptesting.GenerateTestAddrs()

	tests := []struct {
		name          string
		msg           types.MsgBurn
		expectedError string
	}{
		{
			name: "valid inputs",
			msg: types.MsgBurn{
				Burner: validAddress,
				Amount: sdkmath.NewInt(1_200_000),
			},
		},
		{
			name: "invalid address",
			msg: types.MsgBurn{
				Burner: "invalid_address",
				Amount: sdkmath.NewInt(1_200_000),
			},
			expectedError: sdkerrors.ErrInvalidAddress.Error(),
		},
		{
			name: "invalid address: wrong chain's bech32prefix",
			msg: types.MsgBurn{
				Burner: "celestia1yjq0n2ewufluenyyvj2y9sead9jfstpxnqv2xz",
				Amount: sdkmath.NewInt(1_200_000),
			},
			expectedError: sdkerrors.ErrInvalidAddress.Error(),
		},
		{
			name: "amount below threshold",
			msg: types.MsgBurn{
				Burner: validAddress,
				Amount: sdkmath.NewInt(999_999),
			},
			expectedError: "amount (999999ustrd) is below 1 STRD minimum",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actualError := tc.msg.ValidateBasic()
			if tc.expectedError != "" {
				require.ErrorContains(t, actualError, tc.expectedError)
				return
			}
			require.NoError(t, actualError)
		})
	}
}
