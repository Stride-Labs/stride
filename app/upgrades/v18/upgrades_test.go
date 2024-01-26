package v18_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v17/app/apptesting"
	v18 "github.com/Stride-Labs/stride/v17/app/upgrades/v18"
	recordtypes "github.com/Stride-Labs/stride/v17/x/records/types"
	"github.com/Stride-Labs/stride/v17/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v17/x/stakeibc/types"
)

type UserRedemptionRecordTestCases struct {
	Id          string
	EpochNumber uint64
	HostZoneId  string
	StAmount    int64

	// Redemption rate at the time the unbonding was submitted
	UnbondedRR string
	// Redemption rate used in the records
	RecordRR string

	// Native token amount at the time the unbonding was submitted
	InitialNativeAmount sdkmath.Int
	// Expected native amount after the upgrade
	ExpectedNativeAmount sdkmath.Int
}

type HostZoneUnbondingTestCase struct {
	StAmount int64
	URRs     []string

	// Redemption rate at the time the unbonding was submitted
	UnbondedRR string
	// Redemption rate used in the records
	RecordRR string

	// Native token amount at the time the unbonding was submitted
	InitialNativeAmount sdkmath.Int
	// Expected native amount after the upgrade
	ExpectedNativeAmount sdkmath.Int
}

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func (s *UpgradeTestSuite) TestUpgrade() {
	dummyUpgradeHeight := int64(5)

	checkStoreAfterUpgrade := s.SetupTestUnbondingRecords()
	s.ConfirmUpgradeSucceededs("v18", dummyUpgradeHeight)
	checkStoreAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupTestUnbondingRecords() func() {
	chainId := "sommelier-3"
	epochNumberBefore := uint64(501)
	epochNumberAfter := uint64(510)
	stTokenAmount := sdkmath.NewInt(1_000_000)

	// Unbonding #1 (before prop)
	redemptionRateUnbonded := sdk.MustNewDecFromStr("1.0256449837573494") // Rate used when unbonding occurred
	redemptionRateRecords := sdk.MustNewDecFromStr("1.0257449837673494")  // Rate that was on record at the time (+0.0001)

	s.Require().Equal(redemptionRateUnbonded.String(), v18.RedemptionRatesBeforeProp[chainId][epochNumberBefore].String(),
		"example redemption rate from before prop does not match constants - update the test")

	// Unbonding #2 (after prop)
	redemptionRateAtPropTime := sdk.MustNewDecFromStr("1.025900883208774724") // Rate at prop time
	redemptionRateAtUpgradeTime := sdk.MustNewDecFromStr("1.03")              // Rate at upgrade time
	estimatedRedemptionRate := sdk.MustNewDecFromStr("1.027950441604387362")  // Estimated rate used to update record
	unknownRedemptionRate := sdk.MustNewDecFromStr("1.029")                   // Rate that was used in unbonding

	s.Require().Equal(redemptionRateAtPropTime.String(), v18.RedemptionRatesAtTimeOfProp[chainId].String(),
		"example redemption rate from time of prop does not match constants - update the test")

	// Calculate native token in the records before the upgrade
	initialNativeAmount1 := sdk.NewDecFromInt(stTokenAmount).Mul(redemptionRateRecords).TruncateInt()
	initialNativeAmount2 := sdk.NewDecFromInt(stTokenAmount).Mul(unknownRedemptionRate).TruncateInt()

	// Calculate expected native amounts after upgrade
	expectedNativeAmount1 := sdk.NewDecFromInt(stTokenAmount).Mul(redemptionRateUnbonded).TruncateInt()
	expectedNativeAmount2 := sdk.NewDecFromInt(stTokenAmount).Mul(estimatedRedemptionRate).TruncateInt()

	// Create the host zone with redemption rate at time of upgrade
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:        chainId,
		RedemptionRate: redemptionRateAtUpgradeTime,
	})

	// Create redemption records - one before and one after the prop
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:                "A",
		EpochNumber:       epochNumberBefore,
		HostZoneId:        chainId,
		StTokenAmount:     stTokenAmount,
		NativeTokenAmount: initialNativeAmount1,
	})
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:                "B",
		EpochNumber:       epochNumberAfter,
		HostZoneId:        chainId,
		StTokenAmount:     stTokenAmount,
		NativeTokenAmount: initialNativeAmount2,
	})

	// Create epoch unbonding records
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberBefore,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:            chainId,
				Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				StTokenAmount:         stTokenAmount,
				NativeTokenAmount:     initialNativeAmount1,
				UserRedemptionRecords: []string{"A"},
			},
		},
	})
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberAfter,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:            chainId,
				Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				StTokenAmount:         stTokenAmount,
				NativeTokenAmount:     initialNativeAmount2,
				UserRedemptionRecords: []string{"B"},
			},
		},
	})

	// Add a record that should be ignored because the epoch number is too low
	epochNumberIgnore := uint64(3)
	nativeAmountIgnored := stTokenAmount
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:                "C",
		EpochNumber:       epochNumberIgnore,
		HostZoneId:        chainId,
		StTokenAmount:     stTokenAmount,
		NativeTokenAmount: nativeAmountIgnored,
	})
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberIgnore,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:            chainId,
				Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				StTokenAmount:         stTokenAmount,
				NativeTokenAmount:     nativeAmountIgnored,
				UserRedemptionRecords: []string{"C"},
			},
		},
	})

	// Return callback to check store after upgrade
	return func() {
		// Check the user redemption record amount for unbonding 1
		actualRedemptionRecord1, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, "A")
		s.Require().True(found, "record from first unbonding should have been found")
		s.Require().Equal(expectedNativeAmount1.Int64(), actualRedemptionRecord1.NativeTokenAmount.Int64(),
			"native amount on record from first unbonding")

		// Check the user redemption record amount for unbonding 2
		actualRedemptionRecord2, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, "B")
		s.Require().True(found, "record from second unbonding should have been found")
		s.Require().Equal(expectedNativeAmount2.Int64(), actualRedemptionRecord2.NativeTokenAmount.Int64(),
			"native amount on record from second unbonding")

		// Check the host zone unbonding amount for unbonding 1
		actualHostZoneUnbonding1, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochNumberBefore, chainId)
		s.Require().True(found, "host zone unbonding should have been found for second unbonding")
		s.Require().Equal(expectedNativeAmount1.Int64(), actualHostZoneUnbonding1.NativeTokenAmount.Int64(),
			"host zone native amount from first unbonding")

		// Check the host zone unbonding amount for unbonding 2
		actualHostZoneUnbonding2, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochNumberAfter, chainId)
		s.Require().True(found, "host zone unbonding should have been found for second unbonding")
		s.Require().Equal(expectedNativeAmount2.Int64(), actualHostZoneUnbonding2.NativeTokenAmount.Int64(),
			"host zone native amount from first unbonding")

		// Check that the ignored record did not change
		ignoredRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, "C")
		s.Require().True(found, "ignored record should have been found")
		s.Require().Equal(nativeAmountIgnored.Int64(), ignoredRecord.NativeTokenAmount.Int64(),
			"native amount on ignored record should not have changed")
	}
}

