package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/interchainquery/types"
)

// EndBlocker of interchainquery module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)
	_ = k.Logger(ctx)
	events := sdk.Events{}
	// emit events for periodic queries
	k.IterateQueries(ctx, func(_ int64, query types.Query) (stop bool) {
		if !query.RequestSent {
			k.Logger(ctx).Info(fmt.Sprintf("Interchainquery event emitted %s", query.Id))
			// QUESTION: Do we need to emit this event twice?
			event := sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
				sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
				sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
				sdk.NewAttribute(types.AttributeKeyChainId, query.ChainId),
				sdk.NewAttribute(types.AttributeKeyConnectionId, query.ConnectionId),
				sdk.NewAttribute(types.AttributeKeyType, query.QueryType),
				sdk.NewAttribute(types.AttributeKeyHeight, "0"),
				sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(query.Request)),
			)
			events = append(events, event)

			event.Type = "query_request"
			events = append(events, event)

			query.RequestSent = true
			k.SetQuery(ctx, query)
		}
		return false
	})

	if len(events) > 0 {
		ctx.EventManager().EmitEvents(events)
	}
}
