package types_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/Stride-Labs/stride/v17/app/apptesting"
	"github.com/Stride-Labs/stride/v17/x/stakeibc/types"
)

func TestMsgCreateTradeRoute(t *testing.T) {
	apptesting.SetupConfig()

	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	validChainId := "chain-1"
	validConnectionId1 := "connection-1"
	validConnectionId2 := "connection-17"
	validTransferChannelId1 := "channel-2"
	validTransferChannelId2 := "channel-202"
	validTransferChannelId3 := "channel-40"

	validNativeDenom := "denom"
	validIBCDenom := "ibc/denom"

	validPoolId := uint64(1)
	validMaxAllowedSwapLossRate := "0.05"
	validMinSwapAmount := sdkmath.NewInt(100)
	validMaxSwapAmount := sdkmath.NewInt(10000)

	validMessage := types.MsgCreateTradeRoute{
		Authority: authority,

		HostChainId:                validChainId,
		StrideToRewardConnectionId: validConnectionId1,
		StrideToTradeConnectionId:  validConnectionId2,

		HostToRewardTransferChannelId:  validTransferChannelId1,
		RewardToTradeTransferChannelId: validTransferChannelId2,
		TradeToHostTransferChannelId:   validTransferChannelId3,

		RewardDenomOnHost:   validIBCDenom,
		RewardDenomOnReward: validNativeDenom,
		RewardDenomOnTrade:  validIBCDenom,
		HostDenomOnTrade:    validIBCDenom,
		HostDenomOnHost:     validNativeDenom,

		PoolId:                 validPoolId,
		MaxAllowedSwapLossRate: validMaxAllowedSwapLossRate,
		MinSwapAmount:          validMinSwapAmount,
		MaxSwapAmount:          validMaxSwapAmount,
	}

	// Validate successful message
	require.NoError(t, validMessage.ValidateBasic(), "valid message")
	require.Equal(t, validMessage.Route(), types.RouterKey)
	require.Equal(t, validMessage.Type(), "create_trade_route")

	signers := validMessage.GetSigners()
	require.Equal(t, len(signers), 1)
	require.Equal(t, signers[0].String(), authority)

	// Remove authority - confirm invalid
	invalidMessage := validMessage
	invalidMessage.Authority = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid authority address")

	// Set invalid chain ID - confirm invalid
	invalidMessage = validMessage
	invalidMessage.HostChainId = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "host chain ID cannot be empty")

	// Set invalid connection IDs - confirm invalid
	invalidMessage = validMessage
	invalidMessage.StrideToRewardConnectionId = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid stride to reward connection ID")

	invalidMessage = validMessage
	invalidMessage.StrideToTradeConnectionId = "connection-X"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid stride to trade connection ID")

	// Set invalid channel IDs - confirm invalid
	invalidMessage = validMessage
	invalidMessage.HostToRewardTransferChannelId = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid host to reward channel ID")

	invalidMessage = validMessage
	invalidMessage.RewardToTradeTransferChannelId = "channel-"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid reward to trade channel ID")

	invalidMessage = validMessage
	invalidMessage.TradeToHostTransferChannelId = "channel-X"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid trade to host channel ID")

	// Set invalid denom's - confirm invalid
	invalidMessage = validMessage
	invalidMessage.RewardDenomOnHost = "not-ibc-denom"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid reward denom on host")

	invalidMessage = validMessage
	invalidMessage.RewardDenomOnReward = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid reward denom on reward")

	invalidMessage = validMessage
	invalidMessage.RewardDenomOnTrade = "not-ibc-denom"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid reward denom on trade")

	invalidMessage = validMessage
	invalidMessage.HostDenomOnTrade = "not-ibc-denom"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid host denom on trade")

	invalidMessage = validMessage
	invalidMessage.HostDenomOnHost = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid host denom on host")

	// Set invalid pool configurations - confirm invalid
	invalidMessage = validMessage
	invalidMessage.PoolId = 0
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "invalid pool id")

	invalidMessage = validMessage
	invalidMessage.MinSwapAmount = sdkmath.NewInt(10)
	invalidMessage.MaxSwapAmount = sdkmath.NewInt(9)
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "min swap amount cannot be greater than max swap amount")

	invalidMessage = validMessage
	invalidMessage.MaxAllowedSwapLossRate = ""
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "unable to cast max allowed swap loss rate to a decimal")

	invalidMessage = validMessage
	invalidMessage.MaxAllowedSwapLossRate = "-0.01"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "max allowed swap loss rate must be between 0 and 1")

	invalidMessage = validMessage
	invalidMessage.MaxAllowedSwapLossRate = "1.01"
	require.ErrorContains(t, invalidMessage.ValidateBasic(), "max allowed swap loss rate must be between 0 and 1")
}

func TestValidateConnectionId(t *testing.T) {
	require.NoError(t, types.ValidateConnectionId("connection-0"))
	require.NoError(t, types.ValidateConnectionId("connection-10"))
	require.NoError(t, types.ValidateConnectionId("connection-1203"))

	require.ErrorContains(t, types.ValidateConnectionId("connection-X"), "invalid connection-id (connection-X)")
	require.ErrorContains(t, types.ValidateConnectionId(""), "invalid connection-id ()")
}

func TestValidateChannelId(t *testing.T) {
	require.NoError(t, types.ValidateChannelId("channel-0"))
	require.NoError(t, types.ValidateChannelId("channel-10"))
	require.NoError(t, types.ValidateChannelId("channel-1203"))

	require.ErrorContains(t, types.ValidateChannelId("channel-X"), "invalid channel-id (channel-X)")
	require.ErrorContains(t, types.ValidateChannelId(""), "invalid channel-id ()")
}

func TestValidateDenom(t *testing.T) {
	require.NoError(t, types.ValidateDenom("denom", false))
	require.NoError(t, types.ValidateDenom("ibc/denom", false))
	require.NoError(t, types.ValidateDenom("ibc/denom", true))

	require.ErrorContains(t, types.ValidateDenom("", false), "denom is empty")
	require.ErrorContains(t, types.ValidateDenom("", true), "denom is empty")
	require.ErrorContains(t, types.ValidateDenom("denom", true), "denom (denom) should have ibc prefix")
}
