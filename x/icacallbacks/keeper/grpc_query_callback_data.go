package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/icacallbacks/types"
)

func (k Keeper) CallbackDataAll(c context.Context, req *types.QueryAllCallbackDataRequest) (*types.QueryAllCallbackDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var callbackDatas []types.CallbackData
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	callbackDataStore := prefix.NewStore(store, types.KeyPrefix(types.CallbackDataKeyPrefix))

	pageRes, err := query.Paginate(callbackDataStore, req.Pagination, func(key []byte, value []byte) error {
		var callbackData types.CallbackData
		if err := k.cdc.Unmarshal(value, &callbackData); err != nil {
			return err
		}

		callbackDatas = append(callbackDatas, callbackData)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllCallbackDataResponse{CallbackData: callbackDatas, Pagination: pageRes}, nil
}

func (k Keeper) CallbackData(c context.Context, req *types.QueryGetCallbackDataRequest) (*types.QueryGetCallbackDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetCallbackData(
		ctx,
		req.CallbackKey,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetCallbackDataResponse{CallbackData: val}, nil
}
