package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/utils"
	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
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
		k.CleanupEpochUnbondingRecords(ctx)
		// lastly we create an empty unbonding record for this epoch
		k.Logger(ctx).Info("CreateEpochUnbondingRecord")
		k.CreateEpochUnbondingRecord(ctx, epochNumber)
	}

	if epochIdentifier == epochstypes.STRIDE_EPOCH {
		k.Logger(ctx).Info(fmt.Sprintf("Stride Epoch %d", epochNumber))

		// NOTE: We could nest this under `if epochNumber%depositInterval == 0 {`
		// -- should we?
		// e.g. CreateDepositRecordsForDepositInterval
		// Imagine it will be slightly cleaner to track state by epoch, rather than
		// by DepositInterval

		// Create a new deposit record for each host zone for the upcoming epoch
		k.Logger(ctx).Info("CreateDepositRecordsForEpoch")
		k.CreateDepositRecordsForEpoch(ctx, epochNumber)

		k.Logger(ctx).Info("SetWithdrawalAddress")
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

		// NOTE: the stake ICA timeout *must* be l.t. the staking epoch length, otherwise
		// we could send a stake ICA call (which could succeed), without deleting the record.
		// This could happen if the ack doesn't return by the next epoch. We would then send
		// *another* stake ICA call, for a portion of the balance which has *already* been staked,
		// which is very bad! This could result in the protocol becoming insolvent, by staking balances
		// that were earmarked for another purpose, e.g. redemptions.
		// The same holds true for IBC transfers.
		// Given these assumptions, the order of staking / transfers is not important, because stride deposit
		// records always accurately reflect the state of the controller / host chain by the next epoch.
		// Put another way, all outstanding ICA calls / IBC transfers must be settled on the controller
		// chain before the next epoch begins.
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
				if (&hz).WithdrawalAccount != nil { // only process host zones once withdrawal accounts are registered

					// read clock time on host zone
					// k.ReadClockTime(ctx, hz)
					blockTime, found := k.GetLightClientTimeSafely(ctx, hz.ConnectionId)
					if !found {
						k.Logger(ctx).Error(fmt.Sprintf("Could not find blockTime for host zone %s", hz.ConnectionId))
						continue
					} else {
						k.Logger(ctx).Info(fmt.Sprintf("Found blockTime for host zone %s: %d", hz.ConnectionId, blockTime))
					}

					err := k.UpdateWithdrawalBalance(ctx, hz)
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

// -------------------- helper functions --------------------
func (k Keeper) CreateDepositRecordsForEpoch(ctx sdk.Context, epochNumber uint64) {
	// Create one new deposit record / host zone for the next epoch
	createDepositRecords := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("createDepositRecords, index: %d, zoneInfo: %s", index, zoneInfo.ConnectionId))
		// create a deposit record / host zone
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
	// Iterate the zones and apply icaReinvest
	k.IterateHostZones(ctx, createDepositRecords)
}

func (k Keeper) SetWithdrawalAddress(ctx sdk.Context) {
	setWithdrawalAddresses := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("\tsetting withdrawal address for index %v, zoneInfo %v", index, zoneInfo))
		err := k.SetWithdrawalAddressOnHost(ctx, zoneInfo)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Did not set withdrawal address to %s on %s", zoneInfo.GetWithdrawalAccount().GetAddress(), zoneInfo.GetChainId()))
			k.Logger(ctx).Error(fmt.Sprintf("Withdrawal address setting error: %v", err))
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("Successfully set withdrawal address to %s on %s", zoneInfo.GetWithdrawalAccount().GetAddress(), zoneInfo.GetChainId()))
		}
		return nil
	}
	k.IterateHostZones(ctx, setWithdrawalAddresses)
}

