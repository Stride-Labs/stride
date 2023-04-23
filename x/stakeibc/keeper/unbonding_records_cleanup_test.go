package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type CleanupEpochUnbondingRecordsTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
}

func (s *KeeperTestSuite) SetupCleanupEpochUnbondingRecords() CleanupEpochUnbondingRecordsTestCase {
	hostZones := []stakeibc.HostZone{
		{
			ChainId:      HostChainId,
			HostDenom:    Atom,
			Bech32Prefix: GaiaPrefix,
		},
		{
			ChainId:      OsmoChainId,
			HostDenom:    Osmo,
			Bech32Prefix: OsmoPrefix,
		},
	}
	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.ZeroInt(),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.ZeroInt(),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.ZeroInt(),
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

	return CleanupEpochUnbondingRecordsTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
	}
}

func (s *KeeperTestSuite) TestCleanupEpochUnbondingRecords_Successful() {
	tc := s.SetupCleanupEpochUnbondingRecords()

	// Call cleanup on each unbonding record
	for i := range tc.epochUnbondingRecords {
		success := s.App.StakeibcKeeper.CleanupEpochUnbondingRecords(s.Ctx, uint64(i))
		s.Require().True(success, "cleanup unbonding record for epoch %d should succeed", i)
	}

	// Check one record was removed
	finalUnbondingRecords := s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)
	expectedNumUnbondingRecords := len(tc.epochUnbondingRecords) - 1
	s.Require().Len(finalUnbondingRecords, expectedNumUnbondingRecords, "two epoch unbonding records should remain")

	// Confirm it was the last record that was removed
	_, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, uint64(2))
	s.Require().False(found, "removed record should not be found")
}
