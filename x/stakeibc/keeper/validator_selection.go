package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
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
		delegationDelta[validator.GetAddress()] = cast.ToInt64(targetDelegation[validator.GetAddress()]) - cast.ToInt64(validator.DelegationAmt)
	}
	return delegationDelta, nil
}

func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, finalDelegation uint64) (map[string]uint64, error) {
	// This will get the target validator delegation for the given hostZone
	// such that the total validator delegation is equal to the finalDelegation
	// output key is ADDRESS not NAME
	totalWeight := k.GetTotalValidatorWeight(ctx, hostZone)
	if totalWeight == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("No non-zero validators found for host zone %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	if finalDelegation == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("Cannot calculate target delegation if final amount is 0 %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	targetAmount := make(map[string]uint64)
	allocatedAmt := uint64(0)
	for i, validator := range hostZone.Validators {
		if i == len(hostZone.Validators)-1 {
			// for the last element, we need to make sure that the allocatedAmt is equal to the finalDelegation
			targetAmount[validator.GetAddress()] = finalDelegation - allocatedAmt
		} else {
			delegateAmt := cast.ToUint64(float64(validator.Weight*finalDelegation) / float64(totalWeight))
			allocatedAmt += delegateAmt
			targetAmount[validator.GetAddress()] = delegateAmt
		}

	}
	return targetAmount, nil
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
