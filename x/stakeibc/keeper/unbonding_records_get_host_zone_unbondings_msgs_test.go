package keeper_test

import (
	"fmt"
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"

	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

var (
	hostVal1Addr = "cosmos_VALIDATOR_1"
	hostVal2Addr = "cosmos_VALIDATOR_2"
	hostVal3Addr = "cosmos_VALIDATOR_3"
	amtVal1      = sdk.NewInt(1_000_000)
	amtVal2      = sdk.NewInt(2_000_000)
	wgtVal1      = uint64(1)
	wgtVal2      = uint64(2)
	amtToUnbond  = sdk.NewInt(1_000_000)
)

var delegationAccount = stakeibc.ICAAccount{
	Address: "cosmos_DELEGATION",
	Target:  stakeibc.ICAAccountType_DELEGATION,
}

type GetHostZoneUnbondingMsgsTestCase struct {
	valAddrToUnbondAmt    map[string]int64
	HostDenom             string
	Bech32Prefix          string
	DelegationAccount     *stakeibc.ICAAccount
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	chainId               string
	Validators            []*stakeibc.Validator
	totalWeight           uint64
	expectErr             error
	expectPass            bool
}

var defaultUnbondingTestCase = GetHostZoneUnbondingMsgsTestCase{
	epochUnbondingRecords: []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber:        0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
		{
			EpochNumber:        1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
	},
	chainId:           "GAIA",
	HostDenom:         "uatom",
	Bech32Prefix:      "cosmos",
	DelegationAccount: &delegationAccount,

	valAddrToUnbondAmt: map[string]int64{
		"cosmos_VALIDATOR_1": 200000,
		"cosmos_VALIDATOR_2": 400000,
		"cosmos_VALIDATOR_3": 400000,
	},
	Validators: []*stakeibc.Validator{
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
	},
	totalWeight: 5,
	expectPass:  true,
}

func (s *KeeperTestSuite) SetupGetHostZoneUnbondingMsgs(tc GetHostZoneUnbondingMsgsTestCase) (sdk.Int, stakeibc.HostZone) {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)

	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(1660348276)

	//  define the host zone with stakedBal and validators with staked amounts
	validators := tc.Validators

	hostZone := stakeibc.HostZone{
		ChainId:           "GAIA",
		HostDenom:         "uatom",
		Bech32Prefix:      "cosmos",
		Validators:        validators,
		DelegationAccount: &delegationAccount,
	}

	// list of epoch unbonding records
	epochUnbondingRecords := tc.epochUnbondingRecords

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

	if tc.chainId != "" {
		hostZone.ChainId = tc.chainId
	}
	if tc.DelegationAccount != &delegationAccount {
		hostZone.DelegationAccount = tc.DelegationAccount
	}

	return amtToUnbond, hostZone
}

func (s *KeeperTestSuite) VerifyTestCase_TestGetHostZoneUnbondingMsgs(name string, tc GetHostZoneUnbondingMsgsTestCase) {
	s.Setup()
	AmtToUnbond, hostZone := s.SetupGetHostZoneUnbondingMsgs(tc)

	actualUnbondMsgs, actualAmtToUnbond, actualCallbackArgs, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx, hostZone)

	if tc.expectPass {
		s.Require().NoError(err)

		// verify the callback attributes are as expected
		actualCallbackResult, err := s.App.StakeibcKeeper.UnmarshalUndelegateCallbackArgs(s.Ctx, actualCallbackArgs)
		s.Require().NoError(err, "could unmarshal undelegation callback args")
		s.Require().Equal(len(hostZone.Validators), len(actualCallbackResult.SplitDelegations), "number of split delegations in success unbonding case")
		s.Require().Equal(hostZone.ChainId, actualCallbackResult.HostZoneId, "host zone id in success unbonding case")

		// the number of unbonding messages should be (number of validators) * (records to unbond)
		s.Require().Equal(len(tc.Validators), len(actualUnbondMsgs), "number of unbonding messages should be number of records to unbond")

		s.Require().Equal(AmtToUnbond.Int64()*int64(len(tc.epochUnbondingRecords)), actualAmtToUnbond.Int64(), "total amount to unbond should match input amtToUnbond")

		totalWgt := sdk.NewDec(int64(tc.totalWeight))
		actualAmtToUnbondDec := sdk.NewDec(actualAmtToUnbond.Int64())

		for i, validator := range tc.Validators {
			actualUnbondMsg := actualUnbondMsgs[i].String()
			valFraction := sdk.NewDec(int64(validator.Weight)).Quo(totalWgt)
			val1UnbondAmt := valFraction.Mul(actualAmtToUnbondDec).TruncateInt().String()
			valUnbonded := strings.Contains(actualUnbondMsg, val1UnbondAmt)
			// there's rounding in the logic that distributes stake amongst validators, so one or the other of the balances will be correct, depending on the rounding
			// at least one will be correct, and the other will be off by 1 by rounding, so we check and OR condition
			s.Require().True(valUnbonded, "unbonding amt should be the correct amount")
		}
	} else {
		if tc.expectErr != nil {
			s.Require().EqualError(err, tc.expectErr.Error())
		}
		// no messages should be sent
		s.Require().Equal(0, len(actualUnbondMsgs), "no messages should be sent")
		// no value should be unbonded
		s.Require().Equal(int64(0), actualAmtToUnbond.Int64(), "no value should be unbonded")
	}

}
func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_Successful() {

	name := "Successful"
	var test = defaultUnbondingTestCase
	s.VerifyTestCase_TestGetHostZoneUnbondingMsgs(name, test)

}
func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_WrongChainId() {
	name := "Wrong chain id"
	var tc = defaultUnbondingTestCase
	tc.chainId = "nonExistentChainId"
	tc.expectPass = false

	s.VerifyTestCase_TestGetHostZoneUnbondingMsgs(name, tc)
}
func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_NoEpochUnbondingRecords() {
	name := "No epoch unbonding records"
	var tc = defaultUnbondingTestCase
	tc.epochUnbondingRecords = []recordtypes.EpochUnbondingRecord{}
	tc.expectPass = false

	s.VerifyTestCase_TestGetHostZoneUnbondingMsgs(name, tc)
}
func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_UnbodingTooMuch() {
	name := "Unbonding too much"
	var tc = defaultUnbondingTestCase

	for _, validator := range tc.Validators {
		validator.DelegationAmt = sdk.ZeroInt()
	}
	tc.expectPass = false
	tc.expectErr = sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("Could not unbond %d on Host Zone %s, unable to balance the unbond amount across validators", uint64(2_000_000), "GAIA"))

	s.VerifyTestCase_TestGetHostZoneUnbondingMsgs(name, tc)
}
func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs_NoNonzeroWeightValidator() {
	name := "No non-zero weight validator"
	var tc = defaultUnbondingTestCase

	for _, validator := range tc.Validators {
		validator.Weight = 0
	}
	tc.expectErr = sdkerrors.Wrap(stakeibc.ErrNoValidatorAmts, fmt.Sprintf("Error getting target val amts for host zone %s %d: no non-zero validator weights", "GAIA", uint64(2_000_000)))
	tc.expectPass = false

	s.VerifyTestCase_TestGetHostZoneUnbondingMsgs(name, tc)
}

