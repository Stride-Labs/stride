package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	errorsmod "cosmossdk.io/errors"

	oldrecordtypes "github.com/Stride-Labs/stride/v9/x/records/migrations/v2/types"
	recordtypes "github.com/Stride-Labs/stride/v9/x/records/types"
)

func migrateDepositRecord(store sdk.KVStore, cdc codec.BinaryCodec) error {
	depositRecordStore := prefix.NewStore(store, []byte(recordtypes.DepositRecordKey))

	iterator := depositRecordStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldDepositRecord oldrecordtypes.DepositRecord
		err := cdc.Unmarshal(iterator.Value(), &oldDepositRecord)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal deposit record (%v) using old data type", iterator.Key())
		}

		// Convert and serialize using the new type
		newDepositRecord := convertToNewDepositRecord(oldDepositRecord)
		newDepositRecordBz, err := cdc.Marshal(&newDepositRecord)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal deposit record (%v) using new data type", iterator.Key())
		}

		// Store the new type
		depositRecordStore.Set(iterator.Key(), newDepositRecordBz)
	}

	return nil
}

func migrateUserRedemptionRecord(store sdk.KVStore, cdc codec.BinaryCodec) error {
	redemptionRecordStore := prefix.NewStore(store, []byte(recordtypes.UserRedemptionRecordKey))

	iterator := redemptionRecordStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldRedemptionRecord oldrecordtypes.UserRedemptionRecord
		err := cdc.Unmarshal(iterator.Value(), &oldRedemptionRecord)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal redemption record (%v) using old data type", iterator.Key())
		}

		// Convert and serialize using the new type
		newRedemptionRecord := convertToNewUserRedemptionRecord(oldRedemptionRecord)
		newRedemptionRecordBz, err := cdc.Marshal(&newRedemptionRecord)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal redemption record (%v) using new data type", iterator.Key())
		}

		// Store the new type
		redemptionRecordStore.Set(iterator.Key(), newRedemptionRecordBz)
	}

	return nil
}

func migrateEpochUnbondingRecord(store sdk.KVStore, cdc codec.BinaryCodec) error {
	epochUnbondingRecordStore := prefix.NewStore(store, []byte(recordtypes.EpochUnbondingRecordKey))

	iterator := epochUnbondingRecordStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldEpochUnbondingRecord oldrecordtypes.EpochUnbondingRecord
		err := cdc.Unmarshal(iterator.Value(), &oldEpochUnbondingRecord)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to unmarshal epoch unbonding record (%v) using old data type", iterator.Key())
		}

		// Convert and serialize using the new type
		newEpochUnbondingRecord := convertToNewEpochUnbondingRecord(oldEpochUnbondingRecord)
		newEpochUnbondingRecordBz, err := cdc.Marshal(&newEpochUnbondingRecord)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal epoch unbonding record (%v) using new data type", iterator.Key())
		}

		// Store the new type
		epochUnbondingRecordStore.Set(iterator.Key(), newEpochUnbondingRecordBz)
	}

	return nil
}

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)

	err := migrateDepositRecord(store, cdc)
	if err != nil {
		return err
	}

	err = migrateUserRedemptionRecord(store, cdc)
	if err != nil {
		return err
	}

	return migrateEpochUnbondingRecord(store, cdc)
}
