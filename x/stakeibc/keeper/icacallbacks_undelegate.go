package keeper

import (
	"errors"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v14/utils"
	icacallbackstypes "github.com/Stride-Labs/stride/v14/x/icacallbacks/types"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"

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
//	  * Reverts epoch unbonding record status
func (k Keeper) UndelegateCallback(ctx sdk.Context, packet channeltypes.Packet, ackResponse *icacallbackstypes.AcknowledgementResponse, args []byte) error {
	// Fetch callback args
	var undelegateCallback types.UndelegateCallback
	if err := proto.Unmarshal(args, &undelegateCallback); err != nil {
		return errorsmod.Wrapf(types.ErrUnmarshalFailure, fmt.Sprintf("Unable to unmarshal undelegate callback args: %s", err.Error()))
	}
	chainId := undelegateCallback.HostZoneId
	k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
		"Starting undelegate callback for Epoch Unbonding Records: %+v", undelegateCallback.EpochUnbondingRecordIds))

	// Regardless of failure/success/timeout, indicate that this ICA has completed
	hostZone, found := k.GetHostZone(ctx, undelegateCallback.HostZoneId)
	if !found {
		return errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, "Host zone not found: %s", undelegateCallback.HostZoneId)
	}
	for _, splitDelegation := range undelegateCallback.SplitDelegations {
		if err := k.DecrementValidatorDelegationChangesInProgress(&hostZone, splitDelegation.Validator); err != nil {
			// TODO: Revert after v14 upgrade
			if errors.Is(err, types.ErrInvalidValidatorDelegationUpdates) {
				k.Logger(ctx).Error(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
					"Invariant failed - delegation changes in progress fell below 0 for %s", splitDelegation.Validator))
				continue
			}
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)

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

		// Reset unbondings record status
		if err := k.RecordsKeeper.SetHostZoneUnbondings(
			ctx,
			chainId,
			undelegateCallback.EpochUnbondingRecordIds,
			recordstypes.HostZoneUnbonding_UNBONDING_QUEUE,
		); err != nil {
			return err
		}
		return nil
	}

	k.Logger(ctx).Info(utils.LogICACallbackStatusWithHostZone(chainId, ICACallbackID_Undelegate,
		icacallbackstypes.AckResponseStatus_SUCCESS, packet))

	// Update delegation balances
	err := k.UpdateDelegationBalances(ctx, hostZone, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Get the latest transaction completion time (to determine the unbonding time)
	latestCompletionTime, err := k.GetLatestCompletionTime(ctx, ackResponse.MsgResponses)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Burn the stTokens
	stTokenBurnAmount, err := k.UpdateHostZoneUnbondings(ctx, *latestCompletionTime, chainId, undelegateCallback)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}
	err = k.BurnTokens(ctx, hostZone, stTokenBurnAmount)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("UndelegateCallback | %s", err.Error()))
		return err
	}

	// Upon success, add host zone unbondings to the exit transfer queue
	err = k.RecordsKeeper.SetHostZoneUnbondings(ctx, chainId, undelegateCallback.EpochUnbondingRecordIds, recordstypes.HostZoneUnbonding_EXIT_TRANSFER_QUEUE)
	if err != nil {
		return err
	}

	return nil
}

// Decrement the delegation field on the host zone and each validator's delegations after a successful unbonding ICA
func (k Keeper) UpdateDelegationBalances(ctx sdk.Context, hostZone types.HostZone, undelegateCallback types.UndelegateCallback) error {
	// Undelegate from each validator and update host zone staked balance, if successful
	for _, undelegation := range undelegateCallback.SplitDelegations {
		err := k.AddDelegationToValidator(ctx, &hostZone, undelegation.Validator, undelegation.Amount.Neg(), ICACallbackID_Undelegate)
		if err != nil {
			return err
		}
	}
	k.SetHostZone(ctx, hostZone)
	return nil
}

