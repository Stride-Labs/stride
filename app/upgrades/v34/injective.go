package v34

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

const InjectiveChainId = "cosmoshub-test-1" // REHEARSAL ONLY — DO NOT MERGE (mainnet: "injective-1")

// DelegationDelta is the difference between the delegation ICA's actual on-chain
// delegation to a validator and the delegation tracked in the injective-1 host zone.
type DelegationDelta struct {
	Name    string
	Address string
	Delta   sdkmath.Int // actual on-chain delegation minus tracked delegation, in inj (18 decimals)
}

// InjectiveDelegationDeltas trues up the injective-1 host zone's tracked delegations to
// what is actually staked on Injective.
//
// Background: the injective-1 DELEGATION ICA channel died seven times between mid-July and
// early September 2026 (no relayer covered the icacontroller path, so each epoch's ~4h ICA
// timeouts closed the ordered channel). Every death stranded the acks of packets that had
// already been executed on Injective, so their success callbacks never ran and the host zone
// never booked them. The cumulative effect is that actual delegations exceed tracked
// delegations by ~200.48 INJ (one whole redelegation plus a few dozen reinvest delegations),
// and that ~200 INJ of liquid balance the redemption sweep expects to find in the delegation
// ICA is staked instead. The sweep (all-or-nothing) has been failing with "insufficient
// funds" every epoch since, blocking 12 user redemptions back to epoch 1369.
//
// Measured 2026-09-13 at Injective height 182880882 / Stride height 40258613 by diffing
// x/staking delegations of inj16ujqtje2en9ns59hrcjfu2885epsrvus9czdw8ljtd45whxucs5srrejpa
// against the host zone validators. The total delta was byte-identical to a 2026-09-08
// measurement, i.e. it is not drifting now that acks are relayed. Re-measure before the
// upgrade proposal anyway; only entries with |delta| > 0.001 INJ are included.
// REHEARSAL ONLY — DO NOT MERGE: the mainnet table is replaced by the k8s Gaia validators with
// deltas measured by `rehearsal/injective.sh measure` (on-chain delegation of the delegation ICA
// minus tracked delegation), the same method used for the mainnet measurement.
// Measured 2026-09-15 22:53 UTC on the k8s rehearsal network by `rehearsal/injective.sh measure`
// (Σ = 309643420 uatom: the 300 ATOM lost-ack delegate plus reinvests whose acks were withheld).
var InjectiveDelegationDeltas = []DelegationDelta{
	{Name: "val3", Address: "cosmosvaloper1nnurja9zt97huqvsfuartetyjx63tc5zxcyn3n", Delta: mustInt("100253608")},
	{Name: "cosmoshub1", Address: "cosmosvaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrdt795p", Delta: mustInt("100253609")},
	{Name: "val2", Address: "cosmosvaloper17kht2x2ped6qytr2kklevtvmxpw7wq9rarvcqz", Delta: mustInt("100253608")},
}

func mustInt(s string) sdkmath.Int {
	i, ok := sdkmath.NewIntFromString(s)
	if !ok {
		panic("v34: invalid integer constant " + s)
	}
	return i
}

// ReconcileInjectiveDelegations applies InjectiveDelegationDeltas to the injective-1 host zone so
// that tracked validator delegations (and TotalDelegations) match what is staked on Injective.
//
// It returns the total delta actually applied (the same amount TotalDelegations was adjusted by).
// The upgrade handler queues that excess as a pending undelegation, which the day-epoch hook
// submits through the normal undelegate pipeline to move the accidentally staked redemption funds
// back to liquid (see x/stakeibc/keeper/pending_undelegation.go).
//
// The table is applied all-or-nothing, and never as an upgrade error. A missing host zone means a
// non-mainnet environment. A validator missing from the host zone, or a delta that would drive a
// delegation negative, means the constants no longer describe chain state (the validator set
// changed, or acks were lost again): applying the remaining entries would return a partial sum
// (dropping one large negative entry turns the ~200 INJ excess into ~867 INJ) and that sum would
// then be undelegated. Instead nothing is written and zero is returned, so the upgrade completes
// with the sweep unblocked and the reconciliation deferred to a later upgrade. An error here
// would halt the chain, which is disproportionate for an accounting fix.
func ReconcileInjectiveDelegations(ctx sdk.Context, sk stakeibckeeper.Keeper) (appliedDelta sdkmath.Int) {
	hostZone, found := sk.GetHostZone(ctx, InjectiveChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v34: host zone %s not found, skipping delegation reconciliation", InjectiveChainId))
		return sdkmath.ZeroInt()
	}

	// The host zone is only persisted after every entry has been validated
	totalDelta := sdkmath.ZeroInt()
	for _, entry := range InjectiveDelegationDeltas {
		validator, index, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		if !found {
			ctx.Logger().Error(fmt.Sprintf("v34: validator %s (%s) not found on %s; delegation reconciliation NOT applied, "+
				"re-verify constants and reconcile in a later upgrade", entry.Name, entry.Address, InjectiveChainId))
			return sdkmath.ZeroInt()
		}

		if validator.Delegation.IsNil() {
			validator.Delegation = sdkmath.ZeroInt()
		}
		reconciled := validator.Delegation.Add(entry.Delta)
		if reconciled.IsNegative() {
			ctx.Logger().Error(fmt.Sprintf("v34: validator %s tracked delegation %v plus delta %v would be negative; delegation "+
				"reconciliation NOT applied, re-verify constants and reconcile in a later upgrade",
				entry.Name, validator.Delegation, entry.Delta))
			return sdkmath.ZeroInt()
		}

		ctx.Logger().Info(fmt.Sprintf("v34: reconciled %s delegation %v -> %v (%v)",
			entry.Name, validator.Delegation, reconciled, entry.Delta))
		validator.Delegation = reconciled
		hostZone.Validators[index] = &validator
		totalDelta = totalDelta.Add(entry.Delta)
	}

	hostZone.TotalDelegations = hostZone.TotalDelegations.Add(totalDelta)
	sk.SetHostZone(ctx, hostZone)

	ctx.Logger().Info(fmt.Sprintf("v34: %s TotalDelegations adjusted by %v to %v",
		InjectiveChainId, totalDelta, hostZone.TotalDelegations))
	return totalDelta
}
