package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v27/app/apptesting"
	"github.com/Stride-Labs/stride/v27/utils"
	epochtypes "github.com/Stride-Labs/stride/v27/x/epochs/types"
	recordtypes "github.com/Stride-Labs/stride/v27/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v27/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v27/x/staketia/types"
)

const DefaultClaimFundingAmount = 2600 // sum of NativeTokenAmount of records with status UNBONDED

// ----------------------------------------------
//               RedeemStake
// ----------------------------------------------

type Account struct {
	account      sdk.AccAddress
	stTokens     sdk.Coin
	nativeTokens sdk.Coin
}

type RedeemStakeTestCase struct {
	testName string

	userAccount      Account
	hostZone         *types.HostZone
	stakeibcHostZone *stakeibctypes.HostZone
	accUnbondRecord  *types.UnbondingRecord
	redemptionRecord *types.RedemptionRecord
	redeemMsg        types.MsgRedeemStake

	expectedUnbondingRecord            *types.UnbondingRecord
	expectedRedemptionRecord           *types.RedemptionRecord
	expectedStakeibcRedemptionAmount   sdkmath.Int
	expectedStakeibcRedemptionsEnabled bool
	expectedErrorContains              string
}

// Create the correct amounts in accounts, setup the records in store
func (s *KeeperTestSuite) SetupTestRedeemStake(
	userAccount Account,
	hostZone *types.HostZone,
	stakeibcHostZone *stakeibctypes.HostZone,
	accUnbondRecord *types.UnbondingRecord,
	redemptionRecord *types.RedemptionRecord,
) {
	s.FundAccount(userAccount.account, userAccount.nativeTokens)
	s.FundAccount(userAccount.account, userAccount.stTokens)

	if hostZone != nil {
		s.App.StaketiaKeeper.SetHostZone(s.Ctx, *hostZone)
	}

	if stakeibcHostZone != nil {
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, *stakeibcHostZone)
	}

	if accUnbondRecord != nil {
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, *accUnbondRecord)
	}

	if hostZone != nil && accUnbondRecord != nil &&
		accUnbondRecord.StTokenAmount.IsPositive() {
		escrowAccount, err := sdk.AccAddressFromBech32(hostZone.RedemptionAddress)
		if err == nil {
			stTokens := sdk.NewInt64Coin(StDenom, accUnbondRecord.StTokenAmount.Int64())
			s.FundAccount(escrowAccount, stTokens)
		}
	}

	if redemptionRecord != nil {
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, *redemptionRecord)
	}

	// Prepare stakeibc for a redemption
	epochNumber := uint64(1)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, stakeibctypes.EpochTracker{
		EpochIdentifier: epochtypes.DAY_EPOCH,
		EpochNumber:     epochNumber,
	})

	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:        types.CelestiaChainId,
				StTokenAmount:     sdkmath.ZeroInt(),
				NativeTokenAmount: sdkmath.ZeroInt(),
			},
		},
	})
}

// Default values for key variables, different tests will change 1-2 fields for setup
func (s *KeeperTestSuite) getDefaultTestInputs() (
	*Account,
	*types.HostZone,
	*stakeibctypes.HostZone,
	*types.UnbondingRecord,
	*types.RedemptionRecord,
	*types.MsgRedeemStake,
) {
	redeemerAccount := s.TestAccs[0]
	redemptionAccount := s.TestAccs[1]
	receiverAddress := s.TestAccs[2].String()
	depositAddress := s.TestAccs[3].String()

	defaultUserAccount := Account{
		account:      redeemerAccount,
		nativeTokens: sdk.NewInt64Coin(HostNativeDenom, 10_000_000),
		stTokens:     sdk.NewInt64Coin(StDenom, 10_000_000),
	}

	redemptionRate := sdk.MustNewDecFromStr("1.1")
	defaultStaketiaHostZone := types.HostZone{
		NativeTokenDenom:          HostNativeDenom,
		RedemptionAddress:         redemptionAccount.String(),
		RemainingDelegatedBalance: sdkmath.NewInt(1_000_000_000),
		Halted:                    false,
	}
	defaultStakeibcHostZone := stakeibctypes.HostZone{
		ChainId:                types.CelestiaChainId,
		HostDenom:              HostNativeDenom,
		Bech32Prefix:           "stride",
		DepositAddress:         depositAddress,
		RedemptionRate:         redemptionRate,
		MinRedemptionRate:      redemptionRate.Sub(sdk.MustNewDecFromStr("0.2")),
		MinInnerRedemptionRate: redemptionRate.Sub(sdk.MustNewDecFromStr("0.1")),
		MaxInnerRedemptionRate: redemptionRate.Add(sdk.MustNewDecFromStr("0.1")),
		MaxRedemptionRate:      redemptionRate.Add(sdk.MustNewDecFromStr("0.2")),
		TotalDelegations:       sdkmath.NewInt(1_000_000_000),
	}

	defaultAccUnbondingRecord := types.UnbondingRecord{
		Id:            uint64(105),
		Status:        types.ACCUMULATING_REDEMPTIONS,
		StTokenAmount: sdk.NewInt(700_000),
		NativeAmount:  sdk.NewInt(770_000),
	}

	// RR as it would exist for this default user and UnbondingRecord if had previously
	// performed an RedeemStake action this epoch for 400_000 stTokens
	defaultRedemptionRecord := types.RedemptionRecord{
		UnbondingRecordId: defaultAccUnbondingRecord.Id,
		Redeemer:          redeemerAccount.String(),
		StTokenAmount:     sdk.NewInt(400_000),
		NativeAmount:      sdk.NewInt(440_000),
	}

	defaultMsg := types.MsgRedeemStake{
		Redeemer:      redeemerAccount.String(),
		StTokenAmount: sdk.NewInt(1_000_000),
		Receiver:      receiverAddress,
	}

	return &defaultUserAccount, &defaultStaketiaHostZone, &defaultStakeibcHostZone,
		&defaultAccUnbondingRecord, &defaultRedemptionRecord, &defaultMsg
}

