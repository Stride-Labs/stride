package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupTestCreateTradeRoute() (msg types.MsgCreateTradeRoute, expectedTradeRoute types.TradeRoute) {
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
	owner := types.FormatTradeRouteICAOwner(rewardChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_UNWIND)
	s.MockICAChannel(rewardConnectionId, "channel-0", owner, unwindAddress)

	// Create a host zone with an exisiting withdrawal address
	hostZone := types.HostZone{
		ChainId:              HostChainId,
		ConnectionId:         hostConnectionId,
		WithdrawalIcaAddress: withdrawalAddress,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Define a valid message given the parameters above
	msg = types.MsgCreateTradeRoute{
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
	expectedTradeRoute = types.TradeRoute{
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

	return msg, expectedTradeRoute
}

// Helper function to create a trade route and check the created route matched expectations
func (s *KeeperTestSuite) submitCreateTradeRouteAndValidate(msg types.MsgCreateTradeRoute, expectedRoute types.TradeRoute) {
	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when creating trade route")

	actualRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, msg.RewardDenomOnReward, msg.HostDenomOnHost)
	s.Require().True(found, "trade route should have been created")
	s.Require().Equal(expectedRoute, actualRoute, "trade route")
}

// Tests a successful trade route creation
func (s *KeeperTestSuite) TestCreateTradeRoute_Success() {
	msg, expectedRoute := s.SetupTestCreateTradeRoute()
	s.submitCreateTradeRouteAndValidate(msg, expectedRoute)
}

// Tests creating a trade route that uses the default pool config values
func (s *KeeperTestSuite) TestCreateTradeRoute_Success_DefaultPoolConfig() {
	msg, expectedRoute := s.SetupTestCreateTradeRoute()

	// Update the message and remove some trade config parameters
	// so that the defaults are used
	msg.MaxSwapAmount = sdk.ZeroInt()
	msg.MaxAllowedSwapLossRate = ""

	expectedRoute.TradeConfig.MaxAllowedSwapLossRate = sdk.MustNewDecFromStr(keeper.DefaultMaxAllowedSwapLossRate)
	expectedRoute.TradeConfig.MaxSwapAmount = keeper.DefaultMaxSwapAmount

	s.submitCreateTradeRouteAndValidate(msg, expectedRoute)
}

// Tests trying to create a route from an invalid authority
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_Authority() {
	msg, _ := s.SetupTestCreateTradeRoute()

	msg.Authority = "not-gov-address"

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "invalid authority")
}

// Tests creating a duplicate trade route
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_DuplicateTradeRoute() {
	msg, _ := s.SetupTestCreateTradeRoute()

	// Store down a trade route so the tx hits a duplicate trade route error
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
	})

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "Trade route already exists")
}

// Tests creating a trade route when the host zone or withdrawal address does not exist
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_HostZoneNotRegistered() {
	msg, _ := s.SetupTestCreateTradeRoute()

	// Remove the host zone withdrawal address and confirm it fails
	invalidHostZone := s.MustGetHostZone(HostChainId)
	invalidHostZone.WithdrawalIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, invalidHostZone)

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "withdrawal account not initialized on host zone")

	// Remove the host zone completely and check that that also fails
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	_, err = s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "host zone not found")
}

// Tests creating a trade route where the ICA channels cannot be created
// because the ICA connections do not exist
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_ConnectionNotFound() {
	// Test with non-existent reward connection
	msg, _ := s.SetupTestCreateTradeRoute()
	msg.StrideToRewardConnectionId = "connection-X"

	// Remove the host zone completely and check that that also fails
	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to register the unwind ICA account: connection connection-X not found")

	// Setup again, but this time use a non-existent trade connection
	msg, _ = s.SetupTestCreateTradeRoute()
	msg.StrideToTradeConnectionId = "connection-Y"

	_, err = s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to register the trade ICA account: connection connection-Y not found")
}

// Tests creating a trade route where the ICA registration step fails
func (s *KeeperTestSuite) TestCreateTradeRoute_Failure_UnableToRegisterICA() {
	msg, expectedRoute := s.SetupTestCreateTradeRoute()

	// Disable ICA middleware for the trade channel so the ICA fails
	tradeAccount := expectedRoute.TradeAccount
	tradeOwner := types.FormatTradeRouteICAOwner(tradeAccount.ChainId, RewardDenom, HostDenom, types.ICAAccountType_CONVERTER_TRADE)
	tradePortId, _ := icatypes.NewControllerPortID(tradeOwner)
	s.App.ICAControllerKeeper.SetMiddlewareDisabled(s.Ctx, tradePortId, tradeAccount.ConnectionId)

	_, err := s.GetMsgServer().CreateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "unable to register the trade ICA account")
}
