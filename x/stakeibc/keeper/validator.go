package keeper

import (
	"errors"
	"fmt"
	"math"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v18/x/stakeibc/types"
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
