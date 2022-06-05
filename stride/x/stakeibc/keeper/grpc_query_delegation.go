package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Delegation(c context.Context, req *types.QueryGetDelegationRequest) (*types.QueryGetDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetDelegation(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetDelegationResponse{Delegation: val}, nil
}
