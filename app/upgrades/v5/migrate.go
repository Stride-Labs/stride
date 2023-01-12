package v5

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	claimtypes "github.com/Stride-Labs/stride/v4/x/claim/types"
	claimv1types "github.com/Stride-Labs/stride/v4/x/claim/types/v1"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
	recordv1types "github.com/Stride-Labs/stride/v4/x/records/types/v1"
	stakeibctypes "github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	stakeibcv1types "github.com/Stride-Labs/stride/v4/x/stakeibc/types/v1"
)

func migrateClaimParams(store sdk.KVStore, cdc codec.Codec) error {
	// paramsStore := prefix.NewStore(store, []byte(claimtypes.ParamsKey))
	oldBz := store.Get([]byte(claimtypes.ParamsKey))
	var oldProp claimv1types.Params
	err := cdc.UnmarshalJSON(oldBz, &oldProp)
	if err != nil {
		return err
	}
	newProp := convertToNewClaimParams(oldProp)
	newBz, err := cdc.MarshalJSON(&newProp)
	if err != nil {
		return err
	}
	// Set new value on store.
	store.Set([]byte(claimtypes.ParamsKey), newBz)
	return nil
}

func migrateUserRedemptionRecord(store sdk.KVStore, cdc codec.Codec) error {
	paramsStore := prefix.NewStore(store, []byte(recordtypes.UserRedemptionRecordKey))

	iter := paramsStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldProp recordv1types.UserRedemptionRecord
		err := cdc.Unmarshal(iter.Value(), &oldProp)
		if err != nil {
			return err
		}

		newProp := convertToNewUserRedemptionRecord(oldProp)
		bz, err := cdc.Marshal(&newProp)
		if err != nil {
			return err
		}

		// Set new value on store.
		paramsStore.Set(iter.Key(), bz)
	}
	
	return nil
}

func migrateDepositRecord(store sdk.KVStore, cdc codec.Codec) error {
	paramsStore := prefix.NewStore(store, []byte(recordtypes.DepositRecordKey))

	iter := paramsStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldProp recordv1types.DepositRecord
		err := cdc.Unmarshal(iter.Value(), &oldProp)
		if err != nil {
			return err
		}
		newProp := convertToNewDepositRecord(oldProp)
		bz, err := cdc.Marshal(&newProp)
		if err != nil {
			return err
		}

		// Set new value on store.
		paramsStore.Set(iter.Key(), bz)
	}
	
	return nil
}

func migrateEpochUnbondingRecord(store sdk.KVStore, cdc codec.Codec) error {
	paramsStore := prefix.NewStore(store, []byte(recordtypes.EpochUnbondingRecordKey))

	iter := paramsStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldProp recordv1types.EpochUnbondingRecord
		err := cdc.Unmarshal(iter.Value(), &oldProp)
		if err != nil {
			return err
		}

		newProp := convertToNewEpochUnbondingRecord(oldProp)
		bz, err := cdc.Marshal(&newProp)
		if err != nil {
			return err
		}

		// Set new value on store.
		paramsStore.Set(iter.Key(), bz)
	}
	
	return nil
}

func migrateHostZone(store sdk.KVStore, cdc codec.Codec) error {
	paramsStore := prefix.NewStore(store, []byte(stakeibctypes.HostZoneKey))

	iter := paramsStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var oldProp stakeibcv1types.HostZone
		err := cdc.Unmarshal(iter.Value(), &oldProp)
		if err != nil {
			return err
		}

		newProp := convertToNewHostZone(oldProp)
		bz, err := cdc.Marshal(&newProp)
		if err != nil {
			return err
		}

		// Set new value on store.
		paramsStore.Set(iter.Key(), bz)
	}
	
	return nil
}

func MigrateStore(ctx sdk.Context, claimStoreKey storetypes.StoreKey, recordStoreKey storetypes.StoreKey, stakeibcStoreKey storetypes.StoreKey, cdc codec.Codec) error {
	
	// Migrate claim module store
	claimStore := ctx.KVStore(claimStoreKey)
	err := migrateClaimParams(claimStore, cdc)
	if err != nil {
		return err
	}

	// Migrate record module store
	recordStore := ctx.KVStore(recordStoreKey)
	err = migrateUserRedemptionRecord(recordStore, cdc)
	if err != nil {
		return err
	}
	err = migrateDepositRecord(recordStore, cdc)
	if err != nil {
		return err
	}
	err = migrateEpochUnbondingRecord(recordStore, cdc)
	if err != nil {
		return err
	}

	// Migrate stakeibc module store
	stakeibcStore := ctx.KVStore(stakeibcStoreKey)
	err = migrateHostZone(stakeibcStore, cdc)
	if err != nil {
		return err
	}

	return nil
}