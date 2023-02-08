package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k Keeper) EpochTrackerAll(c context.Context, req *types.QueryAllEpochTrackerRequest) (*types.QueryAllEpochTrackerResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var epochTrackers []types.EpochTracker
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	epochTrackerStore := prefix.NewStore(store, types.KeyPrefix(types.EpochTrackerKeyPrefix))

	pageRes, err := query.Paginate(epochTrackerStore, req.Pagination, func(key []byte, value []byte) error {
		var epochTracker types.EpochTracker
		if err := k.cdc.Unmarshal(value, &epochTracker); err != nil {
			return err
		}

		epochTrackers = append(epochTrackers, epochTracker)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllEpochTrackerResponse{EpochTracker: epochTrackers, Pagination: pageRes}, nil
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
