package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
)

func (k msgServer) LSMLiquidStake(goCtx context.Context, msg *types.MsgLSMLiquidStake) (*types.MsgLSMLiquidStakeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	_ = ctx
	// TODO [LSM]

	return &types.MsgLSMLiquidStakeResponse{}, nil
}

func (k Keeper) MintAndSendLSMTokens(ctx sdk.Context, lsmLiquidStake types.LSMLiquidStake) error {
	// TODO [LSM]
	return nil
}
