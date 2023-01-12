package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v4/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

type DelegateCallbackState struct {
	stakedBal      sdk.Int
	val1Bal        sdk.Int
	val2Bal        sdk.Int
	val1RelAmt     sdk.Int
	val2RelAmt     sdk.Int
	balanceToStake sdk.Int
	depositRecord  recordtypes.DepositRecord
	callbackArgs   types.DelegateCallback
}

type DelegateCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type DelegateCallbackTestCase struct {
	initialState DelegateCallbackState
	validArgs    DelegateCallbackArgs
}

func (s *KeeperTestSuite) SetupDelegateCallback() DelegateCallbackTestCase {
	stakedBal := sdk.NewInt(1_000_000)
	val1Bal := sdk.NewInt(400_000)
	val2Bal := stakedBal.Sub(val1Bal)
	balanceToStake := sdk.NewInt(300_000)
	val1RelAmt := sdk.NewInt(120_000)
	val2RelAmt := sdk.NewInt(180_000)

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
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Validators:     []*types.Validator{&val1, &val2},
		StakedBal:      stakedBal,
	}
	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         HostChainId,
		Amount:             balanceToStake,
		Status:             recordtypes.DepositRecord_DELEGATION_QUEUE,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx, depositRecord)

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	msgs = append(msgs, &stakingTypes.MsgDelegate{}, &stakingTypes.MsgDelegate{})
	ack := s.ICAPacketAcknowledgement(msgs, nil)
	val1SplitDelegation := types.SplitDelegation{
		Validator: val1.Address,
		Amount:    val1RelAmt,
	}
	val2SplitDelegation := types.SplitDelegation{
		Validator: val2.Address,
		Amount:    val2RelAmt,
	}
	callbackArgs := types.DelegateCallback{
		HostZoneId:       HostChainId,
		DepositRecordId:  depositRecord.Id,
		SplitDelegations: []*types.SplitDelegation{&val1SplitDelegation, &val2SplitDelegation},
	}
	args, err := s.App.StakeibcKeeper.MarshalDelegateCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return DelegateCallbackTestCase{
		initialState: DelegateCallbackState{
			stakedBal:      stakedBal,
			balanceToStake: balanceToStake,
			depositRecord:  depositRecord,
			callbackArgs:   callbackArgs,
			val1Bal:        val1Bal,
			val2Bal:        val2Bal,
			val1RelAmt:     val1RelAmt,
			val2RelAmt:     val2RelAmt,
		},
		validArgs: DelegateCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestDelegateCallback_Successful() {
	tc := s.SetupDelegateCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx, validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// Confirm stakedBal has increased
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(initialState.stakedBal.Add(initialState.balanceToStake), hostZone.StakedBal, "stakedBal should have increased")

	// Confirm delegations have been added to validators
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(initialState.val1Bal.Add(initialState.val1RelAmt), val1.DelegationAmt, "val1 balance should have increased")
	s.Require().Equal(initialState.val2Bal.Add(initialState.val2RelAmt), val2.DelegationAmt, "val2 balance should have increased")

	// Confirm deposit record has been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 0, "number of deposit records")
}

func (s *KeeperTestSuite) checkDelegateStateIfCallbackFailed(tc DelegateCallbackTestCase) {
	// Confirm stakedBal has not increased
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(tc.initialState.stakedBal, hostZone.StakedBal, "stakedBal should not have increased")

	// Confirm deposit record has NOT been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")
	record := records[0]
	s.Require().Equal(recordtypes.DepositRecord_DELEGATION_QUEUE, record.Status, "deposit record status should not have changed")
}

func (s *KeeperTestSuite) TestDelegateCallback_DelegateCallbackTimeout() {
	tc := s.SetupDelegateCallback()
	invalidArgs := tc.validArgs
	// a nil ack means the request timed out
	invalidArgs.ack = nil
	err := stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_DelegateCallbackErrorOnHost() {
	tc := s.SetupDelegateCallback()
	invalidArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	invalidArgs.ack = &fullAck

	err := stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_WrongCallbackArgs() {
	tc := s.SetupDelegateCallback()
	invalidArgs := tc.validArgs

	err := stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "unexpected EOF")
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_HostNotFound() {
	tc := s.SetupDelegateCallback()
	invalidArgs := tc.validArgs
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)
	err := stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, invalidArgs.args)
	s.Require().EqualError(err, "host zone not found GAIA: invalid request")

	// Confirm deposit record has NOT been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")
	record := records[0]
	s.Require().Equal(recordtypes.DepositRecord_DELEGATION_QUEUE, record.Status, "deposit record status should not have changed")
}

func (s *KeeperTestSuite) TestDelegateCallback_MissingValidator() {
	tc := s.SetupDelegateCallback()
	invalidArgs := tc.validArgs
	badSplitDelegation := types.SplitDelegation{
		Validator: "address_dne",
		Amount:    sdk.NewInt(1234),
	}
	callbackArgs := types.DelegateCallback{
		HostZoneId:       HostChainId,
		DepositRecordId:  1,
		SplitDelegations: []*types.SplitDelegation{&badSplitDelegation},
	}
	args, err := s.App.StakeibcKeeper.MarshalDelegateCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)
	err = stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ack, args)
	s.Require().EqualError(err, "Failed to add delegation to validator: can't change delegation on validator")
	s.checkDelegateStateIfCallbackFailed(tc)
}
