package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Stride-Labs/stride/v12/x/stakeibc/types"
)

func (k Keeper) ConsumerParams(c context.Context, req *types.QueryConsumerParams) (*types.QueryConsumerParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	params := types.ConsumerParams(k.ConsumerKeeper.GetConsumerParams(ctx))

	return &types.QueryConsumerParamsResponse{Params: params}, nil
}
