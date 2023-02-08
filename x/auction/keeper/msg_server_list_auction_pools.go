package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
    "github.com/Stride-Labs/stride/v5/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Perhaps the incoming message to list pools should specify which coin type via the zone id, unsure if this is part of the context already
func (k msgServer) ListAuctionPools(goCtx context.Context,  msg *types.MsgListAuctionPools) (*types.MsgListAuctionPoolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	auctionPools := k.GetAllAuctionPool(ctx)

	return &types.MsgListAuctionPoolsResponse{pools: auctionPools}, nil
}
