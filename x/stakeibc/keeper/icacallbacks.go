package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/v4/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

const (
	ICACallbackID_Delegate   = "delegate"
	ICACallbackID_Claim      = "claim"
	ICACallbackID_Undelegate = "undelegate"
	ICACallbackID_Reinvest   = "reinvest"
	ICACallbackID_Redemption = "redemption"
	ICACallbackID_Rebalance  = "rebalance"
)

// ICACallbacks wrapper struct for stakeibc keeper
type ICACallback func(Keeper, sdk.Context, channeltypes.Packet, *channeltypes.Acknowledgement, []byte) error

type ICACallbacks struct {
	k            Keeper
	icacallbacks map[string]ICACallback
}

var _ icacallbackstypes.ICACallbackHandler = ICACallbacks{}

func (k Keeper) ICACallbackHandler() ICACallbacks {
	return ICACallbacks{k, make(map[string]ICACallback)}
}

func (c ICACallbacks) CallICACallback(ctx sdk.Context, id string, packet channeltypes.Packet, ack *channeltypes.Acknowledgement, args []byte) error {
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
	a := c.
		AddICACallback(ICACallbackID_Delegate, ICACallback(DelegateCallback)).
		AddICACallback(ICACallbackID_Claim, ICACallback(ClaimCallback)).
		AddICACallback(ICACallbackID_Undelegate, ICACallback(UndelegateCallback)).
		AddICACallback(ICACallbackID_Reinvest, ICACallback(ReinvestCallback)).
		AddICACallback(ICACallbackID_Redemption, ICACallback(RedemptionCallback)).
		AddICACallback(ICACallbackID_Rebalance, ICACallback(RebalanceCallback))
	return a.(ICACallbacks)
}
