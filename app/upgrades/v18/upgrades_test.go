package v18_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	"github.com/Stride-Labs/stride/v26/app/apptesting"
	v18 "github.com/Stride-Labs/stride/v26/app/upgrades/v18"
	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinOuterRedemptionRate sdkmath.LegacyDec
	ExpectedMaxOuterRedemptionRate sdkmath.LegacyDec
}

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

	// Setup store before upgrade
	checkDelegationsAfterUpgrade := s.SetupTestResetDelegationChanges()
	checkUnbondingsAfterUpgrade := s.SetupTestUnbondingRecords()
	checkRedemptionRatesAfterUpgrade := s.SetupTestUpdateRedemptionRateBounds()

	// Run through upgrade
	s.ConfirmUpgradeSucceededs("v18", dummyUpgradeHeight)

	// Check store after upgrade
	checkDelegationsAfterUpgrade()
	checkRedemptionRatesAfterUpgrade()
	checkUnbondingsAfterUpgrade()
}

func (s *UpgradeTestSuite) SetupTestResetDelegationChanges() func() {
	validators := []*stakeibctypes.Validator{
		{DelegationChangesInProgress: 3},
		{DelegationChangesInProgress: 3},
		{DelegationChangesInProgress: 3},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:    v18.TerraChainId,
		Validators: validators,
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:    "different-host",
		Validators: validators,
	})

	// Return callback to check store after upgrade
	return func() {
		for _, hostZone := range s.App.StakeibcKeeper.GetAllHostZone(s.Ctx) {
			expected := int64(3)
			if hostZone.ChainId == v18.TerraChainId {
				expected = 0
			}
			for _, validator := range hostZone.Validators {
				s.Require().Equal(expected, validator.DelegationChangesInProgress)
			}
		}
	}
}

func (s *UpgradeTestSuite) SetupTestUpdateRedemptionRateBounds() func() {
	// Define test cases consisting of an initial redemption rate and expected bounds
	testCases := []UpdateRedemptionRateBounds{
		{
			ChainId:                        "chain-0",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.0"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("0.95"), // 1 - 5% = 0.95
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.10"), // 1 + 10% = 1.1
		},
		{
			ChainId:                        "chain-1",
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.1"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.045"), // 1.1 - 5% = 1.045
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.210"), // 1.1 + 10% = 1.21
		},
		{
			// Max outer for osmo uses 12% instead of 10%
			ChainId:                        v18.OsmosisChainId,
			CurrentRedemptionRate:          sdkmath.LegacyMustNewDecFromStr("1.25"),
			ExpectedMinOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.1875"), // 1.25 - 5% = 1.1875
			ExpectedMaxOuterRedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.4000"), // 1.25 + 12% = 1.400
		},
	}

	// Create a host zone for each test case
	for _, tc := range testCases {
		hostZone := stakeibctypes.HostZone{
			ChainId:        tc.ChainId,
			RedemptionRate: tc.CurrentRedemptionRate,
		}
		s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	}

	// Return callback to check store after upgrade
	return func() {
		// Confirm they were all updated
		for _, tc := range testCases {
			hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, tc.ChainId)
			s.Require().True(found)

			s.Require().Equal(tc.ExpectedMinOuterRedemptionRate, hostZone.MinRedemptionRate, "%s - min outer", tc.ChainId)
			s.Require().Equal(tc.ExpectedMaxOuterRedemptionRate, hostZone.MaxRedemptionRate, "%s - max outer", tc.ChainId)
		}
	}
}

