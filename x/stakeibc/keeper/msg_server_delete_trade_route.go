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

	_, found := ms.Keeper.GetTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)
	if !found {
		return nil, errorsmod.Wrapf(types.ErrTradeRouteNotFound,
			"no trade route for rewardDenom %s and hostDenom %s", msg.RewardDenom, msg.HostDenom)
	}

	ms.Keeper.RemoveTradeRoute(ctx, msg.RewardDenom, msg.HostDenom)

	return &types.MsgDeleteTradeRouteResponse{}, nil
}
