package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibctypes "github.com/Stride-Labs/stride/v24/x/stakeibc/types"
	"github.com/Stride-Labs/stride/v24/x/staketia/types"
)

// ----------------------------------------------
//            MsgConfirmDelegation
// ----------------------------------------------

func (s *KeeperTestSuite) SetupDelegationRecordsAndHostZone() {
	s.SetupDelegationRecords()

	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	hostZone := s.initializeHostZone()
	hostZone.OperatorAddressOnStride = operatorAddress
	hostZone.SafeAddressOnStride = safeAddress
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)
}

// Verify that ConfirmDelegation succeeds, and non-admins cannot call it
func (s *KeeperTestSuite) TestConfirmDelegation() {
	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	nonAdminAddress := s.TestAccs[2].String()

	// Confirm that ConfirmDelegation can be called by the operator address
	s.SetupDelegationRecordsAndHostZone()
	msgConfirmDelegationOperator := types.MsgConfirmDelegation{
		Operator: operatorAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err := s.GetMsgServer().ConfirmDelegation(s.Ctx, &msgConfirmDelegationOperator)
	s.Require().NoError(err, "operator should be able to confirm delegation")

	// Confirm that ConfirmDelegation can be called by the safe address
	s.SetupDelegationRecordsAndHostZone()
	msgConfirmDelegationSafe := types.MsgConfirmDelegation{
		Operator: safeAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err = s.GetMsgServer().ConfirmDelegation(s.Ctx, &msgConfirmDelegationSafe)
	s.Require().NoError(err, "safe should be able to confirm delegation")

	// Confirm that ConfirmDelegation cannot be called by a non-admin address
	s.SetupDelegationRecordsAndHostZone()
	msgConfirmDelegationNonAdmin := types.MsgConfirmDelegation{
		Operator: nonAdminAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err = s.GetMsgServer().ConfirmDelegation(s.Ctx, &msgConfirmDelegationNonAdmin)
	s.Require().Error(err, "non-admin should not be able to confirm delegation")
}

// ----------------------------------------------
//           MsgConfirmUndelegation
// ----------------------------------------------

// This function is primarily covered by the keeper's unit test
// This test just validates the address
func (s *KeeperTestSuite) TestConfirmUndelegation() {
	operatorAddress := "operator"
	amountToUndelegate := sdkmath.NewInt(100)
	tc := s.SetupTestConfirmUndelegation(amountToUndelegate)

	// Store the operator address on the host zone
	hostZone := s.MustGetHostZone()
	hostZone.OperatorAddressOnStride = "operator"
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)

	validMsg := types.MsgConfirmUndelegation{
		Operator: operatorAddress,
		RecordId: tc.unbondingRecord.Id,
		TxHash:   ValidTxHashDefault,
	}

	invalidMsg := validMsg
	invalidMsg.Operator = "invalid_address"

	// confirm the valid tx was successful
	_, err := s.GetMsgServer().ConfirmUndelegation(sdk.UnwrapSDKContext(s.Ctx), &validMsg)
	s.Require().NoError(err, "no error expected during confirm undelegation")

	// confirm the invalid tx failed because it was not submitted by the operator
	_, err = s.GetMsgServer().ConfirmUndelegation(sdk.UnwrapSDKContext(s.Ctx), &invalidMsg)
	s.Require().ErrorContains(err, "invalid admin address")
}

// ----------------------------------------------
//         MsgConfirmUnbondingTokensSweep
// ----------------------------------------------

func (s *KeeperTestSuite) SetupUnbondingRecordsAndHostZone() {
	s.SetupTestConfirmUnbondingTokens(DefaultClaimFundingAmount)

	// ger relevant variables
	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	hostZone := s.MustGetHostZone()

	// set host zone
	hostZone.OperatorAddressOnStride = operatorAddress
	hostZone.SafeAddressOnStride = safeAddress
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, hostZone)
}

// Verify that ConfirmUnbondingTokenSweep succeeds, and non-admins cannot call it
func (s *KeeperTestSuite) TestConfirmUnbondingTokenSweep() {
	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	nonAdminAddress := s.TestAccs[2].String()

	// Confirm that ConfirmDelegation can be called by the operator address
	s.SetupUnbondingRecordsAndHostZone()
	MsgConfirmUnbondedTokenSweepOperator := types.MsgConfirmUnbondedTokenSweep{
		Operator: operatorAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err := s.GetMsgServer().ConfirmUnbondedTokenSweep(s.Ctx, &MsgConfirmUnbondedTokenSweepOperator)
	s.Require().NoError(err, "operator should be able to confirm unbonded token sweep")

	// Confirm that ConfirmDelegation can be called by the safe address
	s.SetupUnbondingRecordsAndHostZone()
	msgConfirmUnbondedTokenSweepSafe := types.MsgConfirmUnbondedTokenSweep{
		Operator: safeAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err = s.GetMsgServer().ConfirmUnbondedTokenSweep(s.Ctx, &msgConfirmUnbondedTokenSweepSafe)
	s.Require().NoError(err, "safe should be able to confirm unbonded token sweep")

	// Confirm that ConfirmDelegation cannot be called by a non-admin address
	s.SetupUnbondingRecordsAndHostZone()
	msgConfirmUnbondedTokenSweepNonAdmin := types.MsgConfirmUnbondedTokenSweep{
		Operator: nonAdminAddress,
		RecordId: 6,
		TxHash:   ValidTxHashNew,
	}
	_, err = s.GetMsgServer().ConfirmUnbondedTokenSweep(s.Ctx, &msgConfirmUnbondedTokenSweepNonAdmin)
	s.Require().ErrorIs(err, types.ErrInvalidAdmin, "non-admin should not be able to confirm unbonded token sweep")
}

// ----------------------------------------------
//           MsgAdjustDelegatedBalance
// ----------------------------------------------

func (s *KeeperTestSuite) TestAdjustDelegatedBalance() {

	safeAddress := "safe"

	// Create the host zones
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		SafeAddressOnStride:       safeAddress,
		RemainingDelegatedBalance: sdk.NewInt(0),
	})
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, stakeibctypes.HostZone{
		ChainId:          types.CelestiaChainId,
		TotalDelegations: sdk.NewInt(0),
	})

	// we're halting the zone to test that the tx works even when the host zone is halted
	s.App.StaketiaKeeper.HaltZone(s.Ctx)

	// Call adjust for each test case and confirm the ending delegation
	testCases := []struct {
		address       string
		offset        sdkmath.Int
		endDelegation sdkmath.Int
	}{
		{address: "valA", offset: sdkmath.NewInt(10), endDelegation: sdkmath.NewInt(10)}, // 0 + 10 = 10
		{address: "valB", offset: sdkmath.NewInt(-5), endDelegation: sdkmath.NewInt(5)},  // 10 - 5 = 5
		{address: "valC", offset: sdkmath.NewInt(8), endDelegation: sdkmath.NewInt(13)},  // 5 + 8 = 13
		{address: "valD", offset: sdkmath.NewInt(2), endDelegation: sdkmath.NewInt(15)},  // 13 + 2 = 15
		{address: "valE", offset: sdkmath.NewInt(-6), endDelegation: sdkmath.NewInt(9)},  // 15 - 6 = 9
	}
	for _, tc := range testCases {
		msg := types.MsgAdjustDelegatedBalance{
			Operator:         safeAddress,
			DelegationOffset: tc.offset,
			ValidatorAddress: tc.address,
		}
		_, err := s.GetMsgServer().AdjustDelegatedBalance(s.Ctx, &msg)
		s.Require().NoError(err, "no error expected when adjusting delegated bal properly for %s", tc.address)

		hostZone := s.MustGetHostZone()
		s.Require().Equal(tc.endDelegation, hostZone.RemainingDelegatedBalance, "remaining delegation after change for %s", tc.address)

		stakeibcHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, types.CelestiaChainId)
		s.Require().True(found)
		s.Require().Equal(tc.endDelegation, stakeibcHostZone.TotalDelegations, "total delegation after change for %s", tc.address)
	}

	// Attempt to call it with an amount that would make it negative, it should fail
	_, err := s.GetMsgServer().AdjustDelegatedBalance(s.Ctx, &types.MsgAdjustDelegatedBalance{
		Operator:         safeAddress,
		DelegationOffset: sdk.NewInt(-10000),
	})
	s.Require().ErrorContains(err, "offset would cause the delegated balance to be negative")

	// Attempt to call it from a different address, it should fail
	_, err = s.GetMsgServer().AdjustDelegatedBalance(s.Ctx, &types.MsgAdjustDelegatedBalance{
		Operator: s.TestAccs[0].String(),
	})
	s.Require().ErrorContains(err, "invalid safe address")

	// Remove the host zone and try again, it should fail
	s.App.StaketiaKeeper.RemoveHostZone(s.Ctx)
	_, err = s.GetMsgServer().AdjustDelegatedBalance(s.Ctx, &types.MsgAdjustDelegatedBalance{})
	s.Require().ErrorContains(err, "host zone not found")
}

