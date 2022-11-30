package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v3/utils"
	epochstypes "github.com/Stride-Labs/stride/v3/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/v3/x/records/types"
	"github.com/Stride-Labs/stride/v3/x/stakeibc/types"
)

// TODO [TEST-127]: ensure all timeouts are less than the epoch length
// TODO [TEST-126]: add events from event manager, e.g.
// ctx.EventManager().EmitEvents(sdk.Events{
// 	sdk.NewEvent(
// 		sdk.EventTypeMessage,
// 		sdk.NewAttribute("hostZone", zoneInfo.ChainId),
// 		sdk.NewAttribute("newAmountStaked", balance.String()),
// 	),
// })

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	// every epoch
	epochIdentifier := epochInfo.Identifier
	epochNumber, err := cast.ToUint64E(epochInfo.CurrentEpoch)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch number to uint64: %v", err))
		return
	}

	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch start %s %d", epochIdentifier, epochNumber))
	k.Logger(ctx).Info(fmt.Sprintf("Epoch start time %d", epochInfo.GetCurrentEpochStartTime().UnixNano()))

	ns, err := cast.ToUint64E(epochInfo.GetDuration().Nanoseconds())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return
	}
	nextEpochStartTime, err := cast.ToUint64E(epochInfo.GetCurrentEpochStartTime().Add(epochInfo.GetDuration()).UnixNano())
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Could not convert epoch duration to uint64: %v", err))
		return
	}
	epochTracker := types.EpochTracker{
		EpochIdentifier:    epochIdentifier,
		EpochNumber:        epochNumber,
		Duration:           ns,
		NextEpochStartTime: nextEpochStartTime,
	}
	// deposit records *must* exist for this epoch
	k.Logger(ctx).Info(fmt.Sprintf("Setting epochTracker %v", epochTracker))
	k.SetEpochTracker(ctx, epochTracker)

	// process redemption records
	if epochIdentifier == "day" {
		// here, we process everything we need to for redemptions
		k.Logger(ctx).Info(fmt.Sprintf("Day %d Beginning", epochNumber))
		// first we initiate unbondings from any hostZone where it's appropriate
		k.Logger(ctx).Info("InitiateAllHostZoneUnbondings")
		k.InitiateAllHostZoneUnbondings(ctx, epochNumber)
		// then we check previous epochs to see if unbondings finished, and sweep the tokens if so
		k.Logger(ctx).Info("SweepAllUnbondedTokens")
		k.SweepAllUnbondedTokens(ctx)
		// then we cleanup any records that are no longer needed
		k.Logger(ctx).Info("CleanupEpochUnbondingRecords")
		k.CleanupEpochUnbondingRecords(ctx, epochNumber)
		// lastly we create an empty unbonding record for this epoch
		k.Logger(ctx).Info("CreateEpochUnbondingRecord")
		k.CreateEpochUnbondingRecord(ctx, epochNumber)
	}

	if epochIdentifier == epochstypes.STRIDE_EPOCH {
		k.Logger(ctx).Info(fmt.Sprintf("Stride Epoch %d", epochNumber))

		// Create a new deposit record for each host zone for the upcoming epoch
		k.CreateDepositRecordsForEpoch(ctx, epochNumber)

		// TODO: move this to an external function that anyone can call, so that we don't have to call it every epoch
		k.SetWithdrawalAddress(ctx)

		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)

		// Update the redemption rate
		redemptionRateInterval, err := cast.ToUint64E(k.GetParam(ctx, types.KeyRedemptionRateInterval))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not convert redemptionRateInterval to uint64: %v", err))
			return
		}
		if epochNumber%redemptionRateInterval == 0 {
			k.Logger(ctx).Info("Triggering update redemption rate")
			k.UpdateRedemptionRates(ctx, depositRecords)
		}

		depositInterval, err := cast.ToUint64E(k.GetParam(ctx, types.KeyDepositInterval))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not convert depositInterval to int64: %v", err))
			return
		}
		if epochNumber%depositInterval == 0 {
			// process previous deposit records
			k.Logger(ctx).Info("TransferExistingDepositsToHostZones")
			k.TransferExistingDepositsToHostZones(ctx, epochNumber, depositRecords)
		}

		delegationInterval, err := cast.ToUint64E(k.GetParam(ctx, types.KeyDelegateInterval))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not convert delegationInterval to int64: %v", err))
			return
		}
		if epochNumber%delegationInterval == 0 {
			k.Logger(ctx).Info("StakeExistingDepositsOnHostZones")
			k.StakeExistingDepositsOnHostZones(ctx, epochNumber, depositRecords)
		}

		reinvestInterval, err := cast.ToUint64E(k.GetParam(ctx, types.KeyReinvestInterval))
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not convert reinvestInterval to int64: %v", err))
			return
		}
		if epochNumber%reinvestInterval == 0 { // allow a few blocks from UpdateUndelegatedBal to avoid conflicts
			k.Logger(ctx).Info("Reinvesting tokens")
			for _, hz := range k.GetAllHostZone(ctx) {
				// only process host zones once withdrawal accounts are registered
				withdrawalIca := hz.GetWithdrawalAccount()
				if withdrawalIca != nil {
					// read clock time on host zone
					blockTime, err := k.GetLightClientTimeSafely(ctx, hz.ConnectionId)
					if err != nil {
						k.Logger(ctx).Error(fmt.Sprintf("Could not find blockTime for host zone %s, err: %s", hz.ConnectionId, err.Error()))
						continue
					} else {
						k.Logger(ctx).Info(fmt.Sprintf("Found blockTime for host zone %s: %d", hz.ConnectionId, blockTime))
					}

					err = k.UpdateWithdrawalBalance(ctx, hz)
					if err != nil {
						k.Logger(ctx).Error(fmt.Sprintf("Error updating withdrawal balance for host zone %s: %s", hz.ConnectionId, err.Error()))
						continue
					} else {
						k.Logger(ctx).Info(fmt.Sprintf("Updated withdrawal balance for host zone %s", hz.ConnectionId))
					}
				} else {
					k.Logger(ctx).Info(fmt.Sprintf("Withdrawal account not registered for host zone %s", hz.ChainId))
				}
			}
		}
	}
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	// every epoch
	epochIdentifier := epochInfo.Identifier
	epochNumber := epochInfo.CurrentEpoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch end %s %d", epochIdentifier, epochNumber))
	if epochIdentifier == "day" {
		k.Logger(ctx).Info(fmt.Sprintf("Day %d Ending", epochNumber))
	}
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.AfterEpochEnd(ctx, epochInfo)
}

