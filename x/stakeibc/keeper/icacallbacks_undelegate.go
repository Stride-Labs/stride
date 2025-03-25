package keeper

import (
	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v26/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v26/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v26/x/records/types"
	"github.com/Stride-Labs/stride/v26/x/stakeibc/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

// ICA Callback after undelegating
//
//	If successful:
//	  * Updates epoch unbonding record status
//	  * Records delegation changes on the host zone and validators,
//	  * Burns stTokens
//	If timeout:
//	  * Does nothing
//	If failure:
//	  * Sets epoch unbonding record status to RETRY
func (k Keeper) UndelegateCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	var undelegateCallback types.UndelegateCallback
	if err := proto.Unmarshal(args, &undelegateCallback); err != nil {
		return errorsmod.Wrap(err, "unable to unmarshal undelegate callback args")
	}

	// Fetch the relevant host zone
	chainId := undelegateCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
		"Starting undelegate callback for Epoch Unbonding Records: %+v", undelegateCallback.EpochUnbondingRecordIds))

	hostZone, found := k.GetHostZone(ctx, undelegateCallback.HostZoneId)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", undelegateCallback.HostZoneId)
	}

	// Mark that the ICA completed on the validators and host zone unbonding records
	if err := k.MarkUndelegationAckReceived(ctx, hostZone, undelegateCallback); err != nil {
		return err
	}

	// Check for timeout (ack nil)
	// No need to reset the unbonding record status since it will get reverted when the channel is restored
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_TIMEOUT {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Undelegate,
			icacallbackstypes.AckResponseStatus_TIMEOUT, packet))
		return nil
	}

	// Check for a failed transaction (ack error)
	// Set the status to RETRY_QUEUE if it fails
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Undelegate,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		// Set any IN_PROGRESS records to RETRY_QUEUE
		return k.HandleFailedUndelegation(ctx, chainId, undelegateCallback.EpochUnbondingRecordIds)
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Undelegate,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Calculate the native tokens that were unbonded from the batch and get the latest
	// completion time from the ack response
	nativeTokensUnbonded := k.CalculateTotalUnbondedInBatch(undelegateCallback.SplitUndelegations)
	unbondingTime, err := k.GetLatestUnbondingCompletionTime(ctx, ackResponse.MsgResponses)
	if err != nil {
		return err
	}

	// Update delegation balances on the validators and host zone
	err = k.UpdateDelegationBalances(ctx, hostZone, undelegateCallback)
	if err != nil {
		return err
	}

	// Update the accounting on the host zone unbondings
	stTokensToBurn, err := k.UpdateHostZoneUnbondingsAfterUndelegation(
		ctx,
		chainId,
		undelegateCallback.EpochUnbondingRecordIds,
		nativeTokensUnbonded,
		unbondingTime,
	)
	if err != nil {
		return err
	}

	// Burn the stTokens from the batch
	if err := k.BurnStTokensAfterUndelegation(ctx, hostZone, stTokensToBurn); err != nil {
		return err
	}

	return nil
}

// Regardless of failure/success/timeout, indicate that this ICA has completed on each validator
// on the host zone, and on the epoch unbonding record
func (k Keeper) MarkUndelegationAckReceived(ctx sdk.Context, hostZone types.HostZone, undelegateCallback types.UndelegateCallback) error {
	// Indicate that this ICA has completed on each validator
	for _, splitDelegation := range undelegateCallback.SplitUndelegations {
		if err := k.DecrementValidatorDelegationChangesInProgress(&hostZone, splitDelegation.Validator); err != nil {
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)

	// Indicate that the ICA has completed on the epoch unbonding record
	for _, epochNumber := range undelegateCallback.EpochUnbondingRecordIds {
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochNumber, hostZone.ChainId)
		if !found {
			return recordstypes.ErrHostUnbondingRecordNotFound.Wrapf("epoch number %d, chain %s", epochNumber, hostZone.ChainId)
		}

		if hostZoneUnbonding.UndelegationTxsInProgress == 0 {
			return types.ErrInvalidUndelegationsInProgress.Wrapf("undelegation changes in progress is already 0 and can't be decremented")
		}
		hostZoneUnbonding.UndelegationTxsInProgress -= 1

		if err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, hostZone.ChainId, *hostZoneUnbonding); err != nil {
			return err
		}
	}

	return nil
}

