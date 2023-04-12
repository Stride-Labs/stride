package keeper_test

import (
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	"github.com/golang/protobuf/proto" //nolint:staticcheck

	icacallbackstypes "github.com/Stride-Labs/stride/v8/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
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

/////// NOTE TO REVIEWER ////////
// I'm demonstrating two testing approaches here. This bottom test is redundant with the tests above
// Let me know which test approach you prefer!

func (s *KeeperTestSuite) TestDetokenizeCallback() {
	// Store the LSMDeposit record with status DETOKENIZATION_IN_PROGRESS
	// (This is set in the individual test case since each test is run from a separate go routine)
	deposit := types.LSMTokenDeposit{
		ChainId: HostChainId,
		Denom:   LSMTokenBaseDenom,
		Status:  types.DETOKENIZATION_IN_PROGRESS,
	}

	// Marshal the callback args
	validCallbackBz, err := proto.Marshal(&types.DetokenizeSharesCallback{
		Deposit: &deposit,
	})
	s.Require().NoError(err, "no error expected when marshalling callback args")

	testCases := []struct {
		name                       string
		ackStatus                  icacallbackstypes.AckResponseStatus
		callbackBz                 []byte
		depositFoundAfterCallback  bool
		depositStatusAfterCallback types.LSMDepositStatus
		expectedError              string
	}{
		{
			name:                      "ack success",
			ackStatus:                 icacallbackstypes.AckResponseStatus_SUCCESS,
			callbackBz:                validCallbackBz,
			depositFoundAfterCallback: false,
		},
		{
			name:                       "ack timeout",
			ackStatus:                  icacallbackstypes.AckResponseStatus_TIMEOUT,
			callbackBz:                 validCallbackBz,
			depositFoundAfterCallback:  true,
			depositStatusAfterCallback: types.DETOKENIZATION_IN_PROGRESS,
		},
		{
			name:                       "ack failed",
			ackStatus:                  icacallbackstypes.AckResponseStatus_FAILURE,
			callbackBz:                 validCallbackBz,
			depositFoundAfterCallback:  true,
			depositStatusAfterCallback: types.DETOKENIZATION_FAILED,
		},
		{
			name:          "invalid callback args",
			ackStatus:     icacallbackstypes.AckResponseStatus_SUCCESS,
			callbackBz:    []byte{1, 2, 3},
			expectedError: "unable to unmarshal detokenize callback",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			// Store the LSM deposit with status IN_PROGRESS
			s.App.StakeibcKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

			// Call the Detokenize callback
			ackResponse := &icacallbackstypes.AcknowledgementResponse{
				Status: tc.ackStatus,
			}
			err := keeper.DetokenizeCallback(s.App.StakeibcKeeper, s.Ctx, channeltypes.Packet{}, ackResponse, tc.callbackBz)

			// If the test case is supposed to error, check the error message
			if tc.expectedError != "" {
				s.Require().ErrorContains(err, tc.expectedError)
			} else {
				// Otherwise, confirm there is no error
				s.Require().NoError(err, "no error expected during callback")

				// Check the deposit and status after the callback
				deposit, actualFound := s.App.StakeibcKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
				s.Require().Equal(tc.depositFoundAfterCallback, actualFound, "deposit found after callback")

				if tc.depositFoundAfterCallback {
					s.Require().Equal(tc.depositStatusAfterCallback.String(), deposit.Status.String(), "deposit status")
				}
			}
		})
	}
}