// ----------------------------------------------
//         MsgOverwriteDelgationRecord
// ----------------------------------------------

func (s *KeeperTestSuite) TestOverwriteDelegationRecord() {
	safeAddress := "safe"
	recordId := uint64(1)

	// Create a host zone with a safe admin
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		SafeAddressOnStride: safeAddress,
	})

	// Create an initial delegation record, and a record to be overridden
	initialDelegationRecord := types.DelegationRecord{
		Id:           recordId,
		NativeAmount: sdkmath.NewInt(1000),
		Status:       types.TRANSFER_IN_PROGRESS,
		TxHash:       "initial-hash",
	}
	overrideDelegationRecord := types.DelegationRecord{
		Id:           recordId,
		NativeAmount: sdkmath.NewInt(2000),
		Status:       types.DELEGATION_QUEUE,
		TxHash:       "override-hash",
	}
	s.App.StaketiaKeeper.SetDelegationRecord(s.Ctx, initialDelegationRecord)

	// Attempt to override the delegation record from a non-safe address - it should fail
	msg := types.MsgOverwriteDelegationRecord{
		Creator:          "non-admin",
		DelegationRecord: &overrideDelegationRecord,
	}
	_, err := s.GetMsgServer().OverwriteDelegationRecord(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "invalid safe address")

	// Check that the record was not updated
	recordAfterFailedTx, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, recordId)
	s.Require().True(found, "record should not have been removed")
	s.Require().Equal(initialDelegationRecord, recordAfterFailedTx, "record should not have been overridden")

	// Attempt to override from the safe address - it should succeed
	msg = types.MsgOverwriteDelegationRecord{
		Creator:          safeAddress,
		DelegationRecord: &overrideDelegationRecord,
	}
	_, err = s.GetMsgServer().OverwriteDelegationRecord(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when overriding record")

	// Check that the record was updated
	recordAfterSuccessfulTx, found := s.App.StaketiaKeeper.GetDelegationRecord(s.Ctx, recordId)
	s.Require().True(found, "record should not have been removed")
	s.Require().Equal(overrideDelegationRecord, recordAfterSuccessfulTx, "record should have been overridden")
}

