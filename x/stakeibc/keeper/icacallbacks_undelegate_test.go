package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	icacallbacktypes "github.com/Stride-Labs/stride/v26/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"
)

type UndelegateCallbackState struct {
	totalDelegations   sdkmath.Int
	val1Bal            sdkmath.Int
	val2Bal            sdkmath.Int
	epochNumbers       []uint64
	completionTime     time.Time
	callbackArgs       types.UndelegateCallback
	zoneAccountBalance sdkmath.Int
}

type UndelegateCallbackArgs struct {
	packet      channeltypes.Packet
	ackResponse *icacallbacktypes.AcknowledgementResponse
	args        []byte
}

type UndelegateCallbackTestCase struct {
	initialState           UndelegateCallbackState
	validArgs              UndelegateCallbackArgs
	val1UndelegationAmount sdkmath.Int
	val2UndelegationAmount sdkmath.Int
	totalUndelegated       sdkmath.Int
}

func (s *KeeperTestSuite) SetupUndelegateCallback() UndelegateCallbackTestCase {
	// Test setup is as follows:
	//   Total Stake:     1_000_000
	//    - Val1 Stake:     400_000
	//    - Val2 Stake:     600_000
	//   Total Unbonded:    500_000
	//    - From Val1:      100_000
	//    - From Val2:      400_000
	//   Deposit Account
	//    - Initial Balance: 600_000
	//    - Final Balance:   100_000
	initialTotalDelegations := sdkmath.NewInt(1_000_000)
	initialVal1Delegation := sdkmath.NewInt(400_000)
	initialVal2Delegation := sdkmath.NewInt(600_000)

	totalUndelegated := sdkmath.NewInt(500_000)
	val1UndelegationAmount := sdkmath.NewInt(100_000)
	val2UndelegationAmount := sdkmath.NewInt(400_000)

	initialDepositAccountBalance := sdkmath.NewInt(600_000)

	// Create the host zone and validators
	depositAccount := s.TestAccs[0]
	validators := []*types.Validator{
		{Address: "val1", Delegation: initialVal1Delegation, DelegationChangesInProgress: 1},
		{Address: "val2", Delegation: initialVal2Delegation, DelegationChangesInProgress: 1},
	}
	hostZone := types.HostZone{
		ChainId:          HostChainId,
		HostDenom:        Atom,
		Validators:       validators,
		TotalDelegations: initialTotalDelegations,
		DepositAddress:   depositAccount.String(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Create the host zone unbonding records
	// Record 1 will not have a unbonding time
	// Record 2 will have a unbonding time of 2024-01-02 already
	// The callback will have a unbonding time of 2024-01-01 so it should only update record 1
	completionTimeFromThisBatch := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)                    // 2024-01-01
	completionTimeFromPrevBatch := uint64(time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC).UnixNano()) // 2024-01-02
	hostZoneUnbonding1 := recordtypes.HostZoneUnbonding{
		HostZoneId:                HostChainId,
		Status:                    recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
		NativeTokenAmount:         sdkmath.NewInt(1_000_000),
		StTokenAmount:             sdkmath.NewInt(1_000_000), // Implied RR: 1.0
		NativeTokensToUnbond:      totalUndelegated,
		StTokensToBurn:            totalUndelegated,
		UnbondingTime:             uint64(0),
		UndelegationTxsInProgress: 1,
	}
	hostZoneUnbonding2 := recordtypes.HostZoneUnbonding{
		HostZoneId:                HostChainId,
		Status:                    recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
		NativeTokenAmount:         sdkmath.NewInt(1_000_000),
		StTokenAmount:             sdkmath.NewInt(1_000_000), // Implied RR: 2.0
		NativeTokensToUnbond:      totalUndelegated,
		StTokensToBurn:            totalUndelegated,
		UnbondingTime:             completionTimeFromPrevBatch,
		UndelegationTxsInProgress: 1,
	}

	// Create the epoch unbonding records with the host zone unbonding records
	epochNumbers := []uint64{1, 2}
	epochUnbondingRecords := []recordtypes.EpochUnbondingRecord{
		{
			EpochNumber:        1,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding1},
		},
		{
			EpochNumber:        2,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding2},
		},
	}
	for _, epochUnbondingRecord := range epochUnbondingRecords {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	}

	// Fund the deposit account with stTokens (some of which will be burned)
	s.FundAccount(depositAccount, sdk.NewCoin(StAtom, initialDepositAccountBalance))

	// Mock the ack response
	packet := channeltypes.Packet{}
	msgsUndelegateResponse := &stakingtypes.MsgUndelegateResponse{
		CompletionTime: completionTimeFromThisBatch,
	}
	msgsUndelegateResponseBz, err := proto.Marshal(msgsUndelegateResponse)
	s.Require().NoError(err, "no error expected when marshalling undelegate response")

	ackResponse := icacallbacktypes.AcknowledgementResponse{
		Status:       icacallbacktypes.AckResponseStatus_SUCCESS,
		MsgResponses: [][]byte{msgsUndelegateResponseBz},
	}

	// Build the callback args for each validator
	val1SplitDelegation := types.SplitUndelegation{
		Validator:         validators[0].Address,
		NativeTokenAmount: val1UndelegationAmount,
	}
	val2SplitDelegation := types.SplitUndelegation{
		Validator:         validators[1].Address,
		NativeTokenAmount: val2UndelegationAmount,
	}
	callbackArgs := types.UndelegateCallback{
		HostZoneId: HostChainId,
		SplitUndelegations: []*types.SplitUndelegation{
			&val1SplitDelegation,
			&val2SplitDelegation,
		},
		EpochUnbondingRecordIds: epochNumbers,
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	s.Require().NoError(err, "callback args unmarshalled")

	return UndelegateCallbackTestCase{
		val1UndelegationAmount: val1UndelegationAmount,
		val2UndelegationAmount: val2UndelegationAmount,
		totalUndelegated:       totalUndelegated,
		initialState: UndelegateCallbackState{
			callbackArgs:       callbackArgs,
			totalDelegations:   initialTotalDelegations,
			val1Bal:            initialVal1Delegation,
			val2Bal:            initialVal2Delegation,
			completionTime:     completionTimeFromThisBatch,
			zoneAccountBalance: initialDepositAccountBalance,
		},
		validArgs: UndelegateCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        callbackArgsBz,
		},
	}
}

