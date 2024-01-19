package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	epochstypes "github.com/Stride-Labs/stride/v17/x/epochs/types"
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
	// TODO [sttia]: Add epochly business logic
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