// Set the withdrawal account address for each host zone
func (k Keeper) SetWithdrawalAddress(ctx sdk.Context) {
	k.Logger(ctx).Info("Setting Withdrawal Addresses...")

	setWithdrawalAddresses := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		err := k.SetWithdrawalAddressOnHost(ctx, hostZone)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to set withdrawal address to %s on %s, err: %s", hostZone.WithdrawalAccount.Address, hostZone.ChainId, err))
		}
		return nil
	}

	k.IterateHostZones(ctx, setWithdrawalAddresses)
}

func (k Keeper) UpdateRedemptionRates(ctx sdk.Context, depositRecords []recordstypes.DepositRecord) {
	// Calc redemptionRate for each host zone
	UpdateRedemptionRate := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("index: %d, zoneInfo: %s", index, zoneInfo.ChainId))

		undelegatedBalance, error := k.GetUndelegatedBalance(zoneInfo, depositRecords)
		if error != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not get undelegated balance for host zone %s: %s", zoneInfo.ChainId, error.Error()))
			return error
		}
		k.Logger(ctx).Info(fmt.Sprintf("undelegatedBalance: %d", undelegatedBalance))
		stakedBalance, err := cast.ToInt64E(zoneInfo.GetStakedBal())
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not get staked balance for host zone %s: %s", zoneInfo.ChainId, err.Error()))
			return err
		}
		k.Logger(ctx).Info(fmt.Sprintf("stakedBalance: %d", stakedBalance))
		moduleAcctBalance, error := k.GetModuleAccountBalance(zoneInfo, depositRecords)
		if error != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Could not get module account balance for host zone %s: %s", zoneInfo.ChainId, error.Error()))
			return error
		}
		k.Logger(ctx).Info(fmt.Sprintf("moduleAcctBalance: %d", moduleAcctBalance))
		stSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(zoneInfo.HostDenom)).Amount.Int64()
		if stSupply == 0 {
			k.Logger(ctx).Info(fmt.Sprintf("stSupply: %d", stSupply))
			return nil
		}
		k.Logger(ctx).Info(fmt.Sprintf("stSupply: %d", stSupply))

		// calc redemptionRate = (UB+SB+MA)/stSupply
		k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION-RATE] undelegatedBalance: %d, stakedBalance: %d, moduleAcctBalance: %d, stSupply: %d", undelegatedBalance, stakedBalance, moduleAcctBalance, stSupply))
		redemptionRate := (sdk.NewDec(undelegatedBalance).Add(sdk.NewDec(stakedBalance)).Add(sdk.NewDec(moduleAcctBalance))).Quo(sdk.NewDec(stSupply))
		k.Logger(ctx).Info(fmt.Sprintf("[REDEMPTION-RATE] New Rate is %d (vs prev %d)", redemptionRate, zoneInfo.LastRedemptionRate))

		// set redemptionRate attribute for the hostZone (and update last RedemptionRate)
		zoneInfo.LastRedemptionRate = zoneInfo.RedemptionRate
		zoneInfo.RedemptionRate = redemptionRate
		k.SetHostZone(ctx, zoneInfo)

		return nil
	}
	// Iterate the zones and apply icaReinvest
	k.IterateHostZones(ctx, UpdateRedemptionRate)
}

func (k Keeper) GetUndelegatedBalance(hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) (int64, error) {
	// filter to only the deposit records for the host zone with status DELEGATION_QUEUE
	UndelegatedDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return ((record.Status == recordstypes.DepositRecord_DELEGATION_QUEUE || record.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS) && record.HostZoneId == hostZone.ChainId)
	})

	// sum the amounts of the deposit records
	var totalAmount int64
	for _, depositRecord := range UndelegatedDepositRecords {
		totalAmount += depositRecord.Amount
	}

	return totalAmount, nil
}

func (k Keeper) GetModuleAccountBalance(hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) (int64, error) {
	// filter to only the deposit records for the host zone with status DELEGATION
	ModuleAccountRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return (record.Status == recordstypes.DepositRecord_TRANSFER_QUEUE || record.Status == recordstypes.DepositRecord_TRANSFER_IN_PROGRESS) && record.HostZoneId == hostZone.ChainId
	})

	// sum the amounts of the deposit records
	totalAmount := int64(0)
	for _, depositRecord := range ModuleAccountRecords {
		totalAmount += depositRecord.Amount
	}

	return totalAmount, nil
}
