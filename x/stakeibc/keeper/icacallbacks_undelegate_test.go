package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	icacallbacktypes "github.com/Stride-Labs/stride/v15/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v15/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v15/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v15/x/stakeibc/types"
)

type UndelegateCallbackState struct {
	totalDelegations   sdkmath.Int
	val1Bal            sdkmath.Int
	val2Bal            sdkmath.Int
	epochNumber        uint64
	completionTime     time.Time
	callbackArgs       types.UndelegateCallback
	zoneAccountBalance sdkmath.Int
}

type UndelegateCallbackHostState struct {
	totalDelegations   sdkmath.Int
	val1Bal            sdkmath.Int
	val2Bal            sdkmath.Int
	epochNumber        uint64
	completionTime     time.Time
	callbackArgs       types.UndelegateHostCallback
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
	balanceToUnstake       sdkmath.Int
}

type UndelegateCallbackHostTestCase struct {
	initialState           UndelegateCallbackHostState
	validArgs              UndelegateCallbackArgs
	val1UndelegationAmount sdkmath.Int
	val2UndelegationAmount sdkmath.Int
	balanceToUnstake       sdkmath.Int
}

func (s *KeeperTestSuite) SetupUndelegateCallback() UndelegateCallbackTestCase {
	// Set up host zone and validator state
	totalDelegations := sdkmath.NewInt(1_000_000)
	val1Bal := sdkmath.NewInt(400_000)
	val2Bal := totalDelegations.Sub(val1Bal)
	balanceToUnstake := sdkmath.NewInt(300_000)
	val1UndelegationAmount := sdkmath.NewInt(120_000)
	val2UndelegationAmount := balanceToUnstake.Sub(val1UndelegationAmount)
	epochNumber := uint64(1)
	val1 := types.Validator{
		Name:                        "val1",
		Address:                     "val1_address",
		Delegation:                  val1Bal,
		DelegationChangesInProgress: 1,
	}
	val2 := types.Validator{
		Name:                        "val2",
		Address:                     "val2_address",
		Delegation:                  val2Bal,
		DelegationChangesInProgress: 1,
	}
	depositAddress := types.NewHostZoneDepositAddress(HostChainId)
	zoneAccountBalance := balanceToUnstake.Add(sdkmath.NewInt(10))
	zoneAccount := Account{
		acc:           depositAddress,
		stAtomBalance: sdk.NewCoin(StAtom, zoneAccountBalance), // Add a few extra tokens to make the test more robust
	}
	hostZone := types.HostZone{
		ChainId:          HostChainId,
		HostDenom:        Atom,
		IbcDenom:         IbcAtom,
		RedemptionRate:   sdk.NewDec(1.0),
		Validators:       []*types.Validator{&val1, &val2},
		TotalDelegations: totalDelegations,
		DepositAddress:   depositAddress.String(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Set up EpochUnbondingRecord, HostZoneUnbonding and token state
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:    HostChainId,
		Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		StTokenAmount: balanceToUnstake,
	}
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	// mint stTokens to the zone account, to be burned
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	// Mock ack response
	packet := channeltypes.Packet{}
	completionTime := time.Now()
	msgsUndelegateResponse := &stakingtypes.MsgUndelegateResponse{CompletionTime: completionTime}
	msgsUndelegateResponseBz, err := proto.Marshal(msgsUndelegateResponse)
	s.Require().NoError(err, "no error expected when marshalling undelegate response")

	ackResponse := icacallbacktypes.AcknowledgementResponse{
		Status:       icacallbacktypes.AckResponseStatus_SUCCESS,
		MsgResponses: [][]byte{msgsUndelegateResponseBz},
	}

	// Mock callback args
	val1SplitDelegation := types.SplitDelegation{
		Validator: val1.Address,
		Amount:    val1UndelegationAmount,
	}
	val2SplitDelegation := types.SplitDelegation{
		Validator: val2.Address,
		Amount:    val2UndelegationAmount,
	}
	callbackArgs := types.UndelegateCallback{
		HostZoneId:              HostChainId,
		SplitDelegations:        []*types.SplitDelegation{&val1SplitDelegation, &val2SplitDelegation},
		EpochUnbondingRecordIds: []uint64{epochNumber},
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	s.Require().NoError(err, "callback args unmarshalled")

	return UndelegateCallbackTestCase{
		val1UndelegationAmount: val1UndelegationAmount,
		val2UndelegationAmount: val2UndelegationAmount,
		balanceToUnstake:       balanceToUnstake,
		initialState: UndelegateCallbackState{
			callbackArgs:       callbackArgs,
			totalDelegations:   totalDelegations,
			val1Bal:            val1Bal,
			val2Bal:            val2Bal,
			epochNumber:        epochNumber,
			completionTime:     completionTime,
			zoneAccountBalance: zoneAccountBalance,
		},
		validArgs: UndelegateCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        callbackArgsBz,
		},
	}
}

func (s *KeeperTestSuite) SetupUndelegateHostCallback() UndelegateCallbackHostTestCase {
	// Set up host zone and validator state
	totalDelegations := sdkmath.NewInt(1_000_000)
	val1Bal := sdkmath.NewInt(400_000)
	val2Bal := totalDelegations.Sub(val1Bal)
	balanceToUnstake := sdkmath.NewInt(300_000)
	val1UndelegationAmount := sdkmath.NewInt(120_000)
	val2UndelegationAmount := balanceToUnstake.Sub(val1UndelegationAmount)
	epochNumber := uint64(1)
	val1 := types.Validator{
		Name:                        "val1",
		Address:                     "val1_address",
		Delegation:                  val1Bal,
		DelegationChangesInProgress: 1,
	}
	val2 := types.Validator{
		Name:                        "val2",
		Address:                     "val2_address",
		Delegation:                  val2Bal,
		DelegationChangesInProgress: 1,
	}
	depositAddress := types.NewHostZoneDepositAddress(stakeibckeeper.EvmosHostZoneChainId)
	zoneAccountBalance := balanceToUnstake.Add(sdkmath.NewInt(10))
	zoneAccount := Account{
		acc:           depositAddress,
		stAtomBalance: sdk.NewCoin(StAtom, zoneAccountBalance), // Add a few extra tokens to make the test more robust
	}
	hostZone := types.HostZone{
		ChainId:          stakeibckeeper.EvmosHostZoneChainId,
		HostDenom:        Atom,
		IbcDenom:         IbcAtom,
		RedemptionRate:   sdk.NewDec(1.0),
		Validators:       []*types.Validator{&val1, &val2},
		TotalDelegations: totalDelegations,
		DepositAddress:   depositAddress.String(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Set up EpochUnbondingRecord, HostZoneUnbonding and token state
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:    stakeibckeeper.EvmosHostZoneChainId,
		Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		StTokenAmount: balanceToUnstake,
	}
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)

	// mint stTokens to the zone account, to be burned
	s.FundAccount(zoneAccount.acc, zoneAccount.stAtomBalance)

	// Mock ack response
	packet := channeltypes.Packet{}
	completionTime := time.Now()
	msgsUndelegateResponse := &stakingtypes.MsgUndelegateResponse{CompletionTime: completionTime}
	msgsUndelegateResponseBz, err := proto.Marshal(msgsUndelegateResponse)
	s.Require().NoError(err, "no error expected when marshalling undelegate response")

	ackResponse := icacallbacktypes.AcknowledgementResponse{
		Status:       icacallbacktypes.AckResponseStatus_SUCCESS,
		MsgResponses: [][]byte{msgsUndelegateResponseBz},
	}

	// Mock callback args
	val1SplitDelegation := types.SplitDelegation{
		Validator: val1.Address,
		Amount:    val1UndelegationAmount,
	}
	val2SplitDelegation := types.SplitDelegation{
		Validator: val2.Address,
		Amount:    val2UndelegationAmount,
	}
	callbackArgs := types.UndelegateHostCallback{
		Amt:              balanceToUnstake,
		SplitDelegations: []*types.SplitDelegation{&val1SplitDelegation, &val2SplitDelegation},
	}
	callbackArgsBz, err := proto.Marshal(&callbackArgs)
	s.Require().NoError(err, "callback args unmarshalled")

	return UndelegateCallbackHostTestCase{
		val1UndelegationAmount: val1UndelegationAmount,
		val2UndelegationAmount: val2UndelegationAmount,
		balanceToUnstake:       balanceToUnstake,
		initialState: UndelegateCallbackHostState{
			callbackArgs:       callbackArgs,
			totalDelegations:   totalDelegations,
			val1Bal:            val1Bal,
			val2Bal:            val2Bal,
			epochNumber:        epochNumber,
			completionTime:     completionTime,
			zoneAccountBalance: zoneAccountBalance,
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
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(hostZone.TotalDelegations, initialState.totalDelegations.Sub(tc.balanceToUnstake), "total delegation has decreased on the host zone")

	// Check that Delegations on validators have decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal.Sub(tc.val1UndelegationAmount), val1.Delegation, "val1 delegation has decreased")
	s.Require().Equal(initialState.val2Bal.Sub(tc.val2UndelegationAmount), val2.Delegation, "val2 delegation has decreased")

	// Check that the number of delegation changes in progress was reset to 0
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// Check that the host zone unbonding records have been updated
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1, "1 host zone unbonding found")
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), initialState.completionTime.UnixNano(), "completion time is set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, "hzu status is set to EXIT_TRANSFER_QUEUE")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	s.Require().NoError(err, "zone account address is valid")
	s.Require().Equal(tc.balanceToUnstake, initialState.zoneAccountBalance.Sub(s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount), "tokens are burned")
}

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackFailed(tc UndelegateCallbackTestCase) {
	initialState := tc.initialState

	// Check that total delegation has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(initialState.totalDelegations, hostZone.TotalDelegations, "total delegation has NOT decreased on the host zone")

	// Check that Delegations on validators have NOT decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal, val1.Delegation, "val1 delegation has NOT decreased")
	s.Require().Equal(initialState.val2Bal, val2.Delegation, "val2 delegation has NOT decreased")

	// Check that the number of delegation changes in progress was reset
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// Check that the host zone unbonding records have not been updated
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1, "1 host zone unbonding found")
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), int64(0), "completion time is NOT set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, "hzu status is set to UNBONDING_QUEUE")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	s.Require().NoError(err, "zone account address is valid")
	s.Require().Equal(initialState.zoneAccountBalance, s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "tokens are NOT burned")
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

