package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"

	epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type SweepUnbondedTokensTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
	lightClientTime       uint64
}

func (s *KeeperTestSuite) SetupSweepUnbondedTokens() SweepUnbondedTokensTestCase {
	s.CreateICAChannel("GAIA.DELEGATION")
	//  define the host zone with stakedBal and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: sdk.NewInt(5_000_000),
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
			DelegationAmt: sdk.NewInt(5_000_000),
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
			ChainId:            HostChainId,
			HostDenom:          Atom,
			Bech32Prefix:       GaiaPrefix,
			UnbondingFrequency: 3,
			Validators:         gaiaValidators,
			DelegationAccount:  &gaiaDelegationAccount,
			RedemptionAccount:  &gaiaRedemptionAccount,
			StakedBal:          sdk.NewInt(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
		{
			ChainId:            OsmoChainId,
			HostDenom:          Osmo,
			Bech32Prefix:       OsmoPrefix,
			UnbondingFrequency: 4,
			Validators:         osmoValidators,
			DelegationAccount:  &osmoDelegationAccount,
			RedemptionAccount:  &osmoRedemptionAccount,
			StakedBal:          sdk.NewInt(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
	}
	dayEpochTracker := stakeibc.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(10)
	lightClientTime := unbondingTime + 1
	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdk.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdk.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdk.NewInt(2_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdk.NewInt(2_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdk.NewInt(5_000_000),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdk.NewInt(5_000_000),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, dayEpochTracker)

	return SweepUnbondedTokensTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
		lightClientTime:       lightClientTime,
	}
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_Successful() {
	s.SetupSweepUnbondedTokens()
	success, successfulSweeps, sweepAmounts, failedSweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx)
	s.Require().True(success, "sweep all tokens succeeds")
	s.Require().Len(successfulSweeps, 2, "sweep all tokens succeeds for 2 host zones")
	s.Require().Len(sweepAmounts, 2, "sweep all tokens succeeds for 2 host zones")
	s.Require().Len(failedSweeps, 0, "sweep all tokens fails for no host zone")
	s.Require().Equal([]sdk.Int{sdk.NewInt(2_000_000), sdk.NewInt(3_000_000)}, sweepAmounts, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_HostZoneUnbondingMissing() {
	// If Osmo is missing, make sure that the function still succeeds
	s.SetupSweepUnbondedTokens()
	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	for _, epochUnbonding := range epochUnbondingRecords {
		epochUnbonding.HostZoneUnbondings = []*recordtypes.HostZoneUnbonding{
			epochUnbonding.HostZoneUnbondings[0],
		}
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbonding)
	}
	success, successfulSweeps, sweepAmounts, failedSweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx)
	s.Require().True(success, "sweep all tokens succeeded if osmo missing")
	s.Require().Len(successfulSweeps, 2, "sweep all tokens succeeds for 2 host zones")
	s.Require().Len(sweepAmounts, 2, "sweep all tokens succeeds for 2 host zone")
	s.Require().Len(failedSweeps, 0, "sweep all tokens fails for 0 host zone")
	s.Require().Equal([]sdk.Int{sdk.NewInt(2_000_000), sdk.ZeroInt()}, sweepAmounts, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_RedemptionAccountMissing() {
	s.SetupSweepUnbondedTokens()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	hostZone.RedemptionAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	success, successfulSweeps, sweepAmounts, failedSweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx)
	s.Require().Equal(success, false, "sweep all tokens failed if osmo missing")
	s.Require().Len(successfulSweeps, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal("OSMO", successfulSweeps[0], "sweep all tokens succeeds for osmo")
	s.Require().Len(sweepAmounts, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Len(failedSweeps, 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal("GAIA", failedSweeps[0], "sweep all tokens fails for gaia")
	s.Require().Equal([]sdk.Int{sdk.NewInt(3_000_000)}, sweepAmounts, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_DelegationAccountAddressMissing() {
	s.SetupSweepUnbondedTokens()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "OSMO")
	hostZone.DelegationAccount.Address = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	success, successfulSweeps, sweepAmounts, failedSweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx)
	s.Require().False(success, "sweep all tokens failed if gaia missing")
	s.Require().Len(successfulSweeps, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal("GAIA", successfulSweeps[0], "sweep all tokens succeeds for gaia")
	s.Require().Len(sweepAmounts, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Len(failedSweeps, 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal("OSMO", failedSweeps[0], "sweep all tokens fails for osmo")
	s.Require().Equal([]sdk.Int{sdk.NewInt(2_000_000)}, sweepAmounts, "correct amount of tokens swept for each host zone")
}
