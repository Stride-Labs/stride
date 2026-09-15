package v34

import (
	"fmt"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
)

const InjectiveChainId = "injective-1"

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
var InjectiveDelegationDeltas = []DelegationDelta{
	{Name: "zellic", Address: "injvaloper13v0sulppc8pgtk4907p3vaqgq54vyl9yqmf7av", Delta: mustInt("1205585324378120818521")},
	{Name: "blackpanther", Address: "injvaloper10pe4avat38u38yzj5hvnw235uecfff6c73s8fn", Delta: mustInt("-666635876523567210103")},
	{Name: "autostake", Address: "injvaloper1acgud5qpn3frwzjrayqcdsdr9vkl3p6hrz34ts", Delta: mustInt("-518581018071438645019")},
	{Name: "everstake", Address: "injvaloper134dct56cq5v7uerxcy2cn4m06mqf4dxrlgpp24", Delta: mustInt("15998038353273365575")},
	{Name: "scvsecurity", Address: "injvaloper19a77dzm2lrxt2gehqca3nyzq077kq7qsgvmrp4", Delta: mustInt("13071078955306058570")},
	{Name: "figment", Address: "injvaloper1g4d6dmvnpg7w7yugy6kplndp7jpfmf3krtschp", Delta: mustInt("12730268614446851914")},
	{Name: "informal", Address: "injvaloper10xhy8xurfwts9ckjkq0ga92mrjz9txyygymqzp", Delta: mustInt("10745549570619704629")},
	{Name: "cointelegraph", Address: "injvaloper1vgu49xjgle9k84smvrwlyuc38u0cakuz7q3a5f", Delta: mustInt("9923595219135734648")},
	{Name: "falconx", Address: "injvaloper1da0shwz2mcup5rxkykquc9a7mh4s2hkeah9wuc", Delta: mustInt("9422403541401606634")},
	{Name: "readyblock", Address: "injvaloper1yljq5pdnx84kkg30jfmz6ddu4eyp7twyp4z40f", Delta: mustInt("9081593200542399585")},
	{Name: "innovatingcapital", Address: "injvaloper1rqqpyuka5dxulzjslnzjld2ltcw5095rr0jz07", Delta: mustInt("7377541496246364342")},
	{Name: "chorusone", Address: "injvaloper14yeq3lkajldaggj28hmq8xng9xux7x5g46hezv", Delta: mustInt("6936492819840331699")},
	{Name: "cryptocrew", Address: "injvaloper1nq37nq79w2j76xj8qhjzcn6vh0wlx0qk2r7zm6", Delta: mustInt("6816206817184141088")},
	{Name: "kiln2", Address: "injvaloper1vm2gflv53mzy9ak4f8cr9ml3xus980wywe9fx2", Delta: mustInt("6715968481637315365")},
	{Name: "nansen", Address: "injvaloper1nm48eujr28u3htqrjumfwhytn63rmca2prtklt", Delta: mustInt("6194729136793822234")},
	{Name: "polkachu", Address: "injvaloper125fkz3mq6qxxpkmphdl3ep92t0d3y9695mhclt", Delta: mustInt("5693537459059694194")},
	{Name: "decentriolabs", Address: "injvaloper1mlsg82x0mnw88u2teg0kceautjktgcl08tqfs4", Delta: mustInt("5613346790622233736")},
	{Name: "hextrust2", Address: "injvaloper16jkt2gcm8qfxzjj8j6pvyc045ga82pmthz58mp", Delta: mustInt("5092107445778740603")},
	{Name: "twinstake2", Address: "injvaloper1zt0x89kt3jhflyt69l4cpwcxhtgjaxteyhdgdv", Delta: mustInt("5052012111560010362")},
	{Name: "bitgobytwinstake2", Address: "injvaloper1ae4f9ae2d94kwhn6xlshm34f9fl0p0rkwy9wj5", Delta: mustInt("5011916777341280122")},
	{Name: "mipool", Address: "injvaloper1f68rd44mhx5lu9nz4lkq9wjeucj0gcja66sy8l", Delta: mustInt("4631011102263342832")},
	{Name: "imperator", Address: "injvaloper1esud09zs5754g5nlkmrgxsfdj276xm64cgmd3w", Delta: mustInt("4530772766716516697")},
	{Name: "republiccrypto", Address: "injvaloper1nxq05qdle8w5gm6fagda56euwt4rgvv4ktddcc", Delta: mustInt("4270153094294770661")},
	{Name: "bharvest", Address: "injvaloper1zpy3qf7us3m0pxpqkp72gzjv55t70huy33t47x", Delta: mustInt("3347960407263975116")},
	{Name: "smartstake", Address: "injvaloper1xwsnq88kc8wcrp34qenxf3dvhl5n02yj93u755", Delta: mustInt("3247722071717149514")},
	{Name: "highstakes", Address: "injvaloper1f2kdg34689x93cvw2y59z7y46dvz2fk8lhddfz", Delta: mustInt("3067293067732863711")},
	{Name: "helios", Address: "injvaloper1ffsdugrhfzdyxltjve8v68n6aazyc6p97uhfn0", Delta: mustInt("3007150066404768433")},
	{Name: "keplr", Address: "injvaloper1845wspsvm3z95a2zycz3t49gn7celq48dlyscc", Delta: mustInt("2886864063748577350")},
	{Name: "nttdigital", Address: "injvaloper177zqwtsnuhax28w5xpmf33wsl7l2xxf6y6xshc", Delta: mustInt("2726482726873656382")},
	{Name: "crosnest", Address: "injvaloper1fqrdtx7pyps6eytn3356j9cs4f8zl0eevlt3rt", Delta: mustInt("2526006055780005176")},
	{Name: "deutschetelekom", Address: "injvaloper1nngrhnm65wm8pu7wkah6hfs3vpd0xw463ydd65", Delta: mustInt("2285434050467623729")},
	{Name: "lavenderfive", Address: "injvaloper155yk4wfn0xqye80exlsr6hu4qdfsvsgwg3jckk", Delta: mustInt("2105005046483337596")},
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
