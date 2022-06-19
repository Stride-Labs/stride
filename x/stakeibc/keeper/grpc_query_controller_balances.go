package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ControllerBalancesAll(c context.Context, req *types.QueryAllControllerBalancesRequest) (*types.QueryAllControllerBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var controllerBalancess []types.ControllerBalances
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	controllerBalancesStore := prefix.NewStore(store, types.KeyPrefix(types.ControllerBalancesKeyPrefix))

	pageRes, err := query.Paginate(controllerBalancesStore, req.Pagination, func(key []byte, value []byte) error {
		var controllerBalances types.ControllerBalances
		if err := k.cdc.Unmarshal(value, &controllerBalances); err != nil {
			return err
		}

		controllerBalancess = append(controllerBalancess, controllerBalances)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllControllerBalancesResponse{ControllerBalances: controllerBalancess, Pagination: pageRes}, nil
}

func (k Keeper) ControllerBalances(c context.Context, req *types.QueryGetControllerBalancesRequest) (*types.QueryGetControllerBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetControllerBalances(
	    ctx,
	    req.Index,
        )
	if !found {
	    return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetControllerBalancesResponse{ControllerBalances: val}, nil
}