package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	"github.com/Stride-Labs/stride/v28/x/records/types"
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
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

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
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

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
func (k Keeper) AddHostZoneToEpochUnbondingRecord(
	ctx sdk.Context,
	epochNumber uint64,
	chainId string,
	hzu types.HostZoneUnbonding,
) (eur types.EpochUnbondingRecord, err error) {
	epochUnbondingRecord, found := k.GetEpochUnbondingRecord(ctx, epochNumber)
	if !found {
		return types.EpochUnbondingRecord{}, types.ErrEpochUnbondingRecordNotFound.Wrapf("epoch number %d", epochNumber)
	}

	// Check if the hzu is already in the epoch unbonding record - if so, replace it
	hzuAlreadyExists := false
	for i, hostZoneUnbonding := range epochUnbondingRecord.HostZoneUnbondings {
		if hostZoneUnbonding.HostZoneId == chainId {
			epochUnbondingRecord.HostZoneUnbondings[i] = &hzu
			hzuAlreadyExists = true
			break
		}
	}

	// If the hzu didn't already exist, add a new record
	if !hzuAlreadyExists {
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &hzu)
	}
	return epochUnbondingRecord, nil
}

// Stores a host zone unbonding record - set via an epoch unbonding record
func (k Keeper) SetHostZoneUnbondingRecord(ctx sdk.Context, epochNumber uint64, chainId string, hostZoneUnbonding types.HostZoneUnbonding) error {
	epochUnbondingRecord, err := k.AddHostZoneToEpochUnbondingRecord(ctx, epochNumber, chainId, hostZoneUnbonding)
	if err != nil {
		return err
	}
	k.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return nil
}

// Updates the status for a given host zone across relevant epoch unbonding record IDs
func (k Keeper) SetHostZoneUnbondingStatus(ctx sdk.Context, chainId string, epochUnbondingRecordIds []uint64, status types.HostZoneUnbonding_Status) error {
	for _, epochUnbondingRecordId := range epochUnbondingRecordIds {
		k.Logger(ctx).Info(fmt.Sprintf("Updating host zone unbondings on EpochUnbondingRecord %d to status %s", epochUnbondingRecordId, status.String()))

		// fetch the host zone unbonding
		hostZoneUnbonding, found := k.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecordId, chainId)
		if !found {
			return errorsmod.Wrapf(types.ErrHostUnbondingRecordNotFound, "epoch number %d, chain %s",
				epochUnbondingRecordId, chainId)
		}
		hostZoneUnbonding.Status = status

		// save the updated hzu on the epoch unbonding record
		if err := k.SetHostZoneUnbondingRecord(ctx, epochUnbondingRecordId, chainId, *hostZoneUnbonding); err != nil {
			return err
		}
	}
	return nil
}
