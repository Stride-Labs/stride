package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type UndelegateCallbackState struct {
	stakedBal          sdk.Int
	val1Bal            sdk.Int
	val2Bal            sdk.Int
	epochNumber        uint64
	completionTime     time.Time
	callbackArgs       types.UndelegateCallback
	zoneAccountBalance sdk.Int
}

type UndelegateCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type UndelegateCallbackTestCase struct {
	initialState           UndelegateCallbackState
	validArgs              UndelegateCallbackArgs
	val1UndelegationAmount sdk.Int
	val2UndelegationAmount sdk.Int
	balanceToUnstake       sdk.Int
}

func (s *KeeperTestSuite) SetupUndelegateCallback() UndelegateCallbackTestCase {
	// Set up host zone and validator state
	stakedBal := sdk.NewInt(1_000_000)
	val1Bal := sdk.NewInt(400_000)
	val2Bal := stakedBal.Sub(val1Bal)
	balanceToUnstake := sdk.NewInt(300_000)
	val1UndelegationAmount := sdk.NewInt(120_000)
	val2UndelegationAmount := balanceToUnstake.Sub(val1UndelegationAmount)
	epochNumber := uint64(1)
	val1 := types.Validator{
		Name:          "val1",
		Address:       "val1_address",
		DelegationAmt: val1Bal,
	}
	val2 := types.Validator{
		Name:          "val2",
		Address:       "val2_address",
		DelegationAmt: val2Bal,
	}
	zoneAddress := types.NewZoneAddress(HostChainId)
	zoneAccountBalance := balanceToUnstake.Add(sdk.NewInt(10))
	zoneAccount := Account{
		acc:           zoneAddress,
		stAtomBalance: sdk.NewCoin(StAtom, zoneAccountBalance), // Add a few extra tokens to make the test more robust
	}
	hostZone := stakeibc.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Validators:     []*types.Validator{&val1, &val2},
		StakedBal:      stakedBal,
		Address:        zoneAddress.String(),
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

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	msgs = append(msgs, &stakingTypes.MsgUndelegate{}, &stakingTypes.MsgUndelegate{})
	completionTime := time.Now()
	msgUndelegateResponse := &stakingTypes.MsgUndelegateResponse{CompletionTime: completionTime}
	protoMsgUndelegateResponse := proto.Message(msgUndelegateResponse)
	ack := s.ICAPacketAcknowledgement(msgs, &protoMsgUndelegateResponse)
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
	args, err := s.App.StakeibcKeeper.MarshalUndelegateCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err, "callback args unmarshalled")

	return UndelegateCallbackTestCase{
		val1UndelegationAmount: val1UndelegationAmount,
		val2UndelegationAmount: val2UndelegationAmount,
		balanceToUnstake:       balanceToUnstake,
		initialState: UndelegateCallbackState{
			callbackArgs:       callbackArgs,
			stakedBal:          stakedBal,
			val1Bal:            val1Bal,
			val2Bal:            val2Bal,
			epochNumber:        epochNumber,
			completionTime:     completionTime,
			zoneAccountBalance: zoneAccountBalance,
		},
		validArgs: UndelegateCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestUndelegateCallback_Successful() {
	tc := s.SetupUndelegateCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	// Callback
	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx, validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds")

	// Check that stakedBal has decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(hostZone.StakedBal, initialState.stakedBal.Sub(tc.balanceToUnstake), "stakedBal has decreased on the host zone")

	// Check that Delegations on validators have decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	s.Require().Equal(val1.DelegationAmt, initialState.val1Bal.Sub(tc.val1UndelegationAmount), "val1 delegation has decreased")
	val2 := hostZone.Validators[1]
	// Check that the host zone unbonding records have been updated
	s.Require().Equal(val2.DelegationAmt, initialState.val2Bal.Sub(tc.val2UndelegationAmount), "val2 delegation has decreased")

	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1, "1 host zone unbonding found")
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), initialState.completionTime.UnixNano(), "completion time is set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, "hzu status is set to EXIT_TRANSFER_QUEUE")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	s.Require().NoError(err, "zone account address is valid")
	s.Require().Equal(tc.balanceToUnstake, initialState.zoneAccountBalance.Sub(s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount), "tokens are burned")
}

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackFailed(tc UndelegateCallbackTestCase) {
	initialState := tc.initialState

	// Check that stakedBal has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	s.Require().Equal(hostZone.StakedBal, initialState.stakedBal, "stakedBal has NOT decreased on the host zone")

	// Check that Delegations on validators have NOT decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	s.Require().Equal(val1.DelegationAmt, initialState.val1Bal, "val1 delegation has NOT decreased")
	val2 := hostZone.Validators[1]
	// Check that the host zone unbonding records have not been updated
	s.Require().Equal(val2.DelegationAmt, initialState.val2Bal, "val2 delegation has NOT decreased")

	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, initialState.epochNumber)
	s.Require().True(found, "epoch unbonding record found")
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1, "1 host zone unbonding found")
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), int64(0), "completion time is NOT set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, "hzu status is set to UNBONDING_QUEUE")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	s.Require().NoError(err, "zone account address is valid")
	s.Require().Equal(initialState.zoneAccountBalance, s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "tokens are NOT burned")
}

