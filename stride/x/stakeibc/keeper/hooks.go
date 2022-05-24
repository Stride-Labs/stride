package keeper

import (
	"fmt"

	epochstypes "github.com/Stride-Labs/stride/x/epochs/types"
	sdk "github.com/Stride-Labs/cosmos-sdk/types"
)

func (k Keeper) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
	k.Logger(ctx).Info(fmt.Sprintf("Handling epoch start %s", epochIdentifier))

	if epochIdentifier == "day" {
		k.Logger(ctx).Info("Starting day %d", epochNumber)
	}
	if epochIdentifier == "week" {
		k.Logger(ctx).Info("Starting week %d", epochNumber)
	}
	// k.Logger(ctx).Info(fmt.Sprintf("Handling epoch START TEST %d, %s", uint64(epochNumber), epochIdentifier))
	// if epochIdentifier == "epoch" {
	// 	k.Logger(ctx).Info(fmt.Sprintf("STARTED epoch %d VISH %s", epochNumber, epochIdentifier))
	// }
}

func (k Keeper) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	// every epoch
	k.Logger(ctx).Info("Handling epoch end")
	if epochIdentifier == "day" {
		k.Logger(ctx).Info("Finished day %d", epochNumber)
	}
	if epochIdentifier == "week" {
		k.Logger(ctx).Info("Finished week %d", epochNumber)
	}
}

// Hooks wrapper struct for incentives keeper
type Hooks struct {
	k Keeper
}

var _ epochstypes.EpochHooks = Hooks{}

func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// epochs hooks
func (h Hooks) BeforeEpochStart(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.BeforeEpochStart(ctx, epochIdentifier, epochNumber)
}

func (h Hooks) AfterEpochEnd(ctx sdk.Context, epochIdentifier string, epochNumber int64) {
	h.k.AfterEpochEnd(ctx, epochIdentifier, epochNumber)
}
