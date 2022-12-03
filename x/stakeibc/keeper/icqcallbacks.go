package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	icqtypes "github.com/Stride-Labs/stride/v4/x/interchainquery/types"
)

const (
	ICQCallbackID_WithdrawalBalance = "withdrawalbalance"
	ICQCallbackID_Delegation        = "delegation"
	ICQCallbackID_Validator         = "validator"
)

// ICQCallbacks wrapper struct for stakeibc keeper
type ICQCallback func(Keeper, sdk.Context, []byte, icqtypes.Query) error

type ICQCallbacks struct {
	k         Keeper
	callbacks map[string]ICQCallback
}

var _ icqtypes.QueryCallbacks = ICQCallbacks{}

func (k Keeper) ICQCallbackHandler() ICQCallbacks {
	return ICQCallbacks{k, make(map[string]ICQCallback)}
}

func (c ICQCallbacks) CallICQCallback(ctx sdk.Context, id string, args []byte, query icqtypes.Query) error {
	return c.callbacks[id](c.k, ctx, args, query)
}

func (c ICQCallbacks) HasICQCallback(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c ICQCallbacks) AddICQCallback(id string, fn interface{}) icqtypes.QueryCallbacks {
	c.callbacks[id] = fn.(ICQCallback)
	return c
}

func (c ICQCallbacks) RegisterICQCallbacks() icqtypes.QueryCallbacks {
	return c.
		AddICQCallback(ICQCallbackID_WithdrawalBalance, ICQCallback(WithdrawalBalanceCallback)).
		AddICQCallback(ICQCallbackID_Delegation, ICQCallback(DelegatorSharesCallback)).
		AddICQCallback(ICQCallbackID_Validator, ICQCallback(ValidatorExchangeRateCallback))
}