func (s *KeeperTestSuite) TestUndelegateCallback_Successful() {
	tc := s.SetupUndelegateCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	// Callback
	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, validArgs.packet, validArgs.ackResponse, validArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds")

	// Check that total delegation has decreased on the host zone
	hostZone := s.MustGetHostZone(HostChainId)
	expectedTotalDelegation := initialState.totalDelegations.Sub(tc.totalUndelegated)
	s.Require().Equal(expectedTotalDelegation, hostZone.TotalDelegations, "total delegation has decreased on the host zone")

	// Check that Delegations on validators have decreased
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal.Sub(tc.val1UndelegationAmount), val1.Delegation, "val1 delegation has decreased")
	s.Require().Equal(initialState.val2Bal.Sub(tc.val2UndelegationAmount), val2.Delegation, "val2 delegation has decreased")

	// Check that the number of delegation changes in progress was reset to 0
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// Check that the host zone unbonding records have been updated
	for _, epochNumber := range initialState.epochNumbers {
		hzu := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)
		s.Require().Equal(initialState.completionTime.UnixNano(), int64(hzu.UnbondingTime), "completion time is set on the hzu")
		s.Require().Equal(recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, hzu.Status, "hzu status is set to EXIT_TRANSFER_QUEUE")
		s.Require().Zero(hzu.UndelegationTxsInProgress, "hzu undelegations in progress")
	}

	// Confirm stTokens were removed from the deposit account
	depositAccount := sdk.MustAccAddressFromBech32(hostZone.DepositAddress)
	depositBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAccount, StAtom).Amount
	s.Require().Equal(tc.totalUndelegated, initialState.zoneAccountBalance.Sub(depositBalance), "tokens are burned")
}

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackFailed(tc UndelegateCallbackTestCase, status icacallbacktypes.AckResponseStatus) {
	initialState := tc.initialState

	// Check that total delegation has NOT decreased on the host zone
	hostZone := s.MustGetHostZone(HostChainId)
	s.Require().Equal(initialState.totalDelegations, hostZone.TotalDelegations, "total delegation has NOT decreased on the host zone")

	// Check that Delegations on validators have NOT decreased
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal, val1.Delegation, "val1 delegation has NOT decreased")
	s.Require().Equal(initialState.val2Bal, val2.Delegation, "val2 delegation has NOT decreased")

	// Check that the number of delegation changes in progress was reset
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// Check that the host zone unbonding records have not been updated
	expectedStatus := recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS
	if status == icacallbacktypes.AckResponseStatus_FAILURE {
		expectedStatus = recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE
	}
	for _, epochNumber := range initialState.epochNumbers {
		hzu := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)
		s.Require().Equal(int64(0), int64(hzu.UnbondingTime), "completion time is NOT set on the hzu")
		s.Require().Equal(expectedStatus, hzu.Status, "hzu status was not changed")
		s.Require().Zero(hzu.UndelegationTxsInProgress, "hzu undelegations in progress")
	}

	// Confirm stTokens were NOT burned
	depositAccount := sdk.MustAccAddressFromBech32(hostZone.DepositAddress)
	depositBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAccount, StAtom).Amount
	s.Require().Equal(initialState.zoneAccountBalance, depositBalance, "tokens were not burned")
}

