package v2

import (
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