func (s *KeeperTestSuite) TestRedeemStake() {
	defaultUA, defaultMsHZ, defaultIcaHZ, defaultUR, defaultRR, defaultMsg := s.getDefaultTestInputs()

	testCases := []RedeemStakeTestCase{
		{
			testName: "[Error] Can't find the HostZone",

			userAccount:      *defaultUA,
			hostZone:         nil,
			stakeibcHostZone: defaultIcaHZ,

			expectedErrorContains: types.ErrHostZoneNotFound.Error(),
		},
		{
			testName: "[Error] Can't find the stakeibc HostZone",

			userAccount:      *defaultUA,
			hostZone:         defaultMsHZ,
			stakeibcHostZone: nil,

			expectedErrorContains: types.ErrHostZoneNotFound.Error(),
		},
		{
			testName: "[Error] Can't parse redemption address",

			userAccount: *defaultUA,
			hostZone: func() *types.HostZone {
				_, hz, _, _, _, _ := s.getDefaultTestInputs()
				hz.RedemptionAddress = "nonparsable-address"
				return hz
			}(),
			stakeibcHostZone: defaultIcaHZ,

			expectedErrorContains: "could not bech32 decode redemption address",
		},
		{
			testName: "[Error] HostZone is halted",

			userAccount: *defaultUA,
			hostZone: func() *types.HostZone {
				_, hz, _, _, _, _ := s.getDefaultTestInputs()
				hz.Halted = true
				return hz
			}(),
			stakeibcHostZone: defaultIcaHZ,

			expectedErrorContains: types.ErrHostZoneHalted.Error(),
		},
		{
			testName: "[Error] RedemptionRate outside of bounds",

			userAccount: *defaultUA,
			hostZone:    defaultMsHZ,
			stakeibcHostZone: func() *stakeibctypes.HostZone {
				_, _, hz, _, _, _ := s.getDefaultTestInputs()
				hz.RedemptionRate = sdk.MustNewDecFromStr("5.2")
				return hz
			}(),

			expectedErrorContains: types.ErrRedemptionRateOutsideSafetyBounds.Error(),
		},
		{
			testName: "[Error] No Accumulating UndondingRecord",

			userAccount:      *defaultUA,
			hostZone:         defaultMsHZ,
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  nil,

			expectedErrorContains: types.ErrBrokenUnbondingRecordInvariant.Error(),
		},
		{
			testName: "[Error] Not enough tokens in wallet",

			userAccount: func() Account {
				acc, _, _, _, _, _ := s.getDefaultTestInputs()
				acc.stTokens.Amount = sdk.NewInt(500_000)
				return *acc
			}(),
			hostZone:         defaultMsHZ,
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  defaultUR,
			redeemMsg:        *defaultMsg, // attempt to redeem 1_000_000 stTokens

			expectedErrorContains: sdkerrors.ErrInsufficientFunds.Error(),
		},
		{
			testName: "[Error] Redeeming more than HostZone delegation total",

			userAccount: func() Account {
				acc, _, _, _, _, _ := s.getDefaultTestInputs()
				acc.stTokens.Amount = sdk.NewInt(5_000_000_000)
				return *acc
			}(),
			hostZone:         defaultMsHZ, // 1_000_000_000 total delegation
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  defaultUR,
			redeemMsg: func() types.MsgRedeemStake {
				_, _, _, _, _, msg := s.getDefaultTestInputs()
				msg.StTokenAmount = sdk.NewInt(5_000_000_000)
				return *msg
			}(),

			expectedErrorContains: types.ErrUnbondAmountToLarge.Error(),
		},
		{
			testName: "[Error] Redeeming is disabled",

			userAccount: *defaultUA,
			hostZone: func() *types.HostZone {
				_, hz, _, _, _, _ := s.getDefaultTestInputs()
				hz.RemainingDelegatedBalance = sdkmath.ZeroInt()
				return hz
			}(),
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  defaultUR,
			redeemMsg:        *defaultMsg,

			expectedErrorContains: types.ErrRedemptionsDisabled.Error(),
		},
		{
			testName: "[Error] Spillover fails in stakeibc",

			userAccount: *defaultUA,
			hostZone: func() *types.HostZone {
				_, hz, _, _, _, msg := s.getDefaultTestInputs()
				hz.RemainingDelegatedBalance = msg.StTokenAmount.Sub(sdkmath.OneInt()) // subtract 1 so there's excess
				return hz
			}(),
			stakeibcHostZone: func() *stakeibctypes.HostZone {
				_, _, hz, _, _, _ := s.getDefaultTestInputs()
				hz.RedemptionRate = sdk.OneDec()
				return hz
			}(),
			accUnbondRecord: defaultUR,
			redeemMsg: func() types.MsgRedeemStake {
				// Make receiver address invalid so stakeibc fails
				_, _, _, _, _, msg := s.getDefaultTestInputs()
				msg.Receiver = ""
				return *msg
			}(),

			expectedErrorContains: "unable to execute stakeibc redeem stake",
		},
		{
			testName: "[Success] No RR exists yet, RedeemStake tx creates one",

			userAccount:      *defaultUA,
			hostZone:         defaultMsHZ,
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  defaultUR,
			redemptionRecord: nil,
			redeemMsg:        *defaultMsg, // redeem 1_000_000 stTokens

			expectedUnbondingRecord: func() *types.UnbondingRecord {
				_, _, hz, ur, _, msg := s.getDefaultTestInputs()
				ur.StTokenAmount = ur.StTokenAmount.Add(msg.StTokenAmount)
				nativeDiff := sdk.NewDecFromInt(msg.StTokenAmount).Mul(hz.RedemptionRate).TruncateInt()
				ur.NativeAmount = ur.NativeAmount.Add(nativeDiff)
				return ur
			}(),
			expectedRedemptionRecord: &types.RedemptionRecord{
				UnbondingRecordId: defaultUR.Id,
				Redeemer:          defaultMsg.Redeemer,
				StTokenAmount:     defaultMsg.StTokenAmount,
				NativeAmount:      sdk.NewDecFromInt(defaultMsg.StTokenAmount).Mul(defaultIcaHZ.RedemptionRate).TruncateInt(),
			},
		},
		{
			testName: "[Success] RR exists already for redeemer, RedeemStake tx updates",

			userAccount:      *defaultUA,
			hostZone:         defaultMsHZ,
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  defaultUR,
			redemptionRecord: defaultRR,   // previous redeemption of 400_000
			redeemMsg:        *defaultMsg, // redeem 1_000_000 stTokens

			expectedUnbondingRecord: func() *types.UnbondingRecord {
				_, _, hz, ur, _, msg := s.getDefaultTestInputs()
				ur.StTokenAmount = ur.StTokenAmount.Add(msg.StTokenAmount)
				nativeDiff := sdk.NewDecFromInt(msg.StTokenAmount).Mul(hz.RedemptionRate).TruncateInt()
				ur.NativeAmount = ur.NativeAmount.Add(nativeDiff)
				return ur
			}(),
			expectedRedemptionRecord: func() *types.RedemptionRecord {
				_, _, hz, _, rr, msg := s.getDefaultTestInputs()
				rr.StTokenAmount = rr.StTokenAmount.Add(msg.StTokenAmount)
				nativeDiff := sdk.NewDecFromInt(msg.StTokenAmount).Mul(hz.RedemptionRate).TruncateInt()
				rr.NativeAmount = rr.NativeAmount.Add(nativeDiff)
				return rr
			}(),
		},
		{
			testName: "[Success] Redeems all remaining balance",

			userAccount: *defaultUA,
			hostZone: func() *types.HostZone {
				_, msHz, icaHz, _, _, msg := s.getDefaultTestInputs()
				nativeRedeemAmount := sdk.NewDecFromInt(msg.StTokenAmount).Mul(icaHz.RedemptionRate).TruncateInt()
				msHz.RemainingDelegatedBalance = nativeRedeemAmount
				return msHz
			}(),
			stakeibcHostZone: defaultIcaHZ,
			accUnbondRecord:  defaultUR,
			redeemMsg:        *defaultMsg,

			expectedUnbondingRecord: func() *types.UnbondingRecord {
				_, _, hz, ur, _, msg := s.getDefaultTestInputs()
				ur.StTokenAmount = ur.StTokenAmount.Add(msg.StTokenAmount)
				nativeDiff := sdk.NewDecFromInt(msg.StTokenAmount).Mul(hz.RedemptionRate).TruncateInt()
				ur.NativeAmount = ur.NativeAmount.Add(nativeDiff)
				return ur
			}(),
			expectedRedemptionRecord: &types.RedemptionRecord{
				UnbondingRecordId: defaultUR.Id,
				Redeemer:          defaultMsg.Redeemer,
				StTokenAmount:     defaultMsg.StTokenAmount,
				NativeAmount:      sdk.NewDecFromInt(defaultMsg.StTokenAmount).Mul(defaultIcaHZ.RedemptionRate).TruncateInt(),
			},
			expectedStakeibcRedemptionsEnabled: true,
		},
		{
			testName: "[Success] Redeems with stakeibc spillover",

			userAccount: *defaultUA,
			hostZone: func() *types.HostZone {
				_, hz, _, _, _, msg := s.getDefaultTestInputs()
				hz.RemainingDelegatedBalance = msg.StTokenAmount.Sub(sdkmath.OneInt()) // subtract 1 so there's excess
				return hz
			}(),
			stakeibcHostZone: func() *stakeibctypes.HostZone {
				_, _, hz, _, _, _ := s.getDefaultTestInputs()
				hz.RedemptionRate = sdk.OneDec()
				return hz
			}(),
			accUnbondRecord: defaultUR,
			redeemMsg:       *defaultMsg,

			expectedUnbondingRecord: func() *types.UnbondingRecord {
				_, _, _, ur, _, msg := s.getDefaultTestInputs()
				ur.StTokenAmount = ur.StTokenAmount.Add(msg.StTokenAmount).Sub(sdkmath.OneInt())
				ur.NativeAmount = ur.NativeAmount.Add(msg.StTokenAmount).Sub(sdkmath.OneInt())
				return ur
			}(),
			expectedRedemptionRecord: &types.RedemptionRecord{
				UnbondingRecordId: defaultUR.Id,
				Redeemer:          defaultMsg.Redeemer,
				StTokenAmount:     defaultMsg.StTokenAmount.Sub(sdkmath.OneInt()),
				NativeAmount:      defaultMsg.StTokenAmount.Sub(sdkmath.OneInt()),
			},
			expectedStakeibcRedemptionAmount:   sdkmath.OneInt(),
			expectedStakeibcRedemptionsEnabled: true,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.testName, func() {
			s.checkRedeemStakeTestCase(tc)
		})
	}

}

