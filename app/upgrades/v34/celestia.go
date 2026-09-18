package v34

import (
	"fmt"
	"sort"

	"github.com/cosmos/gogoproto/proto"
	icatypes "github.com/cosmos/ibc-go/v11/modules/apps/27-interchain-accounts/types"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	icacallbackskeeper "github.com/Stride-Labs/stride/v34/x/icacallbacks/keeper"
	recordskeeper "github.com/Stride-Labs/stride/v34/x/records/keeper"
	recordstypes "github.com/Stride-Labs/stride/v34/x/records/types"
	stakeibckeeper "github.com/Stride-Labs/stride/v34/x/stakeibc/keeper"
	stakeibctypes "github.com/Stride-Labs/stride/v34/x/stakeibc/types"
	staketiakeeper "github.com/Stride-Labs/stride/v34/x/staketia/keeper"
)

const CelestiaChainId = "celestia"

// CelestiaDelegationChannelId is the DELEGATION ICA channel that was active when
// CelestiaDelegationDeltas was measured (2026-09-18). Re-measure with the table.
//
// The delta table only changes when a delegate executes on Celestia without its acknowledgement
// being booked, and such a packet leaves its commitment on Stride forever: it can be neither
// acked (the ordered channel closed) nor timed out (the host received it). So an unchanged
// active channel with no packet commitments proves nothing executed unbooked since the table
// was measured, and the reconciliation is exact. A restore would move the active channel and
// hide that evidence, which is why the channel id is pinned rather than looked up.
const CelestiaDelegationChannelId = "channel-862"

