package keeper

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v24/x/icqoracle/types"
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

// AddTokenPrice adds a token to price tracking
func (ms msgServer) AddTokenPrice(goCtx context.Context, msg *types.MsgAddTokenPrice) (*types.MsgAddTokenPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO check admin

	tokenPrice := types.TokenPrice{
		BaseDenom:  msg.BaseDenom,
		QuoteDenom: msg.QuoteDenom,
		UpdatedAt:  time.Time{},
		Price:      sdkmath.LegacyZeroDec(),
	}

	err := ms.Keeper.SetTokenPrice(ctx, tokenPrice)
	if err != nil {
		return nil, err
	}

	return &types.MsgAddTokenPriceResponse{}, nil
}

// RemoveTokenPrice removes a token from price tracking
func (ms msgServer) RemoveTokenPrice(goCtx context.Context, msg *types.MsgRemoveTokenPrice) (*types.MsgRemoveTokenPriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO check admin

	ms.Keeper.RemoveTokenPrice(ctx, msg.BaseDenom, msg.QuoteDenom)

	return &types.MsgRemoveTokenPriceResponse{}, nil
}
