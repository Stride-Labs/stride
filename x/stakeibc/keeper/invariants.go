package keeper

// DONTCOVER

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		return "", false
	}
}

// TODO: Consider removing stride and day epochs completely and using a single hourly epoch
// Confirm the number of stride epochs in 1 day epoch
func (k Keeper) AssertStrideAndDayEpochRelationship(ctx sdk.Context) {
	strideEpoch, found := k.GetEpochTracker(ctx, epochtypes.STRIDE_EPOCH)
	if !found || strideEpoch.Duration == 0 {
		return
	}
	dayEpoch, found := k.GetEpochTracker(ctx, epochtypes.DAY_EPOCH)
	if !found || dayEpoch.Duration == 0 {
		return
	}
	if dayEpoch.Duration/strideEpoch.Duration != StrideEpochsPerDayEpoch {
		panic("The stride epoch must be 1/4th the length of the day epoch")
	}
}
