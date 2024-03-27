package keeper

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v21/x/interchainquery/types"
)

// EndBlocker of interchainquery module
func (k Keeper) EndBlocker(ctx sdk.Context) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	events := sdk.Events{}
	for _, query := range k.AllQueries(ctx) {
		if query.RequestSent {
			continue
		}

		k.Logger(ctx).Info(fmt.Sprintf("Interchainquery event emitted %s", query.Id))

		event := sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeyAction, types.AttributeValueQuery),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
			sdk.NewAttribute(types.AttributeKeyChainId, query.ChainId),
			sdk.NewAttribute(types.AttributeKeyConnectionId, query.ConnectionId),
			sdk.NewAttribute(types.AttributeKeyType, query.QueryType),
			sdk.NewAttribute(types.AttributeKeyHeight, "0"),
			sdk.NewAttribute(types.AttributeKeyRequest, hex.EncodeToString(query.RequestData)),
		)
		events = append(events, event)

		event.Type = "query_request"
		events = append(events, event)

		query.RequestSent = true
		k.SetQuery(ctx, query)
	}

	if len(events) > 0 {
		ctx.EventManager().EmitEvents(events)
	}
}