// CelestiaDelegationDeltas trues up the celestia host zone's tracked delegations to what is
// actually staked on Celestia.
//
// Background: the celestia DELEGATION ICA channel died ~90 times between June and September
// 2026. Each automated restore flipped every DELEGATION_IN_PROGRESS deposit record back to
// DELEGATION_QUEUE, so records whose delegate had already executed on Celestia (ack stranded on
// the dead channel) were re-sent and executed again with other records' liquid. The liquid pool
// drained into validator stake the host zone never booked, while every record stayed open. The
// untracked ("phantom") stake equals the open deposit records that no longer have liquid behind
// them; ReconcileCelestia books the stake and retires the same amount of records, so the
// redemption rate does not move.
//
// Measured 2026-09-18 at Stride height 40362843 / Celestia height 14149694 by diffing x/staking
// delegations of the delegation ICA celestia1rdfn69mf3xey2jlqyp70vljtx6df4lkydndplz9drdueuhqh8geqk9gjkj
// against the host zone validators (89 non-zero entries, all positive, sum 15,439,858,896 utia).
//
// !!! RE-MEASURE right before the proposal !!! The commands:
//
//	on a Celestia node:  GET /cosmos/staking/v1beta1/delegations/<delegation ica>?pagination.limit=500
//	on Stride:           GET /Stride-Labs/stride/stakeibc/host_zone/celestia
//	delta               = on-chain delegation balance minus tracked validator.delegation, per address
//
// Include every validator whose delta != 0.
var CelestiaDelegationDeltas = []DelegationDelta{
	{Name: "go", Address: "celestiavaloper1uvytvhunccudw8fzaxvsrumec53nawyj939gj9", Delta: mustInt("1476091841")},
	{Name: "finoaconsensusservices", Address: "celestiavaloper1e2p4u5vqwgum7pm9vhp0yjvl58gvhfc6yfatw4", Delta: mustInt("1446695576")},
	{Name: "chorusone", Address: "celestiavaloper15urq2dtp9qce4fyc85m6upwm9xul3049gwdz0x", Delta: mustInt("1248473160")},
	{Name: "p2porg", Address: "celestiavaloper1r4kqtye4dzacmrwnh6f057p50pdjm8g59tlhhg", Delta: mustInt("994893330")},
	{Name: "cuck", Address: "celestiavaloper1jkuw8rxxrsgn9pq009987kzelkp46cgcaa7lhj", Delta: mustInt("803032100")},
	{Name: "figment", Address: "celestiavaloper1xqc7w3pe38kg4tswjt7mnvks7gy4p38vtsuycj", Delta: mustInt("532260209")},
	{Name: "p-opsteam", Address: "celestiavaloper1ac4mnwg79gyvd0x5trl2fgjv07lgfas02jf378", Delta: mustInt("516787522")},
	{Name: "cosmostation", Address: "celestiavaloper1uqj5ul7jtpskk9ste9mfv6jvh0y3w34vtpz3gw", Delta: mustInt("468822219")},
	{Name: "twinstake", Address: "celestiavaloper1mhux0vt6qszz8qwv8axggt02jjm7tuvdfhz78j", Delta: mustInt("454896810")},
	{Name: "keplr", Address: "celestiavaloper1uwmf03ke52vld2sa9khs0nslpgzwsm5xs5e4pn", Delta: mustInt("391458818")},
	{Name: "imperator", Address: "celestiavaloper1r5xt7twqmh39ky72f4txxjrhlt2z0qwwmdal8c", Delta: mustInt("375378171")},
	{Name: "nodesguru", Address: "celestiavaloper1hm2d8e6nd5ngtlte3hlded03vgj3rer94vmdml", Delta: mustInt("338851708")},
	{Name: "ai", Address: "celestiavaloper1v5hrqlv8dqgzvy0pwzqzg0gxy899rm4klzxm07", Delta: mustInt("320284498")},
	{Name: "polkachu", Address: "celestiavaloper1jgqewpzn7dww5tlpnkypm72fm8tjznmw7ll7ls", Delta: mustInt("302344735")},
	{Name: "binarybuilders", Address: "celestiavaloper15kpw453rgqrranltr8pcy9muryf3jjd7esw38j", Delta: mustInt("301530838")},
	{Name: "galaxydigital", Address: "celestiavaloper1463wx5xkus5hyugyecvlhv9qpxklz62kyhwcts", Delta: mustInt("284697336")},
	{Name: "upnode|0%fee", Address: "celestiavaloper133t4gpv4vhpqgfn9gr8l4u423zrglg8rkqeupr", Delta: mustInt("283691804")},
	{Name: "qubelabs", Address: "celestiavaloper1j2jq259d3rrc24876gwxg0ksp0lhd8gy49k6st", Delta: mustInt("278508264")},
	{Name: "stakingfacilities", Address: "celestiavaloper1wu24jxpn9j0580ehjz344d58cf3t7lzrrgqmnr", Delta: mustInt("266094035")},
	{Name: "xprv", Address: "celestiavaloper1nwu3ugynh8m6r7aphv0uxnca84t7gnruvyye9c", Delta: mustInt("258015001")},
	{Name: "hextechnologies", Address: "celestiavaloper1ftmw4wh8dq2ljw0xq3xgg00dl7l20se3lrml7q", Delta: mustInt("255299239")},
	{Name: "01node", Address: "celestiavaloper1murrqgqahxevedty0nzqrn5hj434fvffxufxcl", Delta: mustInt("257020047")},
	{Name: "iykyk", Address: "celestiavaloper1amxp3ah9anq4pmpnsknls7sql3kras9hs8pu0g", Delta: mustInt("253234895")},
	{Name: "zkv", Address: "celestiavaloper187avawwq7qhanrkxf45mayztdqsr49hu8lezdh", Delta: mustInt("224125397")},
	{Name: "nocturnallabs", Address: "celestiavaloper1zqjpfxtv3yp6kdlgra4hc9zehxgvpaw82hxr5w", Delta: mustInt("213692077")},
	{Name: "nansen", Address: "celestiavaloper1c9ye54e3pzwm3e0zpdlel6pnavrj9qqvjeqh0m", Delta: mustInt("164728766")},
	{Name: "stakecito", Address: "celestiavaloper1qe8uuf5x69c526h4nzxwv4ltftr73v7q5qhs58", Delta: mustInt("153217828")},
	{Name: "validatus", Address: "celestiavaloper19v94c3z7ckarwsum76kaagma0wqsqhh5nl5zqg", Delta: mustInt("151632281")},
	{Name: "mzonder", Address: "celestiavaloper1demcj83q7nxt6gtqxlu7qsmwqpt4jspj3alr92", Delta: mustInt("134612343")},
	{Name: "blackblocks", Address: "celestiavaloper1knn88yl08ctsdrtxvfp39jywt7rph9ptv7532y", Delta: mustInt("128953032")},
	{Name: "staked", Address: "celestiavaloper1pmn4cjwf26hkcpvyl322glhpdpcemcel8ca2vl", Delta: mustInt("103666974")},
	{Name: "w3hitchhiker", Address: "celestiavaloper1zdrz4w2pwwffdvmpum0626vycel9caay9n3pll", Delta: mustInt("80457954")},
	{Name: "nodestake", Address: "celestiavaloper19f0w9svr905fhefusyx4z8sf83j6et0g57nch8", Delta: mustInt("79027866")},
	{Name: "lavenderfivenodes", Address: "celestiavaloper140l6y2gp3gxvay6qtn70re7z2s0gn57zcvqd22", Delta: mustInt("73025948")},
	{Name: "allnodes", Address: "celestiavaloper1rcm7tth05klgkqpucdhm5hexnk49dfda3l3hak", Delta: mustInt("73311496")},
	{Name: "alphab", Address: "celestiavaloper1v987evnk7hsqct7smdqpxqprhvlcxgt43kyewc", Delta: mustInt("71174356")},
	{Name: "latamnodes", Address: "celestiavaloper14v4ush42xewyeuuldf6jtdz0a7pxg5fwrlumwf", Delta: mustInt("67069831")},
	{Name: "nodeguardians", Address: "celestiavaloper1m58punvt32u07ra4p6x7krgxakye3m90rzgm4c", Delta: mustInt("63438009")},
	{Name: "enigma", Address: "celestiavaloper107lwx458gy345ag2afx9a7e2kkl7x49y3433gj", Delta: mustInt("61890739")},
	{Name: "nakoturk", Address: "celestiavaloper1pnzrk7yzx0nr9xrcjyswj7ram4qxlrfz28xvn6", Delta: mustInt("60745506")},
	{Name: "kjnodes", Address: "celestiavaloper17p8y0sm76zhrtjny98tknevafvlq9860ehykz3", Delta: mustInt("60343485")},
	{Name: "easy2stake", Address: "celestiavaloper1un77nfm6axkhkupe8fk4xl6fd4adz3y59fucpu", Delta: mustInt("55638496")},
	{Name: "forbole", Address: "celestiavaloper1593ns00rftlqp2gyu6wdmrqpgv5frv0hsf4sw2", Delta: mustInt("51059870")},
	{Name: "kiln", Address: "celestiavaloper1djqecw6nn5tydxq0shan7srv8j65clsfmnxcfu", Delta: mustInt("49512600")},
	{Name: "b-harvest", Address: "celestiavaloper14ntfv0qkg8522xe0pvrfgqxcmnj5x466v8z3tl", Delta: mustInt("49512596")},
	{Name: "brightlystake", Address: "celestiavaloper19y52qzj4hxw0u68krfptkjlm77cvth8dgum7yu", Delta: mustInt("47965320")},
	{Name: "stakin", Address: "celestiavaloper1dlsl4u42ycahzjfwc6td6upgsup9tt7cz8vqm4", Delta: mustInt("46418059")},
	{Name: "everstake", Address: "celestiavaloper1eualhqh07w7p45g45hvrjagkcxsfnflzdw5jzg", Delta: mustInt("44870796")},
	{Name: "frens", Address: "celestiavaloper1ej2es5fjztqjcd4pwa0zyvaevtjd2y5wh8xeg4", Delta: mustInt("43323529")},
	{Name: "stakelyio", Address: "celestiavaloper1yknsyf9ws4ugtv3r9g43kwqkne4zmrupcxhlth", Delta: mustInt("43350533")},
	{Name: "dsrv", Address: "celestiavaloper1vje2he3pcq3w5udyvla7zm9qd5yes6hzffsjxj", Delta: mustInt("40228985")},
	{Name: "projectblanc", Address: "celestiavaloper1auqmdc2pnx5gxvakjdsc9v9zc4y2ga0hcaxg9x", Delta: mustInt("39016115")},
	{Name: "larryengineer", Address: "celestiavaloper1ryyzale2qcp3e35k0ze3kc0mpfdtw9jagcss3k", Delta: mustInt("37134456")},
	{Name: "stakelab", Address: "celestiavaloper1lqfqp2w65pjsu6hg3qg4qfp4t2t6da4a3gg8ad", Delta: mustInt("37134456")},
	{Name: "swissstaking", Address: "celestiavaloper1u4vhh70lwlt2va7hw5evzl6sap92t0m9nqzmud", Delta: mustInt("34039919")},
	{Name: "freshstaking", Address: "celestiavaloper17h2x3j7u44qkrq0sk8ul0r2qr440rwgj8g0ng0", Delta: mustInt("34039919")},
	{Name: "stakingcabin", Address: "celestiavaloper1vdp8q3v72mntewqqak56yk3gzz7h5ukmeym9hk", Delta: mustInt("34039919")},
	{Name: "contributiondao", Address: "celestiavaloper17srrapx2cvqyyy4rg3menq0hn86py8f3klelhl", Delta: mustInt("34916059")},
	{Name: "bwarelabs", Address: "celestiavaloper1vl3dkus7g0cj4lg0e2jrqk2reukmlht0ee7cr4", Delta: mustInt("30945383")},
	{Name: "lemniscap", Address: "celestiavaloper12t63cy8kn5n7qw77xvjn00ymcr0uuvz2vh8p79", Delta: mustInt("27850849")},
	{Name: "kek", Address: "celestiavaloper1njzuxja7aa7w2d69ldg2r8c6qhjzfcd42huq0x", Delta: mustInt("26303573")},
	{Name: "staking4all", Address: "celestiavaloper19kr4f4ndyek6kwa0vt3w4un8he0tkekufa8t2g", Delta: mustInt("26303573")},
	{Name: "gunter", Address: "celestiavaloper1pavac9yrlgwyw6v9yx84sttc96n9ee9zrja2u7", Delta: mustInt("26303573")},
	{Name: "cumulo", Address: "celestiavaloper1cs37tvmahavw8xcnzcgyz342sh0al37ma4zqat", Delta: mustInt("26303573")},
	{Name: "cryptocrewvalidators", Address: "celestiavaloper138jl42zlxue4wpvnugcdqhxjmyd2vpt6qhs5ls", Delta: mustInt("25377691")},
	{Name: "validatrium", Address: "celestiavaloper125xazqstxpav7ekrt4w8km7ccdu8ytj40xug70", Delta: mustInt("25377691")},
	{Name: "bitnordic", Address: "celestiavaloper1slnzmhg3kwhc2c5y9atrt5jtmt3sauzzp4tguj", Delta: mustInt("25377691")},
	{Name: "unbonding-pleaseredelegate", Address: "celestiavaloper1qx43f066sh6728avms4qq09cj2a3mg83dgjh22", Delta: mustInt("25377691")},
	{Name: "chainodetech", Address: "celestiavaloper1qxeza0sa037u35p3ze8p7ka7emajvydnyjlp07", Delta: mustInt("23791581")},
	{Name: "encapsulate(fkakingsuper)", Address: "celestiavaloper1s0lankh33kprer2l22nank5rvsuh9ksa2xcd2y", Delta: mustInt("23791581")},
	{Name: "moonlet", Address: "celestiavaloper1q8teur40emyun60et4wh5z6yj5669stgz8xs59", Delta: mustInt("23791581")},
	{Name: "meria", Address: "celestiavaloper1yecxnyegvgm5dwsx0r3jsgr74ju6mlxdwkxx8g", Delta: mustInt("23791581")},
	{Name: "wavefive", Address: "celestiavaloper1cmaga4f6f7pttuwzfldn9067ld3090uw0e6zq8", Delta: mustInt("23791581")},
	{Name: "conqueror", Address: "celestiavaloper10f8l8m4879h40848rsvxat797t3a5ghgdsjgzl", Delta: mustInt("23791581")},
	{Name: "noders", Address: "celestiavaloper139mu0a0ucz0gmrkavm5wjar2lpx7yvxq3e25e5", Delta: mustInt("23791581")},
	{Name: "stakesquid", Address: "celestiavaloper1l0zmpm02u240crndlj7hkvlj5azuglv4emfczt", Delta: mustInt("23791581")},
	{Name: "stakerspace", Address: "celestiavaloper1qdfdh8stxpkj4zz46x2n9ejyyy9h0c86425yjm", Delta: mustInt("22205473")},
	{Name: "spidey", Address: "celestiavaloper1st4h4ymu52hwtl3s0n298t6adj0vputx8jw8xt", Delta: mustInt("22205473")},
	{Name: "strangelove", Address: "celestiavaloper1sl97x54v0u3extuj2zrf7h0qrrtpgpslfrjjry", Delta: mustInt("22205473")},
	{Name: "kooltek68", Address: "celestiavaloper1ax83exaawlmy5p2qn22gcynrchwdntt5xvj0qu", Delta: mustInt("20619377")},
	{Name: "counterpoint", Address: "celestiavaloper1vfydl5r98zev8xc7j0mus28r63jcklsu63vuah", Delta: mustInt("20619377")},
	{Name: "blockscope", Address: "celestiavaloper1cqgzxhn3dqd58xexz8yley2wntdvym4emzvpd7", Delta: mustInt("20619377")},
	{Name: "activenodes", Address: "celestiavaloper1t345w0vxnyyrf4eh43lpd3jl7z378rtsdn9tz3", Delta: mustInt("19033269")},
	{Name: "partnerstaking", Address: "celestiavaloper1clf3nqp89h97umhl4fmcqr642jz6rszcxegjc6", Delta: mustInt("19033269")},
	{Name: "mhventures", Address: "celestiavaloper1q2kaajedxm0r5xc0twdqz6atap96502d67yjyj", Delta: mustInt("15861063")},
	{Name: "kiln2", Address: "celestiavaloper1uxlf7mvr8nep3gm7udf2u9remms2jyjqa02qlj", Delta: mustInt("2491039")},
	{Name: "informal", Address: "celestiavaloper1x20lytyf6zkcrv5edpkfkn8sz578qg5spge2ru", Delta: mustInt("1660692")},
	{Name: "smartstake", Address: "celestiavaloper1tzm96yvupy3egtpg6uw0s2f4c85qayw3afwetd", Delta: mustInt("1266184")},
	{Name: "validao", Address: "celestiavaloper1gfn5m2vjqk4kcdg2zwhzgpvu60jz0f9duhu594", Delta: mustInt("507223")},
}

