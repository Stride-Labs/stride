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

func (k Keeper) CreateEpochUnbondingRecord(ctx sdk.Context, epochNumber uint64) bool {
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
		hostZoneUnbondings = append(hostZoneUnbondings, &hostZoneUnbonding)
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

// return:
// - msgs to send to the host zone
// - total amount to unbond
// - marshalled callback args
// - error
func (k Keeper) GetHostZoneUnbondingMsgs(ctx sdk.Context, hostZone types.HostZone) ([]sdk.Msg, uint64, []byte, error) {
	// this function goes and processes all unbonded records for this hostZone
	// regardless of what epoch they belong to
	totalAmtToUnbond := uint64(0)
	epochUnbondingRecordIds := []uint64{}
	var msgs []sdk.Msg
	for _, epochUnbonding := range k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx) {
		hostZoneRecord, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbonding.EpochNumber, hostZone.ChainId)
		if !found {
			errMsg := fmt.Sprintf("Host zone unbonding record not found for hostZoneId %s in epoch %d",
				hostZone.ChainId, epochUnbonding.GetEpochNumber())
			k.Logger(ctx).Error(errMsg)
			continue
		}
		// mark the epoch unbonding record for processing if it's bonded and the host zone unbonding has an amount g.t. zero
		if hostZoneRecord.Status == recordstypes.HostZoneUnbonding_BONDED && hostZoneRecord.NativeTokenAmount > 0 {
			totalAmtToUnbond += hostZoneRecord.NativeTokenAmount
			epochUnbondingRecordIds = append(epochUnbondingRecordIds, epochUnbonding.EpochNumber)
			k.Logger(ctx).Info(fmt.Sprintf("[SendHostZoneUnbondings] Sending unbondings, host zone: %s, epochUnbonding: %v", hostZone.ChainId, epochUnbonding))

		}
	}
	delegationAccount := hostZone.GetDelegationAccount()
	if delegationAccount == nil || delegationAccount.GetAddress() == "" {
		errMsg := fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
	}
	validators := hostZone.GetValidators()
	if totalAmtToUnbond == 0 {
		return nil, 0, nil, nil
	}
	// we distribute the unbonding based on our target weights
	newUnbondingToValidator, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalAmtToUnbond)
	if err != nil {
		errMsg := fmt.Sprintf("Error getting target val amts for host zone %s %d: %s", hostZone.ChainId, totalAmtToUnbond, err)
		k.Logger(ctx).Error(errMsg)
		return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
	}
	valAddrToUnbondAmt := make(map[string]int64)
	overflowAmt := uint64(0)
	for _, validator := range validators {
		valAddr := validator.GetAddress()
		valUnbondAmt := newUnbondingToValidator[valAddr]
		currentAmtStaked := validator.GetDelegationAmt()
		if err != nil {
			errMsg := fmt.Sprintf("Error casting validator staked amount %d: %s", validator.GetDelegationAmt(), err.Error())
			k.Logger(ctx).Error(errMsg)
			return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
		}
		if valUnbondAmt > currentAmtStaked { // if we don't have enough assets to unbond
			overflowAmt += valUnbondAmt - currentAmtStaked
			valUnbondAmt = currentAmtStaked
		}
		valUnbondAmtInt64, err := cast.ToInt64E(valUnbondAmt)
		if err != nil {
			errMsg := fmt.Sprintf("Error casting validator staked amount %d: %s", validator.GetDelegationAmt(), err.Error())
			k.Logger(ctx).Error(errMsg)
			return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
		}
		valAddrToUnbondAmt[valAddr] = valUnbondAmtInt64
	}
	if overflowAmt > 0 { // if we need to reallocate any weights
		for _, validator := range validators {
			valAddr := validator.GetAddress()
			valUnbondAmt, err := cast.ToUint64E(valAddrToUnbondAmt[valAddr])
			if err != nil {
				errMsg := fmt.Sprintf("Error casting validator staked amount %d: %s", validator.GetDelegationAmt(), err.Error())
				k.Logger(ctx).Error(errMsg)
				return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
			}
			currentAmtStaked := validator.GetDelegationAmt()
			// store how many more tokens we could unbond, if needed
			curAmtStaked := currentAmtStaked
			amtToPotentiallyUnbond := curAmtStaked - valUnbondAmt
			if amtToPotentiallyUnbond > 0 { // if we can afford to unbond more
				if amtToPotentiallyUnbond > overflowAmt { // we can fully cover the overflow
					overflowAmtInt64, err := cast.ToInt64E(overflowAmt)
					if err != nil {
						errMsg := fmt.Sprintf("Error casting overflow amount %d: %s", overflowAmt, err.Error())
						k.Logger(ctx).Error(errMsg)
						return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
					}
					valAddrToUnbondAmt[valAddr] += overflowAmtInt64
					overflowAmt = 0
					break
				} else {
					amtToPotentiallyUnbondInt64, err := cast.ToInt64E(amtToPotentiallyUnbond)
					if err != nil {
						errMsg := fmt.Sprintf("Error casting overflow amount %d: %s", amtToPotentiallyUnbond, err.Error())
						k.Logger(ctx).Error(errMsg)
						return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
					}
					valAddrToUnbondAmt[valAddr] += amtToPotentiallyUnbondInt64
					overflowAmt -= amtToPotentiallyUnbond
				}
			}
		}
	}
	if overflowAmt > 0 { // what?? we still can't cover the overflow? something is very wrong
		errMsg := fmt.Sprintf("Could not unbond %d on Host Zone %s, unable to balance the unbond amount across validators",
			totalAmtToUnbond, hostZone.ChainId)
		k.Logger(ctx).Error(errMsg)
		return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
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
		HostZoneId:              hostZone.ChainId,
		SplitDelegations:        splitDelegations,
		EpochUnbondingRecordIds: epochUnbondingRecordIds,
	}
	k.Logger(ctx).Info(fmt.Sprintf("Marshalling UndelegateCallback args: %v", undelegateCallback))
	marshalledCallbackArgs, err := k.MarshalUndelegateCallbackArgs(ctx, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, 0, nil, sdkerrors.Wrap(sdkerrors.ErrNotFound, err.Error())
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("newAmountUnbonding", strconv.FormatUint(totalAmtToUnbond, 10)),
		),
	)
	return msgs, totalAmtToUnbond, marshalledCallbackArgs, nil
}

