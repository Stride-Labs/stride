package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/icaoracle/types"
)

// Stores/updates an oracle object in the store
func (k Keeper) SetOracle(ctx sdk.Context, oracle types.Oracle) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)

	oracleKey := types.KeyPrefix(oracle.Moniker)
	oracleValue := k.cdc.MustMarshal(&oracle)

	store.Set(oracleKey, oracleValue)
}

// Grabs and returns an oracle object from the store using the moniker
func (k Keeper) GetOracle(ctx sdk.Context, moniker string) (oracle types.Oracle, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)

	oracleKey := types.KeyPrefix(moniker)
	oracleBz := store.Get(oracleKey)

	if len(oracleBz) == 0 {
		return oracle, false
	}

	k.cdc.MustUnmarshal(oracleBz, &oracle)
	return oracle, true
}

// Returns all oracles
func (k Keeper) GetAllOracles(ctx sdk.Context) []types.Oracle {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allOracles := []types.Oracle{}
	for ; iterator.Valid(); iterator.Next() {

		oracle := types.Oracle{}
		k.cdc.MustUnmarshal(iterator.Value(), &oracle)
		allOracles = append(allOracles, oracle)
	}

	return allOracles
}

// Removes an oracle from the store
func (k Keeper) RemoveOracle(ctx sdk.Context, moniker string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OracleKeyPrefix)
	oracleKey := types.KeyPrefix(moniker)
	store.Delete(oracleKey)
}

// Toggle whether an oracle is active
func (k Keeper) ToggleOracle(ctx sdk.Context, moniker string, active bool) error {
	oracle, found := k.GetOracle(ctx, moniker)
	if !found {
		return types.ErrOracleNotFound
	}

	oracle.Active = active
	k.SetOracle(ctx, oracle)
	return nil
}
