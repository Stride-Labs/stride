package keeper

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochtypes "github.com/Stride-Labs/stride/v29/x/epochs/types"
	"github.com/Stride-Labs/stride/v29/x/stakeibc/types"
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

// TODO [cleanup]: Update to be CheckRedemptionRateWithinSafetyBound and only throw an error (instead of a bool)
// safety check: ensure the redemption rate is NOT below our min safety threshold && NOT above our max safety threshold on host zone
func (k Keeper) IsRedemptionRateWithinSafetyBounds(ctx sdk.Context, zone types.HostZone) (bool, error) {
	// Get the wide bounds
	minSafetyThreshold, maxSafetyThreshold := k.GetOuterSafetyBounds(ctx, zone)

	redemptionRate := zone.RedemptionRate

	if redemptionRate.LT(minSafetyThreshold) || redemptionRate.GT(maxSafetyThreshold) {
		return false, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds,
			"redemption rate %v is outside safety bounds [%v, %v]", redemptionRate, minSafetyThreshold, maxSafetyThreshold)
	}

	// Verify the redemption rate is within the inner safety bounds
	// The inner safety bounds should always be within the safety bounds, but
	// the redundancy above is cheap.
	// There is also one scenario where the outer bounds go within the inner bounds - if they're updated as part of a param change proposal.
	minInnerSafetyThreshold, maxInnerSafetyThreshold := k.GetInnerSafetyBounds(ctx, zone)
	if redemptionRate.LT(minInnerSafetyThreshold) || redemptionRate.GT(maxInnerSafetyThreshold) {
		return false, errorsmod.Wrapf(types.ErrRedemptionRateOutsideSafetyBounds,
			"redemption rate %v is outside inner safety bounds [%v, %v]", redemptionRate, minInnerSafetyThreshold, maxInnerSafetyThreshold)
	}

	return true, nil
}
