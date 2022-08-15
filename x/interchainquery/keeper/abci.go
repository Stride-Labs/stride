package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/x/interchainquery/types"
)

// EndBlocker of interchainquery module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)
	_ = k.Logger(ctx)
	events := sdk.Events{}
	// emit events for periodic queries
	k.IterateQueries(ctx, func(_ int64, queryInfo types.Query) (stop bool) {
		if queryInfo.LastHeight.Equal(sdk.ZeroInt()) || queryInfo.LastHeight.Add(queryInfo.Period).Equal(sdk.NewInt(ctx.BlockHeight())) {
			k.Logger(ctx).Info(fmt.Sprintf("Interchainquery event emitted %s", queryInfo.Id))
			event := sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
				sdk.NewAttribute(types.AttributeKeyQueryId, queryInfo.Id),
				sdk.NewAttribute(types.AttributeKeyChainId, queryInfo.ChainId),
				sdk.NewAttribute(types.AttributeKeyConnectionId, queryInfo.ConnectionId),
				sdk.NewAttribute(types.AttributeKeyType, queryInfo.QueryType),
				// TODO: add height to request type
				sdk.NewAttribute(types.AttributeKeyHeight, "0"),
				sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(queryInfo.Request)),
			)

			events = append(events, event)
			queryInfo.LastHeight = sdk.NewInt(ctx.BlockHeight())
			k.SetQuery(ctx, queryInfo)

		}
		return false
	})

	if len(events) > 0 {
		ctx.EventManager().EmitEvents(events)
	}

	k.IterateDatapoints(ctx, func(_ int64, dp types.DataPoint) bool {
		q, found := k.GetQuery(ctx, dp.Id)
		if !found {
			// query was removed; delete datapoint
			k.DeleteDatapoint(ctx, dp.Id)
		} else {
			ttl, err := cast.ToInt64E(q.Ttl)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Error casting ttl to int64, err: %s", err.Error()))
				return false
			}
			lh, err := cast.ToInt64E(dp.LocalHeight)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Error casting local height to int64, err: %s", err.Error()))
				return false
			}
			if lh+ttl > ctx.BlockHeader().Height {
				// gc old data
				k.DeleteDatapoint(ctx, dp.Id)
			}
		}

		return false

	})
}
