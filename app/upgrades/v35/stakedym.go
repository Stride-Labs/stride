package v35

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakedymkeeper "github.com/Stride-Labs/stride/v34/x/stakedym/keeper"
)

// stakedymMaxBoundMultiplier is how far above the current redemption rate the max bounds are widened.
const stakedymMaxBoundMultiplier = 2

// UnhaltStakedym widens stakedym's max redemption-rate bounds and clears its Halted flag (spec §2, §5).
// Stakedym is halted on mainnet because its rate crossed its 1.1 max bounds, and its BeginBlocker
// re-halts (and re-blacklists stadym) every block while the rate is out of bounds, so its unbonding
// records could never be claimed (GetUnhaltedHostZone). Both max bounds are raised to RedemptionRate x 2
// when below that, so the BeginBlocker check keeps passing; the min bounds are left alone. The
// blacklisted denom is removed by the rate-limit step that follows. A missing zone is logged and skipped.
func UnhaltStakedym(ctx sdk.Context, k stakedymkeeper.Keeper) {
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		ctx.Logger().Info(fmt.Sprintf("v35: stakedym host zone not found, skipping unhalt: %s", err))
		return
	}

	// Widen only when needed, so a bound already looser than this is never tightened
	widenedMax := hostZone.RedemptionRate.MulInt64(stakedymMaxBoundMultiplier)
	if hostZone.MaxRedemptionRate.LT(widenedMax) {
		hostZone.MaxRedemptionRate = widenedMax
	}
	if hostZone.MaxInnerRedemptionRate.LT(widenedMax) {
		hostZone.MaxInnerRedemptionRate = widenedMax
	}
	hostZone.Halted = false
	k.SetHostZone(ctx, hostZone)

	ctx.Logger().Info(fmt.Sprintf("v35: stakedym unhalted at rate %s with max bounds %s / %s",
		hostZone.RedemptionRate, hostZone.MaxRedemptionRate, hostZone.MaxInnerRedemptionRate))
}
