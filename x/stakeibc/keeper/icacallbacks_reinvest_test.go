package keeper_test

import (
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"
	_ "github.com/stretchr/testify/suite"

	"github.com/Stride-Labs/stride/v14/app/apptesting"
	epochtypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v14/x/interchainquery/types"

	icacallbacktypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	recordtypes "github.com/Stride-Labs/stride/v14/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v14/x/stakeibc/keeper"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
	stakeibctypes "github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

type ReinvestCallbackState struct {
	hostZone               stakeibctypes.HostZone
	reinvestAmt            sdkmath.Int
	callbackArgs           types.ReinvestCallback
	depositRecord          recordtypes.DepositRecord
	durationUntilNextEpoch time.Duration
}

type ReinvestCallbackArgs struct {
	packet      channeltypes.Packet
	ackResponse *icacallbacktypes.AcknowledgementResponse
	args        []byte
}

type ReinvestCallbackTestCase struct {
	initialState ReinvestCallbackState
	validArgs    ReinvestCallbackArgs
}

func (s *KeeperTestSuite) SetupReinvestCallback() ReinvestCallbackTestCase {
	reinvestAmt := sdkmath.NewInt(1_000)
	feeAddress := apptesting.CreateRandomAccounts(1)[0].String() // must be valid bech32 address

	hostZone := stakeibctypes.HostZone{
		ChainId:        HostChainId,
		HostDenom:      Atom,
		IbcDenom:       IbcAtom,
		RedemptionRate: sdk.NewDec(1.0),
		ConnectionId:   ibctesting.FirstConnectionID,
		FeeIcaAddress:  feeAddress,
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	expectedNewDepositRecord := recordtypes.DepositRecord{
		Id:                 0,
		DepositEpochNumber: 1,
		HostZoneId:         HostChainId,
		Amount:             reinvestAmt,
		Status:             recordtypes.DepositRecord_DELEGATION_QUEUE,
		Source:             recordtypes.DepositRecord_WITHDRAWAL_ICA,
	}

	durationUntilNextEpoch := time.Minute
	blockTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Ctx = s.Ctx.WithBlockTime(blockTime)

	epochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(blockTime.Add(durationUntilNextEpoch).UnixNano()),
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	packet := channeltypes.Packet{}
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_SUCCESS}
	callbackArgs := types.ReinvestCallback{
		HostZoneId:     HostChainId,
		ReinvestAmount: sdk.NewCoin(Atom, reinvestAmt),
	}
	args, err := s.App.StakeibcKeeper.MarshalReinvestCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	// Mock the latest client height for the ICQ submission
	s.MockClientLatestHeight(1)

	return ReinvestCallbackTestCase{
		initialState: ReinvestCallbackState{
			hostZone:               hostZone,
			reinvestAmt:            reinvestAmt,
			callbackArgs:           callbackArgs,
			depositRecord:          expectedNewDepositRecord,
			durationUntilNextEpoch: durationUntilNextEpoch,
		},
		validArgs: ReinvestCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        args,
		},
	}
}

func (s *KeeperTestSuite) TestReinvestCallback_Successful() {
	tc := s.SetupReinvestCallback()
	initialState := tc.initialState
	expectedRecord := initialState.depositRecord
	validArgs := tc.validArgs

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, validArgs.packet, validArgs.ackResponse, validArgs.args)
	s.Require().NoError(err)

	// Confirm deposit record has been added
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 1, "number of deposit records")
	record := records[0]

	// Confirm deposit record fields match those expected
	s.Require().Equal(int64(expectedRecord.Id), int64(record.Id), "deposit record Id")
	s.Require().Equal(expectedRecord.Amount, record.Amount, "deposit record Amount")
	s.Require().Equal(expectedRecord.HostZoneId, record.HostZoneId, "deposit record HostZoneId")
	s.Require().Equal(expectedRecord.Status, record.Status, "deposit record Status")
	s.Require().Equal(expectedRecord.Source, record.Source, "deposit record Source")
	s.Require().Equal(int64(expectedRecord.DepositEpochNumber), int64(record.DepositEpochNumber), "deposit record DepositEpochNumber")

	// Confirm an interchain query was submitted for the fee account balance
	allQueries := s.App.InterchainqueryKeeper.AllQueries(s.Ctx)
	s.Require().Len(allQueries, 1, "should be 1 query submitted")

	query := allQueries[0]
	s.Require().Equal(stakeibckeeper.ICQCallbackID_FeeBalance, query.CallbackId, "query callback ID")
	s.Require().Equal(HostChainId, query.ChainId, "query chain ID")
	s.Require().Equal(ibctesting.FirstConnectionID, query.ConnectionId, "query connection ID")
	s.Require().Equal(icqtypes.BANK_STORE_QUERY_WITH_PROOF, query.QueryType, "query type")
	s.Require().Equal(tc.initialState.durationUntilNextEpoch, query.TimeoutDuration, "query timeout duration")
}

func (s *KeeperTestSuite) checkReinvestStateIfCallbackFailed(tc ReinvestCallbackTestCase) {
	// Confirm deposit record has not been added
	records := s.App.RecordsKeeper.GetAllDepositRecord(s.Ctx)
	s.Require().Len(records, 0, "number of deposit records")
}

func (s *KeeperTestSuite) TestReinvestCallback_ReinvestCallbackTimeout() {
	tc := s.SetupReinvestCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkReinvestStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestReinvestCallback_ReinvestCallbackErrorOnHost() {
	tc := s.SetupReinvestCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkReinvestStateIfCallbackFailed(tc)
}

func (s *KeeperTestSuite) TestReinvestCallback_WrongCallbackArgs() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal reinvest callback args: unexpected EOF: unable to unmarshal data structure")
}

func (s *KeeperTestSuite) TestReinvestCallback_HostZoneNotFound() {
	tc := s.SetupReinvestCallback()

	// Remove the host zone
	s.App.StakeibcKeeper.RemoveHostZone(s.Ctx, HostChainId)

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().ErrorContains(err, "host zone GAIA not found: host zone not found")
}

func (s *KeeperTestSuite) TestReinvestCallback_NoFeeAccount() {
	tc := s.SetupReinvestCallback()

	// Remove the fee account
	badHostZone := tc.initialState.hostZone
	badHostZone.FeeIcaAddress = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().EqualError(err, "no fee account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestReinvestCallback_InvalidFeeAccountAddress() {
	tc := s.SetupReinvestCallback()

	// Remove the fee account
	badHostZone := tc.initialState.hostZone
	badHostZone.FeeIcaAddress = "invalid_fee_account"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().ErrorContains(err, "invalid fee account address, could not decode")
}

func (s *KeeperTestSuite) TestReinvestCallback_MissingEpoch() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs

	// Remove epoch tracker
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().ErrorContains(err, "no number for epoch (stride_epoch)")
}

func (s *KeeperTestSuite) TestReinvestCallback_FailedToSubmitQuery() {
	tc := s.SetupReinvestCallback()
	invalidArgs := tc.validArgs

	// Remove the connection ID from the host zone so that the query submission fails
	badHostZone := tc.initialState.hostZone
	badHostZone.ConnectionId = ""
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := s.App.StakeibcKeeper.ReinvestCallback(s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().EqualError(err, "connection-id cannot be empty: invalid interchain query request")
}
