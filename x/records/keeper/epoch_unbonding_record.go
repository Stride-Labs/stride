package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/records/types"
)

// SetEpochUnbondingRecord set a specific epochUnbondingRecord in the store
func (k Keeper) SetEpochUnbondingRecord(ctx sdk.Context, epochUnbondingRecord types.EpochUnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	b := k.Cdc.MustMarshal(&epochUnbondingRecord)
	store.Set(GetEpochUnbondingRecordIDBytes(epochUnbondingRecord.EpochNumber), b)
}

// GetEpochUnbondingRecord returns a epochUnbondingRecord from its id
func (k Keeper) GetEpochUnbondingRecord(ctx sdk.Context, epochNumber uint64) (val types.EpochUnbondingRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	b := store.Get(GetEpochUnbondingRecordIDBytes(epochNumber))
	if b == nil {
		return val, false
	}
	k.Cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveEpochUnbondingRecord removes a epochUnbondingRecord from the store
func (k Keeper) RemoveEpochUnbondingRecord(ctx sdk.Context, epochNumber uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	store.Delete(GetEpochUnbondingRecordIDBytes(epochNumber))
}

// GetAllEpochUnbondingRecord returns all epochUnbondingRecord
func (k Keeper) GetAllEpochUnbondingRecord(ctx sdk.Context) (list []types.EpochUnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.EpochUnbondingRecord
		k.Cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetEpochUnbondingRecordIDBytes returns the byte representation of the ID
func GetEpochUnbondingRecordIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetEpochUnbondingRecordIDFromBytes returns ID in uint64 format from a byte array
func GetEpochUnbondingRecordIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}

// GetEpochUnbondingRecordByEpoch returns a epochUnbondingRecord from its epochNumber
func (k Keeper) GetHostZoneUnbondingByChainId(ctx sdk.Context, epochNumber uint64, chainId string) (val *types.HostZoneUnbonding, found bool) {
	epochUnbondingRecord, found := k.GetEpochUnbondingRecord(ctx, epochNumber)
	if !found {
		return nil, false
	}
	hostZoneUnbondings := epochUnbondingRecord.HostZoneUnbondings
	for _, hzUnbondingRecord := range hostZoneUnbondings {
		if hzUnbondingRecord.HostZoneId == chainId {
			return hzUnbondingRecord, true
		}
	}
	return &types.HostZoneUnbonding{}, false
}

func (k Keeper) AddHostZoneToEpochUnbondingRecord(ctx sdk.Context, epochNumber uint64, chainId string, hzu *types.HostZoneUnbonding) (val *types.EpochUnbondingRecord, success bool) {
	epochUnbondingRecord, found := k.GetEpochUnbondingRecord(ctx, epochNumber)
	if !found {
		return nil, false
	}
	wasSet := false
	for i, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
		if hostZoneUnbonding.GetHostZoneId() == chainId {
			epochUnbondingRecord.HostZoneUnbondings[i] = hzu
			wasSet = true
			break
		}
	}
	if !wasSet {
		// add new host zone unbonding record
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, hzu)
	}
	return &epochUnbondingRecord, true
}
