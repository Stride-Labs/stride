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

// Rebalance validators according to their validator weights
// Specify whether to rebalance the balanced or unbalanced portion
// The unbalanced portion represents delegations from LSM Tokens that are naturally
//    misaligned with the validator. They must be rebalanced periodically
// The balance portion represents either native delegations or LSM delegation that have
//    already been rebalanced. This portion only requires a to be rebalanced if the validator weights change
func (k Keeper) RebalanceDelegations(ctx sdk.Context, chainId string, delegationType types.DelegationType) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrap(types.ErrHostZoneNotFound, fmt.Sprintf("Host zone %s not found", chainId))
	}
	validatorDeltas, err := k.GetValidatorDelegationDifferences(ctx, hostZone, types.BALANCED_DELEGATION)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get validator deltas for host zone %s", chainId)
	}

	// we convert the above map into a list of tuples
	type valPair struct {
		deltaAmt sdkmath.Int
		valAddr  string
	}
	valDeltaList := make([]valPair, 0)
	for _, valAddr := range utils.StringMapKeys(validatorDeltas) { // DO NOT REMOVE: StringMapKeys fixes non-deterministic map iteration
		deltaAmt := validatorDeltas[valAddr]
		k.Logger(ctx).Info(fmt.Sprintf("Adding deltaAmt: %v to validator: %s", deltaAmt, valAddr))
		valDeltaList = append(valDeltaList, valPair{deltaAmt, valAddr})
	}
	// now we sort that list
	lessFunc := func(i, j int) bool {
		return valDeltaList[i].deltaAmt.LT(valDeltaList[j].deltaAmt)
	}
	sort.SliceStable(valDeltaList, lessFunc)
	// now varDeltaList is sorted by deltaAmt
	overWeightIndex := 0
	underWeightIndex := len(valDeltaList) - 1

	// check if there is a large enough rebalance, if not, just exit
	total_delegation := k.GetTotalValidatorDelegations(hostZone, types.BALANCED_DELEGATION)
	if total_delegation.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"no validator delegations found for Host Zone %s, cannot rebalance 0 delegations!", hostZone.ChainId)
	}

	var msgs []sdk.Msg
	delegationIca := hostZone.GetDelegationAccount()
	if delegationIca == nil || delegationIca.GetAddress() == "" {
		k.Logger(ctx).Error(fmt.Sprintf("Zone %s is missing a delegation address!", hostZone.ChainId))
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidAddress, "Invalid delegation account")
	}

	delegatorAddress := delegationIca.GetAddress()

	// start construction callback
	rebalanceCallback := types.RebalanceCallback{
		HostZoneId:   hostZone.ChainId,
		Rebalancings: []*types.Rebalancing{},
	}

	for i := uint64(1); i <= msg.NumRebalance; i++ {
		underWeightElem := valDeltaList[underWeightIndex]
		overWeightElem := valDeltaList[overWeightIndex]
		if underWeightElem.deltaAmt.LT(sdkmath.ZeroInt()) {
			// if underWeightElem is negative, we're done rebalancing
			break
		}
		if overWeightElem.deltaAmt.GT(sdkmath.ZeroInt()) {
			// if overWeightElem is positive, we're done rebalancing
			break
		}
		// underweight Elem is positive, overweight Elem is negative
		overWeightElemAbs := overWeightElem.deltaAmt.Abs()
		var redelegateMsg *stakingtypes.MsgBeginRedelegate
		if underWeightElem.deltaAmt.GT(overWeightElemAbs) {
			// if the underweight element is more off than the overweight element
			// we transfer stake from the underweight element to the overweight element
			underWeightElem.deltaAmt = underWeightElem.deltaAmt.Sub(overWeightElemAbs)
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    delegatorAddress,
				ValidatorSrcAddress: overWeightElem.valAddr,
				ValidatorDstAddress: underWeightElem.valAddr,
				Amount:              sdk.NewCoin(hostZone.HostDenom, overWeightElemAbs)}
			msgs = append(msgs, redelegateMsg)
			overWeightElem.deltaAmt = sdkmath.ZeroInt()
		} else if underWeightElem.deltaAmt.LT(overWeightElemAbs) {
			// if the overweight element is more overweight than the underweight element
			overWeightElem.deltaAmt = overWeightElem.deltaAmt.Add(underWeightElem.deltaAmt)
			underWeightIndex -= 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    delegatorAddress,
				ValidatorSrcAddress: overWeightElem.valAddr,
				ValidatorDstAddress: underWeightElem.valAddr,
				Amount:              sdk.NewCoin(hostZone.HostDenom, underWeightElem.deltaAmt)}
			msgs = append(msgs, redelegateMsg)
			underWeightElem.deltaAmt = sdkmath.ZeroInt()
		} else {
			// if the two elements are equal, we increment both slices
			underWeightIndex -= 1
			overWeightIndex += 1
			// issue an ICA call to the host zone to rebalance the validator
			redelegateMsg = &stakingtypes.MsgBeginRedelegate{
				DelegatorAddress:    delegatorAddress,
				ValidatorSrcAddress: overWeightElem.valAddr,
				ValidatorDstAddress: underWeightElem.valAddr,
				Amount:              sdk.NewCoin(hostZone.HostDenom, underWeightElem.deltaAmt)}
			msgs = append(msgs, redelegateMsg)
			overWeightElem.deltaAmt = sdkmath.ZeroInt()
			underWeightElem.deltaAmt = sdkmath.ZeroInt()
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
	// marshall the callback
	marshalledCallbackArgs, err := k.MarshalRebalanceCallbackArgs(ctx, rebalanceCallback)
	if err != nil {
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	connectionId := hostZone.GetConnectionId()
	_, err = k.SubmitTxsStrideEpoch(ctx, connectionId, msgs, *hostZone.GetDelegationAccount(), ICACallbackID_Rebalance, marshalledCallbackArgs)
	if err != nil {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "Failed to SubmitTxs for %s, %s, %s, %s", connectionId, hostZone.ChainId, msgs, err.Error())
	}

}

