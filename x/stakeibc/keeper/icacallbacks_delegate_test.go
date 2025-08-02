package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v28/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v28/x/records/types"
	"github.com/Stride-Labs/stride/v28/x/stakeibc/types"
)

type DelegateCallbackTestCase struct {
	initialHostZone       types.HostZone
	initialDepositRecord  recordtypes.DepositRecord
	totalTx1NewDelegation sdkmath.Int
	totalTx2NewDelegation sdkmath.Int
	splitDelegationsTx1   []*types.SplitDelegation
	splitDelegationsTx2   []*types.SplitDelegation
}

func (s *KeeperTestSuite) SetupDelegateCallback() DelegateCallbackTestCase {
	// Test Setup
	//  - 1_000_000 total delegated, 400_000 to val1, 600_000 to val2
	//  - Deposit record: 500_000
	//  - Callback 1: delegated 100_000 to val1 and 100_000 to val2
	//  - Callback 2: delegated 200_000 to val1 and 100_000 to val2
	totalDelegation := sdkmath.NewInt(1_000_000)
	val1InitialDelegation := sdkmath.NewInt(400_000)
	val2InitialDelegation := sdkmath.NewInt(600_000)

	totalNewDelegation := sdkmath.NewInt(500_000)
	totalTx1NewDelegation := sdkmath.NewInt(300_000)
	totalTx2NewDelegation := sdkmath.NewInt(200_000)

	val1Tx1Delegation := sdkmath.NewInt(100_000)
	val2Tx1Delegation := sdkmath.NewInt(200_000)
	val1Tx2Delegation := sdkmath.NewInt(100_000)
	val2Tx2Delegation := sdkmath.NewInt(100_000)

	// Create the validators with their initial delegations and the host zone
	validators := []*types.Validator{
		{Address: "val1", Delegation: val1InitialDelegation, DelegationChangesInProgress: 2},
		{Address: "val2", Delegation: val2InitialDelegation, DelegationChangesInProgress: 2},
	}
	hostZone := types.HostZone{
		ChainId:          HostChainId,
		HostDenom:        Atom,
		Validators:       validators,
		TotalDelegations: totalDelegation,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Create the deposit record
	depositRecord := recordtypes.DepositRecord{
		Id:                      DepositRecordId,
		HostZoneId:              HostChainId,
		Amount:                  totalNewDelegation,
		Status:                  recordtypes.DepositRecord_DELEGATION_QUEUE,
		DelegationTxsInProgress: 2,
	}
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)

	// Mock the callback args
	splitDelegations1 := []*types.SplitDelegation{
		{Validator: validators[0].Address, Amount: val1Tx1Delegation},
		{Validator: validators[1].Address, Amount: val2Tx1Delegation},
	}
	splitDelegations2 := []*types.SplitDelegation{
		{Validator: validators[0].Address, Amount: val1Tx2Delegation},
		{Validator: validators[1].Address, Amount: val2Tx2Delegation},
	}

	return DelegateCallbackTestCase{
		initialHostZone:       hostZone,
		initialDepositRecord:  depositRecord,
		totalTx1NewDelegation: totalTx1NewDelegation,
		totalTx2NewDelegation: totalTx2NewDelegation,
		splitDelegationsTx1:   splitDelegations1,
		splitDelegationsTx2:   splitDelegations2,
	}
}

