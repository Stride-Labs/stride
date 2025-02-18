package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/strdburner/types"
)

var _ types.QueryServer = Keeper{}

// Auction queries the auction info for a specific token
func (k Keeper) StrdBurnerAddress(goCtx context.Context, req *types.QueryStrdBurnerAddressRequest) (*types.QueryStrdBurnerAddressResponse, error) {
	return &types.QueryStrdBurnerAddressResponse{
		Address: k.GetStrdBurnerAddress().String(),
	}, nil
}

func (k Keeper) TotalStrdBurned(goCtx context.Context, req *types.QueryTotalStrdBurnedRequest) (*types.QueryTotalStrdBurnedResponse, error) {
	return &types.QueryTotalStrdBurnedResponse{
		TotalBurned: k.GetTotalStrdBurned(sdk.UnwrapSDKContext(goCtx)),
	}, nil
}
