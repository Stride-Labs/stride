package keeper_test

import (
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	"github.com/gogo/protobuf/proto"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type UndelegateCallbackState struct {
	stakedBal        uint64
	balanceToUnstake int64
	val1Bal          uint64
	val2Bal          uint64
	val1RelAmt       int64
	val2RelAmt       int64
	epochNumber      uint64
	completionTime   time.Time
	callbackArgs     types.UndelegateCallback
}

type UndelegateCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type UndelegateCallbackTestCase struct {
	initialState UndelegateCallbackState
	validArgs    UndelegateCallbackArgs
}

func (s *KeeperTestSuite) SetupUndelegateCallback() UndelegateCallbackTestCase {
	// Set up host zone and validator state
	stakedBal := uint64(1_000_000)
	val1Bal := uint64(400_000)
	val2Bal := uint64(stakedBal) - val1Bal
	balanceToUnstake := int64(300_000)
	val1RelAmt := int64(120_000)
	val2RelAmt := balanceToUnstake - val1RelAmt
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
	zoneAddress := types.NewZoneAddress(chainId)
	zoneAccount := Account{
		acc:           zoneAddress,
		stAtomBalance: sdk.NewInt64Coin(stAtom, balanceToUnstake+10), // Add a few extra tokens to make the test more robust
	}
	hostZone := stakeibc.HostZone{
		ChainId:        chainId,
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Validators:     []*types.Validator{&val1, &val2},
		StakedBal:      stakedBal,
		Address:        zoneAddress.String(),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)

	// Set up EpochUnbondingRecord, HostZoneUnbonding and token state
	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		HostZoneId:    chainId,
		Status:        recordtypes.HostZoneUnbonding_BONDED,
		StTokenAmount: uint64(balanceToUnstake),
	}
	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        epochNumber,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding},
	}
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
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
		Amount:    uint64(val1RelAmt),
	}
	val2SplitDelegation := types.SplitDelegation{
		Validator: val2.Address,
		Amount:    uint64(val2RelAmt),
	}
	callbackArgs := types.UndelegateCallback{
		HostZoneId:              chainId,
		SplitDelegations:        []*types.SplitDelegation{&val1SplitDelegation, &val2SplitDelegation},
		EpochUnbondingRecordIds: []uint64{epochNumber},
	}
	args, err := s.App.StakeibcKeeper.MarshalUndelegateCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)

	return UndelegateCallbackTestCase{
		initialState: UndelegateCallbackState{
			callbackArgs:     callbackArgs,
			stakedBal:        stakedBal,
			balanceToUnstake: balanceToUnstake,
			val1Bal:          val1Bal,
			val2Bal:          val2Bal,
			val1RelAmt:       val1RelAmt,
			val2RelAmt:       val2RelAmt,
			epochNumber:      epochNumber,
			completionTime:   completionTime,
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
	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// Check that stakedBal has decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	s.Require().Equal(int64(hostZone.StakedBal), int64(initialState.stakedBal)-initialState.balanceToUnstake)

	// Check that Delegations on validators have decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	s.Require().Equal(int64(val1.DelegationAmt), int64(initialState.val1Bal)-initialState.val1RelAmt)
	val2 := hostZone.Validators[1]
	s.Require().Equal(int64(val2.DelegationAmt), int64(initialState.val2Bal)-initialState.val2RelAmt)

	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), initialState.epochNumber)
	s.Require().True(found)
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1)
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), initialState.completionTime.UnixNano(), "completion time is set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_UNBONDED, "hzu status is set to UNBONDED")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	s.Require().NoError(err)
	s.Require().Equal(s.App.BankKeeper.GetBalance(s.Ctx(), zoneAccount, stAtom).Amount.Int64(), int64(10), "tokens are burned")
}

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackFailed(tc UndelegateCallbackTestCase) {
	initialState := tc.initialState

	// Check that stakedBal has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	s.Require().Equal(int64(hostZone.StakedBal), int64(initialState.stakedBal))

	// Check that Delegations on validators have NOT decreased
	s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	val1 := hostZone.Validators[0]
	s.Require().Equal(int64(val1.DelegationAmt), int64(initialState.val1Bal))
	val2 := hostZone.Validators[1]
	s.Require().Equal(int64(val2.DelegationAmt), int64(initialState.val2Bal))

	epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), initialState.epochNumber)
	s.Require().True(found)
	s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1)
	hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	s.Require().Equal(int64(hzu.UnbondingTime), int64(0), "completion time is NOT set on the hzu")
	s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_BONDED, "hzu status is set to BONDED")
	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	s.Require().NoError(err)
	s.Require().Equal(s.App.BankKeeper.GetBalance(s.Ctx(), zoneAccount, stAtom).Amount.Int64(), initialState.balanceToUnstake+int64(10), "tokens are NOT burned")
}

