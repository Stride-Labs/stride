package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k Keeper) NextPacketSequence(c context.Context, req *types.QueryGetNextPacketSequenceRequest) (*types.QueryGetNextPacketSequenceResponse, error) {
	if req == nil || req.ChannelId == "" || req.PortId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	sequence, found := k.IBCKeeper.ChannelKeeper.GetNextSequenceSend(ctx, req.PortId, req.ChannelId)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "channel and port combination not found")
	}

	return &types.QueryGetNextPacketSequenceResponse{Sequence: sequence}, nil
}
