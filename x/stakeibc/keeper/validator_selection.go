package keeper

import (
	"fmt"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetValidatorAmtDifferences(ctx sdk.Context, hostZone types.HostZone) (map[string]int64, error) {
	/*
		This function returns a map from Validator Address to how many extra tokens
		need to be given to that validator (posit)
	*/
	validators := hostZone.GetValidators()
	scaled_weights := make(map[string]int64)
	target_weights, err := k.GetTargetValAmtsForHostZone(ctx, hostZone)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target weights for host zone %s", hostZone.ChainId))
		return nil, err
	}
	for _, validator := range validators {
		scaled_weights[validator.Address] = int64(target_weights[validator.Address]) - int64(validator.DelegationAmt)
	}
	return scaled_weights, nil
}

func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone) (map[string]uint64, error) {
	totalDelegated := k.GetTotalValidatorDelegations(ctx, hostZone)
	// grab total weight of all validators
	totalWeight := uint64(0)
	for _, validator := range hostZone.Validators {
		totalWeight += validator.Weight
	}
	if totalWeight == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("Total weight is 0 for host zone %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	var targetWeight map[string]uint64
	for _, validator := range hostZone.Validators {
		targetWeight[validator.Address] = uint64(float64(validator.Weight*totalDelegated) / float64(totalWeight))
	}
	return targetWeight, nil
}

func (k Keeper) GetTotalValidatorDelegations(ctx sdk.Context, hostZone types.HostZone) uint64 {
	validators := hostZone.GetValidators()
	total_weight := uint64(0)
	for _, validator := range validators {
		total_weight += validator.DelegationAmt
	}
	return total_weight
}
