package keeper

// DONTCOVER

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "balance-stake-hostzone-invariant", k.BalanceStakeHostZoneInvariant())
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		// return "", false
		return k.BalanceStakeHostZoneInvariant()(ctx)

	}
}

// BalanceStakeHostZoneInvariant ensure that balance stake of all host zone are equal to of validator's delegation
func (k Keeper) BalanceStakeHostZoneInvariant() sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		listHostZone := k.GetAllHostZone(ctx)

		for _, host := range listHostZone {
			balanceStake := host.StakedBal
			totalDelegateOfVals := k.GetTotalValidatorDelegations(host)
			if !balanceStake.Equal(totalDelegateOfVals) {
				return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant", fmt.Sprintf(
					"\tBalance stake of hostzone %s is not equal to total of validator's delegations \n"+
						"\tBalance stake actually: %d\n"+
						"\t Total of validator's delegations: %d\n",
					host.ChainId, host.StakedBal, totalDelegateOfVals,
				)), true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant", "All host zones have balances stake is equal to total of validator's delegations"), false
	}
}
