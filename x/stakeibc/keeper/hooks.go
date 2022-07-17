package keeper

import (
	"fmt"
	"strconv"

	utils "github.com/Stride-Labs/stride/utils"
	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v3/modules/core/02-client/types"
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

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch start %s %d", epochIdentifier, epochNumber))

	epochTracker := types.EpochTracker{
		EpochIdentifier: epochIdentifier,
		EpochNumber:     uint64(epochNumber),
	}
	// deposit records *must* exist for this epoch
	k.SetEpochTracker(ctx, epochTracker)

	// process redemption records
	if epochIdentifier == "day" {
		// here, we process everything we need to for redemptions
		k.Logger(ctx).Info(fmt.Sprintf("Day %d Beginning", epochNumber))
		// first we initiate unbondings from any hostZone where it's appropriate
		k.InitiateAllHostZoneUnbondings(ctx, uint64(epochNumber))
		// then we check previous epochs to see if unbondings finished, and sweep the tokens if so
		k.SweepAllUnbondedTokens(ctx)
		// then we cleanup any records that are no longer needed
		k.CleanupEpochUnbondingRecords(ctx)
		// lastly we create an empty unbonding record for this epoch
		k.CreateEpochUnbondings(ctx, epochNumber)
	}

	if epochIdentifier == epochstypes.STRIDE_EPOCH {
		k.Logger(ctx).Info(fmt.Sprintf("Stride Epoch %d", epochNumber))

		// NOTE: We could nest this under `if epochNumber%depositInterval == 0 {`
		// -- should we?
		// e.g. CreateDepositRecordsForDepositInterval
		// Imagine it will be slightly cleaner to track state by epoch, rather than
		// by DepositInterval
		if epochNumber < 0 {
			k.Logger(ctx).Error(fmt.Sprintf("Stride Epoch %d negative", epochNumber))
			return
		}

		k.Logger(ctx).Info("Triggering deposits")
		// Create a new deposit record for each host zone for the upcoming epoch
		k.CreateDepositRecordsForEpoch(ctx, epochNumber)

		k.SetWithdrawalAddress(ctx)

		depositRecords := k.RecordsKeeper.GetAllDepositRecord(ctx)

		// Update the redemption rate
		redemptionRateInterval := int64(k.GetParam(ctx, types.KeyDepositInterval))
		if epochNumber%redemptionRateInterval == 0 {
			k.Logger(ctx).Info("Triggeting update redemption rate")
			k.UpdateRedemptionRates(ctx, depositRecords)
		}

		depositInterval := int64(k.GetParam(ctx, types.KeyDepositInterval))
		if epochNumber%depositInterval == 0 {
			// process previous deposit records
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
		delegationInterval := int64(k.GetParam(ctx, types.KeyDelegateInterval))
		if epochNumber%delegationInterval == 0 {
			k.StakeExistingDepositsOnHostZones(ctx, epochNumber, depositRecords)
		}

		reinvestInterval := int64(k.GetParam(ctx, types.KeyReinvestInterval))
		if epochNumber%reinvestInterval == 0 { // allow a few blocks from UpdateUndelegatedBal to avoid conflicts
			for _, hz := range k.GetAllHostZone(ctx) {
				if (&hz).WithdrawalAccount != nil { // only process host zones once withdrawal accounts are registered

					// read clock time on host zome
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
				}
			}
		}
	}
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
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
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}

// -------------------- helper functions --------------------
func (k Keeper) CreateDepositRecordsForEpoch(ctx sdk.Context, epochNumber int64) {
	// Create one new deposit record / host zone for the next epoch
	createDepositRecords := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		// create a deposit record / host zone
		depositRecord := recordstypes.DepositRecord{
			Id:                 0,
			Amount:             0,
			Denom:              zoneInfo.HostDenom,
			HostZoneId:         zoneInfo.ChainId,
			Status:             recordstypes.DepositRecord_TRANSFER,
			DepositEpochNumber: uint64(epochNumber),
		}
		k.RecordsKeeper.AppendDepositRecord(ctx, depositRecord)
		return nil
	}
	// Iterate the zones and apply icaReinvest
	k.IterateHostZones(ctx, createDepositRecords)
}

func (k Keeper) SetWithdrawalAddress(ctx sdk.Context) {
	setWithdrawalAddresses := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info("\tsetting withdrawal addresses on host zones")
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

func (k Keeper) StakeExistingDepositsOnHostZones(ctx sdk.Context, epochNumber int64, depositRecords []recordstypes.DepositRecord) {
	stakeDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_STAKE
	})
	for _, depositRecord := range stakeDepositRecords {
		if depositRecord.DepositEpochNumber < uint64(epochNumber) {
			pstr := fmt.Sprintf("\t[STAKE] Processing deposit ID:{%d} DENOM:{%s} AMT:{%d}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
			k.Logger(ctx).Info(pstr)
			hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
			if !hostZoneFound {
				k.Logger(ctx).Error("[STAKE] Host zone not found for deposit record {%d}", depositRecord.Id)
				continue
			}
			delegateAccount := hostZone.GetDelegationAccount()
			if delegateAccount == nil || delegateAccount.GetAddress() == "" {
				k.Logger(ctx).Error("[STAKE] Zone %s is missing a delegation address!", hostZone.ChainId)
				continue
			}
			k.Logger(ctx).Info(fmt.Sprintf("\tdelegation staking on %s", hostZone.HostDenom))
			processAmount := utils.Int64ToCoinString(depositRecord.Amount, hostZone.HostDenom)
			amt, err := sdk.ParseCoinNormalized(processAmount)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Could not process coin %s: %s", hostZone.HostDenom, err))
				return
			}
			err = k.DelegateOnHost(ctx, hostZone, amt)
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Did not stake %s on %s", processAmount, hostZone.ChainId))
				return
			} else {
				k.Logger(ctx).Info(fmt.Sprintf("Successfully submitted stake for %s on %s", processAmount, hostZone.ChainId))
			}

			ctx.EventManager().EmitEvents(sdk.Events{
				sdk.NewEvent(
					sdk.EventTypeMessage,
					sdk.NewAttribute("hostZone", hostZone.ChainId),
					sdk.NewAttribute("newAmountStaked", strconv.FormatInt(depositRecord.Amount, 10)),
				),
			})
		}
	}
}

