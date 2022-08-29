package keeper_test

import (
	"fmt"

	ibctesting "github.com/cosmos/ibc-go/v3/testing"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type InitiateAllHostZoneUnbondingsTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZones             []stakeibc.HostZone
}

func (s *KeeperTestSuite) SetupInitiateAllHostZoneUnbondings() InitiateAllHostZoneUnbondingsTestCase {
	s.CreateICAChannel("icacontroller-GAIA.DELEGATION")

	gaiaValAddr := "cosmos_VALIDATOR"
	osmoValAddr := "osmo_VALIDATOR"
	gaiaDelegationAddr := "cosmos_DELEGATION"
	osmoDelegationAddr := "osmo_DELEGATION"
	//  define the host zone with stakedBal and validators with staked amounts
	gaiaValidators := []*stakeibc.Validator{
		{
			Address:       gaiaValAddr,
			DelegationAmt: uint64(5_000_000),
			Weight:        uint64(10),
		},
	}
	gaiaDelegationAccount := stakeibc.ICAAccount{
		Address: gaiaDelegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	osmoValidators := []*stakeibc.Validator{
		{
			Address:       osmoValAddr,
			DelegationAmt: uint64(5_000_000),
			Weight:        uint64(10),
		},
	}
	osmoDelegationAccount := stakeibc.ICAAccount{
		Address: osmoDelegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}
	hostZones := []stakeibc.HostZone{
		{
			ChainId:            "GAIA",
			HostDenom:          "uatom",
			Bech32Prefix:       "cosmos",
			UnbondingFrequency: 3,
			Validators:         gaiaValidators,
			DelegationAccount:  &gaiaDelegationAccount,
			StakedBal:          uint64(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
		{
			ChainId:            "OSMO",
			HostDenom:          "uosmo",
			Bech32Prefix:       "osmo",
			UnbondingFrequency: 4,
			Validators:         osmoValidators,
			DelegationAccount:  &osmoDelegationAccount,
			StakedBal:          uint64(5_000_000),
			ConnectionId:       ibctesting.FirstConnectionID,
		},
	}
	// list of epoch unbonding records
	default_unbonding := []*recordtypes.HostZoneUnbonding{
		{
			HostZoneId:        "GAIA",
			StTokenAmount:     1_900_000,
			NativeTokenAmount: 2_000_000,
			Denom:             "uatom",
			Status:            recordtypes.HostZoneUnbonding_BONDED,
		},
		{
			HostZoneId:        "OSMO",
			StTokenAmount:     2_800_000,
			NativeTokenAmount: 3_000_000,
			Denom:             "uosmo",
			Status:            recordtypes.HostZoneUnbonding_BONDED,
		},
	}
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{}
	for _, epochNumber := range []uint64{5} {
		epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
			EpochNumber:        epochNumber,
			HostZoneUnbondings: default_unbonding,
		}
		epochUnbondingRecords = append(epochUnbondingRecords, epochUnbondingRecord)
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	}

	for _, hostZone := range hostZones {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	}

	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx(), stakeibc.EpochTracker{
		EpochIdentifier:    "day",
		EpochNumber:        12,
		NextEpochStartTime: uint64(2661750006000000000),
		Duration:           uint64(1000000000000),
	})

	return InitiateAllHostZoneUnbondingsTestCase{
		epochUnbondingRecords: epochUnbondingRecords,
		hostZones:             hostZones,
	}
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Successful() {
	_ = s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx(), 12)
	s.Require().Equal(success, true, "initiating unbondings returns true")
	s.Require().Equal(len(successful_unbondings), 2, "initiating unbondings returns 2 successful unbondings")
	s.Require().Equal(len(failed_unbondings), 0, "initiating unbondings returns 0 failed unbondings")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_GaiaSuccessful() {
	_ = s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx(), 9)
	s.Require().Equal(success, true, "initiating gaia unbondings returns true")
	s.Require().Equal(len(successful_unbondings), 1, "initiating gaia unbondings returns 1 successful unbondings")
	s.Require().Equal(len(failed_unbondings), 0, "initiating gaia unbondings returns 0 failed unbondings")
	s.Require().Equal(successful_unbondings[0], "GAIA", "initiating gaia unbondings returns gaia")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_OsmoSuccessful() {
	_ = s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx(), 8)
	s.Require().Equal(success, true, "initiating osmo unbondings returns true")
	s.Require().Equal(len(successful_unbondings), 1, "initiating osmo unbondings returns 1 successful unbondings")
	s.Require().Equal(len(failed_unbondings), 0, "initiating osmo unbondings returns 0 failed unbondings")
	s.Require().Equal(successful_unbondings[0], "OSMO", "initiating osmo unbondings returns gaia")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_NoneSuccessful() {
	_ = s.SetupInitiateAllHostZoneUnbondings()
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx(), 10)
	s.Require().Equal(success, true, "initiating no unbondings returns true")
	s.Require().Equal(len(successful_unbondings), 0, "initiating no unbondings returns 0 successful unbondings")
	s.Require().Equal(len(failed_unbondings), 0, "initiating no unbondings returns 0 failed unbondings")
}

func (s *KeeperTestSuite) TestInitiateAllHostZoneUnbondings_Failed() {
	_ = s.SetupInitiateAllHostZoneUnbondings()
	hostZone, _ := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), "GAIA")
	hostZone.Validators = []*stakeibc.Validator{
		{
			Address:       "cosmos_VALIDATOR",
			DelegationAmt: uint64(1_000_000),
			Weight:        uint64(10),
		},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	hostZone, _ = s.App.StakeibcKeeper.GetHostZone(s.Ctx(), "GAIA")
	fmt.Printf("%v\n", hostZone)
	success, successful_unbondings, failed_unbondings := s.App.StakeibcKeeper.InitiateAllHostZoneUnbondings(s.Ctx(), 12)
	s.Require().Equal(success, false, "initiating bad unbondings returns false")
	s.Require().Equal(len(successful_unbondings), 1, "initiating bad unbondings has 1 success")
	s.Require().Equal(len(failed_unbondings), 1, "initiating bad unbondings has 1 failure")
	s.Require().Equal(successful_unbondings[0], "OSMO", "initiating bad unbondings succeeds on osmo")
	s.Require().Equal(failed_unbondings[0], "GAIA", "initiating bad unbondings fails on gaia")
}
