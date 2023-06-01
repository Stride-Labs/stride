package stakeibc

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"

	"github.com/Stride-Labs/stride/v9/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of stakeibc module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper, ak types.AccountKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// Iterate over all host zones and verify redemption rate
	for _, hz := range k.GetAllHostZone(ctx) {
		rrSafe, err := k.IsRedemptionRateWithinSafetyBounds(ctx, hz)
		if !rrSafe {
			hz.Halted = true
			k.SetHostZone(ctx, hz)

			// set rate limit on stAsset
			stDenom := types.StAssetDenomFromHostZoneDenom(hz.HostDenom)
			k.RatelimitKeeper.AddDenomToBlacklist(ctx, stDenom)

			k.Logger(ctx).Error(fmt.Sprintf("[INVARIANT BROKEN!!!] %s's RR is %s. ERR: %v", hz.GetChainId(), hz.RedemptionRate.String(), err.Error()))
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeHostZoneHalt,
					sdk.NewAttribute(types.AttributeKeyHostZone, hz.ChainId),
					sdk.NewAttribute(types.AttributeKeyRedemptionRate, hz.RedemptionRate.String()),
				),
			)
		}
	}
}
