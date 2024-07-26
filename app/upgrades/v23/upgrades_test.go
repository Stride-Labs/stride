package v23_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v23/app/apptesting"
	v23 "github.com/Stride-Labs/stride/v23/app/upgrades/v23"
	recordstypes "github.com/Stride-Labs/stride/v23/x/records/types"
	stakeibctypes "github.com/Stride-Labs/stride/v23/x/stakeibc/types"
)

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
}

func (s *UpgradeTestSuite) TestMigrateTradeRoutes() {
	minTransferAmount := sdkmath.NewInt(100)

	// Create a trade route with the deprecated trade config
	tradeRoutes := stakeibctypes.TradeRoute{
		HostDenomOnHostZone:     "host-denom",
		RewardDenomOnRewardZone: "reward-denom",

		TradeConfig: stakeibctypes.TradeConfig{ //nolint:staticcheck
			MinSwapAmount: minTransferAmount,
		},
	}
	s.App.StakeibcKeeper.SetTradeRoute(s.Ctx, tradeRoutes)

	// Call migration function
	v23.MigrateTradeRoutes(s.Ctx, s.App.StakeibcKeeper)

	// Confirm trade route was migrated
	for _, tradeRoute := range s.App.StakeibcKeeper.GetAllTradeRoutes(s.Ctx) {
		s.Require().Equal(tradeRoute.MinTransferAmount, minTransferAmount)
	}
}

func (s *UpgradeTestSuite) TestMigrateHostZones() {
	// Create a host zone with redemptions enabled set to false
	hostZone := stakeibctypes.HostZone{
		ChainId:            "chain-0",
		RedemptionsEnabled: false,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Call migration function
	v23.MigrateHostZones(s.Ctx, s.App.StakeibcKeeper)

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
	v23.MigrateDepositRecords(s.Ctx, s.App.RecordsKeeper)

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
	v23.MigrateEpochUnbondingRecords(s.Ctx, s.App.RecordsKeeper)

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
