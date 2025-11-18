package v20_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v30/app/apptesting"
	v20 "github.com/Stride-Labs/stride/v30/app/upgrades/v20"
	stakeibctypes "github.com/Stride-Labs/stride/v30/x/stakeibc/types"
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
	// Create a dydx host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: v20.DydxChainId,
	})

	// Run the upgrade
	s.ConfirmUpgradeSucceeded(v20.UpgradeName)

	// Confirm the treasury address was added to dydx
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v20.DydxChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(v20.DydxCommunityPoolTreasuryAddress, hostZone.CommunityPoolTreasuryAddress,
		"community pool treasury address")
}
