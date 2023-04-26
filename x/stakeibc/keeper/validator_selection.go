package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type RebalanceValidatorDelegationChange struct {
	ValidatorAddress string
	Delta            sdkmath.Int
}

// Rebalance validators according to their validator weights
// Specify whether to rebalance the balanced or unbalanced portion
// The unbalanced portion represents delegations from LSM Tokens that are naturally
//    misaligned with the validator. They must be rebalanced periodically
// The balance portion represents either native delegations or LSM delegation that have
//    already been rebalanced. This portion only requires a rebalancing if the validator weights change
func (k Keeper) RebalanceDelegations(ctx sdk.Context, chainId string, numRebalance uint64) error {
	// Get the host zone and confirm the delegation account is initialized
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrap(types.ErrHostZoneNotFound, fmt.Sprintf("Host zone %s not found", chainId))
	}
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}

	// Get the difference between the actual and expected validator delegations
	valDeltaList, err := k.GetValidatorDelegationDifferences(ctx, hostZone)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get validator deltas for host zone %s", chainId)
	}

	msgs, rebalacings := k.GetRebalanceICAMessages(hostZone, valDeltaList, numRebalance)

	// marshall the callback
	rebalanceCallback := types.RebalanceCallback{
		HostZoneId:   hostZone.ChainId,
		Rebalancings: rebalacings,
	}
	rebalanceCallbackBz, err := k.MarshalRebalanceCallbackArgs(ctx, rebalanceCallback)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal rebalance callback args")
	}

	// Submit the rebalance ICA
	_, err = k.SubmitTxsStrideEpoch(
		ctx,
		hostZone.ConnectionId,
		msgs,
		types.ICAAccountType_DELEGATION,
		ICACallbackID_Rebalance,
		rebalanceCallbackBz,
	)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs for %s, messages: %+v", hostZone.ChainId, msgs)
	}

	return nil
}

// Given a list of target delegation changes, builds the individual re-delegation messages by redelegating
// from overweight validators to undersweight validators
// Also returns the callback data for the ICA
func (k Keeper) GetRebalanceICAMessages(
	hostZone types.HostZone,
	validatorDeltas []RebalanceValidatorDelegationChange,
	numRebalance uint64,
) (msgs []sdk.Msg, rebalancings []*types.Rebalancing) {
	// Sort the list of delegation changes by the size of the change
	lessFunc := func(i, j int) bool {
		if !validatorDeltas[i].Delta.Equal(validatorDeltas[j].Delta) {
			return validatorDeltas[i].Delta.LT(validatorDeltas[j].Delta)
		}
		// use name as a tie breaker if deltas are equal
		return validatorDeltas[i].ValidatorAddress < validatorDeltas[j].ValidatorAddress
	}
	sort.SliceStable(validatorDeltas, lessFunc)

	// QUESTION: Does anyone have a preference on the naming convention here? A couple options:
	//  * Overweight/Underweight (current)
	//  * Surplus/Deficit
	//  * Source/Destination
	//  * Giver/Receiver
	// etc.

	// Pair overweight and underweight validators, with a redelegation from the overweight
	// validator to the underweight one
	// The list is sorted with the overweight validators (who should lose stake) at index 0
	//   and the underweight validators (who should gain stake) at index N-1
	// The overweight validator's have a negative delta and the underweight validators have a positive delta
	overWeightIndex := 0
	underWeightIndex := len(validatorDeltas) - 1
	for i := uint64(1); i <= numRebalance; i++ {
		// underweight Elem is positive, overweight Elem is negative
		underWeightValidator := validatorDeltas[underWeightIndex]
		overWeightValidator := validatorDeltas[overWeightIndex]

		// If either delta is 0, we're done rebalancing
		if underWeightValidator.Delta.IsZero() || overWeightValidator.Delta.IsZero() {
			break
		}

		var redelegationAmount sdkmath.Int
		if underWeightValidator.Delta.Abs().GT(overWeightValidator.Delta.Abs()) {
			// If the underweight validator is more underweight than the overweight validator,
			// transfer all the overweight validator's surplus to the underweight validator
			redelegationAmount = overWeightValidator.Delta.Abs()

			// Update the underweight validator, and zero out the overweight validator
			validatorDeltas[underWeightIndex].Delta = underWeightValidator.Delta.Sub(redelegationAmount)
			validatorDeltas[overWeightIndex].Delta = sdkmath.ZeroInt()
			overWeightIndex += 1

		} else if overWeightValidator.Delta.Abs().GT(underWeightValidator.Delta.Abs()) {
			// If the overweight validator is more overweight than the underweight validator,
			// transfer only up to an an amount equal to the underweight validator's deficit
			redelegationAmount = underWeightValidator.Delta

			// Update the overweight validator, and zero out the underweight validator
			validatorDeltas[overWeightIndex].Delta = overWeightValidator.Delta.Add(redelegationAmount)
			validatorDeltas[underWeightIndex].Delta = sdkmath.ZeroInt()
			underWeightIndex -= 1

		} else {
			// if the overweight validator's surplus is equal to the underweight validator's deficit,
			// we'll transfer that amount and both validators will now be balanced
			redelegationAmount = underWeightValidator.Delta

			validatorDeltas[overWeightIndex].Delta = sdkmath.ZeroInt()
			validatorDeltas[underWeightIndex].Delta = sdkmath.ZeroInt()

			overWeightIndex += 1
			underWeightIndex -= 1
		}

		// Append the new Redelegation message and Rebalancing struct for the callback
		// We always send from the overweight validator to the underweight validator
		srcValidator := overWeightValidator.ValidatorAddress
		dstValidator := underWeightValidator.ValidatorAddress

		msgs = append(msgs, &stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    hostZone.DelegationIcaAddress,
			ValidatorSrcAddress: srcValidator,
			ValidatorDstAddress: dstValidator,
			Amount:              sdk.NewCoin(hostZone.HostDenom, redelegationAmount),
		})
		rebalancings = append(rebalancings, &types.Rebalancing{
			SrcValidator: srcValidator,
			DstValidator: dstValidator,
			Amt:          redelegationAmount,
		})
	}

	return msgs, rebalancings
}