func (s *KeeperTestSuite) checkRedeemStakeTestCase(tc RedeemStakeTestCase) {
	s.SetupTest() // reset state
	s.SetupTestRedeemStake(tc.userAccount, tc.hostZone, tc.stakeibcHostZone, tc.accUnbondRecord, tc.redemptionRecord)

	startingStEscrowBalance := sdk.NewInt64Coin(StDenom, 0)
	if tc.hostZone != nil {
		escrowAccount, err := sdk.AccAddressFromBech32(tc.hostZone.RedemptionAddress)
		if err == nil {
			startingStEscrowBalance = s.App.BankKeeper.GetBalance(s.Ctx, escrowAccount, StDenom)
		}
	}

	// Run the RedeemStake, verify expected errors returned or no errors with expected updates to records
	_, err := s.App.StaketiaKeeper.RedeemStake(s.Ctx, tc.redeemMsg.Redeemer, tc.redeemMsg.Receiver, tc.redeemMsg.StTokenAmount)
	if tc.expectedErrorContains == "" {
		// Successful Run Test Case
		s.Require().NoError(err, "No error expected during redeem stake execution")

		// Check the remaining delegated amount was decremented (and should never go below 0)
		hostZone := s.MustGetHostZone()
		redeemAmount := sdk.NewDecFromInt(tc.redeemMsg.StTokenAmount).Mul(tc.stakeibcHostZone.RedemptionRate).TruncateInt()
		expectedRemainingDelegation := tc.hostZone.RemainingDelegatedBalance.Sub(redeemAmount)
		expectedRemainingDelegation = sdkmath.MaxInt(expectedRemainingDelegation, sdkmath.ZeroInt())
		s.Require().Equal(expectedRemainingDelegation.Int64(), hostZone.RemainingDelegatedBalance.Int64(), "remaining delegated balance")

		// check expected updates to Accumulating UnbondingRecord
		currentAUR, err := s.App.StaketiaKeeper.GetAccumulatingUnbondingRecord(s.Ctx)
		s.Require().NoError(err, "No error expected when getting UnbondingRecord")
		s.Require().Equal(*tc.expectedUnbondingRecord, currentAUR, "Accumulating UnbondingRecord did not match expected")

		// check expected updates to RedemptionRecord for this user and current UnbondingRecord
		currentRR, found := s.App.StaketiaKeeper.GetRedemptionRecord(s.Ctx, currentAUR.Id, tc.redeemMsg.Redeemer)
		s.Require().True(found, "No RedemptionRecord found after RedeemStake expected to have created one")
		s.Require().Equal(*tc.expectedRedemptionRecord, currentRR, "RedemptionRecord did not match expected")

		// In test setup the escrow acc was funded with the number of tokens on starting accumulating UnbondingRecord
		// Verify that the redemption account now holds the increased escrowed stTokens matching final UnbondingRecord
		escrowAccount, err := sdk.AccAddressFromBech32(tc.hostZone.RedemptionAddress)
		s.Require().NoError(err, "No error expected when getting escrow account for successful test")
		currentStEscrowBalance := s.App.BankKeeper.GetBalance(s.Ctx, escrowAccount, StDenom)
		s.Require().NotEqual(startingStEscrowBalance, currentStEscrowBalance, "Escrowed balance should have changed")
		s.Require().Equal(currentStEscrowBalance.Amount, currentAUR.StTokenAmount, "Escrowed balance does not match the UnbondingRecord")

		// If all the remaining amount was unbonded, check that redemptions were enabled
		stakeibcHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, types.CelestiaChainId)
		s.Require().True(found)
		s.Require().Equal(tc.expectedStakeibcRedemptionsEnabled, stakeibcHostZone.RedemptionsEnabled, "stakeibc redemptions enabled")

		// If there's stakeibc spillover, check that the amount was stored in the user redemption record
		if !tc.expectedStakeibcRedemptionAmount.IsNil() {
			userRedemptionRecords := s.App.RecordsKeeper.GetAllUserRedemptionRecord(s.Ctx)
			s.Require().Len(userRedemptionRecords, 1, "there should be a stakeibc user redemption record")
			s.Require().Equal(tc.expectedStakeibcRedemptionAmount.Int64(), userRedemptionRecords[0].StTokenAmount.Int64(),
				"stakeibc user redemption record amount")
		}
	} else {
		// Expected Error Test Case
		s.Require().Error(err, "Error expected to be returned but none found")
		s.Require().ErrorContains(err, tc.expectedErrorContains, "Error did not contain expected message")
	}
}

// ----------------------------------------------
//             PrepareUndelegation
// ----------------------------------------------