func (s *UpgradeTestSuite) TestUpdateUnbondingRecords() {
	// We'll create the following scenario across two host zones
	//    T0 - HostZone1 Unbonds from EpochRecord1 - Record RR is 1.10 - Unbonded RR was 1.05
	//    T1 - HostZone2 Unbonds from EpochRecord1 - Record RR is 1.50 - Unbonded RR was 1.40
	//    ...
	//    T2 - HostZone1 Unbonds from EpochRecord2 - Record RR is 1.15 - Unbonded RR was 1.10
	//    T3 - HostZone2 Unbonds from EpochRecord2 - Record RR is 1.65 - Unbonded RR was 1.50
	//    ...
	//    upgrade prop submitted
	//    HostZone1 RR is 1.20
	//    HostZone2 RR is 1.80
	//    ...
	//    T4 - HostZone1 Unbonds from EpochRecord3 - Record RR is 1.32     <- Use implied RR of (1.2 + 1.4) / 2 = 1.30
	//    T5 - HostZone2 Unbonds from EpochRecord3 - Record RR is 1.88     <- Use implied RR of (1.8 + 1.9) / 2 = 1.85
	//    ...
	//    upgrade goes live
	//    HostZone1 RR is 1.40
	//    HostZone2 RR is 1.90

	// Create host zones with the RR at the time that the upgrade goes live
	chainId1 := "chain-1"
	chainId2 := "chain-2"

	hostZone1 := stakeibctypes.HostZone{
		ChainId:        chainId1,
		RedemptionRate: sdk.MustNewDecFromStr("1.4"),
	}
	hostZone2 := stakeibctypes.HostZone{
		ChainId:        chainId2,
		RedemptionRate: sdk.MustNewDecFromStr("1.9"),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone2)

	// Save down the redemption rates before the prop and at the time of the prop
	// These are both hard coded in the upgrade handler
	// For the before rates, use the rate that was actually unbonded (not the one in the record)
	redemptionRatesAtTimeOfProp := map[string]sdk.Dec{
		chainId1: sdk.MustNewDecFromStr("1.2"),
		chainId2: sdk.MustNewDecFromStr("1.8"),
	}
	redemptionRatesBeforeProp := map[string]map[uint64]sdk.Dec{
		chainId1: {
			1: sdk.MustNewDecFromStr("1.05"),
			2: sdk.MustNewDecFromStr("1.10"),
		},
		chainId2: {
			1: sdk.MustNewDecFromStr("1.40"),
			2: sdk.MustNewDecFromStr("1.50"),
		},
	}

	// Define all user redemption records across each of the host zone unbondings
	// We store just the redemption rate now, but will set native amount when setting them
	userRedemptionRecordTestCases := []UserRedemptionRecordTestCases{
		// Epoch 1 - HostZone1 - RR 1.10
		{Id: "A", EpochNumber: 1, HostZoneId: chainId1, StAmount: 1000, RecordRR: "1.1", UnbondedRR: "1.05"},
		{Id: "B", EpochNumber: 1, HostZoneId: chainId1, StAmount: 2000, RecordRR: "1.1", UnbondedRR: "1.05"},
		// Epoch 1 - HostZone2 - RR 1.50
		{Id: "C", EpochNumber: 1, HostZoneId: chainId2, StAmount: 3000, RecordRR: "1.5", UnbondedRR: "1.4"},
		{Id: "D", EpochNumber: 1, HostZoneId: chainId2, StAmount: 4000, RecordRR: "1.5", UnbondedRR: "1.4"},

		// Epoch 2 - HostZone1 - RR 1.15
		{Id: "E", EpochNumber: 2, HostZoneId: chainId1, StAmount: 5000, RecordRR: "1.15", UnbondedRR: "1.1"},
		{Id: "F", EpochNumber: 2, HostZoneId: chainId1, StAmount: 6000, RecordRR: "1.15", UnbondedRR: "1.1"},
		{Id: "G", EpochNumber: 2, HostZoneId: chainId1, StAmount: 7000, RecordRR: "1.15", UnbondedRR: "1.1"},
		// Epoch 2 - HostZone2 - RR 1.65
		{Id: "H", EpochNumber: 2, HostZoneId: chainId2, StAmount: 8000, RecordRR: "1.65", UnbondedRR: "1.5"},
		{Id: "I", EpochNumber: 2, HostZoneId: chainId2, StAmount: 9000, RecordRR: "1.65", UnbondedRR: "1.5"},

		// Epoch 2 - HostZone1 - RR 1.40
		{Id: "J", EpochNumber: 3, HostZoneId: chainId1, StAmount: 10000, RecordRR: "1.40", UnbondedRR: "1.30"},
		{Id: "K", EpochNumber: 3, HostZoneId: chainId1, StAmount: 11000, RecordRR: "1.40", UnbondedRR: "1.30"},
		// Epoch 2 - HostZone2 - RR 1.90
		{Id: "L", EpochNumber: 3, HostZoneId: chainId2, StAmount: 12000, RecordRR: "1.90", UnbondedRR: "1.85"},
		{Id: "M", EpochNumber: 3, HostZoneId: chainId2, StAmount: 13000, RecordRR: "1.90", UnbondedRR: "1.85"},
		{Id: "N", EpochNumber: 3, HostZoneId: chainId2, StAmount: 14000, RecordRR: "1.90", UnbondedRR: "1.85"},
	}

	// Consolidate the totals from above into the host zone unbonding level by
	// grabbing the unbonded redemption rate, summing up the stToken amounts,
	// and aggregating a list of the redemption Ids
	// We'll store it all in a mapping of epochNumber -> chainId -> aggregate values
	hostZoneUnbondingsInfoMap := map[uint64]map[string]HostZoneUnbondingTestCase{
		1: {
			chainId1: {RecordRR: "1.1", UnbondedRR: "1.05", URRs: []string{"A", "B"}, StAmount: 1000 + 2000},
			chainId2: {RecordRR: "1.5", UnbondedRR: "1.4", URRs: []string{"C", "D"}, StAmount: 3000 + 4000},
		},
		2: {
			chainId1: {RecordRR: "1.15", UnbondedRR: "1.1", URRs: []string{"E", "F", "G"}, StAmount: 5000 + 6000 + 7000},
			chainId2: {RecordRR: "1.65", UnbondedRR: "1.5", URRs: []string{"H", "I"}, StAmount: 8000 + 9000},
		},
		3: {
			chainId1: {RecordRR: "1.40", UnbondedRR: "1.30", URRs: []string{"J", "K"}, StAmount: 10000 + 11000},
			chainId2: {RecordRR: "1.90", UnbondedRR: "1.85", URRs: []string{"L", "M", "N"}, StAmount: 12000 + 13000 + 14000},
		},
	}

	// Write all the redemption records to the store, calculating the native amount from
	// the "record" redemption rate
	for i, userTestCase := range userRedemptionRecordTestCases {
		epochNumber := userTestCase.EpochNumber
		hostZoneId := userTestCase.HostZoneId
		stTokenAmount := sdkmath.NewInt(userTestCase.StAmount)

		// Calculate the native amount from the "record RR" - this is the value that will be in the
		// store before the upgrade
		recordRR := sdk.MustNewDecFromStr(userTestCase.RecordRR)
		recordNativeAmount := sdk.NewDecFromInt(stTokenAmount).Mul(recordRR).TruncateInt()

		// Calculate the native amount from the "unbond RR" - this is the implied RR from the
		// actual unbonding
		unbondRR := sdk.MustNewDecFromStr(userTestCase.UnbondedRR)
		actualUnbondAmount := sdk.NewDecFromInt(stTokenAmount).Mul(unbondRR).TruncateInt()

		// Create the user redmeption record
		s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
			Id:                userTestCase.Id,
			EpochNumber:       epochNumber,
			HostZoneId:        hostZoneId,
			StTokenAmount:     stTokenAmount,
			NativeTokenAmount: recordNativeAmount,
		})

		// Update the native amounts in the test case
		userRedemptionRecordTestCases[i].InitialNativeAmount = recordNativeAmount
		userRedemptionRecordTestCases[i].ExpectedNativeAmount = actualUnbondAmount
	}

	// Then do a similar loop for the host zone unbonding records
	for epochNumber, chainToHZUMap := range hostZoneUnbondingsInfoMap {
		for chainId, hostTestCase := range chainToHZUMap {
			stTokenAmount := sdkmath.NewInt(hostTestCase.StAmount)

			// Calculate the native amount from the "record RR" - this is the value that will be in the
			// store before the upgrade
			recordRR := sdk.MustNewDecFromStr(hostTestCase.RecordRR)
			recordNativeAmount := sdk.NewDecFromInt(stTokenAmount).Mul(recordRR).TruncateInt()

			// Calculate the native amount from the "unbond RR" - this is the implied RR from the
			// actual unbonding
			unbondRR := sdk.MustNewDecFromStr(hostTestCase.UnbondedRR)
			actualUnbondAmount := sdk.NewDecFromInt(stTokenAmount).Mul(unbondRR).TruncateInt()

			// Initialize the epoch unbonding record if it hasn't happened already
			if _, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, epochNumber); !found {
				s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
					EpochNumber: epochNumber,
				})
			}

			// Set host zone unbonding record
			hostZoneUnbondingRecord := recordtypes.HostZoneUnbonding{
				Status:                recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
				HostZoneId:            chainId,
				UserRedemptionRecords: hostTestCase.URRs,
				StTokenAmount:         stTokenAmount,
				NativeTokenAmount:     recordNativeAmount,
			}
			err := s.App.RecordsKeeper.SetHostZoneUnbondingRecord(s.Ctx, epochNumber, chainId, hostZoneUnbondingRecord)
			s.Require().NoError(err, "no error expected when setting host zone unbonding")

			// Update the test case so we can check against expectations later
			updatedTestCase := hostZoneUnbondingsInfoMap[epochNumber][chainId]
			updatedTestCase.InitialNativeAmount = recordNativeAmount
			updatedTestCase.ExpectedNativeAmount = actualUnbondAmount
			hostZoneUnbondingsInfoMap[epochNumber][chainId] = updatedTestCase
		}
	}

	// Run the migration
	startingEpochNumber := uint64(1)
	err := v18.UpdateUnbondingRecords(
		s.Ctx,
		s.App.StakeibcKeeper,
		s.App.RecordsKeeper,
		startingEpochNumber,
		redemptionRatesBeforeProp,
		redemptionRatesAtTimeOfProp,
	)
	s.Require().NoError(err, "no error expected when updating records")

	// Confirm user redemption records were updated
	for _, userTestCase := range userRedemptionRecordTestCases {
		actualRedemptionRecord, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, userTestCase.Id)
		s.Require().True(found, "user redemption record %s should have been found after upgrade", userTestCase.Id)
		s.Require().Equal(userTestCase.ExpectedNativeAmount.Int64(), actualRedemptionRecord.NativeTokenAmount.Int64(),
			"user redemption record native amount for %s", userTestCase.Id)
	}

	// Confirm host zone unbonding records were updated
	for epochNumber, chainToHZUMap := range hostZoneUnbondingsInfoMap {
		for chainId, hostTestCase := range chainToHZUMap {
			actualHostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochNumber, chainId)
			s.Require().True(found, "HZU for epoch %d and host zone %s should have been found", epochNumber, chainId)
			s.Require().Equal(hostTestCase.ExpectedNativeAmount.Int64(), actualHostZoneUnbonding.NativeTokenAmount.Int64(),
				"host zone unbonding native amount for epoch %d and host zone %s", epochNumber, chainId)
		}
	}
}

