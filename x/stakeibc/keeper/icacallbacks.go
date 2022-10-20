package keeper

import (
	icacallbackstypes "github.com/Stride-Labs/stride/x/icacallbacks/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
)

const (
	DELEGATE   = "delegate"
	CLAIM      = "claim"
	UNDELEGATE = "undelegate"
	REINVEST   = "reinvest"
	REDEMPTION = "redemption"
  REBALANCE = "rebalance"
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
		AddICACallback(DELEGATE, ICACallback(DelegateCallback)).
		AddICACallback(CLAIM, ICACallback(ClaimCallback)).
		AddICACallback(UNDELEGATE, ICACallback(UndelegateCallback)).
		AddICACallback(REINVEST, ICACallback(ReinvestCallback)).
		AddICACallback(REDEMPTION, ICACallback(RedemptionCallback)).
		AddICACallback(REBALANCE, ICACallback(RebalanceCallback))
	return a.(ICACallbacks)
}
