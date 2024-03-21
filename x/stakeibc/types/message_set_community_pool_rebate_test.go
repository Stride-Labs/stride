package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v20/app/apptesting"
	"github.com/Stride-Labs/stride/v20/x/stakeibc/types"
)

func TestMsgSetCommunityPoolRebate(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"
	validRebateRate := sdk.MustNewDecFromStr("0.1")
	validLiquidStakedAmount := sdk.NewInt(1000)

	tests := []struct {
		name string
		msg  types.MsgSetCommunityPoolRebate
		err  string
	}{
		{
			name: "valid message",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                validRebateRate,
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   invalidAddress,
				ChainId:                   validChainId,
				RebateRate:                validRebateRate,
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
			err: "invalid creator address",
		},
		{
			name: "not admin address",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validNotAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                validRebateRate,
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
			err: "not an admin",
		},
		{
			name: "invalid chain ID",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   "",
				RebateRate:                validRebateRate,
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
			err: "chain ID must be specified",
		},
		{
			name: "invalid rebate rate - nil",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
			err: "rebate rate, must be a decimal between 0 and 1 (inclusive)",
		},
		{
			name: "invalid rebate rate - less than 0",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                sdk.MustNewDecFromStr("0.5").Neg(),
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
			err: "rebate rate, must be a decimal between 0 and 1 (inclusive)",
		},
		{
			name: "valid rebate rate - one",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                sdk.OneDec(),
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
		},
		{
			name: "invalid rebate rate - greater than one",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                sdk.MustNewDecFromStr("1.1"),
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
			err: "rebate rate, must be a decimal between 0 and 1 (inclusive)",
		},
		{
			name: "valid zero rebate",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                sdk.ZeroDec(),
				LiquidStakedStTokenAmount: validLiquidStakedAmount,
			},
		},
		{
			name: "invalid liquid stake amount - nil",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:    validAdminAddress,
				ChainId:    validChainId,
				RebateRate: validRebateRate,
			},
			err: "invalid liquid stake amount, must be greater than or equal to zero",
		},
		{
			name: "invalid liquid stake amount - less than 0",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                validRebateRate,
				LiquidStakedStTokenAmount: sdkmath.NewInt(1).Neg(),
			},
			err: "invalid liquid stake amount, must be greater than or equal to zero",
		},
		{
			name: "valid liquid stake amount - zero",
			msg: types.MsgSetCommunityPoolRebate{
				Creator:                   validAdminAddress,
				ChainId:                   validChainId,
				RebateRate:                validRebateRate,
				LiquidStakedStTokenAmount: sdkmath.ZeroInt(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "set_community_pool_rebate")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
