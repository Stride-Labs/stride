package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type UndelegateCallbackState struct {
	callbackArgs types.UndelegateCallback
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
	stakedBal := uint64(1_000_000)
	val1Bal := uint64(400_000)
	val2Bal := uint64(stakedBal) - val1Bal
	balanceToStake := int64(300_000)
	val1RelAmt := int64(120_000)
	val2RelAmt := int64(180_000)

	// STATE
	// host zone
	// Validators (2)
	// EpochUnbondingRecord (1)
	// HZU (1)
	// Account with tokens

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
	hostZone := stakeibc.HostZone{
		ChainId:        chainId,
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Validators:     []*types.Validator{&val1, &val2},
		StakedBal:      stakedBal,
	}
	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToStake,
		Status:             recordtypes.DepositRecord_STAKE,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx(), depositRecord)

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	// add
	msgs = append(msgs, &stakingTypes.MsgUndelegate{}, &stakingTypes.MsgUndelegate{})
	ack := s.ICAPacketAcknowledgement(msgs)
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
		EpochUnbondingRecordIds: []uint64{1, 2},
	}
	args, err := s.App.StakeibcKeeper.MarshalUndelegateCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)

	return UndelegateCallbackTestCase{
		initialState: UndelegateCallbackState{
			callbackArgs: callbackArgs,
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

	err := stakeibckeeper.UndelegateCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// Check that stakedBal has decreased
	// Check that Delegations on validators have decreased
	// Check that hzu are updated correctly
	// -- Check that the completion time is set on the hzu
	// -- Check that the hzu status is set to UNBONDED
	// Check that tokens are burned
}

func (s *KeeperTestSuite) checkStateIfUndelegateCallbackFailed(tc UndelegateCallbackTestCase) {
	// Confirm stakedBal has not increased
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	s.Require().Equal(int64(tc.initialState.stakedBal), int64(hostZone.StakedBal), "stakedBal should not have increased")

	// Confirm deposit record has NOT been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx())
	s.Require().Len(records, 1, "number of deposit records")
	record := records[0]
	s.Require().Equal(recordtypes.DepositRecord_STAKE, record.Status, "deposit record status should not have changed")
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
	s.Require().EqualError(err, "unexpected EOF")
	s.checkStateIfUndelegateCallbackFailed(tc)
}
