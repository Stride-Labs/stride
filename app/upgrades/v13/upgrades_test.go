package v13_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v13/app/apptesting"
	stakeibctypes "github.com/Stride-Labs/stride/v13/x/stakeibc/types"
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

	// Clear the reward denoms in the consumer keeper
	consumerParams := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	consumerParams.RewardDenoms = []string{"denomA", "denomB"}
	s.App.ConsumerKeeper.SetParams(s.Ctx, consumerParams)

	// Add host zones
	hostZones := []stakeibctypes.HostZone{
		{ChainId: "cosmoshub-4", HostDenom: "uatom"},
		{ChainId: "osmosis-1", HostDenom: "uosmo"},
		{ChainId: "juno-1", HostDenom: "ujuno"},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Submit the upgrade which should register the new host denoms
	s.ConfirmUpgradeSucceededs("v13", dummyUpgradeHeight)

	// Confirm the new reward denoms were registered
	expectedRewardDenoms := []string{"denomA", "denomB", "stuatom", "stuosmo", "stujuno"}
	consumerParams = s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().ElementsMatch(expectedRewardDenoms, consumerParams.RewardDenoms)
}
