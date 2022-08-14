package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/utils"

	recordstypes "github.com/Stride-Labs/stride/x/records/types"
	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

func (k Keeper) CreateEpochUnbondingRecord(ctx sdk.Context, epochNumber int64) bool {
	hostZoneUnbondings := []*recordstypes.HostZoneUnbonding{}
	addEpochUndelegation := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		hostZoneUnbonding := recordstypes.HostZoneUnbonding{
			NativeTokenAmount: uint64(0),
			StTokenAmount:     uint64(0),
			Denom:             hostZone.HostDenom,
			HostZoneId:        hostZone.ChainId,
			Status:            recordstypes.HostZoneUnbonding_BONDED,
		}
		k.Logger(ctx).Info(fmt.Sprintf("Adding hostZoneUnbonding %v to %s", hostZoneUnbonding, hostZone.ChainId))
		hostZoneUnbondings = append(hostZoneUnbondings, &hostZoneUnbonding) // [hostZone.ChainId] = &hostZoneUnbonding
		return nil
	}

	k.IterateHostZones(ctx, addEpochUndelegation)
	epochUnbondingRecord := recordstypes.EpochUnbondingRecord{
		EpochNumber:        cast.ToUint64(epochNumber),
		HostZoneUnbondings: hostZoneUnbondings,
	}
	k.Logger(ctx).Info(fmt.Sprintf("AppendEpochUnbondingRecord %v", epochUnbondingRecord))
	k.RecordsKeeper.SetEpochUnbondingRecord(ctx, epochUnbondingRecord)
	return true
}

func (k Keeper) SendHostZoneUnbondings(ctx sdk.Context, hostZone types.HostZone) bool {
	// this function goes and processes all unbonded records for this hostZone
	// regardless of what epoch they belong to
	totalAmtToUnbond := uint64(0)
	epochUnbondingNumbers := []uint64{}
	var msgs []sdk.Msg
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbonding.EpochNumber, hostZone.ChainId)
		if !found {
			errMsg := fmt.Sprintf("Host zone unbonding record not found for hostZoneId %s in epoch %d",
				hostZone.ChainId, epochUnbonding.GetEpochNumber())
			k.Logger(ctx).Error(errMsg)
			continue
		}
		if hostZoneRecord.Status == recordstypes.HostZoneUnbonding_BONDED { // we only send the ICA call if this hostZone hasn't triggered yet
			totalAmtToUnbond += hostZoneRecord.NativeTokenAmount
			epochUnbondingNumbers = append(epochUnbondingNumbers, epochUnbonding.EpochNumber)
		}
	}
	delegationAccount := hostZone.GetDelegationAccount()
	validators := hostZone.GetValidators()
	if totalAmtToUnbond == 0 {
		return true
	}
	// we distribute the unbonding based on our target weights
	newUnbondingToValidator, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalAmtToUnbond)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target val amts for host zone %s %d: %s", hostZone.ChainId, totalAmtToUnbond, err))
		return false
	}
	valAddrToUnbondAmt := make(map[string]int64)
	overflowAmt := int64(0)
	for _, validator := range validators {
		valAddr := validator.GetAddress()
		valUnbondAmt := cast.ToInt64(newUnbondingToValidator[valAddr])
		currentAmtStaked := cast.ToInt64(validator.GetDelegationAmt())
		if valUnbondAmt > currentAmtStaked { // if we don't have enough assets to unbond
			overflowAmt += valUnbondAmt - currentAmtStaked
			valUnbondAmt = currentAmtStaked
		}
		valAddrToUnbondAmt[valAddr] = valUnbondAmt
	}
	if overflowAmt > 0 { // if we need to reallocate any weights
		for _, validator := range validators {
			valAddr := validator.GetAddress()
			valUnbondAmt := valAddrToUnbondAmt[valAddr]
			currentAmtStaked := validator.GetDelegationAmt()
			// store how many more tokens we could unbond, if needed
			amtToPotentiallyUnbond := cast.ToInt64(currentAmtStaked) - valUnbondAmt
			if amtToPotentiallyUnbond > 0 { // if we can afford to unbond more
				if amtToPotentiallyUnbond > overflowAmt { // we can fully cover the overflow
					valAddrToUnbondAmt[valAddr] += overflowAmt
					overflowAmt = 0
					break
				} else {
					valAddrToUnbondAmt[valAddr] += amtToPotentiallyUnbond
					overflowAmt -= amtToPotentiallyUnbond
				}
			}
		}
	}
	if overflowAmt > 0 { // what?? we still can't cover the overflow? something is very wrong
		k.Logger(ctx).Error(fmt.Sprintf("Could not unbond %d on Host Zone %s, unable to balance the unbond amount across validators",
			totalAmtToUnbond, hostZone.ChainId))
		return false
	}
	var splitDelegations []*types.SplitDelegation
	for _, valAddr := range utils.StringToIntMapKeys(valAddrToUnbondAmt) {
		valUnbondAmt := valAddrToUnbondAmt[valAddr]
		stakeAmt := sdk.NewInt64Coin(hostZone.HostDenom, valUnbondAmt)

		msgs = append(msgs, &stakingtypes.MsgUndelegate{
			DelegatorAddress: delegationAccount.GetAddress(),
			ValidatorAddress: valAddr,
			Amount:           stakeAmt,
		})

		splitDelegations = append(splitDelegations, &types.SplitDelegation{
			Validator: valAddr,
			Amount:    stakeAmt.Amount.Uint64(),
		})
	}

	undelegateCallback := types.UndelegateCallback{
		HostZoneId:            hostZone.ChainId,
		SplitDelegations:      splitDelegations,
		EpochUnbondingNumbers: epochUnbondingNumbers,
	}
	marshalledCallbackArgs, err := k.MarshalUndelegateCallbackArgs(ctx, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return false
	}

	_, err = k.SubmitTxsDayEpoch(ctx, hostZone.GetConnectionId(), msgs, *delegationAccount, UNDELEGATE, marshalledCallbackArgs)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error submitting unbonding tx: %s", err))
		return false
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("newAmountUnbonding", strconv.FormatUint(totalAmtToUnbond, 10)),
		),
	)
	return true
}

