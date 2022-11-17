package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v2/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Validators(c context.Context, req *types.QueryGetValidatorsRequest) (*types.QueryGetValidatorsResponse, error) {
	if req == nil || req.ChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	hostZone, found := k.GetHostZone(ctx, req.ChainId)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetValidatorsResponse{Validators: hostZone.Validators}, nil
}
