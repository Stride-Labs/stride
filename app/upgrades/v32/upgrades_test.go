package v32_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v32/app/apptesting"
	v32 "github.com/Stride-Labs/stride/v32/app/upgrades/v32"
	"github.com/Stride-Labs/stride/v32/utils"
	stakeibctypes "github.com/Stride-Labs/stride/v32/x/stakeibc/types"
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
	checkValidatorWeights := s.SetupTestUpdateValidatorWeights()

	s.ConfirmUpgradeSucceeded(v32.UpgradeName)

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
	for chainId, vals := range v32.OldValidators {
		var validators []*stakeibctypes.Validator
		for _, val := range vals {
			validators = append(validators, &stakeibctypes.Validator{
				Name:                      val.Name,
				Address:                   val.Address,
				Weight:                    val.Weight,
				Delegation:                sdkmath.ZeroInt(),
				SlashQueryProgressTracker: sdkmath.ZeroInt(),
				SlashQueryCheckpoint:      sdkmath.ZeroInt(),
				SharesToTokensRate:        sdkmath.LegacyOneDec(),
			})
		}

		s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
			ChainId:    chainId,
			Validators: validators,
		})
	}

	return func() {
		for chainId, weights := range v32.TargetWeights {
			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
			s.Require().True(found, "host zone %s should exist", chainId)

			actualWeights := map[string]uint64{}
			for _, val := range hostZone.Validators {
				actualWeights[val.Address] = val.Weight
			}

			s.Require().Equal(len(weights), len(actualWeights),
				"%s: validator count mismatch", chainId)

			for _, w := range weights {
				actual, exists := actualWeights[w.Address]
				s.Require().True(exists, "%s: validator %s should exist", chainId, w.Address)
				s.Require().Equal(w.Weight, actual, "%s: weight mismatch for %s", chainId, w.Address)
			}
		}
	}
}
