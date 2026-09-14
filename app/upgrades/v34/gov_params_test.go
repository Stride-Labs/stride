package v34_test

import (
	"time"

	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
)

func (s *UpgradeTestSuite) TestUpgradeUpdatesGovParams() {
	// ----- arrange -----
	s.useTestConsensusKeys()
	s.seedCurrentPOASet()

	// Start from mainnet's pre-upgrade values so the test proves the handler changes them
	initialParams, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	initialVotingPeriod := 3 * 24 * time.Hour
	initialParams.Quorum = "0.334000000000000000"
	initialParams.VotingPeriod = &initialVotingPeriod
	s.Require().NoError(s.App.GovKeeper.Params.Set(s.Ctx, initialParams))

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	// Literal values rather than the v34 constants, so a typo in the constants fails here
	params, err := s.App.GovKeeper.Params.Get(s.Ctx)
	s.Require().NoError(err)
	s.Require().Equal("0.250000000000000000", params.Quorum, "quorum")
	s.Require().Equal(5*24*time.Hour, *params.VotingPeriod, "voting period")

	expectedParams := initialParams
	expectedParams.Quorum = params.Quorum
	expectedParams.VotingPeriod = params.VotingPeriod
	s.Require().Equal(expectedParams, params, "all other gov params should be unchanged")
}
