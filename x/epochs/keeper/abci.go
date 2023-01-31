package keeper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v5/utils"
	"github.com/Stride-Labs/stride/v5/x/epochs/types"
)

// BeginBlocker of epochs module
func (k Keeper) BeginBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	logger := k.Logger(ctx)
	k.IterateEpochInfo(ctx, func(_ int64, epochInfo types.EpochInfo) (stop bool) {
		// Has it not started, and is the block time > initial epoch start time
		shouldInitialEpochStart := !epochInfo.EpochCountingStarted && !epochInfo.StartTime.After(ctx.BlockTime())

		epochEndTime := epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
		shouldEpochStart := ctx.BlockTime().After(epochEndTime) && !shouldInitialEpochStart && !epochInfo.StartTime.After(ctx.BlockTime())

		epochInfo.CurrentEpochStartHeight = sdkmath.NewInt(ctx.BlockHeight())

		switch {
		case shouldInitialEpochStart:
			epochInfo = startInitialEpoch(epochInfo)
			logger.Info(fmt.Sprintf("initial %s epoch", epochInfo.Identifier))
		case shouldEpochStart:
			epochInfo = endEpoch(epochInfo)

			// Capitalize the epoch identifier for the logs
			epochAlias := strings.ToUpper(strings.ReplaceAll(epochInfo.Identifier, "_epoch", ""))
			logger.Info(utils.LogHeader("%s EPOCH %s", epochAlias, epochInfo.CurrentEpoch.String()))
			logger.Info(utils.LogHeader("Epoch Start Time: %s", epochInfo.CurrentEpochStartTime))

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEpochEnd,
					sdk.NewAttribute(types.AttributeEpochNumber, epochInfo.CurrentEpoch.String()),
				),
			)
			k.AfterEpochEnd(ctx, epochInfo)
		default:
			// continue
			return false
		}

		k.SetEpochInfo(ctx, epochInfo)

		k.BeforeEpochStart(ctx, epochInfo)

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeEpochStart,
				sdk.NewAttribute(types.AttributeEpochNumber, epochInfo.CurrentEpoch.String()),
				sdk.NewAttribute(types.AttributeEpochStartTime, strconv.FormatInt(epochInfo.CurrentEpochStartTime.Unix(), 10)),
			),
		)

		return false
	})
}

func startInitialEpoch(epochInfo types.EpochInfo) types.EpochInfo {
	epochInfo.EpochCountingStarted = true
	epochInfo.CurrentEpoch = sdkmath.NewInt(1)
	epochInfo.CurrentEpochStartTime = epochInfo.StartTime
	return epochInfo
}

func endEpoch(epochInfo types.EpochInfo) types.EpochInfo {
	epochInfo.CurrentEpoch = epochInfo.CurrentEpoch.AddRaw(1)
	epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
	return epochInfo
}
