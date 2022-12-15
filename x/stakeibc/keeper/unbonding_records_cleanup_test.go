package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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
					NativeTokenAmount: sdk.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdk.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdk.ZeroInt(),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdk.NewInt(1_000_000),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
			},
		},
		{
			EpochNumber: 2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdk.ZeroInt(),
					Status:            recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				},
				{
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdk.ZeroInt(),
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

func (s *KeeperTestSuite) CleanupEpochUnbondingRecords_Successful() {
	// successfully clean up epoch unbonding records
	tc := s.SetupGetHostZoneUnbondingMsgs()
	// clean up epoch unbonding record 0
	success := s.App.StakeibcKeeper.CleanupEpochUnbondingRecords(s.Ctx, 0)
	s.Require().True(success, "cleanup unbonding records returns true")
	epochUnbondings := tc.epochUnbondingRecords
	s.Require().Len(epochUnbondings, 1, "only one epoch unbonding record should be left")
	epochUnbonding := epochUnbondings[0]
	s.Require().Equal(1, epochUnbonding.EpochNumber, "correct unbonding record remains unprocessed")
}
