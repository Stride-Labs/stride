package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/ratelimit/types"
)

// Stores/Updates a quota object in the store
func (k Keeper) SetQuota(ctx sdk.Context, quota types.Quota) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QuotaKeyPrefix))

	b := k.cdc.MustMarshal(&quota)
	store.Set([]byte(quota.Name), b)
}

// Removes a quota object from the store using quota name
func (k Keeper) RemoveQuota(ctx sdk.Context, name string) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QuotaKeyPrefix))
	store.Delete([]byte(name))
}

// Get a quota from the store using quota name
func (k Keeper) GetQuota(ctx sdk.Context, name string) (quota types.Quota) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QuotaKeyPrefix))

	b := store.Get([]byte(name))
	if b == nil {
		return quota
	}

	k.cdc.MustUnmarshal(b, &quota)
	return quota
}

// Get all quotas from the store
func (k Keeper) GetAllQuotas(ctx sdk.Context) (list []types.Quota) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.QuotaKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		quota := types.Quota{}
		k.cdc.MustUnmarshal(iterator.Value(), &quota)
		list = append(list, quota)
	}

	return list
}

// IsExpired checks relative to current block time
func (k Keeper) IsExpired(ctx sdk.Context, quota types.Quota) bool {
	return ctx.BlockTime().Unix() > int64(quota.PeriodEnd)
}

// CheckExceedsQuota checks if new in/out flow is going to reach the max in/out or not
func (k Keeper) CheckExceedsQuota(direction types.PacketDirection, amount uint64, totalValue uint64, quota types.Quota) bool {
	if direction == types.PACKET_RECV {
		return amount > totalValue*quota.MaxPercentRecv/100
	}

	return amount > totalValue*quota.MaxPercentSend/100
}