// StaketiaRemainingDelegatedBalanceDelta corrects staketia's remaining_delegated_balance (the
// utia the multisig delegation account still holds delegated) down to what is actually staked.
//
// Root cause: a v25 migration straddle. Before v25 (2025-02-06, gov prop 260) RedeemStake left
// the delegated balance alone and ConfirmUndelegation decremented it; after v25 RedeemStake
// decrements staketia's remaining_delegated_balance at redeem time and ConfirmUndelegation
// decrements stakeibc's total_delegations instead. Unbonding record 884 accumulated redemptions
// from 2025-02-03 to 02-07 (mostly under the old code, so no staketia decrement) and was
// confirmed on 2025-02-13 under the new code (stakeibc decrement only). Its native amount,
// 39,829,976,251 utia, is 99.4% of the gap; the residual 246,766,592 utia is rate-lag drift
// (staketia decrements by the redeem-time estimate, stakeibc by the amount finalized up to four
// days later at a slightly higher rate) and is left as-is.
//
// Measured 2026-09-18. Re-measure right before the proposal:
//
//	delta = (multisig on-chain delegation
//	         - native amount owed by the current ACCUMULATING_REDEMPTIONS and UNBONDING_QUEUE records)
//	        - remaining_delegated_balance
//
// The subtraction matters because remaining_delegated_balance is already reduced for
// redemptions the multisig has not undelegated yet (that term was zero on 2026-09-18). The delta
// stays valid as long as no MsgAdjustDelegatedBalance or ConfirmDelegation lands in between.
var StaketiaRemainingDelegatedBalanceDelta = sdkmath.NewInt(-40_076_742_843)

