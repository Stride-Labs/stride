package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v22/x/records/types"

	epochtypes "github.com/Stride-Labs/stride/v22/x/epochs/types"
	stakeibc "github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

type SweepUnbondedTokensTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
	lightClientTime       uint64
}

func (s *KeeperTestSuite) SetupSweepUnbondedTokens() SweepUnbondedTokensTestCase {
	s.CreateICAChannel("GAIA.DELEGATION")
	//  define the host zone with TotalDelegations and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:    "cosmos_VALIDATOR",
			Delegation: sdkmath.NewInt(5_000_000),
			Weight:     uint64(10),
		},
	}
	osmoValidators := []*stakeibc.Validator{
		{
			Address:    "osmo_VALIDATOR",
			Delegation: sdkmath.NewInt(5_000_000),
			Weight:     uint64(10),
		},
	}
	hostZones := []stakeibc.HostZone{
		{
			ChainId:              HostChainId,
			HostDenom:            Atom,
			Bech32Prefix:         GaiaPrefix,
			UnbondingPeriod:      14,
			Validators:           gaiaValidators,
			DelegationIcaAddress: "cosmos_DELEGATION",
			RedemptionIcaAddress: "cosmos_REDEMPTION",
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
		},
		{
			ChainId:              OsmoChainId,
			HostDenom:            Osmo,
			Bech32Prefix:         OsmoPrefix,
			UnbondingPeriod:      21,
			Validators:           osmoValidators,
			DelegationIcaAddress: "osmo_DELEGATION",
			RedemptionIcaAddress: "osmo_REDEMPTION",
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
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
					NativeTokenAmount: sdkmath.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(1_000_000),
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
					NativeTokenAmount: sdkmath.NewInt(2_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(2_000_000),
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
					NativeTokenAmount: sdkmath.NewInt(5_000_000),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(5_000_000),
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
	s.Require().Equal([]sdkmath.Int{sdkmath.NewInt(2_000_000), sdkmath.NewInt(3_000_000)}, sweepAmounts, "correct amount of tokens swept for each host zone")
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
	s.Require().Equal([]sdkmath.Int{sdkmath.NewInt(2_000_000), sdkmath.ZeroInt()}, sweepAmounts, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_RedemptionAccountMissing() {
	s.SetupSweepUnbondedTokens()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	hostZone.RedemptionIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	success, successfulSweeps, sweepAmounts, failedSweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx)
	s.Require().Equal(success, false, "sweep all tokens failed if osmo missing")
	s.Require().Len(successfulSweeps, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal("OSMO", successfulSweeps[0], "sweep all tokens succeeds for osmo")
	s.Require().Len(sweepAmounts, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Len(failedSweeps, 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal("GAIA", failedSweeps[0], "sweep all tokens fails for gaia")
	s.Require().Equal([]sdkmath.Int{sdkmath.NewInt(3_000_000)}, sweepAmounts, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokens_DelegationAccountAddressMissing() {
	s.SetupSweepUnbondedTokens()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "OSMO")
	hostZone.DelegationIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	success, successfulSweeps, sweepAmounts, failedSweeps := s.App.StakeibcKeeper.SweepAllUnbondedTokens(s.Ctx)
	s.Require().False(success, "sweep all tokens failed if gaia missing")
	s.Require().Len(successfulSweeps, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Equal("GAIA", successfulSweeps[0], "sweep all tokens succeeds for gaia")
	s.Require().Len(sweepAmounts, 1, "sweep all tokens succeeds for 1 host zone")
	s.Require().Len(failedSweeps, 1, "sweep all tokens fails for 1 host zone")
	s.Require().Equal("OSMO", failedSweeps[0], "sweep all tokens fails for osmo")
	s.Require().Equal([]sdkmath.Int{sdkmath.NewInt(2_000_000)}, sweepAmounts, "correct amount of tokens swept for each host zone")
}

func (s *KeeperTestSuite) TestGetTotalRedemptionSweepAmountAndRecordsIds() {
	hostBlockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	validUnbondTime := uint64(hostBlockTime.Add(-1 * time.Minute).UnixNano())

	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: uint64(1),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(1),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     validUnbondTime,
				},
				{
					// Different host zone
					HostZoneId:        "different",
					NativeTokenAmount: sdkmath.NewInt(2),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     validUnbondTime,
				},
			},
		},
		{
			EpochNumber: uint64(2),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different host zone
					HostZoneId:        "different",
					NativeTokenAmount: sdkmath.NewInt(3),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     validUnbondTime,
				},
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(4),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     validUnbondTime,
				},
			},
		},
		{
			EpochNumber: uint64(3),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Status
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(5),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
					UnbondingTime:     validUnbondTime,
				},
			},
		},
		{
			EpochNumber: uint64(4),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Unbonding time not set
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(6),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     0,
				},
			},
		},
		{
			EpochNumber: uint64(5),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Unbonding time after block time
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(7),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     uint64(hostBlockTime.Add(time.Minute).UnixNano()),
				},
			},
		},
		{
			EpochNumber: uint64(6),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(8),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     validUnbondTime,
				},
			},
		},
	}

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	expectedUnbondAmount := int64(1 + 4 + 8)
	expectedRecordIds := []uint64{1, 2, 6}

	hostBlockTimeNano := uint64(hostBlockTime.UnixNano())
	actualUnbondAmount, actualRecordIds := s.App.StakeibcKeeper.GetTotalRedemptionSweepAmountAndRecordIds(s.Ctx, HostChainId, hostBlockTimeNano)
	s.Require().Equal(expectedUnbondAmount, actualUnbondAmount.Int64(), "unbonded amount")
	s.Require().Equal(expectedRecordIds, actualRecordIds, "epoch unbonding record IDs")
}
