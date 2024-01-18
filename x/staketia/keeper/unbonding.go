package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO [sttia]
func (k Keeper) RedeemStake(ctx sdk.Context, staker string, stTokenAmount sdkmath.Int) error {
	return nil
}

// TODO [sttia]
func (k Keeper) PrepareUndelegation(ctx sdk.Context, epochNumber uint64) error {
	return nil
}

// TODO [sttia]
func (k Keeper) CheckUnbondingFinished(ctx sdk.Context) error {
	return nil
}

// TODO [sttia]
func (k Keeper) DistributeClaims(ctx sdk.Context) error {
	return nil
}
