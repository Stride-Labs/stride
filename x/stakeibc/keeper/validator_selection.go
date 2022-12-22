package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/utils"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// This function returns a map from Validator Address to how many extra tokens need to be given to that validator
//
//   positive implies extra tokens need to be given,
//   negative impleis tokens need to be taken away
func (k Keeper) GetValidatorDelegationAmtDifferences(ctx sdk.Context, hostZone types.HostZone) (map[string]sdk.Int, error) {
	validators := hostZone.GetValidators()
	delegationDelta := make(map[string]sdk.Int)
	totalDelegatedAmt := k.GetTotalValidatorDelegations(hostZone)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target val amts for host zone %s", hostZone.ChainId))
		return nil, err
	}
	for _, validator := range validators {
		targetDelForVal := targetDelegation[validator.GetAddress()]
		delegationDelta[validator.GetAddress()] = targetDelForVal.Sub(validator.DelegationAmt)
	}
	return delegationDelta, nil
}

// This will get the target validator delegation for the given hostZone
// such that the total validator delegation is equal to the finalDelegation
// output key is ADDRESS not NAME
func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, finalDelegation sdk.Int) (map[string]sdk.Int, error) {
	// Confirm the expected delegation amount is greater than 0
	if finalDelegation == sdk.ZeroInt() {
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
	targetUnbondingsByValidator := make(map[string]sdk.Int)
	totalAllocated := sdk.ZeroInt()
	for i, validator := range validators {
		// For the last element, we need to make sure that the totalAllocated is equal to the finalDelegation
		if i == len(validators)-1 {
			targetUnbondingsByValidator[validator.Address] = finalDelegation.Sub(totalAllocated)
		} else {
			delegateAmt := sdk.NewIntFromUint64(validator.Weight).Mul(finalDelegation).Quo(sdk.NewIntFromUint64(totalWeight))
			totalAllocated = totalAllocated.Add(delegateAmt)
			targetUnbondingsByValidator[validator.Address] = delegateAmt
		}
	}

	return targetUnbondingsByValidator, nil
}

func (k Keeper) GetTotalValidatorDelegations(hostZone types.HostZone) sdk.Int {
	validators := hostZone.GetValidators()
	total_delegation := sdk.ZeroInt()
	for _, validator := range validators {
		total_delegation = total_delegation.Add(validator.DelegationAmt)
	}
	return total_delegation
}

func (k Keeper) GetTotalValidatorWeight(hostZone types.HostZone) uint64 {
	validators := hostZone.GetValidators()
	total_weight := uint64(0)
	for _, validator := range validators {
		total_weight += validator.Weight
	}
	return total_weight
}
