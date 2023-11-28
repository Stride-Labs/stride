package keeper

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (ms msgServer) DeleteTradeRoute(goCtx context.Context, msg *types.MsgDeleteTradeRoute) (*types.MsgDeleteTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ms.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", ms.authority, msg.Authority)
	}

	routes := ms.Keeper.GetAllTradeRoutes(ctx)
	for _, route := range routes {
		if route.HostDenomOnHostZone == msg.HostDenom && route.RewardDenomOnRewardZone == msg.RewardDenom {
			ms.Keeper.RemoveTradeRoute(ctx, route.RewardDenomOnHostZone, route.HostDenomOnHostZone)
		}
		// if no matching trade route was found for the given host-denom and reward-denom... do nothing
	}

	return &types.MsgDeleteTradeRouteResponse{}, nil
}
