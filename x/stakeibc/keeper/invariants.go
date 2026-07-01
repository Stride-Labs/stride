package keeper

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v32/utils"
	epochtypes "github.com/Stride-Labs/stride/v32/x/epochs/types"
	"github.com/Stride-Labs/stride/v32/x/stakeibc/types"
)

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

// CheckDelegationRecordsConsistent verifies, for every active host zone, that TotalDelegations
// equals the sum of each validator's tracked Delegation. This is a cheap internal-bookkeeping
// consistency check that catches code paths which update one value without the other.
//
// It is log-only (never panics or mutates state) so it can run safely in the BeginBlocker.
// NOTE: it does NOT detect tracked-vs-on-chain drift (e.g. delegations that were undelegated on
// the host but whose ack was lost) - that requires an interchain query; see the delegator-shares
// / calibrate reconciliation for that class of drift.
func (k Keeper) CheckDelegationRecordsConsistent(ctx sdk.Context) (consistent bool) {
	consistent = true
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		sum := sdkmath.ZeroInt()
		for _, validator := range hostZone.Validators {
			if validator.Delegation.IsNil() {
				continue
			}
			sum = sum.Add(validator.Delegation)
		}
		if hostZone.TotalDelegations.IsNil() || !sum.Equal(hostZone.TotalDelegations) {
			consistent = false
			k.Logger(ctx).Error(utils.LogWithHostZone(hostZone.ChainId,
				"delegation record inconsistency: TotalDelegations=%v but sum(validator.Delegation)=%v",
				hostZone.TotalDelegations, sum))
		}
	}
	return consistent
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
