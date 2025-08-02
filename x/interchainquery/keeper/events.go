package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v28/x/interchainquery/types"
)

// Emits an event when a ICQ response is submitted
func EmitEventQueryResponse(ctx sdk.Context, query types.Query) {
	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
		),
		sdk.NewEvent(
			types.EventTypeQueryResponse,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(types.AttributeKeyQueryId, query.Id),
			sdk.NewAttribute(types.AttributeKeyChainId, query.ChainId),
		),
	})
}
