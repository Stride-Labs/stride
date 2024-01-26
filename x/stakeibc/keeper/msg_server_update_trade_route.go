package keeper

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Gov tx to update the trade config of a trade route
//
// Example proposal:
//
//		{
//		   "title": "Update a the trade config for host chain X",
//		   "metadata": "Update a the trade config for host chain X",
//		   "summary": "Update a the trade config for host chain X",
//		   "messages":[
//		      {
//		         "@type": "/stride.stakeibc.MsgUpdateTradeRoute",
//		         "authority": "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl",
//
//				 "pool_id": 1,
//				 "max_allowed_swap_loss_rate": "0.05",
//				 "min_swap_amount": "10000000",
//				 "max_swap_amount": "1000000000"
//			  }
//		   ],
//		   "deposit": "2000000000ustrd"
//	   }
//
// >>> strided tx gov submit-proposal {proposal_file.json} --from wallet
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

	maxAllowedSwapLossRate := msg.MaxAllowedSwapLossRate
	if maxAllowedSwapLossRate == "" {
		maxAllowedSwapLossRate = DefaultMaxAllowedSwapLossRate
	}
	maxSwapAmount := msg.MaxSwapAmount
	if maxSwapAmount.IsZero() {
		maxSwapAmount = DefaultMaxSwapAmount
	}

	updatedConfig := types.TradeConfig{
		PoolId: msg.PoolId,

		SwapPrice:            sdk.ZeroDec(),
		PriceUpdateTimestamp: 0,

		MaxAllowedSwapLossRate: sdk.MustNewDecFromStr(maxAllowedSwapLossRate),
		MinSwapAmount:          msg.MinSwapAmount,
		MaxSwapAmount:          maxSwapAmount,
	}

	route.TradeConfig = updatedConfig
	ms.Keeper.SetTradeRoute(ctx, route)

	return &types.MsgUpdateTradeRouteResponse{}, nil
}
