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
	airdrop, foundAirdrop := k.GetAirdrop(ctx, airdropId)
	if !foundAirdrop {
		return types.ErrAirdropNotFound.Wrapf("airdrop %s", airdropId)
	}

	claimerStrideAddress := claimer

	claimerStrideAccount := sdk.MustAccAddressFromBech32(claimerStrideAddress)
	distributorAccount := sdk.MustAccAddressFromBech32(airdrop.DistributionAddress)

	// Fetch the user's linked accounts
	userLinks, foundUserLinks := k.GetUserLinks(ctx, airdropId, claimer)

	claimerAddresses := []string{claimerStrideAddress}
	if foundUserLinks {
		claimerAddresses = append(userLinks.HostAddresses, claimerStrideAddress)
	}

	for _, address := range claimerAddresses {
		userAllocation, foundUserAllocation := k.GetUserAllocation(ctx, airdropId, address)
		if !foundUserAllocation {
			continue
		}

		// Confirm the user has not elected the non-daily claim types
		if userAllocation.ClaimType != types.UNSPECIFIED && userAllocation.ClaimType != types.CLAIM_DAILY {
			return types.ErrClaimTypeUnavailable.Wrapf("user has already elected claim option '%s' for address '%s' (linked to '%s')",
				userAllocation.ClaimType.String(), address, claimerStrideAddress)
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

		// If there are no rewards, continue to check the next allocation
		if todaysRewards.IsZero() {
			continue
		}

		// Update the amount claimed on the allocation record
		userAllocation.Claimed = userAllocation.Claimed.Add(todaysRewards)

		// If this is their first time claiming, flag their decision
		if userAllocation.ClaimType == types.UNSPECIFIED {
			userAllocation.ClaimType = types.CLAIM_DAILY
		}

		// Distribute rewards from the distributor
		rewardsCoin := sdk.NewCoin(airdrop.RewardDenom, todaysRewards)
		if err := k.bankKeeper.SendCoins(ctx, distributorAccount, claimerStrideAccount, sdk.NewCoins(rewardsCoin)); err != nil {
			return errorsmod.Wrapf(err, "unable to distribute rewards to '%s' (linked to '%s')", address, claimerStrideAddress)
		}

		// Update the reward record for to mark the progress
		k.SetUserAllocation(ctx, userAllocation)
	}

	return nil
}

func (k Keeper) ClaimAndStake(ctx sdk.Context, airdropId, claimer, validatorAddress string) error {
	// TODO[airdrop] implement logic

	return nil
}

func (k Keeper) ClaimEarly(ctx sdk.Context, airdropId, claimer string) error {
	// Fetch the airdrop and user's allocations
	airdrop, found := k.GetAirdrop(ctx, airdropId)
	if !found {
		return types.ErrAirdropNotFound.Wrapf("airdrop %s", airdropId)
	}

	claimerStrideAddress := claimer

	claimerStrideAccount := sdk.MustAccAddressFromBech32(claimerStrideAddress)
	distributorAccount := sdk.MustAccAddressFromBech32(airdrop.DistributionAddress)

	// Fetch the user's linked accounts
	userLinks, foundUserLinks := k.GetUserLinks(ctx, airdropId, claimer)

	claimerAddresses := []string{claimerStrideAddress}
	if foundUserLinks {
		claimerAddresses = append(userLinks.HostAddresses, claimerStrideAddress)
	}

	for _, address := range claimerAddresses {
		userAllocation, found := k.GetUserAllocation(ctx, airdropId, address)
		if !found {
			continue
		}

		// Confirm the user has not elected the non-daily claim types
		if userAllocation.ClaimType != types.UNSPECIFIED && userAllocation.ClaimType != types.CLAIM_DAILY {
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
		todaysRewards := sdkmath.ZeroInt()
		for i, rewardsOnDate := range userAllocation.Allocations {
			todaysRewards = todaysRewards.Add(rewardsOnDate)
			userAllocation.Allocations[i] = sdkmath.ZeroInt()
		}

		// If there are no rewards, continue to check the next allocation
		if todaysRewards.IsZero() {
			continue
		}

		// Update the amount claimed on the allocation record
		userAllocation.Claimed = userAllocation.Claimed.Add(todaysRewards)

		// Flag the user's claim type decision
		userAllocation.ClaimType = types.CLAIM_EARLY

		// Distribute rewards from the distributor, deducting the early penalty
		distributedRewards := sdk.NewDecFromInt(todaysRewards).Mul(airdrop.EarlyClaimPenalty).TruncateInt()
		rewardsCoin := sdk.NewCoin(airdrop.RewardDenom, distributedRewards)
		if err := k.bankKeeper.SendCoins(ctx, distributorAccount, claimerStrideAccount, sdk.NewCoins(rewardsCoin)); err != nil {
			return errorsmod.Wrapf(err, "unable to distribute rewards")
		}

		// Update the reward record for to mark the progress
		k.SetUserAllocation(ctx, userAllocation)
	}
	return nil
}
