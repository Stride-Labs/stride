package v26_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v31/app/apptesting"
	v26 "github.com/Stride-Labs/stride/v31/app/upgrades/v26"
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
	s.ConfirmUpgradeSucceeded(v26.UpgradeName)

	params := s.App.ICQOracleKeeper.GetParams(s.Ctx)
	s.Require().Equal(v26.OsmosisChainId, params.OsmosisChainId, "Osmosis chain ID")
	s.Require().Equal(v26.OsmosisConnectionId, params.OsmosisConnectionId, "Osmosis connection ID")
	s.Require().Equal(v26.ICQOracleUpdateIntervalSec, params.UpdateIntervalSec, "Update interval")
	s.Require().Equal(v26.ICQOraclePriceExpirationTimeoutSec, params.PriceExpirationTimeoutSec, "Timeout")
}
