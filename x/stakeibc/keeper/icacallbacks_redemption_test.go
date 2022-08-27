package keeper_test
package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	recordtypes "github.com/Stride-Labs/stride/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	stakeibc "github.com/Stride-Labs/stride/x/stakeibc/types"
)

type RedemptionCallbackState struct {
	callbackArgs          types.RedemptionCallback
	redemptionTransferAmt uint64
}

type RedemptionCallbackArgs struct {
	packet channeltypes.Packet
	ack    *channeltypes.Acknowledgement
	args   []byte
}

type RedemptionCallbackTestCase struct {
	initialState RedemptionCallbackState
	validArgs    RedemptionCallbackArgs
}

func (s *KeeperTestSuite) SetupRedemptionCallback() RedemptionCallbackTestCase {
	redemptionTransferAmt := uint64(1_000_000)

	userRedemptionRecord1 := recordtypes.UserRedemptionRecord{
		Id:          "sender.2.GAIA",
		IsClaimable: true,
	}

	userRedemptionRecord2 := recordtypes.UserRedemptionRecord{
		Id:          "sender.2.GAIA",
		IsClaimable: true,
	}

	hostZoneUnbonding := recordtypes.HostZoneUnbonding{
		StTokenAmount:         100,
		NativeTokenAmount:     100,
		Denom:                 "stake",
		HostZoneId:            "hostZoneId",
		UnbondingTime:         100,
		Status:                recordtypes.HostZoneUnbonding_TRANSFERRED,
		UserRedemptionRecords: []string{"1", "2"},
	}

	epochUnbondingRecord := recordtypes.EpochUnbondingRecord{
		EpochNumber:        1,
		HostZoneUnbondings: []*recordtypes.HostZoneUnbonding{&hostZoneUnbonding},
	}

	hostZone := stakeibc.HostZone{
		ChainId:        chainId,
		HostDenom:      atom,
		IBCDenom:       ibcAtom,
		RedemptionRate: sdk.NewDec(1.0),
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx(), hostZone)
	s.App.RecordsKeeper.SetEpochUnbondingRecord(s.Ctx(), epochUnbondingRecord)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), userRedemptionRecord1)
	s.App.RecordsKeeper.SetUserRedemptionRecord(s.Ctx(), userRedemptionRecord2)

	packet := channeltypes.Packet{}
	var msgs []sdk.Msg
	msgs = append(msgs, &banktypes.MsgSend{})
	ack := s.ICAPacketAcknowledgement(msgs)
	callbackArgs := types.RedemptionCallback{
		HostZoneId:            chainId,
		UnbondingEpochNumbers: []uint64{1},
	}
	args, err := s.App.StakeibcKeeper.MarshalRedemptionCallbackArgs(s.Ctx(), callbackArgs)
	s.Require().NoError(err)

	return RedemptionCallbackTestCase{
		initialState: RedemptionCallbackState{
			callbackArgs:          callbackArgs,
			redemptionTransferAmt: redemptionTransferAmt,
		},
		validArgs: RedemptionCallbackArgs{
			packet: packet,
			ack:    &ack,
			args:   args,
		},
	}
}

func (s *KeeperTestSuite) TestRedemptionCallback_Successful() {
	tc := s.SetupRedemptionCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)

	// check that hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_TRANSFERRED for the epochs in the callbackArgs.UnbondingEpochNumbers
	// check that userRedemptionRecord.IsClaimable = true
	//
}

func (s *KeeperTestSuite) checkRedemptionStateIfCallbackFailed(tc RedemptionCallbackTestCase) {
	// Confirm ...
}

func (s *KeeperTestSuite) TestRedemptionCallback_RedemptionCallbackTimeout() {
	tc := s.SetupRedemptionCallback()
	validArgs := tc.validArgs
	// a nil ack means the request timed out
	validArgs.ack = nil
	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_RedemptionCallbackErrorOnHost() {
	tc := s.SetupRedemptionCallback()
	validArgs := tc.validArgs
	// an error ack means the tx failed on the host
	fullAck := channeltypes.Acknowledgement{Response: &channeltypes.Acknowledgement_Error{Error: "error"}}
	validArgs.ack = &fullAck

	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, validArgs.args)
	s.Require().NoError(err)
	s.checkRedemptionStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestRedemptionCallback_WrongCallbackArgs() {
	tc := s.SetupRedemptionCallback()
	validArgs := tc.validArgs

	err := stakeibckeeper.RedemptionCallback(s.App.StakeibcKeeper, s.Ctx(), validArgs.packet, validArgs.ack, []byte("random bytes"))
	s.Require().EqualError(err, "unexpected EOF")
	s.checkRedemptionStateIfCallbackFailed(tc)
}
