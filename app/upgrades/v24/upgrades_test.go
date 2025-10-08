package v24_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v29/app/apptesting"
	v24 "github.com/Stride-Labs/stride/v29/app/upgrades/v24"
	recordstypes "github.com/Stride-Labs/stride/v29/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v29/x/stakeibc/types"
)

type UpdateRedemptionRateBounds struct {
	ChainId                        string
	CurrentRedemptionRate          sdkmath.LegacyDec
	ExpectedMinOuterRedemptionRate sdkmath.LegacyDec
	ExpectedMaxOuterRedemptionRate sdkmath.LegacyDec
}

type UpgradeTestSuite struct {
	apptesting.AppTestHelper
}

func (s *UpgradeTestSuite) SetupTest() {
	s.Setup()
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}

func (s *UpgradeTestSuite) TestUpgrade() {
	chainId := "chain-0"
	depositRecordId := uint64(1)
	epochNumber := uint64(1)

	// Create a host zone with redemptions enabled set to false
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:            chainId,
		RedemptionsEnabled: false,
	})

	// Create an in-progress deposit record
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, recordstypes.DepositRecord{
		Id:                      depositRecordId,
		Status:                  recordstypes.DepositRecord_DELEGATION_IN_PROGRESS,
		DelegationTxsInProgress: 0,
	})

	// Create an in-progress unbonding record
	stTokens := sdkmath.NewInt(100)
	nativeTokens := sdkmath.NewInt(200)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordstypes.EpochUnbondingRecord{
		EpochNumber: 1,
		HostZoneUnbondings: []*recordstypes.HostZoneUnbonding{
			{
				HostZoneId:        chainId,
				StTokenAmount:     stTokens,
				NativeTokenAmount: nativeTokens,
				Status:            recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
			},
		},
	})

	checkRedemptionRatesAfterUpgrade := s.SetupTestUpdateRedemptionRateBounds()

	// Run the upgrade
	s.ConfirmUpgradeSucceeded(v24.UpgradeName)

	// Confirm the host zone's redemptions enabled were set to true
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, chainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().True(hostZone.RedemptionsEnabled, "redemptions enabled")

	// Confirm the deposit record's delegations in progress were set to 1
	depositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, depositRecordId)
	s.Require().True(found, "deposit record should have been found")
	s.Require().Equal(uint64(1), depositRecord.DelegationTxsInProgress, "delegation txs in progress")

	// Confirm the new unbonding record fields were set
	hostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, epochNumber, chainId)
	s.Require().True(found, "host zone unbonding should have been found")
	s.Require().Equal(uint64(1), hostZoneUnbonding.UndelegationTxsInProgress, "undelegation txs in progress")
	s.Require().Equal(stTokens.Int64(), hostZoneUnbonding.StTokensToBurn.Int64(), "sttokens to burn")
	s.Require().Equal(nativeTokens.Int64(), hostZoneUnbonding.NativeTokensToUnbond.Int64(), "native to unbond")
	s.Require().Zero(hostZoneUnbonding.ClaimableNativeTokens.Int64(), "claimable tokens")

	checkRedemptionRatesAfterUpgrade()
}

func (s *UpgradeTestSuite) TestMigrateHostZones() {
	// Create a host zone with redemptions enabled set to false
	hostZone := stakeibctypes.HostZone{
		ChainId:            "chain-0",
		RedemptionsEnabled: false,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Call migration function
	v24.MigrateHostZones(s.Ctx, s.App.StakeibcKeeper)

	// Confirm host route was migrated
	for _, hostZone := range s.App.StakeibcKeeper.GetAllHostZone(s.Ctx) {
		s.Require().True(hostZone.RedemptionsEnabled)
	}
}

func (s *UpgradeTestSuite) TestMigrateDepositRecords() {
	// Create initial deposit records across each status
	testCases := []struct {
		status                          recordstypes.DepositRecord_Status
		expectedDelegationTxsInProgress uint64
	}{
		{
			status:                          recordstypes.DepositRecord_TRANSFER_QUEUE,
			expectedDelegationTxsInProgress: 0,
		},
		{
			status:                          recordstypes.DepositRecord_TRANSFER_IN_PROGRESS,
			expectedDelegationTxsInProgress: 0,
		},
		{
			status:                          recordstypes.DepositRecord_DELEGATION_QUEUE,
			expectedDelegationTxsInProgress: 0,
		},
		{
			status:                          recordstypes.DepositRecord_DELEGATION_IN_PROGRESS,
			expectedDelegationTxsInProgress: 1,
		},
	}

	for id, tc := range testCases {
		s.App.RecordsKeeper.SetDepositRecord(s.Ctx, recordstypes.DepositRecord{
			Id:     uint64(id),
			Status: tc.status,
		})
	}

	// Migrate the records
	v24.MigrateDepositRecords(s.Ctx, s.App.RecordsKeeper)

	// Confirm the expected status for each
	for id, tc := range testCases {
		depositRecord, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, uint64(id))
		s.Require().True(found, "deposit record %d should have been found", id)
		s.Require().Equal(tc.expectedDelegationTxsInProgress, depositRecord.DelegationTxsInProgress,
			"delegation txs in progress for record %d", id)
	}
}

