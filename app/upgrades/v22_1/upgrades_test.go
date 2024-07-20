package v22_1_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	ibcwasmtypes "github.com/cosmos/ibc-go/modules/light-clients/08-wasm/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v23/app/apptesting"
	v22_1 "github.com/Stride-Labs/stride/v23/app/upgrades/v22_1"
	stakeibctypes "github.com/Stride-Labs/stride/v23/x/stakeibc/types"
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
	dummyUpgradeHeight := int64(4)

	minTransferAmount := sdkmath.NewInt(100)

	// Set the allowed ibc clients to an empty list
	s.App.IBCKeeper.ClientKeeper.SetParams(s.Ctx, ibcclienttypes.Params{AllowedClients: []string{}})

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
	s.ConfirmUpgradeSucceededs(v22_1.UpgradeName, dummyUpgradeHeight)

	// Confirm trade route was migrated
	for _, tradeRoute := range s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx) {
		s.Require().Equal(tradeRoute.MinTransferAmount, minTransferAmount)
	}

	// Confirm the ibc wasm client was added
	params := s.App.IBCKeeper.ClientKeeper.GetParams(s.Ctx)
	s.Require().Equal([]string{ibcwasmtypes.Wasm}, params.AllowedClients, "ibc allowed clients")
}
