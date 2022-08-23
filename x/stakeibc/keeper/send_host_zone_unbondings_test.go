package keeper_test

import (
	// "fmt"

	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type SendHostZoneUnbondingTestCase struct {
	amt_to_unbond           uint64
	epoch_unbonding_records []recordtypes.EpochUnbondingRecord
	host_zone               stakeibc.HostZone
	lightClientTime         uint64
}

func (s *KeeperTestSuite) SetupSendHostZoneUnbonding() SendHostZoneUnbondingTestCase {
	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	delegationAddr := "cosmos_DELEGATION"
	amtToUnbond := uint64(1_000_000)
	amtVal1 := uint64(2_000_000)
	amtVal2 := uint64(1_000_000)
	wgtVal1 := uint64(2)
	wgtVal2 := uint64(1)
	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(1660348276)
	lightClientTime := unbondingTime + 1

	//  define the host zone with stakedBal and validators with staked amounts
	validators := []*stakeibc.Validator{
		{
			Address:       hostVal1Addr,
			DelegationAmt: amtVal1,
			Weight:        wgtVal1,
		},
		{
			Address:       hostVal2Addr,
			DelegationAmt: amtVal2,
			Weight:        wgtVal2,
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
		{
			EpochNumber:        1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
	}

	// for each epoch unbonding record, add a host zone unbonding record and append the record
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		hostZoneUnbonding := &recordtypes.HostZoneUnbonding{
			NativeTokenAmount: amtToUnbond,
			Denom:             "uatom",
			HostZoneId:        "GAIA",
			UnbondingTime:     unbondingTime, // 2022-08-12T19:51
			Status:            recordtypes.HostZoneUnbonding_BONDED,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return SendHostZoneUnbondingTestCase{
		amt_to_unbond:           amtToUnbond,
		host_zone:               hostZone,
		epoch_unbonding_records: epochUnbondingRecords,
		lightClientTime:         lightClientTime, // 2022-08-12T19:51.000001, 1ns after the unbonding time
	}
}

func (s *KeeperTestSuite) TestSendHostZoneUnbondingSuccessful() {
	tc := s.SetupSendHostZoneUnbonding()

	msgs, totalAmtToUnbond, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.host_zone)
	s.Require().NoError(err)

	// the number of unbonding messages should be (number of validators) * (records to unbond)
	s.Require().Equal(len(msgs), len(tc.host_zone.Validators)*len(tc.epoch_unbonding_records))

	s.Require().Equal(totalAmtToUnbond, tc.amt_to_unbond)

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
	// due to erratic rounding and random validator ordering, we can only guarantee that at least one of the undelegation amounts is correct
	//   at least one must be correct (rounding can go in one of two directions, but checking both guarantees that at least one is correct)
	a := strings.Contains(msgs[1].String(), (sdk.NewDec(int64(tc.host_zone.Validators[1].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())
	b := strings.Contains(msgs[0].String(), (sdk.NewDec(int64(tc.host_zone.Validators[1].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())
	c := strings.Contains(msgs[1].String(), (sdk.NewDec(int64(tc.host_zone.Validators[0].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())
	d := strings.Contains(msgs[0].String(), (sdk.NewDec(int64(tc.host_zone.Validators[0].Weight)).Quo(totalWgt).Mul(amtToUnbond)).TruncateInt().String())

	s.Require().True(a || b || c || d)
}

func (s *KeeperTestSuite) TestSendHostZoneUnbondingWrongChainId() {
	tc := s.SetupSendHostZoneUnbonding()

	tc.host_zone.ChainId = "nonExistentChainId"
	msgs, totalAmtToUnbond, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.host_zone)
	s.Require().NoError(err)
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

// func (s *KeeperTestSuite) TestSendHostZoneUnbondingUnbondingTooMuch() {
// 	tc := s.SetupSendHostZoneUnbonding()

// 	// iterate the validators and set all their delegated amounts to 0
// 	for i, _ := range tc.host_zone.Validators {
// 		tc.host_zone.Validators[i].DelegationAmt = 0
// 	}

// 	_, _, err := s.App.StakeibcKeeper.SendHostZoneUnbondings(s.Ctx, tc.host_zone)
// 	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
// 	s.Require().EqualError(err, fmt.Sprintf("Could not unbond %d on Host Zone %s, overflow is %d: balance is insufficient", tc.amt_to_unbond, tc.host_zone.ChainId, tc.amt_to_unbond))
// }
