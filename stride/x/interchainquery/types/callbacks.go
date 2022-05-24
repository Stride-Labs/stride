package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type QueryCallbacks interface {
	AddCallback(id string, fn interface{})
	RemoveCallback(id string)
	//Call(id string, ctx sdk.Context, args proto.Message) error
	Call(ctx sdk.Context, id string, args []byte, query Query) error
	Has(id string) bool
}
