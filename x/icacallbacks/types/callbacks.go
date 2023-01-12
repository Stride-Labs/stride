package types

import (
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ICACallbackHandler interface {
	AddICACallback(id string, fn interface{}) ICACallbackHandler
	RegisterICACallbacks() ICACallbackHandler
	CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error
	HasICACallback(id string) bool
}
