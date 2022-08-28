package keeper_test

import (
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
	// initialState := tc.initialState

	// // Check that stakedBal has NOT decreased on the host zone
	// hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	// s.Require().True(found)
	// s.Require().Equal(int64(hostZone.StakedBal), initialState.stakedBal)

	// // Check that Delegations on validators have NOT decreased
	// s.Require().True(len(hostZone.Validators) == 2, "Expected 2 validators")
	// val1 := hostZone.Validators[0]
	// s.Require().Equal(int64(val1.DelegationAmt), int64(initialState.val1Bal))
	// val2 := hostZone.Validators[1]
	// s.Require().Equal(int64(val2.DelegationAmt), int64(initialState.val2Bal))

	// epochUnbondingRecord, found := s.App.RecordsKeeper.GetEpochUnbondingRecord(s.Ctx(), initialState.epochNumber)
	// s.Require().True(found)
	// s.Require().Equal(len(epochUnbondingRecord.HostZoneUnbondings), 1)
	// hzu := epochUnbondingRecord.HostZoneUnbondings[0]
	// s.Require().Equal(int64(hzu.UnbondingTime), 0, "completion time is NOT set on the hzu")
	// s.Require().Equal(hzu.Status, recordtypes.HostZoneUnbonding_BONDED, "hzu status is set to BONDED")
	// zoneAccount, err := sdk.AccAddressFromBech32(hostZone.Address)
	// s.Require().NoError(err)
	// s.Require().Equal(s.App.BankKeeper.GetBalance(s.Ctx(), zoneAccount, stAtom).Amount.Int64(), initialState.balanceToUnstake, "tokens are NOT burned")
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
