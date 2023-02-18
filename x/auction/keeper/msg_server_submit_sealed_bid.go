package keeper

import (
	"context"
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// Get the name of the target pool, if the end-block is already before the current blockchain height, then we can start a new auction

func (k msgServer) SubmitSealedBid(goCtx context.Context, msg *types.MsgSubmitSealedBid) (*types.MsgSubmitSealedBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	now := cast.ToUint64(ctx.BlockHeight())

	ctx.Logger().Info(fmt.Sprintf("[auction]"))
	ctx.Logger().Info(fmt.Sprintf("[auction] Sealed bid incoming %s at block %d", msg.GetHashedBid(), now))

	auctionPools := k.Keeper.GetAllAuctionPool(ctx)
	for _, pool := range auctionPools {
		if pool.GetId() == msg.GetPoolID() {
			auction := pool.GetLatestAuction()
			if auction != nil && auction.GetAlgorithm() == types.AuctionType_SEALEDBID {
				sealedAuction := auction.GetAuction().(*types.SealedBidAuction)
				//ctx.Logger().Info(fmt.Sprintf("[auction] Sealed Auction %+v", sealedAuction))
				if sealedAuction != nil && sealedAuction.GetStatus() == types.AuctionState_RUNNING {
					creator := msg.GetCreator()
					sealedValue := msg.GetHashedBid()

					// Take in collateral which will be refunded at the end iff they reveal to avoid sybil attack
					// If they submit multiple bids from same address to overwrite, we should only take on the first one

					// check if they already have a sealed bid
					alreadySealedBidForCreator := false
					sealedAuction.GetSealedBids()
					for _, b := range sealedAuction.GetSealedBids() {
						if b.Address == creator {
							alreadySealedBidForCreator = true
						}
					}

					var collateralErr error
					if alreadySealedBidForCreator {
						collateralErr = nil
					} else {
						// Send the collateral to the pool address if this is first sealed message from bidder address
						amount := sealedAuction.GetAuctionProperties().Collateral
						colDenom := pool.GetPoolProperties().BidDenom
						coin := sdk.NewCoin(colDenom, sdkmath.NewIntFromUint64(amount))
						coins := sdk.NewCoins(coin)
						bidderAddress, _ := sdk.AccAddressFromBech32(creator)
						poolAddress, _ := sdk.AccAddressFromBech32(pool.GetPoolProperties().PoolAddress)

						if k.bankKeeper.BlockedAddr(poolAddress) {
							ctx.Logger().Info(fmt.Sprintf("[auction] pool address is blocked!"))
						}
						if k.bankKeeper.BlockedAddr(bidderAddress) {
							ctx.Logger().Info(fmt.Sprintf("[auction] bidder address is blocked!"))
						}

						ctx.Logger().Info(fmt.Sprintf("[auction] About to send collateral %+v %+v %+v", coin, bidderAddress, poolAddress))
						collateralErr = k.bankKeeper.SendCoins(ctx, bidderAddress, poolAddress, coins)
					}
					if collateralErr == nil { // either no need to send, or send was successful
						ctx.Logger().Info(fmt.Sprintf("[auction] Creator %s just set sealed bid %s", creator, sealedValue))
						sb := types.SealedBid{creator, sealedValue}
						sealedAuction.SealedBids = append(sealedAuction.SealedBids, &sb)
						ctx.Logger().Info(fmt.Sprintf("[auction] Here are the sealedbids %+v", pool.GetLatestAuction().GetSealedBidAuction().GetSealedBids()))
						k.Keeper.SetAuctionPool(ctx, pool)
					} else {
						// failed to get collateral
						ctx.Logger().Info(fmt.Sprintf("[auction] problem getting collateral %s", collateralErr.Error()))
					}

				} else {
					ctx.Logger().Info(fmt.Sprintf("[auction] This is not the bidding stage!"))
				}
			} else {
				ctx.Logger().Info(fmt.Sprintf("[auction] This is not a sealed bid auction! Cannot SubmitSealedBid!"))
			}
		}
	}

	return &types.MsgSubmitSealedBidResponse{}, nil
}
