package v22_1_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v22/app/apptesting"
	stakeibctypes "github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)
	minTransferAmount := sdkmath.NewInt(100)

	// Create a trade route with the deprecated trade config
	tradeRoutes := stakeibctypes.TradeRoute{
		HostDenomOnHostZone:     "host-denom",
		RewardDenomOnRewardZone: "reward-denom",

		TradeConfig: stakeibctypes.TradeConfig{ //nolint:staticcheck
			MinSwapAmount: minTransferAmount,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoutes)

	// Run the upgrade
	s.ConfirmUpgradeSucceededs("v23", dummyUpgradeHeight)

	// Confirm trade route was migrated
	for _, tradeRoute := range s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx) {
		s.Require().Equal(tradeRoute.MinTransferAmount, minTransferAmount)
	}
}
