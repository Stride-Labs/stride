package keeper

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/Stride-Labs/stride/v5/x/auction/types"
)

// Bidders send in the bid object with the original salt used to encrypt their sealed bid from earlier messages
// If their sealed bid hash matches the hash we generate then nothing has changed and their revealed bid is entered

func (k msgServer) RevealBid(goCtx context.Context, msg *types.MsgRevealBid) (*types.MsgRevealBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ctx.Logger().Info(fmt.Sprintf("[auction]"))
	ctx.Logger().Info(fmt.Sprintf("[auction] Reveal request for %+v", msg))

	auctionPools := k.Keeper.GetAllAuctionPool(ctx)
	for _, pool := range auctionPools {
		if pool.GetId() == msg.GetPoolID() {
			auction := pool.GetLatestAuction()
			if auction != nil && auction.GetAlgorithm() == types.AuctionType_SEALEDBID {
				sealedAuction := auction.GetAuction().(*types.SealedBidAuction)
				if sealedAuction != nil && sealedAuction.GetStatus() == types.AuctionState_REVEAL {

					creator := msg.GetCreator()
					sealedBids := sealedAuction.GetSealedBids()
					ctx.Logger().Info(fmt.Sprintf("[auction] Here are the sealedAuction %+v", pool.GetLatestAuction().GetSealedBidAuction()))
					ctx.Logger().Info(fmt.Sprintf("[auction] Here are the sealedbids map %+v", sealedBids))

					inSealedBids := false
					sealedValue := ""
					for _, sealedBid := range sealedBids {
						if sealedBid.GetAddress() == creator {
							inSealedBids = true
							sealedValue = sealedBid.GetHashedBid()
						}
					}

					if inSealedBids {
						//ctx.Logger().Info(fmt.Sprintf("[auction] Before the hash Bid is %s", msg.GetBid()+msg.GetSalt()))
						hash := hex.EncodeToString(crypto.Sha256([]byte(msg.GetBid() + msg.GetSalt())))

						if sealedValue == hash {
							ctx.Logger().Info(fmt.Sprintf("[auction] Hash Match! On hash %s", hash))

							var revealedBid types.Bid
							json.Unmarshal([]byte(msg.GetBid()), &revealedBid)
							ctx.Logger().Info(fmt.Sprintf("[auction] Bid revealed as %+v", revealedBid))
							// Update the unhashed bid into the data structure which is what resolve will use
							ob := types.OpenBid{creator, &revealedBid}
							sealedAuction.RevealedBids = append(sealedAuction.RevealedBids, &ob)
							k.Keeper.SetAuctionPool(ctx, pool)
							//}
						} else {
							ctx.Logger().Info(fmt.Sprintf("[auction] The hash doesn't match! %s != %s", sealedValue, hash))
						}
					} else {
						ctx.Logger().Info(fmt.Sprintf("[auction] No sealed bid for creator %s", creator))
					}
				} else {
					ctx.Logger().Info(fmt.Sprintf("[auction] This is not the reveal stage yet!"))
				}
			} else {
				ctx.Logger().Info(fmt.Sprintf("[auction] This is not a sealed bid auction!"))
			}
		}
	}

	return &types.MsgRevealBidResponse{}, nil
}
