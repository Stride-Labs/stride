package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

func (k Keeper) NextHubUnbonding(c context.Context, req *types.QueryNextHubUnbonding) (*types.QueryNextHubUnbondingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	// query the next hub unbonding keeper
	k.GetNextHubUnbonding(sdk.UnwrapSDKContext(c))

	return &types.QueryNextHubUnbondingResponse{}, nil
}