func (k Keeper) InitiateAllHostZoneUnbondings(ctx sdk.Context, dayNumber uint64) bool {
	// this function goes through each host zone, and if it's the right time to
	// initiate an unbonding, it goes and tries to unbond all outstanding records
	for _, hostZone := range k.GetAllHostZone(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for host zone %s", hostZone.GetChainId()))
		// we only send the ICA call if this hostZone is supposed to be triggered
		if dayNumber%hostZone.UnbondingFrequency == 0 {
			k.Logger(ctx).Info(fmt.Sprintf("Sending unbondings for host zone %s", hostZone.ChainId))
			k.SendHostZoneUnbondings(ctx, hostZone)
		}
	}
	return true
}

func (k Keeper) CleanupEpochUnbondingRecords(ctx sdk.Context) bool {
	// this function goes through each EpochUnbondingRecord
	// if any of them don't have any hostZones, then it deletes the record
	for _, epochUnbondingRecord := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Cleaning up epoch unbondings for epoch unbonding record from epoch %d", epochUnbondingRecord.GetEpochNumber()))
		shouldDeleteRecord := true
		hostZoneUnbondings := epochUnbondingRecord.GetHostZoneUnbondings()
		for _, hostZoneUnbonding := range hostZoneUnbondings {
			k.Logger(ctx).Info(fmt.Sprintf("processing hostZoneUnbonding %v", hostZoneUnbonding))
			if (hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_TRANSFERRED) && (hostZoneUnbonding.GetNativeTokenAmount() != 0) {
				shouldDeleteRecord = false
				break
			}
		}
		if shouldDeleteRecord {
			k.Logger(ctx).Info(fmt.Sprintf("removing EpochUnbondingRecord %v", epochUnbondingRecord.GetEpochNumber()))
			k.RecordsKeeper.RemoveEpochUnbondingRecord(ctx, epochUnbondingRecord.GetEpochNumber())
		}
	}
	return true
}

