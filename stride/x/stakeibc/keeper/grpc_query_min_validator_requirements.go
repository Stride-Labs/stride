package keeper

import (
	"context"

	"github.com/Stride-labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MinValidatorRequirements(c context.Context, req *types.QueryGetMinValidatorRequirementsRequest) (*types.QueryGetMinValidatorRequirementsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMinValidatorRequirements(ctx)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMinValidatorRequirementsResponse{MinValidatorRequirements: val}, nil
}
