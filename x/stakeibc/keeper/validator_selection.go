package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetValidatorDelegationAmtDifferences(ctx sdk.Context, hostZone types.HostZone) (map[string]int64, error) {
	/*
		This function returns a map from Validator Address to how many extra tokens
		need to be given to that validator

		positive implies extra tokens need to be given,
		negative impleis tokens need to be taken away
	*/
	validators := hostZone.GetValidators()
	delegationDelta := make(map[string]int64)
	totalDelegatedAmt := k.GetTotalValidatorDelegations(ctx, hostZone)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target weights for host zone %s", hostZone.ChainId))
		return nil, err
	}
	for _, validator := range validators {
		delegationDelta[validator.GetAddress()] = int64(targetDelegation[validator.GetAddress()]) - int64(validator.DelegationAmt)
	}
	return delegationDelta, nil
}

func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, finalDelegation uint64) (map[string]uint64, error) {
	// This will get the target validator delegation for the given hostZone
	// such that the total validator delegation is equal to the finalDelegation
	// output key is ADDRESS not NAME
	totalWeight := k.GetTotalValidatorWeight(ctx, hostZone)
	if finalDelegation == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("Cannot calculate target delegation if final amount is 0 %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	targetWeight := make(map[string]uint64)
	allocatedAmt := uint64(0)
	for i, validator := range hostZone.Validators {
		if i == len(hostZone.Validators)-1 {
			// for the last element, we need to make sure that the allocatedAmt is equal to the finalDelegation
			targetWeight[validator.GetAddress()] = finalDelegation - allocatedAmt
		} else {
			delegateAmt := uint64(float64(validator.Weight*finalDelegation) / float64(totalWeight))
			allocatedAmt += delegateAmt
			targetWeight[validator.GetAddress()] = delegateAmt
		}

	}
	return targetWeight, nil
}

func (k Keeper) GetTotalValidatorDelegations(ctx sdk.Context, hostZone types.HostZone) uint64 {
	validators := hostZone.GetValidators()
	total_delegation := uint64(0)
	for _, validator := range validators {
		total_delegation += validator.DelegationAmt
	}
	return total_delegation
}

func (k Keeper) GetTotalValidatorWeight(ctx sdk.Context, hostZone types.HostZone) uint64 {
	validators := hostZone.GetValidators()
	total_weight := uint64(0)
	for _, validator := range validators {
		total_weight += validator.Weight
	}
	return total_weight
}