// UpdateDelegationBalances tests
func (s *KeeperTestSuite) TestUpdateDelegationBalances_Success() {
	tc := s.SetupUndelegateCallback()
	// Check that total delegation has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	err := s.App.StakeibcKeeper.UpdateDelegationBalances(s.Ctx, hostZone, tc.initialState.callbackArgs)
	s.Require().NoError(err, "update delegation balances succeeds")

	updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	// Check that Delegations on validators have decreased
	s.Require().True(len(updatedHostZone.Validators) == 2, "Expected 2 validators")
	val1 := updatedHostZone.Validators[0]
	s.Require().Equal(val1.Delegation, tc.initialState.val1Bal.Sub(tc.val1UndelegationAmount), "val1 delegation has decreased")
	val2 := updatedHostZone.Validators[1]
	s.Require().Equal(val2.Delegation, tc.initialState.val2Bal.Sub(tc.val2UndelegationAmount), "val2 delegation has decreased")
}

// GetLatestCompletionTime tests
func (s *KeeperTestSuite) TestGetLatestCompletionTime_Success() {
	s.SetupUndelegateCallback()

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
	latestCompletionTime, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx, msgResponses)
	s.Require().NoError(err, "get latest completion time succeeds")
	s.Require().Equal(secondCompletionTime.Unix(), latestCompletionTime.Unix(), "latest completion time is the second completion time")
}