func (s *UpgradeTestSuite) TestDecrementTerraDelegationChangesInProgress() {
	// Create list of validators
	validators := []*types.Validator{}
	for i := 0; i < 5; i++ {
		address := fmt.Sprintf("val-%d", i)
		validators = append(validators, &types.Validator{Address: address, DelegationChangesInProgress: int64(i)})
	}

	// set the host zone
	hostZone1 := stakeibctypes.HostZone{
		ChainId:    v18.TerraChainId,
		Validators: validators,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone1)

	err := v18.DecrementTerraDelegationChangesInProgress(s.Ctx, s.App.StakeibcKeeper)
	s.Require().NoError(err, "no error decrementing terra delegation changes in progress")

	hostZoneAfter, err := s.App.StakeibcKeeper.GetActiveHostZone(s.Ctx, v18.TerraChainId)
	s.Require().NoError(err, "get host zone")

	// check each val
	expectedVals := []int64{0, 0, 0, 0, 1, 2}
	for i := 0; i < 5; i++ {
		s.Require().Equal(expectedVals[i], hostZoneAfter.Validators[i].DelegationChangesInProgress)
	}
}

func (s *UpgradeTestSuite) TestDecrementTerraDelegationChangesInProgress_ZoneNotFound() {
	// test host zone not found
	hostZoneWrongChainId := stakeibctypes.HostZone{
		ChainId: "not-terra",
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZoneWrongChainId)

	err := v18.DecrementTerraDelegationChangesInProgress(s.Ctx, s.App.StakeibcKeeper)
	s.Require().Error(err, "host zone not found")
}
