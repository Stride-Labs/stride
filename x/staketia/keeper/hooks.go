package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/Stride-Labs/stride/v23/x/epochs/types"
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
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	epochNumber := uint64(epochInfo.CurrentEpoch)

	// Every day, refresh the redemption rate and prepare delegations
	// Every 4 days, prepare undelegations
	if epochInfo.Identifier == epochstypes.DAY_EPOCH {
		fmt.Println("STAKETIA DAY EPOCH")

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
}

type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {
	h.k.BeforeEpochStart(ctx, epochInfo)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochInfo epochstypes.EpochInfo) {}
