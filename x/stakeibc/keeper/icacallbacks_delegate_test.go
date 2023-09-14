package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"

	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type DelegateCallbackState struct {
	totalDelegation sdkmath.Int
	val1Bal         sdkmath.Int
	val2Bal         sdkmath.Int
	val1RelAmt      sdkmath.Int
	val2RelAmt      sdkmath.Int
	balanceToStake  sdkmath.Int
	depositRecord   recordtypes.DepositRecord
	callbackArgs    types.DelegateCallback
}

type DelegateCallbackArgs struct {
	packet      channeltypes.Packet
	ackResponse *icacallbacktypes.AcknowledgementResponse
	args        []byte
}

type DelegateCallbackTestCase struct {
	initialState DelegateCallbackState
	validArgs    DelegateCallbackArgs
}

func (s *KeeperTestSuite) SetupDelegateCallback() DelegateCallbackTestCase {
	totalDelegation := sdkmath.NewInt(1_000_000)
	val1Bal := sdkmath.NewInt(400_000)
	val2Bal := totalDelegation.Sub(val1Bal)
	balanceToStake := sdkmath.NewInt(300_000)
	val1RelAmt := sdkmath.NewInt(120_000)
	val2RelAmt := sdkmath.NewInt(180_000)

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
	hostZone := types.HostZone{
		ChainId:          HostChainId,
		HostDenom:        Atom,
		IbcDenom:         IbcAtom,
		RedemptionRate:   sdk.NewDec(1.0),
		Validators:       []*types.Validator{&val1, &val2},
		TotalDelegations: totalDelegation,
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

	// Mock the ack response
	packet := channeltypes.Packet{}
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_SUCCESS}

	// Mock the callback args
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
	callbackArgsBz, err := s.App.StakeibcKeeper.MarshalDelegateCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return DelegateCallbackTestCase{
		initialState: DelegateCallbackState{
			totalDelegation: totalDelegation,
			balanceToStake:  balanceToStake,
			depositRecord:   depositRecord,
			callbackArgs:    callbackArgs,
			val1Bal:         val1Bal,
			val2Bal:         val2Bal,
			val1RelAmt:      val1RelAmt,
			val2RelAmt:      val2RelAmt,
		},
		validArgs: DelegateCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        callbackArgsBz,
		},
	}
}

func (s *KeeperTestSuite) TestDelegateCallback_Successful() {
	tc := s.SetupDelegateCallback()
	initialState := tc.initialState
	validArgs := tc.validArgs

	err := s.App.StakeibcKeeper.DelegateCallback(s.Ctx, validArgs.packet, validArgs.ackResponse, validArgs.args)
	s.Require().NoError(err)

	// Confirm total delegation has increased
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(initialState.totalDelegation.Add(initialState.balanceToStake), hostZone.TotalDelegations, "total delegation should have increased")

	// Confirm delegations have been added to validators and number delegation changes in progress was reduced
	val1 := hostZone.Validators[0]
	val2 := hostZone.Validators[1]
	s.Require().Equal(val1.Delegation, initialState.val1Bal.Add(initialState.val1RelAmt), "val1 balance should have increased")
	s.Require().Equal(val2.Delegation, initialState.val2Bal.Add(initialState.val2RelAmt), "val2 balance should have increased")

	// Confirm the number of delegations in progress has decreased
	s.Require().Equal(0, int(val1.DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(val2.DelegationChangesInProgress), "val2 delegation changes in progress")

	// Confirm deposit record has been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 0, "number of deposit records")
}

func (s *KeeperTestSuite) checkDelegateStateIfCallbackFailed(tc DelegateCallbackTestCase) {
	// Confirm total delegation has not increased
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found)
	s.Require().Equal(tc.initialState.totalDelegation, hostZone.TotalDelegations, "total delegation should not have increased")

	// Confirm the number of delegations in progress has decreased
	s.Require().Equal(0, int(hostZone.Validators[0].DelegationChangesInProgress), "val1 delegation changes in progress")
	s.Require().Equal(0, int(hostZone.Validators[1].DelegationChangesInProgress), "val2 delegation changes in progress")

	// Confirm deposit record has NOT been removed
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")
	record := records[0]
	s.Require().Equal(recordtypes.DepositRecord_DELEGATION_QUEUE, record.Status, "deposit record status should not have changed")
}

func (s *KeeperTestSuite) TestDelegateCallback_DelegateCallbackTimeout() {
	tc := s.SetupDelegateCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := s.App.StakeibcKeeper.DelegateCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_DelegateCallbackErrorOnHost() {
	tc := s.SetupDelegateCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := s.App.StakeibcKeeper.DelegateCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegateStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestDelegateCallback_WrongCallbackArgs() {
	tc := s.SetupDelegateCallback()

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := s.App.StakeibcKeeper.DelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal delegate callback args: unexpected EOF: unable to unmarshal data structure")
}

func (s *KeeperTestSuite) TestDelegateCallback_HostNotFound() {
	tc := s.SetupDelegateCallback()

	// Remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	err := s.App.StakeibcKeeper.DelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, "host zone not found GAIA: invalid request")
}

func (s *KeeperTestSuite) TestDelegateCallback_MissingValidator() {
	tc := s.SetupDelegateCallback()

	// Update the callback args such that a validator is missing
	badSplitDelegation := types.SplitDelegation{
		Validator: "address_dne",
		Amount:    sdkmath.NewInt(1234),
	}
	callbackArgs := types.DelegateCallback{
		HostZoneId:       HostChainId,
		DepositRecordId:  1,
		SplitDelegations: []*types.SplitDelegation{&badSplitDelegation},
	}
	invalidCallbackArgs, err := s.App.StakeibcKeeper.MarshalDelegateCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	err = s.App.StakeibcKeeper.DelegateCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidCallbackArgs)
	s.Require().ErrorContains(err, "validator not found")
}