func (s *KeeperTestSuite) TestUndelegateCallback_AckTimeout() {
	tc := s.SetupUndelegateCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds on timeout")
	s.checkStateIfUndelegateCallbackFailed(tc, invalidArgs.ackResponse.Status)
}

func (s *KeeperTestSuite) TestUndelegateCallback_AckFailure() {
	tc := s.SetupUndelegateCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds with error on host")
	s.checkStateIfUndelegateCallbackFailed(tc, invalidArgs.ackResponse.Status)
}

func (s *KeeperTestSuite) TestUndelegateCallback_WrongCallbackArgs() {
	tc := s.SetupUndelegateCallback()

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal undelegate callback args")
}

func (s *KeeperTestSuite) TestUndelegateCallback_HostNotFound() {
	tc := s.SetupUndelegateCallback()

	// remove the host zone from the store to trigger a host not found error
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, "Host zone not found: GAIA: key not found")
}

func (s *KeeperTestSuite) TestMarkUndelegationAckReceived() {
	// Setup 3 validators, two of which will have their delegation changes decremented
	initialHostZone := types.HostZone{
		ChainId: HostChainId,
		Validators: []*types.Validator{
			{Address: "val1", DelegationChangesInProgress: 1},
			{Address: "val2", DelegationChangesInProgress: 2},
			{Address: "val3", DelegationChangesInProgress: 3},
		},
	}
	splitUndelegations := []*types.SplitUndelegation{
		{Validator: "val2"},
		{Validator: "val3"},
	}
	expectedFinalValidators := []*types.Validator{
		{Address: "val1", DelegationChangesInProgress: 1},
		{Address: "val2", DelegationChangesInProgress: 1}, // decremented
		{Address: "val3", DelegationChangesInProgress: 2}, // decremented
	}

	// Create three host zone unbonding records, two of which will have the counter decremented
	initialHostZoneUnbondings := []recordtypes.HostZoneUnbonding{
		{HostZoneId: HostChainId, UndelegationTxsInProgress: 1},
		{HostZoneId: HostChainId, UndelegationTxsInProgress: 2},
		{HostZoneId: HostChainId, UndelegationTxsInProgress: 3},
	}
	allEpochs := []uint64{0, 1, 2}
	ackedEpochs := []uint64{1, 2}
	for i, epochNumber := range allEpochs {
		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: epochNumber,
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
				&initialHostZoneUnbondings[i],
			},
		})
	}
	expectedFinalUnbondings := []recordtypes.HostZoneUnbonding{
		{UndelegationTxsInProgress: 1},
		{UndelegationTxsInProgress: 1}, // decremented
		{UndelegationTxsInProgress: 2}, // decremented
	}

	// Call MarkAckReceived
	callback := types.UndelegateCallback{
		SplitUndelegations:      splitUndelegations,
		EpochUnbondingRecordIds: ackedEpochs,
	}
	err := s.App.StakeibcKeeper.MarkUndelegationAckReceived(s.Ctx, initialHostZone, callback)
	s.Require().NoError(err)

	// Check validator counts against expectations
	hostZone := s.MustGetHostZone(HostChainId)
	actualValidators := hostZone.Validators
	for i, actualValidator := range actualValidators {
		expectedValidator := expectedFinalValidators[i]
		s.Require().Equal(expectedValidator.DelegationChangesInProgress,
			actualValidator.DelegationChangesInProgress, "validator delegation changs in progress")
	}

	// Check host zone unbonding counts against expectations
	for i, epochNumber := range allEpochs {
		actualHostZoneUnbonding := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)
		expectedHostZoneUnbonding := expectedFinalUnbondings[i]
		s.Require().Equal(expectedHostZoneUnbonding.UndelegationTxsInProgress,
			actualHostZoneUnbonding.UndelegationTxsInProgress, "hzu undelegation txs in progress")
	}
}

