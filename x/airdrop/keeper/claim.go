package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/x/airdrop/types"
)

// User transaction to claim all the pending airdrop rewards up to the current day
func (k Keeper) ClaimDaily(ctx sdk.Context, airdropId, claimer string) error {
	// Fetch the airdrop and user's allocations
	airdrop, airdropFound := k.GetAirdrop(ctx, airdropId)
	if !airdropFound {
		return types.ErrAirdropNotFound.Wrapf("airdrop %s", airdropId)
	}
	userAllocation, userFound := k.GetUserAllocation(ctx, airdropId, claimer)
	if !userFound {
		return types.ErrUserAllocationNotFound.Wrapf("user %s for airdrop %s", claimer, airdropId)
	}

	// Confirm the airdrop started
	currentTime := ctx.BlockTime().Unix()
	if currentTime < airdrop.DistributionStartDate.Unix() {
		return types.ErrAirdropNotStarted
	}

	// Confirm we're not passed the clawback date
	if currentTime >= airdrop.ClawbackDate.Unix() {
		return types.ErrAirdropEnded
	}

	// Get the index in the allocations array from the current date
	// E.g. on the 10th day of distribution, this will map to the 9th index in the list
	periodLengthSeconds := k.GetParams(ctx).PeriodLengthSeconds
	todaysIndex, err := airdrop.GetCurrentDateIndex(ctx, periodLengthSeconds)
	if err != nil {
		return err
	}

	// Sum the rewards up to that date and 0 them out in the process
	todaysRewards := sdkmath.ZeroInt()
	for i := 0; i <= todaysIndex; i++ {
		rewardsOnDate := userAllocation.Allocations[i]
		todaysRewards = todaysRewards.Add(rewardsOnDate)
		userAllocation.Allocations[i] = sdkmath.ZeroInt()
	}

	// If there are no rewards, alert the user with an error
	if todaysRewards.IsZero() {
		return types.ErrNoUnclaimedRewards
	}

	// Update the amount claimed on the allocation record
	userAllocation.Claimed = userAllocation.Claimed.Add(todaysRewards)

	// Distribute rewards from the distributor
	distributorAccount := sdk.MustAccAddressFromBech32(airdrop.DistributorAddress)
	claimerAccount := sdk.MustAccAddressFromBech32(userAllocation.Address)
	rewardsCoin := sdk.NewCoin(airdrop.RewardDenom, todaysRewards)

	// Update the reward record for to mark the progress
	k.SetUserAllocation(ctx, userAllocation)

	if err := k.bankKeeper.SendCoins(ctx, distributorAccount, claimerAccount, sdk.NewCoins(rewardsCoin)); err != nil {
		return errorsmod.Wrapf(err, "unable to distribute rewards")
	}

	return nil
}

