package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/utils"
)

func (k Keeper) HaltZone(ctx sdk.Context) {
	// Set the halted flag on the zone
	hostZone, err := k.GetHostZone(ctx)
	if err != nil {
		// No panic - we don't want to halt the chain! Just the zone.
		// log the error
		k.Logger(ctx).Error(fmt.Sprintf("Unable to get host zone: %s", err.Error()))
		return
	}
	hostZone.Halted = true
	k.SetHostZone(ctx, hostZone)

	// set rate limit on stAsset
	stDenom := utils.StAssetDenomFromHostZoneDenom(hostZone.NativeTokenDenom)
	k.ratelimitKeeper.AddDenomToBlacklist(ctx, stDenom)

	k.Logger(ctx).Error(fmt.Sprintf("[INVARIANT BROKEN!!!] %s's RR is %s.", hostZone.GetChainId(), hostZone.RedemptionRate.String()))

	EmitHaltZoneEvent(ctx, hostZone)
}