func (s *KeeperTestSuite) VerifyTestCase_TestGetUnbondingAmountAndRecords(name string, tc GetHostZoneUnbondingMsgsTestCase) {
	s.Setup()
	amtToUnbond, hostZone := s.SetupGetHostZoneUnbondingMsgs(tc)
	actualAmtToUnbond, actualUnbondRecords := s.App.StakeibcKeeper.GetUnbondingAmountAndRecords(s.Ctx, hostZone)
	if tc.expectPass {
		s.Require().Equal(amtToUnbond.Int64()*int64(len(tc.epochUnbondingRecords)), actualAmtToUnbond.Int64(), "total amount to unbond should match input amtToUnbond")
		s.Require().Equal(len(tc.epochUnbondingRecords), len(actualUnbondRecords))
	} else {
		// no messages should be sent
		s.Require().Equal(0, len(actualUnbondRecords), "no messages should be sent")
		// no value should be unbonded
		s.Require().Equal(int64(0), actualAmtToUnbond.Int64(), "no value should be unbonded")
	}

}
func (s *KeeperTestSuite) TestGetUnbondingAmountAndRecords_WrongChainId() {
	name := "Wrong chain id"
	var test = defaultUnbondingTestCase
	test.chainId = "nonExistentChainId"
	test.expectPass = false
	s.VerifyTestCase_TestGetUnbondingAmountAndRecords(name, test)

}