func (s *KeeperTestSuite) TestUndelegateCallback_UndelegateCallbackTimeout() {
	tc := s.SetupUndelegateCallback()
	invalidArgs := tc.validArgs
	// a nil ack means the request timed out
	invalidArgs.ack = nil
	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_UndelegateCallbackErrorOnHost() {
	tc := s.SetupUndelegateCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	invalidArgs.ack = &fullAck

	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_WrongCallbackArgs() {
	tc := s.SetupUndelegateCallback()
	invalidArgs := tc.validArgs

	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx(), invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "Unable to unmarshal undelegate callback args | unexpected EOF: unable to unmarshal data structure")
	s.checkStateIfUndelegateCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestUndelegateCallback_HostNotFound() {
	tc := s.SetupUndelegateCallback()
	valid := tc.validArgs
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx(), chainId)
	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx(), valid.packet, valid.ack, valid.args)
	s.Require().EqualError(err, "Host zone not found: GAIA: key not found")
}

// UpdateDelegationBalances tests
func (s *KeeperTestSuite) TestUpdateDelegationBalances_Success() {
	tc := s.SetupUndelegateCallback()
	// Check that stakedBal has NOT decreased on the host zone
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	err := s.App.StakeibcKeeper.UpdateDelegationBalances(s.Ctx(), hostZone, tc.initialState.callbackArgs)
	s.Require().NoError(err)

	updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)

	// Check that Delegations on validators have decreased
	s.Require().True(len(updatedHostZone.Validators) == 2, "Expected 2 validators")
	val1 := updatedHostZone.Validators[0]
	s.Require().Equal(int64(val1.DelegationAmt), int64(tc.initialState.val1Bal)-tc.initialState.val1RelAmt)
	val2 := updatedHostZone.Validators[1]
	s.Require().Equal(int64(val2.DelegationAmt), int64(tc.initialState.val2Bal)-tc.initialState.val2RelAmt)
}

func (s *KeeperTestSuite) TestUpdateDelegationBalances_BigDelegation() {
	_ = s.SetupUndelegateCallback()
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	splitDelegation := types.SplitDelegation{
		Amount: math.MaxUint64,
	}
	invalidCallbackArgs := types.UndelegateCallback{
		HostZoneId:              chainId,
		SplitDelegations:        []*types.SplitDelegation{&splitDelegation},
		EpochUnbondingRecordIds: []uint64{},
	}
	err := s.App.StakeibcKeeper.UpdateDelegationBalances(s.Ctx(), hostZone, invalidCallbackArgs)
	s.Require().EqualError(err, "Could not convert undelegate amount to int64 in undelegation callback | overflow: unable to cast 18446744073709551615 of type uint64 to int64: unable to cast to safe cast int")
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
	latestCompletionTime, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx(), txMsgData)
	s.Require().NoError(err)
	s.Require().Equal(secondCompletionTime.Unix(), latestCompletionTime.Unix())
}

func (s *KeeperTestSuite) TestGetLatestCompletionTime_Failure() {
	_ = s.SetupUndelegateCallback()
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, 2),
	}
	// Check that the second completion time (the later of the two) is returned
	_, err := s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx(), txMsgData)
	s.Require().EqualError(err, "msgResponseBytes or msgResponseBytes.Data is nil: TxMsgData invalid")

	txMsgData = &sdk.TxMsgData{}
	// Check that the second completion time (the later of the two) is returned
	_, err = s.App.StakeibcKeeper.GetLatestCompletionTime(s.Ctx(), txMsgData)
	s.Require().EqualError(err, "invalid packet completion time")
}

