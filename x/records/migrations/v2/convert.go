package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	oldrecordstypes "github.com/Stride-Labs/stride/v4/x/records/migrations/v2/types"
	recordstypes "github.com/Stride-Labs/stride/v4/x/records/types"
)

func convertToNewDepositRecord(oldDepositRecord oldrecordstypes.DepositRecord) recordstypes.DepositRecord {
	return recordstypes.DepositRecord{
		Id:                 oldDepositRecord.Id,
		Amount:             sdk.NewInt(oldDepositRecord.Amount),
		Denom:              oldDepositRecord.Denom,
		HostZoneId:         oldDepositRecord.HostZoneId,
		Status:             recordstypes.DepositRecord_Status(oldDepositRecord.Status),
		DepositEpochNumber: oldDepositRecord.DepositEpochNumber,
		Source:             recordstypes.DepositRecord_Source(oldDepositRecord.Source),
	}
}

func convertToNewEpochUnbondingRecord(oldEpochUnbondingRecord oldrecordstypes.EpochUnbondingRecord) recordstypes.EpochUnbondingRecord {
	var epochUnbondingRecord recordstypes.EpochUnbondingRecord
	for _, hz := range oldEpochUnbondingRecord.HostZoneUnbondings {
		newHz := recordstypes.HostZoneUnbonding{
			StTokenAmount:         sdk.NewIntFromUint64(hz.StTokenAmount),
			NativeTokenAmount:     sdk.NewIntFromUint64(hz.NativeTokenAmount),
			Denom:                 hz.Denom,
			HostZoneId:            hz.HostZoneId,
			UnbondingTime:         hz.UnbondingTime,
			Status:                recordstypes.HostZoneUnbonding_Status(hz.Status),
			UserRedemptionRecords: hz.UserRedemptionRecords,
		}
		epochUnbondingRecord.HostZoneUnbondings = append(epochUnbondingRecord.HostZoneUnbondings, &newHz)
	}
	return epochUnbondingRecord
}

func convertToNewUserRedemptionRecord(oldRedemptionRecord oldrecordstypes.UserRedemptionRecord) recordstypes.UserRedemptionRecord {
	return recordstypes.UserRedemptionRecord{
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