func (s *KeeperTestSuite) TestPrepareUndelegation() {
	accumulatingUnbondingRecordId := uint64(4)
	epochNumber := uint64(5)

	// Create the accumulating unbonding record
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, types.UnbondingRecord{
		Id:     accumulatingUnbondingRecordId,
		Status: types.ACCUMULATING_REDEMPTIONS,
	})

	// Set a host zone with a 1.999 redemption rate
	// (an uneven number is used to test rounding/truncation)
	oldRedemptionRate := sdk.MustNewDecFromStr("1.9")
	redemptionRate := sdk.MustNewDecFromStr("1.999")
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:        types.CelestiaChainId,
		RedemptionRate: redemptionRate,
	})

	// Define the expected redemption records after the function is called
	expectedRedemptionRecords := []types.RedemptionRecord{
		// StTokenAmount: 1000 * 1.999 = 1999 Native
		{UnbondingRecordId: 4, Redeemer: "A", StTokenAmount: sdkmath.NewInt(1000), NativeAmount: sdkmath.NewInt(1999)},
		// StTokenAmount: 999 * 1.999 = 1997.001, Truncated to 1997 Native
		{UnbondingRecordId: 4, Redeemer: "B", StTokenAmount: sdkmath.NewInt(999), NativeAmount: sdkmath.NewInt(1997)},
		// StTokenAmount: 100 * 1.999 = 199.9, Truncated to 199 Native
		{UnbondingRecordId: 4, Redeemer: "C", StTokenAmount: sdkmath.NewInt(100), NativeAmount: sdkmath.NewInt(199)},

		// Different unbonding records, should be excluded
		{UnbondingRecordId: 1, Redeemer: "D", StTokenAmount: sdkmath.NewInt(100), NativeAmount: sdkmath.NewInt(100)},
		{UnbondingRecordId: 2, Redeemer: "E", StTokenAmount: sdkmath.NewInt(200), NativeAmount: sdkmath.NewInt(200)},
		{UnbondingRecordId: 3, Redeemer: "F", StTokenAmount: sdkmath.NewInt(300), NativeAmount: sdkmath.NewInt(300)},
	}
	expectedTotalNativeAmount := sdkmath.NewInt(1999 + 1997 + 199)

	// Create the initial records, setting the native amount to be slightly less than expected
	for _, expectedUserRedemptionRecord := range expectedRedemptionRecords {
		initialRedemptionRecord := expectedUserRedemptionRecord
		initialRedemptionRecord.NativeAmount = sdk.NewDecFromInt(initialRedemptionRecord.StTokenAmount).Mul(oldRedemptionRate).TruncateInt()
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, initialRedemptionRecord)
	}

	// Call prepare undelegation
	err := s.App.StaketiaKeeper.PrepareUndelegation(s.Ctx, epochNumber)
	s.Require().NoError(err, "no error expected when calling prepare undelegation")

	// Confirm the total and status was updated on the unbonding record
	unbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, accumulatingUnbondingRecordId)
	s.Require().True(found)
	s.Require().Equal(unbondingRecord.Status, types.UNBONDING_QUEUE, "unbonding record status should have updated")
	s.Require().Equal(expectedTotalNativeAmount.Int64(), unbondingRecord.NativeAmount.Int64(),
		"total native tokens on unbonding record")

	// Confirm the summation is correct and the redemption records were updated
	for _, expectedRecord := range expectedRedemptionRecords {
		if expectedRecord.UnbondingRecordId != accumulatingUnbondingRecordId {
			continue
		}

		unbondingRecordId := expectedRecord.UnbondingRecordId
		redeemer := expectedRecord.Redeemer
		actualRecord, found := s.App.StaketiaKeeper.GetRedemptionRecord(s.Ctx, unbondingRecordId, redeemer)
		s.Require().True(found, "record %d %s should have been found", unbondingRecordId, redeemer)
		s.Require().Equal(expectedRecord.NativeAmount.Int64(), actualRecord.NativeAmount.Int64(),
			"record %s %d native amount", unbondingRecordId, redeemer)
	}

	// Confirm a new record was created
	newUnbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, epochNumber)
	s.Require().True(found, "new unbonding record should have been created")
	s.Require().Equal(newUnbondingRecord.Status, types.ACCUMULATING_REDEMPTIONS, "new unbonding record status")

	// Call prepare undelegation again with the new unbonding record
	// Since there are no new unbondings, the record should get archived immediately
	err = s.App.StaketiaKeeper.PrepareUndelegation(s.Ctx, epochNumber+1)
	s.Require().NoError(err, "no error expected during second undelegation")

	archivedRecords := s.App.StaketiaKeeper.GetAllArchivedUnbondingRecords(s.Ctx)
	s.Require().Len(archivedRecords, 1, "record should have been archived")
	s.Require().Equal(epochNumber, archivedRecords[0].Id, "archived record ID")

	// Create an unbonding record in non-ACCUMULATING Status
	duplicateRecordId := uint64(10)
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, types.UnbondingRecord{
		Id:     duplicateRecordId,
		Status: types.UNBONDING_QUEUE,
	})

	// Check that if we tried to run prepare with that ID, it would error because the record already exists
	err = s.App.StaketiaKeeper.PrepareUndelegation(s.Ctx, duplicateRecordId)
	s.Require().ErrorContains(err, "unbonding record already exists")

	// Create another accumulating record and check that this would break an invariant and error
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, types.UnbondingRecord{
		Id:     99,
		Status: types.ACCUMULATING_REDEMPTIONS,
	})

	err = s.App.StaketiaKeeper.PrepareUndelegation(s.Ctx, epochNumber+10) // any epoch in future
	s.Require().ErrorContains(err, "more than one record in status ACCUMULATING")

	// Halt the host zone and try again - it should fail
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, types.CelestiaChainId)
	s.Require().True(found)
	hostZone.Halted = true
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	err = s.App.StaketiaKeeper.PrepareUndelegation(s.Ctx, epochNumber)
	s.Require().ErrorContains(err, "host zone celestia is halted")
}

// ----------------------------------------------
//             ConfirmUndelegation
// ----------------------------------------------

type ConfirmUndelegationTestCase struct {
	operatorAddress          string
	redemptionAddress        sdk.AccAddress
	stTokenAmountToBurn      sdkmath.Int
	unbondingRecord          types.UnbondingRecord
	amountToUndelegate       sdkmath.Int
	delegatedBalance         sdkmath.Int
	redemptionAccountBalance sdkmath.Int
	staketiaHostZone         types.HostZone
	stakeibcHostZone         stakeibctypes.HostZone
	expectedUnbondingTime    uint64
}

// Helper function to mock relevant state before testing a confirm undelegation
func (s *KeeperTestSuite) SetupTestConfirmUndelegation(amountToUndelegate sdkmath.Int) ConfirmUndelegationTestCase {
	redemptionAddress := s.TestAccs[0]
	operatorAddress := s.TestAccs[1]

	redemptionAccountBalance := sdkmath.NewInt(500)
	delegatedBalance := sdkmath.NewInt(1000)
	unbondingPeriodSeconds := uint64(120) // 2 minutes
	expectedUnbondingTime := uint64(s.Ctx.BlockTime().Add(time.Minute * 2).Unix())

	// Create a host zone with delegatedBalance and RedemptionAddresses
	staketiaHostZone := types.HostZone{
		RedemptionAddress:      redemptionAddress.String(),
		NativeTokenDenom:       HostNativeDenom,
		UnbondingPeriodSeconds: unbondingPeriodSeconds,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, staketiaHostZone)

	stakeibcHostZone := stakeibctypes.HostZone{
		ChainId:          types.CelestiaChainId,
		RedemptionRate:   sdk.MustNewDecFromStr("1.1"),
		TotalDelegations: delegatedBalance,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibcHostZone)

	// Fund the redemption account with tokens that will be burned
	stTokensInRedemption := sdk.NewCoin(StDenom, redemptionAccountBalance)
	s.FundAccount(redemptionAddress, stTokensInRedemption)

	// create an unbonding record in status UNBONDING_QUEUE
	// - stToken amount to burn as if the RR is 1.1
	stTokenAmountToBurn := sdk.NewDecFromInt(amountToUndelegate).Mul(stakeibcHostZone.RedemptionRate).TruncateInt()
	unbondingRecord := types.UnbondingRecord{
		Id:            1,
		Status:        types.UNBONDING_QUEUE,
		StTokenAmount: stTokenAmountToBurn,
		NativeAmount:  amountToUndelegate,
	}
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)

	tc := ConfirmUndelegationTestCase{
		redemptionAddress:        redemptionAddress,
		operatorAddress:          operatorAddress.String(),
		stTokenAmountToBurn:      stTokenAmountToBurn,
		unbondingRecord:          unbondingRecord,
		amountToUndelegate:       amountToUndelegate,
		delegatedBalance:         delegatedBalance,
		redemptionAccountBalance: redemptionAccountBalance,
		stakeibcHostZone:         stakeibcHostZone,
		staketiaHostZone:         staketiaHostZone,
		expectedUnbondingTime:    expectedUnbondingTime,
	}
	return tc
}

