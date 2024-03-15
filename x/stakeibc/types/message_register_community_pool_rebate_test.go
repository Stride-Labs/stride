package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v19/app/apptesting"
	"github.com/Stride-Labs/stride/v19/x/stakeibc/types"
)

func TestMsgRegisterCommunityPoolRebate(t *testing.T) {
	apptesting.SetupConfig()

	validNotAdminAddress, invalidAddress := apptesting.GenerateTestAddrs()
	validAdminAddress, ok := apptesting.GetAdminAddress()
	require.True(t, ok)

	validChainId := "chain-0"
	validRebatePercentage := sdk.MustNewDecFromStr("0.1")
	validLiquidStakeAmount := sdk.NewInt(1000)

	tests := []struct {
		name string
		msg  types.MsgRegisterCommunityPoolRebate
		err  string
	}{
		{
			name: "valid message",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  validRebatePercentage,
				LiquidStakeAmount: validLiquidStakeAmount,
			},
		},
		{
			name: "invalid address",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           invalidAddress,
				ChainId:           validChainId,
				RebatePercentage:  validRebatePercentage,
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "invalid creator address",
		},
		{
			name: "not admin address",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validNotAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  validRebatePercentage,
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "not an admin",
		},
		{
			name: "invalid chain ID",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           "",
				RebatePercentage:  validRebatePercentage,
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "chain ID must be specified",
		},
		{
			name: "invalid rebate percentage - nil",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "rebate percentage, must be between [0, 1)",
		},
		{
			name: "invalid rebate percentage - less than 0",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  sdk.MustNewDecFromStr("0.5").Neg(),
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "rebate percentage, must be between [0, 1)",
		},
		{
			name: "invalid rebate percentage - one",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  sdk.OneDec(),
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "rebate percentage, must be between [0, 1)",
		},
		{
			name: "invalid rebate percentage - greater than one",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  sdk.MustNewDecFromStr("1.1"),
				LiquidStakeAmount: validLiquidStakeAmount,
			},
			err: "rebate percentage, must be between [0, 1)",
		},
		{
			name: "valid zero rebate",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  sdk.ZeroDec(),
				LiquidStakeAmount: validLiquidStakeAmount,
			},
		},
		{
			name: "invalid liquid stake amount - nil",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:          validAdminAddress,
				ChainId:          validChainId,
				RebatePercentage: validRebatePercentage,
			},
			err: "invalid liquid stake amount, must be greater than 0",
		},
		{
			name: "invalid liquid stake amount - less than 0",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  validRebatePercentage,
				LiquidStakeAmount: sdkmath.NewInt(1).Neg(),
			},
			err: "invalid liquid stake amount, must be greater than 0",
		},
		{
			name: "invalid liquid stake amount - zero",
			msg: types.MsgRegisterCommunityPoolRebate{
				Creator:           validAdminAddress,
				ChainId:           validChainId,
				RebatePercentage:  validRebatePercentage,
				LiquidStakeAmount: sdkmath.ZeroInt(),
			},
			err: "invalid liquid stake amount, must be greater than 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.err == "" {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
				require.Equal(t, test.msg.Route(), types.RouterKey)
				require.Equal(t, test.msg.Type(), "register_community_pool_rebate")

				signers := test.msg.GetSigners()
				require.Equal(t, len(signers), 1)
				require.Equal(t, signers[0].String(), validAdminAddress)
			} else {
				require.ErrorContains(t, test.msg.ValidateBasic(), test.err, "test: %v", test.name)
			}
		})
	}
}