// ReconcileCelestia books the Celestia stake that executed without an acknowledgement and retires
// exactly the same amount of phantom deposit records, leaving the redemption rate unchanged.
//
// The phantom amount is the sum of CelestiaDelegationDeltas, never a separate constant: using one
// number for both halves is what guarantees the rate does not move. The record set is not a
// constant because phantom records churn between epochs (queue records are fungible claims on
// one liquid pool, so removing any set summing to the phantom amount leaves the survivors equal
// to the real liquid).
//
// Order matters and nothing is written until every check passes:
//  1. Host zone present (absent means non-mainnet).
//  2. The delta table applies in full (every validator present, no delegation driven negative).
//  3. Celestia DELEGATION_QUEUE then DELEGATION_IN_PROGRESS records, each ascending by id, cover
//     the phantom amount. TRANSFER_QUEUE records are never touched (their coins are on Stride).
//  4. Every delegate callback on the celestia DELEGATION port, on any channel, decodes cleanly.
//     This is checked before any write so a malformed entry causes a clean skip instead of
//     deleting the deposit record its (unreadable) callback references, which would otherwise
//     leave the callback behind to fail unmarshalling on its eventual ack rather than being the
//     no-op a removed record's callback is supposed to be.
//  5. Apply the validator deltas (raises TotalDelegations by the phantom amount).
//  6. Remove the phantom amount from the collected records: whole records are deleted while
//     they fit; the first record that would overshoot is shrunk if it is a queue record.
//  7. An in-progress record is never shrunk: its in-flight callback carries the original split
//     amounts and a success ack would subtract more than the shrunk record holds. It is deleted
//     instead and a fresh DELEGATION_QUEUE record is appended for the leftover. Every deleted
//     in-progress record's (already-decoded) delegate callbacks are removed (on any channel, so
//     the eventual ack is a no-op instead of failing on a missing record), and for callbacks on
//     the active DELEGATION channel the validators' DelegationChangesInProgress are decremented
//     as the callback would have.
//  8. Log the removals and totals.
//
// Any failed check is logged and the function returns false with nothing written. It never
// returns an error: halting the chain is disproportionate for an accounting fix.
func ReconcileCelestia(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
	rk recordskeeper.Keeper,
	ick icacallbackskeeper.Keeper,
) (applied bool) {
	hostZone, found := sk.GetHostZone(ctx, CelestiaChainId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v34: host zone %s not found, skipping delegation reconciliation", CelestiaChainId))
		return false
	}

	// Validate the delta table before touching anything; its sum is the phantom amount
	phantomAmount, valid := validateDelegationDeltas(ctx, hostZone, CelestiaDelegationDeltas)
	if !valid {
		return false
	}
	if !phantomAmount.IsPositive() {
		ctx.Logger().Error(fmt.Sprintf("v34: %s delta table sums to %v, expected a positive phantom amount; "+
			"reconciliation NOT applied", CelestiaChainId, phantomAmount))
		return false
	}

	// The records must cover the phantom amount, otherwise the two halves would not balance
	records := collectCelestiaDelegationRecords(ctx, rk)
	coverable := sdkmath.ZeroInt()
	for _, record := range records {
		coverable = coverable.Add(record.Amount)
	}
	if coverable.LT(phantomAmount) {
		ctx.Logger().Error(fmt.Sprintf("v34: %s delegation records total %v, less than the phantom amount %v; "+
			"reconciliation NOT applied, re-verify constants and reconcile in a later upgrade",
			CelestiaChainId, coverable, phantomAmount))
		return false
	}

	// Decode every delegate callback on the port before the first write, so a malformed entry
	// skips the whole reconciliation instead of deleting a record whose callback later fails to ack
	owner := stakeibctypes.FormatHostZoneICAOwner(CelestiaChainId, stakeibctypes.ICAAccountType_DELEGATION)
	portId, err := icatypes.NewControllerPortID(owner)
	if err != nil {
		ctx.Logger().Error(fmt.Sprintf("v34: unable to build %s delegation port id: %s", CelestiaChainId, err))
		return false
	}
	delegateCallbacks, decodable := collectCelestiaDelegateCallbacks(ctx, ick, portId)
	if !decodable {
		return false
	}

	// Only apply when no delegate can have executed unbooked since the table was measured
	// (see CelestiaDelegationChannelId)
	if !celestiaDelegationChannelIsQuiet(ctx, sk, hostZone, portId) {
		return false
	}

	// Every check passed; book the stake on the validators first
	appliedDelta, applied := reconcileHostZoneDelegations(ctx, sk, CelestiaChainId, CelestiaDelegationDeltas)
	if !applied {
		return false
	}

	// Then retire the same amount of records, cleaning up the callbacks of deleted in-progress records
	deletedInProgress, removedCount := removeAmountFromDepositRecords(ctx, rk, records, appliedDelta)
	hostZone, _ = sk.GetHostZone(ctx, CelestiaChainId)
	removeDelegateCallbacks(ctx, sk, ick, &hostZone, portId, delegateCallbacks, deletedInProgress)
	sk.SetHostZone(ctx, hostZone)

	ctx.Logger().Info(fmt.Sprintf("v34: %s reconciled: booked %v %s of unacknowledged stake and removed the same amount "+
		"from %d delegation records", CelestiaChainId, appliedDelta, hostZone.HostDenom, removedCount))
	return true
}

