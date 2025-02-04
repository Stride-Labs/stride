package keeper

import (
	"context"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v25/x/icqoracle/types"
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

// RegisterTokenPriceQuery registers a token to price tracking
func (ms msgServer) RegisterTokenPriceQuery(goCtx context.Context, msg *types.MsgRegisterTokenPriceQuery) (*types.MsgRegisterTokenPriceQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_, err := ms.Keeper.GetTokenPrice(ctx, msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)
	if err == nil {
		return nil, types.ErrTokenPriceAlreadyExists.Wrapf("token price BaseDenom='%s' QuoteDenom='%s' OsmosisPoolId='%s'", msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)
	}

	tokenPrice := types.TokenPrice{
		BaseDenom:         msg.BaseDenom,
		QuoteDenom:        msg.QuoteDenom,
		OsmosisPoolId:     msg.OsmosisPoolId,
		OsmosisBaseDenom:  msg.OsmosisBaseDenom,
		OsmosisQuoteDenom: msg.OsmosisQuoteDenom,
		LastQueryTime:     time.Time{},
		SpotPrice:         sdkmath.LegacyZeroDec(),
		QueryInProgress:   false,
	}

	err = ms.Keeper.SetTokenPrice(ctx, tokenPrice)
	if err != nil {
		return nil, err
	}

	return &types.MsgRegisterTokenPriceQueryResponse{}, nil
}

// RemoveTokenPriceQuery removes a token from price tracking
func (ms msgServer) RemoveTokenPriceQuery(goCtx context.Context, msg *types.MsgRemoveTokenPriceQuery) (*types.MsgRemoveTokenPriceQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ms.Keeper.RemoveTokenPrice(ctx, msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)

	return &types.MsgRemoveTokenPriceQueryResponse{}, nil
}
