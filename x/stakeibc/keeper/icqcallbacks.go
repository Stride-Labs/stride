package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	icqtypes "github.com/Stride-Labs/stride/x/interchainquery/types"
)

type Callback func(Keeper, sdk.Context, []byte, icqtypes.Query) error

type Callbacks struct {
	k         Keeper
	callbacks map[string]Callback
}

var _ icqtypes.QueryCallbacks = Callbacks{}

func (k Keeper) CallbackHandler() Callbacks {
	return Callbacks{k, make(map[string]Callback)}
}

func (c Callbacks) Call(ctx sdk.Context, id string, args []byte, query icqtypes.Query) error {
	return c.callbacks[id](c.k, ctx, args, query)
}

func (c Callbacks) Has(id string) bool {
	_, found := c.callbacks[id]
	return found
}

func (c Callbacks) AddCallback(id string, fn interface{}) icqtypes.QueryCallbacks {
	c.callbacks[id] = fn.(Callback)
	return c
}

func (c Callbacks) RegisterCallbacks() icqtypes.QueryCallbacks {
	return c.
		AddCallback("withdrawalbalance", Callback(WithdrawalBalanceCallback)).
		AddCallback("delegation", Callback(DelegatorSharesCallback)).
		AddCallback("validator", Callback(ValidatorExchangeRateCallback))
}
