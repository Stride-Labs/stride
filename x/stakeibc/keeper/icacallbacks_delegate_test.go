package keeper_test

import (
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


type DelegateCallbackState struct {
	stakedBal 		uint64
	balanceToStake  int64
	val1Bal 		uint64
	val2Bal 		uint64
	val1RelAmt 		int64
	val2RelAmt 		int64
	depositRecord	recordtypes.DepositRecord
	callbackArgs	types.DelegateCallback
}

type DelegateCallbackArgs struct {
	packet		channeltypes.Packet
	ack			*channeltypes.Acknowledgement
	args		[]byte
}

type DelegateCallbackTestCase struct {
	initialState	DelegateCallbackState
	validArgs		DelegateCallbackArgs
}

func (s *KeeperTestSuite) SetupDelegateCallback() DelegateCallbackTestCase {
	stakedBal := uint64(1_000_000)
	val1Bal := uint64(400_000)
	val2Bal := uint64(stakedBal) - val1Bal
	balanceToStake := int64(300_000)
	val1RelAmt := int64(120_000)
	val2RelAmt := int64(180_000)
	delegatorAddress := "delegator_address"

	val1 := types.Validator{
		Name: 			   	"val1",
		Address:            "val1_address",
		DelegationAmt:      val1Bal,
	}

	val2 := types.Validator{
		Name: 			   	"val2",
		Address:            "val2_address",
		DelegationAmt:      val2Bal,
	}

	hostZone := stakeibc.HostZone{
		ChainId:        chainId,
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		Validators:		[]*types.Validator{&val1, &val2},
		StakedBal: 		stakedBal,
	}

	depositRecord := recordtypes.DepositRecord{
		Id:                 1,
		DepositEpochNumber: 1,
		HostZoneId:         chainId,
		Amount:             balanceToStake,
		Status:				recordtypes.DepositRecord_STAKE,
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.RecordsKeeper.SetDepositRecord(s.Ctx(), depositRecord)
	
	packet := channeltypes.Packet{}

	// serialize TxMsgData to get the acknowledgement
	var msgs []sdk.Msg
	msgs = append(msgs, &stakingTypes.MsgDelegate{
		DelegatorAddress: delegatorAddress,
		ValidatorAddress: val1.Address,
		Amount:           sdk.NewCoin(atom, sdk.NewInt(val1RelAmt)),
	})
	msgs = append(msgs, &stakingTypes.MsgDelegate{
		DelegatorAddress: delegatorAddress,
		ValidatorAddress: val1.Address,
		Amount:			  sdk.NewCoin(atom, sdk.NewInt(val2RelAmt)),
	})
	txMsgData := &sdk.TxMsgData{
		Data: make([]*sdk.MsgData, len(msgs)),
	}
	for i, msg := range msgs {
		msgResponse := []byte("msg_response")
		txMsgData.Data[i] = &sdk.MsgData{
			MsgType: sdk.MsgTypeURL(msg),
			Data:    msgResponse,
		}

	}
	marshalledTxMsgData, err := proto.Marshal(txMsgData)
	s.Require().NoError(err)
	ack := channeltypes.NewResultAcknowledgement(marshalledTxMsgData)
	s.Require().NoError(err)

	val1SplitDelegation := types.SplitDelegation{
		Validator: 		val1.Address,
		Amount:			uint64(val1RelAmt),
	}
	val2SplitDelegation := types.SplitDelegation{
		Validator: 		val2.Address,
		Amount:			uint64(val2RelAmt),
	}
	callbackArgs := types.DelegateCallback{
		HostZoneId:				chainId,
		DepositRecordId:		depositRecord.Id,
		SplitDelegations:		[]*types.SplitDelegation{&val1SplitDelegation, &val2SplitDelegation},
	}
	args, err := s.App.StakeibcKeeper.MarshalDelegateCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)

	return DelegateCallbackTestCase{
		initialState: DelegateCallbackState{
			stakedBal:		stakedBal,
			balanceToStake: balanceToStake,
			depositRecord:	depositRecord,
			callbackArgs:	callbackArgs,
			val1Bal: 		val1Bal,
			val2Bal: 		val2Bal,
			val1RelAmt:		val1RelAmt,
			val2RelAmt:		val2RelAmt,
		},
		validArgs: DelegateCallbackArgs{
			packet:   packet,
			ack: 	  &ack,
			args:     args,
		},
	}
}

func (s *KeeperTestSuite) TestDelegateCallback_Successful() {
	tc := s.SetupDelegateCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := stakeibckeeper.DelegateCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// Confirm stakedBal has increased
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx(), chainId)
	s.Require().True(found)
	s.Require().Equal(int64(initialState.stakedBal) + initialState.balanceToStake, int64(hostZone.StakedBal), "stakedBal should have increased")

	// Confirm delegations have been added to validators
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(int64(initialState.val1Bal) + initialState.val1RelAmt, int64(val1.DelegationAmt), "val1 balance should have increased")
	s.Require().Equal(int64(initialState.val2Bal) + initialState.val2RelAmt, int64(val2.DelegationAmt), "val2 balance should have increased")
	
	// Confirm deposit record has been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx())
	s.Require().Len(records, 0, "number of deposit records")
}
