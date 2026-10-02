package v35

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
)

// HaqqDelegationDeltas trues up the haqq_11235-1 host zone's tracked delegations to what is
// actually staked on Haqq (spec §5 "Haqq delegation reconciliation", §9a).
//
// Twelve validators are over-recorded (undetected downtime slashes and sub-token rounding,
// the largest 853.8 ISLM) and four are under-recorded by sub-token dust. Both signs are
// applied so every tracked delegation equals the chain's; the net is a decrease of about
// 1,393 ISLM, so TotalDelegations drops. The rate update was deleted in PR 2, so this no
// longer reaches the stISLM redemption rate. The stored SharesToTokensRate is deliberately
// left as is: the day-0 refresh updates it, and the slash callback then finds tracked
// delegation equal to on-chain shares × the refreshed rate, so nothing is applied twice.
//
// Generated 2026-09-30 01:41 UTC by scripts/wind-down/gen_delta_table.py from
// measure_delegation_drift.py (Haqq REST rest.cosmos.haqq.network; 52 tracked validators).
// The 2026-09-29 21:36 UTC run (net -1,758 ISLM) differed:
// gmocoin's 364.79 ISLM over-recording no longer shows (its slash was booked on chain between
// the two measurements). Regenerate right before the upgrade proposal;
// the mainnet-export suite (PR 6) fails if this table no longer matches state.
var HaqqDelegationDeltas = []DelegationDelta{
	{Name: "neuler", Address: "haqqvaloper1a57vprf7lswm3aqy2g5gy509235wmtsfvf9q73", Delta: mustInt("-853800913870811785762")},
	{Name: "kioqqhaqqsh", Address: "haqqvaloper1a4qnqnk5ag0um6z3unkdth92v9c46x0dcpe2n0", Delta: mustInt("-448303270428234546889")},
	{Name: "takamulfivalidator", Address: "haqqvaloper1gcw6ru5akcmzvr3anqre8f3m8eml2rmrz8m6sl", Delta: mustInt("-91337544757772840484")},
	{Name: "stakingcabin", Address: "haqqvaloper1f6hy5d68hkx9wtp8er3mgdfwjfjr7tun5yu7us", Delta: mustInt("-29944721483185573")},
	{Name: "islamicstaking", Address: "haqqvaloper1p8k6xk94u24vv9dmxu3vkgg43fs3v72grkpjhm", Delta: mustInt("-3216140")},
	{Name: "foreststaking", Address: "haqqvaloper1ktu3f367c0j0yet6apuefy8xt5n7eswns7yuff", Delta: mustInt("677576")},
	{Name: "gmocoin", Address: "haqqvaloper1f2j8t0ddtak6z9td28wmv60xj6mlykzx65jr8w", Delta: mustInt("258119")},
	{Name: "surestake", Address: "haqqvaloper16hy887wxzjmmkkfrdxzgz9dlv6mfru56q539cw", Delta: mustInt("-203557")},
	{Name: "masterblox", Address: "haqqvaloper1ja29wpj6l5t42jj67vgqcuj8pu046uqklss524", Delta: mustInt("31310")},
	{Name: "stakeme", Address: "haqqvaloper1p02zk5ecdanap637e2wtt82cucjlxtkrhus623", Delta: mustInt("-1233")},
	{Name: "haqqassociation", Address: "haqqvaloper16lp0xpq87cre5z4jkfddq78r5l4vcd7el2jlmj", Delta: mustInt("555")},
	{Name: "noders", Address: "haqqvaloper1hgggrfgjeu4d5nveh03c6w37magsuqcy84p44t", Delta: mustInt("-4")},
	{Name: "staketake", Address: "haqqvaloper1tnm7y48w5nh8wt0s2u0fxwu607xtqhk6v99773", Delta: mustInt("-2")},
	{Name: "alxvoyanodeteam", Address: "haqqvaloper1wgm35c4nzs6ssgktd0zj4pefdr3h8ms5252mqc", Delta: mustInt("-2")},
	{Name: "segastakers", Address: "haqqvaloper10jqmd8rvggegva0r5smarr7avwwa8vce0q5gsh", Delta: mustInt("-1")},
	{Name: "palamar", Address: "haqqvaloper1xp597fjhgu6dx3a525htulkn36fqqntjaqvhct", Delta: mustInt("-1")},
}

