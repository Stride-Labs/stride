package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v17/x/staketia/types"
)

// TODO [sttia]
func (k Keeper) UpdateRedemptionRate(ctx sdk.Context) {

}

// Checks whether the redemption rate has exceeded the inner or outer safety bounds
func (k Keeper) CheckRedemptionRateExceedsBounds(ctx sdk.Context) error {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}
	redemptionRate := hostZone.RedemptionRate

	// Validate the safety bounds (e.g. that the inner is inside the outer)
	if err := k.ValidateRedemptionRateBoundsInitalized(hostZone); err != nil {
		return err
	}

	// Check if the redemption rate is outside the outer bounds
	if redemptionRate.LT(hostZone.MinRedemptionRate) || redemptionRate.GT(hostZone.MaxRedemptionRate) {
		return types.ErrRedemptionRateOutsideSafetyBounds.Wrapf("redemption rate outside outer safety bounds")
	}

	// Check if it's outside the inner bounds
	if redemptionRate.LT(hostZone.MinInnerRedemptionRate) || redemptionRate.GT(hostZone.MaxInnerRedemptionRate) {
		return types.ErrRedemptionRateOutsideSafetyBounds.Wrapf("redemption rate outside inner safety bounds")
	}

	return nil
}

// Verify the redemption rate bounds are set properly on the host zone
func (k Keeper) ValidateRedemptionRateBoundsInitalized(hostZone types.HostZone) error {
	// Validate outer bounds are set
	if hostZone.MinRedemptionRate.IsNil() || !hostZone.MinRedemptionRate.IsPositive() {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("min outer redemption rate bound not set")
	}
	if hostZone.MaxRedemptionRate.IsNil() || !hostZone.MaxRedemptionRate.IsPositive() {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("max outer redemption rate bound not set")
	}

	// Validate inner bounds set
	if hostZone.MinInnerRedemptionRate.IsNil() || !hostZone.MinInnerRedemptionRate.IsPositive() {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("min inner redemption rate bound not set")
	}
	if hostZone.MaxInnerRedemptionRate.IsNil() || !hostZone.MaxInnerRedemptionRate.IsPositive() {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("max inner redemption rate bound not set")
	}

	// Validate inner bounds are within outer bounds
	if hostZone.MinInnerRedemptionRate.LT(hostZone.MinRedemptionRate) {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("min inner redemption rate bound outside of min outer bound")
	}
	if hostZone.MaxInnerRedemptionRate.GT(hostZone.MaxRedemptionRate) {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("max inner redemption rate bound outside of max outer bound")
	}
	if hostZone.MinInnerRedemptionRate.GT(hostZone.MaxInnerRedemptionRate) {
		return types.ErrInvalidRedemptionRateBounds.Wrapf("min inner redemption rate greater than max inner bound")
	}

	return nil
}
