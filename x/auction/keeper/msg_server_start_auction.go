package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// Get the name of the target pool, if the end-block is already before the current blockchain height, then we can start a new auction

func (k msgServer) StartAuction(goCtx context.Context, msg *types.MsgStartAuction) (*types.MsgStartAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	auctionPools := k.Keeper.GetAllAuctionPool(ctx)
	for _, pool := range auctionPools {
		// find which pool the incoming message is targeting
		if pool.GetPoolName() == msg.GetPoolName() {
			now := cast.ToUint64(ctx.BlockHeight())
			if pool.LastBlock < now {
				// TODO: get auctionDuration as well as auction type from a config
				k.Keeper.StartNewAuction(ctx, pool, 150)

			}
		}
	}

	return &types.MsgStartAuctionResponse{}, nil
}