// ----------------------------------------------
//         MsgOverwriteUnbondingRecord
// ----------------------------------------------

func (s *KeeperTestSuite) TestOverwriteUnbondingRecord() {
	safeAddress := "safe"
	recordId := uint64(1)

	// Create a host zone with a safe admin
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		SafeAddressOnStride: safeAddress,
	})

	// Create an initial unbonding record, and a record to be overridden
	initialUnbondingRecord := types.UnbondingRecord{
		Id:                             recordId,
		NativeAmount:                   sdkmath.NewInt(1000),
		StTokenAmount:                  sdkmath.NewInt(1000),
		Status:                         types.UNBONDING_IN_PROGRESS,
		UnbondingCompletionTimeSeconds: 100,
		UndelegationTxHash:             "initial-hash-1",
		UnbondedTokenSweepTxHash:       "initial-hash-2",
	}
	overrideUnbondingRecord := types.UnbondingRecord{
		Id:                             recordId,
		NativeAmount:                   sdkmath.NewInt(2000),
		StTokenAmount:                  sdkmath.NewInt(2000),
		Status:                         types.UNBONDED,
		UnbondingCompletionTimeSeconds: 200,
		UndelegationTxHash:             "override-hash-1",
		UnbondedTokenSweepTxHash:       "override-hash-2",
	}
	s.App.StaketiaKeeper.SetUnbondingRecord(s.Ctx, initialUnbondingRecord)

	// Attempt to override the unbonding record from a non-safe address - it should fail
	msg := types.MsgOverwriteUnbondingRecord{
		Creator:         "non-admin",
		UnbondingRecord: &overrideUnbondingRecord,
	}
	_, err := s.GetMsgServer().OverwriteUnbondingRecord(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "invalid safe address")

	// Check that the record was not updated
	recordAfterFailedTx, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, recordId)
	s.Require().True(found, "record should not have been removed")
	s.Require().Equal(initialUnbondingRecord, recordAfterFailedTx, "record should not have been overridden")

	// Attempt to override from the safe address - it should succeed
	msg = types.MsgOverwriteUnbondingRecord{
		Creator:         safeAddress,
		UnbondingRecord: &overrideUnbondingRecord,
	}
	_, err = s.GetMsgServer().OverwriteUnbondingRecord(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when overriding record")

	// Check that the record was updated
	recordAfterSuccessfulTx, found := s.App.StaketiaKeeper.GetUnbondingRecord(s.Ctx, recordId)
	s.Require().True(found, "record should not have been removed")
	s.Require().Equal(overrideUnbondingRecord, recordAfterSuccessfulTx, "record should have been overridden")
}

