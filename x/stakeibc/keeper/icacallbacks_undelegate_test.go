package keeper_test

import (
	"time"

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
		stAtomBalance: sdk.NewInt64Coin(stAtom, balanceToUnstake),
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
		HostZoneId: chainId,
		Status:     recordtypes.HostZoneUnbonding_UNBONDED,
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
	// TODO add an unbonding time to the response
	msgUndelegateResponse := &stakingTypes.MsgUndelegateResponse{CompletionTime: time.Now()}
	ack := s.ICAPacketAcknowledgement(msgs, &msgUndelegateResponse)
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
	// initialState := tc.initialState
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
	// Check that stakedBal has not decreased
	// Check that Delegations on validators have not decreased
	// Check that hzu are not updated correctly
	// -- Check that the completion time is not set on the hzu
	// -- Check that the hzu status is set to UNBONDING
	// Check that tokens are not burned
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