func (s *UpgradeTestSuite) SetupTestUnbondingRecords() func() {
	chainId := "sommelier-3"
	epochNumberBefore := uint64(501)
	epochNumberAfter := uint64(510)
	stTokenAmount := sdkmath.NewInt(1_000_000)

	// Unbonding #1 (before prop)
	redemptionRateUnbonded := sdkmath.LegacyMustNewDecFromStr("1.025236900070852") // Rate used when unbonding occurred
	redemptionRateRecords := sdkmath.LegacyMustNewDecFromStr("1.025136900070852")  // Rate that was on record at the time (+0.0001)

	s.Require().Equal(redemptionRateUnbonded.String(), v18.RedemptionRatesBeforeProp[chainId][epochNumberBefore].String(),
		"example redemption rate from before prop does not match constants - update the test")

	// Unbonding #2 (after prop)
	redemptionRateAtPropTime := sdkmath.LegacyMustNewDecFromStr("1.025900897761638723") // Rate at prop time
	redemptionRateAtUpgradeTime := sdkmath.LegacyMustNewDecFromStr("1.03")              // Rate at upgrade time
	estimatedRedemptionRate := sdkmath.LegacyMustNewDecFromStr("1.027950448880819361")  // Estimated rate used to update record
	unknownRedemptionRate := sdkmath.LegacyMustNewDecFromStr("1.029")                   // Rate that was used in unbonding

	s.Require().Equal(redemptionRateAtPropTime.String(), v18.RedemptionRatesAtTimeOfProp[chainId].String(),
		"example redemption rate from time of prop does not match constants - update the test")

	// Calculate native token in the records before the upgrade
	initialNativeAmount1 := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(redemptionRateRecords).TruncateInt()
	initialNativeAmount2 := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(unknownRedemptionRate).TruncateInt()

	// Calculate expected native amounts after upgrade
	expectedNativeAmount1 := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(redemptionRateUnbonded).TruncateInt()
	expectedNativeAmount2 := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(estimatedRedemptionRate).TruncateInt()

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
	epochNumberIgnore1 := uint64(3)
	nativeAmountIgnored := stTokenAmount
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:                "C",
		EpochNumber:       epochNumberIgnore1,
		HostZoneId:        chainId,
		StTokenAmount:     stTokenAmount,
		NativeTokenAmount: nativeAmountIgnored,
	})
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberIgnore1,
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

	// Add another record that should be ignored - this one because the status is not EXIT_TRANSFER_IN_QUEUE
	epochNumberIgnore2 := uint64(505) // should be in constants
	_, ok := v18.RedemptionRatesBeforeProp[chainId][epochNumberIgnore2]
	s.Require().True(ok, "example epoch should be in redemption rate map - update the test")

	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx, recordtypes.UserRedemptionRecord{
		Id:                "D",
		EpochNumber:       epochNumberIgnore2,
		HostZoneId:        chainId,
		StTokenAmount:     stTokenAmount,
		NativeTokenAmount: nativeAmountIgnored,
	})
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
		EpochNumber: epochNumberIgnore2,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
			{
				HostZoneId:            chainId,
				Status:                recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
				StTokenAmount:         stTokenAmount,
				NativeTokenAmount:     nativeAmountIgnored,
				UserRedemptionRecords: []string{"D"},
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

		// Check that the ignored record 1 did not change
		ignoredRecord1, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, "C")
		s.Require().True(found, "ignored record should have been found")
		s.Require().Equal(nativeAmountIgnored.Int64(), ignoredRecord1.NativeTokenAmount.Int64(),
			"native amount on ignored record should not have changed")

		// Check that the ignored record 2 did not change
		ignoredRecord2, found := s.App.RecordsKeeper.GetUserRedemptionRecord(s.Ctx, "D")
		s.Require().True(found, "ignored record should have been found")
		s.Require().Equal(nativeAmountIgnored.Int64(), ignoredRecord2.NativeTokenAmount.Int64(),
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
		RedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.4"),
	}
	hostZone2 := stakeibctypes.HostZone{
		ChainId:        chainId2,
		RedemptionRate: sdkmath.LegacyMustNewDecFromStr("1.9"),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone1)
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone2)

	// Save down the redemption rates before the prop and at the time of the prop
	// These are both hard coded in the upgrade handler
	// For the before rates, use the rate that was actually unbonded (not the one in the record)
	redemptionRatesAtTimeOfProp := map[string]sdkmath.LegacyDec{
		chainId1: sdkmath.LegacyMustNewDecFromStr("1.2"),
		chainId2: sdkmath.LegacyMustNewDecFromStr("1.8"),
	}
	redemptionRatesBeforeProp := map[string]map[uint64]sdkmath.LegacyDec{
		chainId1: {
			1: sdkmath.LegacyMustNewDecFromStr("1.05"),
			2: sdkmath.LegacyMustNewDecFromStr("1.10"),
		},
		chainId2: {
			1: sdkmath.LegacyMustNewDecFromStr("1.40"),
			2: sdkmath.LegacyMustNewDecFromStr("1.50"),
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
		recordRR := sdkmath.LegacyMustNewDecFromStr(userTestCase.RecordRR)
		recordNativeAmount := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(recordRR).TruncateInt()

		// Calculate the native amount from the "unbond RR" - this is the implied RR from the
		// actual unbonding
		unbondRR := sdkmath.LegacyMustNewDecFromStr(userTestCase.UnbondedRR)
		actualUnbondAmount := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(unbondRR).TruncateInt()

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
			recordRR := sdkmath.LegacyMustNewDecFromStr(hostTestCase.RecordRR)
			recordNativeAmount := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(recordRR).TruncateInt()

			// Calculate the native amount from the "unbond RR" - this is the implied RR from the
			// actual unbonding
			unbondRR := sdkmath.LegacyMustNewDecFromStr(hostTestCase.UnbondedRR)
			actualUnbondAmount := sdkmath.LegacyNewDecFromInt(stTokenAmount).Mul(unbondRR).TruncateInt()

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

func (s *UpgradeTestSuite) TestExecuteProp228IfPassed() {
	sender := sdk.MustAccAddressFromBech32(v18.IncentiveProgramAddress)
	receiver := sdk.MustAccAddressFromBech32(v18.StrideFoundationAddress_F4)

	// Fund the sender
	s.FundAccount(sender, sdk.NewCoin(v18.Strd, v18.Prop228SendAmount))

	// Attempt to run when the prop has not been created yet - it should error
	err := v18.ExecuteProp228IfPassed(s.Ctx, s.App.BankKeeper, s.App.GovKeeper)
	s.Require().ErrorContains(err, "Prop 228 not found")

	// Store the prop in status rejected
	err = s.App.GovKeeper.SetProposal(s.Ctx, govtypes.Proposal{
		Id:     v18.Prop228ProposalId,
		Status: govtypes.ProposalStatus_PROPOSAL_STATUS_REJECTED,
	})
	s.Require().NoError(err)

	// Attempt to run when it's been rejected, it should not error but no funds
	// should be sent
	err = v18.ExecuteProp228IfPassed(s.Ctx, s.App.BankKeeper, s.App.GovKeeper)
	s.Require().NoError(err, "no error expected after rejected prop")

	senderBalance := s.App.BankKeeper.GetBalance(s.Ctx, sender, v18.Strd).Amount
	receiverBalance := s.App.BankKeeper.GetBalance(s.Ctx, receiver, v18.Strd).Amount

	s.Require().Zero(receiverBalance.Int64(), "receiver balance should not have changed")
	s.Require().Equal(v18.Prop228SendAmount.Int64(), senderBalance.Int64(),
		"sender balance should not have changed")

	// Update the prop to be successful
	err = s.App.GovKeeper.SetProposal(s.Ctx, govtypes.Proposal{
		Id:     v18.Prop228ProposalId,
		Status: govtypes.ProposalStatus_PROPOSAL_STATUS_PASSED,
	})
	s.Require().NoError(err)

	// Execute the prop again and confirm balances were updated
	err = v18.ExecuteProp228IfPassed(s.Ctx, s.App.BankKeeper, s.App.GovKeeper)
	s.Require().NoError(err, "no error expected after passed prop")

	senderBalance = s.App.BankKeeper.GetBalance(s.Ctx, sender, v18.Strd).Amount
	receiverBalance = s.App.BankKeeper.GetBalance(s.Ctx, receiver, v18.Strd).Amount

	s.Require().Zero(senderBalance.Int64(), "sender balance should be zero")
	s.Require().Equal(v18.Prop228SendAmount.Int64(), receiverBalance.Int64(),
		"receiver address should have recieved prop funds")
}
