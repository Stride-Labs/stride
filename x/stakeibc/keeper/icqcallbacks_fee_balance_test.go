package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctesting "github.com/cosmos/ibc-go/v7/testing"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"

	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icqtypes "github.com/Stride-Labs/stride/v9/x/interchainquery/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type FeeBalanceICQCallbackState struct {
	hostZone         stakeibctypes.HostZone
	feeChannel       Channel
	feeBalance       int64
	startICASequence uint64
}

type FeeBalanceICQCallbackArgs struct {
	query        icqtypes.Query
	callbackArgs []byte
}

type FeeBalanceICQCallbackTestCase struct {
	initialState FeeBalanceICQCallbackState
	validArgs    FeeBalanceICQCallbackArgs
}

func (s *KeeperTestSuite) SetupFeeBalanceCallbackTest() FeeBalanceICQCallbackTestCase {
	feeAccountOwner := fmt.Sprintf("%s.%s", HostChainId, "FEE")
	feeChannelId := s.CreateICAChannel(feeAccountOwner)
	feeAddress := s.IcaAddresses[feeAccountOwner]

	hostZone := stakeibctypes.HostZone{
		ChainId:      HostChainId,
		HostDenom:    Atom,
		ConnectionId: ibctesting.FirstConnectionID,
		FeeAccount: &stakeibctypes.ICAAccount{
			Address: feeAddress,
			Target:  stakeibctypes.ICAAccountType_FEE,
		},
		TransferChannelId: ibctesting.FirstChannelID,
	}

	strideEpochTracker := stakeibctypes.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        1,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}

	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, strideEpochTracker)

	// Get the next sequence number to confirm if an ICA was sent
	feePortId := icatypes.ControllerPortPrefix + feeAccountOwner
	startSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found before ICA")

	feeBalance := int64(100)
	queryResponse := s.CreateBalanceQueryResponse(feeBalance, Atom)

	return FeeBalanceICQCallbackTestCase{
		initialState: FeeBalanceICQCallbackState{
			hostZone: hostZone,
			feeChannel: Channel{
				PortID:    feePortId,
				ChannelID: feeChannelId,
			},
			feeBalance:       feeBalance,
			startICASequence: startSequence,
		},
		validArgs: FeeBalanceICQCallbackArgs{
			query: icqtypes.Query{
				Id:      "0",
				ChainId: HostChainId,
			},
			callbackArgs: queryResponse,
		},
	}
}

// Helper function to check that no ICA was submitted in the case of the function exiting prematurely
func (s *KeeperTestSuite) CheckNoICASubmitted(tc FeeBalanceICQCallbackTestCase) {
	feeChannel := tc.initialState.feeChannel
	feePortId := feeChannel.PortID
	feeChannelId := feeChannel.ChannelID

	// The sequence number should not have incremented
	expectedSequence := tc.initialState.startICASequence
	endSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found after ICA")
	s.Require().Equal(expectedSequence, endSequence, "sequence number after ICA")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_Successful() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Get the sequence number before the ICA is submitted to confirm it incremented
	feeChannel := tc.initialState.feeChannel
	feePortId := feeChannel.PortID
	feeChannelId := feeChannel.ChannelID

	// Call the ICQ callback
	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm the sequence number was incremented
	expectedSequence := tc.initialState.startICASequence + 1
	actualSequence, found := s.App.IBCKeeper.ChannelKeeper.GetNextSequenceSend(s.Ctx, feePortId, feeChannelId)
	s.Require().True(found, "sequence number not found after ICA")
	s.Require().Equal(expectedSequence, actualSequence, "sequence number after ICA")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_EmptyCallbackArgs() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Replace the query response an empty byte array (this happens when the account has not been registered yet)
	emptyCallbackArgs := []byte{}

	// It should short circuit but not throw an error
	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, emptyCallbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// No ICA should have been submitted
	s.CheckNoICASubmitted(tc)
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_ZeroBalance() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Replace the query response with a coin that has a nil amount
	tc.validArgs.callbackArgs = s.CreateBalanceQueryResponse(0, Atom)

	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.CheckNoICASubmitted(tc)
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_ZeroBalanceImplied() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Replace the query response with a coin that has a nil amount
	coin := sdk.Coin{}
	coinBz := s.App.RecordsKeeper.Cdc.MustMarshal(&coin)
	tc.validArgs.callbackArgs = coinBz

	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().NoError(err)

	// Confirm revinvestment callback was not created
	s.Require().Len(s.App.IcacallbacksKeeper.GetAllCallbackData(s.Ctx), 0, "number of callbacks found")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_HostZoneNotFound() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Submit callback with incorrect host zone
	invalidQuery := tc.validArgs.query
	invalidQuery.ChainId = "fake_host_zone"
	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, invalidQuery)
	s.Require().EqualError(err, "no registered zone for queried chain ID (fake_host_zone): host zone not found")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_InvalidArgs() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Submit callback with invalid callback args (so that it can't unmarshal into a coin)
	invalidArgs := []byte("random bytes")
	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs, tc.validArgs.query)

	s.Require().ErrorContains(err, "unable to determine balance from query response")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_NoFeeAccount() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Remove the fee account
	badHostZone := tc.initialState.hostZone
	badHostZone.FeeAccount = nil
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "no fee account found for GAIA: ICA acccount not found on host zone")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_FailedToCalculatedTimeout() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Remove the epoch tracker so that it cannot calculate the ICA timeout
	s.App.StakeibcKeeper.RemoveEpochTracker(s.Ctx, epochtypes.STRIDE_EPOCH)

	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "Failed to get ICATimeout from stride_epoch epoch:")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_NoTransferChannel() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Set an invalid transfer channel so that the counterparty channel cannot be found
	badHostZone := tc.initialState.hostZone
	badHostZone.TransferChannelId = "channel-X"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().EqualError(err, "transfer channel channel-X not found: channel not found")
}

func (s *KeeperTestSuite) TestFeeBalanceCallback_FailedSubmitTx() {
	tc := s.SetupFeeBalanceCallbackTest()

	// Remove connectionId from host zone so the ICA tx fails
	badHostZone := tc.initialState.hostZone
	badHostZone.ConnectionId = "connection-X"
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, badHostZone)

	err := stakeibckeeper.FeeBalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.callbackArgs, tc.validArgs.query)
	s.Require().ErrorContains(err, "Failed to SubmitTxs")
	s.Require().ErrorContains(err, "invalid connection id, connection-X not found")
}
