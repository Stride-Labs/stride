package keeper_test

import (
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type CleanupEpochUnbondingRecordsTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
}

func (s *KeeperTestSuite) SetupCleanupEpochUnbondingRecords() CleanupEpochUnbondingRecordsTestCase {
	hostZones := []stakeibc.HostZone{
		{
			ChainId:      "GAIA",
			HostDenom:    "uatom",
			Bech32Prefix: "cosmos",
		},
		{
			ChainId:      "OSMO",
			HostDenom:    "uosmo",
			Bech32Prefix: "osmo",
		},
	}
	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: 0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        "GAIA",
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_TRANSFERRED,
				},
				{
					HostZoneId:        "OSMO",
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_UNBONDED,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        "GAIA",
					NativeTokenAmount: 0,
					Status:            recordtypes.HostZoneUnbonding_UNBONDED,
				},
				{
					HostZoneId:        "OSMO",
					NativeTokenAmount: 1_000_000,
					Status:            recordtypes.HostZoneUnbonding_TRANSFERRED,
				},
			},
		},
		{
			EpochNumber: 1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					HostZoneId:        "GAIA",
					NativeTokenAmount: 0,
					Status:            recordtypes.HostZoneUnbonding_UNBONDED,
				},
				{
					HostZoneId:        "OSMO",
					NativeTokenAmount: 0,
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

	return CleanupEpochUnbondingRecordsTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
	}
}

func (s *KeeperTestSuite) SetupCleanupEpochUnbondingRecords_Successful() {
	tc := s.SetupSendHostZoneUnbonding()
	success := s.App.StakeibcKeeper.CleanupEpochUnbondingRecords(s.Ctx())
	s.Require().Equal(success, true, "cleanup unbonding records returns true")
	epochUnbondings := tc.epochUnbondingRecords
	s.Require().Equal(len(epochUnbondings), 1, "only one epoch unbonding record should be left")
	epochUnbonding := epochUnbondings[0]
	s.Require().Equal(epochUnbonding.EpochNumber, 1, "correct unbonding record remains unprocessed")
}
