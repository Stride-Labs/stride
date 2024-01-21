package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/app/apptesting"
	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

type DistributeClaimsTestCase struct {
	claimAddress              sdk.AccAddress
	claimableRecordIds        []uint64
	expectedFinalClaimBalance sdkmath.Int
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

	// Loop again and confirm all users received their tokens
	for i, redemption := range redemptionRecords {
		if redemption.UnbondingRecordId != distributedUnbondingId {
			continue
		}
		redeemer := accounts[i]
		userBalance := s.App.BankKeeper.GetBalance(s.Ctx, redeemer, HostIBCDenom)
		s.Require().Equal(redemption.NativeAmount.Int64(), userBalance.Amount.Int64(),
			"user %d should have received their native tokens", i)
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
