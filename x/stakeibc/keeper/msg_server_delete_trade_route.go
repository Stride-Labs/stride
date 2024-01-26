package keeper

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Gov tx to remove a trade route
//
// Example proposal:
//
//		{
//		   "title": "Remove a new trade route for host chain X",
//		   "metadata": "Remove a new trade route for host chain X",
//		   "summary": "Remove a new trade route for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgDeleteTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//				 "reward_denom": "rewardToken",
//				 "host_denom": "hostToken
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
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