func (s *KeeperTestSuite) TestHandleFailedUndelegation() {
	// Create two HZU records
	// One should be in EXIT_TRANSFER_QUEUE (because it's already submitted the full undelegation)
	// And the other should be in status IN_PROGRESS
	initialStatuses := []recordtypes.HostZoneUnbonding_Status{
		recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
	}

	// After the failed undelegation, only the second record should be set to RETRY_QUEUE
	expectedStatuses := []recordtypes.HostZoneUnbonding_Status{
		recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE,
		recordtypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
	}

	// Create the initial records
	epochNumbers := []uint64{}
	for i, initialStatus := range initialStatuses {
		epochNumbers = append(epochNumbers, uint64(i))

		s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
			EpochNumber: uint64(i),
			HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{{
				HostZoneId: HostChainId,
				Status:     initialStatus,
			}},
		})
	}

	// Call HandleFailedUndelegation
	err := s.App.StakeibcKeeper.HandleFailedUndelegation(s.Ctx, HostChainId, epochNumbers)
	s.Require().NoError(err, "no error expected when handling undelegation")

	// Check that the status of the second record was set to RETRY_QUEUE
	for _, epochNumber := range epochNumbers {
		hostZoneUnbonding := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)
		expectedStatus := expectedStatuses[int(epochNumber)]
		s.Require().Equal(expectedStatus, hostZoneUnbonding.Status, "status after update")
	}
}

func (s *KeeperTestSuite) TestUpdateDelegationBalances() {
	tc := s.SetupUndelegateCallback()

	// Check that total delegation has NOT decreased on the host zone
	hostZone := s.MustGetHostZone(HostChainId)
	err := s.App.StakeibcKeeper.UpdateDelegationBalances(s.Ctx, hostZone, tc.initialState.callbackArgs)
	s.Require().NoError(err, "update delegation balances succeeds")

	// Check that Delegations on validators have decreased
	updatedHostZone := s.MustGetHostZone(HostChainId)
	val1 := updatedHostZone.Validators[0]
	val2 := updatedHostZone.Validators[1]
	s.Require().Equal(val1.Delegation, tc.initialState.val1Bal.Sub(tc.val1UndelegationAmount), "val1 delegation has decreased")
	s.Require().Equal(val2.Delegation, tc.initialState.val2Bal.Sub(tc.val2UndelegationAmount), "val2 delegation has decreased")
}

