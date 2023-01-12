package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type GetHostZoneUnbondingMsgsTestCase struct {
	amtToUnbond           sdk.Int
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	hostZone              stakeibc.HostZone
	lightClientTime       uint64
	totalWgt              uint64
	valNames              []string
}

func (s *KeeperTestSuite) SetupGetHostZoneUnbondingMsgs() GetHostZoneUnbondingMsgsTestCase {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)

	hostVal1Addr := "cosmos_VALIDATOR_1"
	hostVal2Addr := "cosmos_VALIDATOR_2"
	hostVal3Addr := "cosmos_VALIDATOR_3"
	valNames := []string{hostVal1Addr, hostVal2Addr, hostVal3Addr}
	delegationAddr := "cosmos_DELEGATION"
	amtToUnbond := sdk.NewInt(1_000_000)
	amtVal1 := sdk.NewInt(1_000_000)
	amtVal2 := sdk.NewInt(2_000_000)
	wgtVal1 := uint64(1)
	wgtVal2 := uint64(2)
	totalWgt := uint64(5)
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
		{
			Address: hostVal3Addr,
			// DelegationAmt and Weight are the same as Val2, to test tie breaking
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
			UnbondingTime:     unbondingTime, // 2022-08-12T19:52
			Status:            recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hostZoneUnbonding)
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	return GetHostZoneUnbondingMsgsTestCase{
		amtToUnbond:           amtToUnbond,
		hostZone:              hostZone,
		epochUnbondingRecords: epochUnbondingRecords,
		lightClientTime:       lightClientTime, // 2022-08-12T19:51.000001, 1ns after the unbonding time
		totalWgt:              totalWgt,
		valNames:              valNames,
	}
}

func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_Successful() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	// TODO: check epoch unbonding record ids here
	actualUnbondMsgs, actualAmtToUnbond, actualCallbackArgs, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	s.Require().NoError(err)

	// verify the callback attributes are as expected
	actualCallbackResult, err := s.App.StakeibcKeeper.UnmarshalUndelegateCallbackArgs(s.Ctx, actualCallbackArgs)
	s.Require().NoError(err, "could unmarshal undelegation callback args")
	s.Require().Equal(len(tc.hostZone.Validators), len(actualCallbackResult.SplitDelegations), "number of split delegations in success unbonding case")
	s.Require().Equal(tc.hostZone.ChainId, actualCallbackResult.HostZoneId, "host zone id in success unbonding case")

	// TODO add case that checks the *marshaled* callback args against expectations

	// the number of unbonding messages should be (number of validators) * (records to unbond)
	s.Require().Equal(len(tc.valNames), len(actualUnbondMsgs), "number of unbonding messages should be number of records to unbond")

	s.Require().Equal(tc.amtToUnbond.Mul(sdk.NewInt(int64(len(tc.epochUnbondingRecords)))), actualAmtToUnbond, "total amount to unbond should match input amtToUnbond")
	
	totalWgt := sdk.NewDec(int64(tc.totalWgt))
	actualAmtToUnbondDec := sdk.NewDecFromInt(actualAmtToUnbond)
	actualUnbondMsg1 := actualUnbondMsgs[0].String()
	actualUnbondMsg2 := actualUnbondMsgs[1].String()
	val1Fraction := sdk.NewDec(int64(tc.hostZone.Validators[0].Weight)).Quo(totalWgt)
	val2Fraction := sdk.NewDec(int64(tc.hostZone.Validators[1].Weight)).Quo(totalWgt)
	val1UnbondAmt := val1Fraction.Mul(actualAmtToUnbondDec).TruncateInt().String()
	val2UnbondAmt := val2Fraction.Mul(actualAmtToUnbondDec).TruncateInt().String()

	val1Unbonded := strings.Contains(actualUnbondMsg1, val1UnbondAmt)
	val2Unbonded := strings.Contains(actualUnbondMsg2, val2UnbondAmt)

	// there's rounding in the logic that distributes stake amongst validators, so one or the other of the balances will be correct, depending on the rounding
	// at least one will be correct, and the other will be off by 1 by rounding, so we check and OR condition
	s.Require().True(val1Unbonded || val2Unbonded, "unbonding amt should be the correct amount")
}

func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_WrongChainId() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	tc.hostZone.ChainId = "nonExistentChainId"
	// TODO: check epoch unbonding record ids here
	msgs, totalAmtToUnbond, _, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().Nil(err, "error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages")
	// no messages should be sent
	s.Require().Equal(0, len(msgs), "no messages should be sent")
	// no value should be unbonded
	s.Require().Equal(sdk.ZeroInt(), totalAmtToUnbond, "no value should be unbonded")
}

func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_NoEpochUnbondingRecords() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	// iterate epoch unbonding records and delete them
	for i := range tc.epochUnbondingRecords {
		s.App.RecordsKeeper.RemoveEpochUnbondingRecord(s.Ctx, uint64(i))
	}

	s.Require().Equal(0, len(s.App.RecordsKeeper.GetAllEpochUnbondingRecord(s.Ctx)), "number of epoch unbonding records should be 0 after deletion")

	// TODO: check epoch unbonding record ids here
	msgs, totalAmtToUnbond, _, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	// error should be nil -- we do NOT raise an error on a non-existent chain id, we simply do not send any messages
	s.Require().Nil(err, "error should be nil -- we do NOT raise an error when no records exist, we simply do not send any messages")
	// no messages should be sent
	s.Require().Equal(0, len(msgs), "no messages should be sent")
	// no value should be unbonded
	s.Require().Equal(sdk.ZeroInt(), totalAmtToUnbond, "no value should be unbonded")
}

func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_UnbondingTooMuch() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	// iterate the validators and set all their delegated amounts to 0
	for i := range tc.hostZone.Validators {
		tc.hostZone.Validators[i].DelegationAmt = sdk.ZeroInt()
	}
	// write the host zone with zero-delegation validators back to the store
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, tc.hostZone)

	// TODO: check epoch unbonding record ids here
	_, _, _, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, tc.hostZone)
	s.Require().EqualError(err, fmt.Sprintf("Could not unbond %v on Host Zone %s, unable to balance the unbond amount across validators: not found", tc.amtToUnbond.Mul(sdk.NewInt(int64(len(tc.epochUnbondingRecords)))), tc.hostZone.ChainId))
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_Success() {
	tc := s.SetupGetHostZoneUnbondingMsgs()

	// verify the total amount is expected
	unbond := sdk.NewInt(1_000_000)
	totalAmt, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, tc.hostZone, unbond)
	s.Require().Nil(err)

	// sum up totalAmt
	actualAmount := sdk.ZeroInt()
	for _, amt := range totalAmt {
		actualAmount = actualAmount.Add(amt)
	}
	s.Require().Equal(unbond, actualAmount, "total amount unbonded matches input")

	// verify the order of the validators is expected
	// GetTargetValAmtsForHostZone first reverses the list, then sorts by weight using SliceStable
	// E.g. given A:1, B:2, C:2
	// 1. C:2, B:2, A:1
	// 2. A:1, C:2, B:2
	s.Require().Equal(tc.valNames[0], tc.hostZone.Validators[0].Address)
	s.Require().Equal(tc.valNames[1], tc.hostZone.Validators[2].Address)
	s.Require().Equal(tc.valNames[2], tc.hostZone.Validators[1].Address)
}
