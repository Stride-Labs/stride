package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v2/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ICAAccount(c context.Context, req *types.QueryGetICAAccountRequest) (*types.QueryGetICAAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetICAAccount(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetICAAccountResponse{ICAAccount: val}, nil
}
