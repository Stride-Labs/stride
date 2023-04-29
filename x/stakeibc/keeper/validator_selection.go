package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/Stride-Labs/stride/v9/utils"
	"github.com/Stride-Labs/stride/v9/x/stakeibc/types"
)

type RebalanceValidatorDelegationChange struct {
	ValidatorAddress string
	Delta            sdkmath.Int
}

// Iterate each active host zone and issues redelegation messages to rebalance each
//   validator's stake according to their weights
// This is required when accepting LSM LiquidStakes as the distribution of stake
//   from the LSM Tokens will be inconsistend with the host zone's validator set
// Note: this cannot be run more than once in a single unbonding period
func (k Keeper) RebalanceAllHostZones(ctx sdk.Context, dayNumber uint64) {
	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		numRebalance := uint64(len(hostZone.Validators))

		if dayNumber%hostZone.UnbondingPeriod != 0 {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
				"Host does not rebalance this epoch (Unbonding Period: %d, Epoch: %d)", hostZone.UnbondingPeriod, dayNumber))
			continue
		}

		if err := k.RebalanceDelegationsForHostZone(ctx, hostZone.ChainId, numRebalance); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to rebalance delegations for %s: %s", hostZone.ChainId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Successfully rebalanced delegations"))
	}
}

// Rebalance validators according to their validator weights for a specific host zone
func (k Keeper) RebalanceDelegationsForHostZone(ctx sdk.Context, chainId string, numRebalance uint64) error {
	// Get the host zone and confirm the delegation account is initialized
	hostZone, found := k.GetHostZone(ctx, chainId)
	if !found {
		return errorsmod.Wrap(types.ErrHostZoneNotFound, fmt.Sprintf("Host zone %s not found", chainId))
	}
	if hostZone.DelegationIcaAddress == "" {
		return errorsmod.Wrapf(types.ErrICAAccountNotFound, "no delegation account found for %s", chainId)
	}

	// Get the difference between the actual and expected validator delegations
	valDeltaList, err := k.GetValidatorDelegationDifferences(ctx, hostZone)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to get validator deltas for host zone %s", chainId)
	}

	msgs, rebalacings := k.GetRebalanceICAMessages(hostZone, valDeltaList, numRebalance)

	// marshall the callback
	rebalanceCallback := types.RebalanceCallback{
		HostZoneId:   hostZone.ChainId,
		Rebalancings: rebalacings,
	}
	rebalanceCallbackBz, err := k.MarshalRebalanceCallbackArgs(ctx, rebalanceCallback)
	if err != nil {
		return errorsmod.Wrapf(err, "unable to marshal rebalance callback args")
	}

	// Submit the rebalance ICA
	_, err = k.SubmitTxsStrideEpoch(
		ctx,
		hostZone.ConnectionId,
		msgs,
		types.ICAAccountType_DELEGATION,
		ICACallbackID_Rebalance,
		rebalanceCallbackBz,
	)
	if err != nil {
		return errorsmod.Wrapf(err, "Failed to SubmitTxs for %s, messages: %+v", hostZone.ChainId, msgs)
	}

	return nil
}

