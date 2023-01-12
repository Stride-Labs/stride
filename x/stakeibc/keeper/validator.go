package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// SetValidator set validator in the store
func (k Keeper) SetValidator(ctx sdk.Context, validator types.Validator) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ValidatorKey))
	b := k.cdc.MustMarshal(&validator)
	store.Set([]byte{0}, b)
}

// GetValidator returns validator
func (k Keeper) GetValidator(ctx sdk.Context) (val types.Validator, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ValidatorKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveValidator removes validator from the store
func (k Keeper) RemoveValidator(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ValidatorKey))
	store.Delete([]byte{0})
}

// get a validator and its index from a list of validators, by address
func GetValidatorFromAddress(validators []*types.Validator, address string) (val types.Validator, index int64, found bool) {
	for i, v := range validators {
		if v.Address == address {
			return *v, int64(i), true
		}
	}
	return types.Validator{}, 0, false
}
