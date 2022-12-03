package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// SetMinValidatorRequirements set minValidatorRequirements in the store
func (k Keeper) SetMinValidatorRequirements(ctx sdk.Context, minValidatorRequirements types.MinValidatorRequirements) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinValidatorRequirementsKey))
	b := k.cdc.MustMarshal(&minValidatorRequirements)
	store.Set([]byte{0}, b)
}

// GetMinValidatorRequirements returns minValidatorRequirements
func (k Keeper) GetMinValidatorRequirements(ctx sdk.Context) (val types.MinValidatorRequirements, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinValidatorRequirementsKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMinValidatorRequirements removes minValidatorRequirements from the store
func (k Keeper) RemoveMinValidatorRequirements(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MinValidatorRequirementsKey))
	store.Delete([]byte{0})
}
