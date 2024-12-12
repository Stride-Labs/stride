package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/auction/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// PlaceBid places a bid to buy a token off an auction
func (ms msgServer) PlaceBid(goCtx context.Context, msg *types.MsgPlaceBid) (*types.MsgPlaceBidResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := ms.Keeper.PlaceBid(ctx, msg)
	if err != nil {
		return nil, err
	}

	return &types.MsgPlaceBidResponse{}, nil
}

// CreateAuction creates a new auction
func (ms msgServer) CreateAuction(goCtx context.Context, msg *types.MsgCreateAuction) (*types.MsgCreateAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO check admin

	_, err := ms.Keeper.GetAuction(ctx, msg.Denom)
	if err == nil {
		return nil, types.ErrAuctionAlreadyExists.Wrapf("auction for token '%s' already exists", msg.Denom)
	}

	auction := types.Auction{
		Denom:           msg.Denom,
		Enabled:         msg.Enabled,
		PriceMultiplier: msg.PriceMultiplier,
		MinBidAmount:    msg.MinBidAmount,
	}

	err = ms.Keeper.SetAuction(ctx, auction)
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateAuctionResponse{}, nil
}

// CreateAuction updates an existing auction
func (ms msgServer) UpdateAuction(goCtx context.Context, msg *types.MsgUpdateAuction) (*types.MsgUpdateAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO check admin

	auction, err := ms.Keeper.GetAuction(ctx, msg.Denom)
	if err != nil {
		return nil, types.ErrAuctionDoesntExist.Wrapf("cannot find auction for token '%s'", msg.Denom)
	}

	auction.Enabled = msg.Enabled
	auction.MinBidAmount = msg.MinBidAmount
	auction.PriceMultiplier = msg.PriceMultiplier

	err = ms.Keeper.SetAuction(ctx, auction)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateAuctionResponse{}, nil
}
