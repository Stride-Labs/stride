package keeper_test

import (
	_ "github.com/stretchr/testify/suite"
)

func (s *KeeperTestSuite) TestEnableStrictUnbondingCap_CapNotSet() {

	// make sure StrictUnbondingCap is not set
	s.Require().False(s.App.StakeibcKeeper.IsUndelegateHostPrevented(s.Ctx), "undelegate host prevented")
}

func (s *KeeperTestSuite) TestEnableStrictUnbondingCap_CapSet() {

	// set undelegate Prevented
	err := s.App.StakeibcKeeper.SetUndelegateHostPrevented(s.Ctx)
	s.Require().NoError(err, "set undelegate host prevented")

	// make sure StrictUnbondingCap is set
	s.Require().True(s.App.StakeibcKeeper.IsUndelegateHostPrevented(s.Ctx), "strict unbonding cap set to true")
}
