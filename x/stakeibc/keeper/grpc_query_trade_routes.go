package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

func (k Keeper) AllTradeRoutes(c context.Context, req *types.QueryAllTradeRoutes) (*types.QueryAllTradeRoutesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	routes := k.GetAllTradeRoutes(ctx)

	return &types.QueryAllTradeRoutesResponse{TradeRoutes: routes}, nil
}