func (s *KeeperTestSuite) TestCalculateTotalUnbondedInBatch() {
	splitUndelegations := []*types.SplitUndelegation{
		{NativeTokenAmount: sdkmath.NewInt(10)},
		{NativeTokenAmount: sdkmath.NewInt(20)},
		{NativeTokenAmount: sdkmath.NewInt(30)},
	}
	expectedNativeAmount := sdkmath.NewInt(10 + 20 + 30)

	actualNativeAmount := s.App.StakeibcKeeper.CalculateTotalUnbondedInBatch(splitUndelegations)
	s.Require().Equal(expectedNativeAmount, actualNativeAmount, "native total")

	// Zero case
	actualNativeAmount = s.App.StakeibcKeeper.CalculateTotalUnbondedInBatch([]*types.SplitUndelegation{})
	s.Require().Zero(actualNativeAmount.Int64(), "native zero")
}

func (s *KeeperTestSuite) TestGetLatestUnbondingCompletionTime() {
	// Construct TxMsgData
	firstCompletionTime := time.Now().Add(time.Second * time.Duration(10))
	secondCompletionTime := time.Now().Add(time.Second * time.Duration(20))

	var err error
	msgResponses := make([][]byte, 2)
	msgResponses[0], err = proto.Marshal(&stakingtypes.MsgUndelegateResponse{CompletionTime: firstCompletionTime})
	s.Require().NoError(err, "marshal error")
	msgResponses[1], err = proto.Marshal(&stakingtypes.MsgUndelegateResponse{CompletionTime: secondCompletionTime})
	s.Require().NoError(err, "marshal error")

	// Check that the second completion time (the later of the two) is returned
	latestCompletionTime, err := s.App.StakeibcKeeper.GetLatestUnbondingCompletionTime(s.Ctx, msgResponses)
	s.Require().NoError(err, "get latest completion time succeeds")
	s.Require().Equal(uint64(secondCompletionTime.UnixNano()), latestCompletionTime, "latest completion time is the second completion time")

	// Calling latest completion time with random message responses will provoke an unmarshal failure
	msgResponses = [][]byte{{1}, {2}, {3}}
	_, err = s.App.StakeibcKeeper.GetLatestUnbondingCompletionTime(s.Ctx, msgResponses)
	s.Require().ErrorContains(err, "Unable to unmarshal undelegation tx response")

	// Calling latest completion time with an no msg responses will cause the completion time to be 0
	msgResponses = [][]byte{}
	_, err = s.App.StakeibcKeeper.GetLatestUnbondingCompletionTime(s.Ctx, msgResponses)
	s.Require().ErrorContains(err, "invalid packet completion time")
}

