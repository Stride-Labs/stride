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

	routes := ms.Keeper.GetAllTradeRoutes(ctx)
	for _, route := range routes {
		if route.HostDenomOnHostZone == msg.HostDenom && route.RewardDenomOnRewardZone == msg.RewardDenom {
			tradeConfig := types.TradeConfig{
				PoolId: msg.PoolId,

				SwapPrice:            sdk.ZeroDec(),
				PriceUpdateTimestamp: 0,

				MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(msg.MaxAllowedSwapLossRate),
				MinSwapAmount:          sdkmath.NewIntFromUint64(msg.MinSwapAmount),
				MaxSwapAmount:          sdkmath.NewIntFromUint64(msg.MaxSwapAmount),
			}

			route.TradeConfig = tradeConfig
			ms.Keeper.SetTradeRoute(ctx, route)
		}
		// if no matching trade route was found for the given host-denom and reward-denom... do nothing
	}

	return &types.MsgUpdateTradeRouteResponse{}, nil
}
