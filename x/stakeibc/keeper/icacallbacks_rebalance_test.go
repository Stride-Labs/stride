package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	ibctesting "github.com/cosmos/ibc-go/v5/testing"
	_ "github.com/stretchr/testify/suite"

	epochtypes "github.com/Stride-Labs/stride/v9/x/epochs/types"
	icacallbacktypes "github.com/Stride-Labs/stride/v9/x/icacallbacks/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
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

func (s *KeeperTestSuite) SetupRebalanceCallback() RebalanceCallbackTestCase {
	// Setup IBC
	delegationIcaOwner := "GAIA.DELEGATION"
	s.CreateICAChannel(delegationIcaOwner)
	delegationAddr := s.IcaAddresses[delegationIcaOwner]

	// setup epochs
	epochNumber := uint64(1)
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochtypes.STRIDE_EPOCH,
		EpochNumber:        epochNumber,
		NextEpochStartTime: uint64(s.Coordinator.CurrentTime.UnixNano() + 30_000_000_000), // dictates timeouts
	}
	s.App.StakeibcKeeper.SetEpochTracker(s.Ctx, epochTracker)

	// define validators for host zone
	initialValidators := []*types.Validator{
		{
			Name:       "val1",
			Address:    "stride_VAL1",
			Weight:     100,
			Delegation: sdkmath.NewInt(100),
		},
		{
			Name:       "val2",
			Address:    "stride_VAL2",
			Weight:     500,
			Delegation: sdkmath.NewInt(500),
		},
		{
			Name:       "val3",
			Address:    "stride_VAL3",
			Weight:     200,
			Delegation: sdkmath.NewInt(200),
		},
		{
			Name:       "val4",
			Address:    "stride_VAL4",
			Weight:     400,
			Delegation: sdkmath.NewInt(400),
		},
		{
			Name:       "val5",
			Address:    "stride_VAL5",
			Weight:     400,
			Delegation: sdkmath.NewInt(400),
		},
	}

	// setup host zone
	hostZone := types.HostZone{
		ChainId:              "GAIA",
		Validators:           initialValidators,
		TotalDelegations:     sdkmath.NewInt(1000),
		ConnectionId:         ibctesting.FirstConnectionID,
		DelegationIcaAddress: delegationAddr,
		HostDenom:            "uatom",
	}
	s.App.StakeibcKeeper.SetHostZone(s.Ctx, hostZone)

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
	}
	args, err := s.App.StakeibcKeeper.MarshalRebalanceCallbackArgs(s.Ctx, callbackArgs)
	s.Require().NoError(err)

	return RebalanceCallbackTestCase{
		initialState: RebalanceCallbackState{
			hostZone:          hostZone,
			initialValidators: initialValidators,
		},
		validArgs: RebalanceCallbackArgs{
			packet:      packet,
			ackResponse: &ackResponse,
			args:        args,
		},
	}
}

func (s *KeeperTestSuite) TestRebalanceCallback_Successful() {
	tc := s.SetupRebalanceCallback()

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, tc.validArgs.packet, tc.validArgs.ackResponse, tc.validArgs.args)
	s.Require().NoError(err, "rebalance callback succeeded")

	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")

	validators := hz.GetValidators()
	s.Require().Len(validators, 5, "host zone has 5 validators")

	// TODO: Improve these tests
	// These expected values are hard coded - and you have to reference a separate file to see where they come from
	s.Require().Equal(int64(217), validators[0].Delegation.Int64(), "validator 1 stake")
	s.Require().Equal(int64(500), validators[1].Delegation.Int64(), "validator 2 stake")
	s.Require().Equal(int64(96), validators[2].Delegation.Int64(), "validator 3 stake")
	s.Require().Equal(int64(387), validators[3].Delegation.Int64(), "validator 4 stake")
	s.Require().Equal(int64(400), validators[4].Delegation.Int64(), "validator 5 stake")
}

func (s *KeeperTestSuite) checkDelegationStateIfCallbackFailed() {
	hz, found := s.App.StakeibcKeeper.GetHostZone(s.Ctx, "GAIA")
	s.Require().True(found, "host zone found")

	validators := hz.GetValidators()
	s.Require().Len(validators, 5, "host zone has 5 validators")

	s.Require().Equal(int64(100), validators[0].Delegation.Int64(), "validator 1 stake")
	s.Require().Equal(int64(500), validators[1].Delegation.Int64(), "validator 2 stake")
	s.Require().Equal(int64(200), validators[2].Delegation.Int64(), "validator 3 stake")
	s.Require().Equal(int64(400), validators[3].Delegation.Int64(), "validator 4 stake")
	s.Require().Equal(int64(400), validators[4].Delegation.Int64(), "validator 5 stake")
}

func (s *KeeperTestSuite) TestRebalanceCallback_Timeout() {
	tc := s.SetupRebalanceCallback()

	// Update the ack response to indicate a timeout
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_TIMEOUT

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegationStateIfCallbackFailed()
}

func (s *KeeperTestSuite) TestRebalanceCallback_ErrorOnHost() {
	tc := s.SetupRebalanceCallback()

	// an error ack means the tx failed on the host
	invalidArgs := tc.validArgs
	invalidArgs.ackResponse.Status = icacallbacktypes.AckResponseStatus_FAILURE

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidArgs.args)
	s.Require().NoError(err)
	s.checkDelegationStateIfCallbackFailed()
}

func (s *KeeperTestSuite) TestRebalanceCallback_WrongCallbackArgs() {
	tc := s.SetupRebalanceCallback()
	invalidArgs := tc.validArgs

	// random args should cause the callback to fail
	invalidCallbackArgs := []byte("random bytes")

	err := stakeibckeeper.RebalanceCallback(s.App.StakeibcKeeper, s.Ctx, invalidArgs.packet, invalidArgs.ackResponse, invalidCallbackArgs)
	s.Require().EqualError(err, "Unable to unmarshal rebalance callback args: unexpected EOF: unable to unmarshal data structure")
	s.checkDelegationStateIfCallbackFailed()
}

func (s *KeeperTestSuite) TestRebalanceCallback_WrongValidator() {
	tc := s.SetupRebalanceCallback()

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