// collectCelestiaDelegationRecords returns the celestia deposit records that can be retired:
// DELEGATION_QUEUE records first, then DELEGATION_IN_PROGRESS, each group ascending by id.
func collectCelestiaDelegationRecords(ctx sdk.Context, rk recordskeeper.Keeper) []recordstypes.DepositRecord {
	queued := []recordstypes.DepositRecord{}
	inProgress := []recordstypes.DepositRecord{}
	for _, record := range rk.GetAllDepositRecord(ctx) {
		if record.HostZoneId != CelestiaChainId {
			continue
		}
		switch record.Status {
		case recordstypes.DepositRecord_DELEGATION_QUEUE:
			queued = append(queued, record)
		case recordstypes.DepositRecord_DELEGATION_IN_PROGRESS:
			inProgress = append(inProgress, record)
		}
	}

	byId := func(records []recordstypes.DepositRecord) {
		sort.Slice(records, func(i, j int) bool { return records[i].Id < records[j].Id })
	}
	byId(queued)
	byId(inProgress)
	return append(queued, inProgress...)
}

// removeAmountFromDepositRecords removes exactly `amount` from the records in order, deleting
// whole records while they fit and shrinking (queue) or splitting (in-progress) the first record
// that would overshoot. The caller guarantees the records cover the amount. It returns the ids of
// the deleted in-progress records, whose in-flight callbacks the caller must clean up, and the
// count of records it deleted, shrunk or split.
func removeAmountFromDepositRecords(
	ctx sdk.Context,
	rk recordskeeper.Keeper,
	records []recordstypes.DepositRecord,
	amount sdkmath.Int,
) (deletedInProgress map[uint64]bool, removedCount int) {
	deletedInProgress = map[uint64]bool{}
	remaining := amount
	for _, record := range records {
		if remaining.IsZero() {
			break
		}
		inProgress := record.Status == recordstypes.DepositRecord_DELEGATION_IN_PROGRESS

		// Whole records are deleted while they fit
		if remaining.GTE(record.Amount) {
			rk.RemoveDepositRecord(ctx, record.Id)
			if inProgress {
				deletedInProgress[record.Id] = true
			}
			remaining = remaining.Sub(record.Amount)
			removedCount++
			ctx.Logger().Info(fmt.Sprintf("v34: removed %s deposit record %d (%s, %v %s)",
				CelestiaChainId, record.Id, record.Status, record.Amount, record.Denom))
			continue
		}

		// The first record that would overshoot absorbs the remainder. A queue record is shrunk in
		// place; an in-progress record is replaced by a fresh queue record for the leftover
		leftover := record.Amount.Sub(remaining)
		if !inProgress {
			record.Amount = leftover
			rk.SetDepositRecord(ctx, record)
			removedCount++
			ctx.Logger().Info(fmt.Sprintf("v34: shrunk %s deposit record %d (%s) by %v to %v %s",
				CelestiaChainId, record.Id, record.Status, remaining, leftover, record.Denom))
			return deletedInProgress, removedCount
		}

		rk.RemoveDepositRecord(ctx, record.Id)
		deletedInProgress[record.Id] = true
		removedCount++
		replacementId := rk.AppendDepositRecord(ctx, recordstypes.DepositRecord{
			Amount:             leftover,
			Denom:              record.Denom,
			HostZoneId:         record.HostZoneId,
			Status:             recordstypes.DepositRecord_DELEGATION_QUEUE,
			Source:             record.Source,
			DepositEpochNumber: record.DepositEpochNumber,
		})
		ctx.Logger().Info(fmt.Sprintf("v34: removed %s deposit record %d (%s, %v %s) and re-queued its leftover %v as record %d",
			CelestiaChainId, record.Id, record.Status, record.Amount, record.Denom, leftover, replacementId))
		return deletedInProgress, removedCount
	}
	return deletedInProgress, removedCount
}

