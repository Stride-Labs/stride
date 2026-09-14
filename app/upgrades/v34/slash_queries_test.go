package v34_test

import (
	v34 "github.com/Stride-Labs/stride/v34/app/upgrades/v34"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

const (
	otherHostZoneChainId  = "other-chain"
	unlistedValidatorName = "unlisted-validator"
	unlistedQueryId       = "unlisted-query"
)

// TestUpgradeClearsStuckSlashQueries runs the full handler (rather than the
// helpers directly) so the wiring in CreateUpgradeHandler is covered too.
func (s *UpgradeTestSuite) TestUpgradeClearsStuckSlashQueries() {
	// ----- arrange -----
	s.useTestConsensusKeys()
	s.seedCurrentPOASet()

	// The last listed validator and query are left out of state to confirm the
	// handler skips entries that no longer exist instead of failing
	presentValidatorNames := v34.StuckSlashQueryValidators[:len(v34.StuckSlashQueryValidators)-1]
	presentQueryIds := v34.StuckQueryIds[:len(v34.StuckQueryIds)-1]

	validators := []*stakeibctypes.Validator{{Name: unlistedValidatorName, SlashQueryInProgress: true}}
	for _, name := range presentValidatorNames {
		validators = append(validators, &stakeibctypes.Validator{Name: name, SlashQueryInProgress: true})
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:    v34.StuckSlashQueryHostZone,
		Validators: validators,
	})

	// A validator sharing a listed name on a different host zone must be left alone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:    otherHostZoneChainId,
		Validators: []*stakeibctypes.Validator{{Name: v34.StuckSlashQueryValidators[0], SlashQueryInProgress: true}},
	})

	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqtypes.Query{Id: unlistedQueryId})
	for _, queryId := range presentQueryIds {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqtypes.Query{Id: queryId})
	}

	// ----- act -----
	s.ConfirmUpgradeSucceeded(v34.UpgradeName)

	// ----- assert -----
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.StuckSlashQueryHostZone)
	s.Require().True(found)
	s.Require().Len(hostZone.Validators, len(presentValidatorNames)+1, "no validators should be added or removed")
	for _, validator := range hostZone.Validators {
		expectedInProgress := validator.Name == unlistedValidatorName
		s.Require().Equal(expectedInProgress, validator.SlashQueryInProgress,
			"validator %s slash query in progress", validator.Name)
	}

	otherHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, otherHostZoneChainId)
	s.Require().True(found)
	s.Require().True(otherHostZone.Validators[0].SlashQueryInProgress, "validators on other host zones should not be reset")

	for _, queryId := range presentQueryIds {
		_, found := s.App.InterchainqueryKeeper.GetQuery(s.Ctx, queryId)
		s.Require().False(found, "stuck query %s should have been deleted", queryId)
	}
	_, found = s.App.InterchainqueryKeeper.GetQuery(s.Ctx, unlistedQueryId)
	s.Require().True(found, "queries not in StuckQueryIds should not be deleted")
}

func (s *UpgradeTestSuite) TestResetStuckSlashQueriesMissingHostZone() {
	s.Require().NotPanics(func() {
		v34.ResetStuckSlashQueries(s.Ctx, s.App.StakeibcKeeper)
	})

	_, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v34.StuckSlashQueryHostZone)
	s.Require().False(found, "the reset should not create the host zone")
}
