package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

const DELEGATE = "delegate"
const CLAIM = "claim"
const UNDELEGATE = "undelegate"

// ICACallbacks wrapper struct for stakeibc keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, *channeltypes.Acknowledgement_Result, []byte) error

type ICACallbacks struct {
	k            Keeper
	icacallbacks map[string]ICACallback
}

var _ icacallbackstypes.ICACallbackHandler = ICACallbacks{}

func (k Keeper) ICACallbackHandler() ICACallbacks {
	return ICACallbacks{k, make(map[string]ICACallback)}
}

func (c ICACallbacks) CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ack *channeltypes.Acknowledgement_Result, args []byte) error {
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
	return c.
		AddICACallback(DELEGATE, ICACallback(DelegateCallback)).
		AddICACallback(CLAIM, ICACallback(ReinvestCallback)).
		AddICACallback(UNDELEGATE, ICACallback(UndelegateCallback))
}