func (s *KeeperTestSuite) TestUpdateHostZoneUnbondingsAfterUndelegation() {
	// Abbreviated struct and statues for readability
	type HostZoneUnbonding struct {
		RecordNative     int64
		RecordStToken    int64
		RemainingNative  int64
		RemainingStToken int64
		UnbondTime       uint64
		Status           recordtypes.HostZoneUnbonding_Status
	}
	inProgress := recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS
	complete := recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE

	testCases := []struct {
		name                        string
		batchNativeUnbonded         sdkmath.Int
		expectedBatchStTokensBurned sdkmath.Int
		unbondingTimeFromResponse   uint64
		initialRecords              []HostZoneUnbonding
		finalRecords                []HostZoneUnbonding
	}{
		{
			// One Record, full unbonding
			// 1000 total native, 1000 total sttoken, implied RR of 1.0
			// Both remaining amounts decrement to 0,
			// Unbonding time is updated to 2, Status updates to EXIT_TRANSFER_QUEUE
			name:                        "one unbonding record full amount",
			batchNativeUnbonded:         sdkmath.NewInt(1000),
			expectedBatchStTokensBurned: sdkmath.NewInt(1000),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
			},
		},
		{
			// One Record, parital unbonding
			// 2000 total native, 1000 total sttoken, implied RR of 2.0
			// Batch 1000 native decremented from record, implies 500 sttokens burned
			// Unbonding time is updated to 2, Status doesn't change
			name:                        "one unbonding record parital amount",
			batchNativeUnbonded:         sdkmath.NewInt(1000),
			expectedBatchStTokensBurned: sdkmath.NewInt(500),
			unbondingTimeFromResponse:   1,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 2000, RecordStToken: 1000, UnbondTime: 2, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 1000, RemainingStToken: 500, UnbondTime: 2, Status: inProgress},
			},
		},
		{
			// Two records, parital unbonding on first
			// Record 1: 1000 total native, 1000 total st, implied RR of 1.0
			// Record 2: 2000 total native, 1000 total st, implied RR of 2.0
			// Record 1: Batch 400 native decremented, implies 400 sttokens burned
			// Record 2: Untouched
			// Unbonding time updated, Status doesn't change
			name:                        "two unbonding records partial on first",
			batchNativeUnbonded:         sdkmath.NewInt(400),
			expectedBatchStTokensBurned: sdkmath.NewInt(400),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 1, Status: inProgress},
				{RecordNative: 2000, RecordStToken: 1000, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 600, RemainingStToken: 600, UnbondTime: 2, Status: inProgress},
				{RemainingNative: 2000, RemainingStToken: 1000, UnbondTime: 2, Status: inProgress},
			},
		},
		{
			// Two records, full unbonding on first
			// Record 1: 1000 total native, 1000 total st, implied RR of 1.0
			// Record 2: 2000 total native, 1000 total st, implied RR of 2.0
			// Record 1: Batch 1000 native decremented, implies 1000 sttokens burned
			// Record 2: Untouched
			// Unbonding time not changed, Status changes on first record
			name:                        "two unbonding records partial on first",
			batchNativeUnbonded:         sdkmath.NewInt(1000),
			expectedBatchStTokensBurned: sdkmath.NewInt(1000),
			unbondingTimeFromResponse:   1,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 2, Status: inProgress},
				{RecordNative: 2000, RecordStToken: 1000, UnbondTime: 2, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 2000, RemainingStToken: 1000, UnbondTime: 2, Status: inProgress},
			},
		},
		{
			// Two records, full unbonding on first, parital on second
			// Record 1: 1000 total native, 1000 total st, implied RR of 1.0
			// Record 2: 2000 total native, 1000 total st, implied RR of 2.0
			// Total batch unbonded: 2200
			// Record 1: Batch 1000 native decremented, implies 1000 sttokens burned
			// Record 2: Batch 1200 native decremented, implies 600 sttokens burned
			// Unbonding time updated, Status changes on first record
			name:                        "two unbonding records partial on first",
			batchNativeUnbonded:         sdkmath.NewInt(1000 + 1200),
			expectedBatchStTokensBurned: sdkmath.NewInt(1000 + 600),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 1, Status: inProgress},
				{RecordNative: 2000, RecordStToken: 1000, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 800, RemainingStToken: 400, UnbondTime: 2, Status: inProgress},
			},
		},
		{
			// Two records, full unbonding on both
			// Record 1: 1000 total native, 1000 total st, implied RR of 1.0
			// Record 2: 2000 total native, 1000 total st, implied RR of 2.0
			// Total batch unbonded: 3000
			// Record 1: Batch 1000 native decremented, implies 1000 sttokens burned
			// Record 2: Batch 2000 native decremented, implies 1000 sttokens burned
			// Unbonding time not changed, Status changes on both records
			name:                        "two unbonding records partial on first",
			batchNativeUnbonded:         sdkmath.NewInt(1000 + 2000),
			expectedBatchStTokensBurned: sdkmath.NewInt(1000 + 1000),
			unbondingTimeFromResponse:   1,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 2, Status: inProgress},
				{RecordNative: 2000, RecordStToken: 1000, UnbondTime: 2, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
			},
		},
		{
			// Two records, partial starting point, batch finishes
			// Record 1: 1000 total native, 1000 total st, implied RR of 1.0
			// Record 2: 2000 total native, 1000 total st, implied RR of 2.0
			// Previously unbonded 800, now unbonding remaining 2200
			// Record 1: Decrements 200 down to 0, implies 200 sttokens burned
			// Record 2: Decrements 2000, implies 1000 sttokens burned
			// Unbonding time updated, Status changes on both records
			name:                        "two unbonding records partial on first",
			batchNativeUnbonded:         sdkmath.NewInt(200 + 2000),
			expectedBatchStTokensBurned: sdkmath.NewInt(200 + 1000),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, RemainingNative: 200, RemainingStToken: 200, UnbondTime: 1, Status: inProgress},
				{RecordNative: 2000, RecordStToken: 1000, RemainingNative: 2000, RemainingStToken: 1000, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
			},
		},
		{
			// Three records, full unbond on all, precision error
			// All records decremented to 0, All records status updated
			// Record 1: 1000 native, 1000 sttokens, implied RR of 1.0
			// Record 2: 1500 native, 1000 sttokens, implied RR of 1.5
			// Record 3: 2000 native, 1000 sttokens, implied RR of 2.0
			// Each record will start with an extra sttoken remaining, to test that it gets rounded down to 0
			// when the native amount goes to 0
			name:                        "three unbonding records full on all with precision error",
			batchNativeUnbonded:         sdkmath.NewInt(1000 + 1500 + 2000),
			expectedBatchStTokensBurned: sdkmath.NewInt(1001 + 1001 + 1001),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, RemainingStToken: 1001, UnbondTime: 1, Status: inProgress},
				{RecordNative: 1500, RecordStToken: 1000, RemainingStToken: 1001, UnbondTime: 1, Status: inProgress},
				{RecordNative: 2000, RecordStToken: 1000, RemainingStToken: 1001, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
			},
		},
		{
			// Four records records, full unbond on three, 1 remaining on last
			// Random sttoken values to similate random redemption rates
			// First three records decremented to 0, last record partially decremented
			// Status update on first three records
			name:                        "four unbonding records partial on last",
			batchNativeUnbonded:         sdkmath.NewInt(834 + 234 + 1093 + 2379 - 1),
			expectedBatchStTokensBurned: sdkmath.NewInt(923 + 389 + 654 + 2379 - 1),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 834, RecordStToken: 923, UnbondTime: 1, Status: inProgress},
				{RecordNative: 234, RecordStToken: 389, UnbondTime: 1, Status: inProgress},
				{RecordNative: 1093, RecordStToken: 654, UnbondTime: 1, Status: inProgress},
				{RecordNative: 2379, RecordStToken: 2379, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 2, Status: complete},
				{RemainingNative: 1, RemainingStToken: 1, UnbondTime: 2, Status: inProgress},
			},
		},
		{
			// Two records, first one already finished
			// Time should only update on the last record
			name:                        "first record completed in previous callback",
			batchNativeUnbonded:         sdkmath.NewInt(500),
			expectedBatchStTokensBurned: sdkmath.NewInt(500),
			unbondingTimeFromResponse:   2,
			initialRecords: []HostZoneUnbonding{
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 1, Status: complete},
				{RecordNative: 1000, RecordStToken: 1000, UnbondTime: 1, Status: inProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{RemainingNative: 0, RemainingStToken: 0, UnbondTime: 1, Status: complete},
				{RemainingNative: 500, RemainingStToken: 500, UnbondTime: 2, Status: inProgress},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Create host zone unbonding records
			epochUnbondingRecordIds := []uint64{}
			for i, hostZoneUnbondingTc := range tc.initialRecords {
				epochNumber := uint64(i)
				epochUnbondingRecordIds = append(epochUnbondingRecordIds, epochNumber)

				// For brevity, the remaining amount was excluded on certain records where
				// the remaining was equal to the full record amount
				remainingNative := hostZoneUnbondingTc.RemainingNative
				remainingStToken := hostZoneUnbondingTc.RemainingStToken
				if hostZoneUnbondingTc.RemainingNative == 0 && hostZoneUnbondingTc.Status == inProgress {
					remainingNative = hostZoneUnbondingTc.RecordNative
				}
				if hostZoneUnbondingTc.RemainingStToken == 0 && hostZoneUnbondingTc.Status == inProgress {
					remainingStToken = hostZoneUnbondingTc.RecordStToken
				}

				s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
					EpochNumber: epochNumber,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
						{
							HostZoneId:           HostChainId,
							NativeTokenAmount:    sdkmath.NewInt(hostZoneUnbondingTc.RecordNative),
							NativeTokensToUnbond: sdkmath.NewInt(remainingNative),
							StTokenAmount:        sdkmath.NewInt(hostZoneUnbondingTc.RecordStToken),
							StTokensToBurn:       sdkmath.NewInt(remainingStToken),
							UnbondingTime:        hostZoneUnbondingTc.UnbondTime,
							Status:               hostZoneUnbondingTc.Status,
						},
					},
				})
			}

			// Call the Update function
			actualStTokensBurned, err := s.App.StakeibcKeeper.UpdateHostZoneUnbondingsAfterUndelegation(
				s.Ctx,
				HostChainId,
				epochUnbondingRecordIds,
				tc.batchNativeUnbonded,
				tc.unbondingTimeFromResponse,
			)
			s.Require().NoError(err, "no error expected during update")
			s.Require().Equal(tc.expectedBatchStTokensBurned.Int64(), actualStTokensBurned.Int64(), "total sttokens burned")

			// Confirm the new host zone unbonding records match expectations
			for i, epochNumber := range epochUnbondingRecordIds {
				expectedHostZoneUnbonding := tc.finalRecords[i]
				actualHostZoneUnbonding := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)

				s.Require().Equal(expectedHostZoneUnbonding.Status, actualHostZoneUnbonding.Status,
					"status for record %d", i)
				s.Require().Equal(expectedHostZoneUnbonding.RemainingNative, actualHostZoneUnbonding.NativeTokensToUnbond.Int64(),
					"native tokens for record %d", i)
				s.Require().Equal(expectedHostZoneUnbonding.RemainingStToken, actualHostZoneUnbonding.StTokensToBurn.Int64(),
					"sttokens for record %d", i)
				s.Require().Equal(expectedHostZoneUnbonding.UnbondTime, actualHostZoneUnbonding.UnbondingTime,
					"unbonding time for record %d", i)
			}
		})
	}
}

