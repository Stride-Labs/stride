package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
)

const redemptionRateInvariantName = "redemption-rate-above-0.9"

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, redemptionRateInvariantName, RedemptionRateInvariant(k))
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		return msg, broke
	}
}

// RedemptionRateInvariant checks that all hostZone redemption rates are above a fixed threshold
func RedemptionRateInvariant(k Keeper) sdk.Invariant {

	// threshold is 0.9
	threshold := sdk.NewDec(9).Quo(sdk.NewDec(10))

	return func(ctx sdk.Context) (string, bool) {
		for _, hz := range k.GetAllHostZone(ctx) {
			if hz.RedemptionRate.LT(threshold) {
				return sdk.FormatInvariant(types.ModuleName, redemptionRateInvariantName,
					fmt.Sprintf("[INVARIANT BROKEN!!!] %s's RR is %s", hz.GetChainId(), hz.RedemptionRate.String())), true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, redemptionRateInvariantName,
			fmt.Sprintf("All RR's are GT %s", threshold.String())), false
	}
}
