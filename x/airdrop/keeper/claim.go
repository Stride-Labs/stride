package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v22/x/airdrop/types"
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

	// Confirm the user has not elected the non-daily claim types
	if userAllocation.ClaimType != types.CLAIM_DAILY {
		return types.ErrClaimTypeUnavailable.Wrapf("user has already elected claim option %s",
			userAllocation.ClaimType.String())
	}

	// Get the index in the allocations array from the current date
	// E.g. on the 10th day of distribution, this will map to the 9th index in the list
	todaysIndex, err := airdrop.GetCurrentDateIndex(ctx)
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

	// If this is their first time claiming, flag their decision
	userAllocation.ClaimType = types.CLAIM_DAILY

	// Distribute rewards from the distributor
	distributorAccount := sdk.MustAccAddressFromBech32(airdrop.DistributionAddress)
	claimerAccount := sdk.MustAccAddressFromBech32(userAllocation.Address)
	rewardsCoin := sdk.NewCoin(airdrop.RewardDenom, todaysRewards)

	if err := k.bankKeeper.SendCoins(ctx, distributorAccount, claimerAccount, sdk.NewCoins(rewardsCoin)); err != nil {
		return errorsmod.Wrapf(err, "unable to distribute rewards")
	}

	// Update the reward record for to mark the progress
	k.SetUserAllocation(ctx, userAllocation)

	return nil
}

func (k Keeper) ClaimAndStake(ctx sdk.Context, airdropId, claimer, validatorAddress string) error {
	// TODO[airdrop] implement logic

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

	// Confirm the user has not elected the non-daily claim types
	if userAllocation.ClaimType != types.CLAIM_DAILY {
		return types.ErrClaimTypeUnavailable.Wrapf("user has already elected claim option %s",
			userAllocation.ClaimType.String())
	}

	// Confirm we're not past the decision date
	currentTime := ctx.BlockTime().Unix()
	if currentTime > airdrop.ClaimTypeDeadlineDate.Unix() {
		return types.ErrAfterDecisionDeadline
	}

	// Confirm the airdrop started
	if currentTime < airdrop.DistributionStartDate.Unix() {
		return types.ErrDistributionNotStarted
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
	rewardsRemainingRate := sdk.OneDec().Sub(airdrop.EarlyClaimPenalty)
	distributedRewards := sdk.NewDecFromInt(totalAccruedRewards).Mul(rewardsRemainingRate).TruncateInt()

	// Update the amount claimed on the allocation record
	// claimed += distributedRewards
	userAllocation.Claimed = userAllocation.Claimed.Add(distributedRewards)

	// Update the amount forfeited on the allocation record
	// forfeited += totalAccruedRewards - distributedRewards
	// Note: forfeited should be zero before the next operation,
	// but we're doing += just in case there's a scenario where it's not zero in the future
	userAllocation.Forfeited = userAllocation.Forfeited.Add(totalAccruedRewards.Sub(distributedRewards))

	// Flag the user's claim type decision
	userAllocation.ClaimType = types.CLAIM_EARLY

	// Distribute rewards from the distributor, deducting the early penalty
	distributorAccount := sdk.MustAccAddressFromBech32(airdrop.DistributionAddress)
	claimerAccount := sdk.MustAccAddressFromBech32(userAllocation.Address)

	rewardsCoin := sdk.NewCoin(airdrop.RewardDenom, distributedRewards)
	if err := k.bankKeeper.SendCoins(ctx, distributorAccount, claimerAccount, sdk.NewCoins(rewardsCoin)); err != nil {
		return errorsmod.Wrapf(err, "unable to distribute rewards")
	}

	// Update the reward record for to mark the progress
	k.SetUserAllocation(ctx, userAllocation)

	return nil
}