// User transaction to claim a portion of their total amount now, and forfeit the
// remainder to be clawed back
func (k Keeper) ClaimEarly(ctx sdk.Context, airdropId, claimer string) error {
	// Fetch the airdrop and user's allocations
	airdrop, airdropFound := k.GetAirdrop(ctx, airdropId)
	if !airdropFound {
		return types.ErrAirdropNotFound.Wrapf("airdrop %s", airdropId)
	}
	userAllocation, userFound := k.GetUserAllocation(ctx, airdropId, claimer)
	if !userFound {
		return types.ErrUserAllocationNotFound.Wrapf("user %s for airdrop %s", claimer, airdropId)
	}

	// Confirm the airdrop started
	currentTime := ctx.BlockTime().Unix()
	if currentTime < airdrop.DistributionStartDate.Unix() {
		return types.ErrAirdropNotStarted
	}

	// Confirm we're not past the decision date
	if currentTime > airdrop.ClaimTypeDeadlineDate.Unix() {
		return types.ErrAfterDecisionDeadline
	}

	// Sum the total rewards 0 them out in the process
	totalAccruedRewards := sdkmath.ZeroInt()
	for i, rewardsOnDate := range userAllocation.Allocations {
		totalAccruedRewards = totalAccruedRewards.Add(rewardsOnDate)
		userAllocation.Allocations[i] = sdkmath.ZeroInt()
	}

	// If there are no rewards, alert the user with an error
	if totalAccruedRewards.IsZero() {
		return types.ErrNoUnclaimedRewards
	}

	// Calculate rewards after claim early penalty
	rewardsRemainingRate := sdkmath.LegacyOneDec().Sub(airdrop.EarlyClaimPenalty)
	distributedRewards := sdk.NewDecFromInt(totalAccruedRewards).Mul(rewardsRemainingRate).TruncateInt()

	// Update the amount claimed on the allocation record by the amount distributed
	userAllocation.Claimed = userAllocation.Claimed.Add(distributedRewards)

	// Update the amount forfeited on the allocation record by (total rewards amount - distributed rewards)
	// Note: forfeited should be zero before the next operation,
	// but we're doing += just in case there's a scenario where it's not zero in the future
	forfeitedRewards := totalAccruedRewards.Sub(distributedRewards)
	userAllocation.Forfeited = userAllocation.Forfeited.Add(forfeitedRewards)

	// Distribute rewards from the distributor, deducting the early penalty
	distributorAccount := sdk.MustAccAddressFromBech32(airdrop.DistributorAddress)
	claimerAccount := sdk.MustAccAddressFromBech32(userAllocation.Address)
	rewardsCoin := sdk.NewCoin(airdrop.RewardDenom, distributedRewards)

	// Update the reward record for to mark the progress
	k.SetUserAllocation(ctx, userAllocation)

	if err := k.bankKeeper.SendCoins(ctx, distributorAccount, claimerAccount, sdk.NewCoins(rewardsCoin)); err != nil {
		return errorsmod.Wrapf(err, "unable to distribute rewards")
	}

	return nil
}

// Admin transaction to merge allocations between a stride and non-stride address
// If the stride address does not yet have an allocation, the host allocation will be overwritten
// with the stride address
// Otherwise, if the stride allocation already exists, the two allocations will be merged and set
// in on the stride allocation
// There's no need to merge the Claimed or Forfeited amounts because the host allocations cannot
// be claimed through a non-stride address
func (k Keeper) LinkAddresses(ctx sdk.Context, airdropId, strideAddress, hostAddress string) error {
	// Fetch the user's allocation across both addresses
	hostAllocation, hostFound := k.GetUserAllocation(ctx, airdropId, hostAddress)
	if !hostFound {
		return types.ErrUserAllocationNotFound.Wrapf("user %s for airdrop %s", hostAddress, airdropId)
	}
	strideAllocation, strideFound := k.GetUserAllocation(ctx, airdropId, strideAddress)

	// If the stride user doesn't exist yet, just update the address in the host allocation
	// to the the stride address overwrite it
	if !strideFound {
		hostAllocation.Address = strideAddress
		k.RemoveUserAllocation(ctx, airdropId, hostAddress)
		k.SetUserAllocation(ctx, hostAllocation)
		return nil
	}

	// Confirm the stride and host allocations are the same length
	if len(strideAllocation.Allocations) != len(hostAllocation.Allocations) {
		return errorsmod.Wrapf(types.ErrFailedToLinkAddresses,
			"stride (%s) and host (%s) allocations are not the same length", strideAddress, hostAllocation.Address)
	}

	// If the stride user does exist, merge the two allocations into the stride user
	for i, strideRewards := range strideAllocation.Allocations {
		hostReward := hostAllocation.Allocations[i]
		strideAllocation.Allocations[i] = strideRewards.Add(hostReward)
	}

	// Use the stride allocation as the canonical one moving forward and remove the host allocation
	k.SetUserAllocation(ctx, strideAllocation)
	k.RemoveUserAllocation(ctx, airdropId, hostAddress)

	return nil
}