// Given a list of target delegation changes, builds the individual re-delegation messages by redelegating
// from surplus validators to deficit validators
// Returns the list of messages and the callback data for the ICA
func (k Keeper) GetRebalanceICAMessages(
	hostZone types.HostZone,
	validatorDeltas []RebalanceValidatorDelegationChange,
	numRebalance uint64,
) (msgs []sdk.Msg, rebalancings []*types.Rebalancing) {
	// Sort the list of delegation changes by the size of the change
	// Sort descending so the surplus validators appear first
	lessFunc := func(i, j int) bool {
		if !validatorDeltas[i].Delta.Equal(validatorDeltas[j].Delta) {
			return validatorDeltas[i].Delta.GT(validatorDeltas[j].Delta)
		}
		// use name as a tie breaker if deltas are equal
		return validatorDeltas[i].ValidatorAddress < validatorDeltas[j].ValidatorAddress
	}
	sort.SliceStable(validatorDeltas, lessFunc)

	// Pair surplus and deficit validators, with a redelegation from the surplus
	// validator to the deficit one
	// The list is sorted with the surplus validators (who should lose stake) at index 0
	//   and the deficit validators (who should gain stake) at index N-1
	// The surplus validator's have a positive delta and the deficit validators have a negative delta
	surplusIndex := 0
	deficitIndex := len(validatorDeltas) - 1
	for i := uint64(1); i <= numRebalance; i++ {
		// surplus validator is positive, deficit validator is negative
		deficitValidator := validatorDeltas[deficitIndex]
		surplusValidator := validatorDeltas[surplusIndex]

		// If the indicies flipped, or either delta is 0, we're done rebalancing
		if surplusIndex > deficitIndex || deficitValidator.Delta.IsZero() || surplusValidator.Delta.IsZero() {
			break
		}

		var redelegationAmount sdkmath.Int
		if deficitValidator.Delta.Abs().GT(surplusValidator.Delta.Abs()) {
			// If the deficit validator needs more stake than the surplus validator has to give,
			// transfer the full surplus to deficit validator
			redelegationAmount = surplusValidator.Delta.Abs()

			// Update the deficit validator, and zero out the surplus validator
			validatorDeltas[deficitIndex].Delta = deficitValidator.Delta.Add(redelegationAmount)
			validatorDeltas[surplusIndex].Delta = sdkmath.ZeroInt()
			surplusIndex += 1

		} else if surplusValidator.Delta.Abs().GT(deficitValidator.Delta.Abs()) {
			// If one validator's deficit is less than the other validator's surplus,
			// move only enough of the surplus to cover the shortage
			redelegationAmount = deficitValidator.Delta.Abs()

			// Update the surplus validator, and zero out the deficit validator
			validatorDeltas[surplusIndex].Delta = surplusValidator.Delta.Sub(redelegationAmount)
			validatorDeltas[deficitIndex].Delta = sdkmath.ZeroInt()
			deficitIndex -= 1

		} else {
			// if one validator's surplus is equal to the other validator's deficit,
			// we'll transfer that amount and both validators will now be balanced
			redelegationAmount = deficitValidator.Delta.Abs()

			validatorDeltas[surplusIndex].Delta = sdkmath.ZeroInt()
			validatorDeltas[deficitIndex].Delta = sdkmath.ZeroInt()

			surplusIndex += 1
			deficitIndex -= 1
		}

		// Append the new Redelegation message and Rebalancing struct for the callback
		// We always send from the surplus validator to the deficit validator
		srcValidator := surplusValidator.ValidatorAddress
		dstValidator := deficitValidator.ValidatorAddress

		msgs = append(msgs, &stakingtypes.MsgBeginRedelegate{
			DelegatorAddress:    hostZone.DelegationIcaAddress,
			ValidatorSrcAddress: srcValidator,
			ValidatorDstAddress: dstValidator,
			Amount:              sdk.NewCoin(hostZone.HostDenom, redelegationAmount),
		})
		rebalancings = append(rebalancings, &types.Rebalancing{
			SrcValidator: srcValidator,
			DstValidator: dstValidator,
			Amt:          redelegationAmount,
		})
	}

	return msgs, rebalancings
}

// This function returns a list with the number of extra tokens that should be sent to each validator
//   - Positive delta implies the validator has a surplus (and should lose stake)
//   - Negative delta implies the validator has a deficit (and should gain stake)
func (k Keeper) GetValidatorDelegationDifferences(ctx sdk.Context, hostZone types.HostZone) ([]RebalanceValidatorDelegationChange, error) {
	// Get the target delegation amount for each validator
	totalDelegatedAmt := k.GetTotalValidatorDelegations(hostZone)
	targetDelegation, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, totalDelegatedAmt)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId)
	}

	// For each validator, store the amount that their delegation should change
	delegationDeltas := []RebalanceValidatorDelegationChange{}
	totalDelegationChange := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		// Compare the target with either the current delegation
		delegationChange := validator.Delegation.Sub(targetDelegation[validator.Address])

		// Only include validators who's delegation should change
		if !delegationChange.IsZero() {
			delegationDeltas = append(delegationDeltas, RebalanceValidatorDelegationChange{
				ValidatorAddress: validator.Address,
				Delta:            delegationChange,
			})
			totalDelegationChange = totalDelegationChange.Add(delegationChange)
		}
		k.Logger(ctx).Info(fmt.Sprintf("Adding delegation: %v to validator: %s", delegationChange, validator.Address))
	}

	// Sanity check that the sum of all the delegation change's is equal to 0
	// (meaning the total delegation across ALL validators has not changed)
	if !totalDelegationChange.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"non-zero net delegation change (%v) across validators during rebalancing", totalDelegationChange)
	}

	return delegationDeltas, nil
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
	validators := hostZone.Validators
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
func (k Keeper) GetTotalValidatorDelegations(hostZone types.HostZone) sdkmath.Int {
	validators := hostZone.Validators
	totalDelegation := sdkmath.ZeroInt()
	for _, validator := range validators {
		totalDelegation = totalDelegation.Add(validator.Delegation)
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