// ----------------------------------------------
//         MsgOverwriteRedemptionRecord
// ----------------------------------------------

func (s *KeeperTestSuite) TestOverwriteRedemptionRecord() {
	safeAddress := "safe"
	recordId := uint64(1)
	redeemer := "redeemer"

	// Create a host zone with a safe admin
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, types.HostZone{
		SafeAddressOnStride: safeAddress,
	})

	// Create an initial redemption record, and a record to be overridden
	initialRedemptionRecord := types.RedemptionRecord{
		UnbondingRecordId: recordId,
		Redeemer:          redeemer,
		NativeAmount:      sdkmath.NewInt(1000),
		StTokenAmount:     sdkmath.NewInt(1000),
	}
	overrideRedemptionRecord := types.RedemptionRecord{
		UnbondingRecordId: recordId,
		Redeemer:          redeemer,
		NativeAmount:      sdkmath.NewInt(2000),
		StTokenAmount:     sdkmath.NewInt(2000),
	}
	s.App.StaketiaKeeper.SetRedemptionRecord(s.Ctx, initialRedemptionRecord)

	// Attempt to override the redemption record from a non-safe address - it should fail
	msg := types.MsgOverwriteRedemptionRecord{
		Creator:          "non-admin",
		RedemptionRecord: &overrideRedemptionRecord,
	}
	_, err := s.GetMsgServer().OverwriteRedemptionRecord(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().ErrorContains(err, "invalid safe address")

	// Check that the record was not updated
	recordAfterFailedTx, found := s.App.StaketiaKeeper.GetRedemptionRecord(s.Ctx, recordId, redeemer)
	s.Require().True(found, "record should not have been removed")
	s.Require().Equal(initialRedemptionRecord, recordAfterFailedTx, "record should not have been overridden")

	// Attempt to override from the safe address - it should succeed
	msg = types.MsgOverwriteRedemptionRecord{
		Creator:          safeAddress,
		RedemptionRecord: &overrideRedemptionRecord,
	}
	_, err = s.GetMsgServer().OverwriteRedemptionRecord(sdk.UnwrapSDKContext(s.Ctx), &msg)
	s.Require().NoError(err, "no error expected when overriding record")

	// Check that the record was updated
	recordAfterSuccessfulTx, found := s.App.StaketiaKeeper.GetRedemptionRecord(s.Ctx, recordId, redeemer)
	s.Require().True(found, "record should not have been removed")
	s.Require().Equal(overrideRedemptionRecord, recordAfterSuccessfulTx, "record should have been overridden")
}

// ----------------------------------------------
//            MsgSetOperatorAddress
// ----------------------------------------------

// Verify that operator address can be set successfully
func (s *KeeperTestSuite) TestSetOperatorAddress() {

	safeAddress := s.TestAccs[0].String()
	operatorAddress := s.TestAccs[1].String()
	nonAdminAddress := s.TestAccs[2].String()

	// set the host zone
	zone := types.HostZone{
		SafeAddressOnStride:     safeAddress,
		OperatorAddressOnStride: operatorAddress,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)

	// Set the operator address, signed by the SAFE address
	msgSetOperatorAddress := types.MsgSetOperatorAddress{
		Signer:   safeAddress,
		Operator: nonAdminAddress,
	}

	_, err := s.GetMsgServer().SetOperatorAddress(s.Ctx, &msgSetOperatorAddress)
	s.Require().NoError(err, "should not throw an error")

	// Confirm the operator address was updated
	zone, err = s.App.StaketiaKeeper.GetHostZone(s.Ctx)
	s.Require().NoError(err, "should not throw an error")
	s.Require().Equal(s.TestAccs[2].String(), zone.OperatorAddressOnStride, "operator address should be set")

	// Confirm the operator address cannot be set by a non-safe address
	msgSetOperatorAddressWrongSafe := types.MsgSetOperatorAddress{
		Signer:   operatorAddress,
		Operator: nonAdminAddress,
	}
	s.App.StaketiaKeeper.SetHostZone(s.Ctx, zone)
	_, err = s.GetMsgServer().SetOperatorAddress(s.Ctx, &msgSetOperatorAddressWrongSafe)
	s.Require().Error(err, "invalid safe address")
}