func (s *KeeperTestSuite) TestUndelegateCallback_UndelegateCallbackTimeout() {
	tc := s.SetupUndelegateCallback()
	invalidArgs := tc.validArgs
	// a nil ack means the request timed out
	invalidArgs.ack = nil
	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds on timeout")
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_UndelegateCallbackErrorOnHost() {
	tc := s.SetupUndelegateCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	invalidArgs.ack = &fullAck

	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err, "undelegate callback succeeds with error on host")
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_WrongCallbackArgs() {
	tc := s.SetupUndelegateCallback()
	invalidArgs := tc.validArgs

	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "Unable to unmarshal undelegate callback args | unexpected EOF: unable to unmarshal data structure")
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_HostNotFound() {
	tc := s.SetupUndelegateCallback()
	valid := tc.validArgs
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)
	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx, valid.packet, valid.ack, valid.args)
	s.Require().EqualError(err, "Host zone not found: GAIA: key not found")
}

// UpdateDelegationBalances tests
func (s *KeeperTestSuite) TestUpdateDelegationBalances_Success() {
	tc := s.SetupUndelegateCallback()
	// Check that stakedBal has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	err := s.App.StakeibcKeeper.UpdateDelegationBalances(s.Ctx, hostZone, tc.initialState.callbackArgs)
	s.Require().NoError(err, "update delegation balances succeeds")

	updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")

	// Check that Delegations on validators have decreased
	s.Require().True(len(updatedHostZone.Validators) == 2, "Expected 2 validators")
	val1 := updatedHostZone.Validators[0]
	s.Require().Equal(val1.DelegationAmt, tc.initialState.val1Bal.Sub(tc.val1UndelegationAmount), "val1 delegation has decreased")
	val2 := updatedHostZone.Validators[1]
	s.Require().Equal(val2.DelegationAmt, tc.initialState.val2Bal.Sub(tc.val2UndelegationAmount), "val2 delegation has decreased")
}

// GetLatestCompletionTime tests
func (s *KeeperTestSuite) TestGetLatestCompletionTime_Success() {
	_ = s.SetupUndelegateCallback()
	// Construct TxMsgData
	firstCompletionTime := time.Now().Add(time.Second * time.Duration(10))
	secondCompletionTime := time.Now().Add(time.Second * time.Duration(20))
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, 2),
	}
	data, err := proto.Marshal(&stakingTypes.MsgUndelegateResponse{CompletionTime: firstCompletionTime})
	s.Require().NoError(err, "marshal error")
	txMsgData.Data[0] = &sdk.MsgData{
		MsgType: sdk.MsgTypeURL(&stakingTypes.MsgUndelegate{}),
		Data:    data,
	}
	data, err = proto.Marshal(&stakingTypes.MsgUndelegateResponse{CompletionTime: secondCompletionTime})
	s.Require().NoError(err, "marshal error")
	txMsgData.Data[1] = &sdk.MsgData{
		MsgType: sdk.MsgTypeURL(&stakingTypes.MsgUndelegate{}),
		Data:    data,
	}
	// Check that the second completion time (the later of the two) is returned
	latestCompletionTime, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx, txMsgData)
	s.Require().NoError(err, "get latest completion time succeeds")
	s.Require().Equal(secondCompletionTime.Unix(), latestCompletionTime.Unix(), "latest completion time is the second completion time")
}

