package keeper_test

import (
	// "fmt"

	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type SendHostZoneUnbondingTestCase struct {
	amtToUnbond           uint64
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZone              stakeibc.HostZone
	lightClientTime       uint64
	totalWgt              uint64
}

func (s *KeeperTestSuite) SetupSendHostZoneUnbonding() SendHostZoneUnbondingTestCase {
	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	delegationAddr := "cosmos_DELEGATION"
	amtToUnbond := uint64(1_000_000)
	amtVal1 := uint64(1_000_000)
	amtVal2 := uint64(2_000_000)
	wgtVal1 := uint64(1)
	wgtVal2 := uint64(2)
	totalWgt := uint64(3)
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
			EpochNumber:        0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
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
		amtToUnbond:           amtToUnbond,
		hostZone:              hostZone,
		epochUnbondingRecords: epochUnbondingRecords,
		lightClientTime:       lightClientTime, // 2022-08-12T19:51.000001, 1ns after the unbonding time
		totalWgt:              totalWgt,
	}
}

func (s *KeeperTestSuite) TestSendHostZoneUnbonding_Successful() {
	tc := s.SetupSendHostZoneUnbonding()

	actualUnbondMsgs, actualAmtToUnbond, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	s.Require().NoError(err)

	// the number of unbonding messages should be (number of validators) * (records to unbond)
	s.Require().Equal(len(tc.epochUnbondingRecords), len(actualUnbondMsgs), "number of unbonding messages should be number of records to unbond")

	s.Require().Equal(int64(actualAmtToUnbond), int64(tc.amtToUnbond)*int64(len(tc.epochUnbondingRecords)), "total amount to unbond should match input amtToUnbond")

	totalWgt := sdk.NewDec(int64(tc.totalWgt)) //getTotalWgt(tc.hostZone)
	actualAmtToUnbondDec := sdk.NewDec(int64(actualAmtToUnbond))
	actualUnbondMsg1 := actualUnbondMsgs[0].String()
	actualUnbondMsg2 := actualUnbondMsgs[1].String()
	val1Fraction := sdk.NewDec(int64(tc.hostZone.Validators[0].Weight)).Quo(totalWgt)
	val2Fraction := sdk.NewDec(int64(tc.hostZone.Validators[1].Weight)).Quo(totalWgt)
	val1UnbondAmt := val1Fraction.Mul(actualAmtToUnbondDec).TruncateInt().String()
	val2UnbondAmt := val2Fraction.Mul(actualAmtToUnbondDec).TruncateInt().String()

	val1Unbonded := strings.Contains(actualUnbondMsg1, val1UnbondAmt)
	val2Unbonded := strings.Contains(actualUnbondMsg2, val2UnbondAmt)

	s.Require().True(val1Unbonded || val2Unbonded)
}

func (s *KeeperTestSuite) TestSendHostZoneUnbonding_WrongChainId() {
	tc := s.SetupSendHostZoneUnbonding()

	tc.hostZone.ChainId = "nonExistentChainId"
	msgs, totalAmtToUnbond, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().Nil(err, "error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages")
	// no messages should be sent
	s.Require().Equal(0, len(msgs), "no messages should be sent")
	// no value should be unbonded
	s.Require().Equal(int64(0), int64(totalAmtToUnbond), "no value should be unbonded")
}

func (s *KeeperTestSuite) TestSendHostZoneUnbonding_NoEpochUnbondingRecords() {
	tc := s.SetupSendHostZoneUnbonding()

	// iterate epoch unbonding records and delete them
	for i := range tc.epochUnbondingRecords {
		s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, uint64(i))
	}

	s.Require().Equal(0, len(s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)), "number of epoch unbonding records should be 0 after deletion")

	msgs, totalAmtToUnbond, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().Nil(err, "error should be nil -- we do NOT raise an error when no records exist, we simply do not send any messages")
	// no messages should be sent
	s.Require().Equal(0, len(msgs), "no messages should be sent")
	// no value should be unbonded
	s.Require().Equal(int64(0), int64(totalAmtToUnbond), "no value should be unbonded")
}

func (s *KeeperTestSuite) TestSendHostZoneUnbonding_UnbondingTooMuch() {
	tc := s.SetupSendHostZoneUnbonding()

	// iterate the validators and set all their delegated amounts to 0
	for i := range tc.hostZone.Validators {
		tc.hostZone.Validators[i].DelegationAmt = 0
	}
	// write the host zone with zero-delegation validators back to the store
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	_, _, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	s.Require().EqualError(err, fmt.Sprintf("Could not unbond %d on Host Zone %s, unable to balance the unbond amount across validators: not found", tc.amtToUnbond*uint64(len(tc.epochUnbondingRecords)), tc.hostZone.ChainId))
}
