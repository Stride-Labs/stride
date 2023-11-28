package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ms msgServer) UpdateTradeRoute(goCtx context.Context, msg *types.MsgUpdateTradeRoute) (*types.MsgUpdateTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	route, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteNotFound,
			"no trade route for rewardDenom %s and hostDenom %s", msg.RewardDenom, msg.HostDenom)
	}

	updatedConfig := types.TradeConfig{
		PoolId: msg.PoolId,

		SwapPrice:            sdk.ZeroDec(),
		PriceUpdateTimestamp: 0,

		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(msg.MaxAllowedSwapLossRate),
		MinSwapAmount:          sdkmath.NewIntFromUint64(msg.MinSwapAmount),
		MaxSwapAmount:          sdkmath.NewIntFromUint64(msg.MaxSwapAmount),
	}

	route.TradeConfig = updatedConfig
	ms.Keeper.SetTradeRoute(ctx, route)

	return &types.MsgUpdateTradeRouteResponse{}, nil
}
