package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/staketia/types"
)

// Checks whether the redemption rate has exceeded the inner or outer safety bounds
// and returns an error if so
func (k Keeper) CheckRedemptionRateExceedsBounds(ctx sdk.Context) error {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		return err
	}
	redemptionRate := hostZone.RedemptionRate

	// Validate the safety bounds (e.g. that the inner is inside the outer)
	if err := hostZone.ValidateRedemptionRateBoundsInitalized(); err != nil {
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
