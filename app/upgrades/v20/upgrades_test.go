package v20_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v19/app/apptesting"
	v20 "github.com/Stride-Labs/stride/v19/app/upgrades/v20"
	stakeibctypes "github.com/Stride-Labs/stride/v19/x/stakeibc/types"
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

	// Create a dydx host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: v20.DydxChainId,
	})

	// Run the upgrade
	s.ConfirmUpgradeSucceededs("v20", dummyUpgradeHeight)

	// Confirm the treasury address was added to dydx
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v20.DydxChainId)
	s.Require().True(found, "host ozne should have been found")
	s.Require().Equal(v20.DydxCommunityPoolTreasuryAddress, hostZone.CommunityPoolTreasuryAddress,
		"community pool treasury address")
}
