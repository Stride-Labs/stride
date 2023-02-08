package keeper

import (
	"context"

    "github.com/Stride-Labs/stride/v5/x/auction/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)


func (k msgServer) StartAuction(goCtx context.Context,  msg *types.MsgStartAuction) (*types.MsgStartAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

    // TODO: Handling the message
    _ = ctx

	return &types.MsgStartAuctionResponse{}, nil
}
