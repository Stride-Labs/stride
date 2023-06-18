package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v10/x/stakeibc/types"
)

func (k Keeper) ModuleAddress(goCtx context.Context, req *types.QueryModuleAddressRequest) (*types.QueryModuleAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := k.AccountKeeper.GetModuleAccount(ctx, req.Name).GetAddress().String()

	return &types.QueryModuleAddressResponse{Addr: addr}, nil
}
