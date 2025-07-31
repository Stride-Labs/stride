package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	_ "github.com/stretchr/testify/suite"

	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v27/x/records/types"

	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestCreateDepositRecordsForEpoch_Successful() {
	// Set host zones
	hostZones := []types.HostZone{
		{
			ChainId:   "HOST1",
			HostDenom: "denom1",
		},
		{
			ChainId:   "HOST2",
			HostDenom: "denom2",
		},
		{
			ChainId:   "HOST3",
			HostDenom: "denom3",
		},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Create depoist records for two epochs
	s.App.StakeibcKeeper.CreateDepositRecordsForEpoch(s.Ctx, 1)
	s.App.StakeibcKeeper.CreateDepositRecordsForEpoch(s.Ctx, 2)

	expectedDepositRecords := []recordstypes.DepositRecord{
		// Epoch 1
		{
			Id:                 0,
			Amount:             sdkmath.ZeroInt(),
			Denom:              "denom1",
			HostZoneId:         "HOST1",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 1,
		},
		{
			Id:                 1,
			Amount:             sdkmath.ZeroInt(),
			Denom:              "denom2",
			HostZoneId:         "HOST2",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 1,
		},
		{
			Id:                 2,
			Amount:             sdkmath.ZeroInt(),
			Denom:              "denom3",
			HostZoneId:         "HOST3",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 1,
		},
		// Epoch 2
		{
			Id:                 3,
			Amount:             sdkmath.ZeroInt(),
			Denom:              "denom1",
			HostZoneId:         "HOST1",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 2,
		},
		{
			Id:                 4,
			Amount:             sdkmath.ZeroInt(),
			Denom:              "denom2",
			HostZoneId:         "HOST2",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 2,
		},
		{
			Id:                 5,
			Amount:             sdkmath.ZeroInt(),
			Denom:              "denom3",
			HostZoneId:         "HOST3",
			Status:             recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber: 2,
		},
	}

	// Confirm deposit records
	actualDepositRecords := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Equal(len(expectedDepositRecords), len(actualDepositRecords), "number of deposit records")
	s.Require().Equal(expectedDepositRecords, actualDepositRecords, "deposit records")
}

func (s *KeeperTestSuite) TestCleanupEpochUnbondingRecords() {
	// Epoch unbonding records with different amounts and statuses
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			// Has a non-CLAIMABLE record, should not be removed
			EpochNumber: 0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:            HostChainId,
					ClaimableNativeTokens: sdkmath.ZeroInt(),
					Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:            OsmoChainId,
					ClaimableNativeTokens: sdkmath.NewInt(1_000_000),
					Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				},
			},
		},
		{
			// Has a non-zero CLAIMABLE record, should not be removed
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:            HostChainId,
					ClaimableNativeTokens: sdkmath.NewInt(1_000_000),
					Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:            OsmoChainId,
					ClaimableNativeTokens: sdkmath.ZeroInt(),
					Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
			},
		},
		{
			// Has only CLAIMABLE and zero-amounts - should be removed
			EpochNumber: 2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:            HostChainId,
					ClaimableNativeTokens: sdkmath.ZeroInt(),
					Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:            OsmoChainId,
					ClaimableNativeTokens: sdkmath.ZeroInt(),
					Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
			},
		},
		{
			// Has a non-CLAIMABLE record, should not be removed
			EpochNumber: 3,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:            HostChainId,
					ClaimableNativeTokens: sdkmath.ZeroInt(),
					Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				},
				{
					HostZoneId:            OsmoChainId,
					ClaimableNativeTokens: sdkmath.NewInt(1_000_000),
					Status:                recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
			},
		},
	}

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	// Call cleanup on each unbonding record
	for i := range epochUnbondingRecords {
		s.App.StakeibcKeeper.CleanupEpochUnbondingRecords(s.Ctx, uint64(i))
	}

	// Check one record was removed
	finalUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	expectedNumUnbondingRecords := len(epochUnbondingRecords) - 1
	s.Require().Len(finalUnbondingRecords, expectedNumUnbondingRecords, "two epoch unbonding records should remain")

	// Confirm it was the last record that was removed
	_, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(2))
	s.Require().False(found, "removed record should not be found")
}
