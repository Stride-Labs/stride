package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryCallbacks interface {
	AddICQCallback(id string, fn interface{}) QueryCallbacks
	RegisterICQCallbacks() QueryCallbacks
	CallICQCallback(ctx sdk.Context, id string, args []byte, query Query) error
	HasICQCallback(id string) bool
}