// ReconcileHaqqDelegations applies HaqqDelegationDeltas to the haqq_11235-1 host zone so that
// tracked validator delegations (and TotalDelegations) match what is staked on Haqq. It
// returns the net delta applied and whether any row was applied; it never returns an upgrade
// error (see reconcileHostZoneDelegations). A row whose tracked delegation differs from
// HaqqExpectedTrackedDelegations is skipped on its own; the remaining rows are applied
// all-or-nothing. Nothing is queued afterwards: the wind-down drains every delegation by admin tx.
func ReconcileHaqqDelegations(ctx sdk.Context, sk stakeibckeeper.Keeper) (appliedDelta sdkmath.Int, applied bool) {
	deltas := HaqqDelegationDeltas
	if hostZone, found := sk.GetHostZone(ctx, HaqqChainId); found {
		deltas = haqqRowsMatchingPins(ctx, hostZone)
	}
	if len(deltas) == 0 {
		ctx.Logger().Error("v35: no haqq delegation table row matches its pin, nothing applied")
		return sdkmath.ZeroInt(), false
	}
	return reconcileHostZoneDelegations(ctx, sk, HaqqChainId, deltas)
}

// HaqqExpectedTrackedDelegations pins HaqqDelegationDeltas to the state it was measured against:
// each table validator's tracked Delegation (host base denom) as the 2026-09-30 drift measurement
// recorded it. A slash booked between generation and the upgrade changes that validator's
// tracked value, and applying its delta on top would double-apply the slash, so
// ReconcileHaqqDelegations skips that row. Emitted by gen_delta_table.py from the same drift.json
// rows as the deltas (their `recorded` value).
var HaqqExpectedTrackedDelegations = map[string]sdkmath.Int{
	"haqqvaloper1a57vprf7lswm3aqy2g5gy509235wmtsfvf9q73": mustInt("8550624701423372833667126"),
	"haqqvaloper1a4qnqnk5ag0um6z3unkdth92v9c46x0dcpe2n0": mustInt("1977324198402384201312159"),
	"haqqvaloper1gcw6ru5akcmzvr3anqre8f3m8eml2rmrz8m6sl": mustInt("942107178025082374262442"),
	"haqqvaloper1f6hy5d68hkx9wtp8er3mgdfwjfjr7tun5yu7us": mustInt("2436918471544559988644179"),
	"haqqvaloper1p8k6xk94u24vv9dmxu3vkgg43fs3v72grkpjhm": mustInt("3404566862077802114673425"),
	"haqqvaloper1ktu3f367c0j0yet6apuefy8xt5n7eswns7yuff": mustInt("3049476711114360580564391"),
	"haqqvaloper1f2j8t0ddtak6z9td28wmv60xj6mlykzx65jr8w": mustInt("1218094444964848572284210"),
	"haqqvaloper16hy887wxzjmmkkfrdxzgz9dlv6mfru56q539cw": mustInt("514296911029389874078629"),
	"haqqvaloper1ja29wpj6l5t42jj67vgqcuj8pu046uqklss524": mustInt("0"),
	"haqqvaloper1p02zk5ecdanap637e2wtt82cucjlxtkrhus623": mustInt("5566110280315914752203104"),
	"haqqvaloper16lp0xpq87cre5z4jkfddq78r5l4vcd7el2jlmj": mustInt("19837445550880585209244"),
	"haqqvaloper1hgggrfgjeu4d5nveh03c6w37magsuqcy84p44t": mustInt("1631295331676570357712739"),
	"haqqvaloper1tnm7y48w5nh8wt0s2u0fxwu607xtqhk6v99773": mustInt("459348774183198602894980"),
	"haqqvaloper1wgm35c4nzs6ssgktd0zj4pefdr3h8ms5252mqc": mustInt("10713097828733603828494"),
	"haqqvaloper10jqmd8rvggegva0r5smarr7avwwa8vce0q5gsh": mustInt("6512322804323302301356"),
	"haqqvaloper1xp597fjhgu6dx3a525htulkn36fqqntjaqvhct": mustInt("2539004185409864107364697"),
}

// haqqRowsMatchingPins returns the table rows whose validator is tracked at exactly its expected
// value, logging an error for each row it drops. A validator missing from the host zone is kept,
// so reconcileHostZoneDelegations sees it and skips the table, as for any missing validator.
func haqqRowsMatchingPins(ctx sdk.Context, hostZone stakeibctypes.HostZone) []DelegationDelta {
	matching := []DelegationDelta{}
	for _, entry := range HaqqDelegationDeltas {
		validator, _, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, entry.Address)
		if !found {
			matching = append(matching, entry)
			continue
		}

		// A nil delegation is zero, as in reconcileHostZoneDelegations, so it matches a pin of 0
		tracked := validator.Delegation
		if tracked.IsNil() {
			tracked = sdkmath.ZeroInt()
		}
		expected, hasExpected := HaqqExpectedTrackedDelegations[entry.Address]
		if !hasExpected || !tracked.Equal(expected) {
			ctx.Logger().Error(fmt.Sprintf("v35: validator %s (%s) tracked delegation is %v, expected %v; its haqq delegation "+
				"delta (%v) was measured against different state and is NOT applied, reconcile it after the upgrade",
				entry.Name, entry.Address, tracked, expected, entry.Delta))
			continue
		}
		matching = append(matching, entry)
	}
	return matching
}
