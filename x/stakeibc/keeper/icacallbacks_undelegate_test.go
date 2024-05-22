package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	icacallbacktypes "github.com/Stride-Labs/stride/v22/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v22/x/records/types"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"
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
		HostZoneId:           HostChainId,
		Status:               recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
		NativeTokensToUnbond: totalUndelegated,
		StTokensToBurn:       totalUndelegated,
		UnbondingTime:        uint64(0),
	}
	hostZoneUnbonding2 := recordtypes.HostZoneUnbonding{
		HostZoneId:           HostChainId,
		Status:               recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
		NativeTokensToUnbond: totalUndelegated,
		StTokensToBurn:       totalUndelegated,
		UnbondingTime:        completionTimeFromPrevBatch,
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
		StTokenAmount:     val1UndelegationAmount,
	}
	val2SplitDelegation := types.SplitUndelegation{
		Validator:         validators[1].Address,
		NativeTokenAmount: val2UndelegationAmount,
		StTokenAmount:     val2UndelegationAmount,
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

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackFailed(tc UndelegateCallbackTestCase) {
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
	for _, epochNumber := range initialState.epochNumbers {
		hzu := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)
		s.Require().Equal(int64(0), int64(hzu.UnbondingTime), "completion time is NOT set on the hzu")
		s.Require().Equal(recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS, hzu.Status, "hzu status was not changed")
	}

	// Confirm stTokens were NOT burned
	depositAccount := sdk.MustAccAddressFromBech32(hostZone.DepositAddress)
	depositBalance := s.App.BankKeeper.GetBalance(s.Ctx, depositAccount, StAtom).Amount
	s.Require().Equal(initialState.zoneAccountBalance, depositBalance, "tokens were not burned")
}