func (s *UpgradeTestSuite) TestMigrateEpochUnbondingRecords() {
	recordTestCases := []struct {
		epochNumber                       uint64
		chainId                           string
		status                            recordstypes.HostZoneUnbonding_Status
		stTokenAmount                     int64
		nativeTokenAmount                 int64
		expectedStTokensToBurn            int64
		expectedNativeTokensToUnbond      int64
		expectedClaimableNativeTokens     int64
		expectedUndelegationTxsInProgress uint64
	}{
		{
			epochNumber: 1,
			chainId:     "chain-1",
			status:      recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,

			stTokenAmount:     1,
			nativeTokenAmount: 2,

			expectedStTokensToBurn:            0,
			expectedNativeTokensToUnbond:      0,
			expectedClaimableNativeTokens:     0,
			expectedUndelegationTxsInProgress: 0,
		},
		{
			epochNumber: 1,
			chainId:     "chain-2",
			status:      recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,

			stTokenAmount:     3,
			nativeTokenAmount: 4,

			expectedStTokensToBurn:            3,
			expectedNativeTokensToUnbond:      4,
			expectedClaimableNativeTokens:     0,
			expectedUndelegationTxsInProgress: 1,
		},
		{
			epochNumber: 2,
			chainId:     "chain-3",
			status:      recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,

			stTokenAmount:     5,
			nativeTokenAmount: 6,

			expectedStTokensToBurn:            0,
			expectedNativeTokensToUnbond:      0,
			expectedClaimableNativeTokens:     0,
			expectedUndelegationTxsInProgress: 0,
		},
		{
			epochNumber: 2,
			chainId:     "chain-4",
			status:      recordstypes.HostZoneUnbonding_EXIT_TRANSFER_IN_PROGRESS,

			stTokenAmount:     7,
			nativeTokenAmount: 8,

			expectedStTokensToBurn:            0,
			expectedNativeTokensToUnbond:      0,
			expectedClaimableNativeTokens:     0,
			expectedUndelegationTxsInProgress: 0,
		},
		{
			epochNumber: 4,
			chainId:     "chain-5",
			status:      recordstypes.HostZoneUnbonding_CLAIMABLE,

			stTokenAmount:     9,
			nativeTokenAmount: 10,

			expectedStTokensToBurn:            0,
			expectedNativeTokensToUnbond:      0,
			expectedClaimableNativeTokens:     10,
			expectedUndelegationTxsInProgress: 0,
		},
	}

	// Create the initial host zone unbonding and epoch unbonding records
	for _, tc := range recordTestCases {
		if _, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, tc.epochNumber); !found {
			s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordstypes.EpochUnbondingRecord{
				EpochNumber: tc.epochNumber,
			})
		}

		hostZoneUnbonding := recordstypes.HostZoneUnbonding{
			HostZoneId:        tc.chainId,
			Status:            tc.status,
			StTokenAmount:     sdkmath.NewInt(tc.stTokenAmount),
			NativeTokenAmount: sdkmath.NewInt(tc.nativeTokenAmount),
		}
		err := s.App.RecordsKeeper.SetHostZoneUnbondingRecord(s.Ctx, tc.epochNumber, tc.chainId, hostZoneUnbonding)
		s.Require().NoError(err, "no error expected when creating epoch unbonding records")
	}

	// Call migration function
	v24.MigrateEpochUnbondingRecords(s.Ctx, s.App.RecordsKeeper)

	// Confirm new fields were added
	for _, tc := range recordTestCases {
		hostZoneUnbonding, found := s.App.RecordsKeeper.GetHostZoneUnbondingByChainId(s.Ctx, tc.epochNumber, tc.chainId)
		s.Require().True(found, "host zone unbonding should have been found for %d and %s", tc.epochNumber, tc.chainId)

		s.Require().Equal(tc.expectedStTokensToBurn, hostZoneUnbonding.StTokensToBurn.Int64(),
			"%s stTokens to burn", tc.chainId)
		s.Require().Equal(tc.expectedNativeTokensToUnbond, hostZoneUnbonding.NativeTokensToUnbond.Int64(),
			"%s native to unbond", tc.chainId)
		s.Require().Equal(tc.expectedClaimableNativeTokens, hostZoneUnbonding.ClaimableNativeTokens.Int64(),
			"%s claimable native", tc.chainId)
		s.Require().Equal(tc.expectedUndelegationTxsInProgress, hostZoneUnbonding.UndelegationTxsInProgress,
			"%s undelegation txs in progress", tc.chainId)
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
			ChainId:                        v24.OsmosisChainId,
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