// celestiaDelegateCallback pairs a decoded delegate callback with the key and channel its
// CallbackData was stored under, so removeDelegateCallbacks can act on it without re-fetching or
// re-decoding CallbackData once collectCelestiaDelegateCallbacks has already done so.
type celestiaDelegateCallback struct {
	Key       string
	ChannelId string
	Callback  stakeibctypes.DelegateCallback
}

// celestiaDelegationChannelIsQuiet checks that the active DELEGATION channel is still the one the
// delta table was measured on and that it has no outstanding packet commitments. Either failure
// means a delegate may have executed on Celestia after the measurement, so the table can no
// longer be trusted and the reconciliation is deferred to a later upgrade.
func celestiaDelegationChannelIsQuiet(ctx sdk.Context, sk stakeibckeeper.Keeper, hostZone stakeibctypes.HostZone, portId string) bool {
	activeChannelId, found := sk.ICAControllerKeeper.GetActiveChannelID(ctx, hostZone.ConnectionId, portId)
	if !found || activeChannelId != CelestiaDelegationChannelId {
		ctx.Logger().Error(fmt.Sprintf("v34: %s active delegation channel is %q, expected %s; "+
			"reconciliation NOT applied, re-measure constants and reconcile in a later upgrade",
			CelestiaChainId, activeChannelId, CelestiaDelegationChannelId))
		return false
	}

	pending := sk.IBCKeeper.ChannelKeeper.GetAllPacketCommitmentsAtChannel(ctx, portId, activeChannelId)
	if len(pending) > 0 {
		ctx.Logger().Error(fmt.Sprintf("v34: %s delegation channel %s has %d unacknowledged packets; "+
			"reconciliation NOT applied, clear the channel and reconcile in a later upgrade",
			CelestiaChainId, activeChannelId, len(pending)))
		return false
	}
	return true
}

