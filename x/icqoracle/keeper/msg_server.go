package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v29/x/icqoracle/types"
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
		return nil, types.ErrTokenPriceAlreadyExists.Wrapf("token price BaseDenom='%s' QuoteDenom='%s' OsmosisPoolId='%d'", msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)
	}

	tokenPrice := types.TokenPrice{
		BaseDenom:         msg.BaseDenom,
		QuoteDenom:        msg.QuoteDenom,
		OsmosisPoolId:     msg.OsmosisPoolId,
		OsmosisBaseDenom:  msg.OsmosisBaseDenom,
		OsmosisQuoteDenom: msg.OsmosisQuoteDenom,
		LastRequestTime:   time.Time{},
		SpotPrice:         sdkmath.LegacyZeroDec(),
		QueryInProgress:   false,
	}
	ms.Keeper.SetTokenPrice(ctx, tokenPrice)

	return &types.MsgRegisterTokenPriceQueryResponse{}, nil
}

// RemoveTokenPriceQuery removes a token from price tracking
func (ms msgServer) RemoveTokenPriceQuery(goCtx context.Context, msg *types.MsgRemoveTokenPriceQuery) (*types.MsgRemoveTokenPriceQueryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ms.Keeper.RemoveTokenPrice(ctx, msg.BaseDenom, msg.QuoteDenom, msg.OsmosisPoolId)

	return &types.MsgRemoveTokenPriceQueryResponse{}, nil
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	k.SetParams(ctx, req.Params)

	return &types.MsgUpdateParamsResponse{}, nil
}
