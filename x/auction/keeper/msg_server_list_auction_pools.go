package keeper

import (
	"context"

    "github.com/Stride-Labs/stride/v5/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) ListAuctionPools(goCtx context.Context,  msg *types.MsgListAuctionPools) (*types.MsgListAuctionPoolsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgListAuctionPoolsResponse{}, nil
}
