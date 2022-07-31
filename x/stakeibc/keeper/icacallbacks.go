package keeper

import (
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/icacallbacks/types"
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
)

// ___________________________________________________________________________________________________

// ICACallbacks wrapper struct for interchainstaking keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, []byte, []byte) error

type ICACallbacks struct {
	k         Keeper
	icacallbacks map[string]ICACallback
}

var _ types.ICACallbackHandler = ICACallbacks{}

func (k Keeper) ICACallbackHandler() ICACallbacks {
	return ICACallbacks{k, make(map[string]ICACallback)}
}

//callback handler
func (c ICACallbacks) CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ack []byte, args []byte) error {
	return c.icacallbacks[id](c.k, ctx, packet, ack, args)
}

func (c ICACallbacks) HasICACallback(id string) bool {
	_, found := c.icacallbacks[id]
	return found
}

func (c ICACallbacks) AddICACallback(id string, fn interface{}) icacallbackstypes.ICACallbackHandler {
	c.icacallbacks[id] = fn.(ICACallback)
	return c
}

func (c ICACallbacks) RegisterICACallbacks() icacallbackstypes.ICACallbackHandler {
	a := c.AddICACallback("samplecallback", ICACallback(SampleCallback))
	return a.(ICACallbacks)
}

// -----------------------------------
// ICACallback Handlers
// -----------------------------------

func SampleCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack []byte, args []byte) error {
	return nil
}
