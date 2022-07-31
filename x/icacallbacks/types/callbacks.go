package types

import (
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ICACallbackHandler interface {
	AddCallback(id string, fn interface{}) ICACallbackHandler
	RegisterCallbacks() ICACallbackHandler
	Call(ctx sdk.Context, id string, packet channeltypes.Packet, ack []byte, args []byte) error
	Has(id string) bool
}
