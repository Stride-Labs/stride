package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v26/utils"
	epochstypes "github.com/Stride-Labs/stride/v26/x/epochs/types"
)

// This module has the following epochly triggers
//   - Handle delegations daily
//   - Handle undelegations every 4 days
//   - Updates the redemption rate daily
//   - Check for completed unbondings hourly
//   - Process claims (if applicable) hourly
//
// Note: The hourly processes are meant for actions that should run ASAP,
// but the hourly buffer makes it less expensive
func (k Keeper) BeforeEpochStart(context context.Context, epochInfo epochstypes.EpochInfo) {
	ctx := sdk.UnwrapSDKContext(context)

	epochNumber := utils.IntToUint(epochInfo.CurrentEpoch)

	// Every day, refresh the redemption rate and prepare delegations
	// Every 4 days, prepare undelegations
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		// Update the redemption rate
		// If this fails, do not proceed to the delegation or undelegation step
		// Note: This must be run first because it is used when refreshing the native token
		// balance in prepare undelegation
		if err := k.UpdateRedemptionRate(ctx); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable update redemption rate: %s", err.Error()))
			return
		}

		// Post the redemption rate to the oracle (if it doesn't exceed the bounds)
		if err := k.PostRedemptionRateToOracles(ctx); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to post redemption rate to oracle: %s", err.Error()))
		}

		// Prepare delegations by transferring the deposited tokens to the host zone
		if err := k.SafelyPrepareDelegation(ctx, epochNumber, epochInfo.Duration); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to prepare delegation for epoch %d: %s", epochNumber, err.Error()))
		}

		// Every few days (depending on the unbonding frequency) prepare undelegations which
		// freezes the accumulating unbonding record and refreshes the native token amount
		// TODO [cleanup]: replace with unbonding frequency
		if epochInfo.CurrentEpoch%4 == 0 {
			if err := k.SafelyPrepareUndelegation(ctx, epochNumber); err != nil {
				k.Logger(ctx).Error(fmt.Sprintf("Unable to prepare undelegations for epoch %d: %s", epochNumber, err.Error()))
			}
		}
	}

	// Every hour, annotate finished unbondings and distribute claims
	// The hourly epoch is meant for actions that should be executed asap, but have a
	// relaxed SLA. It makes it slightly less expensive than running every block
	if epochInfo.Identifier == epochstypes.HOUR_EPOCH {
		k.MarkFinishedUnbondings(ctx)

		if err := k.SafelyDistributeClaims(ctx); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to distribute claims for epoch %d: %s", epochNumber, err.Error()))
		}
	}

	// Every mint epoch, liquid stake fees and distribute to fee collector
	if epochInfo.Identifier == epochstypes.MINT_EPOCH {
		if err := k.SafelyLiquidStakeAndDistributeFees(ctx); err != nil {
			k.Logger(ctx).Error(fmt.Sprintf("Unable to liquid stake and distribute fees this epoch %d: %s", epochNumber, err.Error()))
		}
	}
}

type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(context context.Context, epochInfo epochstypes.EpochInfo) {
	ctx := sdk.UnwrapSDKContext(context)
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(context context.Context, epochInfo epochstypes.EpochInfo) {}