func (s *KeeperTestSuite) TestGetLatestCompletionTime_UnmarshalFailure() {
	s.SetupUndelegateCallback()

	// Calling latest completion time with random message responses will provoke an unmarshal failure
	msgResponses := [][]byte{{1}, {2}, {3}}
	_, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx, msgResponses)
	s.Require().ErrorContains(err, "Unable to unmarshal undelegation tx response")
}

func (s *KeeperTestSuite) TestGetLatestCompletionTime_Failure() {
	s.SetupUndelegateCallback()

	// Calling latest completion time with an no msg responses will cause the completion time to be 0
	msgResponses := [][]byte{}
	_, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx, msgResponses)
	s.Require().ErrorContains(err, "invalid packet completion time")
}

// UpdateHostZoneUnbondings tests
func (s *KeeperTestSuite) TestUpdateHostZoneUnbondings_Success() {
	totalBalance := sdkmath.NewInt(1_500_000)
	stAmtHzu1 := sdkmath.NewInt(600_000)
	stAmtHzu2 := sdkmath.NewInt(700_000)
	stAmtHzu3 := sdkmath.NewInt(200_000)
	s.Require().Equal(totalBalance, stAmtHzu1.Add(stAmtHzu2).Add(stAmtHzu3), "total balance is correct")

	// Set up EpochUnbondingRecord, HostZoneUnbonding and token state
	hostZoneUnbonding1 := recordtypes.HostZoneUnbonding{
		HostZoneId:    HostChainId,
		Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		StTokenAmount: stAmtHzu1,
	}
	hostZoneUnbonding2 := recordtypes.HostZoneUnbonding{
		HostZoneId:    "not_gaia",
		Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		StTokenAmount: stAmtHzu2,
	}
	hostZoneUnbonding3 := recordtypes.HostZoneUnbonding{
		HostZoneId:    HostChainId,
		Status:        recordtypes.HostZoneUnbonding_UNBONDING_QUEUE,
		StTokenAmount: stAmtHzu3,
	}

	// Create two epoch unbonding records (status UNBONDING_QUEUE, completion time originally unset)
	epochUnbondingRecord1 := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding1, &hostZoneUnbonding2},
	}
	epochUnbondingRecord2 := recordtypes.EpochUnbondingRecord{
		EpochNumber:        2,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding3},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord1)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord2)
	callbackArgs := types.UndelegateCallback{
		EpochUnbondingRecordIds: []uint64{1, 2},
	}

	// Update host zone unbonding record status and calculate how many stTokens to burn
	completionTime := time.Now().Add(time.Second * time.Duration(10))
	burnAmount, err := s.App.StakeibcKeeper.UpdateHostZoneUnbondings(s.Ctx, completionTime, HostChainId, callbackArgs)
	s.Require().NoError(err)
	s.Require().Equal(stAmtHzu1.Add(stAmtHzu3), burnAmount, "burn amount is correct")

	// Verify that 2 hzus have status EXIT_TRANSFER_QUEUE, while the third has status UNBONDING_QUEUE
	// Verify that 2 hzus have completion time set, while the third has no completion time
	epochUnbondingRecord1, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, 1)
	s.Require().True(found)
	epochUnbondingRecord2, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, 2)
	s.Require().True(found)

	hzu1 := epochUnbondingRecord1.HostZoneUnbondings[0]
	s.Require().Equal(completionTime.UnixNano(), int64(hzu1.UnbondingTime), "hzu1 completion time is set")

	hzu2 := epochUnbondingRecord1.HostZoneUnbondings[1]
	s.Require().Equal(recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, hzu2.Status, "hzu2 status is UNBONDING_QUEUE")
	s.Require().Equal(int64(0), int64(hzu2.UnbondingTime), "hzu2 completion time is NOT set")

	hzu3 := epochUnbondingRecord2.HostZoneUnbondings[0]
	s.Require().Equal(completionTime.UnixNano(), int64(hzu3.UnbondingTime), "hzu3 completion time is set")
}

