package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	epochtypes "github.com/Stride-Labs/stride/v22/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v22/x/records/types"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
)

type SweepUnbondedTokensTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []types.HostZone
	delegationChannelID   string
	delegationPortID      string
	channelStartSequence  uint64
}

func (s *KeeperTestSuite) SetupSweepUnbondedTokens() SweepUnbondedTokensTestCase {
	delegationChannelId, delegationPortId := s.CreateICAChannel("GAIA.DELEGATION")

	hostZones := []types.HostZone{
		{
			ChainId:              HostChainId,
			HostDenom:            Atom,
			UnbondingPeriod:      14,
			DelegationIcaAddress: "cosmos_DELEGATION",
			RedemptionIcaAddress: "cosmos_REDEMPTION",
			ConnectionId:         ibctesting.FirstConnectionID,
		},
		{
			ChainId:              OsmoChainId,
			HostDenom:            Osmo,
			UnbondingPeriod:      21,
			DelegationIcaAddress: "osmo_DELEGATION",
			RedemptionIcaAddress: "osmo_REDEMPTION",
			ConnectionId:         "connection-2",
		},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	dayEpochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.DAY_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, dayEpochTracker)

	unbondingTime := uint64(s.Ctx.BlockTime().Add(-1 * time.Minute).UnixNano())
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(1_000_000),
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
					NativeTokenAmount: sdkmath.NewInt(3_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(4_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
					UnbondingTime:     unbondingTime,
				},
			},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	// Get the sequence number before sweep ICAs are sent to confirm it increments after the ICA
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, delegationPortId, delegationChannelId)
	s.Require().True(found, "sequence number not found before transfer")

	return SweepUnbondedTokensTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
		delegationChannelID:   delegationChannelId,
		delegationPortID:      delegationPortId,
		channelStartSequence:  startSequence,
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

func (s *KeeperTestSuite) TestSweepUnbondedTokensForHostZone_Successful() {
	tc := s.SetupSweepUnbondedTokens()
	hostZone := tc.hostZones[0]

	// Call redemption sweep
	err := s.App.StakeibcKeeper.SweepUnbondedTokensForHostZone(s.Ctx, hostZone)
	s.Require().NoError(err, "no error expected when sweeping")

	// Confirm ICA was submitted (by checking sequence number was incremented)
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, tc.delegationPortID, tc.delegationChannelID)
	s.Require().True(found, "sequence number not found after after redemption ICA")
	s.Require().Equal(tc.channelStartSequence+1, endSequence, "tx sequence number after redemption ICA")

	// Confirm callback data was stored
	allCallbackData := s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx)
	s.Require().Len(allCallbackData, 1, "length of callback data")

	redemptionCallback, err := s.App.StakeibcKeeper.UnmarshalRedemptionCallbackArgs(s.Ctx, allCallbackData[0].CallbackArgs)
	s.Require().NoError(err, "no error expected when unmarshaling redemption callback")

	s.Require().Equal(HostChainId, redemptionCallback.HostZoneId, "callback chain ID")
	s.Require().Equal([]uint64{1, 2}, redemptionCallback.EpochUnbondingRecordIds, "callback epoch unbonding IDs")

	// Confirm epoch unbonding record status was updated
	epochUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		for _, hostZoneUnbondingRecord := range epochUnbondingRecord.HostZoneUnbondings {
			expectedStatus := recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE.String()
			if hostZoneUnbondingRecord.HostZoneId == HostChainId {
				expectedStatus = recordtypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS.String()
			}
			s.Require().Equal(expectedStatus, hostZoneUnbondingRecord.Status.String(),
				"epoch unbonding record status for record %d and host zone %s",
				epochUnbondingRecord.EpochNumber, hostZoneUnbondingRecord.HostZoneId)
		}
	}

	// Confirm sweep amount was correct
	s.CheckEventValueEmitted(types.EventTypeRedemptionSweep, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeRedemptionSweep, types.AttributeKeySweptAmount, "4000000")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokensForHostZone_MissingDelegationAccount() {
	tc := s.SetupSweepUnbondedTokens()
	hostZone := tc.hostZones[0]

	// Remove the delegation account from the host chain, it should cause the redemption to fail
	hostZone.DelegationIcaAddress = ""
	err := s.App.StakeibcKeeper.SweepUnbondedTokensForHostZone(s.Ctx, hostZone)
	s.Require().ErrorContains(err, "no delegation account found")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokensForHostZone_MissingRedemptionAccount() {
	tc := s.SetupSweepUnbondedTokens()
	hostZone := tc.hostZones[0]

	// Remove the redemption account from the host chain, it should cause the redemption to fail
	hostZone.DelegationIcaAddress = ""
	err := s.App.StakeibcKeeper.SweepUnbondedTokensForHostZone(s.Ctx, hostZone)
	s.Require().ErrorContains(err, "no redemption account found")
}

func (s *KeeperTestSuite) TestSweepUnbondedTokensForHostZone_FailedToGetLightClientTime() {
	tc := s.SetupSweepUnbondedTokens()
	hostZone := tc.hostZones[0]

	// Change the connection ID on the host zone so that the light client time cannot be found
	// It should cause the redemption to fail
	hostZone.ConnectionId = "invalid-connection-id"
	err := s.App.StakeibcKeeper.SweepUnbondedTokensForHostZone(s.Ctx, hostZone)
	s.Require().ErrorContains(err, "could not get light client block time for host zone")
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
