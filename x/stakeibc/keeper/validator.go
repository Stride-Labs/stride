package keeper

import (
	"errors"
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v20/utils"
	"github.com/Stride-Labs/stride/v20/x/stakeibc/types"
)

// Get a validator and its index from a list of validators, by address
func GetValidatorFromAddress(validators []*types.Validator, address string) (val types.Validator, index int64, found bool) {
	for i, v := range validators {
		if v.Address == address {
			return *v, int64(i), true
		}
	}
	return types.Validator{}, 0, false
}

// Appends a validator to host zone (if the host zone is not already at capacity)
// If the validator is added through governance, the weight is equal to the minimum weight across the set
// If the validator is added through an admin transactions, the weight is specified in the message
func (k Keeper) AddValidatorToHostZone(ctx sdk.Context, chainId string, validator types.Validator, fromGovernance bool) error {
	// Get the corresponding host zone
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, "Host Zone (%s) not found", chainId)
	}

	// Check that we don't already have this validator
	// Grab the minimum weight in the process (to assign to validator's added through governance)
	var minWeight uint64 = math.MaxUint64
	for _, existingValidator := range hostZone.Validators {
		if existingValidator.Address == validator.Address {
			return errorsmod.Wrapf(types.ErrValidatorAlreadyExists, "Validator address (%s) already exists on Host Zone (%s)", validator.Address, chainId)
		}
		if existingValidator.Name == validator.Name {
			return errorsmod.Wrapf(types.ErrValidatorAlreadyExists, "Validator name (%s) already exists on Host Zone (%s)", validator.Name, chainId)
		}
		// Store the min weight to assign to new validator added through governance (ignore zero-weight validators)
		if existingValidator.Weight < minWeight && existingValidator.Weight > 0 {
			minWeight = existingValidator.Weight
		}
	}

	// If the validator was added via governance, set the weight to the min validator weight of the host zone
	valWeight := validator.Weight
	if fromGovernance {
		valWeight = minWeight
	}

	// Determine the slash query checkpoint for LSM liquid stakes
	checkpoint := k.GetUpdatedSlashQueryCheckpoint(ctx, hostZone.TotalDelegations)

	// Finally, add the validator to the host
	hostZone.Validators = append(hostZone.Validators, &types.Validator{
		Name:                      validator.Name,
		Address:                   validator.Address,
		Weight:                    valWeight,
		Delegation:                sdkmath.ZeroInt(),
		SlashQueryProgressTracker: sdkmath.ZeroInt(),
		SlashQueryCheckpoint:      checkpoint,
	})

	// Finally, confirm none of the validator's exceed the weight cap
	if err := k.CheckValidatorWeightsBelowCap(ctx, hostZone.Validators); err != nil {
		return err
	}

	k.SetHostZone(ctx, hostZone)

	return nil
}

// Removes a validator from a host zone
// The validator must be zero-weight and have no delegations in order to be removed
// There must also be no LSMTokenDeposits in progress since this would update the delegation on completion
func (k Keeper) RemoveValidatorFromHostZone(ctx sdk.Context, chainId string, validatorAddress string) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		errMsg := fmt.Sprintf("HostZone (%s) not found", chainId)
		k.Logger(ctx).Error(errMsg)
		return errorsmod.Wrapf(types.ErrHostZoneNotFound, errMsg)
	}

	// Check for LSMTokenDeposit records with this specific validator address
	lsmTokenDeposits := k.RecordsKeeper.GetAllLSMTokenDeposit(ctx)
	for _, lsmTokenDeposit := range lsmTokenDeposits {
		if lsmTokenDeposit.ValidatorAddress == validatorAddress {
			return errorsmod.Wrapf(types.ErrUnableToRemoveValidator, "Validator (%s) still has at least one LSMTokenDeposit (%+v)", validatorAddress, lsmTokenDeposit)
		}
	}

	for i, val := range hostZone.Validators {
		if val.GetAddress() == validatorAddress {
			if val.Delegation.IsZero() && val.Weight == 0 {
				hostZone.Validators = append(hostZone.Validators[:i], hostZone.Validators[i+1:]...)
				k.SetHostZone(ctx, hostZone)
				return nil
			}
			errMsg := fmt.Sprintf("Validator (%s) has non-zero delegation (%v) or weight (%d)", validatorAddress, val.Delegation, val.Weight)
			k.Logger(ctx).Error(errMsg)
			return errors.New(errMsg)
		}
	}
	errMsg := fmt.Sprintf("Validator address (%s) not found on host zone (%s)", validatorAddress, chainId)
	k.Logger(ctx).Error(errMsg)
	return errorsmod.Wrapf(types.ErrValidatorNotFound, errMsg)
}