// Get the latest completion time across each MsgUndelegate in the ICA transaction
// The time is used to set the
func (k Keeper) GetLatestCompletionTime(ctx sdk.Context, msgResponses [][]byte) (*time.Time, error) {
	// Update the completion time using the latest completion time across each message within the transaction
	latestCompletionTime := time.Time{}

	for _, msgResponse := range msgResponses {
		// unmarshall the ack response into a MsgUndelegateResponse and grab the completion time
		var undelegateResponse stakingtypes.MsgUndelegateResponse
		err := proto.Unmarshal(msgResponse, &undelegateResponse)
		if err != nil {
			return nil, errorsmod.Wrapf(types.ErrUnmarshalFailure, "Unable to unmarshal undelegation tx response: %s", err.Error())
		}
		if undelegateResponse.CompletionTime.After(latestCompletionTime) {
			latestCompletionTime = undelegateResponse.CompletionTime
		}
	}

	if latestCompletionTime.IsZero() {
		return nil, errorsmod.Wrapf(types.ErrInvalidPacketCompletionTime, "Invalid completion time (%s) from txMsg", latestCompletionTime.String())
	}
	return &latestCompletionTime, nil
}

// UpdateHostZoneUnbondings does two things:
//  1. Update the time of each hostZoneUnbonding on each epochUnbondingRecord
//  2. Return the number of stTokens that need to be burned
func (k Keeper) UpdateHostZoneUnbondings(
	ctx sdk.Context,
	latestCompletionTime time.Time,
	chainId string,
	undelegateCallback types.UndelegateCallback,
) (stTokenBurnAmount sdkmath.Int, err error) {
	stTokenBurnAmount = sdkmath.ZeroInt()
	for _, epochNumber := range undelegateCallback.EpochUnbondingRecordIds {
		epochUnbondingRecord, found := k.RecordsKeeper.GetEpochUnbondingRecord(ctx, epochNumber)
		if !found {
			errMsg := fmt.Sprintf("Unable to find epoch unbonding record for epoch: %d", epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdkmath.ZeroInt(), errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}
		hostZoneUnbonding, found := k.RecordsKeeper.GetHostZoneUnbondingByChainId(ctx, epochUnbondingRecord.EpochNumber, chainId)
		if !found {
			errMsg := fmt.Sprintf("Host zone unbonding not found (%s) in epoch unbonding record: %d", chainId, epochNumber)
			k.Logger(ctx).Error(errMsg)
			return sdkmath.ZeroInt(), errorsmod.Wrapf(sdkerrors.ErrKeyNotFound, errMsg)
		}

		// Keep track of the stTokens that need to be burned
		stTokenAmount := hostZoneUnbonding.StTokenAmount
		stTokenBurnAmount = stTokenBurnAmount.Add(stTokenAmount)

		// Update the bonded time
		hostZoneUnbonding.UnbondingTime = cast.ToUint64(latestCompletionTime.UnixNano())
		updatedEpochUnbondingRecord, success := k.RecordsKeeper.AddHostZoneToEpochUnbondingRecord(ctx, epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding)
		if !success {
			k.Logger(ctx).Error(fmt.Sprintf("Failed to set host zone epoch unbonding record: epochNumber %d, chainId %s, hostZoneUnbonding %+v",
				epochUnbondingRecord.EpochNumber, chainId, hostZoneUnbonding))
			return sdkmath.ZeroInt(), errorsmod.Wrapf(types.ErrEpochNotFound, "couldn't set host zone epoch unbonding record. err: %s", err.Error())
		}
		k.RecordsKeeper.SetEpochUnbondingRecord(ctx, *updatedEpochUnbondingRecord)

		k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(chainId, ICACallbackID_Undelegate,
			"Epoch Unbonding Record: %d - Seting unbonding time to %s", epochNumber, latestCompletionTime.String()))
	}
	return stTokenBurnAmount, nil
}

// Burn stTokens after they've been unbonded
func (k Keeper) BurnTokens(ctx sdk.Context, hostZone types.HostZone, stTokenBurnAmount sdkmath.Int) error {
	// Build the coin from the stDenom on the host zone
	stCoinDenom := types.StAssetDenomFromHostZoneDenom(hostZone.HostDenom)
	stCoinString := stTokenBurnAmount.String() + stCoinDenom
	stCoin, err := sdk.ParseCoinNormalized(stCoinString)
	if err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidCoins, "could not parse burnCoin: %s. err: %s", stCoinString, err.Error())
	}

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
