package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) BeginBlocker(ctx sdk.Context) {
	// Check invariants

	// Check redemption rate is within safety bounds
	if err := k.CheckRedemptionRateExceedsBounds(ctx); err != nil {
		k.Logger(ctx).Error(err.Error())
		// If not, halt the zone
		k.HaltZone(ctx)
	}
}
