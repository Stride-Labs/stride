package v35_test

import (
	sdkmath "cosmossdk.io/math"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	v35 "github.com/Stride-Labs/stride/v34/app/upgrades/v35"
	icqkeeper "github.com/Stride-Labs/stride/v34/x/interchainquery/keeper"
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

func (s *UpgradeTestSuite) TestUpgrade_PurgedCalibrationResponse() {
	s.seedQueries()
	s.App.InterchainqueryKeeper.SetQuery(s.Ctx, icqtypes.Query{
		Id: "other-comdex-calibrate", ChainId: v35.ComdexChainId,
		CallbackModule: "records", CallbackId: stakeibckeeper.ICQCallbackID_Calibrate,
	})
	hostZone := withInBoundsRates(stakeibctypes.HostZone{
		ChainId: v35.ComdexChainId, Deprecated: true,
		TotalDelegations: sdkmath.NewInt(100), Validators: []*stakeibctypes.Validator{{
			Address: "comdexvaloper1", Delegation: sdkmath.NewInt(100), SharesToTokensRate: sdkmath.LegacyOneDec(),
		}},
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.ConfirmUpgradeSucceeded(v35.UpgradeName)
	s.Require().ElementsMatch([]string{
		"haqq-fee", "juno-delegation", "other-haqq-delegation",
		"other-juno-withdrawal", "other-comdex-calibrate",
	}, s.queryIds(), "calibrations on every zone are purged without module collisions")
	before, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	response := s.App.AppCodec().MustMarshal(&stakingtypes.Delegation{ValidatorAddress: "comdexvaloper1", Shares: sdkmath.LegacyNewDec(200)})
	_, err := icqkeeper.NewMsgServerImpl(s.App.InterchainqueryKeeper).SubmitQueryResponse(s.Ctx,
		&icqtypes.MsgSubmitQueryResponse{QueryId: "comdex-calibrate", Result: response})
	s.Require().NoError(err, "late response for a purged query remains a successful no-op")
	after, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, v35.ComdexChainId)
	s.Require().True(found)
	s.Require().Equal(before, after, "purged calibration cannot change delegation accounting")
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
