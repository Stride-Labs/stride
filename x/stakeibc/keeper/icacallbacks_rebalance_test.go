package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	_ "github.com/stretchr/testify/suite"

	icacallbacktypes "github.com/Stride-Labs/stride/v8/x/icacallbacks/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v8/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

type RebalanceCallbackState struct {
	hostZone          types.HostZone
	initialValidators []*types.Validator
}

type RebalanceCallbackArgs struct {
	packet      channeltypes.Packet
	ackResponse *icacallbacktypes.AcknowledgementResponse
	args        []byte
}

type RebalanceCallbackTestCase struct {
	initialState RebalanceCallbackState
	validArgs    RebalanceCallbackArgs
}

func (s *KeeperTestSuite) SetupRebalanceCallback(delegationType types.DelegationType) RebalanceCallbackTestCase {
	rebalanceValidatorsTestCase := s.SetupRebalanceValidators()

	packet := channeltypes.Packet{}
	ackResponse := icacallbacktypes.AcknowledgementResponse{Status: icacallbacktypes.AckResponseStatus_SUCCESS}
	callbackArgs := types.RebalanceCallback{
		HostZoneId: HostChainId,
		Rebalancings: []*types.Rebalancing{
			{
				SrcValidator: "stride_VAL3",
				DstValidator: "stride_VAL1",
				Amt:          sdkmath.NewInt(104),
			},
			{
				SrcValidator: "stride_VAL4",
				DstValidator: "stride_VAL1",
				Amt:          sdkmath.NewInt(13),
			},
		},
		DelegationType: delegationType,
	}
	args, err := s.App.StakeibcKeeper.MarshalRebalanceCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return RebalanceCallbackTestCase{
		initialState: RebalanceCallbackState{
			hostZone:          rebalanceValidatorsTestCase.hostZone,
			initialValidators: rebalanceValidatorsTestCase.initialValidators,
		},
		validArgs: RebalanceCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        args,
		},
	}
}

func (s *KeeperTestSuite) TestRebalanceCallback_Balanced_Successful() {
	tc := s.SetupRebalanceCallback(types.DelegationType_BALANCED)

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().NoError(err, "rebalance callback succeeded")

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")

	validators := hz.GetValidators()
	s.Require().Len(validators, 5, "host zone has 5 validators")

	// TODO: Improve these tests
	// These expected values are hard coded - and you have to reference a separate file to see where they come from
	s.Require().Equal(int64(217), validators[0].BalancedDelegation.Int64(), "validator 1 stake")
	s.Require().Equal(int64(500), validators[1].BalancedDelegation.Int64(), "validator 2 stake")
	s.Require().Equal(int64(96), validators[2].BalancedDelegation.Int64(), "validator 3 stake")
	s.Require().Equal(int64(387), validators[3].BalancedDelegation.Int64(), "validator 4 stake")
	s.Require().Equal(int64(400), validators[4].BalancedDelegation.Int64(), "validator 5 stake")
}

func (s *KeeperTestSuite) TestRebalanceCallback_Unbalanced_Successful() {
	tc := s.SetupRebalanceCallback(types.DelegationType_UNBALANCED)

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().NoError(err, "rebalance callback succeeded")

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")

	validators := hz.GetValidators()
	s.Require().Len(validators, 5, "host zone has 5 validators")

	s.Require().Equal(int64(317), validators[0].UnbalancedDelegation.Int64(), "validator 1 stake")
	s.Require().Equal(int64(1000), validators[1].UnbalancedDelegation.Int64(), "validator 2 stake")
	s.Require().Equal(int64(296), validators[2].UnbalancedDelegation.Int64(), "validator 3 stake")
	s.Require().Equal(int64(787), validators[3].UnbalancedDelegation.Int64(), "validator 4 stake")
	s.Require().Equal(int64(800), validators[4].UnbalancedDelegation.Int64(), "validator 5 stake")
}

