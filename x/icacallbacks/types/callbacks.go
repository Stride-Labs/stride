package types

import (
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ICACallbackHandler interface {
	AddICACallback(id string, fn interface{}) ICACallbackHandler
	RegisterICACallbacks() ICACallbackHandler
	CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ackResponse *AcknowledgementResponse, args []byte) error
	HasICACallback(id string) bool
}
