package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/records/types"
)

// SetUserRedemptionRecord set a specific userRedemptionRecord in the store
func (k Keeper) SetUserRedemptionRecord(ctx sdk.Context, userRedemptionRecord types.UserRedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	b := k.Cdc.MustMarshal(&userRedemptionRecord)
	store.Set([]byte(userRedemptionRecord.Id), b)
}

// GetUserRedemptionRecord returns a userRedemptionRecord from its id
func (k Keeper) GetUserRedemptionRecord(ctx sdk.Context, id string) (val types.UserRedemptionRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	b := store.Get([]byte(id))
	if b == nil {
		return val, false
	}
	k.Cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveUserRedemptionRecord removes a userRedemptionRecord from the store
func (k Keeper) RemoveUserRedemptionRecord(ctx sdk.Context, id string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	store.Delete([]byte(id))
}

// GetAllUserRedemptionRecord returns all userRedemptionRecord
func (k Keeper) GetAllUserRedemptionRecord(ctx sdk.Context) (list []types.UserRedemptionRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.UserRedemptionRecord
		k.Cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// IterateUserRedemptionRecords iterates zones
func (k Keeper) IterateUserRedemptionRecords(ctx sdk.Context,
	fn func(index int64, userRedemptionRecord types.UserRedemptionRecord) (stop bool),
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.UserRedemptionRecordKey))

	iterator := sdk.KVStorePrefixIterator(store, nil)
	defer iterator.Close()

	i := int64(0)

	for ; iterator.Valid(); iterator.Next() {
		userRedRecord := types.UserRedemptionRecord{}
		k.Cdc.MustUnmarshal(iterator.Value(), &userRedRecord)

		stop := fn(i, userRedRecord)

		if stop {
			break
		}
		i++
	}
}