func (s *KeeperTestSuite) TestGetUnbondingAmountAndRecords_Successful() {
	name := "Successful"
	var test = defaultUnbondingTestCase

	s.VerifyTestCase_TestGetUnbondingAmountAndRecords(name, test)

}
func (s *KeeperTestSuite) TestGetUnbondingAmountAndRecords_NoEpochUnbondingRecords() {
	name := "No Epoch Unbonding Records"
	var test = defaultUnbondingTestCase
	test.expectPass = false
	test.epochUnbondingRecords = []recordtypes.EpochUnbondingRecord{}

	s.VerifyTestCase_TestGetUnbondingAmountAndRecords(name, test)

}
func (s *KeeperTestSuite) VerifyTestCase_TestDistributeUnbondingAmountToValidators(name string, tc GetHostZoneUnbondingMsgsTestCase) {
	s.Setup()
	AmtToUnbond, hostZone := s.SetupGetHostZoneUnbondingMsgs(tc)
	actualAmtToUnbond, err := s.App.StakeibcKeeper.DistributeUnbondingAmountToValidators(s.Ctx, hostZone, amtToUnbond)
	// fmt.Println(actualAmtToUnbond)

	if tc.expectPass {
		s.Require().NoError(err)
		totalWgt := sdk.NewDec(int64(tc.totalWeight))
		actualAmtToUnbondDec := sdk.NewDec(AmtToUnbond.Int64())

		for _, validator := range tc.Validators {
			valFraction := sdk.NewDec(int64(validator.Weight)).Quo(totalWgt)
			valUnbondAmt := valFraction.Mul(actualAmtToUnbondDec).TruncateInt()
			s.Require().Equal(valUnbondAmt, sdk.NewInt(actualAmtToUnbond[validator.Address]))
		}

	} else {
		s.Require().Error(err)

		s.Require().EqualError(err, tc.expectErr.Error())
	}

}
func (s *KeeperTestSuite) TestDistributeUnbondingAmountToValidators_Successful() {

	name := "Successful"
	var test = defaultUnbondingTestCase

	s.VerifyTestCase_TestDistributeUnbondingAmountToValidators(name, test)

}
func (s *KeeperTestSuite) TestDistributeUnbondingAmountToValidators_EmptyValidators() {
	name := "Empty validators"
	var test = defaultUnbondingTestCase
	test.expectPass = false
	test.Validators = []*stakeibc.Validator{}
	test.expectErr = sdkerrors.Wrap(stakeibc.ErrNoValidatorAmts, fmt.Sprintf("Error getting target val amts for host zone %s %d: no non-zero validator weights", "GAIA", uint64(1_000_000)))

	s.VerifyTestCase_TestDistributeUnbondingAmountToValidators(name, test)

}
func (s *KeeperTestSuite) TestDistributeUnbondingAmountToValidators_TotalWeightIsZero() {
	name := "Total weight is zero"
	var test = defaultUnbondingTestCase

	for _, validator := range test.Validators {
		validator.Weight = 0
	}
	test.expectPass = false
	test.expectErr = sdkerrors.Wrap(stakeibc.ErrNoValidatorAmts, fmt.Sprintf("Error getting target val amts for host zone %s %d: no non-zero validator weights", "GAIA", uint64(1_000_000)))

	s.VerifyTestCase_TestDistributeUnbondingAmountToValidators(name, test)

}
func (s *KeeperTestSuite) TestDistributeUnbondingAmountToValidators_UnbondingTooMuch() {
	name := "Unbonding too much"
	var test = defaultUnbondingTestCase
	test.expectPass = false
	for _, validator := range test.Validators {
		validator.DelegationAmt = sdk.ZeroInt()
	}
	test.expectErr = sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("Could not unbond %d on Host Zone %s, unable to balance the unbond amount across validators", uint64(1_000_000), "GAIA"))

	s.VerifyTestCase_TestDistributeUnbondingAmountToValidators(name, test)

}

func (s *KeeperTestSuite) VerifyTestCase_TestSplitDelegationMsg(name string, tc GetHostZoneUnbondingMsgsTestCase) {
	s.Setup()
	_, hostZone := s.SetupGetHostZoneUnbondingMsgs(tc) //amtToUnbond
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	msgs, splitDelegations, err := s.App.StakeibcKeeper.SplitDelegationMsg(s.Ctx, tc.valAddrToUnbondAmt, hostZone)
	if tc.expectPass {
		s.Require().NoError(err)

		s.Require().Equal(len(tc.Validators), len(splitDelegations), "number of split delegations in success unbonding case")

		// the number of unbonding messages should be (number of validators) * (records to unbond)
		s.Require().Equal(len(tc.Validators), len(msgs), "number of unbonding messages should be number of records to unbond")
		for i, validator := range tc.Validators {
			actualUnbondMsg := msgs[i].String()
			valUnbonded := strings.Contains(actualUnbondMsg, strconv.Itoa(int(tc.valAddrToUnbondAmt[validator.Address])))
			// there's rounding in the logic that distributes stake amongst validators, so one or the other of the balances will be correct, depending on the rounding
			// at least one will be correct, and the other will be off by 1 by rounding, so we check and OR condition
			s.Require().True(valUnbonded, "unbonding amt should be the correct amount")
		}
	} else {
		s.Require().Error(err)
		s.Require().EqualError(err, "Zone GAIA is missing a delegation address!: not found") //tc.expectErr.Error())
	}
}
func (s *KeeperTestSuite) TestSplitDelegationMsg_Successful() {
	name := "Successful"
	var test = defaultUnbondingTestCase

	s.VerifyTestCase_TestSplitDelegationMsg(name, test)
}
func (s *KeeperTestSuite) TestSplitDelegationMsg_MissingDelegationAddress() {
	name := "Missing a delegation address"
	var test = defaultUnbondingTestCase
	test.expectPass = false
	test.DelegationAccount = nil

	s.VerifyTestCase_TestSplitDelegationMsg(name, test)
}

func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_Success() {
	_, hostZone := s.SetupGetHostZoneUnbondingMsgs(defaultUnbondingTestCase)

	// verify the total amount is expected
	unbond := sdk.NewInt(1_000_000)
	totalAmt, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx, hostZone, unbond)
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
	s.Require().Equal(hostVal1Addr, hostZone.Validators[0].Address)
	s.Require().Equal(hostVal2Addr, hostZone.Validators[2].Address)
	s.Require().Equal(hostVal3Addr, hostZone.Validators[1].Address)
}
