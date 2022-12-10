package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/epochs/types"
)

// AfterEpochEnd executes the indicated hook after epochs ends
func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochInfo types.EpochInfo) {
	k.hooks.AfterEpochEnd(ctx, epochInfo)
}

// BeforeEpochStart executes the indicated hook before the epochs
func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochInfo types.EpochInfo) {
	k.hooks.BeforeEpochStart(ctx, epochInfo)
}
