package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v18/x/records/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

func (s *KeeperTestSuite) SetupInitiateAllHostZoneUnbondings() {
	s.CreateICAChannel("GAIA.DELEGATION")

	gaiaValAddr := "cosmos_VALIDATOR"
	osmoValAddr := "osmo_VALIDATOR"
	gaiaDelegationAddr := "cosmos_DELEGATION"
	osmoDelegationAddr := "osmo_DELEGATION"

	//  define the host zone with total delegation and validators with staked amounts
	gaiaValidators := []*types.Validator{
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
	osmoValidators := []*types.Validator{
		{
			Address:    osmoValAddr,
			Delegation: sdkmath.NewInt(5_000_000),
			Weight:     uint64(10),
		},
	}
	hostZones := []types.HostZone{
		{
			ChainId:              HostChainId,
			HostDenom:            Atom,
			Bech32Prefix:         GaiaPrefix,
			UnbondingPeriod:      14,
			Validators:           gaiaValidators,
			DelegationIcaAddress: gaiaDelegationAddr,
			TotalDelegations:     sdkmath.NewInt(5_000_000),
			ConnectionId:         ibctesting.FirstConnectionID,
			RedemptionRate:       sdk.OneDec(),
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
			RedemptionRate:       sdk.OneDec(),
		},
	}
	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// list of epoch unbonding records
	epochNumber := uint64(5)

	redemptionRecordId1 := recordtypes.UserRedemptionRecordKeyFormatter(HostChainId, epochNumber, "receiver")
	redemptionRecordId2 := recordtypes.UserRedemptionRecordKeyFormatter(OsmoChainId, epochNumber, "receiver")

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:            HostChainId,
				StTokenAmount:         sdkmath.NewInt(1_900_000),
				NativeTokenAmount:     sdkmath.NewInt(2_000_000),
				Denom:                 Atom,
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				UserRedemptionRecords: []string{redemptionRecordId1},
			},
			{
				HostZoneId:            OsmoChainId,
				StTokenAmount:         sdkmath.NewInt(2_800_000),
				NativeTokenAmount:     sdkmath.NewInt(3),
				Denom:                 Osmo,
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				UserRedemptionRecords: []string{redemptionRecordId2},
			},
		},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:            redemptionRecordId1,
		HostZoneId:    HostChainId,
		EpochNumber:   epochNumber,
		StTokenAmount: epochUnbondingRecord.HostZoneUnbondings[0].StTokenAmount,
	})

	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:            redemptionRecordId2,
		HostZoneId:    OsmoChainId,
		EpochNumber:   epochNumber,
		StTokenAmount: epochUnbondingRecord.HostZoneUnbondings[1].StTokenAmount,
	})

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, types.EpochTracker{
		EpochIdentifier:    "day",
		EpochNumber:        12,
		NextEpochStartTime: uint64(2661750006000000000), // arbitrary time in the future, year 2056 I believe
		Duration:           uint64(1000000000000),       // 16 min 40 sec
	})
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Successful() {
	// tests that we can successful initiate a host zone unbonding for GAIA and OSMO
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)

	// An event should be emitted for each if they were successful
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_GaiaSuccessful() {
	// Tests that if we initiate unbondings a day where only Gaia is supposed to unbond, it succeeds and Osmo is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 9)

	// An event should only be emitted for Gaia
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_OsmoSuccessful() {
	// Tests that if we initiate unbondings a day where only Osmo is supposed to unbond, it succeeds and Gaia is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 8)

	// An event should only be emitted for Osmo
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_NoneSuccessful() {
	// Tests that if we initiate unbondings a day where none are supposed to unbond, it works successfully
	s.SetupInitiateAllHostZoneUnbondings()
	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 10)

	// No event should be emitted for either host
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Failed() {
	// Tests that if Gaia doesn't have enough delegated stake to unbond, it fails
	// but Osmo does and is successful
	s.SetupInitiateAllHostZoneUnbondings()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	hostZone.Validators = []*types.Validator{
		{
			Address:    "cosmos_VALIDATOR",
			Delegation: sdkmath.NewInt(1_000_000),
			Weight:     uint64(10),
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)

	s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)

	// An event should only be emitted for Osmo
	s.CheckEventValueNotEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, HostChainId)
	s.CheckEventValueEmitted(types.EventTypeUndelegation, types.AttributeKeyHostZone, OsmoChainId)

}