// collectCelestiaDelegateCallbacks decodes every delegate callback on portId, on any channel,
// before any accounting write begins. If any entry fails to unmarshal it logs an error and
// returns ok=false so the caller skips the whole reconciliation: leaving a malformed callback in
// place while deleting the deposit record it references would strand it to fail unmarshalling on
// its eventual ack, instead of the no-op a removed record's callback is supposed to become.
func collectCelestiaDelegateCallbacks(
	ctx sdk.Context,
	ick icacallbackskeeper.Keeper,
	portId string,
) (callbacks []celestiaDelegateCallback, ok bool) {
	for _, callbackData := range ick.GetAllCallbackData(ctx) {
		if callbackData.PortId != portId || callbackData.CallbackId != stakeibckeeper.ICACallbackID_Delegate {
			continue
		}
		delegateCallback := stakeibctypes.DelegateCallback{}
		if err := proto.Unmarshal(callbackData.CallbackArgs, &delegateCallback); err != nil {
			ctx.Logger().Error(fmt.Sprintf("v34: unable to unmarshal delegate callback %s on %s: %s; "+
				"reconciliation NOT applied, re-verify constants and reconcile in a later upgrade",
				callbackData.CallbackKey, portId, err))
			return nil, false
		}
		callbacks = append(callbacks, celestiaDelegateCallback{
			Key:       callbackData.CallbackKey,
			ChannelId: callbackData.ChannelId,
			Callback:  delegateCallback,
		})
	}
	return callbacks, true
}

