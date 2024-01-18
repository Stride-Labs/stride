package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO [sttia]
func (k Keeper) LiquidStake(ctx sdk.Context, staker string, nativeAmount sdkmath.Int) error {
	return nil
}

// TODO [sttia]
func (k Keeper) PrepareDelegation(ctx sdk.Context, epochNumber uint64) error {
	return nil
}
