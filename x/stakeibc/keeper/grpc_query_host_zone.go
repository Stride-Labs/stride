package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

func (k Keeper) HostZoneAll(c context.Context, req *types.QueryAllHostZoneRequest) (*types.QueryAllHostZoneResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var hostZones []types.HostZone
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	hostZoneStore := prefix.NewStore(store, types.KeyPrefix(types.HostZoneKey))

	pageRes, err := query.Paginate(hostZoneStore, req.Pagination, func(key []byte, value []byte) error {
		var hostZone types.HostZone
		if err := k.cdc.Unmarshal(value, &hostZone); err != nil {
			return err
		}

		hostZones = append(hostZones, hostZone)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllHostZoneResponse{HostZone: hostZones, Pagination: pageRes}, nil
}

func (k Keeper) HostZone(c context.Context, req *types.QueryGetHostZoneRequest) (*types.QueryGetHostZoneResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	hostZone, found := k.GetHostZone(ctx, req.ChainId)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetHostZoneResponse{HostZone: hostZone}, nil
}
