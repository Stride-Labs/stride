package keeper

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/Stride-Labs/stride/v14/utils"
	epochstypes "github.com/Stride-Labs/stride/v14/x/epochs/types"
	"github.com/Stride-Labs/stride/v14/x/stakeibc/types"
)

const RebalanceIcaBatchSize = 5

type RebalanceValidatorDelegationChange struct {
	ValidatorAddress string
	Delta            sdkmath.Int
}

// Iterate each active host zone and issues redelegation messages to rebalance each
// validator's stake according to their weights
//
// This is required when accepting LSM LiquidStakes as the distribution of stake
// from the LSM Tokens will be inconsistend with the host zone's validator set
//
// Note: this cannot be run more than once in a single unbonding period
func (k Keeper) RebalanceAllHostZones(ctx sdk.Context) {
	dayEpoch, found := k.GetEpochTracker(ctx, epochstypes.DAY_EPOCH)
	if !found {
		k.Logger(ctx).Error("Unable to get day epoch tracker")
		return
	}

	for _, hostZone := range k.GetAllActiveHostZone(ctx) {
		if dayEpoch.EpochNumber%hostZone.UnbondingPeriod != 0 {
			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
				"Host does not rebalance this epoch (Unbonding Period: %d, Epoch: %d)", hostZone.UnbondingPeriod, dayEpoch.EpochNumber))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Rebalancing delegations"))

		if err := k.RebalanceDelegationsForHostZone(ctx, hostZone.ChainId); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to rebalance delegations for %s: %s", hostZone.ChainId, err.Error()))
			continue
		}
		k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Successfully rebalanced delegations"))
	}
}

// Rebalance validators according to their validator weights for a specific host zone
func (k Keeper) RebalanceDelegationsForHostZone(ctx sdk.Context, chainId string) error {
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

	msgs, rebalancings := k.GetRebalanceICAMessages(hostZone, valDeltaList)

	for start := 0; start < len(msgs); start += RebalanceIcaBatchSize {
		end := start + RebalanceIcaBatchSize
		if end > len(msgs) {
			end = len(msgs)
		}

		msgsBatch := msgs[start:end]
		rebalancingsBatch := rebalancings[start:end]

		// marshall the callback
		rebalanceCallback := types.RebalanceCallback{
			HostZoneId:   hostZone.ChainId,
			Rebalancings: rebalancingsBatch,
		}
		rebalanceCallbackBz, err := k.MarshalRebalanceCallbackArgs(ctx, rebalanceCallback)
		if err != nil {
			return errorsmod.Wrapf(err, "unable to marshal rebalance callback args")
		}

		// Submit the rebalance ICA
		_, err = k.SubmitTxsStrideEpoch(
			ctx,
			hostZone.ConnectionId,
			msgsBatch,
			types.ICAAccountType_DELEGATION,
			ICACallbackID_Rebalance,
			rebalanceCallbackBz,
		)
		if err != nil {
			return errorsmod.Wrapf(err, "Failed to SubmitTxs for %s, messages: %+v", hostZone.ChainId, msgs)
		}

		// flag the delegation change in progress on each validator
		for _, rebalancing := range rebalancingsBatch {
			if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, rebalancing.SrcValidator); err != nil {
				return err
			}
			if err := k.IncrementValidatorDelegationChangesInProgress(&hostZone, rebalancing.DstValidator); err != nil {
				return err
			}
		}
		k.SetHostZone(ctx, hostZone)
	}

	return nil
}