// Test failure case - epoch unbonding record DNE
func (s *KeeperTestSuite) TestUpdateHostZoneUnbondings_EpochUnbondingRecordDNE() {
	callbackArgs := types.UndelegateCallback{
		EpochUnbondingRecordIds: []uint64{1},
	}
	completionTime := s.Ctx.BlockTime()
	_, err := s.App.StakeibcKeeper.UpdateHostZoneUnbondings(s.Ctx, completionTime, HostChainId, callbackArgs)
	s.Require().EqualError(err, "Unable to find epoch unbonding record for epoch: 1: key not found")
}

// Test failure case - HostZoneUnbonding DNE
func (s *KeeperTestSuite) TestUpdateHostZoneUnbondings_HostZoneUnbondingDNE() {
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber: 1,
		// No hzu!
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	callbackArgs := types.UndelegateCallback{
		EpochUnbondingRecordIds: []uint64{1},
	}
	completionTime := s.Ctx.BlockTime()
	_, err := s.App.StakeibcKeeper.UpdateHostZoneUnbondings(s.Ctx, completionTime, HostChainId, callbackArgs)
	s.Require().EqualError(err, "Host zone unbonding not found (GAIA) in epoch unbonding record: 1: key not found")
}

// BurnTokens Tests
func (s *KeeperTestSuite) TestBurnTokens_Success() {
	tc := s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	s.Require().NoError(err, "zoneAccount is valid")
	s.Require().Equal(tc.initialState.zoneAccountBalance, s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "initial token balance is 300_010")

	burnAmt := sdkmath.NewInt(123456)
	err = s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, burnAmt)
	s.Require().NoError(err)

	s.Require().Equal(tc.initialState.zoneAccountBalance.Sub(burnAmt), s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "post burn amount is 176_554")
}

// Test failure case - could not parse coin
func (s *KeeperTestSuite) TestBurnTokens_CouldNotParseCoin() {
	s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	hostZone.HostDenom = ","

	burnAmt := sdkmath.NewInt(123456)
	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, burnAmt)
	s.Require().EqualError(err, "could not parse burnCoin: 123456st,. err: invalid decimal coin expression: 123456st,: invalid coins")
}

