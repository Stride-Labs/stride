package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "balance-stake-hostzone-invariant", k.BalanceStakeHostZoneInvariant())
	ir.RegisterRoute(types.ModuleName, "validator-weight-hostzone-invariant", k.ValidatorWeightHostZoneInvariant())
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		// return "", false
		res, stop := k.ValidatorWeightHostZoneInvariant()(ctx)
		if !stop {
			return res, stop
		}
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

// ValidatorWeightHostZoneInvariant ensure that sum of all validator delegation weight is equal to balance stake
func (k Keeper) ValidatorWeightHostZoneInvariant() sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		listHostZone := k.GetAllHostZone(ctx)

		for _, host := range listHostZone {
			totalDelegateOfVals := k.GetTotalValidatorDelegations(host)
			totalWeightOfHostZone := k.GetTotalValidatorWeight(host)
			totalAllocated := sdk.ZeroInt()
			validators := host.Validators
			for _, validator := range validators {
				delegateAmt := validator.DelegationAmt
				delegateAmtFromWeight := sdk.NewIntFromUint64(validator.Weight).Mul(totalDelegateOfVals).Quo(sdk.NewIntFromUint64(totalWeightOfHostZone))
				if !delegateAmtFromWeight.Equal(delegateAmt) {
					return sdk.FormatInvariant(types.ModuleName, "validator-weight-hostzone-invariant", fmt.Sprintf(
						"\tAmount of delegate of validator %s is not inconsistent with the ratio of weight \n"+
							"\tAmount actually of delegate: %d\n"+
							"\tAmount of delegate by ratio of weight: %d\n",
						validator.Name, delegateAmt, delegateAmtFromWeight,
					)), true
				}
				totalAllocated = totalAllocated.Add(delegateAmt)
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "validator-weight-hostzone-invariant", "All host zones have sum of all validator delegation weight is equal to balance stake of all host zone"), false
	}
}
