package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/Stride-Labs/stride/v14/utils"
)

func (k Keeper) QueryNextHubUnbonding(ctx sdk.Context) error {
	hubChainId := "cosmoshub-4"
	hostZone, found := k.GetHostZone(ctx, hubChainId)
	if !found {
		return fmt.Errorf("host zone %s not found", hubChainId)
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

	// log the messages
	for _, msg := range msgs {
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
			"Undelegate ICA message: %s", msg.String()))
	}
	// print the full msgs array
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Undelegate ICA messages: %v", msgs))

	// print the unbondings
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Undelegate ICA unbondings: %v", unbondings))

	// print the epochUnbondingRecordIds
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
		"Undelegate ICA epochUnbondingRecordIds: %v", epochUnbondingRecordIds))

	EmitUndelegationEvent(ctx, hostZone, totalUnbondAmount)

	return nil
}
