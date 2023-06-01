package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v9/x/interchainquery/keeper"
	"github.com/Stride-Labs/stride/v9/x/interchainquery/types"
)

func TestUnmarshalAmountFromBalanceQuery(t *testing.T) {
	type InputType int64
	const (
		rawBytes InputType = iota
		coinType
		intType
	)

	testCases := []struct {
		name           string
		inputType      InputType
		raw            []byte
		coin           sdk.Coin
		integer        sdkmath.Int
		expectedAmount sdkmath.Int
		expectedError  string
	}{
		{
			name:           "full_coin",
			inputType:      coinType,
			coin:           sdk.Coin{Denom: "denom", Amount: sdkmath.NewInt(50)},
			expectedAmount: sdkmath.NewInt(50),
		},
		{
			name:           "coin_no_denom",
			inputType:      coinType,
			coin:           sdk.Coin{Amount: sdkmath.NewInt(60)},
			expectedAmount: sdkmath.NewInt(60),
		},
		{
			name:           "coin_no_amount",
			inputType:      coinType,
			coin:           sdk.Coin{Denom: "denom"},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "zero_coin",
			inputType:      coinType,
			coin:           sdk.Coin{Amount: sdkmath.NewInt(0)},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "empty_coin",
			inputType:      coinType,
			coin:           sdk.Coin{},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "positive_int",
			inputType:      intType,
			integer:        sdkmath.NewInt(20),
			expectedAmount: sdkmath.NewInt(20),
		},
		{
			name:           "zero_int",
			inputType:      intType,
			integer:        sdkmath.NewInt(0),
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "empty_int",
			inputType:      intType,
			integer:        sdkmath.Int{},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:           "empty_bytes",
			inputType:      rawBytes,
			raw:            []byte{},
			expectedAmount: sdkmath.NewInt(0),
		},
		{
			name:          "invalid_bytes",
			inputType:     rawBytes,
			raw:           []byte{1, 2},
			expectedError: "unable to unmarshal balance query response",
		},
		{
			name:          "nil_bytes",
			inputType:     rawBytes,
			raw:           nil,
			expectedError: "query response is nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var args []byte
			var err error
			switch tc.inputType {
			case rawBytes:
				args = tc.raw
			case coinType:
				args, err = tc.coin.Marshal()
			case intType:
				args, err = tc.integer.Marshal()
			}
			require.NoError(t, err)

			if tc.expectedError == "" {
				actualAmount, err := keeper.UnmarshalAmountFromBalanceQuery(types.ModuleCdc, args)
				require.NoError(t, err)
				require.Equal(t, tc.expectedAmount.Int64(), actualAmount.Int64())
			} else {
				_, err := keeper.UnmarshalAmountFromBalanceQuery(types.ModuleCdc, args)
				require.ErrorContains(t, err, tc.expectedError)
			}
		})
	}
}
