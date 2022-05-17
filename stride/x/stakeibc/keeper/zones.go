package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-labs/stride/x/stakeibc/types"
)

// GetHostZoneInfo returns zone info by chain_id
func (k Keeper) GetHostZoneInfo(ctx sdk.Context, chain_id string) (types.HostZone, bool) {
	zone := types.HostZone{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixZone)
	bz := store.Get([]byte(chain_id))
	if len(bz) == 0 {
		return zone, false
	}

	k.cdc.MustUnmarshal(bz, &zone)
	return zone, true
}

// IterateHostZones iterates zones
func (k Keeper) IterateHostZones(ctx sdk.Context, fn func(index int64, zoneInfo types.HostZone) (stop bool)) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefixZone)

	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		zone := types.HostZone{}
		k.cdc.MustUnmarshal(iterator.Value(), &zone)

		stop := fn(i, zone)

		if stop {
			break
		}
		i++
	}
}
