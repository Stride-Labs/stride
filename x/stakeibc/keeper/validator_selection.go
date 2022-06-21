package keeper

import (
	"github.com/Stride-Labs/stride/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) GetValidatorAmtDifferences(ctx sdk.Context, hostZone types.HostZone) map[string]int64 {
	/*
		This function returns a map from Validator Address to how many extra tokens
		need to be given to that validator (posit)
	*/
	validators := hostZone.GetValidators()
	scaled_weights := make(map[string]int64)
	target_weights := k.GetTargetValAmtsForHostZone(ctx, hostZone)
	for _, validator := range validators {
		scaled_weights[validator.Address] = target_weights[validator.Address] - validator.DelegationAmt
	}
	return scaled_weights
}

func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone) map[string]int64 {
	var out map[string]types.ValidatorWeights
	k.paramstore.Get(ctx, types.KeyHostZoneValidatorWeights, &out)
	valWeights := out[hostZone.ChainId]
	totalWeight := int64(0)
	totalDelegatedWeight := k.GetTotalValidatorWeight()
	for _, weight := range valWeights.ValidatorWeights {
		totalWeight += weight
	}
	var targetWeight map[string]int64
	for valAddr, weight := range valWeights.ValidatorWeights {
		targetWeight[valAddr] = int64(float64(weight*totalDelegatedWeight) / float64(totalWeight))
	}
	return targetWeight
}

func (k Keeper) GetTotalValidatorWeight(ctx sdk.Context, hostZone types.HostZone) int64 {
	validators := hostZone.GetValidators()
	total_weight := int64(0)
	for _, validator := range validators {
		// QUESTION do we need to do some error handling here if delegaitonAmt is missing?
		total_weight += validator.DelegationAmt
	}
	return total_weight
}
