package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"
)

type GeneralAuction interface {
	CreateNew(ctx sdk.Context)
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

func (a *AscendingAuction) CreateNew(ctx sdk.Context) {
	return
}

func (a *AscendingAuction) CheckBlock(ctx sdk.Context) {
	return
}

func (a *AscendingAuction) ResolveAuction(ctx sdk.Context) {
	return
}

func (a *DescendingAuction) CreateNew(ctx sdk.Context) {
	return
}

func (a *DescendingAuction) CheckBlock(ctx sdk.Context) {
	return
}

func (a *DescendingAuction) ResolveAuction(ctx sdk.Context) {
	return
}

func (a *SealedBidAuction) CreateNew(ctx sdk.Context) {
	// reset the bids data structure
	return
}

func (a *SealedBidAuction) CheckBlock(ctx sdk.Context) {
	now := cast.ToUint64(ctx.BlockHeight())

	if a != nil {
		// if the current block has now >= lastBlock+1 and Status is Running
		// then Status needs to be set to reveal
		if a.Status == AuctionState_RUNNING &&
			a.GetLastBlock()+1 <= now {
			a.Status = AuctionState_REVEAL
		}
		// if the current block has now >= revealBlock+1 and Status is Reveal
		// then Status needs to be set to payout and we need to call resolve
		if a.Status == AuctionState_REVEAL &&
			a.GetRevealBlock()+1 <= now {
			a.Status = AuctionState_PAYOUT
			a.ResolveAuction(ctx)
		}

		return
	}
	return
}

func (a *SealedBidAuction) ResolveAuction(ctx sdk.Context) {
	a.Status = AuctionState_COMPLETE
	return
}
