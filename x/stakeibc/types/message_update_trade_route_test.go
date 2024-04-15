package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

func TestMsgUpdateTradeRoute(t *testing.T) {
	apptesting.SetupConfig()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	validDenom := "denom"
	validPoolId := uint64(1)
	validMaxAllowedSwapLossRate := "0.05"
	validMinSwapAmount := sdkmath.NewInt(100)
	validMaxSwapAmount := sdkmath.NewInt(10000)

	tests := []struct {
		name string
		msg  types.MsgUpdateTradeRoute
		err  string
	}{
		{
			name: "successful message",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
		},
		{
			name: "invalid authority",
			msg: types.MsgUpdateTradeRoute{
				Authority:              "",
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "invalid authority address",
		},
		{
			name: "invalid host denom",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              "",
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "missing host denom",
		},
		{
			name: "invalid reward denom",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            "",
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "missing reward denom",
		},
		{
			name: "invalid pool id",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 0,
				MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "invalid pool id",
		},
		{
			name: "invalid swap loss rate - negative",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: "-0.01",
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "max allowed swap loss rate must be between 0 and 1",
		},
		{
			name: "invalid swap loss rate - greater than 1",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: "1.01",
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "max allowed swap loss rate must be between 0 and 1",
		},
		{
			name: "invalid swap loss rate - can't cast",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: "",
				MinSwapAmount:          validMinSwapAmount,
				MaxSwapAmount:          validMaxSwapAmount,
			},
			err: "unable to cast max allowed swap loss rate to a decimal",
		},
		{
			name: "invalid min/max swap amount",
			msg: types.MsgUpdateTradeRoute{
				Authority:              authority,
				HostDenom:              validDenom,
				RewardDenom:            validDenom,
				PoolId:                 validPoolId,
				MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
				MinSwapAmount:          sdkmath.NewInt(10),
				MaxSwapAmount:          sdkmath.NewInt(5),
			},
			err: "min swap amount cannot be greater than max swap amount",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "update_trade_route")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), authority)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