// If the undelegation failed, set the unbonding status to RETRY_QUEUE, but only
// for records that are currently in status UNBONDING_IN_PROGRESS
// There may be some epoch numbers in this batch from records that have already had a full unbonding
// and have moved onto status EXIT_TRANSFER_QUEUE
func (k Keeper) HandleFailedUndelegation(ctx sdk.Context, chainId string, epochNumbers []uint64) error {
	for _, epochNumber := range epochNumbers {
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochNumber, chainId)
		if !found {
			return errorsmod.Wrapf(recordstypes.ErrHostUnbondingRecordNotFound, "epoch number %d, chain %s",
				epochNumber, chainId)
		}

		if hostZoneUnbonding.Status != recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS {
			continue
		}
		hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE

		err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, chainId, *hostZoneUnbonding)
		if err != nil {
			return err
		}
	}
	return nil
}

// Decrement the delegation field on the host zone and each validator's delegations after a successful unbonding ICA
func (k Keeper) UpdateDelegationBalances(ctx sdk.Context, hostZone types.HostZone, undelegateCallback types.UndelegateCallback) error {
	// Undelegate from each validator and update host zone staked balance, if successful
	for _, undelegation := range undelegateCallback.SplitUndelegations {
		err := k.AddDelegationToValidator(ctx, &hostZone, undelegation.Validator, undelegation.NativeTokenAmount.Neg(), ICACallbackID_Undelegate)
		if err != nil {
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)
	return nil
}

// Calculates the tokens unbonded for this batch by summing from each validator
func (k Keeper) CalculateTotalUnbondedInBatch(undelegations []*types.SplitUndelegation) (nativeTokens sdkmath.Int) {
	nativeTokens = sdkmath.ZeroInt()
	for _, undelegation := range undelegations {
		nativeTokens = nativeTokens.Add(undelegation.NativeTokenAmount)
	}
	return nativeTokens
}

// Get the latest completion time across each MsgUndelegate in the ICA transaction
// The time is later stored on the unbonding record
func (k Keeper) GetLatestUnbondingCompletionTime(ctx sdk.Context, msgResponses [][]byte) (latestCompletionTime uint64, err error) {
	for _, msgResponse := range msgResponses {
		var undelegateResponse stakingtypes.MsgUndelegateResponse
		if err := proto.Unmarshal(msgResponse, &undelegateResponse); err != nil {
			return 0, errorsmod.Wrapf(types.ErrUnmarshalFailure, "Unable to unmarshal undelegation tx response: %s", err.Error())
		}

		responseCompletionTime := utils.IntToUint(undelegateResponse.CompletionTime.UnixNano())
		if responseCompletionTime > latestCompletionTime {
			latestCompletionTime = responseCompletionTime
		}
	}

	if latestCompletionTime == 0 {
		return 0, errorsmod.Wrapf(types.ErrInvalidPacketCompletionTime, "Invalid completion time 0 from txMsg")
	}
	return latestCompletionTime, nil
}

// Updates the host zone unbonding records after a successful undelegation batch
// The StTokensToBurn and the NativeTokensToUnbond amounts on the records are
// decremented in a cascading fashion starting from the earliest record
// The latest completion times is also set on each record if the time from the
// batch is later than what's currently on the record
func (k Keeper) UpdateHostZoneUnbondingsAfterUndelegation(
	ctx sdk.Context,
	chainId string,
	epochUnbondingRecordIds []uint64,
	totalNativeTokensUnbonded sdkmath.Int,
	unbondingTime uint64,
) (totalStTokensToBurn sdkmath.Int, err error) {
	// As we process the accounting changes, keep track of the stTokens that should be burned later
	totalStTokensToBurn = sdkmath.ZeroInt()

	// Loop each epoch unbonding record starting from the earliest
	for _, epochNumber := range epochUnbondingRecordIds {
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochNumber, chainId)
		if !found {
			return totalStTokensToBurn, errorsmod.Wrapf(recordstypes.ErrHostUnbondingRecordNotFound,
				"host zone unbonding not found for epoch %d and %s", epochNumber, chainId)
		}

		// If the record was already completed by a previous callback, continue to the next record
		if hostZoneUnbonding.Status == recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE {
			continue
		}

		// Determine the native amount to decrement from the record, capping at the amount in the record
		// Also decrement the total for the next loop
		nativeTokensUnbonded := sdkmath.MinInt(hostZoneUnbonding.NativeTokensToUnbond, totalNativeTokensUnbonded)
		hostZoneUnbonding.NativeTokensToUnbond = hostZoneUnbonding.NativeTokensToUnbond.Sub(nativeTokensUnbonded)
		totalNativeTokensUnbonded = totalNativeTokensUnbonded.Sub(nativeTokensUnbonded)

		// Calculate the relative stToken portion using the implied RR from the record
		// If the native amount has already been decremented to 0, just use the full stToken remainder
		// from the record to prevent any precision error
		var stTokensToBurn sdkmath.Int
		if hostZoneUnbonding.NativeTokensToUnbond.IsZero() {
			stTokensToBurn = hostZoneUnbonding.StTokensToBurn
		} else {
			impliedRedemptionRate := sdk.NewDecFromInt(hostZoneUnbonding.NativeTokenAmount).Quo(sdk.NewDecFromInt(hostZoneUnbonding.StTokenAmount))
			stTokensToBurn = sdk.NewDecFromInt(nativeTokensUnbonded).Quo(impliedRedemptionRate).TruncateInt()
		}

		k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
			"Epoch Unbonding Record: %d - Native Unbonded: %v, StTokens Burned: %v",
			epochNumber, nativeTokensUnbonded, stTokensToBurn))

		// Decrement st amount on the record and increment the total
		hostZoneUnbonding.StTokensToBurn = hostZoneUnbonding.StTokensToBurn.Sub(stTokensToBurn)
		totalStTokensToBurn = totalStTokensToBurn.Add(stTokensToBurn)

		// If there are no more tokens to unbond or burn after this batch, iterate the record to the next status
		if hostZoneUnbonding.StTokensToBurn.IsZero() && hostZoneUnbonding.NativeTokensToUnbond.IsZero() {
			hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
		}

		// Update the unbonding time if the time from this batch is later than what's on the record
		if unbondingTime > hostZoneUnbonding.UnbondingTime {
			hostZoneUnbonding.UnbondingTime = unbondingTime

			k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
				"Epoch Unbonding Record: %d - Setting unbonding time to %d", epochNumber, unbondingTime))
		}

		// Persist the record changes
		if err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, chainId, *hostZoneUnbonding); err != nil {
			return totalStTokensToBurn, err
		}
	}
	return totalStTokensToBurn, nil
}

// Burn stTokens after they've been unbonded
func (k Keeper) BurnStTokensAfterUndelegation(ctx sdk.Context, hostZone types.HostZone, stTokenBurnAmount sdkmath.Int) error {
	// Build the coin from the stDenom on the host zone
	stCoinDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stCoin := sdk.NewCoin(stCoinDenom, stTokenBurnAmount)

	// Send the stTokens from the host zone module account to the stakeibc module account
	depositAddress, err := sdk.AccAddressFromBech32(hostZone.DepositAddress)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to convert deposit address")
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositAddress, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		return errorsmod.Wrapf(err, "unable to send sttokens from deposit account for burning")
	}

	// Finally burn the stTokens
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		return errorsmod.Wrapf(err, "unable to burn %v%s tokens", stTokenBurnAmount, stCoinDenom)
	}
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(hostZone.ChainId, ICACallbackID_Undelegate,
		"Total Burned from Batch %v", stCoin))
	return nil
}
