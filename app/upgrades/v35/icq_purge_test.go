package v35_test

import (
	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icqtypes "github.com/Stride-Labs/stride/v34/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

func (s *UpgradeTestSuite) seedQueries() {
	for _, query := range []icqtypes.Query{
		{Id: "haqq-delegation", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Delegation},
		{Id: "haqq-validator", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Validator},
		{Id: "haqq-calibrate", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Calibrate},
		{Id: "haqq-withdrawal", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
		{Id: "haqq-fee", ChainId: v35.HaqqChainId, CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_FeeBalance},
		{Id: "juno-withdrawal", ChainId: "juno-1", CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
		{Id: "juno-delegation", ChainId: "juno-1", CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Delegation},
		{Id: "comdex-calibrate", ChainId: "comdex-1", CallbackModule: stakeibctypes.ModuleName, CallbackId: stakeibckeeper.ICQCallbackID_Calibrate},
		// Another module's queries whose callback ids collide with stakeibc's: never deleted
		{Id: "other-haqq-delegation", ChainId: v35.HaqqChainId, CallbackModule: "records", CallbackId: stakeibckeeper.ICQCallbackID_Delegation},
		{Id: "other-juno-withdrawal", ChainId: "juno-1", CallbackModule: "records", CallbackId: stakeibckeeper.ICQCallbackID_WithdrawalHostBalance},
	} {
		s.App.InterchainqueryKeeper.SetQuery(s.Ctx, query)
	}
}

func (s *UpgradeTestSuite) queryIds() []string {
	ids := []string{}
	for _, query := range s.App.InterchainqueryKeeper.AllQueries(s.Ctx) {
		ids = append(ids, query.Id)
	}
	return ids
}

func (s *UpgradeTestSuite) TestPurgeHaqqSlashQueries() {
	s.seedQueries()
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: v35.HaqqChainId,
		Validators: []*stakeibctypes.Validator{
			{Address: "haqqvaloper1", SlashQueryInProgress: true},
			{Address: "haqqvaloper2", SlashQueryInProgress: false},
		},
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:    "juno-1",
		Validators: []*stakeibctypes.Validator{{Address: "junovaloper1", SlashQueryInProgress: true}},
	})

	v35.PurgeHaqqSlashQueries(s.Ctx, s.App.InterchainqueryKeeper, s.App.StakeibcKeeper)

	s.Require().ElementsMatch(
		[]string{
			"haqq-withdrawal", "haqq-fee", "juno-withdrawal", "juno-delegation", "comdex-calibrate",
			"other-haqq-delegation", "other-juno-withdrawal",
		},
		s.queryIds(), "only stakeibc's haqq slash-path queries are deleted")

	haqq, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.HaqqChainId)
	for _, validator := range haqq.Validators {
		s.Require().False(validator.SlashQueryInProgress, "%s flag cleared", validator.Address)
	}
	juno, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "juno-1")
	s.Require().True(juno.Validators[0].SlashQueryInProgress, "other zones' flags untouched")
}

func (s *UpgradeTestSuite) TestPurgeHaqqSlashQueries_NoZone() {
	s.seedQueries()
	s.Require().NotPanics(func() { v35.PurgeHaqqSlashQueries(s.Ctx, s.App.InterchainqueryKeeper, s.App.StakeibcKeeper) })
	s.Require().Len(s.queryIds(), 7, "queries are still purged when the zone is missing")
}

func (s *UpgradeTestSuite) TestPurgeWithdrawalBalanceQueries() {
	s.seedQueries()

	v35.PurgeWithdrawalBalanceQueries(s.Ctx, s.App.InterchainqueryKeeper)

	s.Require().ElementsMatch(
		[]string{
			"haqq-delegation", "haqq-validator", "haqq-calibrate", "haqq-fee", "juno-delegation", "comdex-calibrate",
			"other-haqq-delegation", "other-juno-withdrawal",
		},
		s.queryIds(), "only stakeibc's withdrawal-balance queries are deleted, on every chain")
}
