package stakeibc

import (
	"time"

	"github.com/Stride-Labs/stride/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of stakeibc module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	
	if ctx.BlockHeight()%int64(types.DelegateInterval) == 0 {
		icaStake := func(index int64, zoneInfo types.HostZone) (stop bool) {
			// Verify the delegation ICA is registered
			delegationIca := zoneInfo.GetDelegationAccount()
			if delegationIca == nil || delegationIca.Address == "" {
				k.Logger(ctx).Error("Zone %s is missing a delegation address!", zoneInfo.ChainId)
				return true
			}

			// TODO(TEST-46): Query process amount (unstaked balance) on host zone using ICQ
			processAmount := "1" + zoneInfo.BaseDenom
			amt, err := sdk.ParseCoinNormalized(processAmount)
			// Do we want to panic here? All unprocessed zones would also fail
			if err != nil {
				panic(err)
			}
			err = k.DelegateOnHost(ctx, zoneInfo, amt)
			if err != nil {
				k.Logger(ctx).Error("Did not stake %s on %s", processAmount, zoneInfo.ChainId)
				return true
			} else {
				k.Logger(ctx).Info("Successfully staked %s on %s", processAmount, zoneInfo.ChainId)
			}
			return false
		}

		// Iterate the zones and apply icaStake
		k.IterateHostZones(ctx, icaStake)
	}
}
