package types

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
)

type GeneralAuction interface {
	CreateNew(ctx sdk.Context, properties interface{})
	CheckBlock(ctx sdk.Context)
	ResolveAuction(ctx sdk.Context)
	GetStatus() AuctionState
}

func (a *Auction) GetAuction() GeneralAuction {
	switch a.GetAlgorithm() {
	case AuctionType_ASCENDING:
		return a.GetAscendingAuction()
	case AuctionType_DESCENDING:
		return a.GetDescendingAuction()
	case AuctionType_SEALEDBID:
		return a.GetSealedBidAuction()
	}
	return nil
}

func (a *AscendingAuction) CreateNew(ctx sdk.Context, properties interface{}) {
	return
}

func (a *AscendingAuction) CheckBlock(ctx sdk.Context) {
	return
}

func (a *AscendingAuction) ResolveAuction(ctx sdk.Context) {
	a.Status = AuctionState_COMPLETE
	return
}

func (a *DescendingAuction) CreateNew(ctx sdk.Context, properties interface{}) {
	return
}

func (a *DescendingAuction) CheckBlock(ctx sdk.Context) {
	return
}

func (a *DescendingAuction) ResolveAuction(ctx sdk.Context) {
	a.Status = AuctionState_COMPLETE
	return
}

func (a *SealedBidAuction) CreateNew(ctx sdk.Context, properties interface{}) {
	props := properties.(*SealedBidAuctionProperties)
	a.AuctionProperties = props

	//a.Supply = params.GetSupply()
	//a.AuctionDuration = params.GetSealedAuctionDuration()
	//a.RevealDuration = params.GetSealedRevealDuration()

	now := cast.ToUint64(ctx.BlockHeight())
	a.FirstBlock = now
	a.LastBlock = a.FirstBlock + props.GetAuctionDuration()
	a.RevealBlock = a.LastBlock + props.GetRevealDuration()

	a.SealedBids = []*SealedBid{}
	a.RevealedBids = []*OpenBid{}

	a.Status = AuctionState_RUNNING
	ctx.Logger().Info(fmt.Sprintf("[auction] New auction starting, state RUNNING at block %d", now))
	ctx.Logger().Info(fmt.Sprintf("[auction] New auction properties: %+v", *a))
}

func (a *SealedBidAuction) CheckBlock(ctx sdk.Context) {
	now := cast.ToUint64(ctx.BlockHeight())

	if a != nil {
		// if the current block has now >= lastBlock+1 and Status is Running
		// then Status needs to be set to reveal
		if a.Status == AuctionState_RUNNING &&
			a.GetLastBlock()+1 <= now {
			a.Status = AuctionState_REVEAL
			ctx.Logger().Info(fmt.Sprintf("[auction] State Change from RUNNING to REVEAL at block %d", now))
		}
		// if the current block has now >= revealBlock+1 and Status is Reveal
		// then Status needs to be set to payout and we need to call resolve
		if a.Status == AuctionState_REVEAL &&
			a.GetRevealBlock()+1 <= now {
			a.Status = AuctionState_PAYOUT
			ctx.Logger().Info(fmt.Sprintf("[auction] State Change from REVEAL to PAYOUT at block %d", now))
			a.ResolveAuction(ctx)
		}
	}
}

func (a *SealedBidAuction) ResolveAuction(ctx sdk.Context) {
	ctx.Logger().Info(fmt.Sprintf("[auction]"))
	ctx.Logger().Info(fmt.Sprintf("[auction] Running resolve to determine payouts..."))

	redempRate := a.GetAuctionProperties().GetRedemptionRate() // If 0, ratio style price
	supply := a.GetAuctionProperties().GetSupply()

	min := func(i, j uint64) uint64 {
		if i < j {
			return i
		}
		return j
	}

	type row struct {
		price  uint64
		volume uint64
		bidder string
	}
	orders := []row{}

	for _, openBid := range a.GetRevealedBids() {
		address := openBid.GetAddress()
		bid := openBid.GetBid()
		for _, order := range bid.GetOrders() {
			ctx.Logger().Info(fmt.Sprintf("[auction] %s --> %+v", address, *order))
			orders = append(orders, row{order.Price, order.Volume, address})
		}
	}
	sort.SliceStable(orders, func(i, j int) bool {
		// this is a trick to avoid needing floats or doing division
		// price per volume is i-price/i-volume > j-price/j-volume so multiply denominators
		return orders[i].price*orders[j].volume > orders[j].price*orders[i].volume
	})
	ctx.Logger().Info(fmt.Sprintf("[auction] Orderbook sorted by desc price per volume is now %+v", orders))

	give := map[string]uint64{}
	take := map[string]uint64{}
	for _, o := range orders {
		if supply > 0 {
			delta := min(supply, o.volume)
			percentFilled := cast.ToFloat64(delta) / cast.ToFloat64(o.volume)
			supply -= delta
			take[o.bidder] += delta + uint64(cast.ToFloat64(o.price)*percentFilled)             // In bidDenom
			give[o.bidder] += cast.ToUint64(cast.ToFloat64(delta) * cast.ToFloat64(redempRate)) // in auctionDenom
		}
	}

	for bidder, _ := range take {
		ctx.Logger().Info(fmt.Sprintf("[auction] From %s take %d and give %d", bidder, take[bidder], give[bidder]))
	}
	ctx.Logger().Info(fmt.Sprintf("[auction] %d supply remains", supply))

	a.Status = AuctionState_COMPLETE
	ctx.Logger().Info(fmt.Sprintf("[auction] state now COMPLETE"))
}
