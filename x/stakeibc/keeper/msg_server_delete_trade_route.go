package keeper

import (
	"context"
	"fmt"

	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k msgServer) DeleteTradeRoute(goCtx context.Context, msg *types.MsgDeleteTradeRoute) (*types.MsgDeleteTradeRouteResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info(fmt.Sprintf("delete trade route: %s", msg.String()))

	// get our addresses, make sure they're valid
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "creator address is invalid: %s. err: %s", msg.Creator, err.Error())
	}

	routes := k.GetAllTradeRoutes(ctx)
	for _, route := range routes {
		if route.HostDenomOnHostZone == msg.HostDenom && route.RewardDenomOnRewardZone == msg.RewardDenom {
			k.RemoveTradeRoute(ctx, route.RewardDenomOnHostZone, route.HostDenomOnHostZone)
		}
		// if no matching trade route was found for the given host-denom and reward-denom... do nothing
	}

	k.Logger(ctx).Info(fmt.Sprintf("delete trade route: %s", msg.String()))
	return &types.MsgDeleteTradeRouteResponse{}, nil
}
