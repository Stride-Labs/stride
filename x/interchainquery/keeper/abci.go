package keeper

import (
	"time"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	RetryInterval = 25
)

// EndBlocker of interchainquery module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

}
