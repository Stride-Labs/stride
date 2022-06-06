package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) RedeemStake(goCtx context.Context, msg *types.MsgRedeemStake) (*types.MsgRedeemStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	_ = ctx

	// Implement!

	return &types.MsgRedeemStakeResponse{}, nil
}
