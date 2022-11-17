package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v2/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ModuleAddress(goCtx context.Context, req *types.QueryModuleAddressRequest) (*types.QueryModuleAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := k.accountKeeper.GetModuleAccount(ctx, req.Name).GetAddress().String()

	return &types.QueryModuleAddressResponse{Addr: addr}, nil
}
