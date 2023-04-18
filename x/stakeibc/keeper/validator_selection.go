package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v8/utils"
	"github.com/Stride-Labs/stride/v8/x/stakeibc/types"
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
func (k Keeper) RebalanceDelegations(ctx sdk.Context, chainId string, delegationType types.DelegationType, numRebalance uint64) error {
	// Get the host zone and confirm the delegation account is initialized
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrap(types.ErrHostZoneNotFound, fmt.Sprintf("Host zone %s not found", chainId))
	}
	delegationIca := hostZone.DelegationAccount
	if delegationIca == nil || delegationIca.Address == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}

	// Get the difference between the actual and expected validator delegations
	valDeltaList, err := k.GetValidatorDelegationDifferences(ctx, hostZone, delegationType)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get validator deltas for host zone %s", chainId)
	}

	msgs, rebalanceCallback := k.GetRebalanceICAMessages(hostZone, valDeltaList, delegationType, numRebalance)

	// marshall the callback
	marshalledCallbackArgs, err := k.MarshalRebalanceCallbackArgs(ctx, rebalanceCallback)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal rebalance callback args")
	}

	// Submit the rebalance ICA
	_, err = k.SubmitTxsStrideEpoch(
		ctx,
		hostZone.ConnectionId,
		msgs,
		*hostZone.DelegationAccount,
		ICACallbackID_Rebalance,
		marshalledCallbackArgs,
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
	delegationType types.DelegationType,
	numRebalance uint64,
) (messages []sdk.Msg, callback types.RebalanceCallback) {
	// Sort the list of delegation changes by the size of the change
	lessFunc := func(i, j int) bool {
		if !validatorDeltas[i].Delta.Equal(validatorDeltas[j].Delta) {
			return validatorDeltas[i].Delta.LT(validatorDeltas[j].Delta)
		}
		// use name as a tie breaker if deltas are equal
		return validatorDeltas[i].ValidatorAddress < validatorDeltas[j].ValidatorAddress
	}
	sort.SliceStable(validatorDeltas, lessFunc)

	// start construction callback
	rebalanceCallback := types.RebalanceCallback{
		HostZoneId:     hostZone.ChainId,
		DelegationType: delegationType,
		Rebalancings:   []*types.Rebalancing{},
	}

	var msgs []sdk.Msg
	overWeightIndex := 0
	underWeightIndex := len(validatorDeltas) - 1
	for i := uint64(1); i <= numRebalance; i++ {
		underWeightElem := validatorDeltas[underWeightIndex]
		overWeightElem := validatorDeltas[overWeightIndex]
		if underWeightElem.Delta.LT(sdkmath.ZeroInt()) {
			// if underWeightElem is negative, we're done rebalancing
			break
		}
		if overWeightElem.Delta.GT(sdkmath.ZeroInt()) {
			// if overWeightElem is positive, we're done rebalancing
			break
		}
		// underweight Elem is positive, overweight Elem is negative
		overWeightElemAbs := overWeightElem.Delta.Abs()
		var redelegateMsg *stakingtypes.MsgBeginRedelegate
		if underWeightElem.Delta.GT(overWeightElemAbs) {
			// if the underweight element is more off than the overweight element
			// we transfer stake from the underweight element to the overweight element
			underWeightElem.Delta = underWeightElem.Delta.Sub(overWeightElemAbs)
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    hostZone.DelegationAccount.Address,
				ValidatorSrcAddress: overWeightElem.ValidatorAddress,
				ValidatorDstAddress: underWeightElem.ValidatorAddress,
				Amount:              sdk.NewCoin(hostZone.HostDenom, overWeightElemAbs)}
			msgs = append(msgs, redelegateMsg)
			overWeightElem.Delta = sdkmath.ZeroInt()
		} else if underWeightElem.Delta.LT(overWeightElemAbs) {
			// if the overweight element is more overweight than the underweight element
			overWeightElem.Delta = overWeightElem.Delta.Add(underWeightElem.Delta)
			underWeightIndex -= 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    hostZone.DelegationAccount.Address,
				ValidatorSrcAddress: overWeightElem.ValidatorAddress,
				ValidatorDstAddress: underWeightElem.ValidatorAddress,
				Amount:              sdk.NewCoin(hostZone.HostDenom, underWeightElem.Delta)}
			msgs = append(msgs, redelegateMsg)
			underWeightElem.Delta = sdkmath.ZeroInt()
		} else {
			// if the two elements are equal, we increment both slices
			underWeightIndex -= 1
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    hostZone.DelegationAccount.Address,
				ValidatorSrcAddress: overWeightElem.ValidatorAddress,
				ValidatorDstAddress: underWeightElem.ValidatorAddress,
				Amount:              sdk.NewCoin(hostZone.HostDenom, underWeightElem.Delta)}
			msgs = append(msgs, redelegateMsg)
			overWeightElem.Delta = sdkmath.ZeroInt()
			underWeightElem.Delta = sdkmath.ZeroInt()
		}
		// add the rebalancing to the callback
		// lastMsg grabs rebalanceMsg from above (due to the type, it's hard to )
		// lastMsg := (msgs[len(msgs)-1]).(*stakingTypes.MsgBeginRedelegate)
		rebalanceCallback.Rebalancings = append(rebalanceCallback.Rebalancings, &types.Rebalancing{
			SrcValidator: redelegateMsg.ValidatorSrcAddress,
			DstValidator: redelegateMsg.ValidatorDstAddress,
			Amt:          redelegateMsg.Amount.Amount,
		})
	}

	return msgs, rebalanceCallback
}

// This function returns a list with the number of extra tokens that needs to be sent to each validator
//   positive implies extra tokens need to be given,
//   negative implies tokens need to be taken away
func (k Keeper) GetValidatorDelegationDifferences(
	ctx sdk.Context,
	hostZone types.HostZone,
	delegationType types.DelegationType, // QUESTION: Is delegationType more clear as an enum or as a boolean?
) ([]RebalanceValidatorDelegationChange, error) {
	totalDelegatedAmt := k.GetTotalValidatorDelegations(hostZone, delegationType)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId)
	}

	delegationDelta := make([]RebalanceValidatorDelegationChange, len(hostZone.Validators))
	for i, validator := range hostZone.Validators {
		targetDelForVal := targetDelegation[validator.Address]

		var delegationChange sdkmath.Int
		if delegationType == types.DelegationType_BALANCED {
			delegationChange = targetDelForVal.Sub(validator.BalancedDelegation)
		} else if delegationType == types.DelegationType_UNBALANCED {
			delegationChange = targetDelForVal.Sub(validator.UnbalancedDelegation)
		}

		delegationDelta[i] = RebalanceValidatorDelegationChange{
			ValidatorAddress: validator.Address,
			Delta:            delegationChange,
		}
		k.Logger(ctx).Info(fmt.Sprintf("Adding delegation: %v to validator: %s", delegationChange, validator.Address))
	}

	return delegationDelta, nil
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
func (k Keeper) GetTotalValidatorDelegations(hostZone types.HostZone, delegationType types.DelegationType) sdkmath.Int {
	validators := hostZone.Validators
	totalDelegation := sdkmath.ZeroInt()
	for _, validator := range validators {
		if delegationType == types.DelegationType_BALANCED {
			totalDelegation = totalDelegation.Add(validator.BalancedDelegation)
		} else if delegationType == types.DelegationType_UNBALANCED {
			totalDelegation = totalDelegation.Add(validator.UnbalancedDelegation)
		}
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
