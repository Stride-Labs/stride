package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/auction/types"
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

	_, err := ms.Keeper.GetAuction(ctx, msg.AuctionName)
	if err == nil {
		return nil, types.ErrAuctionAlreadyExists.Wrapf("auction with name '%s' already exists", msg.AuctionName)
	}

	auction := types.Auction{
		Type:                      msg.AuctionType,
		Name:                      msg.AuctionName,
		SellingDenom:              msg.SellingDenom,
		PaymentDenom:              msg.PaymentDenom,
		Enabled:                   msg.Enabled,
		MinPriceMultiplier:        msg.MinPriceMultiplier,
		MinBidAmount:              msg.MinBidAmount,
		Beneficiary:               msg.Beneficiary,
		TotalPaymentTokenReceived: math.ZeroInt(),
		TotalSellingTokenSold:     math.ZeroInt(),
	}
	ms.Keeper.SetAuction(ctx, &auction)

	return &types.MsgCreateAuctionResponse{}, nil
}

// CreateAuction updates an existing auction
func (ms msgServer) UpdateAuction(goCtx context.Context, msg *types.MsgUpdateAuction) (*types.MsgUpdateAuctionResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	auction, err := ms.Keeper.GetAuction(ctx, msg.AuctionName)
	if err != nil {
		return nil, types.ErrAuctionDoesntExist.Wrapf("cannot find auction with name '%s'", msg.AuctionName)
	}

	auction.Type = msg.AuctionType
	auction.Enabled = msg.Enabled
	auction.MinBidAmount = msg.MinBidAmount
	auction.MinPriceMultiplier = msg.MinPriceMultiplier
	auction.Beneficiary = msg.Beneficiary
	ms.Keeper.SetAuction(ctx, auction)

	return &types.MsgUpdateAuctionResponse{}, nil
}
