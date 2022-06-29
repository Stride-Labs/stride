package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryCallbacks interface {
	AddCallback(id string, fn interface{}) QueryCallbacks
	RegisterCallbacks() QueryCallbacks
	Call(ctx sdk.Context, id string, args []byte, query Query) error
	Has(id string) bool
}
