package keeper

import (
	"context"

	"github.com/Stride-Labs/stride/v30/x/epochs/types"
)

// AfterEpochEnd executes the indicated hook after epochs ends
func (k Keeper) AfterEpochEnd(context context.Context, epochInfo types.EpochInfo) {
	k.hooks.AfterEpochEnd(context, epochInfo)
}

// BeforeEpochStart executes the indicated hook before the epochs
func (k Keeper) BeforeEpochStart(context context.Context, epochInfo types.EpochInfo) {
	k.hooks.BeforeEpochStart(context, epochInfo)
}
