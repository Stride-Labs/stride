package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/utils"
	recordstypes "github.com/Stride-Labs/stride/v14/x/records/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

// Submits undelegation ICA messages for a given host zone
//
// First, the total unbond amount is determined from the epoch unbonding records
// Then that unbond amount is allowed to cascade across the validators in order of how proportionally
// different their current delegations are from the weight implied target delegation,
// until their capacities have consumed the full amount
// As a result, unbondings lead to a more balanced distribution of stake across validators
//
// Context: Over time, as LSM Liquid stakes are accepted, the total stake managed by the protocol becomes unbalanced
// as liquid stakes are not aligned with the validator weights. This is only rebalanced once per unbonding period
func (k Keeper) UnbondFromHostZone(ctx sdk.Context, hostZone types.HostZone) error {
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Preparing MsgUndelegates from the delegation account to each validator"))

	// Confirm the delegation account was registered
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", hostZone.ChainId)
	}

	// Iterate through every unbonding record and sum the total amount to unbond for the given host zone
	totalUnbondAmount, epochUnbondingRecordIds := k.GetTotalUnbondAmountAndRecordsIds(ctx, hostZone.ChainId)
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Total unbonded amount: %v%s", totalUnbondAmount, hostZone.HostDenom))

	// If there's nothing to unbond, return and move on to the next host zone
	if totalUnbondAmount.IsZero() {
		return nil
	}

	// Determine the ideal balanced delegation for each validator after the unbonding
	//   (as if we were to unbond and then rebalance)
	// This will serve as the starting point for determining how much to unbond each validator
	delegationAfterUnbonding := hostZone.TotalDelegations.Sub(totalUnbondAmount)
	balancedDelegationsAfterUnbonding, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, delegationAfterUnbonding)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId)
	}

	// Determine the unbond capacity for each validator
	// Each validator can only unbond up to the difference between their current delegation and their balanced delegation
	// The validator's current delegation will be above their balanced delegation if they've received LSM Liquid Stakes
	//   (which is only rebalanced once per unbonding period)
	validatorUnbondCapacity := k.GetValidatorUnbondCapacity(ctx, hostZone.Validators, balancedDelegationsAfterUnbonding)
	if len(validatorUnbondCapacity) == 0 {
		return fmt.Errorf("there are no validators on %s with sufficient unbond capacity", hostZone.ChainId)
	}

	// Sort the unbonding capacity by priority
	// Priority is determined by checking the how proportionally unbalanced each validator is
	// Zero weight validators will come first in the list
	prioritizedUnbondCapacity, err := SortUnbondingCapacityByPriority(validatorUnbondCapacity)
	if err != nil {
		return err
	}

	// Get the undelegation ICA messages and split delegations for the callback
	msgs, unbondings, err := k.GetUnbondingICAMessages(hostZone, totalUnbondAmount, prioritizedUnbondCapacity)
	if err != nil {
		return err
	}

	// Shouldn't be possible, but if all the validator's had a target unbonding of zero, do not send an ICA
	if len(msgs) == 0 {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "Target unbonded amount was 0 for each validator")
	}

	// Send the messages in batches so the gas limit isn't exceedeed
	for start := 0; start < len(msgs); start += UndelegateICABatchSize {
		end := start + UndelegateICABatchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		msgsBatch := msgs[start:end]
		unbondingsBatch := unbondings[start:end]

		// Store the callback data
		undelegateCallback := types.UndelegateCallback{
			HostZoneId:              hostZone.ChainId,
			SplitDelegations:        unbondingsBatch,
			EpochUnbondingRecordIds: epochUnbondingRecordIds,
		}
		callbackArgsBz, err := proto.Marshal(&undelegateCallback)
		if err != nil {
			return errorsmod.Wrap(err, "unable to marshal undelegate callback args")
		}

		// Submit the undelegation ICA
		if _, err := k.SubmitTxsDayEpoch(
			ctx,
			hostZone.ConnectionId,
			msgsBatch,
			types.ICAAccountType_DELEGATION,
			ICACallbackID_Undelegate,
			callbackArgsBz,
		); err != nil {
			return errorsmod.Wrapf(err, "unable to submit unbonding ICA for %s", hostZone.ChainId)
		}

		// flag the delegation change in progress on each validator
		for _, unbonding := range unbondingsBatch {
			if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, unbonding.Validator); err != nil {
				return err
			}
		}
		k.SetHostZone(ctx, hostZone)
	}

	// Update the epoch unbonding record status
	if err := k.RecordsKeeper.SetHostZoneUnbondings(
		ctx,
		hostZone.ChainId,
		epochUnbondingRecordIds,
		recordstypes.HostZoneUnbonding_UNBONDING_IN_PROGRESS,
	); err != nil {
		return err
	}

	EmitUndelegationEvent(ctx, hostZone, totalUnbondAmount)

	return nil
}