func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_STAKE
	})
	for _, depositRecord := range stakeDepositRecords {
		if depositRecord.DepositEpochNumber < cast.ToUint64(epochNumber) {
			pstr := fmt.Sprintf("\t[StakeExistingDepositsOnHostZones] Processing deposit ID:{%d} DENOM:{%s} AMT:{%d}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
			k.Logger(ctx).Info(pstr)
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
			k.Logger(ctx).Info(fmt.Sprintf("\tdelegation staking on %s", hostZone.HostDenom))
			processAmount := utils.Int64ToCoinString(depositRecord.Amount, hostZone.HostDenom)
			amt, err := sdk.ParseCoinNormalized(processAmount)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Could not process coin %s: %s", hostZone.HostDenom, err.Error()))
				return
			}
			err = k.DelegateOnHost(ctx, hostZone, amt, depositRecord.Id)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s | err: %s", processAmount, hostZone.ChainId, err.Error()))
				return
			} else {
				k.Logger(ctx).Info(fmt.Sprintf("Successfully submitted stake for %s on %s", processAmount, hostZone.ChainId))
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
}

func (k Keeper) TransferExistingDepositsToHostZones(ctx sdk.Context, epochNumber uint64, depositRecords []recordstypes.DepositRecord) {
	transferDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_TRANSFER
	})
	ibcTimeoutBlocks := k.GetParam(ctx, types.KeyIbcTimeoutBlocks)
	addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress().String()
	var emptyRecords []uint64
	for _, depositRecord := range transferDepositRecords {
		if depositRecord.DepositEpochNumber < cast.ToUint64(epochNumber) {
			pstr := fmt.Sprintf("\t[TransferExistingDepositsToHostZones] Processing deposits {%d} {%s} {%d}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
			k.Logger(ctx).Info(pstr)

			// skip empty records
			if depositRecord.Amount <= 0 {
				k.Logger(ctx).Info("[TransferExistingDepositsToHostZones] Empty deposit record! Skipping")
				emptyRecords = append(emptyRecords, depositRecord.Id)
				continue
			}

			hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
			if !hostZoneFound {
				k.Logger(ctx).Error(fmt.Sprintf("[TransferExistingDepositsToHostZones] Host zone not found for deposit record id %d", depositRecord.Id))
				continue
			}
			delegateAccount := hostZone.GetDelegationAccount()
			if delegateAccount == nil || delegateAccount.GetAddress() == "" {
				k.Logger(ctx).Error(fmt.Sprintf("[TransferExistingDepositsToHostZones] Zone %s is missing a delegation address!", hostZone.ChainId))
				continue
			}
			delegateAddress := delegateAccount.GetAddress()
			// TODO(TEST-90): why do we have two gaia LCs?
			blockHeight, found := k.GetLightClientHeightSafely(ctx, hostZone.ConnectionId)
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("Could not find blockHeight for host zone %s, aborting transfers to host zone this epoch", hostZone.ConnectionId))
				continue
			} else {
				k.Logger(ctx).Info(fmt.Sprintf("Found blockHeight for host zone %s: %d", hostZone.ConnectionId, blockHeight))
			}
			timeoutHeight := clienttypes.NewHeight(0, blockHeight+ibcTimeoutBlocks)
			transferCoin := sdk.NewCoin(hostZone.GetIBCDenom(), sdk.NewInt(depositRecord.Amount))
			goCtx := sdk.WrapSDKContext(ctx)

			msg := ibctypes.NewMsgTransfer("transfer", hostZone.TransferChannelId, transferCoin, addr, delegateAddress, timeoutHeight, 0)
			k.Logger(ctx).Info(fmt.Sprintf("TransferExistingDepositsToHostZones msg %v", msg))
			_, err := k.TransferKeeper.Transfer(goCtx, msg)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] ERROR WITH DEPOSIT RECEIPT %s %v %s %s %v", hostZone.TransferChannelId, transferCoin, addr, delegateAddress, timeoutHeight))
				k.Logger(ctx).Error(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] err {%s}", err.Error()))
				return
			}
		}
	}
	// clear empty records
	for _, recordId := range emptyRecords {
		k.Logger(ctx).Info(fmt.Sprintf("\t[TransferExistingDepositsToHostZones] clear empty deposit record record %v", recordId))
		k.RecordsKeeper.RemoveDepositRecord(ctx, recordId)
	}
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
	// filter to only the deposit records for the host zone with status STAKE
	UndelegatedDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_STAKE && record.HostZoneId == hostZone.ChainId
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
		return record.Status == recordstypes.DepositRecord_TRANSFER && record.HostZoneId == hostZone.ChainId
	})

	// sum the amounts of the deposit records
	totalAmount := int64(0)
	for _, depositRecord := range ModuleAccountRecords {
		totalAmount += depositRecord.Amount
	}

	return totalAmount, nil
}