// Given a list of target delegation changes, builds the individual re-delegation messages by redelegating
// from surplus validators to deficit validators
// Returns the list of messages and the callback data for the ICA
func (k Keeper) GetRebalanceICAMessages(
	hostZone types.HostZone,
	validatorDeltas []RebalanceValidatorDelegationChange,
) (msgs []proto.Message, rebalancings []*types.Rebalancing) {
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
	for surplusIndex <= deficitIndex {
		// surplus validator delta is positive, deficit validator delta is negative
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
	// The total rebalance amount consists of all delegations from validator's without a slash query in progress
	// Validators with a slash query in progress will be excluded from rebalancing
	targetRebalanceAmount := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		if !validator.SlashQueryInProgress {
			targetRebalanceAmount = targetRebalanceAmount.Add(validator.Delegation)
		}
	}

	// Get the target delegation amount for each validator
	targetDelegations, err := k.GetTargetValAmtsForHostZone(ctx, hostZone, targetRebalanceAmount)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "unable to get target val amounts for host zone %s", hostZone.ChainId)
	}

	// For each validator, store the amount that their delegation should change
	delegationDeltas := []RebalanceValidatorDelegationChange{}
	totalDelegationChange := sdkmath.ZeroInt()
	for _, validator := range hostZone.Validators {
		// Compare the target with the current delegation
		targetDelegation, ok := targetDelegations[validator.Address]
		if !ok {
			continue
		}
		delegationChange := validator.Delegation.Sub(targetDelegation)

		// Only include validators who's delegation should change
		if !delegationChange.IsZero() {
			delegationDeltas = append(delegationDeltas, RebalanceValidatorDelegationChange{
				ValidatorAddress: validator.Address,
				Delta:            delegationChange,
			})
			totalDelegationChange = totalDelegationChange.Add(delegationChange)

			k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId,
				"Validator %s delegation surplus/deficit: %v", validator.Address, delegationChange))
		}
	}

	// Sanity check that the sum of all the delegation change's is equal to 0
	// (meaning the total delegation across ALL validators has not changed)
	if !totalDelegationChange.IsZero() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"non-zero net delegation change (%v) across validators during rebalancing", totalDelegationChange)
	}

	return delegationDeltas, nil
}

// This will split a total delegation amount across validators, according to weights
// It returns a map of each portion, key'd on validator address
// Validator's with a slash query in progress are excluded
func (k Keeper) GetTargetValAmtsForHostZone(ctx sdk.Context, hostZone types.HostZone, totalDelegation sdkmath.Int) (map[string]sdkmath.Int, error) {
	// Confirm the expected delegation amount is greater than 0
	if !totalDelegation.IsPositive() {
		return nil, errorsmod.Wrapf(sdkerrors.ErrInvalidRequest,
			"Cannot calculate target delegation if final amount is less than or equal to zero (%v)", totalDelegation)
	}

	// Ignore any validators with a slash query in progress
	validators := []types.Validator{}
	for _, validator := range hostZone.Validators {
		if !validator.SlashQueryInProgress {
			validators = append(validators, *validator)
		}
	}

	// Sum the total weight across all validators
	totalWeight := k.GetTotalValidatorWeight(validators)
	if totalWeight == 0 {
		return nil, errorsmod.Wrapf(types.ErrNoValidatorWeights,
			"No non-zero validators found for host zone %s", hostZone.ChainId)
	}
	k.Logger(ctx).Info(utils.LogWithHostZone(hostZone.ChainId, "Total Validator Weight: %d", totalWeight))

	// sort validators by weight ascending
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
			targetUnbondingsByValidator[validator.Address] = totalDelegation.Sub(totalAllocated)
		} else {
			delegateAmt := sdkmath.NewIntFromUint64(validator.Weight).Mul(totalDelegation).Quo(sdkmath.NewIntFromUint64(totalWeight))
			totalAllocated = totalAllocated.Add(delegateAmt)
			targetUnbondingsByValidator[validator.Address] = delegateAmt
		}
	}

	return targetUnbondingsByValidator, nil
}

// Sum the total weights across each validator for a host zone
func (k Keeper) GetTotalValidatorWeight(validators []types.Validator) uint64 {
	totalWeight := uint64(0)
	for _, validator := range validators {
		totalWeight += validator.Weight
	}
	return totalWeight
}
