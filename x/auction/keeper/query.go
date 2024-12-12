package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/auction/types"
)

var _ types.QueryServer = Keeper{}

// Auction queries the auction info for a specific token
func (k Keeper) Auction(goCtx context.Context, req *types.QueryAuctionRequest) (*types.QueryAuctionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	auction, err := k.GetAuction(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryAuctionResponse{
		Auction: auction,
	}, nil
}

// Auctions queries the auction info for a specific token
func (k Keeper) Auctions(goCtx context.Context, req *types.QueryAuctionsRequest) (*types.QueryAuctionsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	auctions := k.GetAllAuctions(ctx)

	// TODO impl paging

	return &types.QueryAuctionsResponse{
		Auctions: auctions,
	}, nil
}

// Stats queries the auction stats for a specific auction
func (k Keeper) Stats(goCtx context.Context, req *types.QueryStatsRequest) (*types.QueryStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	stats, err := k.GetStats(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryStatsResponse{
		Stats: stats,
	}, nil
}

// AllStats queries the auction stats for all auctions
func (k Keeper) AllStats(goCtx context.Context, req *types.QueryAllStatsRequest) (*types.QueryAllStatsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	allStats := k.GetAllStats(ctx)

	// TODO impl paging

	return &types.QueryAllStatsResponse{
		AllStats: allStats,
	}, nil
}
