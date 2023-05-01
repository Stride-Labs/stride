package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type InitiateAllHostZoneUnbondingsTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
}

func (s *KeeperTestSuite) SetupInitiateAllHostZoneUnbondings() InitiateAllHostZoneUnbondingsTestCase {
	s.CreateICAChannel("GAIA.DELEGATION")

	gaiaValAddr := "cosmos_VALIDATOR"
	osmoValAddr := "osmo_VALIDATOR"
	gaiaDelegationAddr := "cosmos_DELEGATION"
	osmoDelegationAddr := "osmo_DELEGATION"
	//  define the host zone with total delegation and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:    gaiaValAddr,
			Delegation: sdkmath.NewInt(5_000_000),
			Weight:     uint64(10),
		},
		{
			Address:    gaiaValAddr + "2",
			Delegation: sdkmath.NewInt(3_000_000),
			Weight:     uint64(6),
		},
	}
	osmoValidators := []*stakeibc.Validator{
		{
			Address:    osmoValAddr,
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
			DelegationIcaAddress: gaiaDelegationAddr,
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
		},
		{
			ChainId:              OsmoChainId,
			HostDenom:            Osmo,
			Bech32Prefix:         OsmoPrefix,
			UnbondingPeriod:      21,
			Validators:           osmoValidators,
			DelegationIcaAddress: osmoDelegationAddr,
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
		},
	}
	// list of epoch unbonding records
	default_unbonding := []*recordtypes.HostZoneUnbonding{
		{
			HostZoneId:        HostChainId,
			StTokenAmount:     sdkmath.NewInt(1_900_000),
			NativeTokenAmount: sdkmath.NewInt(2_000_000),
			Denom:             Atom,
			Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		},
		{
			HostZoneId:        OsmoChainId,
			StTokenAmount:     sdkmath.NewInt(2_800_000),
			NativeTokenAmount: sdkmath.NewInt(3),
			Denom:             Osmo,
			Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		},
	}
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{}
	for _, epochNumber := range []uint64{5} {
		epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
			EpochNumber:        epochNumber,
			HostZoneUnbondings: default_unbonding,
		}
		epochUnbondingRecords = append(epochUnbondingRecords, epochUnbondingRecord)
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibc.EpochTracker{
		EpochIdentifier:    "day",
		EpochNumber:        12,
		NextEpochStartTime: uint64(2661750006000000000), // arbitrary time in the future, year 2056 I believe
		Duration:           uint64(1000000000000),       // 16 min 40 sec
	})

	return InitiateAllHostZoneUnbondingsTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
	}
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Successful() {
	// tests that we can successful initiate a host zone unbonding for ATOM and OSMO
	s.SetupInitiateAllHostZoneUnbondings()
	success, successfulUnbondings, failedUnbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)
	s.Require().True(success, "initiating unbondings returns true")
	s.Require().Len(successfulUnbondings, 2, "initiating unbondings returns 2 successful unbondings")
	s.Require().Len(failedUnbondings, 0, "initiating unbondings returns 0 failed unbondings")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_GaiaSuccessful() {
	// Tests that if we initiate unbondings a day where only Gaia is supposed to unbond, it succeeds and Osmo is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	success, successfulUnbondings, failedUnbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 9)
	s.Require().True(success, "initiating gaia unbondings returns true")
	s.Require().Len(successfulUnbondings, 1, "initiating gaia unbondings returns 1 successful unbondings")
	s.Require().Len(failedUnbondings, 0, "initiating gaia unbondings returns 0 failed unbondings")
	s.Require().Equal("GAIA", successfulUnbondings[0], "initiating gaia unbondings returns gaia")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_OsmoSuccessful() {
	// Tests that if we initiate unbondings a day where only Osmo is supposed to unbond, it succeeds and Gaia is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	success, successfulUnbondings, failedUnbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 8)
	s.Require().True(success, "initiating osmo unbondings returns true")
	s.Require().Len(successfulUnbondings, 1, "initiating osmo unbondings returns 1 successful unbondings")
	s.Require().Len(failedUnbondings, 0, "initiating osmo unbondings returns 0 failed unbondings")
	s.Require().Equal("OSMO", successfulUnbondings[0], "initiating osmo unbondings returns gaia")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_NoneSuccessful() {
	// Tests that if we initiate unbondings a day where none are supposed to unbond, it works successfully
	s.SetupInitiateAllHostZoneUnbondings()
	success, successfulUnbondings, failedUnbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 10)
	s.Require().True(success, "initiating no unbondings returns true")
	s.Require().Len(successfulUnbondings, 0, "initiating no unbondings returns 0 successful unbondings")
	s.Require().Len(failedUnbondings, 0, "initiating no unbondings returns 0 failed unbondings")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Failed() {
	// Tests that if Gaia doesn't have enough delegated stake to unbond, it fails
	// but Osmo does and is successful
	s.SetupInitiateAllHostZoneUnbondings()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	hostZone.Validators = []*stakeibc.Validator{
		{
			Address:    "cosmos_VALIDATOR",
			Delegation: sdkmath.NewInt(1_000_000),
			Weight:     uint64(10),
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	success, successfulUnbondings, failedUnbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)
	s.Require().False(success, "initiating bad unbondings returns false")
	s.Require().Len(successfulUnbondings, 1, "initiating bad unbondings has 1 success")
	s.Require().Len(failedUnbondings, 1, "initiating bad unbondings has 1 failure")
	s.Require().Equal("OSMO", successfulUnbondings[0], "initiating bad unbondings succeeds on osmo")
	s.Require().Equal("GAIA", failedUnbondings[0], "initiating bad unbondings fails on gaia")
}

func (s *KeeperTestSuite) TestGetTotalUnbondAmountAndRecordsIds() {
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber: uint64(1),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(1),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(2),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
		{
			EpochNumber: uint64(2),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(3),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
				{
					// Different host zone
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(4),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
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
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
				{
					// Different Status
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(6),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
				},
			},
		},
		{
			EpochNumber: uint64(4),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				{
					// Different Host and Status
					HostZoneId:        OsmoChainId,
					NativeTokenAmount: sdkmath.NewInt(7),
					Status:            recordtypes.HostZoneUnbonding_CLAIMABLE,
				},
				{
					// Summed
					HostZoneId:        HostChainId,
					NativeTokenAmount: sdkmath.NewInt(8),
					Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				},
			},
		},
	}

	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	expectedUnbondAmount := int64(1 + 3 + 8)
	expectedRecordIds := []uint64{1, 2, 4}

	actualUnbondAmount, actualRecordIds := s.App.StakeibcKeeper.GetTotalUnbondAmountAndRecordsIds(s.Ctx, HostChainId)
	s.Require().Equal(expectedUnbondAmount, actualUnbondAmount.Int64(), "unbonded amount")
	s.Require().Equal(expectedRecordIds, actualRecordIds, "epoch unbonding record IDs")
}
