package keeper_test

import (
	// "fmt"

	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	// "github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type SendHostZoneUnbondingTestCase struct {
	amt_to_unbond           uint64
	epoch_unbonding_records []recordtypes.EpochUnbondingRecord
	host_zone               stakeibc.HostZone
	lightClientTime         uint64
}

func (s *KeeperTestSuite) SetupSendHostZoneUnbonding() SendHostZoneUnbondingTestCase {
	host_val1_addr := "cosmos_VALIDATOR_1"
	host_val2_addr := "cosmos_VALIDATOR_2"
	delegationAddr := "cosmos_DELEGATION"
	amt_to_unbond := uint64(1_000_000)
	amt_val1 := uint64(2_000_000)
	amt_val2 := uint64(1_000_000)
	wgt_val1 := uint64(2)
	wgt_val2 := uint64(1)
	unbonding_time := uint64(1660348276)
	light_client_time := unbonding_time + 1

	//  define the host zone with stakedBal and validators with staked amounts
	validators := []*stakeibc.Validator{
		{
			Address:       host_val1_addr,
			DelegationAmt: amt_val1,
			Weight:        wgt_val1,
		},
		{
			Address:       host_val2_addr,
			DelegationAmt: amt_val2,
			Weight:        wgt_val2,
		},
	}

	delegationAccount := stakeibc.ICAAccount{
		Address: delegationAddr,
		Target:  stakeibc.ICAAccountType_DELEGATION,
	}

	hostZone := stakeibc.HostZone{
		ChainId:           "GAIA",
		HostDenom:         "uatom",
		Bech32Prefix:      "cosmos",
		Validators:        validators,
		DelegationAccount: &delegationAccount,
	}

	// list of epoch unbonding records
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		recordtypes.EpochUnbondingRecord{
			Id:                   1,
			UnbondingEpochNumber: 1,
			HostZoneUnbondings:   make(map[string]*recordtypes.HostZoneUnbonding),
		},
	}

	// for each epoch unbonding record, add a host zone unbonding record and append the record
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		// note: we're using the same hostZoneUnbonding for each epoch unbonding record in this test
		epochUnbondingRecords[0].HostZoneUnbondings["GAIA"] = &recordtypes.HostZoneUnbonding{
			Amount:        amt_to_unbond,
			Denom:         "uatom",
			HostZoneId:    "GAIA",
			UnbondingTime: unbonding_time, // 2022-08-12T19:51
			Status:        recordtypes.HostZoneUnbonding_BONDED,
		}

		s.App.RecordsKeeper.AppendEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return SendHostZoneUnbondingTestCase{
		amt_to_unbond:           amt_to_unbond,
		host_zone:               hostZone,
		epoch_unbonding_records: epochUnbondingRecords,
		lightClientTime:         light_client_time, // 2022-08-13T19:51, 1d after the unbonding time
	}
}

func (s *KeeperTestSuite) TestSendHostZoneUnbondingSuccessful() {
	tc := s.SetupSendHostZoneUnbonding()

	msgs, totalAmtToUnbond, err := s.App.StakeibcKeeper.SendHostZoneUnbondings(s.Ctx, tc.host_zone)
	s.Require().Nil(err)

	// function to get total wgt of all validators on a host zone
	getTotalWgt := func(hostZone stakeibc.HostZone) sdk.Dec {
		totalWgt := sdk.ZeroDec()
		for _, validator := range hostZone.Validators {
			totalWgt = totalWgt.Add(sdk.NewDec(int64(validator.Weight)))
		}
		return totalWgt
	}
	totalWgt := getTotalWgt(tc.host_zone)
	amtToUnbond := sdk.NewDec(int64(tc.amt_to_unbond))
	// due to erratic rounding and random validator ordering, we can only guarantee that at least one of the undelegations is correct
	a := strings.Contains(msgs[1].String(), (sdk.NewDec(int64(tc.host_zone.Validators[1].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())
	b := strings.Contains(msgs[0].String(), (sdk.NewDec(int64(tc.host_zone.Validators[1].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())
	c := strings.Contains(msgs[1].String(), (sdk.NewDec(int64(tc.host_zone.Validators[0].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())
	d := strings.Contains(msgs[0].String(), (sdk.NewDec(int64(tc.host_zone.Validators[0].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())

	s.Require().True(a || b || c || d)
	s.Require().Equal(int64(tc.amt_to_unbond), int64(totalAmtToUnbond))
}

func (s *KeeperTestSuite) TestSendHostZoneUnbondingWrongChainId() {
	tc := s.SetupSendHostZoneUnbonding()

	tc.host_zone.ChainId = "nonExistentChainId"
	msgs, totalAmtToUnbond, err := s.App.StakeibcKeeper.SendHostZoneUnbondings(s.Ctx, tc.host_zone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().Nil(err)
	// no messages should be sent
	s.Require().Equal(0, len(msgs))
	// no value should be unbonded
	s.Require().Equal(int64(0), int64(totalAmtToUnbond))
}

func (s *KeeperTestSuite) TestSendHostZoneUnbondingNoEpochUnbondingRecords() {
	tc := s.SetupSendHostZoneUnbonding()

	// iterate epoch unbonding records and delete them
	for i, _ := range tc.epoch_unbonding_records {
		s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, uint64(i))
	}

	msgs, totalAmtToUnbond, err := s.App.StakeibcKeeper.SendHostZoneUnbondings(s.Ctx, tc.host_zone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().Nil(err)
	// no messages should be sent
	s.Require().Equal(0, len(msgs))
	// no value should be unbonded
	s.Require().Equal(int64(0), int64(totalAmtToUnbond))
}

func (s *KeeperTestSuite) TestSendHostZoneUnbondingUnbondingTooMuch() {
	tc := s.SetupSendHostZoneUnbonding()

	// iterate the validators and set all their delegated amounts to 0
	for i, _ := range tc.host_zone.Validators {
		tc.host_zone.Validators[i].DelegationAmt = 0
	}

	_, _, err := s.App.StakeibcKeeper.SendHostZoneUnbondings(s.Ctx, tc.host_zone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().EqualError(err, fmt.Sprintf("Could not unbond %d on Host Zone %s, overflow is %d: balance is insufficient", tc.amt_to_unbond, tc.host_zone.ChainId, tc.amt_to_unbond))
}
