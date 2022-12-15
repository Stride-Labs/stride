package keeper

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/utils"
	"github.com/Stride-Labs/stride/v4/x/epochs/types"
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

		epochInfo.CurrentEpochStartHeight = ctx.BlockHeight()

		switch {
		case shouldInitialEpochStart:
			epochInfo = startInitialEpoch(epochInfo)
			logger.Info(fmt.Sprintf("initial %s epoch", epochInfo.Identifier))
		case shouldEpochStart:
			epochInfo = endEpoch(epochInfo)

			// Capitalize the epoch identifier for the logs
			epochAlias := strings.ToUpper(strings.ReplaceAll(epochInfo.Identifier, "_epoch", ""))
			logger.Info(utils.LogHeader("%s EPOCH %d", epochAlias, epochInfo.CurrentEpoch))
			logger.Info(utils.LogHeader("Epoch Start Time: %s", epochInfo.CurrentEpochStartTime))

			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeEpochEnd,
					sdk.NewAttribute(types.AttributeEpochNumber, strconv.FormatInt(epochInfo.CurrentEpoch, 10)),
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
				sdk.NewAttribute(types.AttributeEpochNumber, strconv.FormatInt(epochInfo.CurrentEpoch, 10)),
				sdk.NewAttribute(types.AttributeEpochStartTime, strconv.FormatInt(epochInfo.CurrentEpochStartTime.Unix(), 10)),
			),
		)

		return false
	})
}

func startInitialEpoch(epochInfo types.EpochInfo) types.EpochInfo {
	epochInfo.EpochCountingStarted = true
	epochInfo.CurrentEpoch = 1
	epochInfo.CurrentEpochStartTime = epochInfo.StartTime
	return epochInfo
}

func endEpoch(epochInfo types.EpochInfo) types.EpochInfo {
	epochInfo.CurrentEpoch++
	epochInfo.CurrentEpochStartTime = epochInfo.CurrentEpochStartTime.Add(epochInfo.Duration)
	return epochInfo
}
