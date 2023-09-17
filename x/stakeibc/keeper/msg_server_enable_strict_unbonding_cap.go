package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const EnableStrictUnbondingCapKey = "strict-unbonding-cap"

func (k msgServer) EnableStrictUnbondingCap(goCtx context.Context, msg *types.MsgEnableStrictUnbondingCap) (*types.MsgEnableStrictUnbondingCapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// change the state in the store
	store := ctx.KVStore(k.storeKey)

	if k.IsStrictUnbondingEnabled(ctx) { // set the cap to false
		store.Set([]byte(EnableStrictUnbondingCapKey), []byte{0})
	} else { // set the cap to true
		store.Set([]byte(EnableStrictUnbondingCapKey), []byte{1})
	}

	return &types.MsgEnableStrictUnbondingCapResponse{}, nil
}

func (k Keeper) IsStrictUnbondingEnabled(ctx sdk.Context) bool {
	store := ctx.KVStore(k.storeKey)
	if !store.Has([]byte(EnableStrictUnbondingCapKey)) {
		return false
	}

	value := store.Get([]byte(EnableStrictUnbondingCapKey))
	return len(value) == 1 && value[0] == 1
}