func (s *KeeperTestSuite) TestUndelegateCallback_UndelegateCallbackTimeout() {
	tc := s.SetupUndelegateCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds on timeout")
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_UndelegateCallbackErrorOnHost() {
	tc := s.SetupUndelegateCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds with error on host")
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_WrongCallbackArgs() {
	tc := s.SetupUndelegateCallback()

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal undelegate callback args: unexpected EOF: unable to unmarshal data structure")
}

func (s *KeeperTestSuite) TestUndelegateCallback_HostNotFound() {
	tc := s.SetupUndelegateCallback()

	// remove the host zone from the store to trigger a host not found error
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	err := s.App.StakeibcKeeper.UndelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, "Host zone not found: GAIA: key not found")
}

func (s *KeeperTestSuite) TestMarkUndelegationAckReceived() {

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

func (s *KeeperTestSuite) TestCalculateTokensFromBatch() {

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
		Native        int64
		StToken       int64
		UnbondingTime uint64
		Status        recordtypes.HostZoneUnbonding_Status
	}
	statusInProgress := recordtypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS
	statusComplete := recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE

	testCases := []struct {
		name                      string
		totalNativeUnbonded       sdkmath.Int
		totalStBurned             sdkmath.Int
		unbondingTimeFromResponse uint64
		initialRecords            []HostZoneUnbonding
		finalRecords              []HostZoneUnbonding
	}{
		{
			// One Record, full unbonding, time updated
			// Both amounts decrement to 0, unbonding time is updated to 2
			// Status updates to EXIT_TRANSFER_QUEUE
			name:                      "one unbonding record full amount",
			totalNativeUnbonded:       sdkmath.NewInt(1000),
			totalStBurned:             sdkmath.NewInt(500),
			unbondingTimeFromResponse: 2,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
			},
		},
		{
			// One Record, parital unbonding, time not updated
			// Both amounts decrement paritally, unbonding time is updated to 2
			// Status doesn't change
			name:                      "one unbonding record parital amount",
			totalNativeUnbonded:       sdkmath.NewInt(500),
			totalStBurned:             sdkmath.NewInt(250),
			unbondingTimeFromResponse: 1,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 500, StToken: 250, UnbondingTime: 2, Status: statusInProgress},
			},
		},
		{
			// Two records, parital unbonding on first, time updated
			// First record decremented, second record untouched
			// Both records have time updated
			// Status doesn't change
			name:                      "two unbonding records partial on first",
			totalNativeUnbonded:       sdkmath.NewInt(500),
			totalStBurned:             sdkmath.NewInt(250),
			unbondingTimeFromResponse: 2,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 500, StToken: 250, UnbondingTime: 2, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
		},
		{
			// Two records, full unbond on first, time not updated
			// First record decremented to 0, second record untouched
			// First record status updated
			name:                      "two unbonding records full on first",
			totalNativeUnbonded:       sdkmath.NewInt(1000),
			totalStBurned:             sdkmath.NewInt(500),
			unbondingTimeFromResponse: 1,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
		},
		{
			// Two records, full unbond on first, partial on second, time updated
			// First record decremented to 0, second record decremented partially
			// First record status updated
			name:                      "two unbonding records full on first",
			totalNativeUnbonded:       sdkmath.NewInt(1500),
			totalStBurned:             sdkmath.NewInt(750),
			unbondingTimeFromResponse: 2,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 500, StToken: 250, UnbondingTime: 2, Status: statusInProgress},
			},
		},
		{
			// Two records, full unbond on both, time not updated
			// Both records decremented to 0
			// Both records status updated
			name:                      "two unbonding records full on both",
			totalNativeUnbonded:       sdkmath.NewInt(2000),
			totalStBurned:             sdkmath.NewInt(1000),
			unbondingTimeFromResponse: 1,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
			},
		},
		{
			// Three records, full unbond on all, time not updated
			// All records decremented to 0
			// ALl records status updated
			name:                      "three unbonding records full on all",
			totalNativeUnbonded:       sdkmath.NewInt(3000),
			totalStBurned:             sdkmath.NewInt(1500),
			unbondingTimeFromResponse: 1,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
			},
		},
		{
			// Four records records, full unbond on three, partial on last, time updated
			// First three records decremented to 0, last record partially decremented
			// Status update on first three records
			name:                      "four unbonding records partial on last",
			totalNativeUnbonded:       sdkmath.NewInt(3999),
			totalStBurned:             sdkmath.NewInt(1999),
			unbondingTimeFromResponse: 2,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
				{Native: 1000, StToken: 500, UnbondingTime: 1, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 0, StToken: 0, UnbondingTime: 2, Status: statusComplete},
				{Native: 1, StToken: 1, UnbondingTime: 2, Status: statusInProgress},
			},
		},
		{
			// One record, native decremented to 0, stTokens has remainder
			// This should not be possible, but tests that the record status doesn't change
			name:                      "native and sttoken mismatch",
			totalNativeUnbonded:       sdkmath.NewInt(1000),
			totalStBurned:             sdkmath.NewInt(499),
			unbondingTimeFromResponse: 2,
			initialRecords: []HostZoneUnbonding{
				{Native: 1000, StToken: 500, UnbondingTime: 2, Status: statusInProgress},
			},
			finalRecords: []HostZoneUnbonding{
				{Native: 0, StToken: 1, UnbondingTime: 2, Status: statusInProgress},
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

				s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, recordtypes.EpochUnbondingRecord{
					EpochNumber: epochNumber,
					HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{
						{
							HostZoneId:           HostChainId,
							NativeTokensToUnbond: sdkmath.NewInt(hostZoneUnbondingTc.Native),
							StTokensToBurn:       sdkmath.NewInt(hostZoneUnbondingTc.StToken),
							UnbondingTime:        hostZoneUnbondingTc.UnbondingTime,
							Status:               hostZoneUnbondingTc.Status,
						},
					},
				})
			}

			// Call the Update function
			err := s.App.StakeibcKeeper.UpdateHostZoneUnbondingsAfterUndelegation(
				s.Ctx,
				HostChainId,
				epochUnbondingRecordIds,
				tc.totalStBurned,
				tc.totalNativeUnbonded,
				tc.unbondingTimeFromResponse,
			)
			s.Require().NoError(err, "no error expected during update")

			// Confirm the new host zone unbonding records match expectations
			for i, epochNumber := range epochUnbondingRecordIds {
				expectedHostZoneUnbonding := tc.finalRecords[i]
				actualHostZoneUnbonding := s.MustGetHostZoneUnbonding(epochNumber, HostChainId)

				s.Require().Equal(expectedHostZoneUnbonding.Status, actualHostZoneUnbonding.Status,
					"status for record %d", i)
				s.Require().Equal(expectedHostZoneUnbonding.Native, actualHostZoneUnbonding.NativeTokensToUnbond.Int64(),
					"native tokens for record %d", i)
				s.Require().Equal(expectedHostZoneUnbonding.StToken, actualHostZoneUnbonding.StTokensToBurn.Int64(),
					"sttokens for record %d", i)
				s.Require().Equal(expectedHostZoneUnbonding.UnbondingTime, actualHostZoneUnbonding.UnbondingTime,
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
			expectedError:  "could not bech32 decode address",
		},
		{
			name:           "insufficient funds",
			depositAccount: validDepositAccount.String(),
			initialBalance: sdkmath.NewInt(10_000),
			burnAmount:     sdkmath.NewInt(10_001),
			expectedError:  "could not send coins from account",
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
