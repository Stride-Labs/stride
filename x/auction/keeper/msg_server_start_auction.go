package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// Get the name of the target pool, if the end-block is already before the current blockchain height, then we can start a new auction

func (k msgServer) StartAuction(goCtx context.Context, msg *types.MsgStartAuction) (*types.MsgStartAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params, _ := k.Keeper.GetParams(ctx)
	now := cast.ToUint64(ctx.BlockHeight())

	ctx.Logger().Info(fmt.Sprintf("[auction] Request to start auction for pool %d at block %d", msg.GetPoolID(), now))

	auctionPools := k.Keeper.GetAllAuctionPool(ctx)
	for _, pool := range auctionPools {
		// find which pool the incoming message is targeting
		if pool.GetId() == msg.GetPoolID() {
			auction := pool.GetLatestAuction().GetAuction()
			if auction == nil || auction.GetStatus() == types.AuctionState_COMPLETE {
				ctx.Logger().Info(fmt.Sprintf("[auction] params auctionDuration %d and revealDuration %d", params.GetSealedAuctionDuration(), params.GetSealedRevealDuration()))
				k.Keeper.StartNewAuction(ctx, pool, params)
			} else {
				ctx.Logger().Info(fmt.Sprintf("[auction] Request to start auction for pool %d at block %d failed because already a running auction!", msg.GetPoolID(), now))
			}
		}
	}

	return &types.MsgStartAuctionResponse{}, nil
}