func (k Keeper) TransferExistingDepositsToHostZones(ctx sdk.Context, epochNumber int64, depositRecords []recordstypes.DepositRecord) {
	transferDepositRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_TRANSFER
	})
	addr := k.accountKeeper.GetModuleAccount(ctx, types.ModuleName).GetAddress().String()
	var emptyRecords []uint64
	for _, depositRecord := range transferDepositRecords {
		if depositRecord.DepositEpochNumber < uint64(epochNumber) {
			pstr := fmt.Sprintf("\t[TRANSFER] Processing deposits {%d} {%s} {%d}", depositRecord.Id, depositRecord.Denom, depositRecord.Amount)
			k.Logger(ctx).Info(pstr)

			// skip empty records
			if depositRecord.Amount <= 0 {
				k.Logger(ctx).Info("[TRANSFER] Empty deposit record! Skipping")
				emptyRecords = append(emptyRecords, depositRecord.Id)
				continue
			}

			hostZone, hostZoneFound := k.GetHostZone(ctx, depositRecord.HostZoneId)
			if !hostZoneFound {
				k.Logger(ctx).Error("[TRANSFER] Host zone not found for deposit record {%d}", depositRecord.Id)
				continue
			}
			delegateAccount := hostZone.GetDelegationAccount()
			if delegateAccount == nil || delegateAccount.GetAddress() == "" {
				k.Logger(ctx).Error("[TRANSFER] Zone %s is missing a delegation address!", hostZone.ChainId)
				continue
			}
			delegateAddress := delegateAccount.GetAddress()
			// TODO(TEST-89): Set NewHeight relative to the most recent known gaia height (based on the LC)
			// TODO(TEST-90): why do we have two gaia LCs?
			timeoutHeight := clienttypes.NewHeight(0, 1000000)
			transferCoin := sdk.NewCoin(hostZone.GetIBCDenom(), sdk.NewInt(int64(depositRecord.Amount)))
			goCtx := sdk.WrapSDKContext(ctx)

			msg := ibctypes.NewMsgTransfer("transfer", hostZone.TransferChannelId, transferCoin, addr, delegateAddress, timeoutHeight, 0)
			k.Logger(ctx).Info("TransferExistingDepositsToHostZones msg:", msg)
			_, err := k.TransferKeeper.Transfer(goCtx, msg)
			if err != nil {
				k.Logger(ctx).Error("\t[TRANSFER] ERROR WITH DEPOSIT RECEIPT", hostZone.TransferChannelId, transferCoin, addr, delegateAddress, timeoutHeight)
				pstr = fmt.Sprintf("\t[TRANSFER] ERROR WITH DEPOSIT RECEIPT {%v}", err)
				k.Logger(ctx).Error(pstr)
				return
			}
		}
	}
	// clear empty records
	for _, recordId := range emptyRecords {
		k.RecordsKeeper.RemoveDepositRecord(ctx, recordId)
	}
}

func (k Keeper) UpdateRedemptionRates(ctx sdk.Context, depositRecords []recordstypes.DepositRecord) {
	// Calc redemptionRate for each host zone
	UpdateRedemptionRate := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {

		undelegatedBalance, error := k.GetUndelegatedBalance(ctx, zoneInfo, depositRecords)
		if error != nil {
			return error
		}
		stakedBalance := zoneInfo.StakedBal
		modeuleAcctBalance, error := k.GetModuleAccountBalance(ctx, zoneInfo, depositRecords)
		if error != nil {
			return error
		}
		stSupply := k.bankKeeper.GetSupply(ctx, types.StAssetDenomFromHostZoneDenom(zoneInfo.HostDenom)).Amount.Int64()
		if stSupply == 0 {
			return fmt.Errorf("stSupply is 0")
		}

		// calc redemptionRate = (UB+SB+MA)/stSupply
		redemptionRate := (sdk.NewDec(undelegatedBalance).Add(sdk.NewDec(stakedBalance)).Add(sdk.NewDec(modeuleAcctBalance))).Quo(sdk.NewDec(stSupply))
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

func (k Keeper) GetUndelegatedBalance(ctx sdk.Context, hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) (int64, error) {
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

func (k Keeper) GetModuleAccountBalance(ctx sdk.Context, hostZone types.HostZone, depositRecords []recordstypes.DepositRecord) (int64, error) {
	// filter to only the deposit records for the host zone with status DELEGATION
	ModuleAccountRecords := utils.FilterDepositRecords(depositRecords, func(record recordstypes.DepositRecord) (condition bool) {
		return record.Status == recordstypes.DepositRecord_TRANSFER && record.HostZoneId == hostZone.ChainId
	})

	// sum the amounts of the deposit records
	var totalAmount int64
	for _, depositRecord := range ModuleAccountRecords {
		totalAmount += depositRecord.Amount
	}

	return totalAmount, nil
}
