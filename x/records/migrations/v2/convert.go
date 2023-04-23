package v2

import (
	sdkmath "cosmossdk.io/math"

	oldrecordstypes "github.com/Stride-Labs/stride/v9/x/records/migrations/v2/types"
	recordstypes "github.com/Stride-Labs/stride/v9/x/records/types"
)

func convertToNewDepositRecord(oldDepositRecord oldrecordstypes.DepositRecord) recordstypes.DepositRecord {
	return recordstypes.DepositRecord{
		Id:                 oldDepositRecord.Id,
		Amount:             sdkmath.NewInt(oldDepositRecord.Amount),
		Denom:              oldDepositRecord.Denom,
		HostZoneId:         oldDepositRecord.HostZoneId,
		Status:             recordstypes.DepositRecord_Status(oldDepositRecord.Status),
		DepositEpochNumber: oldDepositRecord.DepositEpochNumber,
		Source:             recordstypes.DepositRecord_Source(oldDepositRecord.Source),
	}
}

func convertToNewHostZoneUnbonding(oldHostZoneUnbondings oldrecordstypes.HostZoneUnbonding) recordstypes.HostZoneUnbonding {
	return recordstypes.HostZoneUnbonding{
		StTokenAmount:         sdkmath.NewIntFromUint64(oldHostZoneUnbondings.StTokenAmount),
		NativeTokenAmount:     sdkmath.NewIntFromUint64(oldHostZoneUnbondings.NativeTokenAmount),
		Denom:                 oldHostZoneUnbondings.Denom,
		HostZoneId:            oldHostZoneUnbondings.HostZoneId,
		UnbondingTime:         oldHostZoneUnbondings.UnbondingTime,
		Status:                recordstypes.HostZoneUnbonding_Status(oldHostZoneUnbondings.Status),
		UserRedemptionRecords: oldHostZoneUnbondings.UserRedemptionRecords,
	}
}

func convertToNewEpochUnbondingRecord(oldEpochUnbondingRecord oldrecordstypes.EpochUnbondingRecord) recordstypes.EpochUnbondingRecord {
	var epochUnbondingRecord recordstypes.EpochUnbondingRecord
	for _, oldHostZoneUnbonding := range oldEpochUnbondingRecord.HostZoneUnbondings {
		newHostZoneUnbonding := convertToNewHostZoneUnbonding(*oldHostZoneUnbonding)
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &newHostZoneUnbonding)
	}
	return epochUnbondingRecord
}

func convertToNewUserRedemptionRecord(oldRedemptionRecord oldrecordstypes.UserRedemptionRecord) recordstypes.UserRedemptionRecord {
	return recordstypes.UserRedemptionRecord{
		Id:             oldRedemptionRecord.Id,
		Sender:         oldRedemptionRecord.Sender,
		Receiver:       oldRedemptionRecord.Receiver,
		Amount:         sdkmath.NewIntFromUint64(oldRedemptionRecord.Amount),
		Denom:          oldRedemptionRecord.Denom,
		HostZoneId:     oldRedemptionRecord.HostZoneId,
		EpochNumber:    oldRedemptionRecord.EpochNumber,
		ClaimIsPending: oldRedemptionRecord.ClaimIsPending,
	}
}