// Helper function to call the delegate callback function with relevant args
func (s *KeeperTestSuite) delegateCallback(status icacallbacktypes.AckResponseStatus, splitDelegations []*types.SplitDelegation) error {
	packet := channeltypes.Packet{}
	ackResponse := &icacallbacktypes.AcknowledgementResponse{
		Status: status,
	}
	callbackArgs := types.DelegateCallback{
		HostZoneId:       HostChainId,
		DepositRecordId:  DepositRecordId,
		SplitDelegations: splitDelegations,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	s.Require().NoError(err)

	return s.App.StakeibcKeeper.DelegateCallback(s.Ctx, packet, ackResponse, callbackArgsBz)
}

func (s *KeeperTestSuite) TestDelegateCallback_Successful() {
	// This test will test two consecutive callbacks from different ICAs
	// The first callback will partially decrement the deposit record and the second
	// callback will remove the deposit record
	tc := s.SetupDelegateCallback()

	// Execute the callback the first time
	err := s.delegateCallback(icacallbacktypes.AckResponseStatus_SUCCESS, tc.splitDelegationsTx1)
	s.Require().NoError(err)

	// Confirm total delegation has increased
	hostZone := s.MustGetHostZone(HostChainId)
	expectedUpdatedDelegation := tc.initialHostZone.TotalDelegations.Add(tc.totalTx1NewDelegation)
	s.Require().Equal(expectedUpdatedDelegation.Int64(), hostZone.TotalDelegations.Int64(), "total delegation after first callback")

	// Confirm delegations have been added to validators
	validator1 := hostZone.Validators[0]
	validator2 := hostZone.Validators[1]
	expectedVal1Delegation := tc.initialHostZone.Validators[0].Delegation.Add(tc.splitDelegationsTx1[0].Amount)
	expectedVal2Delegation := tc.initialHostZone.Validators[1].Delegation.Add(tc.splitDelegationsTx1[1].Amount)
	s.Require().Equal(expectedVal1Delegation.Int64(), validator1.Delegation.Int64(), "val1 delegation after first callback")
	s.Require().Equal(expectedVal2Delegation.Int64(), validator2.Delegation.Int64(), "val2 delegation after first callback")

	// Confirm the number of delegations in progress has decreased
	s.Require().Equal(1, int(validator1.DelegationChangesInProgress), "val1 delegation changes in progress first callback")
	s.Require().Equal(1, int(validator1.DelegationChangesInProgress), "val2 delegation changes in progress first callback")

	// Confirm deposit record was decremented
	depositRecord := s.MustGetDepositRecord(DepositRecordId)
	expectedDepositRecordAmount := tc.initialDepositRecord.Amount.Sub(tc.totalTx1NewDelegation)
	s.Require().Equal(expectedDepositRecordAmount.Int64(), depositRecord.Amount.Int64(), "deposit record after first callback")
	s.Require().Equal(1, int(depositRecord.DelegationTxsInProgress), "deposit record delegation txs in progress first callback")

	// Execute the callback again for the second tx, it should delegate the remainder
	err = s.delegateCallback(icacallbacktypes.AckResponseStatus_SUCCESS, tc.splitDelegationsTx2)
	s.Require().NoError(err)

	// Confirm total delegation has increased
	hostZone = s.MustGetHostZone(HostChainId)
	expectedUpdatedDelegation = expectedUpdatedDelegation.Add(tc.totalTx2NewDelegation)
	s.Require().Equal(expectedUpdatedDelegation.Int64(), hostZone.TotalDelegations.Int64(), "total delegation after second callback")

	// Confirm delegations have been added to validators
	validator1 = hostZone.Validators[0]
	validator2 = hostZone.Validators[1]
	expectedVal1Delegation = expectedVal1Delegation.Add(tc.splitDelegationsTx2[0].Amount)
	expectedVal2Delegation = expectedVal2Delegation.Add(tc.splitDelegationsTx2[1].Amount)
	s.Require().Equal(expectedVal1Delegation.Int64(), validator1.Delegation.Int64(), "val1 delegation after second callback")
	s.Require().Equal(expectedVal2Delegation.Int64(), validator2.Delegation.Int64(), "val2 delegation after second callback")

	// Confirm the number of delegations in progress has decreased
	s.Require().Zero(validator1.DelegationChangesInProgress, "val1 delegation changes in progress second callback")
	s.Require().Zero(validator1.DelegationChangesInProgress, "val2 delegation changes in progress second callback")

	// Confirm deposit record was removed
	_, found := s.App.RecordsKeeper.GetDepositRecord(s.Ctx, DepositRecordId)
	s.Require().False(found, "deposit record should have been removed")
}

func (s *KeeperTestSuite) checkDelegateStateIfCallbackFailed(tc DelegateCallbackTestCase) {
	// Confirm total delegation has not increased
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(tc.initialHostZone.TotalDelegations, hostZone.TotalDelegations, "total delegation should not have increased")

	// Confirm the validator delegations did not change
	validator1 := hostZone.Validators[0]
	validator2 := hostZone.Validators[1]
	expectedVal1Delegation := tc.initialHostZone.Validators[0].Delegation
	expectedVal2Delegation := tc.initialHostZone.Validators[1].Delegation
	s.Require().Equal(expectedVal1Delegation.Int64(), validator1.Delegation.Int64(), "val1 delegation should not change")
	s.Require().Equal(expectedVal2Delegation.Int64(), validator2.Delegation.Int64(), "val2 delegation should not change")

	// Confirm the number of delegations in progress has decreased
	s.Require().Equal(1, int(hostZone.Validators[0].DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(1, int(hostZone.Validators[1].DelegationChangesInProgress), "val2 delegation changes in progress")

	// Confirm deposit record has had the txs in progress decremented, but the amount untouched
	depositRecord := s.MustGetDepositRecord(DepositRecordId)
	s.Require().Equal(1, int(depositRecord.DelegationTxsInProgress), "delegation tx in progress should have decremented")
	s.Require().Equal(tc.initialDepositRecord.Amount, depositRecord.Amount, "deposit record amount should not have changed")
	s.Require().Equal(tc.initialDepositRecord.Status, depositRecord.Status, "deposit record status should not have changed")
}

func (s *KeeperTestSuite) TestDelegateCallback_AckTimeout() {
	tc := s.SetupDelegateCallback()

	// Call the callback with a timeout ack
	err := s.delegateCallback(icacallbacktypes.AckResponseStatus_TIMEOUT, tc.splitDelegationsTx1)
	s.Require().NoError(err)
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_AckError() {
	tc := s.SetupDelegateCallback()

	// Call the callback with a failure ack
	err := s.delegateCallback(icacallbacktypes.AckResponseStatus_FAILURE, tc.splitDelegationsTx1)
	s.Require().NoError(err)
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_WrongCallbackArgs() {
	s.SetupDelegateCallback()

	// random args should cause the callback to fail
	packet := channeltypes.Packet{}
	ackResponse := &icacallbacktypes.AcknowledgementResponse{}
	invalidCallbackArgs := []byte("random bytes")

	err := s.App.StakeibcKeeper.DelegateCallback(s.Ctx, packet, ackResponse, invalidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal delegate callback")
}

func (s *KeeperTestSuite) TestDelegateCallback_HostNotFound() {
	tc := s.SetupDelegateCallback()

	// Remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	err := s.delegateCallback(icacallbacktypes.AckResponseStatus_SUCCESS, tc.splitDelegationsTx1)
	s.Require().ErrorContains(err, "host zone not found GAIA")
}

func (s *KeeperTestSuite) TestDelegateCallback_MissingValidator() {
	tc := s.SetupDelegateCallback()

	// Update the callback args such that a validator is missing
	invalidSplitDelegation := tc.splitDelegationsTx1
	invalidSplitDelegation[0].Validator = "validator_dne"

	err := s.delegateCallback(icacallbacktypes.AckResponseStatus_SUCCESS, invalidSplitDelegation)
	s.Require().ErrorContains(err, "validator not found")
}
