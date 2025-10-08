package keeper

import (
	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v29/x/airdrop/types"
)

// Writes an airdrop configuration to the store
func (k Keeper) SetAirdrop(ctx sdk.Context, airdrop types.Airdrop) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropKeyPrefix)

	key := types.KeyPrefix(airdrop.Id)
	value := k.cdc.MustMarshal(&airdrop)

	store.Set(key, value)
}

// Retrieves an airdrop configuration from the store
func (k Keeper) GetAirdrop(ctx sdk.Context, airdropId string) (airdrop types.Airdrop, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropKeyPrefix)

	key := types.KeyPrefix(airdropId)
	airdropBz := store.Get(key)

	if len(airdropBz) == 0 {
		return airdrop, false
	}

	k.cdc.MustUnmarshal(airdropBz, &airdrop)
	return airdrop, true
}

// Removes an airdrop configuration from the store
func (k Keeper) RemoveAirdrop(ctx sdk.Context, airdropId string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropKeyPrefix)
	key := types.KeyPrefix(airdropId)
	store.Delete(key)
}

// Retrieves all airdrop configurations from the store
func (k Keeper) GetAllAirdrops(ctx sdk.Context) (airdrops []types.Airdrop) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		airdrop := types.Airdrop{}
		k.cdc.MustUnmarshal(iterator.Value(), &airdrop)
		airdrops = append(airdrops, airdrop)
	}

	return airdrops
}