func (k Keeper) SweepAllUnbondedTokens(ctx sdk.Context) {
	// SweepAllUnbondedTokens iterates host zones and transfers unbonded tokens to the redemption account

	sweepUnbondedTokens := func(ctx sdk.Context, index int64, zoneInfo types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("sweepUnbondedTokens for host zone %s", zoneInfo.ChainId))

		// get latest epoch unbonding record
		unbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
		totalAmtTransferToRedemptionAcct := uint64(0)
		for _, unbondingRecord := range unbondingRecords {
			k.Logger(ctx).Info(fmt.Sprintf("processing unbondingRecord %v", unbondingRecord.EpochNumber))

			// iterate through all host zone unbondings and process them if they're ready to be swept
			unbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, unbondingRecord.EpochNumber, zoneInfo.ChainId)
			if !found {
				return sdkerrors.Wrapf(types.ErrInvalidHostZone, "host zone not found in unbondings: %s", zoneInfo.ChainId)
			}
			k.Logger(ctx).Info(fmt.Sprintf("\tProcessing batch SweepAllUnbondedTokens for host zone %s", zoneInfo.ChainId))
			zone, found := k.GetHostZone(ctx, unbonding.HostZoneId)
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("\t\tHost zone not found for hostZoneId %s", unbonding.HostZoneId))
				continue
			}

			// get latest blockTime from light client
			blockTime, found := k.GetLightClientTimeSafely(ctx, zone.ConnectionId)
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("\t\tCould not find blockTime for host zone %s", zone.ChainId))
				return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "\t\tCould not find blockTime for host zone")
			}

			shouldProcess := (unbonding.Status == recordstypes.HostZoneUnbonding_PENDING_TRANSFER || unbonding.Status == recordstypes.HostZoneUnbonding_UNBONDED)
			// if the unbonding period has elapsed, then we can send the ICA call to sweep this hostZone's unbondings to the rewards account (in a batch)
			k.Logger(ctx).Info(fmt.Sprintf("\tUnbonding time:  %d blockTime %d, shouldProcess %v", unbonding.UnbondingTime, blockTime, shouldProcess))
			if (unbonding.UnbondingTime < blockTime) && shouldProcess {
				// we have a match, so we can process this unbonding
				logMsg := fmt.Sprintf("\t\tAdding %d to amt to batch transfer from delegation acct to rewards acct for host zone %s, epoch %v",
					unbonding.NativeTokenAmount, zone.ChainId, unbondingRecord.EpochNumber)
				k.Logger(ctx).Info(logMsg)
				totalAmtTransferToRedemptionAcct += unbonding.NativeTokenAmount
				unbonding.Status = recordstypes.HostZoneUnbonding_PENDING_TRANSFER
				k.RecordsKeeper.SetEpochUnbondingRecord(ctx, unbondingRecord)
			}

		}
		// if we have any amount to sweep, then we can send the ICA call to sweep them
		if totalAmtTransferToRedemptionAcct > 0 {
			k.Logger(ctx).Info(fmt.Sprintf("\tSending batch SweepAllUnbondedTokens for %d amt to host zone %s", totalAmtTransferToRedemptionAcct, zoneInfo.ChainId))
			// Issue ICA bank send from delegation account to rewards account
			if (&zoneInfo).DelegationAccount != nil && (&zoneInfo).RedemptionAccount != nil { // only process host zones once withdrawal accounts are registered

				// get the delegation account and rewards account
				delegationAccount := zoneInfo.GetDelegationAccount()
				redemptionAccount := zoneInfo.GetRedemptionAccount()

				sweepCoin := sdk.NewCoin(zoneInfo.HostDenom, sdk.NewInt(cast.ToInt64(totalAmtTransferToRedemptionAcct)))
				var msgs []sdk.Msg
				// construct the msg
				msgs = append(msgs, &banktypes.MsgSend{FromAddress: delegationAccount.GetAddress(),
					ToAddress: redemptionAccount.GetAddress(), Amount: sdk.NewCoins(sweepCoin)})

				ctx.Logger().Info(fmt.Sprintf("Bank sending unbonded tokens batch, from delegation to redemption account. Msg: %v", msgs))

				// Send the transaction through SubmitTx
				_, err := k.SubmitTxsDayEpoch(ctx, zoneInfo.ConnectionId, msgs, *delegationAccount, "", nil)
				if err != nil {
					ctx.Logger().Info(fmt.Sprintf("Failed to SubmitTxs for %s", zoneInfo.ChainId))
				}
				ctx.Logger().Info(fmt.Sprintf("Successfully completed unbonded token sweep ICA call for %s, %s, %v", zoneInfo.ConnectionId, zoneInfo.ChainId, msgs))
			}
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("\tNo unbonded tokens this day to sweep for host zone %s", zoneInfo.ChainId))
		}

		return nil
	}
	// Iterate the zones and sweep their unbonded tokens
	k.IterateHostZones(ctx, sweepUnbondedTokens)
}
