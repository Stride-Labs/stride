package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v16/utils"
	"github.com/Stride-Labs/stride/v16/x/stakeibc/types"
)

const (
	MaxNumTokensUnbondableStr = "2500000000000000000000000" // 2,500,000e18
	EvmosHostZoneChainId      = "evmos_9001-2"
)

// Submits undelegation ICA message for Evmos
// The total unbond amount is input, capped at MaxNumTokensUnbondable.
func (k Keeper) UndelegateHostEvmos(ctx sdk.Context, totalUnbondAmount math.Int) error {

	// if the total unbond amount is greater than the max, exit
	maxNumTokensUnbondable, ok := math.NewIntFromString(MaxNumTokensUnbondableStr)
	if !ok {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "unable to parse maxNumTokensUnbondable %s", maxNumTokensUnbondable)
	}
	if totalUnbondAmount.GT(maxNumTokensUnbondable) {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "total unbond amount %v is greater than maxNumTokensUnbondable %v",
			totalUnbondAmount, maxNumTokensUnbondable)
	}

	// Get the host zone
	evmosHost, found := k.GetHostZone(ctx, EvmosHostZoneChainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "host zone %s not found", EvmosHostZoneChainId)
	}

	k.Logger(ctx).Info(utils.LogWithHostZone(evmosHost.ChainId,
		"Total unbonded amount: %v%s", totalUnbondAmount, evmosHost.HostDenom))

	// If there's nothing to unbond, return and move on to the next host zone
	if totalUnbondAmount.IsZero() {
		return nil
	}

	k.Logger(ctx).Info("Preparing MsgUndelegates from the delegation account to each validator on Evmos")

	// Confirm the delegation account was registered
	if evmosHost.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", evmosHost.ChainId)
	}

	// Determine the ideal balanced delegation for each validator after the unbonding
	//   (as if we were to unbond and then rebalance)
	// This will serve as the starting point for determining how much to unbond each validator
	delegationAfterUnbonding := evmosHost.TotalDelegations.Sub(totalUnbondAmount)
	balancedDelegationsAfterUnbonding, err := k.GetTargetValAmtsForHostZone(ctx, evmosHost, delegationAfterUnbonding)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", evmosHost.ChainId)
	}

	// Determine the unbond capacity for each validator
	// Each validator can only unbond up to the difference between their current delegation and their balanced delegation
	// The validator's current delegation will be above their balanced delegation if they've received LSM Liquid Stakes
	//   (which is only rebalanced once per unbonding period)
	validatorUnbondCapacity := k.GetValidatorUnbondCapacity(ctx, evmosHost.Validators, balancedDelegationsAfterUnbonding)
	if len(validatorUnbondCapacity) == 0 {
		return fmt.Errorf("there are no validators on %s with sufficient unbond capacity", evmosHost.ChainId)
	}

	// Sort the unbonding capacity by priority
	// Priority is determined by checking the how proportionally unbalanced each validator is
	// Zero weight validators will come first in the list
	prioritizedUnbondCapacity, err := SortUnbondingCapacityByPriority(validatorUnbondCapacity)
	if err != nil {
		return err
	}

	// Get the undelegation ICA messages and split delegations for the callback
	msgs, unbondings, err := k.GetUnbondingICAMessages(evmosHost, totalUnbondAmount, prioritizedUnbondCapacity, UndelegateICABatchSize)
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
		undelegateHostCallback := types.UndelegateHostCallback{
			Amt:              totalUnbondAmount,
			SplitDelegations: unbondingsBatch,
		}
		callbackArgsBz, err := proto.Marshal(&undelegateHostCallback)
		if err != nil {
			return errorsmod.Wrap(err, "unable to marshal undelegate callback args")
		}

		// Submit the undelegation ICA
		if _, err := k.SubmitTxsDayEpoch(
			ctx,
			evmosHost.ConnectionId,
			msgsBatch,
			types.ICAAccountType_DELEGATION,
			ICACallbackID_UndelegateHost,
			callbackArgsBz,
		); err != nil {
			return errorsmod.Wrapf(err, "unable to submit unbonding ICA for %s", evmosHost.ChainId)
		}

		// flag the delegation change in progress on each validator
		for _, unbonding := range unbondingsBatch {
			if err := k.IncrementValidatorDelegationChangesInProgress(&evmosHost, unbonding.Validator); err != nil {
				return err
			}
		}
		k.SetHostZone(ctx, evmosHost)
	}

	EmitUndelegationEvent(ctx, evmosHost, totalUnbondAmount)

	return nil
}
