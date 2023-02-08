package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/Stride-Labs/stride/v5/x/auction/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) AuctionPoolAll(goCtx context.Context, req *types.QueryAllAuctionPoolRequest) (*types.QueryAllAuctionPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var auctionPools []types.AuctionPool
	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	auctionPoolStore := prefix.NewStore(store, types.KeyPrefix(types.AuctionPoolKey))

	pageRes, err := query.Paginate(auctionPoolStore, req.Pagination, func(key []byte, value []byte) error {
		var auctionPool types.AuctionPool
		if err := k.cdc.Unmarshal(value, &auctionPool); err != nil {
			return err
		}

		auctionPools = append(auctionPools, auctionPool)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAuctionPoolResponse{AuctionPool: auctionPools, Pagination: pageRes}, nil
}

func (k Keeper) AuctionPool(goCtx context.Context, req *types.QueryGetAuctionPoolRequest) (*types.QueryGetAuctionPoolResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	auctionPool, found := k.GetAuctionPool(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetAuctionPoolResponse{AuctionPool: auctionPool}, nil
}