// Updates a validator's individual delegation, and the corresponding total delegation on the host zone
// Note: This modifies the original host zone struct. The calling function must Set this host zone
// for changes to persist
func (k Keeper) AddDelegationToValidator(
	ctx sdk.Context,
	hostZone *types.HostZone,
	validatorAddress string,
	amount sdkmath.Int,
	callbackId string,
) error {
	for _, validator := range hostZone.Validators {
		if validator.Address == validatorAddress {
			k.Logger(ctx).Info(utils.LogICACallbackWithHostZone(hostZone.ChainId, callbackId,
				"  Validator %s, Current Delegation: %v, Delegation Change: %v", validator.Address, validator.Delegation, amount))

			// If the delegation change is negative, make sure it wont cause the delegation to fall below zero
			if amount.IsNegative() {
				if amount.Abs().GT(validator.Delegation) {
					return errorsmod.Wrapf(types.ErrValidatorDelegationChg,
						"Delegation change (%v) is greater than validator (%s) delegation %v",
						amount.Abs(), validatorAddress, validator.Delegation)
				}
				if amount.Abs().GT(hostZone.TotalDelegations) {
					return errorsmod.Wrapf(types.ErrValidatorDelegationChg,
						"Delegation change (%v) is greater than total delegation amount on host %s (%v)",
						amount.Abs(), hostZone.ChainId, hostZone.TotalDelegations)
				}
			}

			validator.Delegation = validator.Delegation.Add(amount)
			hostZone.TotalDelegations = hostZone.TotalDelegations.Add(amount)

			return nil
		}
	}

	return errorsmod.Wrapf(types.ErrValidatorNotFound,
		"Could not find validator %s on host zone %s", validatorAddress, hostZone.ChainId)
}

// Sum the total weights across each validator for a host zone
func (k Keeper) GetTotalValidatorWeight(validators []types.Validator) uint64 {
	totalWeight := uint64(0)
	for _, validator := range validators {
		totalWeight += validator.Weight
	}
	return totalWeight
}

// Increments the validators slash query progress tracker
func (k Keeper) IncrementValidatorSlashQueryProgress(
	ctx sdk.Context,
	chainId string,
	validatorAddress string,
	amount sdkmath.Int,
) error {
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return types.ErrHostZoneNotFound
	}

	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return types.ErrValidatorNotFound
	}

	// Increment the progress tracker
	oldProgress := validator.SlashQueryProgressTracker
	newProgress := validator.SlashQueryProgressTracker.Add(amount)
	validator.SlashQueryProgressTracker = newProgress

	// If the checkpoint is zero, it implies the TVL was 0 last time it was set, and we should
	// update it here
	// If the checkpoint is non-zero, only update it if it was just breached
	shouldUpdateCheckpoint := true
	if !validator.SlashQueryCheckpoint.IsZero() {
		oldInterval := oldProgress.Quo(validator.SlashQueryCheckpoint)
		newInterval := newProgress.Quo(validator.SlashQueryCheckpoint)
		shouldUpdateCheckpoint = oldInterval.LT(newInterval)
	}

	// Optionally re-calculate the checkpoint
	// Threshold of 1% means once 1% of TVL has been breached, the query is issued
	if shouldUpdateCheckpoint {
		validator.SlashQueryCheckpoint = k.GetUpdatedSlashQueryCheckpoint(ctx, hostZone.TotalDelegations)
	}

	hostZone.Validators[valIndex] = &validator
	k.SetHostZone(ctx, hostZone)

	return nil
}

// Increments the number of validator delegation changes in progress by 1
// Note: This modifies the original host zone struct. The calling function must Set this host zone
// for changes to persist
func (k Keeper) IncrementValidatorDelegationChangesInProgress(hostZone *types.HostZone, validatorAddress string) error {
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "validator %s not found", validatorAddress)
	}
	validator.DelegationChangesInProgress += 1
	hostZone.Validators[valIndex] = &validator
	return nil
}

// Decrements the number of validator delegation changes in progress by 1
// Note: This modifies the original host zone struct. The calling function must Set this host zone
// for changes to persist
func (k Keeper) DecrementValidatorDelegationChangesInProgress(hostZone *types.HostZone, validatorAddress string) error {
	validator, valIndex, found := GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		return errorsmod.Wrapf(types.ErrValidatorNotFound, "validator %s not found", validatorAddress)
	}
	if validator.DelegationChangesInProgress == 0 {
		return errorsmod.Wrapf(types.ErrInvalidValidatorDelegationUpdates,
			"cannot decrement the number of delegation updates if the validator has 0 updates in progress")
	}
	validator.DelegationChangesInProgress -= 1
	hostZone.Validators[valIndex] = &validator
	return nil
}

// Checks if any validator's portion of the weight is greater than the cap
func (k Keeper) CheckValidatorWeightsBelowCap(ctx sdk.Context, validators []*types.Validator) error {
	// If there's only a few validators, don't enforce this yet
	if len(validators) < MinValidatorsBeforeWeightCapCheck {
		return nil
	}

	// The weight cap in params is an int representing a percentage (e.g. 10 is 10%)
	params := k.GetParams(ctx)
	validatorWeightCap := float64(params.ValidatorWeightCap)

	// Store a map of each validator weight, as well as the total
	totalWeight := float64(0)
	weightsByValidator := map[string]float64{}
	for _, validator := range validators {
		weightsByValidator[validator.Address] = float64(validator.Weight)
		totalWeight += float64(validator.Weight)
	}

	// If the total validator weights are 0, exit prematurely
	if totalWeight == 0 {
		return nil
	}

	// Check if any validator exceeds the cap
	for _, address := range utils.StringMapKeys[float64](weightsByValidator) {
		weightPercentage := weightsByValidator[address] / totalWeight * 100
		if weightPercentage > validatorWeightCap {
			return errorsmod.Wrapf(types.ErrValidatorExceedsWeightCap,
				"validator %s exceeds weight cap, has %v%% of the total weight when the cap is %v%%",
				address, weightPercentage, validatorWeightCap)
		}
	}

	return nil
}