func (s *KeeperTestSuite) TestBurnStTokensAfterUndelegation() {
	validDepositAccount := s.TestAccs[0]

	testCases := []struct {
		name                     string
		depositAccount           string
		initialBalance           sdkmath.Int
		burnAmount               sdkmath.Int
		expectedRemainingBalance sdkmath.Int
		expectedError            string
	}{
		{
			name:                     "successful partial burn",
			depositAccount:           validDepositAccount.String(),
			initialBalance:           sdkmath.NewInt(10_000),
			burnAmount:               sdkmath.NewInt(8_000),
			expectedRemainingBalance: sdkmath.NewInt(2_000),
		},
		{
			name:                     "successful full burn",
			depositAccount:           validDepositAccount.String(),
			initialBalance:           sdkmath.NewInt(10_000),
			burnAmount:               sdkmath.NewInt(10_000),
			expectedRemainingBalance: sdkmath.NewInt(0),
		},
		{
			name:           "invalid deposit account",
			depositAccount: "invalid-account",
			initialBalance: sdkmath.NewInt(10_000),
			burnAmount:     sdkmath.NewInt(10_000),
			expectedError:  "unable to convert deposit address",
		},
		{
			name:           "insufficient funds",
			depositAccount: validDepositAccount.String(),
			initialBalance: sdkmath.NewInt(10_000),
			burnAmount:     sdkmath.NewInt(10_001),
			expectedError:  "unable to send sttokens from deposit account for burning",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			s.SetupTest() // resets account balances and token supply

			hostZone := types.HostZone{
				HostDenom:      Atom,
				DepositAddress: tc.depositAccount,
			}

			stTokens := sdk.NewCoin(StAtom, tc.initialBalance)
			s.FundAccount(validDepositAccount, stTokens)

			actualError := s.App.StakeibcKeeper.BurnStTokensAfterUndelegation(s.Ctx, hostZone, tc.burnAmount)
			if tc.expectedError != "" {
				s.Require().ErrorContains(actualError, tc.expectedError)
			} else {
				s.Require().NoError(actualError, "no error expected when burning")

				finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, validDepositAccount, StAtom).Amount
				s.Require().Equal(tc.expectedRemainingBalance.Int64(), finalBalance.Int64(), "remaining balance")
			}
		})
	}
}
