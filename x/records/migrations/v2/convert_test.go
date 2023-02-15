package v2

import (
	"testing"

	sdkmath "cosmossdk.io/math"

	"github.com/stretchr/testify/require"

	oldrecordstypes "github.com/Stride-Labs/stride/v5/x/records/migrations/v2/types"
	recordstypes "github.com/Stride-Labs/stride/v5/x/records/types"
)

func TestConvertDepositRecord(t *testing.T) {
	id := uint64(1)
	denom := "denom"
	hostZoneId := "hz"
	epochNumber := uint64(2)

	// Only the Amount field of the DepositRecord should change
	oldDepositRecord := oldrecordstypes.DepositRecord{
		Id:                 id,
		Amount:             int64(1),
		Denom:              denom,
		HostZoneId:         hostZoneId,
		Status:             oldrecordstypes.DepositRecord_DELEGATION_QUEUE,
		DepositEpochNumber: epochNumber,
		Source:             oldrecordstypes.DepositRecord_WITHDRAWAL_ICA,
	}
	expectedNewDepositRecord := recordstypes.DepositRecord{
		Id:                 id,
		Amount:             sdkmath.NewInt(1),
		Denom:              denom,
		HostZoneId:         hostZoneId,
		Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
		DepositEpochNumber: epochNumber,
		Source:             recordstypes.DepositRecord_WITHDRAWAL_ICA,
	}

	actualNewDepositRecord := convertToNewDepositRecord(oldDepositRecord)
	require.Equal(t, expectedNewDepositRecord, actualNewDepositRecord)
}

func TestConvertHostZoneUnbonding(t *testing.T) {
	denom := "denom"
	hostZoneId := "hz"
	unbondingTime := uint64(3)
	userRedemptionRecords := []string{"a", "b", "c"}

	// The StTokenAmount and NativeTokenAmount should change
	oldHostZoneUnbonding := oldrecordstypes.HostZoneUnbonding{
		StTokenAmount:         uint64(1),
		NativeTokenAmount:     uint64(2),
		Denom:                 denom,
		HostZoneId:            hostZoneId,
		UnbondingTime:         unbondingTime,
		Status:                oldrecordstypes.HostZoneUnbonding_CLAIMABLE,
		UserRedemptionRecords: userRedemptionRecords,
	}
	expectedNewHostZoneUnbonding := recordstypes.HostZoneUnbonding{
		StTokenAmount:         sdkmath.NewInt(1),
		NativeTokenAmount:     sdkmath.NewInt(2),
		Denom:                 denom,
		HostZoneId:            hostZoneId,
		UnbondingTime:         unbondingTime,
		Status:                recordstypes.HostZoneUnbonding_CLAIMABLE,
		UserRedemptionRecords: userRedemptionRecords,
	}

	actualNewHostZoneUnbonding := convertToNewHostZoneUnbonding(oldHostZoneUnbonding)
	require.Equal(t, expectedNewHostZoneUnbonding, actualNewHostZoneUnbonding)
}

func TestConvertEpochUnbondingRecord(t *testing.T) {
	numHostZoneUnbondings := 3

	// Build a list of old hostZoneUnbondings as well as the new expected type
	oldEpochUnbondingRecord := oldrecordstypes.EpochUnbondingRecord{}
	expectedNewEpochUnbondingRecord := recordstypes.EpochUnbondingRecord{}
	for i := 0; i <= numHostZoneUnbondings-1; i++ {
		oldEpochUnbondingRecord.HostZoneUnbondings = append(oldEpochUnbondingRecord.HostZoneUnbondings, &oldrecordstypes.HostZoneUnbonding{
			StTokenAmount:     uint64(i),
			NativeTokenAmount: uint64(i * 10),
		})

		expectedNewEpochUnbondingRecord.HostZoneUnbondings = append(expectedNewEpochUnbondingRecord.HostZoneUnbondings, &recordstypes.HostZoneUnbonding{
			StTokenAmount:     sdkmath.NewInt(int64(i)),
			NativeTokenAmount: sdkmath.NewInt(int64(i * 10)),
		})
	}

	// Convert epoch unbonding record
	actualNewEpochUnbondingRecord := convertToNewEpochUnbondingRecord(oldEpochUnbondingRecord)

	// Confirm new host zone unbondings align with expectations
	require.Equal(t, len(expectedNewEpochUnbondingRecord.HostZoneUnbondings), len(actualNewEpochUnbondingRecord.HostZoneUnbondings))
	for i := 0; i <= numHostZoneUnbondings-1; i++ {
		require.Equal(t, expectedNewEpochUnbondingRecord.HostZoneUnbondings[i], actualNewEpochUnbondingRecord.HostZoneUnbondings[i], "index: %d", i)
	}
}

func TestConvertUserRedemptionRecord(t *testing.T) {
	id := "id"
	sender := "sender"
	receiver := "receiver"
	denom := "denom"
	hostZoneId := "hz"
	epochNumber := uint64(1)
	claimIsPending := true

	// Only the Amount field of the UserRedemptionRecord should change
	oldUserRedemptionRecord := oldrecordstypes.UserRedemptionRecord{
		Id:             id,
		Sender:         sender,
		Receiver:       receiver,
		Amount:         uint64(1),
		Denom:          denom,
		HostZoneId:     hostZoneId,
		EpochNumber:    epochNumber,
		ClaimIsPending: claimIsPending,
	}
	expectedNewUserRedemptionRecord := recordstypes.UserRedemptionRecord{
		Id:             id,
		Sender:         sender,
		Receiver:       receiver,
		Amount:         sdkmath.NewInt(1),
		Denom:          denom,
		HostZoneId:     hostZoneId,
		EpochNumber:    epochNumber,
		ClaimIsPending: claimIsPending,
	}

	actualNewUserRedemptionRecord := convertToNewUserRedemptionRecord(oldUserRedemptionRecord)
	require.Equal(t, expectedNewUserRedemptionRecord, actualNewUserRedemptionRecord)
}
