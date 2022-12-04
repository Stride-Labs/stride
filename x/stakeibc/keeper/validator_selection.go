package keeper

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cast"

	"github.com/Stride-Labs/stride/v4/utils"
	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// This function returns a map from Validator Address to how many extra tokens need to be given to that validator
//
//   positive implies extra tokens need to be given,
//   negative impleis tokens need to be taken away
func (k Keeper) GetValidatorDelegationAmtDifferences(ctx sdk.Context, hostZone types.HostZone) (map[string]int64, error) {
	validators := hostZone.GetValidators()
	delegationDelta := make(map[string]int64)
	totalDelegatedAmt := k.GetTotalValidatorDelegations(hostZone)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		k.Logger(ctx).Error(fmt.Sprintf("Error getting target val amts for host zone %s", hostZone.ChainId))
		return nil, err
	}
	for _, validator := range validators {
		targetDelForVal, err := cast.ToInt64E(targetDelegation[validator.GetAddress()])
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error getting target weight for host zone %s, targetDelForVal: %d", hostZone.ChainId, targetDelForVal))
			return nil, err
		}
		delAmt, err := cast.ToInt64E(validator.DelegationAmt)
		if err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Error getting target delAmt for host zone %s, amt: %d", hostZone.ChainId, delAmt))
			return nil, err
		}
		delegationDelta[validator.GetAddress()] = targetDelForVal - delAmt
	}
	return delegationDelta, nil
}

// This will get the target validator delegation for the given hostZone
// such that the total validator delegation is equal to the finalDelegation
// output key is ADDRESS not NAME
func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, finalDelegation uint64) (map[string]uint64, error) {
	totalWeight := k.GetTotalValidatorWeight(hostZone)
	if totalWeight == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("No non-zero validators found for host zone %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Total Validator Weight: %d", totalWeight))

	if finalDelegation == 0 {
		k.Logger(ctx).Error(fmt.Sprintf("Cannot calculate target delegation if final amount is 0 %s", hostZone.ChainId))
		return nil, types.ErrNoValidatorWeights
	}
	targetAmount := make(map[string]uint64)
	allocatedAmt := uint64(0)

	// sort validators by weight ascending, this is inplace sorting!
	validators := hostZone.GetValidators()

	for i, j := 0, len(validators)-1; i < j; i, j = i+1, j-1 {
		validators[i], validators[j] = validators[j], validators[i]
	}

	// Do not use `Slice` here, it is stochastic
	sort.SliceStable(validators, func(i, j int) bool {
		return validators[i].Weight < validators[j].Weight
	})

	for i, validator := range validators {
		if i == len(validators)-1 {
			// for the last element, we need to make sure that the allocatedAmt is equal to the finalDelegation
			targetAmount[validator.GetAddress()] = finalDelegation - allocatedAmt
		} else {
			delegateAmt, err := cast.ToUint64E(float64(validator.Weight*finalDelegation) / float64(totalWeight))
			if err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Error getting target weights for host zone %s", hostZone.ChainId))
				return nil, err
			}
			allocatedAmt += delegateAmt
			targetAmount[validator.GetAddress()] = delegateAmt
		}
	}
	return targetAmount, nil
}

func (k Keeper) GetTotalValidatorDelegations(hostZone types.HostZone) uint64 {
	validators := hostZone.GetValidators()
	total_delegation := uint64(0)
	for _, validator := range validators {
		total_delegation += validator.DelegationAmt
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
