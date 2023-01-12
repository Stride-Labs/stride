package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldrecordtypes "github.com/Stride-Labs/stride/v4/x/records/migrations/v2/types"
	recordtypes "github.com/Stride-Labs/stride/v4/x/records/types"
)

func convertToNewDepositRecord(oldDepositRecord oldrecordtypes.DepositRecord) recordtypes.DepositRecord {
	return recordtypes.DepositRecord{
		Id:                 oldDepositRecord.Id,
		Amount:             sdk.NewInt(oldDepositRecord.Amount),
		Denom:              oldDepositRecord.Denom,
		HostZoneId:         oldDepositRecord.HostZoneId,
		Status:             recordtypes.DepositRecord_Status(oldDepositRecord.Status),
		DepositEpochNumber: oldDepositRecord.DepositEpochNumber,
		Source:             recordtypes.DepositRecord_Source(oldDepositRecord.Source),
	}
}

func convertToNewEpochUnbondingRecord(oldEpochUnbondingRecord oldrecordtypes.EpochUnbondingRecord) recordtypes.EpochUnbondingRecord {
	var epochUnbondingRecord recordtypes.EpochUnbondingRecord
	for _, hz := range oldEpochUnbondingRecord.HostZoneUnbondings {
		newHz := recordtypes.HostZoneUnbonding{
			StTokenAmount:         sdk.NewIntFromUint64(hz.StTokenAmount),
			NativeTokenAmount:     sdk.NewIntFromUint64(hz.NativeTokenAmount),
			Denom:                 hz.Denom,
			HostZoneId:            hz.HostZoneId,
			UnbondingTime:         hz.UnbondingTime,
			Status:                recordtypes.HostZoneUnbonding_Status(hz.Status),
			UserRedemptionRecords: hz.UserRedemptionRecords,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &newHz)
	}
	return epochUnbondingRecord
}

func convertToNewUserRedemptionRecord(oldRedemptionRecord oldrecordtypes.UserRedemptionRecord) recordtypes.UserRedemptionRecord {
	return recordtypes.UserRedemptionRecord{
		Id:             oldRedemptionRecord.Id,
		Sender:         oldRedemptionRecord.Sender,
		Receiver:       oldRedemptionRecord.Receiver,
		Amount:         sdk.NewIntFromUint64(oldRedemptionRecord.Amount),
		Denom:          oldRedemptionRecord.Denom,
		HostZoneId:     oldRedemptionRecord.HostZoneId,
		EpochNumber:    oldRedemptionRecord.EpochNumber,
		ClaimIsPending: oldRedemptionRecord.ClaimIsPending,
	}
}

func migrateDepositRecord(store sdk.KVStore, cdc codec.BinaryCodec) error {
	paramsStore := prefix.NewStore(store, []byte(recordtypes.DepositRecordKey))

	iterator := paramsStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldDepositRecord oldrecordtypes.DepositRecord
		err := cdc.Unmarshal(iterator.Value(), &oldDepositRecord)
		if err != nil {
			return err
		}

		// Convert and serialize using the new type
		newDepositRecord := convertToNewDepositRecord(oldDepositRecord)
		newDepositRecordBz, err := cdc.Marshal(&newDepositRecord)
		if err != nil {
			return err
		}

		// Store the new type
		paramsStore.Set(iterator.Key(), newDepositRecordBz)
	}

	return nil
}

func migrateUserRedemptionRecord(store sdk.KVStore, cdc codec.BinaryCodec) error {
	paramsStore := prefix.NewStore(store, []byte(recordtypes.UserRedemptionRecordKey))

	iterator := paramsStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldRedemptionRecord oldrecordtypes.UserRedemptionRecord
		err := cdc.Unmarshal(iterator.Value(), &oldRedemptionRecord)
		if err != nil {
			return err
		}

		// Convert and serialize using the new type
		newRedemptionRecord := convertToNewUserRedemptionRecord(oldRedemptionRecord)
		newRedemptionRecordBz, err := cdc.Marshal(&newRedemptionRecord)
		if err != nil {
			return err
		}

		// Store the new type
		paramsStore.Set(iterator.Key(), newRedemptionRecordBz)
	}

	return nil
}

func migrateEpochUnbondingRecord(store sdk.KVStore, cdc codec.BinaryCodec) error {
	paramsStore := prefix.NewStore(store, []byte(recordtypes.EpochUnbondingRecordKey))

	iterator := paramsStore.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		// Deserialize using the old type
		var oldEpochUnbondingRecord oldrecordtypes.EpochUnbondingRecord
		err := cdc.Unmarshal(iterator.Value(), &oldEpochUnbondingRecord)
		if err != nil {
			return err
		}

		// Convert and serialize using the new type
		newEpochUnbondingRecord := convertToNewEpochUnbondingRecord(oldEpochUnbondingRecord)
		newEpochUnbondingRecordBz, err := cdc.Marshal(&newEpochUnbondingRecord)
		if err != nil {
			return err
		}

		// Store the new type
		paramsStore.Set(iterator.Key(), newEpochUnbondingRecordBz)
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
