package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	_ "github.com/stretchr/testify/suite"

	recordstypes "github.com/Stride-Labs/stride/v21/x/records/types"
	recordtypes "github.com/Stride-Labs/stride/v21/x/records/types"

	"github.com/Stride-Labs/stride/v21/x/stakeibc/types"
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

// TODO [cleanup]: Combine this all into one test
type CleanupEpochUnbondingRecordsTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []types.HostZone
}

func (s *KeeperTestSuite) SetupCleanupEpochUnbondingRecords() CleanupEpochUnbondingRecordsTestCase {
	hostZones := []types.HostZone{
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
