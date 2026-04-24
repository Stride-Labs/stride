package v32_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	v32 "github.com/Stride-Labs/stride/v32/app/upgrades/v32"
	"github.com/Stride-Labs/stride/v32/utils"
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
	// Setup state for steps that require it
	checkValidatorWeights := s.SetupTestUpdateValidatorWeights()

	// Run upgrade
	s.ConfirmUpgradeSucceeded(v32.UpgradeName)

	// Verify post-upgrade state
	s.VerifyMinDepositIncreased()
	s.VerifyMaxValidatorWeightIncreased()
	checkValidatorWeights()
}

func (s *UpgradeTestSuite) VerifyMinDepositIncreased() {
	params, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err, "no error expected when getting gov params")
	s.Require().Equal(utils.BaseStrideDenom, params.MinDeposit[0].Denom, "min deposit denom")
	s.Require().Equal(v32.MinDeposit, params.MinDeposit[0].Amount, "min deposit amount")
}

func (s *UpgradeTestSuite) VerifyMaxValidatorWeightIncreased() {
	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	s.Require().Equal(v32.ValidatorWeightCap, params.ValidatorWeightCap, "validator weight cap")
}

func (s *UpgradeTestSuite) SetupTestUpdateValidatorWeights() func() {
	// TODO: seed host zones with validators whose weights should be updated

	// Return callback to check store after upgrade
	return func() {
		// TODO: assert validator weights were updated as expected
	}
}
