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
func (k Keeper) TokenPrice(goCtx context.Context, req *types.QueryTokenPriceRequest) (*types.QueryTokenPriceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tokenPrice, err := k.GetTokenPrice(ctx, req.BaseDenom, req.QuoteDenom, req.PoolId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}

	return &types.QueryTokenPriceResponse{
		TokenPrice: k.TokenPriceToTokenPriceResponse(ctx, tokenPrice)[0],
	}, nil
}

// TokenPrices queries all token prices
func (k Keeper) TokenPrices(goCtx context.Context, req *types.QueryTokenPricesRequest) (*types.QueryTokenPricesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO impl paging

	prices := k.GetAllTokenPrices(ctx)

	return &types.QueryTokenPricesResponse{
		TokenPrices: k.TokenPriceToTokenPriceResponse(ctx, prices...),
	}, nil
}

// Params queries the oracle parameters
func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

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

// TokenPriceToTokenPriceResponse converts TokenPrices to TokenPriceResponses format
func (k Keeper) TokenPriceToTokenPriceResponse(ctx sdk.Context, tokenPrices ...types.TokenPrice) []types.TokenPriceResponse {
	responses := make([]types.TokenPriceResponse, len(tokenPrices))

	for i, tokenPrice := range tokenPrices {
		baseDenomUnwrapped := k.unwrapIBCDenom(ctx, tokenPrice.BaseDenom)
		quoteDenomUnwrapped := k.unwrapIBCDenom(ctx, tokenPrice.QuoteDenom)

		responses[i] = types.TokenPriceResponse{
			BaseDenomUnwrapped:  baseDenomUnwrapped,
			QuoteDenomUnwrapped: quoteDenomUnwrapped,
			BaseDenom:           tokenPrice.BaseDenom,
			QuoteDenom:          tokenPrice.QuoteDenom,
			BaseDenomDecimals:   tokenPrice.BaseDenomDecimals,
			QuoteDenomDecimals:  tokenPrice.QuoteDenomDecimals,
			OsmosisBaseDenom:    tokenPrice.OsmosisBaseDenom,
			OsmosisQuoteDenom:   tokenPrice.OsmosisQuoteDenom,
			OsmosisPoolId:       tokenPrice.OsmosisPoolId,
			SpotPrice:           tokenPrice.SpotPrice,
			UpdatedAt:           tokenPrice.UpdatedAt,
			QueryInProgress:     tokenPrice.QueryInProgress,
		}
	}

	return responses
}
