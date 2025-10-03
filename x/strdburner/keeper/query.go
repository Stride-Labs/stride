package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/strdburner/types"
)

var _ types.QueryServer = Keeper{}

// Auction queries the auction info for a specific token
func (k Keeper) StrdBurnerAddress(goCtx context.Context, req *types.QueryStrdBurnerAddressRequest) (*types.QueryStrdBurnerAddressResponse, error) {
	return &types.QueryStrdBurnerAddressResponse{
		Address: k.GetStrdBurnerAddress().String(),
	}, nil
}

func (k Keeper) TotalStrdBurned(goCtx context.Context, req *types.QueryTotalStrdBurnedRequest) (*types.QueryTotalStrdBurnedResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryTotalStrdBurnedResponse{
		TotalBurned:     k.GetTotalStrdBurned(ctx),
		ProtocolBurned:  k.GetProtocolStrdBurned(ctx),
		TotalUserBurned: k.GetTotalUserStrdBurned(ctx),
	}, nil
}

func (k Keeper) StrdBurnedByAddress(goCtx context.Context, req *types.QueryStrdBurnedByAddressRequest) (*types.QueryStrdBurnedByAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	address, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	return &types.QueryStrdBurnedByAddressResponse{
		BurnedAmount: k.GetStrdBurnedByAddress(ctx, address),
	}, nil
}