func (s *KeeperTestSuite) TestConfirmUndelegation_Success() {
	amountToUndelegate := sdkmath.NewInt(100)
	tc := s.SetupTestConfirmUndelegation(amountToUndelegate)

	// we're halting the zone to test that the tx works even when the host zone is halted
	s.App.StaketiaKeeper.HaltZone(s.Ctx)

	// confirm the tx was successful
	err := s.App.StaketiaKeeper.ConfirmUndelegation(s.Ctx, tc.unbondingRecord.Id, ValidTxHashDefault, tc.operatorAddress)
	s.Require().NoError(err)

	// check that the unbonding record was updated
	unbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, tc.unbondingRecord.Id)
	s.Require().True(found, "unbonding record should have been found")
	s.Require().Equal(types.UNBONDING_IN_PROGRESS, unbondingRecord.Status, "unbonding record status")
	s.Require().Equal(ValidTxHashDefault, unbondingRecord.UndelegationTxHash, "unbonding record tx hash")
	s.Require().Equal(tc.expectedUnbondingTime, unbondingRecord.UnbondingCompletionTimeSeconds, "unbonding record completion time")

	// check that the balance on the redemption account was updated
	expectedRedemptionAccountBalance := tc.redemptionAccountBalance.Sub(tc.stTokenAmountToBurn)
	actualRedemptionAccountBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.redemptionAddress, StDenom).Amount
	s.Require().Equal(expectedRedemptionAccountBalance, actualRedemptionAccountBalance, "redemption account balance")

	// check that delegated balance was updated
	stakeibcHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, types.CelestiaChainId)
	s.Require().True(found)
	s.Require().Equal(tc.delegatedBalance.Sub(tc.amountToUndelegate), stakeibcHostZone.TotalDelegations, "delegated balance")
}

// unit test ConfirmUndelegation with nothing to unbond
func (s *KeeperTestSuite) TestConfirmUndelegation_Failure_NothingToUnbond() {
	// test undelegating nothing
	amountToUndelegate := sdkmath.ZeroInt()
	tc := s.SetupTestConfirmUndelegation(amountToUndelegate)

	// require both stTokenAmountToBurn and amountToUndelegate are 0
	s.Require().Zero(tc.stTokenAmountToBurn.Int64())
	s.Require().Zero(tc.amountToUndelegate.Int64())

	txHash := "" // hash must be empty for nothing to unbond case
	err := s.App.StaketiaKeeper.ConfirmUndelegation(s.Ctx, tc.unbondingRecord.Id, txHash, tc.operatorAddress)
	s.Require().Error(err, "no tokens to unbond")
}

// unit test ConfirmUndelegation with nothing to unbond
func (s *KeeperTestSuite) TestConfirmUndelegation_Failure_NegativeAmountToUnbond() {
	// test undelegating nothing
	amountToUndelegateNegative := sdkmath.ZeroInt().Sub(sdkmath.NewInt(1))
	tc := s.SetupTestConfirmUndelegation(amountToUndelegateNegative)

	// require both stTokenAmountToBurn and amountToUndelegate are -1
	s.Require().Negative(tc.stTokenAmountToBurn.Int64())
	s.Require().Negative(tc.amountToUndelegate.Int64())

	txHash := "" // hash must be empty for nothing to unbond case
	err := s.App.StaketiaKeeper.ConfirmUndelegation(s.Ctx, tc.unbondingRecord.Id, txHash, tc.operatorAddress)
	s.Require().Error(err, "no tokens to unbond (or negative)")
}

func (s *KeeperTestSuite) TestConfirmUndelegation_Failure_RecordWasNotQueued() {
	amountToUndelegate := sdkmath.NewInt(100)
	tc := s.SetupTestConfirmUndelegation(amountToUndelegate)

	// set the unbonding record status to something other than UNBONDING_QUEUE
	tc.unbondingRecord.Status = types.UNBONDING_IN_PROGRESS
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, tc.unbondingRecord)

	err := s.App.StaketiaKeeper.ConfirmUndelegation(s.Ctx, tc.unbondingRecord.Id, ValidTxHashDefault, tc.operatorAddress)
	s.Require().Error(err, "not ready to be undelegated")
}

func (s *KeeperTestSuite) TestConfirmUndelegation_Failure_NoRecordWithId() {
	amountToUndelegate := sdkmath.NewInt(100)
	tc := s.SetupTestConfirmUndelegation(amountToUndelegate)

	// archive the record (this is the only way to remove it from the active store)
	tc.unbondingRecord.Status = types.UNBONDING_IN_PROGRESS
	s.App.StaketiaKeeper.ArchiveUnbondingRecord(s.Ctx, tc.unbondingRecord)

	err := s.App.StaketiaKeeper.ConfirmUndelegation(s.Ctx, tc.unbondingRecord.Id, ValidTxHashDefault, tc.operatorAddress)
	s.Require().Error(err, "couldn't find unbonding record")
}

func (s *KeeperTestSuite) TestConfirmUndelegation_Failure_RecordHashAlreadySet() {
	amountToUndelegate := sdkmath.NewInt(100)
	tc := s.SetupTestConfirmUndelegation(amountToUndelegate)

	// set the unbonding record tx hash
	tc.unbondingRecord.UndelegationTxHash = ValidTxHashDefault
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, tc.unbondingRecord)

	err := s.App.StaketiaKeeper.ConfirmUndelegation(s.Ctx, tc.unbondingRecord.Id, ValidTxHashDefault, tc.operatorAddress)
	s.Require().Error(err, "already has a tx hash set")
}

func (s *KeeperTestSuite) TestBurnRedeemedStTokens_Success() {

	redemptionAddress := s.TestAccs[0]
	redemptionAccountBalance := sdkmath.NewInt(500)

	// Create a host zone with delegatedBalance and RedemptionAddresses
	hostZone := types.HostZone{
		RedemptionAddress: redemptionAddress.String(),
		NativeTokenDenom:  HostNativeDenom,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// Fund the redemption account with tokens that will be burned
	stTokensInRedemption := sdk.NewCoin(StDenom, redemptionAccountBalance)
	s.FundAccount(redemptionAddress, stTokensInRedemption)

	// Store down the stToken supply for checking against after burn
	stTokenSupplyBefore := s.App.BankKeeper.GetSupply(s.Ctx, StDenom).Amount

	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)

	// burn redeemed stTokens
	tokensToBurn := sdk.NewCoin(stDenom, sdkmath.NewInt(100))
	err := s.App.StaketiaKeeper.BurnRedeemedStTokens(s.Ctx, sdk.NewCoins(tokensToBurn), redemptionAddress.String())
	s.Require().NoError(err)

	// check that stTIA supply decremented
	stTokenSupplyAfter := s.App.BankKeeper.GetSupply(s.Ctx, StDenom).Amount
	s.Require().Equal(stTokenSupplyBefore.Sub(tokensToBurn.Amount), stTokenSupplyAfter, "stToken supply")

	// check that the balance on the redemption account was updated
	expectedRedemptionAccountBalance := redemptionAccountBalance.Sub(tokensToBurn.Amount)
	actualRedemptionAccountBalance := s.App.BankKeeper.GetBalance(s.Ctx, redemptionAddress, StDenom).Amount
	s.Require().Equal(expectedRedemptionAccountBalance, actualRedemptionAccountBalance, "redemption account balance")
}

