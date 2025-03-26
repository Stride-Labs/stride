package v22_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	v22 "github.com/Stride-Labs/stride/v26/app/upgrades/v22"
	stakeibckeeper "github.com/Stride-Labs/stride/v26/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
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
	otherHostChainId := "chain-0"

	// Create three host zones
	chainIds := []string{
		v22.OsmosisChainId,
		v22.DydxChainId,
		otherHostChainId,
	}
	for _, chainId := range chainIds {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{ChainId: chainId})
	}

	// Run the upgrade
	s.ConfirmUpgradeSucceeded(v22.UpgradeName)

	// Confirm the max ICA messages on each host zone
	for _, chainId := range chainIds {
		expectedMaxMessages, ok := v22.MaxMessagesPerIcaByHost[chainId]
		if !ok {
			expectedMaxMessages = stakeibckeeper.DefaultMaxMessagesPerIcaTx
		}
		hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
		s.Require().True(found, "host zone %s should have been found", chainId)
		s.Require().Equal(expectedMaxMessages, hostZone.MaxMessagesPerIcaTx, "max messages for %s", chainId)
	}
}