// This function returns a map from Validator Address to how many extra tokens need to be given to that validator
//   positive implies extra tokens need to be given,
//   negative implies tokens need to be taken away
func (k Keeper) GetValidatorDelegationDifferences(
	ctx sdk.Context,
	hostZone types.HostZone,
	delegationType types.DelegationType,
) (map[string]sdkmath.Int, error) {
	delegationDelta := make(map[string]sdkmath.Int)
	totalDelegatedAmt := k.GetTotalValidatorDelegations(hostZone, delegationType)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target val amts for host zone %s", hostZone.ChainId))
		return nil, err
	}

	for _, validator := range hostZone.Validators {
		targetDelForVal := targetDelegation[validator.Address]

		if delegationType == types.BALANCED_DELEGATION {
			delegationDelta[validator.Address] = targetDelForVal.Sub(validator.BalancedDelegation)
		} else if delegationType == types.UNBALANCED_DELEGATION {
			delegationDelta[validator.Address] = targetDelForVal.Sub(validator.UnbalancedDelegation)
		}
	}

	return delegationDelta, nil
}

// This will get the target validator delegation for the given hostZone
// such that the total validator delegation is equal to the finalDelegation
// output key is ADDRESS not NAME
func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, finalDelegation sdkmath.Int) (map[string]sdkmath.Int, error) {
	// Confirm the expected delegation amount is greater than 0
	if finalDelegation.Equal(sdkmath.ZeroInt()) {
		k.Logger(ctx).Error(fmt.Sprintf("Cannot calculate target delegation if final amount is 0 %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}

	// Sum the total weight across all validators
	totalWeight := k.GetTotalValidatorWeight(hostZone)
	if totalWeight == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("No non-zero validators found for host zone %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Total Validator Weight: %d", totalWeight))

	// sort validators by weight ascending, this is inplace sorting!
	validators := hostZone.Validators

	for i, j := 0, len(validators)-1; i < j; i, j = i+1, j-1 {
		validators[i], validators[j] = validators[j], validators[i]
	}

	sort.SliceStable(validators, func(i, j int) bool { // Do not use `Slice` here, it is stochastic
		return validators[i].Weight < validators[j].Weight
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
	validators := hostZone.GetValidators()
	total_delegation := sdkmath.ZeroInt()
	for _, validator := range validators {
		if delegationType == types.BALANCED_DELEGATION {
			total_delegation = total_delegation.Add(validator.BalancedDelegation)
		} else if delegationType == types.UNBALANCED_DELEGATION {
			total_delegation = total_delegation.Add(validator.UnbalancedDelegation)
		}
	}
	return total_delegation
}

// Sum the total weights across each validator for a host zone
func (k Keeper) GetTotalValidatorWeight(hostZone types.HostZone) uint64 {
	validators := hostZone.GetValidators()
	total_weight := uint64(0)
	for _, validator := range validators {
		total_weight += validator.Weight
	}
	return total_weight
}