func (s *KeeperTestSuite) TestGetLatestCompletionTime_Failure() {
	_ = s.SetupUndelegateCallback()
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, 2),
	}
	_, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx, txMsgData)
	s.Require().EqualError(err, "msgResponseBytes or msgResponseBytes.Data is nil: TxMsgData invalid")

	txMsgData = &sdk.TxMsgData{}
	_, err = s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx, txMsgData)
	s.Require().EqualError(err, "invalid packet completion time")
}

// UpdateHostZoneUnbondings tests
func (s *KeeperTestSuite) TestUpdateHostZoneUnbondings_Success() {
	totalBalance := sdk.NewInt(1_500_000)
	stAmtHzu1 := sdk.NewInt(600_000)
	stAmtHzu2 := sdk.NewInt(700_000)
	stAmtHzu3 := sdk.NewInt(200_000)
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
	// Create two epoch unbonding records (status UNBONDING_QUEUE, completion time unset)
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding1, &hostZoneUnbonding2},
	}
	epochUnbondingRecord2 := recordtypes.EpochUnbondingRecord{
		EpochNumber:        2,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding3},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx, epochUnbondingRecord2)
	callbackArgs := types.UndelegateCallback{
		EpochUnbondingRecordIds: []uint64{1, 2},
	}
	completionTime := time.Now().Add(time.Second * time.Duration(10))
	burnAmount, err := s.App.StakeibcKeeper.UpdateHostZoneUnbondings(s.Ctx, completionTime, HostChainId, callbackArgs)
	s.Require().NoError(err)
	s.Require().Equal(stAmtHzu1.Add(stAmtHzu3), burnAmount, "burn amount is correct")

	// Verify that 2 hzus have status EXIT_TRANSFER_QUEUE, while the third has status UNBONDING_QUEUE
	// Verify that 2 hzus have completion time set, while the third has no completion time
	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, 1)
	s.Require().True(found)
	epochUnbondingRecord2, found = s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx, 2)
	s.Require().True(found)
	hzu1 := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, hzu1.Status, "hzu1 status is EXIT_TRANSFER_QUEUE")
	s.Require().Equal(completionTime.UnixNano(), int64(hzu1.UnbondingTime), "hzu1 completion time is set")

	hzu2 := epochUnbondingRecord.HostZoneUnbondings[1]
	s.Require().Equal(recordtypes.HostZoneUnbonding_UNBONDING_QUEUE, hzu2.Status, "hzu2 status is UNBONDING_QUEUE")
	s.Require().Equal(int64(0), int64(hzu2.UnbondingTime), "hzu2 completion time is NOT set")

	hzu3 := epochUnbondingRecord2.HostZoneUnbondings[0]
	s.Require().Equal(recordtypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE, hzu3.Status, "hzu3 status is EXIT_TRANSFER_QUEUE")
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

	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	s.Require().NoError(err, "zoneAccount is valid")
	s.Require().Equal(tc.initialState.zoneAccountBalance, s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "initial token balance is 300_010")

	burnAmt := sdk.NewInt(123456)
	err = s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, burnAmt)
	s.Require().NoError(err)

	s.Require().Equal(tc.initialState.zoneAccountBalance.Sub(burnAmt), s.App.BankKeeper.GetBalance(s.Ctx, zoneAccount, StAtom).Amount, "post burn amount is 176_554")
}

// Test failure case - could not parse coin
func (s *KeeperTestSuite) TestBurnTokens_CouldNotParseCoin() {
	s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	hostZone.HostDenom = ":"

	burnAmt := sdk.NewInt(123456)
	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, burnAmt)
	s.Require().EqualError(err, "could not parse burnCoin: 123456st:. err: invalid decimal coin expression: 123456st:: invalid coins")
}

// Test failure case - could not decode address
func (s *KeeperTestSuite) TestBurnTokens_CouldNotParseAddress() {
	s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	hostZone.Address = "invalid"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, sdk.NewInt(123456))
	s.Require().EqualError(err, "could not bech32 decode address invalid of zone with id: GAIA")
}

// Test failure case - could not send coins from account to module
func (s *KeeperTestSuite) TestBurnTokens_CouldNotSendCoinsFromAccountToModule() {
	s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone found")
	hostZone.HostDenom = "coinDNE"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx, hostZone, sdk.NewInt(123456))
	s.Require().EqualError(err, "could not send coins from account stride1755g4dkhpw73gz9h9nwhlcefc6sdf8kcmvcwrk4rxfrz8xpxxjms7savm8 to module stakeibc. err: 0stcoinDNE is smaller than 123456stcoinDNE: insufficient funds")
}
