package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/utils"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k Keeper) CreateDepositRecordsForEpoch(ctx sdk.Context, epochNumber uint64) {
	// Create one new deposit record / host zone for the next epoch
	createDepositRecords := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("createDepositRecords, index: %d, zoneInfo: %s", index, zoneInfo.ConnectionId))
		depositRecord := recordstypes.DepositRecord{
			Id:                 0,
			Amount:             0,
			Denom:              zoneInfo.HostDenom,
			HostZoneId:         zoneInfo.ChainId,
			Status:             recordstypes.DepositRecord_TRANSFER,
			DepositEpochNumber: epochNumber,
		}
		k.RecordsKeeper.AppendDepositRecord(ctx, depositRecord)
		return nil
	}
	k.IterateHostZones(ctx, createDepositRecords)
}

func (k Keeper) TransferExistingDepositsToHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	transferDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isTransferRecord := record.Status == recordstypes.DepositRecord_TRANSFER
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		return isTransferRecord && isBeforeCurrentEpoch
	})

	ibcTimeoutBlocks := k.GetParam(ctx, types.KeyIbcTimeoutBlocks)

	for _, depositRecord := range transferDepositRecords {
		pstr := fmt.Sprintf("\t[TransferExistingDepositsToHostZones] Processing deposits {%d} {%s} {%d}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
		k.Logger(ctx).Info(pstr)

		// if a TRANSFER record has 0 balance and was created in the previous epoch, it's safe to remove since it will never be updated or used"
		if depositRecord.Amount <= 0 {
			k.Logger(ctx).Info("[TransferExistingDepositsToHostZones] Empty deposit record (ID: %s)! Removing.", depositRecord.Id)
			k.RecordsKeeper.RemoveDepositRecord(ctx, depositRecord.Id)
			continue
		}

		hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
		if !hostZoneFound {
			k.Logger(ctx).Error(fmt.Sprintf("[TransferExistingDepositsToHostZones] Host zone not found for deposit record id %d", depositRecord.Id))
			continue
		}

		hostZoneModuleAddress := hostZone.GetAddress()
		delegateAccount := hostZone.GetDelegationAccount()
		if delegateAccount == nil || delegateAccount.GetAddress() == "" {
			k.Logger(ctx).Error(fmt.Sprintf("[TransferExistingDepositsToHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
			continue
		}
		delegateAddress := delegateAccount.GetAddress()

		blockHeight, err := k.GetLightClientHeightSafely(ctx, hostZone.ConnectionId)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not find blockHeight for host zone %s, aborting transfers to host zone this epoch", hostZone.ConnectionId))
			continue
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("Found blockHeight for host zone %s: %d", hostZone.ConnectionId, blockHeight))
		}

		timeoutHeight := clienttypes.NewHeight(0, blockHeight+ibcTimeoutBlocks)
		transferCoin := sdk.NewCoin(hostZone.GetIBCDenom(), sdk.NewInt(depositRecord.Amount))
		// QUESTION: Is there a good place to store "tranfer" as a constant?
		msg := ibctypes.NewMsgTransfer("transfer", hostZone.TransferChannelId, transferCoin, hostZoneModuleAddress, delegateAddress, timeoutHeight, 0)
		k.Logger(ctx).Info(fmt.Sprintf("TransferExistingDepositsToHostZones msg %v", msg))

		err = k.RecordsKeeper.Transfer(ctx, msg, depositRecord.Id)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] ERROR WITH DEPOSIT RECEIPT %s %v %s %s %v", hostZone.TransferChannelId, transferCoin, hostZoneModuleAddress, delegateAddress, timeoutHeight))
			k.Logger(ctx).Error(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] err {%s}", err.Error()))
			continue
		}
	}
}

func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		isStakeRecord := record.Status == recordstypes.DepositRecord_STAKE
		isBeforeCurrentEpoch := record.DepositEpochNumber < epochNumber
		return isStakeRecord && isBeforeCurrentEpoch
	})

	// limit the number of staking deposits to process per epoch
	maxDepositRecordsToStake := utils.Min(len(stakeDepositRecords), cast.ToInt(k.GetParam(ctx, types.KeyMaxStakeICACallsPerEpoch)))
	k.Logger(ctx).Info(fmt.Sprintf("Staking %d out of %d deposit records", maxDepositRecordsToStake, len(stakeDepositRecords)))

	for _, depositRecord := range stakeDepositRecords[:maxDepositRecordsToStake] {
		k.Logger(ctx).Info(fmt.Sprintf("\t[StakeExistingDepositsOnHostZones] Processing deposit ID:{%d} DENOM:{%s} AMT:{%d}",
			depositRecord.Id, depositRecord.Denom, depositRecord.Amount))

		hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
		if !hostZoneFound {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Host zone not found for deposit record {%d}", depositRecord.Id))
			continue
		}

		delegateAccount := hostZone.GetDelegationAccount()
		if delegateAccount == nil || delegateAccount.GetAddress() == "" {
			k.Logger(ctx).Error(fmt.Sprintf("[StakeExistingDepositsOnHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
			continue
		}

		k.Logger(ctx).Info(fmt.Sprintf("\t[StakeExistingDepositsOnHostZones] Staking %d on %s", depositRecord.Amount, hostZone.HostDenom))
		stakeAmount := sdk.NewCoin(hostZone.HostDenom, sdk.NewInt(depositRecord.Amount))

		err := k.DelegateOnHost(ctx, hostZone, stakeAmount, depositRecord.Id)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s | err: %s", stakeAmount.String(), hostZone.ChainId, err.Error()))
			continue
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("Successfully submitted stake for %s on %s", stakeAmount.String(), hostZone.ChainId))
		}

		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				sdk.EventTypeMessage,
				sdk.NewAttribute("hostZone", hostZone.ChainId),
				sdk.NewAttribute("newAmountStaked", strconv.FormatInt(depositRecord.Amount, 10)),
			),
		)
	}
}