// removeDelegateCallbacks deletes every pre-decoded delegate callback that references a deleted
// in-progress record, on any channel. For callbacks on the host zone's active DELEGATION channel
// (the packet is genuinely in flight), the validators' DelegationChangesInProgress are
// decremented as the callback would have done, never below zero. Callbacks on dead channels are
// orphans and are just deleted. The host zone is modified in memory; the caller persists it.
func removeDelegateCallbacks(
	ctx sdk.Context,
	sk stakeibckeeper.Keeper,
	ick icacallbackskeeper.Keeper,
	hostZone *stakeibctypes.HostZone,
	portId string,
	delegateCallbacks []celestiaDelegateCallback,
	deletedRecordIds map[uint64]bool,
) {
	if len(deletedRecordIds) == 0 {
		return
	}

	activeChannelId, found := sk.ICAControllerKeeper.GetActiveChannelID(ctx, hostZone.ConnectionId, portId)
	if !found {
		ctx.Logger().Info(fmt.Sprintf("v34: no active channel on %s; deleted records' callbacks are orphans", portId))
	}

	for _, entry := range delegateCallbacks {
		if !deletedRecordIds[entry.Callback.DepositRecordId] {
			continue
		}

		// Only a callback on the active channel still has a packet in flight whose completion
		// the validators are waiting on
		if found && entry.ChannelId == activeChannelId {
			for _, split := range entry.Callback.SplitDelegations {
				decrementDelegationChangesInProgress(ctx, hostZone, split.Validator)
			}
		}

		ick.RemoveCallbackData(ctx, entry.Key)
		ctx.Logger().Info(fmt.Sprintf("v34: removed delegate callback %s for deleted deposit record %d (active channel: %t)",
			entry.Key, entry.Callback.DepositRecordId, found && entry.ChannelId == activeChannelId))
	}
}

// decrementDelegationChangesInProgress mirrors DecrementValidatorDelegationChangesInProgress but
// logs and skips instead of erroring when the validator is missing or already at zero.
func decrementDelegationChangesInProgress(ctx sdk.Context, hostZone *stakeibctypes.HostZone, validatorAddress string) {
	validator, index, found := stakeibckeeper.GetValidatorFromAddress(hostZone.Validators, validatorAddress)
	if !found {
		ctx.Logger().Error(fmt.Sprintf("v34: validator %s referenced by a removed delegate callback is not on %s, skipping",
			validatorAddress, hostZone.ChainId))
		return
	}
	if validator.DelegationChangesInProgress == 0 {
		ctx.Logger().Error(fmt.Sprintf("v34: validator %s already has 0 delegation changes in progress, not decrementing",
			validatorAddress))
		return
	}
	validator.DelegationChangesInProgress -= 1
	hostZone.Validators[index] = &validator
}

// AdjustStaketiaRemainingDelegatedBalance applies a signed delta to the staketia host zone's
// remaining_delegated_balance only. Unlike MsgAdjustDelegatedBalance it deliberately does not
// mirror the change to stakeibc's TotalDelegations, which is already correct (see
// StaketiaRemainingDelegatedBalanceDelta).
//
// Skips with a log, writing nothing, if the staketia host zone is absent (non-mainnet) or the
// result would be negative (the constant no longer describes chain state). Never returns an error.
func AdjustStaketiaRemainingDelegatedBalance(ctx sdk.Context, stk staketiakeeper.Keeper, delta sdkmath.Int) (applied bool) {
	hostZone, err := stk.GetHostZone(ctx)
	if err != nil {
		ctx.Logger().Info("v34: staketia host zone not found, skipping remaining delegated balance adjustment")
		return false
	}

	current := hostZone.RemainingDelegatedBalance
	if current.IsNil() {
		current = sdkmath.ZeroInt()
	}
	adjusted := current.Add(delta)
	if adjusted.IsNegative() {
		ctx.Logger().Error(fmt.Sprintf("v34: staketia remaining delegated balance %v plus delta %v would be negative; "+
			"adjustment NOT applied, re-verify constants and adjust in a later upgrade", current, delta))
		return false
	}

	hostZone.RemainingDelegatedBalance = adjusted
	stk.SetHostZone(ctx, hostZone)
	ctx.Logger().Info(fmt.Sprintf("v34: staketia remaining delegated balance adjusted by %v: %v -> %v", delta, current, adjusted))
	return true
}
