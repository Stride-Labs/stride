package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/Stride-Labs/stride/v22/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v22/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v22/x/records/types"
	"github.com/Stride-Labs/stride/v22/x/stakeibc/types"

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
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal undelegate callback args: %s", err.Error()))
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
	// Reset the unbonding record status upon failure
	if ackResponse.Status == icacallbackstypes.AckResponseStatus_FAILURE {
		k.Logger(ctx).Error(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Undelegate,
			icacallbackstypes.AckResponseStatus_FAILURE, packet))

		// Set the unbonding status to RETRY
		if err := k.RecordsKeeper.SetHostZoneUnbondingStatus(
			ctx,
			chainId,
			undelegateCallback.EpochUnbondingRecordIds,
			recordstypes.HostZoneUnbonding_UNBONDING_RETRY_QUEUE,
		); err != nil {
			return err
		}
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Undelegate,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Calculate the native tokens that were unbonded from the batch and get the latest
	// completion time from the ack response
	nativeTokensUnbonded := k.CalculateTotalUnbondedInBatch(undelegateCallback.SplitUndelegations)
	unbondingTime, err := k.GetLatestUnbondingCompletionTime(ctx, ackResponse.MsgResponses)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Update delegation balances on the validators and host zone
	err = k.UpdateDelegationBalances(ctx, hostZone, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
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
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Burn the stTokens from the batch
	if err := k.BurnStTokensAfterUndelegation(ctx, hostZone, stTokensToBurn); err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
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
	// Update the completion time using the latest completion time across each message within the transaction
	for _, msgResponse := range msgResponses {
		// unmarshall the ack response into a MsgUndelegateResponse and grab the completion time
		var undelegateResponse stakingtypes.MsgUndelegateResponse
		err := proto.Unmarshal(msgResponse, &undelegateResponse)
		if err != nil {
			return 0, errorsmod.Wrapf(types.ErrUnmarshalFailure, "Unable to unmarshal undelegation tx response: %s", err.Error())
		}
		responseCompletionTime := uint64(undelegateResponse.CompletionTime.UnixNano())
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

		// Decrement st amount on the record and increment the tota
		hostZoneUnbonding.StTokensToBurn = hostZoneUnbonding.StTokensToBurn.Sub(stTokensToBurn)
		totalStTokensToBurn = totalStTokensToBurn.Add(stTokensToBurn)

		// If there are no more tokens to unbond or burn after this batch, iterate the record to the next status
		if hostZoneUnbonding.StTokensToBurn.IsZero() && hostZoneUnbonding.NativeTokensToUnbond.IsZero() {
			hostZoneUnbonding.Status = recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE
		}

		// Update the unbonding time if the time from this batch is later than what's on the record
		if unbondingTime > hostZoneUnbonding.UnbondingTime {
			hostZoneUnbonding.UnbondingTime = unbondingTime
		}

		// Persist the record changes
		if err := k.RecordsKeeper.SetHostZoneUnbondingRecord(ctx, epochNumber, chainId, *hostZoneUnbonding); err != nil {
			return totalStTokensToBurn, err
		}

		k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
			"Epoch Unbonding Record: %d - Seting unbonding time to %d", epochNumber, unbondingTime))
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
		return fmt.Errorf("could not bech32 decode address %s of zone with id: %s", hostZone.DepositAddress, hostZone.ChainId)
	}
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, depositAddress, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		return fmt.Errorf("could not send coins from account %s to module %s. err: %s", hostZone.DepositAddress, types.ModuleName, err.Error())
	}

	// Finally burn the stTokens
	err = k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(stCoin))
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Failed to burn stAssets upon successful unbonding %s", err.Error()))
		return errorsmod.Wrapf(types.ErrInsufficientFunds, "couldn't burn %v%s tokens in module account. err: %s", stTokenBurnAmount, stCoinDenom, err.Error())
	}
	k.Logger(ctx).Info(fmt.Sprintf("Total supply %s", k.bankKeeper.GetSupply(ctx, stCoinDenom)))
	return nil
}
