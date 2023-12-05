package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestMsgCreateTradeRoute() {
	rewardChainId := "reward-0"
	tradeChainId := "trade-0"

	hostConnectionId := "connection-0"
	rewardConnectionId := "connection-1"
	tradeConnectionId := "connection-2"

	hostToRewardChannelId := "channel-100"
	rewardToTradeChannelId := "channel-200"
	tradeToHostChannelId := "channel-300"

	rewardDenomOnHost := "ibc/reward-on-host"
	rewardDenomOnReward := RewardDenom
	rewardDenomOnTrade := "ibc/reward-on-trade"
	hostDenomOnTrade := "ibc/host-on-trade"
	hostDenomOnHost := HostDenom

	withdrawalAddress := "withdrawal-address"
	unwindAddress := "unwind-address"

	poolId := uint64(100)
	maxAllowedSwapLossRate := "0.05"
	minSwapAmount := sdkmath.NewInt(100)
	maxSwapAmount := sdkmath.NewInt(1_000)

	// Mock out connections for the reward an trade chain so that an ICA registration can be submitted
	s.MockClientAndConnection(rewardChainId, "07-tendermint-0", rewardConnectionId)
	s.MockClientAndConnection(tradeChainId, "07-tendermint-1", tradeConnectionId)

	// Register an exisiting ICA account for the unwind ICA to test that
	// existing accounts are re-used
	// TODO [DYDX]: Replace with trade route owner
	owner := types.FormatICAAccountOwner(rewardChainId, types.ICAAccountType_CONVERTER_UNWIND)
	s.MockICAChannel(rewardConnectionId, "channel-0", owner, unwindAddress)

	// Create a host zone with an exisiting withdrawal address
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         hostConnectionId,
		WithdrawalIcaAddress: withdrawalAddress,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Define a valid message given the parameters above
	msg := types.MsgCreateTradeRoute{
		Authority:   Authority,
		HostChainId: HostChainId,

		StrideToRewardConnectionId: rewardConnectionId,
		StrideToTradeConnectionId:  tradeConnectionId,

		HostToRewardTransferChannelId:  hostToRewardChannelId,
		RewardToTradeTransferChannelId: rewardToTradeChannelId,
		TradeToHostTransferChannelId:   tradeToHostChannelId,

		RewardDenomOnHost:   rewardDenomOnHost,
		RewardDenomOnReward: rewardDenomOnReward,
		RewardDenomOnTrade:  rewardDenomOnTrade,
		HostDenomOnTrade:    hostDenomOnTrade,
		HostDenomOnHost:     hostDenomOnHost,

		PoolId:                 poolId,
		MaxAllowedSwapLossRate: maxAllowedSwapLossRate,
		MinSwapAmount:          minSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	// Build out the expected trade route given the above
	expectedTradeRoute := types.TradeRoute{
		RewardDenomOnHostZone:   rewardDenomOnHost,
		RewardDenomOnRewardZone: rewardDenomOnReward,
		RewardDenomOnTradeZone:  rewardDenomOnTrade,
		HostDenomOnTradeZone:    hostDenomOnTrade,
		HostDenomOnHostZone:     hostDenomOnHost,

		HostAccount: types.ICAAccount{
			ChainId:      HostChainId,
			Type:         types.ICAAccountType_WITHDRAWAL,
			ConnectionId: hostConnectionId,
			Address:      withdrawalAddress,
		},
		RewardAccount: types.ICAAccount{
			ChainId:      rewardChainId,
			Type:         types.ICAAccountType_CONVERTER_UNWIND,
			ConnectionId: rewardConnectionId,
			Address:      unwindAddress,
		},
		TradeAccount: types.ICAAccount{
			ChainId:      tradeChainId,
			Type:         types.ICAAccountType_CONVERTER_TRADE,
			ConnectionId: tradeConnectionId,
		},

		HostToRewardChannelId:  hostToRewardChannelId,
		RewardToTradeChannelId: rewardToTradeChannelId,
		TradeToHostChannelId:   tradeToHostChannelId,

		TradeConfig: types.TradeConfig{
			PoolId:               poolId,
			SwapPrice:            sdk.ZeroDec(),
			PriceUpdateTimestamp: 0,

			MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(maxAllowedSwapLossRate),
			MinSwapAmount:          minSwapAmount,
			MaxSwapAmount:          maxSwapAmount,
		},
	}

	// Create the trade route and confirm the created route matches expectations
	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when creating trade route")

	actualTradeRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been created")
	s.Require().Equal(expectedTradeRoute, actualTradeRoute, "trade route")
}
