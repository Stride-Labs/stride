package keeper_test

import (
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type SweepUnbondedTokensTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
	lightClientTime       uint64
}

func (s *KeeperTestSuite) SetupSweepUnbondedTokens() SweepUnbondedTokensTestCase {
	s.CreateICAChannel("icacontroller-GAIA.DELEGATION")
	//  define the host zone with stakedBal and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: uint64(5_000_000),
			Weight:        uint64(10),
		},
	}
	gaiaDelegationAccount := stakeibc.ICAAccount{
		Address: "cosmos_DELEGATION",
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	gaiaRedemptionAccount := stakeibc.ICAAccount{
		Address: "cosmos_REDEMPTION",
		Target:  stakeibc.ICAAccountType_REDEMPTION,
	}
	osmoValidators := []*stakeibc.Validator{
		{
			Address:       "osmo_VALIDATOR",
			DelegationAmt: uint64(5_000_000),
			Weight:        uint64(10),
		},
	}
	osmoDelegationAccount := stakeibc.ICAAccount{
		Address: "osmo_DELEGATION",
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	osmoRedemptionAccount := stakeibc.ICAAccount{
		Address: "osmo_REDEMPTION",
		Target:  stakeibc.ICAAccountType_REDEMPTION,
	}
	hostZones := []stakeibc.HostZone{
		{
			ChainId:            "GAIA",
			HostDenom:          "uatom",
			Bech32Prefix:       "cosmos",
			UnbondingFrequency: 3,
			Validators:         gaiaValidators,
			DelegationAccount:  &gaiaDelegationAccount,
			RedemptionAccount:  &gaiaRedemptionAccount,
			StakedBal:          uint64(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
		{
			ChainId:            "OSMO",
			HostDenom:          "uosmo",
			Bech32Prefix:       "osmo",
			UnbondingFrequency: 4,
			Validators:         osmoValidators,
			DelegationAccount:  &osmoDelegationAccount,
			RedemptionAccount:  &osmoRedemptionAccount,
			StakedBal:          uint64(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
	}
	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(1660348276)
	lightClientTime := unbondingTime + 1
	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        "GAIA",
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_BONDED,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        "OSMO",
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_UNBONDED,
					UnbondingTime:     unbondingTime,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        "GAIA",
					NativeTokenAmount: 2_000_000,
					Status:            recordtypes.HostZoneUnbonding_UNBONDED,
				},
				{
					HostZoneId:        "OSMO",
					NativeTokenAmount: 2_000_000,
					Status:            recordtypes.HostZoneUnbonding_UNBONDED,
				},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        "GAIA",
					NativeTokenAmount: 5_000_000,
					Status:            recordtypes.HostZoneUnbonding_TRANSFERRED,
				},
				{
					HostZoneId:        "OSMO",
					NativeTokenAmount: 5_000_000,
					Status:            recordtypes.HostZoneUnbonding_BONDED,
				},
			},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	}

	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	}

	return SweepUnbondedTokensTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
		lightClientTime:       lightClientTime,
	}
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_Successful() {
	tc := s.SetupSweepUnbondedTokens()
	_ = tc
	success, successful_sweeps, sweep_amounts, failed_sweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx())
	s.Require().Equal(success, true, "sweep all tokens succeeds")
	s.Require().Equal(len(successful_sweeps), 2, "sweep all tokens succeeds for 2 host zones")
	s.Require().Equal(len(sweep_amounts), 2, "sweep all tokens succeeds for 2 host zones")
	s.Require().Equal(len(failed_sweeps), 0, "sweep all tokens fails for no host zone")
	s.Require().Equal(sweep_amounts, []int64{2_000_000, 3_000_000}, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_HostZoneUnbondingMissing() {
	tc := s.SetupSweepUnbondedTokens()
	_ = tc
	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx())
	for _, epochUnbonding := range epochUnbondingRecords {
		epochUnbonding.HostZoneUnbondings = []*recordtypes.HostZoneUnbonding{
			epochUnbonding.HostZoneUnbondings[0],
		}
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbonding)
	}
	success, successful_sweeps, sweep_amounts, failed_sweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx())
	s.Require().Equal(success, false, "sweep all tokens failed if osmo missing")
	s.Require().Equal(len(successful_sweeps), 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal(successful_sweeps[0], "GAIA", "sweep all tokens succeeds for gaia")
	s.Require().Equal(len(sweep_amounts), 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal(len(failed_sweeps), 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal(failed_sweeps[0], "OSMO", "sweep all tokens fails for osmo")
	s.Require().Equal(sweep_amounts, []int64{2_000_000}, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_RedemptionAccountMissing() {
	tc := s.SetupSweepUnbondedTokens()
	_ = tc
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), "GAIA")
	hostZone.RedemptionAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	success, successful_sweeps, sweep_amounts, failed_sweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx())
	s.Require().Equal(success, false, "sweep all tokens failed if osmo missing")
	s.Require().Equal(len(successful_sweeps), 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal(successful_sweeps[0], "OSMO", "sweep all tokens succeeds for osmo")
	s.Require().Equal(len(sweep_amounts), 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal(len(failed_sweeps), 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal(failed_sweeps[0], "GAIA", "sweep all tokens fails for gaia")
	s.Require().Equal(sweep_amounts, []int64{3_000_000}, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_DelegationAccountAddressMissing() {
	tc := s.SetupSweepUnbondedTokens()
	_ = tc
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), "OSMO")
	hostZone.DelegationAccount.Address = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	success, successful_sweeps, sweep_amounts, failed_sweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx())
	s.Require().Equal(success, false, "sweep all tokens failed if gaia missing")
	s.Require().Equal(len(successful_sweeps), 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal(successful_sweeps[0], "GAIA", "sweep all tokens succeeds for gaia")
	s.Require().Equal(len(sweep_amounts), 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal(len(failed_sweeps), 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal(failed_sweeps[0], "OSMO", "sweep all tokens fails for osmo")
	s.Require().Equal(sweep_amounts, []int64{2_000_000}, "correct amount of tokens swept for each host zone")
}
