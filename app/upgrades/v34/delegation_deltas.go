package v34

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// Chain-agnostic helpers for truing up a host zone's tracked validator delegations to what is
// actually staked on the host. Each chain's delta table and entry point live in its own file
// (injective.go, celestia.go).

// DelegationDelta is the difference between the delegation ICA's actual on-chain
// delegation to a validator and the delegation tracked in a host zone. Deltas are in the
// host's base denom.
type DelegationDelta struct {
	Name    string
	Address string
	Delta   sdkmath.Int // actual on-chain delegation minus tracked delegation
}

func mustInt(s string) sdkmath.Int {
	i, ok := sdkmath.NewIntFromString(s)
	if !ok {
		panic("v34: invalid integer constant " + s)
	}
	return i
}

// reconcileHostZoneDelegations applies a delta table to a host zone's tracked validator
// delegations and raises TotalDelegations by the table's sum. It is shared by the Injective and
// Celestia reconciliations.
//
// The table is applied all-or-nothing, and never as an upgrade error. A missing host zone means a
// non-mainnet environment. A validator missing from the host zone, or a delta that would drive a
// delegation negative, means the constants no longer describe chain state (the validator set
// changed, or acks were lost again). In every failure case nothing is written, zero is returned
// and applied is false, so the reconciliation is deferred to a later upgrade. An error here would
// halt the chain, which is disproportionate for an accounting fix.
func reconcileHostZoneDelegations(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
	chainId string,
	deltas []DelegationDelta,
) (appliedDelta sdkmath.Int, applied bool) {
	hostZone, found := sk.GetHostZone(ctx, chainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v34: host zone %s not found, skipping delegation reconciliation", chainId))
		return sdkmath.ZeroInt(), false
	}

	// The host zone is only persisted after every entry has been validated
	totalDelta, valid := validateDelegationDeltas(ctx, hostZone, deltas)
	if !valid {
		return sdkmath.ZeroInt(), false
	}

	for _, entry := range deltas {
		validator, index, _ := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		if validator.Delegation.IsNil() {
			validator.Delegation = sdkmath.ZeroInt()
		}
		reconciled := validator.Delegation.Add(entry.Delta)

		ctx.Logger().Info(fmt.Sprintf("v34: reconciled %s delegation %v -> %v (%v)",
			entry.Name, validator.Delegation, reconciled, entry.Delta))
		validator.Delegation = reconciled
		hostZone.Validators[index] = &validator
	}

	hostZone.TotalDelegations = hostZone.TotalDelegations.Add(totalDelta)
	sk.SetHostZone(ctx, hostZone)

	ctx.Logger().Info(fmt.Sprintf("v34: %s TotalDelegations adjusted by %v to %v",
		chainId, totalDelta, hostZone.TotalDelegations))
	return totalDelta, true
}

// validateDelegationDeltas checks that every entry of a delta table can be applied to the host
// zone (validator present, delegation not driven negative) without writing anything. It returns
// the table's sum and whether the whole table is applicable; failures are logged as errors.
func validateDelegationDeltas(ctx sdk.Context, hostZone stakeibctypes.HostZone, deltas []DelegationDelta) (totalDelta sdkmath.Int, valid bool) {
	totalDelta = sdkmath.ZeroInt()
	for _, entry := range deltas {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("v34: validator %s (%s) not found on %s; delegation reconciliation NOT applied, "+
				"re-verify constants and reconcile in a later upgrade", entry.Name, entry.Address, hostZone.ChainId))
			return sdkmath.ZeroInt(), false
		}

		tracked := validator.Delegation
		if tracked.IsNil() {
			tracked = sdkmath.ZeroInt()
		}
		if tracked.Add(entry.Delta).IsNegative() {
			ctx.Logger().Error(fmt.Sprintf("v34: validator %s tracked delegation %v plus delta %v would be negative; delegation "+
				"reconciliation NOT applied, re-verify constants and reconcile in a later upgrade",
				entry.Name, tracked, entry.Delta))
			return sdkmath.ZeroInt(), false
		}
		totalDelta = totalDelta.Add(entry.Delta)
	}
	return totalDelta, true
}
