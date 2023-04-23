package keeper_test

import (
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	icacallbackstypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

// Helper function to setup the detokenize ICA callback test
// Returns the serialized callback args which will be an input parameter
// to the callback
func (s *KeeperTestSuite) SetupTestDetokenizeCallback() []byte {
	// Store the LSMDeposit record with status DETOKENIZATION_IN_PROGRESS
	deposit := types.LSMTokenDeposit{
		ChainId: HostChainId,
		Denom:   LSMTokenBaseDenom,
		Status:  types.DETOKENIZATION_IN_PROGRESS,
	}
	s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

	// Return the deposit as callback args
	callbackBz, err := proto.Marshal(&types.DetokenizeSharesCallback{
		Deposit: &deposit,
	})
	s.Require().NoError(err, "no error expected when marshalling callback args")

	return callbackBz
}

func (s *KeeperTestSuite) TestDetokenizeCallback_Successful() {
	callbackBz := s.SetupTestDetokenizeCallback()

	// Call the callback with a successful response
	ackSuccess := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_SUCCESS,
	}
	err := keeper.DetokenizeCallback(s.App.StakeibcKeeper, s.Ctx, channeltypes.Packet{}, ackSuccess, callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// Check that the deposit was removed
	_, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().False(found, "deposit should have been removed")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_InvalidCallbackArgs() {
	s.SetupTestDetokenizeCallback()

	// Call the callback with a successful ack, but invalid callback args
	invalidCallbackArgs := []byte{1, 2, 3}
	ackSuccess := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_SUCCESS,
	}
	err := keeper.DetokenizeCallback(s.App.StakeibcKeeper, s.Ctx, channeltypes.Packet{}, ackSuccess, invalidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal detokenize callback")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_AckTimeout() {
	callbackBz := s.SetupTestDetokenizeCallback()

	// Call the callback with a timed-out response
	ackTimeout := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_TIMEOUT,
	}
	err := keeper.DetokenizeCallback(s.App.StakeibcKeeper, s.Ctx, channeltypes.Packet{}, ackTimeout, callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// The deposit should still be there in status IN_PROGRESS
	deposit, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().True(found, "deposit should not have been removed")
	s.Require().Equal(types.DETOKENIZATION_IN_PROGRESS.String(), deposit.Status.String(), "deposit status")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_AckFailure() {
	callbackBz := s.SetupTestDetokenizeCallback()

	// Call the callback with an ack-failure response
	ackFailure := &icacallbackstypes.AcknowledgementResponse{
		Status: icacallbackstypes.AckResponseStatus_FAILURE,
	}
	err := keeper.DetokenizeCallback(s.App.StakeibcKeeper, s.Ctx, channeltypes.Packet{}, ackFailure, callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// The deposit status should be FAILED
	deposit, found := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().True(found, "deposit should not have been removed")
	s.Require().Equal(types.DETOKENIZATION_FAILED.String(), deposit.Status.String(), "deposit status")
}
