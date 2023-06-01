package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (s *KeeperTestSuite) TestNextPacketSequenceQuery() {
	portId := "transfer"
	channelId := "channel-0"
	sequence := uint64(10)
	context := sdk.WrapSDKContext(s.Ctx)

	// Set a channel sequence
	s.App.IBCKeeper.ChannelKeeper.SetNextSequenceSend(s.Ctx, portId, channelId, sequence)

	// Test a successful query
	response, err := s.App.StakeibcKeeper.NextPacketSequence(context, &types.QueryGetNextPacketSequenceRequest{
		ChannelId: channelId,
		PortId:    portId,
	})
	s.Require().NoError(err)
	s.Require().Equal(sequence, response.Sequence)

	// Test querying a non-existent channel (should fail)
	_, err = s.App.StakeibcKeeper.NextPacketSequence(context, &types.QueryGetNextPacketSequenceRequest{
		ChannelId: "fake-channel",
		PortId:    portId,
	})
	s.Require().ErrorContains(err, "channel and port combination not found")
}