// UpdateHostZoneUnbondings tests
// WIP
// func (s *KeeperTestSuite) TestUpdateHostZoneUnbondings_Success() {
// 	tc := s.SetupUndelegateCallback()
// 	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
// 	s.Require().True(found)

// 	totalBalance := 1_500_000
// 	hzu1 := 600_000
// 	hzu2 := 700_000
// 	hzu3 := 200_000
// 	// Set up EpochUnbondingRecord, HostZoneUnbonding and token state
// 	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
// 		HostZoneId:    chainId,
// 		Status:        recordtypes.HostZoneUnbonding_BONDED,
// 		StTokenAmount: uint64(hzu1),
// 	}
// 	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
// 		HostZoneId:    chainId,
// 		Status:        recordtypes.HostZoneUnbonding_BONDED,
// 		StTokenAmount: uint64(hzu2),
// 	}
// 	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
// 		HostZoneId:    chainId,
// 		Status:        recordtypes.HostZoneUnbonding_BONDED,
// 		StTokenAmount: uint64(hzu3),
// 	}
// 	// Create two epoch unbonding records (status BONDED, completion time unset)
// 	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
// 		EpochNumber:        epochNumber,
// 		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding},
// 	}
// 	// Save them

// 	err := s.App.StakeibcKeeper.UpdateHostZoneUnbondings(s.Ctx(), hostZone, tc.initialState.callbackArgs)
// 	s.Require().NoError(err)

// 	updatedHostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
// 	s.Require().True(found)

// 	// Verify that 2 hzus have status UNBONDED, while the third has status BONDED
// 	// Verify that 2 hzus have completion time set, while the third has no completion time

// 	// verify that the stTokenBurnAmount is what we expect, based on the validators on the hzu
// }

// TODO: BELOW TESTS
// Test success case - update unbonding records and get tokens to burn
// Test failure case - epoch unbonding record DNE
// Test failure case - HostZoneUnbonding DNE
// Test failure case - Amount too big

// BurnTokens Tests
func (s *KeeperTestSuite) TestBurnTokens_Success() {
	_ = s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)

	zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	s.Require().NoError(err)
	s.Require().Equal(s.App.BankKeeper.GetBalance(s.Ctx(), zoneAccount, stAtom).Amount.Int64(), int64(300_010), "initial token balance is 300_010")

	err = s.App.StakeibcKeeper.BurnTokens(s.Ctx(), hostZone, int64(123456))
	s.Require().NoError(err)

	s.Require().Equal(s.App.BankKeeper.GetBalance(s.Ctx(), zoneAccount, stAtom).Amount.Int64(), int64(176_554), "post burn amount is 176_554")
}

// Test failure case - could not parse coin
func (s *KeeperTestSuite) TestBurnTokens_CouldNotParseCoin() {
	_ = s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	hostZone.HostDenom = ":"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx(), hostZone, int64(123456))
	s.Require().EqualError(err, "could not parse burnCoin: 123456.000000000000000000st:. err: invalid decimal coin expression: 123456.000000000000000000st:: invalid coins")
}

// Test failure case - could not decode address
func (s *KeeperTestSuite) TestBurnTokens_CouldNotParseAddress() {
	_ = s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	hostZone.Address = "invalid"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx(), hostZone, int64(123456))
	s.Require().EqualError(err, "could not bech32 decode address invalid of zone with id: GAIA")
}

// Test failure case - could not send coins from account to module
func (s *KeeperTestSuite) TestBurnTokens_CouldNotSendCoinsFromAccountToModule() {
	_ = s.SetupUndelegateCallback()

	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	hostZone.HostDenom = "coinDNE"

	err := s.App.StakeibcKeeper.BurnTokens(s.Ctx(), hostZone, int64(123456))
	s.Require().EqualError(err, "could not send coins from account stride1755g4dkhpw73gz9h9nwhlcefc6sdf8kcmvcwrk4rxfrz8xpxxjms7savm8 to module stakeibc. err: 0stcoinDNE is smaller than 123456stcoinDNE: insufficient funds")
}
