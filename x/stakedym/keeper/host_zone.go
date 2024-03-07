package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v18/x/stakedym/types"
)

// Writes a host zone to the store
func (k Keeper) SetHostZone(ctx sdk.Context, hostZone types.HostZone) {
	store := ctx.KVStore(k.storeKey)
	hostZoneBz := k.cdc.MustMarshal(&hostZone)
	store.Set(types.HostZoneKey, hostZoneBz)
}

// Reads a host zone from the store
// There should always be a host zone, so this should error if it is not found
func (k Keeper) GetHostZone(ctx sdk.Context) (hostZone types.HostZone, err error) {
	store := ctx.KVStore(k.storeKey)
	hostZoneBz := store.Get(types.HostZoneKey)

	if len(hostZoneBz) == 0 {
		return hostZone, types.ErrHostZoneNotFound.Wrapf("No HostZone found, there must be exactly one HostZone!")
	}

	k.cdc.MustUnmarshal(hostZoneBz, &hostZone)
	return hostZone, nil
}

// Removes a host zone from the store
// Note: This is only for testing - it should never be used elsewhere
func (k Keeper) RemoveHostZone(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.HostZoneKey)
}

// Reads a host zone from the store and errors if the host zone is halted
func (k Keeper) GetUnhaltedHostZone(ctx sdk.Context) (hostZone types.HostZone, err error) {
	hostZone, err = k.GetHostZone(ctx)
	if err != nil {
		return hostZone, err
	}
	if hostZone.Halted {
		return hostZone, types.ErrHostZoneHalted.Wrapf("host zone %s is halted", hostZone.ChainId)
	}
	return hostZone, nil
}
