package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const EnableStrictUnbondingCapKey = "admin-gate"

func (k msgServer) EnableStrictUnbondingCap(goCtx context.Context, msg *types.MsgEnableStrictUnbondingCap) (*types.MsgEnableStrictUnbondingCapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// change the state in the store
	store := ctx.KVStore(k.storeKey)
	store.Set([]byte(EnableStrictUnbondingCapKey), []byte{1})

	return &types.MsgEnableStrictUnbondingCapResponse{}, nil
}

func (k Keeper) IsStrictUnbondingEnabled(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(EnableStrictUnbondingCapKey)) {
		return false
	}
	return store.Get([]byte(EnableStrictUnbondingCapKey))[0] == 1
}
