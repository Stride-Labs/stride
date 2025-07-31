package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	icacallbackstypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v27/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v27/x/records/types"
	"github.com/Stride-Labs/stride/v27/x/stakeibc/types"
)

type DetokenizeCallbackTestCase struct {
	callbackBz                  []byte
	expectedValidatorDelegation int64
	expectedTotalDelegation     int64
	ackResponse                 *icacallbacktypes.AcknowledgementResponse
}

// Helper function to setup the detokenize ICA callback test
// Returns the serialized callback args which will be an input parameter
// to the callback
func (s *KeeperTestSuite) SetupTestDetokenizeCallback(ackStatus icacallbacktypes.AckResponseStatus) DetokenizeCallbackTestCase {
	stakeAmount := sdkmath.NewInt(1000)
	detokenizedAmount := sdkmath.NewInt(999) // mimics SDK rounding bug
	initialValidatorDelegation := sdkmath.NewInt(5000)
	initialTotalDelegation := sdkmath.NewInt(10000)

	expectedValidatorDelegation := int64(5999) // 5000 + 999
	expectedTotalDelegation := int64(10999)    // 10000 + 999

	// Store host zone with validator
	hostZone := types.HostZone{
		ChainId:          HostChainId,
		TotalDelegations: initialTotalDelegation,
		Validators: []*types.Validator{{
			Address:                     ValAddress,
			Delegation:                  initialValidatorDelegation,
			DelegationChangesInProgress: 1,
		}},
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

	// Store the LSMDeposit record with status DETOKENIZATION_IN_PROGRESS
	deposit := recordstypes.LSMTokenDeposit{
		ChainId:          HostChainId,
		Denom:            LSMTokenBaseDenom,
		Status:           recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS,
		ValidatorAddress: ValAddress,
		Amount:           stakeAmount,
	}
	s.App.RecordsKeeper.SetLSMTokenDeposit(s.Ctx, deposit)

	// Return the deposit as callback args
	callbackBz, err := proto.Marshal(&types.DetokenizeSharesCallback{
		Deposit: &deposit,
	})
	s.Require().NoError(err, "no error expected when marshalling callback args")

	// Message response includes the redeemed amount
	detokenizeResponse := types.MsgRedeemTokensForSharesResponse{
		Amount: sdk.NewCoin(Atom, detokenizedAmount),
	}
	detokenizeResponseBz, err := proto.Marshal(&detokenizeResponse)
	s.Require().NoError(err, "no error expected when marshaling detokenize response")

	// Build the ack response with the detokenized amount and ack status
	ackResponse := &icacallbacktypes.AcknowledgementResponse{
		Status:       ackStatus,
		MsgResponses: [][]byte{detokenizeResponseBz},
	}

	return DetokenizeCallbackTestCase{
		callbackBz:                  callbackBz,
		expectedValidatorDelegation: expectedValidatorDelegation,
		expectedTotalDelegation:     expectedTotalDelegation,
		ackResponse:                 ackResponse,
	}
}

func (s *KeeperTestSuite) TestDetokenizeCallback_Successful() {
	tc := s.SetupTestDetokenizeCallback(icacallbackstypes.AckResponseStatus_SUCCESS)

	// Call the callback with a successful response
	err := s.App.StakeibcKeeper.DetokenizeCallback(s.Ctx, channeltypes.Packet{}, tc.ackResponse, tc.callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// Check that the deposit was removed
	_, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().False(found, "deposit should have been removed")

	// Check that the delegation was updated
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(tc.expectedTotalDelegation, hostZone.TotalDelegations.Int64(), "host zone total delegation")
	s.Require().Equal(tc.expectedValidatorDelegation, hostZone.Validators[0].Delegation.Int64(), "validator delegation")

	// Check that the number of delegations in progress was decremented
	s.Require().Equal(0, int(hostZone.Validators[0].DelegationChangesInProgress), "delegation change in progress")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_InvalidCallbackArgs() {
	tc := s.SetupTestDetokenizeCallback(icacallbackstypes.AckResponseStatus_SUCCESS)

	// Call the callback with a successful ack, but invalid callback args
	invalidCallbackArgs := []byte{1, 2, 3}
	err := s.App.StakeibcKeeper.DetokenizeCallback(s.Ctx, channeltypes.Packet{}, tc.ackResponse, invalidCallbackArgs)
	s.Require().ErrorContains(err, "unable to unmarshal detokenize callback")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_HostNotFound() {
	tc := s.SetupTestDetokenizeCallback(icacallbackstypes.AckResponseStatus_SUCCESS)

	// Call the callback with a host zone that does not exist - it should fail
	invalidCallbackArgs, err := proto.Marshal(&types.DetokenizeSharesCallback{
		Deposit: &recordstypes.LSMTokenDeposit{
			ChainId: "fake_chain",
		},
	})
	s.Require().NoError(err, "no error expected when marshalling callback data")

	err = s.App.StakeibcKeeper.DetokenizeCallback(s.Ctx, channeltypes.Packet{}, tc.ackResponse, invalidCallbackArgs)
	s.Require().ErrorContains(err, "Host zone not found")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_AckTimeout() {
	tc := s.SetupTestDetokenizeCallback(icacallbackstypes.AckResponseStatus_TIMEOUT)

	// Call the callback with a timed-out response
	err := s.App.StakeibcKeeper.DetokenizeCallback(s.Ctx, channeltypes.Packet{}, tc.ackResponse, tc.callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// The deposit should still be there in status IN_PROGRESS
	deposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().True(found, "deposit should not have been removed")
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_IN_PROGRESS.String(), deposit.Status.String(), "deposit status")

	// Check that the number of delegations in progress was decremented
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(0, int(hostZone.Validators[0].DelegationChangesInProgress), "delegation change in progress")
}

func (s *KeeperTestSuite) TestDetokenizeCallback_AckFailure() {
	tc := s.SetupTestDetokenizeCallback(icacallbackstypes.AckResponseStatus_FAILURE)

	// Call the callback with an ack-failure response
	err := s.App.StakeibcKeeper.DetokenizeCallback(s.Ctx, channeltypes.Packet{}, tc.ackResponse, tc.callbackBz)
	s.Require().NoError(err, "no error expected during callback")

	// The deposit status should be FAILED
	deposit, found := s.App.RecordsKeeper.GetLSMTokenDeposit(s.Ctx, HostChainId, LSMTokenBaseDenom)
	s.Require().True(found, "deposit should not have been removed")
	s.Require().Equal(recordstypes.LSMTokenDeposit_DETOKENIZATION_FAILED.String(), deposit.Status.String(), "deposit status")

	// Check that the number of delegations in progress was decremented
	hostZone, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, HostChainId)
	s.Require().True(found, "host zone should have been found")
	s.Require().Equal(0, int(hostZone.Validators[0].DelegationChangesInProgress), "delegation change in progress")
}
