package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v18/utils"
	recordstypes "github.com/Stride-Labs/stride/v18/x/records/types"
	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
)

// Iterate each deposit record marked DELEGATION_QUEUE and use the delegation ICA to delegate on the host zone
func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	k.Logger(ctx).Info("Staking deposit records...")

	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isStakeRecord := record.Status == recordstypes.DepositRecord_DELEGATION_QUEUE
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		return isStakeRecord && isBeforeCurrentEpoch
	})

	if len(stakeDepositRecords) == 0 {
		k.Logger(ctx).Info("No deposit records in state DELEGATION_QUEUE")
		return
	}

	// limit the number of staking deposits to process per epoch
	maxDepositRecordsToStake := utils.Min(len(stakeDepositRecords), cast.ToInt(k.GetParam(ctx, types.KeyMaxStakeICACallsPerEpoch)))
	if maxDepositRecordsToStake < len(stakeDepositRecords) {
		k.Logger(ctx).Info(fmt.Sprintf("  MaxStakeICACallsPerEpoch limit reached - Only staking %d out of %d deposit records", maxDepositRecordsToStake, len(stakeDepositRecords)))
	}

	for _, depositRecord := range stakeDepositRecords[:maxDepositRecordsToStake] {
		if depositRecord.Amount.IsZero() {
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId,
			"Processing deposit record %d: %v%s", depositRecord.Id, depositRecord.Amount, depositRecord.Denom))

		hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
		if !hostZoneFound {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Host zone not found for deposit record {%d}", depositRecord.Id))
			continue
		}

		if hostZone.Halted {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Host zone halted for deposit record {%d}", depositRecord.Id))
			continue
		}

		if hostZone.DelegationIcaAddress == "" {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
			continue
		}

		k.Logger(ctx).Info(utils.LogWithHostZone(depositRecord.HostZoneId, "Staking %v%s", depositRecord.Amount, hostZone.HostDenom))
		stakeAmount := sdk.NewCoin(hostZone.HostDenom, depositRecord.Amount)

		err := k.DelegateOnHost(ctx, hostZone, stakeAmount, depositRecord)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s | err: %s", stakeAmount.String(), hostZone.ChainId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Successfully submitted stake"))

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("hostZone", hostZone.ChainId),
				sdk.NewAttribute("newAmountStaked", depositRecord.Amount.String()),
			),
		)
	}
}
