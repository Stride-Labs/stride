package keeper_test

import (
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestChangeValidatorWeights() {
	tc := s.SetupAddValidators()

	// Add validators
	err := s.App.StakeibcKeeper.AddValidatorsProposal(s.Ctx, &tc.validMsg)
	s.Require().NoError(err)

	// Change the val1 weight to zero
	err = s.App.StakeibcKeeper.ChangeValidatorWeightsProposal(s.Ctx, &stakeibctypes.ChangeValidatorWeightsProposal{
		HostZone: tc.validMsg.HostZone,
		ValAddrs: []string{tc.validMsg.Validators[0].Address},
		Weights:  []uint64{0},
	})
	s.Require().NoError(err)

	// restrict max number of validators
	params := s.App.StakeibcKeeper.GetParams(s.Ctx)
	params.SafetyNumValidators = uint64(len(tc.validMsg.Validators)) - 1
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	// Change the val1 weight to non-zero fails for SafetyNumValidators
	err = s.App.StakeibcKeeper.ChangeValidatorWeightsProposal(s.Ctx, &stakeibctypes.ChangeValidatorWeightsProposal{
		HostZone: tc.validMsg.HostZone,
		ValAddrs: []string{tc.validMsg.Validators[0].Address},
		Weights:  []uint64{1},
	})
	s.Require().Error(err)

	// increase max number of validators restriction
	params.SafetyNumValidators = uint64(len(tc.validMsg.Validators))
	s.App.StakeibcKeeper.SetParams(s.Ctx, params)

	// Change the val1 weight to non-zero success for SafetyNumValidators increase
	err = s.App.StakeibcKeeper.ChangeValidatorWeightsProposal(s.Ctx, &stakeibctypes.ChangeValidatorWeightsProposal{
		HostZone: tc.validMsg.HostZone,
		ValAddrs: []string{tc.validMsg.Validators[0].Address},
		Weights:  []uint64{1},
	})
	s.Require().NoError(err)
}
