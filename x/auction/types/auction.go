package types

import (
	"fmt"
	"sort"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
)

// Perhaps these functions should be moved to the keeper instead of the type
// Without access to the keeper the resolve function can't execute payouts

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

func min(i, j uint64) uint64 {
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

func (a *AscendingAuction) CreateNew(ctx sdk.Context, properties interface{}) {
	props := properties.(*AscendingAuctionProperties)
	a.AuctionProperties = props

	now := cast.ToUint64(ctx.BlockHeight())
	a.LastBlock = now + props.GetAuctionDuration()

	a.Bids = []*OpenBid{}

	a.Status = AuctionState_RUNNING
	ctx.Logger().Info(fmt.Sprintf("[auction] New ascending auction starting, state RUNNING at block %d", now))
	ctx.Logger().Info(fmt.Sprintf("[auction] New auction properties: %+v", *a))
}

func (a *AscendingAuction) CheckBlock(ctx sdk.Context) {
	now := cast.ToUint64(ctx.BlockHeight())

	if a != nil {
		if a.Status == AuctionState_RUNNING &&
			a.GetLastBlock()+1 <= now {
			a.Status = AuctionState_PAYOUT
			ctx.Logger().Info(fmt.Sprintf("[auction] State Change from REVEAL to PAYOUT at block %d", now))
			a.ResolveAuction(ctx)
		}
	}
}

func (a *AscendingAuction) ResolveAuction(ctx sdk.Context) {
	a.Status = AuctionState_COMPLETE
	ctx.Logger().Info(fmt.Sprintf("[auction]"))
	ctx.Logger().Info(fmt.Sprintf("[auction] Running resolve to determine payouts..."))

	redempRate := a.GetAuctionProperties().GetRedemptionRate() // If 0, ratio style price
	supply := a.GetAuctionProperties().GetSupply()
	orders := []row{}

	for _, openBid := range a.GetBids() {
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

func (a *DescendingAuction) CreateNew(ctx sdk.Context, properties interface{}) {
	props := properties.(*DescendingAuctionProperties)
	a.AuctionProperties = props

	a.CurrentSupply = props.Supply
	a.CurrentBid = props.StartingBid

	now := cast.ToUint64(ctx.BlockHeight())
	a.NextStep = now + props.BidStepDuration

	a.Bids = []*OpenBid{}

	a.Status = AuctionState_RUNNING
	ctx.Logger().Info(fmt.Sprintf("[auction] New descending auction starting, state RUNNING at block %d", now))
	ctx.Logger().Info(fmt.Sprintf("[auction] New auction properties: %+v", *a))
}

func (a *DescendingAuction) CheckBlock(ctx sdk.Context) {
	now := cast.ToUint64(ctx.BlockHeight())

	if a != nil {
		// If there are bids from the last block, immediately resolve them
		// If there is still supply, keep state Running, if no supply then Complete
		if a.Status == AuctionState_RUNNING &&
			len(a.GetBids()) > 0 {
			a.ResolveAuction(ctx)
			if a.GetCurrentSupply() > 0 {
				a.Status = AuctionState_RUNNING // technically no change here
				a.Bids = []*OpenBid{}           // empty out the bids which were resolved
			} else {
				a.Status = AuctionState_COMPLETE
			}
		}

		// if the current block is time for the next step,
		// update the current bid and time for next step
		if a.Status == AuctionState_RUNNING &&
			a.GetNextStep()+1 <= now {
			a.NextStep += a.GetAuctionProperties().GetBidStepDuration()
			a.CurrentBid -= a.GetAuctionProperties().GetBidStepSize()

			if a.CurrentBid < a.GetAuctionProperties().GetMinAllowedBid() ||
				a.CurrentBid > a.GetAuctionProperties().GetStartingBid() { // ie. uint wrapped around
				a.Status = AuctionState_COMPLETE
				ctx.Logger().Info(fmt.Sprintf("[auction] Completed auction because bid is too low, remaining supply %d", a.GetCurrentSupply()))
			} else {
				ctx.Logger().Info(fmt.Sprintf("[auction] Currently block %d, Bid step down to  %d, next step at %d", now, a.GetCurrentBid(), a.GetNextStep()))
			}
		}
	}
}

func (a *DescendingAuction) ResolveAuction(ctx sdk.Context) {
	// In descending auction, we immedaitely pay out any bids which came in
	// The price in the bid doesn't matter because the price is set by the auction itself
	ctx.Logger().Info(fmt.Sprintf("[auction] Running resolve to determine payouts on current bids..."))
	ctx.Logger().Info(fmt.Sprintf("[auction] Incoming price ignored, auction bid is at %d", a.CurrentBid))
	redempRate := a.GetAuctionProperties().GetRedemptionRate() // If 0, ratio style price
	supply := a.GetCurrentSupply()
	orders := []row{}

	for _, openBid := range a.GetBids() {
		address := openBid.GetAddress()
		bid := openBid.GetBid()
		for _, order := range bid.GetOrders() {
			// the price sent in doesn't matter, the auction has the price
			orders = append(orders, row{a.CurrentBid, order.Volume, address})
		}
	}

	give := map[string]uint64{}
	take := map[string]uint64{}
	for _, o := range orders {
		if supply > 0 {
			ctx.Logger().Info(fmt.Sprintf("[auction] order: price %d volume %d from %s", o.price, o.volume, o.bidder))
			delta := min(supply, o.volume)
			percentFilled := cast.ToFloat64(delta) / cast.ToFloat64(o.volume)
			supply -= delta
			take[o.bidder] += delta + uint64(cast.ToFloat64(o.price)*percentFilled) // In bidDenom
			ctx.Logger().Info(fmt.Sprintf("[auction] taking %d for redemption and %d for margin cost from %s", delta, uint64(cast.ToFloat64(o.price)*percentFilled), o.bidder))
			give[o.bidder] += cast.ToUint64(cast.ToFloat64(delta) * cast.ToFloat64(redempRate)) // in auctionDenom
		}
	}

	for bidder, _ := range take {
		ctx.Logger().Info(fmt.Sprintf("[auction] From %s take %d and give %d", bidder, take[bidder], give[bidder]))
		a.CurrentSupply = supply
	}
	ctx.Logger().Info(fmt.Sprintf("[auction] %d supply remains", supply))

}

func (a *SealedBidAuction) CreateNew(ctx sdk.Context, properties interface{}) {
	props := properties.(*SealedBidAuctionProperties)
	a.AuctionProperties = props

	now := cast.ToUint64(ctx.BlockHeight())
	a.FirstBlock = now
	a.LastBlock = a.FirstBlock + props.GetAuctionDuration()
	a.RevealBlock = a.LastBlock + props.GetRevealDuration()

	a.SealedBids = []*SealedBid{}
	a.RevealedBids = []*OpenBid{}

	a.Status = AuctionState_RUNNING
	ctx.Logger().Info(fmt.Sprintf("[auction] New sealed bid auction starting, state RUNNING at block %d", now))
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

	// Need access to the keeper -- should these methods be defined on the keeper instead of the auction types?

	//Use bank keeper to send the coins as result of the auction

	// Refund collateral -- if the address has a bid in the revealed bids table,
	// then they had collateral taken when they sealed bid and they later revealed successfully
	for _, rb := range a.GetRevealedBids() {
		amount := a.GetAuctionProperties().Collateral
		colDenom := a.GetPoolProperties().BidDenom
		coin := sdk.NewCoin(colDenom, sdkmath.NewIntFromUint64(amount))
		//coins := sdk.NewCoins(coin)
		bidderAddress, _ := sdk.AccAddressFromBech32(rb.Address)
		poolAddress, _ := sdk.AccAddressFromBech32(a.GetPoolProperties().PoolAddress)

		ctx.Logger().Info(fmt.Sprintf("[auction] About to refund collateral %+v %+v %+v", coin, bidderAddress, poolAddress))
		//collateralErr = k.bankKeeper.SendCoins(ctx, poolAddress, bidderAddress, coins)
	}

	a.Status = AuctionState_COMPLETE
	ctx.Logger().Info(fmt.Sprintf("[auction] state now COMPLETE"))
}
