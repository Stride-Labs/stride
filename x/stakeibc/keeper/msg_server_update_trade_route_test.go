package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Helper function to update a trade route and check the updated route matched expectations
func (s *KeeperTestSuite) submitUpdateTradeRouteAndValidate(msg types.MsgUpdateTradeRoute, expectedRoute types.TradeRoute) {
	_, err := s.GetMsgServer().UpdateTradeRoute(sdk.WrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when updating trade route")

	actualRoute, found := s.App.StakeibcKeeper.GetTradeRoute(s.Ctx, RewardDenom, HostDenom)
	s.Require().True(found, "trade route should have been updated")
	s.Require().Equal(expectedRoute, actualRoute, "trade route")
}

func (s *KeeperTestSuite) TestUpdateTradeRoute() {
	poolId := uint64(100)
	maxAllowedSwapLossRate := "0.05"
	minSwapAmount := sdkmath.NewInt(100)
	maxSwapAmount := sdkmath.NewInt(1_000)

	// Create a trade route with no parameters
	initialRoute := types.TradeRoute{
		RewardDenomOnRewardZone: RewardDenom,
		HostDenomOnHostZone:     HostDenom,
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, initialRoute)

	// Define a valid message given the parameters above
	msg := types.MsgUpdateTradeRoute{
		Authority: Authority,

		RewardDenom: RewardDenom,
		HostDenom:   HostDenom,

		PoolId:                 poolId,
		MaxAllowedSwapLossRate: maxAllowedSwapLossRate,
		MinSwapAmount:          minSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	// Build out the expected trade route given the above
	expectedRoute := initialRoute
	expectedRoute.TradeConfig = types.TradeConfig{
		PoolId:               poolId,
		SwapPrice:            sdk.ZeroDec(),
		PriceUpdateTimestamp: 0,

		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(maxAllowedSwapLossRate),
		MinSwapAmount:          minSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	// Update the route and confirm the changes persisted
	s.submitUpdateTradeRouteAndValidate(msg, expectedRoute)

	// Update it again, this time using default args
	defaultMsg := msg
	defaultMsg.MaxAllowedSwapLossRate = ""
	defaultMsg.MaxSwapAmount = sdkmath.ZeroInt()

	expectedRoute.TradeConfig.MaxAllowedSwapLossRate = sdk.MustNewDecFromStr(keeper.DefaultMaxAllowedSwapLossRate)
	expectedRoute.TradeConfig.MaxSwapAmount = keeper.DefaultMaxSwapAmount

	s.submitUpdateTradeRouteAndValidate(defaultMsg, expectedRoute)

	// Test that an error is thrown if the correct authority is not specified
	invalidMsg := msg
	invalidMsg.Authority = "not-gov-address"

	_, err := s.GetMsgServer().UpdateTradeRoute(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid authority")

	// Test that an error is thrown if the route doesn't exist
	invalidMsg = msg
	invalidMsg.RewardDenom = "invalid-reward-denom"

	_, err = s.GetMsgServer().UpdateTradeRoute(sdk.WrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "trade route not found")
}
