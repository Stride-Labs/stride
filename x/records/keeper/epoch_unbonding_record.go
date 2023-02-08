package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"

	"github.com/Stride-Labs/stride/v4/x/records/types"
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

// GetAllPreviousEpochUnbondingRecords returns all epochUnbondingRecords prior to a given epoch
func (k Keeper) GetAllPreviousEpochUnbondingRecords(ctx sdk.Context, epochNumber uint64) (list []types.EpochUnbondingRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.EpochUnbondingRecordKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	// these aren't guaranteed to be ordered
	for ; iterator.Valid(); iterator.Next() {
		var val types.EpochUnbondingRecord
		k.Cdc.MustUnmarshal(iterator.Value(), &val)
		if val.EpochNumber < epochNumber {
			list = append(list, val)
		}
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

// Adds a HostZoneUnbonding to an EpochUnbondingRecord
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

// Updates the status for a given host zone across relevant epoch unbonding record IDs
func (k Keeper) SetHostZoneUnbondings(ctx sdk.Context, chainId string, epochUnbondingRecordIds []uint64, status types.HostZoneUnbonding_Status) error {
	for _, epochUnbondingRecordId := range epochUnbondingRecordIds {
		k.Logger(ctx).Info(fmt.Sprintf("Updating host zone unbondings on EpochUnbondingRecord %d to status %s", epochUnbondingRecordId, status.String()))
		// fetch the host zone unbonding
		hostZoneUnbonding, found := k.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecordId, chainId)
		if !found {
			errMsg := fmt.Sprintf("Error fetching host zone unbonding record for epoch: %d, host zone: %s", epochUnbondingRecordId, chainId)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrapf(stakeibctypes.ErrHostZoneNotFound, errMsg)
		}
		hostZoneUnbonding.Status = status
		// save the updated hzu on the epoch unbonding record
		updatedRecord, success := k.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecordId, chainId, hostZoneUnbonding)
		if !success {
			errMsg := fmt.Sprintf("Error adding host zone unbonding record to epoch unbonding record: %d, host zone: %s", epochUnbondingRecordId, chainId)
			k.Logger(ctx).Error(errMsg)
			return sdkerrors.Wrap(types.ErrAddingHostZone, errMsg)
		}
		k.SetEpochUnbondingRecord(ctx, *updatedRecord)
	}
	return nil
}
