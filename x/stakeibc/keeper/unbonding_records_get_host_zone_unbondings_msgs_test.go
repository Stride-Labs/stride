package keeper_test

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

const (
	hostVal1Addr = "cosmos_VALIDATOR_1"
	hostVal2Addr = "cosmos_VALIDATOR_2"
	hostVal3Addr = "cosmos_VALIDATOR_3"
	amtVal1 = uint64(1_000_000)
	amtVal2 = uint64(2_000_000)
	wgtVal1 = uint64(1)
	wgtVal2 = uint64(2)
	amtToUnbond = uint64(1_000_000)
)

type GetHostZoneUnbondingMsgsTestCase struct {
	epochUnbondingRecords []recordtypes.EpochUnbondingRecord
	chainId              string
	validators             []*stakeibc.Validator
	totalWeight uint64
	expectErr	error
	expectPass	bool
}

var defaultUnbondingTestCase = GetHostZoneUnbondingMsgsTestCase{
	epochUnbondingRecords:         []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber:        0,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
		{
			EpochNumber:        1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
		},
	},
	chainId: "GAIA",
	validators: []*stakeibc.Validator{
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
	expectPass: true,
}

func (s *KeeperTestSuite) SetupGetHostZoneUnbondingMsgs(tc GetHostZoneUnbondingMsgsTestCase) (uint64, types.HostZone) {
	delegationAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "DELEGATION")
	s.CreateICAChannel(delegationAccountOwner)

	delegationAddr := "cosmos_DELEGATION"
	
	// 2022-08-12T19:51, a random time in the past
	unbondingTime := uint64(1660348276)

	//  define the host zone with stakedBal and validators with staked amounts
	validators := tc.validators

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
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	if tc.chainId != "" {
		hostZone.ChainId = tc.chainId
	}

	return amtToUnbond,hostZone
}

func (s *KeeperTestSuite) TestGetHostZoneUnbondingMsgs() {
	testCases := map[string]GetHostZoneUnbondingMsgsTestCase{
		"Successful": defaultUnbondingTestCase,
		"Wrong chain id": {
			epochUnbondingRecords:         []recordtypes.EpochUnbondingRecord{
				{
					EpochNumber:        0,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
				},
				{
					EpochNumber:        1,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
				},
			},
			chainId: "nonExistentChainId",
			validators: []*stakeibc.Validator{
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
			expectPass: false,
		},
		"No epoch unbonding records": {
			epochUnbondingRecords:         []recordtypes.EpochUnbondingRecord{},
			chainId: "nonExistentChainId",
			validators: []*stakeibc.Validator{
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
			expectPass: false,
		},
		"Unbonding too much": {
			epochUnbondingRecords:         []recordtypes.EpochUnbondingRecord{
				{
					EpochNumber:        0,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
				},
				{
					EpochNumber:        1,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
				},
			},
			chainId: "GAIA",
			validators: []*stakeibc.Validator{
				{
					Address:       hostVal1Addr,
					DelegationAmt: 0,
					Weight:        wgtVal1,
				},
				{
					Address:       hostVal2Addr,
					DelegationAmt: 0,
					Weight:        wgtVal2,
				},
				{
					Address: hostVal3Addr,
					// DelegationAmt and Weight are the same as Val2, to test tie breaking
					DelegationAmt: 0,
					Weight:        wgtVal2,
				},
			},
			expectPass: false,
			expectErr: sdkerrors.Wrap(sdkerrors.ErrNotFound, fmt.Sprintf("Could not unbond %d on Host Zone %s, unable to balance the unbond amount across validators", uint64(2_000_000), "GAIA")),
		},
		"No non-zero weight validator": {
			epochUnbondingRecords:         []recordtypes.EpochUnbondingRecord{
				{
					EpochNumber:        0,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
				},
				{
					EpochNumber:        1,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
				},
			},
			chainId: "GAIA",
			validators: []*stakeibc.Validator{
				{
					Address:       hostVal1Addr,
					DelegationAmt: amtVal1,
					Weight:        0,
				},
				{
					Address:       hostVal2Addr,
					DelegationAmt: amtVal2,
					Weight:        0,
				},
				{
					Address: hostVal3Addr,
					DelegationAmt: amtVal2,
					Weight:        0,
				},
			},
			expectPass: false,
			expectErr: sdkerrors.Wrap(types.ErrNoValidatorAmts, fmt.Sprintf("Error getting target val amts for host zone %s %d: no non-zero validator weights", "GAIA", uint64(2_000_000))),
		},
	}

	for name, test := range testCases {
		s.Run(name, func() {
			s.Setup()
			amtToUnbond, hostZone := s.SetupGetHostZoneUnbondingMsgs(test)
			actualUnbondMsgs, actualAmtToUnbond, actualCallbackArgs, _, err := s.App.StakeibcKeeper.GetHostZoneUnbondingMsgs(s.Ctx(), hostZone)
			
			if test.expectPass {
				s.Require().NoError(err)

				// verify the callback attributes are as expected
				actualCallbackResult, err := s.App.StakeibcKeeper.UnmarshalUndelegateCallbackArgs(s.Ctx(), actualCallbackArgs)
				s.Require().NoError(err, "could unmarshal undelegation callback args")
				s.Require().Equal(len(hostZone.Validators), len(actualCallbackResult.SplitDelegations), "number of split delegations in success unbonding case")
				s.Require().Equal(hostZone.ChainId, actualCallbackResult.HostZoneId, "host zone id in success unbonding case")

				// the number of unbonding messages should be (number of validators) * (records to unbond)
				s.Require().Equal(len(test.validators), len(actualUnbondMsgs), "number of unbonding messages should be number of records to unbond")

				s.Require().Equal(int64(amtToUnbond)*int64(len(test.epochUnbondingRecords)), int64(actualAmtToUnbond), "total amount to unbond should match input amtToUnbond")

				totalWgt := sdk.NewDec(int64(test.totalWeight))
				actualAmtToUnbondDec := sdk.NewDec(int64(actualAmtToUnbond))

				for i, validator := range test.validators {
					actualUnbondMsg := actualUnbondMsgs[i].String()
					valFraction := sdk.NewDec(int64(validator.Weight)).Quo(totalWgt)
					val1UnbondAmt := valFraction.Mul(actualAmtToUnbondDec).TruncateInt().String()
					valUnbonded := strings.Contains(actualUnbondMsg, val1UnbondAmt)
					// there's rounding in the logic that distributes stake amongst validators, so one or the other of the balances will be correct, depending on the rounding
					// at least one will be correct, and the other will be off by 1 by rounding, so we check and OR condition
					s.Require().True(valUnbonded, "unbonding amt should be the correct amount")
				}
			} else {
				if test.expectErr != nil {
					s.Require().EqualError(err, test.expectErr.Error())
				}
				// no messages should be sent
				s.Require().Equal(0, len(actualUnbondMsgs), "no messages should be sent")
				// no value should be unbonded
				s.Require().Equal(int64(0), int64(actualAmtToUnbond), "no value should be unbonded")
			}
		})
		
	}
}


func (s *KeeperTestSuite) TestGetTargetValAmtsForHostZone_Success() {
	_, hostZone := s.SetupGetHostZoneUnbondingMsgs(defaultUnbondingTestCase)

	// verify the total amount is expected
	unbond := uint64(1_000_000)
	totalAmt, err := s.App.StakeibcKeeper.GetTargetValAmtsForHostZone(s.Ctx(), hostZone, unbond)
	s.Require().Nil(err)

	// sum up totalAmt
	actualAmount := uint64(0)
	for _, amt := range totalAmt {
		actualAmount += amt
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
