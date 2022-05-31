package stakeibc

import (
	"time"

	"github.com/Stride-Labs/stride/x/stakeibc/keeper"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BeginBlocker of stakeibc module
func BeginBlocker(ctx sdk.Context, k keeper.Keeper, bk types.BankKeeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	
	if ctx.BlockHeight()%int64(types.DelegateInterval) == 0 {
		// Assume ICA is registered
		// Should we call iterate registered zones here?
		// If so, function called on each zone should:
		// ICA stake the current unstaked balance on the HostZone
	}

}
