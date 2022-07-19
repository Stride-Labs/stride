package keeper

import (
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetterSafetyGasFeeFlag set SafetyGasFeeFlag in the store
func (k Keeper) SetterSafetyGasFeeFlag(ctx sdk.Context, safetyGasFeeFlag types.SafetyGasFeeFlag) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SafetyGasFeeFlagKey))
	b := k.cdc.MustMarshal(&safetyGasFeeFlag)
	store.Set([]byte{0}, b)
}

// GetSafetyGasFeeFlag returns SafetyGasFeeFlag
func (k Keeper) GetSafetyGasFeeFlag(ctx sdk.Context) (val types.SafetyGasFeeFlag, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SafetyGasFeeFlagKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}
