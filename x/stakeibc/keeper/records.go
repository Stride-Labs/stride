package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v26/utils"
	recordstypes "github.com/Stride-Labs/stride/v26/x/records/types"
)

// Create a new deposit record for each host zone for the given epoch
func (k Keeper) CreateDepositRecordsForEpoch(ctx sdk.Context, epochNumber uint64) {
	k.Logger(ctx).Info(fmt.Sprintf("Creating Deposit Records for Epoch %d", epochNumber))

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Creating Deposit Record"))

		depositRecord := recordstypes.DepositRecord{
			Amount:                  sdkmath.ZeroInt(),
			Denom:                   hostZone.HostDenom,
			HostZoneId:              hostZone.ChainId,
			Status:                  recordstypes.DepositRecord_TRANSFER_QUEUE,
			DepositEpochNumber:      epochNumber,
			DelegationTxsInProgress: 0,
		}
		k.RecordsKeeper.AppendDepositRecord(ctx, depositRecord)
	}
}

// Creates a new epoch unbonding record for the epoch
func (k Keeper) CreateEpochUnbondingRecord(ctx sdk.Context, epochNumber uint64) bool {
	k.Logger(ctx).Info(fmt.Sprintf("Creating Epoch Unbonding Records for Epoch %d", epochNumber))

	hostZoneUnbondings := []*recordstypes.HostZoneUnbonding{}

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Creating Epoch Unbonding Record"))

		hostZoneUnbonding := recordstypes.HostZoneUnbonding{
			NativeTokenAmount:         sdkmath.ZeroInt(),
			StTokenAmount:             sdkmath.ZeroInt(),
			StTokensToBurn:            sdkmath.ZeroInt(),
			NativeTokensToUnbond:      sdkmath.ZeroInt(),
			ClaimableNativeTokens:     sdkmath.ZeroInt(),
			Denom:                     hostZone.HostDenom,
			HostZoneId:                hostZone.ChainId,
			Status:                    recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,
			UndelegationTxsInProgress: 0,
		}
		hostZoneUnbondings = append(hostZoneUnbondings, &hostZoneUnbonding)
	}

	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		EpochNumber:        cast.ToUint64(epochNumber),
		HostZoneUnbondings: hostZoneUnbondings,
	}
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return true
}

// Deletes any epoch unbonding records that have had all unbondings claimed
func (k Keeper) CleanupEpochUnbondingRecords(ctx sdk.Context, epochNumber uint64) {
	k.Logger(ctx).Info("Cleaning Claimed Epoch Unbonding Records...")

	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		shouldDeleteEpochUnbondingRecord := true
		hostZoneUnbondings := epochUnbondingRecord.HostZoneUnbondings

		for _, hostZoneUnbonding := range hostZoneUnbondings {
			// if an EpochUnbondingRecord has any HostZoneUnbonding with non-zero balances, we don't delete the EpochUnbondingRecord
			// because it has outstanding tokens that need to be claimed
			notClaimable := hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_CLAIMABLE
			hasUnclaimedTokens := !hostZoneUnbonding.ClaimableNativeTokens.Equal(sdkmath.ZeroInt())
			if notClaimable || hasUnclaimedTokens {
				shouldDeleteEpochUnbondingRecord = false
				break
			}
		}
		if shouldDeleteEpochUnbondingRecord {
			k.RecordsKeeper.RemoveEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber)
		}
	}
}
