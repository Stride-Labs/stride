package keeper

// DONTCOVER

import (
	"fmt"

	"github.com/Stride-Labs/stride/v4/x/stakeibc/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "balance-stake-hostzone-invariant", BalanceStakeHostZoneInvariant(k))
	ir.RegisterRoute(types.ModuleName, "amount-of-delagate-of-validator-invariant", AmountDelegateOfValidatorInvariant(k))
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		// return "", false
		res, stop := BalanceStakeHostZoneInvariant(k)(ctx)
		if !stop {
			return res, stop
		}
		return AmountDelegateOfValidatorInvariant(k)(ctx)

	}
}

// BalanceStakeHostZoneInvariant ensure that balance stake of all host zone are equal to of validator's delegation
func BalanceStakeHostZoneInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		listHostZone := k.GetAllHostZone(ctx)

		for _, host := range listHostZone {
			balanceStake := host.StakedBal
			totalDelegateOfVals := k.GetTotalValidatorDelegations(host)
			if !balanceStake.Equal(totalDelegateOfVals) {
				return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant",
					fmt.Sprintf("\tBalance stake of hostzone %s is not equal to total of validator's delegations \n\tBalance stake actually: %d\n\t Total of validator's delegations: %d\n",
						host.ChainId, host.StakedBal, totalDelegateOfVals,
					)), true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant", "All host zones have balances stake is equal to total of validator's delegations"), false
	}
}

func AmountDelegateOfValidatorInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		listHostZone := k.GetAllHostZone(ctx)

		for _, host := range listHostZone {
			totalWeightOfHostZone := int64(k.GetTotalValidatorWeight(host))
			totalDelegateOfVals := k.GetTotalValidatorDelegations(host)
			for _, val := range host.Validators {
				weightOfVal := int64(val.Weight)
				amoutDelegateOfVal := val.DelegationAmt
				// TODO: check Tolerance for calculation below
				amoutDelegateOfValFromWeight := totalDelegateOfVals.Mul(sdk.NewInt(weightOfVal)).Quo(sdk.NewInt(totalWeightOfHostZone))
				if !amoutDelegateOfValFromWeight.Equal(amoutDelegateOfVal) {
					return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant",
						fmt.Sprintf("\tAmount of delegate of validator %s is not inconsistent with the ratio of weight \n\tAmount actually of delegate: %d\n\t Amount of delegate by ratio of weight: %d\n",
							val.Name, val.DelegationAmt, amoutDelegateOfValFromWeight,
						)), true
				}
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "amount-of-delagate-of-validator-invariant", "All validators have amount of delegate inconsistent with the ratio of weight"), false
	}
}
