package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

func (k Keeper) EpochTrackerAll(c context.Context, req *types.QueryAllEpochTrackerRequest) (*types.QueryAllEpochTrackerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	epochTrackers := k.GetAllEpochTracker(ctx)

	return &types.QueryAllEpochTrackerResponse{EpochTracker: epochTrackers}, nil
}

func (k Keeper) EpochTracker(c context.Context, req *types.QueryGetEpochTrackerRequest) (*types.QueryGetEpochTrackerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetEpochTracker(
		ctx,
		req.EpochIdentifier,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetEpochTrackerResponse{EpochTracker: val}, nil
}
