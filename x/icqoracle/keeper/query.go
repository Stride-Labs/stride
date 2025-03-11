package keeper

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	"github.com/Stride-Labs/stride/v26/x/icqoracle/types"
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

	store := ctx.KVStore(k.storeKey)
	tokenPriceStore := prefix.NewStore(store, types.TokenPricePrefix)

	responses := []types.TokenPriceResponse{}
	pageRes, err := query.Paginate(tokenPriceStore, req.Pagination, func(key []byte, value []byte) error {
		var tokenPrice types.TokenPrice
		if err := k.cdc.Unmarshal(value, &tokenPrice); err != nil {
			return err
		}

		responses = append(responses, types.TokenPriceResponse{
			BaseDenomUnwrapped:  k.unwrapIBCDenom(ctx, tokenPrice.BaseDenom),
			QuoteDenomUnwrapped: k.unwrapIBCDenom(ctx, tokenPrice.QuoteDenom),
			TokenPrice:          tokenPrice,
		})
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryTokenPricesResponse{
		TokenPrices: responses,
		Pagination:  pageRes,
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