func (s *KeeperTestSuite) TestBurnRedeemedStTokens_BadRedemptionAddress() {

	redemptionAddress := "INVALID_ADDRESS"

	// Create a host zone with delegatedBalance and RedemptionAddresses
	hostZone := types.HostZone{
		RedemptionAddress: redemptionAddress,
		NativeTokenDenom:  HostNativeDenom,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)

	// burn redeemed stTokens
	tokensToBurn := sdk.NewCoin(stDenom, sdkmath.NewInt(100))
	err := s.App.StaketiaKeeper.BurnRedeemedStTokens(s.Ctx, sdk.NewCoins(tokensToBurn), redemptionAddress)
	s.Require().Error(err, "could not bech32 decode addres")
}

// ----------------------------------------------
//          ConfirmUnbondedTokensSwept
// ----------------------------------------------

func (s *KeeperTestSuite) GetDefaultUnbondingRecords() []types.UnbondingRecord {
	unbondingRecords := []types.UnbondingRecord{ // verify no issue if these are out of order
		{
			Id:                             1,
			Status:                         types.UNBONDING_QUEUE,
			StTokenAmount:                  sdk.NewInt(100),
			NativeAmount:                   sdk.NewInt(200),
			UnbondingCompletionTimeSeconds: 0,
			UndelegationTxHash:             "",
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             7,
			Status:                         types.CLAIMABLE,
			StTokenAmount:                  sdk.NewInt(200),
			NativeAmount:                   sdk.NewInt(400),
			UnbondingCompletionTimeSeconds: 10,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       ValidTxHashDefault,
		},
		{
			Id:                             5,
			Status:                         types.UNBONDING_IN_PROGRESS,
			StTokenAmount:                  sdk.NewInt(500),
			NativeAmount:                   sdk.NewInt(1000),
			UnbondingCompletionTimeSeconds: 20,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             3,
			Status:                         types.ACCUMULATING_REDEMPTIONS,
			StTokenAmount:                  sdk.NewInt(300),
			NativeAmount:                   sdk.NewInt(600),
			UnbondingCompletionTimeSeconds: 0,
			UndelegationTxHash:             "",
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             6,
			Status:                         types.UNBONDED,
			StTokenAmount:                  sdk.NewInt(600),
			NativeAmount:                   sdk.NewInt(1200),
			UnbondingCompletionTimeSeconds: 15,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       "",
		},
		{
			Id:                             4,
			Status:                         types.CLAIMABLE,
			StTokenAmount:                  sdk.NewInt(400),
			NativeAmount:                   sdk.NewInt(800),
			UnbondingCompletionTimeSeconds: 5,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       ValidTxHashDefault,
		},
		{
			Id:                             2,
			Status:                         types.UNBONDED,
			StTokenAmount:                  sdk.NewInt(700),
			NativeAmount:                   sdk.NewInt(1400),
			UnbondingCompletionTimeSeconds: 18,
			UndelegationTxHash:             ValidTxHashDefault,
			UnbondedTokenSweepTxHash:       "",
		},
	}
	return unbondingRecords
}

// Helper function to setup unbonding records, returns a list of records
func (s *KeeperTestSuite) SetupTestConfirmUnbondingTokens(amount int64) {
	unbondingRecords := s.GetDefaultUnbondingRecords()

	// loop through and set each record
	for _, unbondingRecord := range unbondingRecords {
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}

	// setup host zone, to fund claim address
	claimAddress := s.TestAccs[0]
	hostZone := types.HostZone{
		NativeTokenIbcDenom: HostIBCDenom,
		ClaimAddress:        claimAddress.String(),
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// fund claim address
	liquidStakeToken := sdk.NewCoin(hostZone.NativeTokenIbcDenom, sdk.NewInt(amount))
	s.FundAccount(claimAddress, liquidStakeToken)
}

func (s *KeeperTestSuite) VerifyUnbondingRecordsAfterConfirmSweep(verifyUpdatedFieldsIdentical bool) {
	defaultUnbondingRecords := s.GetDefaultUnbondingRecords()
	for _, defaultUnbondingRecord := range defaultUnbondingRecords {
		// grab relevant record in store
		loadedUnbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, defaultUnbondingRecord.Id)
		s.Require().True(found)

		// verify record is correct
		s.Require().Equal(defaultUnbondingRecord.Id, loadedUnbondingRecord.Id)
		s.Require().Equal(defaultUnbondingRecord.NativeAmount, loadedUnbondingRecord.NativeAmount)
		s.Require().Equal(defaultUnbondingRecord.StTokenAmount, loadedUnbondingRecord.StTokenAmount)
		s.Require().Equal(defaultUnbondingRecord.UnbondingCompletionTimeSeconds, loadedUnbondingRecord.UnbondingCompletionTimeSeconds)
		s.Require().Equal(defaultUnbondingRecord.UndelegationTxHash, loadedUnbondingRecord.UndelegationTxHash)

		// if relevant, verify status and tx hash
		if (defaultUnbondingRecord.Status != types.UNBONDED) ||
			verifyUpdatedFieldsIdentical {
			s.Require().Equal(defaultUnbondingRecord.Status, loadedUnbondingRecord.Status)
			s.Require().Equal(defaultUnbondingRecord.UnbondedTokenSweepTxHash, loadedUnbondingRecord.UnbondedTokenSweepTxHash)
		}
	}
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_Successful() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	// we're halting the zone to test that the tx works even when the host zone is halted
	s.App.StaketiaKeeper.HaltZone(s.Ctx)

	// process record 6
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyUnbondingRecordsAfterConfirmSweep(false)

	// verify record 6 modified
	loadedUnbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, 6)
	s.Require().True(found)
	s.Require().Equal(types.CLAIMABLE, loadedUnbondingRecord.Status, "unbonding record should be updated to status CLAIMABLE")
	s.Require().Equal(ValidTxHashNew, loadedUnbondingRecord.UnbondedTokenSweepTxHash, "unbonding record should be updated with token sweep txHash")

	// process record 2
	err = s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 2, ValidTxHashNew, ValidOperator)
	s.Require().NoError(err)
	s.VerifyUnbondingRecordsAfterConfirmSweep(false)

	// verify record 2 modified
	loadedUnbondingRecord, found = s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, 2)
	s.Require().True(found)
	s.Require().Equal(types.CLAIMABLE, loadedUnbondingRecord.Status, "unbonding record should be updated to status CLAIMABLE")
	s.Require().Equal(ValidTxHashNew, loadedUnbondingRecord.UnbondedTokenSweepTxHash, "unbonding record should be updated with token sweep txHash")

}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_NotFunded() {
	s.SetupTestConfirmUnbondingTokens(10)

	// try setting with no hash
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrInsufficientFunds, "should error when claim account doesn't have enough funds")
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_InvalidClaimAddress() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	hostZone := s.MustGetHostZone()
	hostZone.ClaimAddress = "strideinvalidaddress" // random address
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// try setting with no hash
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorContains(err, "decoding bech32 failed", "should error when claim address is invalid")
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_RecordDoesntExist() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	// process record 15
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 15, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrUnbondingRecordNotFound, "should error when record doesn't exist")
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_RecordIncorrectState() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	// get list of ids to try
	ids := []uint64{1, 3, 5, 7}
	for _, id := range ids {
		err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, id, ValidTxHashNew, ValidOperator)
		s.Require().ErrorIs(err, types.ErrInvalidUnbondingRecord, "should error when record is in incorrect state")
	}
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_ZeroSweepAmount() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	// update the sweep record so that the native amount is zero
	unbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, 6)
	s.Require().True(found)
	unbondingRecord.NativeAmount = sdk.NewInt(0)
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)

	// try confirming with zero token amount on record
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrInvalidUnbondingRecord, "should error when record has zero sweep amount")
}

