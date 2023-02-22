package keeper

// DONTCOVER

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// epochtypes "github.com/Stride-Labs/stride/v4/x/epochs/types"
	"github.com/Stride-Labs/stride/v5/x/stakeibc/types"
)

const (
	DelegationsSumToStakedBalName = "delegations-sum-to-stakedbal"
)

// RegisterInvariants registers all governance invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, DelegationsSumToStakedBalName, k.DelegationsSumToStakedBal())
}

// AllInvariants runs all invariants of the stakeibc module
func AllInvariants(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		// msg, broke := RedemptionRateInvariant(k)(ctx)
		// note: once we have >1 invariant here, follow the pattern from staking module invariants here: https://github.com/cosmos/cosmos-sdk/blob/v0.46.0/x/staking/keeper/invariants.go
		// return "", false
		res, stop := k.DelegationsSumToStakedBal()(ctx)
		if !stop {
			return res, stop
		}
		return "", false
	}
}

// DelegationsSumToStakedBal ensure that sum of balances staked to each validator (as recorded on the host_zone struct) sums to the total staked balance of the host zone (as recorded on the host_zone struct)
// NOTE: this invariant does not query any *actual* staked balances on the host chain, only the staked balances recorded on the host_zone struct
func (k Keeper) DelegationsSumToStakedBal() sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		listHostZone := k.GetAllHostZone(ctx)

		for _, host := range listHostZone {
			sumOfDelegations := k.GetTotalValidatorDelegations(host)
			if !(host.StakedBal).Equal(sumOfDelegations) {
				return sdk.FormatInvariant(types.ModuleName, "balance-stake-hostzone-invariant", fmt.Sprintf(
					"\tStakedBal %s (as recorded on the host_zone struct) is not equal to sum of validator's delegations (as recorded on the host_zone struct) \n"+
						"\tStakedBal: %d\n"+
						"\t Sum of validator's delegations: %d\n",
					host.ChainId, host.StakedBal, sumOfDelegations,
				)), true
			}
		}
		return sdk.FormatInvariant(types.ModuleName, "delegations-sum-to-stakedbal", "sum of balances staked to each validator (as recorded on the host_zone struct) sums to the total staked balance of the host zone (as recorded on the host_zone struct)"), false
	}
}