func (k Keeper) SubmitHostZoneUnbondingMsg(ctx sdk.Context, msgs []sdk.Msg, totalAmtToUnbond uint64, marshalledCallbackArgs []byte, hostZone types.HostZone) error {
	delegationAccount := hostZone.GetDelegationAccount()

	// safety check: if msgs is nil, error
	if msgs == nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "no msgs to submit for host zone unbondings")
	}

	_, err := k.SubmitTxsDayEpoch(ctx, hostZone.GetConnectionId(), msgs, *delegationAccount, UNDELEGATE, marshalledCallbackArgs)
	if err != nil {
		errMsg := fmt.Sprintf("Error submitting unbonding tx: %s", err)
		k.Logger(ctx).Error(errMsg)
		return sdkerrors.Wrap(sdkerrors.ErrNotFound, errMsg)
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute("hostZone", hostZone.ChainId),
			sdk.NewAttribute("newAmountUnbonding", strconv.FormatUint(totalAmtToUnbond, 10)),
		),
	)

	return nil
}

func (k Keeper) InitiateAllHostZoneUnbondings(ctx sdk.Context, dayNumber uint64) bool {
	// this function goes through each host zone, and if it's the right time to
	// initiate an unbonding, it goes and tries to unbond all outstanding records
	for _, hostZone := range k.GetAllHostZone(ctx) {
		k.Logger(ctx).Info(fmt.Sprintf("Processing epoch unbondings for host zone %s", hostZone.GetChainId()))
		// we only send the ICA call if this hostZone is supposed to be triggered
		if dayNumber%hostZone.UnbondingFrequency == 0 {
			k.Logger(ctx).Info(fmt.Sprintf("Sending unbondings for host zone %s", hostZone.ChainId))
			msgs, totalAmtToUnbond, marshalledCallbackArgs, err := k.GetHostZoneUnbondingMsgs(ctx, hostZone)
			if err != nil {
				errMsg := fmt.Sprintf("Error getting unbonding msgs for host zone %s: %s", hostZone.ChainId, err.Error())
				k.Logger(ctx).Error(errMsg)
				return false
			}
			err = k.SubmitHostZoneUnbondingMsg(ctx, msgs, totalAmtToUnbond, marshalledCallbackArgs, hostZone)
			if err != nil {
				errMsg := fmt.Sprintf("Error submitting unbonding tx for host zone %s: %s", hostZone.ChainId, err.Error())
				k.Logger(ctx).Error(errMsg)
				return false
			}
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
	sweepUnbondedTokens := func(ctx sdk.Context, index int64, hostZone types.HostZone) error {
		k.Logger(ctx).Info(fmt.Sprintf("sweepUnbondedTokens for host zone %s", hostZone.ChainId))

		epochUnbondingRecords := k.RecordsKeeper.GetAllEpochUnbondingRecord(ctx)
		totalAmtTransferToRedemptionAcct := int64(0)
		epochUnbondingRecordIds := []uint64{}
		for _, epochUnbondingRecord := range epochUnbondingRecords {
			k.Logger(ctx).Info(fmt.Sprintf("processing epochUnbondingRecord %v", epochUnbondingRecord.EpochNumber))

			// iterate through all host zone unbondings and process them if they're ready to be swept
			hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, hostZone.ChainId)
			if !found {
				k.Logger(ctx).Error(fmt.Sprintf("Could not find host zone unbonding %d for host zone %s", epochUnbondingRecord.EpochNumber, hostZone.ChainId))
				return sdkerrors.Wrapf(sdkerrors.ErrNotFound, "Could not find host zone unbonding %d for host zone %s", epochUnbondingRecord.EpochNumber, hostZone.ChainId)
			}
			k.Logger(ctx).Info(fmt.Sprintf("\tProcessing batch SweepAllUnbondedTokens for host zone %s", hostZone.ChainId))

			// get latest blockTime from light client
			blockTime, found := k.GetLightClientTimeSafely(ctx, hostZone.ConnectionId)
			if !found {
				errMsg := fmt.Sprintf("\tCould not find blockTime for host zone %s", hostZone.ChainId)
				k.Logger(ctx).Error(errMsg)
				return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, errMsg)
			}

			shouldProcess := hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_UNBONDED
			k.Logger(ctx).Info(fmt.Sprintf("\tUnbonding time:  %d blockTime %d, shouldProcess %v", hostZoneUnbonding.UnbondingTime, blockTime, shouldProcess))

			// if the unbonding period has elapsed, then we can send the ICA call to sweep this hostZone's unbondings to the redemption account (in a batch)
			if (hostZoneUnbonding.UnbondingTime < blockTime) && shouldProcess {
				// we have a match, so we can process this unbonding
				logMsg := fmt.Sprintf("\t\tAdding %d to amt to batch transfer from delegation acct to rewards acct for host zone %s, epoch %v",
					hostZoneUnbonding.NativeTokenAmount, hostZone.ChainId, epochUnbondingRecord.EpochNumber)
				k.Logger(ctx).Info(logMsg)

				nativeTokenAmount, err := cast.ToInt64E(hostZoneUnbonding.NativeTokenAmount)
				if err != nil {
					errMsg := fmt.Sprintf("Could not convert native token amount to int64 | %s", err.Error())
					k.Logger(ctx).Error(errMsg)
					return sdkerrors.Wrapf(types.ErrIntCast, errMsg)
				}
				totalAmtTransferToRedemptionAcct += nativeTokenAmount
				epochUnbondingRecordIds = append(epochUnbondingRecordIds, epochUnbondingRecord.EpochNumber)
			}
		}
		// if we have any amount to sweep, then we can send the ICA call to sweep them
		if totalAmtTransferToRedemptionAcct > 0 {
			k.Logger(ctx).Info(fmt.Sprintf("\tSending batch SweepAllUnbondedTokens for %d amt to host zone %s", totalAmtTransferToRedemptionAcct, hostZone.ChainId))
			// Issue ICA bank send from delegation account to redemption account
			if (&hostZone).DelegationAccount != nil && (&hostZone).RedemptionAccount != nil { // only process host zones once withdrawal accounts are registered

				// get the delegation account and rewards account
				delegationAccount := hostZone.GetDelegationAccount()
				if delegationAccount == nil || delegationAccount.Address == "" {
					k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegation account")
				}
				redemptionAccount := hostZone.GetRedemptionAccount()
				if redemptionAccount == nil || redemptionAccount.Address == "" {
					k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a redemption address!", hostZone.ChainId))
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid redemption account")
				}

				// build transfer message from delegation account to redemption account
				sweepCoin := sdk.NewCoin(hostZone.HostDenom, sdk.NewInt(totalAmtTransferToRedemptionAcct))
				var msgs []sdk.Msg
				msgs = append(msgs, &banktypes.MsgSend{
					FromAddress: delegationAccount.GetAddress(),
					ToAddress:   redemptionAccount.GetAddress(),
					Amount:      sdk.NewCoins(sweepCoin),
				})
				ctx.Logger().Info(fmt.Sprintf("Bank sending unbonded tokens batch, from delegation to redemption account. Msg: %v", msgs))

				// store the epoch numbers in the callback to identify the epoch unbonding records
				redemptionCallback := types.RedemptionCallback{
					HostZoneId:              hostZone.ChainId,
					EpochUnbondingRecordIds: epochUnbondingRecordIds,
				}

				marshalledCallbackArgs, err := k.MarshalRedemptionCallbackArgs(ctx, redemptionCallback)
				if err != nil {
					k.Logger(ctx).Error(err.Error())
					return err
				}

				// Send the transaction through SubmitTx
				_, err = k.SubmitTxsDayEpoch(ctx, hostZone.ConnectionId, msgs, *delegationAccount, REDEMPTION, marshalledCallbackArgs)
				if err != nil {
					ctx.Logger().Info(fmt.Sprintf("Failed to SubmitTxs, transfer to redemption account on %s", hostZone.ChainId))
				}
				ctx.Logger().Info(fmt.Sprintf("Successfully completed unbonded token sweep ICA call for %s, %s, %v", hostZone.ConnectionId, hostZone.ChainId, msgs))
			}
		} else {
			k.Logger(ctx).Info(fmt.Sprintf("\tNo unbonded tokens this day to sweep for host zone %s", hostZone.ChainId))
		}

		return nil
	}
	// Iterate the zones and sweep their unbonded tokens
	k.IterateHostZones(ctx, sweepUnbondedTokens)
}