func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep_NegativeSweepAmount() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	// update the sweep record so that the native amount is negative
	unbondingRecord, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, 6)
	s.Require().True(found)
	unbondingRecord.StTokenAmount = sdk.NewInt(-10)
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)

	// try confirming with negative token amount on record
	err := s.App.StaketiaKeeper.ConfirmUnbondedTokenSweep(s.Ctx, 6, ValidTxHashNew, ValidOperator)
	s.Require().ErrorIs(err, types.ErrInvalidUnbondingRecord, "should error when record has zero sweep amount")
}

// ----------------------------------------------
//              DistributeClaims
// ----------------------------------------------

type DistributeClaimsTestCase struct {
	claimAddress              sdk.AccAddress
	claimableRecordIds        []uint64
	expectedFinalClaimBalance sdkmath.Int
}

func (s *KeeperTestSuite) TestMarkFinishedUnbondings() {
	currentTime := uint64(s.Ctx.BlockTime().Unix())

	// Set unbonding records across different statuses and times
	finishedUnbondingIds := map[uint64]bool{7: true, 8: true, 9: true}
	initialUnbondingRecords := []types.UnbondingRecord{
		{Id: 1, Status: types.ACCUMULATING_REDEMPTIONS, UnbondingCompletionTimeSeconds: currentTime - 1},
		{Id: 2, Status: types.ACCUMULATING_REDEMPTIONS, UnbondingCompletionTimeSeconds: currentTime},
		{Id: 3, Status: types.ACCUMULATING_REDEMPTIONS, UnbondingCompletionTimeSeconds: currentTime + 1},

		{Id: 4, Status: types.UNBONDING_QUEUE, UnbondingCompletionTimeSeconds: currentTime - 1},
		{Id: 5, Status: types.UNBONDING_QUEUE, UnbondingCompletionTimeSeconds: currentTime},
		{Id: 6, Status: types.UNBONDING_QUEUE, UnbondingCompletionTimeSeconds: currentTime + 1},

		{Id: 7, Status: types.UNBONDING_IN_PROGRESS, UnbondingCompletionTimeSeconds: currentTime - 3}, // finished
		{Id: 8, Status: types.UNBONDING_IN_PROGRESS, UnbondingCompletionTimeSeconds: currentTime - 2}, // finished
		{Id: 9, Status: types.UNBONDING_IN_PROGRESS, UnbondingCompletionTimeSeconds: currentTime - 1}, // finished

		{Id: 10, Status: types.UNBONDING_IN_PROGRESS, UnbondingCompletionTimeSeconds: currentTime},     // still unbonding
		{Id: 11, Status: types.UNBONDING_IN_PROGRESS, UnbondingCompletionTimeSeconds: currentTime + 1}, // still unbonding

		{Id: 12, Status: types.UNBONDED, UnbondingCompletionTimeSeconds: currentTime + 1},
	}
	for _, unbondingRecord := range initialUnbondingRecords {
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}

	// Call check unbonding finished
	s.App.StaketiaKeeper.MarkFinishedUnbondings(s.Ctx)

	// Check that the relevant records were updated
	for i, actualUnbondingRecord := range s.App.StaketiaKeeper.GetAllActiveUnbondingRecords(s.Ctx) {
		if _, ok := finishedUnbondingIds[actualUnbondingRecord.Id]; ok {
			s.Require().Equal(actualUnbondingRecord.Status, types.UNBONDED,
				"record %d should have been marked as unbonded", actualUnbondingRecord.Id)
		} else {
			initialRecord := initialUnbondingRecords[i]
			s.Require().Equal(initialRecord.Status, actualUnbondingRecord.Status,
				"record %d status should not have changed", actualUnbondingRecord.Id)
		}
	}
}

// Helper function to mock the state required to test distribute claims
func (s *KeeperTestSuite) SetupTestDistributeClaims() DistributeClaimsTestCase {
	// Fund the claim account
	claimAddress := s.TestAccs[0]
	initialClaimBalance := sdkmath.NewInt(400)
	s.FundAccount(claimAddress, sdk.NewCoin(HostIBCDenom, initialClaimBalance))

	// Create the host zone with a claim address and token denom
	hostZone := types.HostZone{
		ClaimAddress:        claimAddress.String(),
		NativeTokenIbcDenom: HostIBCDenom,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	// Create unhalted stakeibc host zone
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId: types.CelestiaChainId,
	})

	// Define unbonding records with different statuses
	claimableRecordIds := []uint64{1, 3}
	unbondingRecords := []types.UnbondingRecord{
		{Id: 1, Status: types.CLAIMABLE},
		{Id: 2, Status: types.UNBONDING_IN_PROGRESS},
		{Id: 3, Status: types.CLAIMABLE},
		{Id: 4, Status: types.UNBONDING_QUEUE},
	}
	for _, unbondingRecord := range unbondingRecords {
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}

	// Define redmeption records across different unbonding records
	redemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: 1, NativeAmount: sdkmath.NewInt(10)}, // claimable
		{UnbondingRecordId: 1, NativeAmount: sdkmath.NewInt(20)}, // claimable
		{UnbondingRecordId: 2, NativeAmount: sdkmath.NewInt(30)},
		{UnbondingRecordId: 2, NativeAmount: sdkmath.NewInt(40)},
		{UnbondingRecordId: 3, NativeAmount: sdkmath.NewInt(50)}, // claimable
		{UnbondingRecordId: 3, NativeAmount: sdkmath.NewInt(60)}, // claimable
		{UnbondingRecordId: 3, NativeAmount: sdkmath.NewInt(70)}, // claimable
		{UnbondingRecordId: 4, NativeAmount: sdkmath.NewInt(80)},
		{UnbondingRecordId: 4, NativeAmount: sdkmath.NewInt(90)},
	}
	accounts := apptesting.CreateRandomAccounts(len(redemptionRecords))
	expectedFinalClaimBalance := initialClaimBalance.SubRaw(10 + 20 + 50 + 60 + 70)

	// Create a record for each redemption
	for i, redemptionRecord := range redemptionRecords {
		redemptionRecord.Redeemer = accounts[i].String()
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	return DistributeClaimsTestCase{
		claimAddress:              claimAddress,
		claimableRecordIds:        claimableRecordIds,
		expectedFinalClaimBalance: expectedFinalClaimBalance,
	}
}

