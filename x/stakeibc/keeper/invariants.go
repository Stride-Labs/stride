package keeper

// DONTCOVER

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		return "", false
	}
}