// This function returns a list with the number of extra tokens that needs to be sent to each validator
//   positive implies extra tokens need to be given,
//   negative implies tokens need to be taken away
func (k Keeper) GetValidatorDelegationDifferences(ctx sdk.Context, hostZone types.HostZone) ([]RebalanceValidatorDelegationChange, error) {
	// Get the target delegation amount for each validator
	totalDelegatedAmt := k.GetTotalValidatorDelegations(hostZone)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId)
	}

	// For each validator, store the amount that their delegation should change
	delegationDeltas := []RebalanceValidatorDelegationChange{}
	totalDelegationChange := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		// Compare the target with either the current delegation
		delegationChange := targetDelegation[validator.Address].Sub(validator.Delegation)

		// Only include validators who's delegation should change
		if !delegationChange.IsZero() {
			delegationDeltas = append(delegationDeltas, RebalanceValidatorDelegationChange{
				ValidatorAddress: validator.Address,
				Delta:            delegationChange,
			})
			totalDelegationChange = totalDelegationChange.Add(delegationChange)
		}
		k.Logger(ctx).Info(fmt.Sprintf("Adding delegation: %v to validator: %s", delegationChange, validator.Address))
	}

	// Sanity check that the sum of all the delegation change's is equal to 0
	// (meaning the total delegation across ALL validators has not changed)
	if !totalDelegationChange.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"non-zero net delegation change (%v) across validators during rebalancing", totalDelegationChange)
	}

	return delegationDeltas, nil
}

// This will get the target validator delegation for the given hostZone
// such that the total validator delegation is equal to the finalDelegation
// output key is ADDRESS not NAME
func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, finalDelegation sdkmath.Int) (map[string]sdkmath.Int, error) {
	// Confirm the expected delegation amount is greater than 0
	if finalDelegation.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"Cannot calculate target delegation if final amount is 0 %s", hostZone.ChainId)
	}

	// Sum the total weight across all validators
	totalWeight := k.GetTotalValidatorWeight(hostZone)
	if totalWeight == 0 {
		return nil, errorsmod.Wrapf(types.ErrNoValidatorWeights,
			"No non-zero validators found for host zone %s", hostZone.ChainId)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Total Validator Weight: %d", totalWeight))

	// sort validators by weight ascending, this is inplace sorting!
	// QUESTION: Now that we're sorting by weight + name, should we remove this?
	validators := hostZone.Validators
	for i, j := 0, len(validators)-1; i < j; i, j = i+1, j-1 {
		validators[i], validators[j] = validators[j], validators[i]
	}

	sort.SliceStable(validators, func(i, j int) bool { // Do not use `Slice` here, it is stochastic
		if validators[i].Weight != validators[j].Weight {
			return validators[i].Weight < validators[j].Weight
		}
		// use name for tie breaker if weights are equal
		return validators[i].Address < validators[j].Address
	})

	// Assign each validator their portion of the delegation (and give any overflow to the last validator)
	targetUnbondingsByValidator := make(map[string]sdkmath.Int)
	totalAllocated := sdkmath.ZeroInt()
	for i, validator := range validators {
		// For the last element, we need to make sure that the totalAllocated is equal to the finalDelegation
		if i == len(validators)-1 {
			targetUnbondingsByValidator[validator.Address] = finalDelegation.Sub(totalAllocated)
		} else {
			delegateAmt := sdkmath.NewIntFromUint64(validator.Weight).Mul(finalDelegation).Quo(sdkmath.NewIntFromUint64(totalWeight))
			totalAllocated = totalAllocated.Add(delegateAmt)
			targetUnbondingsByValidator[validator.Address] = delegateAmt
		}
	}

	return targetUnbondingsByValidator, nil
}

// Sum the total delegation across each validator for a host zone
// Must specify whether to sum the balanced or unbalanced delegation
func (k Keeper) GetTotalValidatorDelegations(hostZone types.HostZone) sdkmath.Int {
	validators := hostZone.Validators
	totalDelegation := sdkmath.ZeroInt()
	for _, validator := range validators {
		totalDelegation = totalDelegation.Add(validator.Delegation)
	}
	return totalDelegation
}

// Sum the total weights across each validator for a host zone
func (k Keeper) GetTotalValidatorWeight(hostZone types.HostZone) uint64 {
	validators := hostZone.Validators
	totalWeight := uint64(0)
	for _, validator := range validators {
		totalWeight += validator.Weight
	}
	return totalWeight
}
