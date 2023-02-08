package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/records/types"
)

func (k Keeper) UserRedemptionRecordAll(c context.Context, req *types.QueryAllUserRedemptionRecordRequest) (*types.QueryAllUserRedemptionRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var userRedemptionRecords []types.UserRedemptionRecord
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	userRedemptionRecordStore := prefix.NewStore(store, types.KeyPrefix(types.UserRedemptionRecordKey))

	pageRes, err := query.Paginate(userRedemptionRecordStore, req.Pagination, func(key []byte, value []byte) error {
		var userRedemptionRecord types.UserRedemptionRecord
		if err := k.Cdc.Unmarshal(value, &userRedemptionRecord); err != nil {
			return err
		}

		userRedemptionRecords = append(userRedemptionRecords, userRedemptionRecord)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllUserRedemptionRecordResponse{UserRedemptionRecord: userRedemptionRecords, Pagination: pageRes}, nil
}

func (k Keeper) UserRedemptionRecord(c context.Context, req *types.QueryGetUserRedemptionRecordRequest) (*types.QueryGetUserRedemptionRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	userRedemptionRecord, found := k.GetUserRedemptionRecord(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetUserRedemptionRecordResponse{UserRedemptionRecord: userRedemptionRecord}, nil
}
