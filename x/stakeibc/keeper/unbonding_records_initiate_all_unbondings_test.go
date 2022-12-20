package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
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
	//  define the host zone with stakedBal and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:       gaiaValAddr,
			DelegationAmt: sdk.NewInt(5_000_000),
			Weight:        uint64(10),
		},
		{
			Address:       gaiaValAddr + "2",
			DelegationAmt: sdk.NewInt(3_000_000),
			Weight:        uint64(6),
		},
	}
	gaiaDelegationAccount := stakeibc.ICAAccount{
		Address: gaiaDelegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	osmoValidators := []*stakeibc.Validator{
		{
			Address:       osmoValAddr,
			DelegationAmt: sdk.NewInt(5_000_000),
			Weight:        uint64(10),
		},
	}
	osmoDelegationAccount := stakeibc.ICAAccount{
		Address: osmoDelegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	hostZones := []stakeibc.HostZone{
		{
			ChainId:            HostChainId,
			HostDenom:          Atom,
			Bech32Prefix:       GaiaPrefix,
			UnbondingFrequency: 3,
			Validators:         gaiaValidators,
			DelegationAccount:  &gaiaDelegationAccount,
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
			StakedBal:          sdk.NewInt(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
	}
	// list of epoch unbonding records
	default_unbonding := []*recordtypes.HostZoneUnbonding{
		{
			HostZoneId:        HostChainId,
			StTokenAmount:     sdk.NewInt(1_900_000),
			NativeTokenAmount:  sdk.NewInt(2_000_000),
			Denom:             Atom,
			Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		},
		{
			HostZoneId:        OsmoChainId,
			StTokenAmount:      sdk.NewInt(2_800_000),
			NativeTokenAmount: sdk.NewInt(3),
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
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)
	s.Require().True(success, "initiating unbondings returns true")
	s.Require().Len(successful_unbondings, 2, "initiating unbondings returns 2 successful unbondings")
	s.Require().Len(failed_unbondings, 0, "initiating unbondings returns 0 failed unbondings")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_GaiaSuccessful() {
	// Tests that if we initiate unbondings a day where only Gaia is supposed to unbond, it succeeds and Osmo is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 9)
	s.Require().True(success, "initiating gaia unbondings returns true")
	s.Require().Len(successful_unbondings, 1, "initiating gaia unbondings returns 1 successful unbondings")
	s.Require().Len(failed_unbondings, 0, "initiating gaia unbondings returns 0 failed unbondings")
	s.Require().Equal("GAIA", successful_unbondings[0], "initiating gaia unbondings returns gaia")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_OsmoSuccessful() {
	// Tests that if we initiate unbondings a day where only Osmo is supposed to unbond, it succeeds and Gaia is ignored
	s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 8)
	s.Require().True(success, "initiating osmo unbondings returns true")
	s.Require().Len(successful_unbondings, 1, "initiating osmo unbondings returns 1 successful unbondings")
	s.Require().Len(failed_unbondings, 0, "initiating osmo unbondings returns 0 failed unbondings")
	s.Require().Equal("OSMO", successful_unbondings[0], "initiating osmo unbondings returns gaia")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_NoneSuccessful() {
	// Tests that if we initiate unbondings a day where none are supposed to unbond, it works successfully
	s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 10)
	s.Require().True(success, "initiating no unbondings returns true")
	s.Require().Len(successful_unbondings, 0, "initiating no unbondings returns 0 successful unbondings")
	s.Require().Len(failed_unbondings, 0, "initiating no unbondings returns 0 failed unbondings")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Failed() {
	// Tests that if Gaia doesn't have enough delegated stake to unbond, it fails
	// but Osmo does and is successful
	s.SetupInitiateAllHostZoneUnbondings()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	hostZone.Validators = []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: sdk.NewInt(1_000_000),
			Weight:        uint64(10),
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx, 12)
	s.Require().False(success, "initiating bad unbondings returns false")
	s.Require().Len(successful_unbondings, 1, "initiating bad unbondings has 1 success")
	s.Require().Len(failed_unbondings, 1, "initiating bad unbondings has 1 failure")
	s.Require().Equal("OSMO", successful_unbondings[0], "initiating bad unbondings succeeds on osmo")
	s.Require().Equal("GAIA", failed_unbondings[0], "initiating bad unbondings fails on gaia")
}
