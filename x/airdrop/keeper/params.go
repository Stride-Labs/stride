package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v27/x/airdrop/types"
)

// Writes params to the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	store := ctx.KVStore(k.storeKey)
	paramsBz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsPrefix, paramsBz)
}

// Retrieves the module parameters
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	store := ctx.KVStore(k.storeKey)
	paramsBz := store.Get(types.ParamsPrefix)
	if len(paramsBz) == 0 {
		panic("module parameters not set")
	}
	k.cdc.MustUnmarshal(paramsBz, &params)
	return params
}
