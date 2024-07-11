package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
)

func (k Keeper) GetAirdropRecords(ctx sdk.Context) []types.AirdropRecord {
	// TODO[airdrop] add pagination?
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropRecordsKeyPrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	allAirdrops := []types.AirdropRecord{}
	for ; iterator.Valid(); iterator.Next() {

		airdrop := types.AirdropRecord{}
		k.cdc.MustUnmarshal(iterator.Value(), &airdrop)
		allAirdrops = append(allAirdrops, airdrop)
	}

	return allAirdrops
}

func (k Keeper) SetAirdropRecords(ctx sdk.Context, airdropRecords []types.AirdropRecord) {
	for _, airdrop := range airdropRecords {
		store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AirdropRecordsKeyPrefix)

		key := types.AirdropRecordKeyPrefix(airdrop.Id)
		value := k.cdc.MustMarshal(&airdrop)

		store.Set(key, value)
	}
}