// The granularity at the redemption record level is covered by TestDistributeClaimsForUnbondingRecord
func (s *KeeperTestSuite) TestDistributeClaims_Success() {
	tc := s.SetupTestDistributeClaims()

	// Call distribute again, it should succeed
	err := s.App.StaketiaKeeper.DistributeClaims(s.Ctx)
	s.Require().NoError(err, "no error expected during claim")

	// Confirm the claim balance was depleted
	actualClaimBalance := s.App.BankKeeper.GetBalance(s.Ctx, tc.claimAddress, HostIBCDenom)
	s.Require().Equal(tc.expectedFinalClaimBalance.Int64(), actualClaimBalance.Amount.Int64(),
		"claim balance should have been depleted")

	// Confirm the CLAIMABLE records were archived
	activeRecords := s.App.StaketiaKeeper.GetAllActiveUnbondingRecords(s.Ctx)
	archivedRecords := s.App.StaketiaKeeper.GetAllArchivedUnbondingRecords(s.Ctx)
	s.Require().Len(activeRecords, 2, "there should only be two remaining active records")
	s.Require().Len(archivedRecords, 2, "there should be two archived records")

	archivedIds := []uint64{archivedRecords[0].Id, archivedRecords[1].Id}
	s.Require().ElementsMatch(tc.claimableRecordIds, archivedIds, "claimable records should now be archived")
}

func (s *KeeperTestSuite) TestDistributeClaims_HostHalted() {
	s.SetupTestDistributeClaims()

	// Halt the host zone, then attempt to call distribute claims, it should fail
	hostZone := s.MustGetHostZone()
	hostZone.Halted = true
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	err := s.App.StaketiaKeeper.DistributeClaims(s.Ctx)
	s.Require().ErrorContains(err, "host zone is halted")
}

func (s *KeeperTestSuite) TestDistributeClaims_InsufficientFunds() {
	s.SetupTestDistributeClaims()

	// Pass through the records again and make them all claimable
	for _, unbondingRecord := range s.App.StaketiaKeeper.GetAllActiveUnbondingRecords(s.Ctx) {
		unbondingRecord.Status = types.CLAIMABLE
		s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, unbondingRecord)
	}

	// Attempt to distribute, it will error cause there will not be enough funds to cover all records
	err := s.App.StaketiaKeeper.DistributeClaims(s.Ctx)
	s.Require().ErrorContains(err, "insufficient funds")
}

func (s *KeeperTestSuite) TestDistributeClaims_InvalidClaimAddress() {
	s.SetupTestDistributeClaims()

	// Update the claim address so that it is invalid
	invalidHostZone := s.MustGetHostZone()
	invalidHostZone.ClaimAddress = "invalid_address"
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, invalidHostZone)

	err := s.App.StaketiaKeeper.DistributeClaims(s.Ctx)
	s.Require().ErrorContains(err, "invalid host zone claim address invalid_address")
}

func (s *KeeperTestSuite) TestDistributeClaimsForUnbondingRecord() {
	// Fund the claim account
	claimAddress := s.TestAccs[0]
	redeemerAddress := s.TestAccs[1]
	initialClaimBalance := sdkmath.NewInt(100)
	s.FundAccount(claimAddress, sdk.NewCoin(HostIBCDenom, initialClaimBalance))

	// Define all the redemptions
	// Unbonding record 1 will be distributed
	distributedUnbondingId := uint64(1)
	redemptionRecords := []types.RedemptionRecord{
		{UnbondingRecordId: distributedUnbondingId, NativeAmount: sdkmath.NewInt(10)}, // 100 (initial) - 10 = 90 remaining
		{UnbondingRecordId: distributedUnbondingId, NativeAmount: sdkmath.NewInt(15)}, // 90 - 15 = 75 remaining
		{UnbondingRecordId: 2, NativeAmount: sdkmath.NewInt(10)},                      // Different unbonding record ID, skipped
		{UnbondingRecordId: distributedUnbondingId, NativeAmount: sdkmath.NewInt(30)}, // 75 - 30 = 45 remaining
		{UnbondingRecordId: 3, NativeAmount: sdkmath.NewInt(10)},                      // Different unbonding record ID, skipped
		{UnbondingRecordId: distributedUnbondingId, NativeAmount: sdkmath.NewInt(8)},  // 45 - 8 = 37 remaining
		{UnbondingRecordId: 4, NativeAmount: sdkmath.NewInt(10)},                      // Different unbonding record ID, skipped
		{UnbondingRecordId: distributedUnbondingId, NativeAmount: sdkmath.NewInt(27)}, // 37 - 27 = 10 remaining (final)
	}
	accounts := apptesting.CreateRandomAccounts(len(redemptionRecords) + 1)
	expectedFinalClaimBalance := initialClaimBalance.SubRaw(10 + 15 + 30 + 8 + 27)

	// Create a record for each redemption
	for i, redemptionRecord := range redemptionRecords {
		redemptionRecord.Redeemer = accounts[i].String()
		s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, redemptionRecord)
	}

	// Call distribute on the unbonding record in question
	err := s.App.StaketiaKeeper.DistributeClaimsForUnbondingRecord(
		s.Ctx,
		HostIBCDenom,
		claimAddress,
		distributedUnbondingId,
	)
	s.Require().NoError(err, "no error expected when distributing claims")

	// Confirm the claim balance was depleted
	actualClaimBalance := s.App.BankKeeper.GetBalance(s.Ctx, claimAddress, HostIBCDenom)
	s.Require().Equal(expectedFinalClaimBalance.Int64(), actualClaimBalance.Amount.Int64(),
		"claim balance should have been depleted")

	// Loop again, confirm all users received their tokens, and that all redemption records were removed
	for i, redemption := range redemptionRecords {
		if redemption.UnbondingRecordId != distributedUnbondingId {
			continue
		}
		redeemer := accounts[i]
		userBalance := s.App.BankKeeper.GetBalance(s.Ctx, redeemer, HostIBCDenom)
		s.Require().Equal(redemption.NativeAmount.Int64(), userBalance.Amount.Int64(),
			"user %d should have received their native tokens", i)

		_, found := s.App.StaketiaKeeper.GetRedemptionRecord(s.Ctx, redemption.UnbondingRecordId, redemption.Redeemer)
		s.Require().False(found, "redemption record for unbonding record %d should have been removed",
			redemption.UnbondingRecordId)
	}

	// Add a record with an amount that would exceed the claim account's remaining balance
	exceedBalanceUnbondingId := uint64(100)
	s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, types.RedemptionRecord{
		UnbondingRecordId: exceedBalanceUnbondingId,
		Redeemer:          redeemerAddress.String(),
		NativeAmount:      initialClaimBalance,
	})

	// Add a record with an invalid address
	invalidAddressUnbondingId := uint64(200)
	s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, types.RedemptionRecord{
		UnbondingRecordId: invalidAddressUnbondingId,
		Redeemer:          "invalid_address",
		NativeAmount:      initialClaimBalance,
	})

	// Attempt to distribute for that record, it should fail
	err = s.App.StaketiaKeeper.DistributeClaimsForUnbondingRecord(
		s.Ctx,
		HostIBCDenom,
		claimAddress,
		exceedBalanceUnbondingId,
	)
	s.Require().ErrorContains(err, "insufficient funds")

	// Attempt to distribute for that record, it should fail
	err = s.App.StaketiaKeeper.DistributeClaimsForUnbondingRecord(
		s.Ctx,
		HostIBCDenom,
		claimAddress,
		invalidAddressUnbondingId,
	)
	s.Require().ErrorContains(err, "invalid redeemer address")
}
