package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// Get the name of the target pool, if the end-block is already before the current blockchain height, then we can start a new auction

func (k msgServer) SubmitOpenBid(goCtx context.Context, msg *types.MsgSubmitOpenBid) (*types.MsgSubmitOpenBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	now := cast.ToUint64(ctx.BlockHeight())

	ctx.Logger().Info(fmt.Sprintf("[auction]"))
	ctx.Logger().Info(fmt.Sprintf("[auction] Open bid incoming %s at block %d", msg.GetBid(), now))

	auctionPools := k.Keeper.GetAllAuctionPool(ctx)
	for _, pool := range auctionPools {
		if pool.GetId() == msg.GetPoolID() {
			auction := pool.GetLatestAuction()
			if auction != nil && auction.GetAlgorithm() == types.AuctionType_ASCENDING {
				ascendingAuction := auction.GetAuction().(*types.AscendingAuction)
				//ctx.Logger().Info(fmt.Sprintf("[auction] Sealed Auction %+v", sealedAuction))
				if ascendingAuction != nil && ascendingAuction.GetStatus() == types.AuctionState_RUNNING {
					// Should add check that bid is greater than previous bid
					timeToEnd := ascendingAuction.LastBlock - now
					if timeToEnd < ascendingAuction.GetAuctionProperties().ExtendDuration {
						ascendingAuction.LastBlock += ascendingAuction.GetAuctionProperties().ExtendDuration
						ctx.Logger().Info(fmt.Sprintf("[auction] New bid causes auction end to extend to %d", ascendingAuction.GetLastBlock()))
					}

					creator := msg.GetCreator()

					var openBid types.Bid
					json.Unmarshal([]byte(msg.GetBid()), &openBid)

					ctx.Logger().Info(fmt.Sprintf("[auction] Creator %s just set bid %s", creator, msg.GetBid()))
					ob := types.OpenBid{creator, &openBid}
					ascendingAuction.Bids = append(ascendingAuction.Bids, &ob)
					ctx.Logger().Info(fmt.Sprintf("[auction] Here are the bids %+v", pool.GetLatestAuction().GetAscendingAuction().GetBids()))
					k.Keeper.SetAuctionPool(ctx, pool)

				} else {
					ctx.Logger().Info(fmt.Sprintf("[auction] This is not the bidding stage!"))
				}
			} else if auction != nil && auction.GetAlgorithm() == types.AuctionType_DESCENDING {
				descendingAuction := auction.GetAuction().(*types.DescendingAuction)
				//ctx.Logger().Info(fmt.Sprintf("[auction] Sealed Auction %+v", sealedAuction))
				if descendingAuction != nil && descendingAuction.GetStatus() == types.AuctionState_RUNNING {
					creator := msg.GetCreator()

					var openBid types.Bid
					json.Unmarshal([]byte(msg.GetBid()), &openBid)

					ob := types.OpenBid{creator, &openBid}
					descendingAuction.Bids = append(descendingAuction.Bids, &ob)
					k.Keeper.SetAuctionPool(ctx, pool)

				} else {
					ctx.Logger().Info(fmt.Sprintf("[auction] This is not the bidding stage!"))
				}
			} else {
				ctx.Logger().Info(fmt.Sprintf("[auction] This is not a Ascending or Descending auction! Cannot SubmitOpenBid!"))
			}
		}
	}

	return &types.MsgSubmitOpenBidResponse{}, nil
}
