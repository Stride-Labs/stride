package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v24/x/strdburner/types"
)

var _ types.QueryServer = Keeper{}

// Auction queries the auction info for a specific token
func (k Keeper) StrdBurnerAddress(goCtx context.Context, req *types.QueryStrdBurnerAddressRequest) (*types.QueryStrdBurnerAddressResponse, error) {
	return &types.QueryStrdBurnerAddressResponse{
		Address: k.GetStrdBurnerAddress().String(),
	}, nil
}
