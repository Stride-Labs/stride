package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
)

var _ types.QueryServer = Keeper{}

// TokenPrice queries the current price for a specific token
func (k Keeper) TokenPrice(goCtx context.Context, req *types.QueryTokenPriceRequest) (*types.QueryTokenPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	tokenPrice := types.TokenPrice{
		BaseDenom:     req.BaseDenom,
		QuoteDenom:    req.QuoteDenom,
		OsmosisPoolId: req.PoolId,
	}
	price, err := k.GetTokenPrice(ctx, tokenPrice)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryTokenPriceResponse{
		TokenPrice: price,
	}, nil
}

// TokenPrices queries all token prices
func (k Keeper) TokenPrices(goCtx context.Context, req *types.QueryTokenPricesRequest) (*types.QueryTokenPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	prices := k.GetAllTokenPrices(ctx)

	return &types.QueryTokenPricesResponse{
		TokenPrices: prices,
	}, nil
}

// Params queries the oracle parameters
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)

	return &types.QueryParamsResponse{
		Params: params,
	}, nil
}

// TokenPriceForQuoteDenom queries the exchange rate between two tokens
func (k Keeper) TokenPriceForQuoteDenom(goCtx context.Context, req *types.QueryTokenPriceForQuoteDenomRequest) (*types.QueryTokenPriceForQuoteDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	price, err := k.GetTokenPriceForQuoteDenom(ctx, req.BaseDenom, req.QuoteDenom)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryTokenPriceForQuoteDenomResponse{
		Price: price,
	}, nil
}
