package keeper

import (
	"fmt"

	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	proto "github.com/gogo/protobuf/proto"

	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

// ___________________________________________________________________________________________________

// ICACallbacks wrapper struct for interchainstaking keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, []byte, []byte) error

type ICACallbacks struct {
	k         Keeper
	icacallbacks map[string]ICACallback
}

var _ icacallbackstypes.ICACallbackHandler = ICACallbacks{}

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
	a := c.AddICACallback("delegate", ICACallback(DelegateCallback))
	return a.(ICACallbacks)
}

// -----------------------------------
// ICACallback Handlers
// -----------------------------------

// ----------------------------------- Delegate Callback ----------------------------------- //
func (k Keeper) MarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback types.DelegateCallback) []byte {
	out, err := proto.Marshal(&delegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
	}
	return out
}

func (k Keeper) UnmarshalDelegateCallbackArgs(ctx sdk.Context, delegateCallback []byte) types.DelegateCallback {
	unmarshalledDelegateCallback := types.DelegateCallback{}
	if err := proto.Unmarshal(delegateCallback, &unmarshalledDelegateCallback); err != nil {
        k.Logger(ctx).Error(fmt.Sprintf("UnmarshalDelegateCallbackArgs %v", err.Error()))
	}
	return unmarshalledDelegateCallback
}

func DelegateCallback(k Keeper, ctx sdk.Context, packet channeltypes.Packet, ack []byte, args []byte) error {
	k.Logger(ctx).Info("DOGE DelegateCallback executing", "packet", packet, "ack", ack, "args", args)
	// deserialize the ack
	// deserialize the args
	delegateCallback := k.UnmarshalDelegateCallbackArgs(ctx, args)
	k.Logger(ctx).Info(fmt.Sprintf("DelegateCallback %v", delegateCallback))
	return nil
}