// Test failure case - could not decode address
func (s *KeeperTestSuite) TestBurnTokens_CouldNotParseAddress() {
	s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	hostZone.DepositAddress = "invalid"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, sdkmath.NewInt(123456))
	s.Require().EqualError(err, "could not bech32 decode address invalid of zone with id: GAIA")
}

// Test failure case - could not send coins from account to module
func (s *KeeperTestSuite) TestBurnTokens_CouldNotSendCoinsFromAccountToModule() {
	s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	hostZone.HostDenom = "coinDNE"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, sdkmath.NewInt(123456))
	s.Require().EqualError(err, "could not send coins from account stride1755g4dkhpw73gz9h9nwhlcefc6sdf8kcmvcwrk4rxfrz8xpxxjms7savm8 to module stakeibc. err: spendable balance  is smaller than 123456stcoinDNE: insufficient funds")
}

func (s *KeeperTestSuite) TestUndelegateCallbackHost_Successful() {
	tc := s.SetupUndelegateHostCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	// Ensure IsUndelegateHostPrevented(ctx) is not yet flipped
	s.Require().False(s.App.StakeibcKeeper.IsUndelegateHostPrevented(s.Ctx))

	// Callback
	err := s.App.StakeibcKeeper.UndelegateHostCallback(s.Ctx, validArgs.packet, validArgs.ackResponse, validArgs.args)
	s.Require().NoError(err, "undelegate host callback succeeds")

	// Check that total delegation has decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, stakeibckeeper.EvmosHostZoneChainId)
	s.Require().True(found)
	s.Require().Equal(hostZone.TotalDelegations, initialState.totalDelegations.Sub(tc.balanceToUnstake), "total delegation has decreased on the host zone")

	// Check that Delegations on validators have decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal.Sub(tc.val1UndelegationAmount), val1.Delegation, "val1 delegation has decreased")
	s.Require().Equal(initialState.val2Bal.Sub(tc.val2UndelegationAmount), val2.Delegation, "val2 delegation has decreased")

	// Check that the number of delegation changes in progress was reset to 0
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// ensure UndelegateHostPrevented has been flipped to true
	s.Require().True(s.App.StakeibcKeeper.IsUndelegateHostPrevented(s.Ctx))
}

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackHostFailed(tc UndelegateCallbackHostTestCase) {
	initialState := tc.initialState

	// Check that total delegation has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, stakeibckeeper.EvmosHostZoneChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(initialState.totalDelegations, hostZone.TotalDelegations, "total delegation has NOT decreased on the host zone")

	// Check that Delegations on validators have NOT decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal, val1.Delegation, "val1 delegation has NOT decreased")
	s.Require().Equal(initialState.val2Bal, val2.Delegation, "val2 delegation has NOT decreased")

	// Check that the number of delegation changes in progress was reset
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// Check that the host zone unbonding records have not been updated
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1, "1 host zone unbonding found")
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), int64(0), "completion time is NOT set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, "hzu status is set to UNBONDING_QUEUE")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	s.Require().NoError(err, "zone account address is valid")
	s.Require().Equal(initialState.zoneAccountBalance, s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "tokens are NOT burned")
}

func (s *KeeperTestSuite) TestUndelegateCallbackHost_UndelegateCallbackTimeout() {
	tc := s.SetupUndelegateHostCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := s.App.StakeibcKeeper.UndelegateHostCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds on timeout")
	s.checkStateIfUndelegateCallbackHostFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallbackHost_UndelegateCallbackErrorOnHost() {
	tc := s.SetupUndelegateHostCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := s.App.StakeibcKeeper.UndelegateHostCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds with error on host")
	s.checkStateIfUndelegateCallbackHostFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallbackHost_WrongCallbackArgs() {
	tc := s.SetupUndelegateHostCallback()

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := s.App.StakeibcKeeper.UndelegateHostCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal undelegate host callback args: unexpected EOF: unable to unmarshal data structure")
}

func (s *KeeperTestSuite) TestUndelegateCallbackHost_HostNotFound() {
	tc := s.SetupUndelegateHostCallback()

	// remove the host zone from the store to trigger a host not found error
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, stakeibckeeper.EvmosHostZoneChainId)

	err := s.App.StakeibcKeeper.UndelegateHostCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, "Host zone not found: evmos_9001-2: key not found")
}
