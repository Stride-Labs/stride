package v27_test

import (
	"testing"

	"github.com/cometbft/cometbft/libs/os"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	v27 "github.com/Stride-Labs/stride/v26/app/upgrades/v27"
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
	upgradeHeight := int64(4)

	s.ConfirmUpgradeSucceededs(v27.UpgradeName, upgradeHeight)

	// Confirm consumer ID is set to 1
	params := s.App.ConsumerKeeper.GetConsumerParams(s.Ctx)
	s.Require().Equal(params.ConsumerId, "1")
}

func (s *UpgradeTestSuite) TestDistributionFix() {
	jsonData := os.MustReadFile("test_dist_genesis.json")

	var genState disttypes.GenesisState
	s.App.AppCodec().MustUnmarshalJSON(jsonData, &genState)

	s.App.DistrKeeper.InitGenesis(s.Ctx, genState)

	// TODO:
	// test that things are failing
	// fix state
	// test that things are not failing
}
