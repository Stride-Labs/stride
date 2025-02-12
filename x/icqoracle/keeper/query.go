package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
)

var _ types.QueryServer = Keeper{}

// TokenPrice queries the current price for a specific token
func (k Keeper) TokenPrice(goCtx context.Context, req *types.QueryTokenPriceRequest) (*types.TokenPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tokenPrice, err := k.GetTokenPrice(ctx, req.BaseDenom, req.QuoteDenom, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.TokenPriceResponse{
		BaseDenomUnwrapped:  k.unwrapIBCDenom(ctx, tokenPrice.BaseDenom),
		QuoteDenomUnwrapped: k.unwrapIBCDenom(ctx, tokenPrice.QuoteDenom),
		TokenPrice:          tokenPrice,
	}, nil
}

// TokenPrices queries all token prices
func (k Keeper) TokenPrices(goCtx context.Context, req *types.QueryTokenPricesRequest) (*types.QueryTokenPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO impl paging

	responses := []types.TokenPriceResponse{}
	for _, tokenPrice := range k.GetAllTokenPrices(ctx) {
		responses = append(responses, types.TokenPriceResponse{
			BaseDenomUnwrapped:  k.unwrapIBCDenom(ctx, tokenPrice.BaseDenom),
			QuoteDenomUnwrapped: k.unwrapIBCDenom(ctx, tokenPrice.QuoteDenom),
			TokenPrice:          tokenPrice,
		})
	}

	return &types.QueryTokenPricesResponse{TokenPrices: responses}, nil
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

func (k Keeper) unwrapIBCDenom(ctx sdk.Context, denom string) string {
	if !strings.HasPrefix(denom, "ibc/") {
		return denom
	}

	hash, err := ibctransfertypes.ParseHexHash(strings.TrimPrefix(denom, "ibc/"))
	if err == nil {
		denomTrace, found := k.ibcTransferKeeper.GetDenomTrace(ctx, hash)
		if found {
			return denomTrace.BaseDenom
		}
	}
	return denom
}