func (s *KeeperTestSuite) checkDelegationStateIfCallbackFailed() {
	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")

	validators := hz.GetValidators()
	s.Require().Len(validators, 5, "host zone has 5 validators")

	s.Require().Equal(int64(100), validators[0].BalancedDelegation.Int64(), "validator 1 balanced stake")
	s.Require().Equal(int64(500), validators[1].BalancedDelegation.Int64(), "validator 2 balanced stake")
	s.Require().Equal(int64(200), validators[2].BalancedDelegation.Int64(), "validator 3 balanced stake")
	s.Require().Equal(int64(400), validators[3].BalancedDelegation.Int64(), "validator 4 balanced stake")
	s.Require().Equal(int64(400), validators[4].BalancedDelegation.Int64(), "validator 5 balanced stake")

	s.Require().Equal(int64(200), validators[0].UnbalancedDelegation.Int64(), "validator 1 unbalanced stake")
	s.Require().Equal(int64(1000), validators[1].UnbalancedDelegation.Int64(), "validator 2 unbalanced stake")
	s.Require().Equal(int64(400), validators[2].UnbalancedDelegation.Int64(), "validator 3 unbalanced stake")
	s.Require().Equal(int64(800), validators[3].UnbalancedDelegation.Int64(), "validator 4 unbalanced stake")
	s.Require().Equal(int64(800), validators[4].UnbalancedDelegation.Int64(), "validator 5 unbalanced stake")
}

func (s *KeeperTestSuite) TestRebalanceCallback_Timeout() {
	tc := s.SetupRebalanceCallback(types.DelegationType_BALANCED)

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegationStateIfCallbackFailed()
}

func (s *KeeperTestSuite) TestRebalanceCallback_ErrorOnHost() {
	tc := s.SetupRebalanceCallback(types.DelegationType_BALANCED)

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegationStateIfCallbackFailed()
}

func (s *KeeperTestSuite) TestRebalanceCallback_WrongCallbackArgs() {
	tc := s.SetupRebalanceCallback(types.DelegationType_BALANCED)
	invalidArgs := tc.validArgs

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal rebalance callback args: unexpected EOF: unable to unmarshal data structure")
	s.checkDelegationStateIfCallbackFailed()
}

func (s *KeeperTestSuite) TestRebalanceCallback_WrongValidator() {
	tc := s.SetupRebalanceCallback(types.DelegationType_BALANCED)

	callbackArgs := types.RebalanceCallback{
		HostZoneId: HostChainId,
		Rebalancings: []*types.Rebalancing{
			{
				SrcValidator: "stride_VAL3",
				DstValidator: "stride_VAL1",
				Amt:          sdkmath.NewInt(104),
			},
			{
				SrcValidator: "stride_VAL4_WRONG",
				DstValidator: "stride_VAL1",
				Amt:          sdkmath.NewInt(13),
			},
		},
	}
	invalidArgsOne, err := s.App.StakeibcKeeper.MarshalRebalanceCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	callbackArgs = types.RebalanceCallback{
		HostZoneId: HostChainId,
		Rebalancings: []*types.Rebalancing{
			{
				SrcValidator: "stride_VAL3",
				DstValidator: "stride_VAL1_WRONG",
				Amt:          sdkmath.NewInt(104),
			},
			{
				SrcValidator: "stride_VAL4",
				DstValidator: "stride_VAL1",
				Amt:          sdkmath.NewInt(13),
			},
		},
	}
	invalidArgsTwo, err := s.App.StakeibcKeeper.MarshalRebalanceCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	err = stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidArgsOne)
	s.Require().ErrorContains(err, "source validator not found stride_VAL4_WRONG")
	s.checkDelegationStateIfCallbackFailed()

	err = stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, invalidArgsTwo)
	s.Require().ErrorContains(err, "validator not found stride_VAL1_WRONG: invalid request")
	s.checkDelegationStateIfCallbackFailed()
}
