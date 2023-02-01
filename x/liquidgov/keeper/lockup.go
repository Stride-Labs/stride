package keeper

import (
	"time"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v5/x/liquidgov/types"
)

// GetLockup returns a specific lockup.
func (k Keeper) GetLockup(ctx sdk.Context, creatorAddr sdk.AccAddress, denom string) (lockup types.Lockup, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetLockupKey(creatorAddr, denom)

	value := store.Get(key)
	if value == nil {
		return lockup, false
	}

	lockup = types.MustUnmarshalLockup(k.cdc, value)

	return lockup, true
}

// IterateAllLockups iterates through all of the lockup.
func (k Keeper) IterateAllLockups(ctx sdk.Context, cb func(lockup types.Lockup) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.LockupKey)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		lockup := types.MustUnmarshalLockup(k.cdc, iterator.Value())
		if cb(lockup) {
			break
		}
	}
}

// GetAlllockups returns all lockups used during genesis dump.
func (k Keeper) GetAllLockups(ctx sdk.Context) (lockups []types.Lockup) {
	k.IterateAllLockups(ctx, func(lockup types.Lockup) bool {
		lockups = append(lockups, lockup)
		return false
	})

	return lockups
}

// GetCreatorLockups returns a given amount of all the lockups from a
// creator.
func (k Keeper) GetCreatorLockups(ctx sdk.Context, creator sdk.AccAddress, maxRetrieve uint16) (lockups []types.Lockup) {
	lockups = make([]types.Lockup, maxRetrieve)
	store := ctx.KVStore(k.storeKey)
	lockupPrefixKey := types.GetLockupsKey(creator)

	iterator := sdk.KVStorePrefixIterator(store, lockupPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		lockup := types.MustUnmarshalLockup(k.cdc, iterator.Value())
		lockups[i] = lockup
		i++
	}

	return lockups[:i] // trim if the array length < maxRetrieve
}

// SetLockup sets a lockup.
func (k Keeper) SetLockup(ctx sdk.Context, lockup types.Lockup) {
	creatorAddress := sdk.MustAccAddressFromBech32(lockup.Creator)

	store := ctx.KVStore(k.storeKey)
	b := types.MustMarshalLockup(k.cdc, lockup)
	store.Set(types.GetLockupKey(creatorAddress, lockup.GetDenom()), b)
}

// RemoveLockup removes a lockup
func (k Keeper) RemoveLockup(ctx sdk.Context, lockup types.Lockup) error {
	creatorAddress := sdk.MustAccAddressFromBech32(lockup.Creator)

	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetLockupKey(creatorAddress, lockup.GetDenom()))
	return nil
}

// GetUnlockingRecords returns a given amount of all the creator unlocking-records.
func (k Keeper) GetUnlockingRecords(ctx sdk.Context, creator sdk.AccAddress, maxRetrieve uint16) (unlockingRecords []types.UnlockingRecord) {
	unlockingRecords = make([]types.UnlockingRecord, maxRetrieve)

	store := ctx.KVStore(k.storeKey)
	creatorPrefixKey := types.GetURsKey(creator)

	iterator := sdk.KVStorePrefixIterator(store, creatorPrefixKey)
	defer iterator.Close()

	i := 0
	for ; iterator.Valid() && i < int(maxRetrieve); iterator.Next() {
		unlockingRecord := types.MustUnmarshalUR(k.cdc, iterator.Value())
		unlockingRecords[i] = unlockingRecord
		i++
	}

	return unlockingRecords[:i] // trim if the array length < maxRetrieve
}

// GetUnlockingRecord returns a unlocking record.
func (k Keeper) GetUnlockingRecord(ctx sdk.Context, creator sdk.AccAddress, denom string) (ur types.UnlockingRecord, found bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetURKey(creator, denom)
	value := store.Get(key)

	if value == nil {
		return ur, false
	}

	ur = types.MustUnmarshalUR(k.cdc, value)

	return ur, true
}

// TODO: use in invariant?
// IterateUnlockingRecords iterates through all of the unlocking records.
func (k Keeper) IterateUnlockingRecords(ctx sdk.Context, fn func(index int64, ur types.UnlockingRecord) (stop bool)) {
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.UnlockingRecordKey)
	defer iterator.Close()

	for i := int64(0); iterator.Valid(); iterator.Next() {
		ur := types.MustUnmarshalUR(k.cdc, iterator.Value())
		if stop := fn(i, ur); stop {
			break
		}
		i++
	}
}

/* TODO: Necessary?
// HasMaxUnlockingRecordEntries - check if unlocking record has maximum number of entries.
func (k Keeper) HasMaxUnlockingRecordEntries(ctx sdk.Context, creator sdk.AccAddress, denom string) bool {
	ur, found := k.GetUnlockingRecord(ctx, creator, denom)
	if !found {
		return false
	}

	return len(ur.Entries) >= int(k.MaxEntries(ctx))
}
*/
// SetUnlockingRecord sets the unlocking record
func (k Keeper) SetUnlockingRecord(ctx sdk.Context, ur types.UnlockingRecord) {
	creatorAddress := sdk.MustAccAddressFromBech32(ur.Creator)

	store := ctx.KVStore(k.storeKey)
	bz := types.MustMarshalUR(k.cdc, ur)
	key := types.GetURKey(creatorAddress, ur.Denom)
	store.Set(key, bz)
}

// RemoveUnlockingRecord removes the unlocking record object.
func (k Keeper) RemoveUnlockingRecord(ctx sdk.Context, ur types.UnlockingRecord) {
	creatorAddress := sdk.MustAccAddressFromBech32(ur.Creator)

	store := ctx.KVStore(k.storeKey)
	key := types.GetURKey(creatorAddress, ur.Denom)
	store.Delete(key)
}

// SetUnlockingRecordEntry adds an entry to the unlocking record at
// the given addresses. It creates the unlocking record if it does not exist.
func (k Keeper) SetUnlockingRecordEntry(
	ctx sdk.Context, creatorAddr sdk.AccAddress, denom string,
	creationHeight int64, minTime time.Time, balance math.Int,
) types.UnlockingRecord {
	ur, found := k.GetUnlockingRecord(ctx, creatorAddr, denom)
	if found {
		ur.AddEntry(creationHeight, minTime, balance)
	} else {
		ur = types.NewUnlockingRecord(creatorAddr, denom, creationHeight, minTime, balance)
	}

	k.SetUnlockingRecord(ctx, ur)

	return ur
}
