# Protocol Wind-Down: Consolidated Single-Upgrade Design

Status: approved design (brainstorm of 2026-09-28), on branch `wind-down-design-consolidation`. This document revisits the
sequencing of `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` (the "two-upgrade
spec"). It changes only what must change to ship one upgrade instead of two; everything it does
not mention (facts §3, operator addresses §3a, the Osmosis side §7, drift measurement §8a,
accounting §9, validator compensation §9a, the exact removal lists) is inherited unchanged.

## §1. Why revisit

The two-upgrade spec separates "close the doors" (v35) from "halt, unbond, migrate" (v36) so that
the second upgrade lands on a chain with nothing in flight. That costs a full unbonding period of
waiting: window 1 exists only to flush the redemptions that were open at v35, and only then does
window 2 start the real unbonding. End to end is roughly 70 days.

Two things have changed since that split was chosen:

- Withdrawal mode is gone. There is no on-chain pool assertion and no on-chain redeem, so no
  upgrade needs the chain to be in a known, fully-drained state. The coverage check lives off
  chain and runs whenever ops want it to.
- Redemptions in flight no longer need a special path. They can finish through the pipeline
  that already exists (undelegate at the day epoch, sweep on completion, claim), in parallel with
  the admin drain of everything else, and the pool is funded from what is left. Nobody has to
  wait for them.

Combining the upgrades therefore removes window 1 and one governance cycle. The wind-down becomes
one upgrade and one ops window of roughly 40 days: the length of the longest unbonding period
plus a few days of ops on each side.

## §2. The design in one paragraph

One upgrade, v35, ships everything the two-upgrade spec spread across v35 and v36, with one
substitution: instead of halting the in-scope zones, the binary stops the flows that compound or
move stake by deleting their epoch-hook calls, and leaves the unbond, sweep and claim flows
running. The redemption rate is therefore frozen by code at the upgrade height, the redemptions
open at the upgrade finish on their own, and the admin drains every validator as soon as the next
day epoch has submitted those redemptions. Everything downstream (transfers to the Osmosis vault,
pool creation, the token sweep, the halt) is unchanged.

## §3. What the single upgrade contains

From the two-upgrade spec's v35 (§5 there), unchanged: removal of every liquid-stake, redeem and
create-things message across stakeibc, staketia, stakedym, icaoracle, icqoracle, auction, airdrop
and claim; autopilot `StakeibcActive = false`; wasm upload access and the four contract admins to
gov; comdex-1 `Deprecated`; the dYdX trade route deleted; the haqq slash-query purge and
delegation delta table.

From the two-upgrade spec's v36 (§6 there), unchanged: removal of `RebalanceValidators`,
`ClearBalance` and `ResumeHostZone` (stakeibc), `ResumeHostZone` (staketia, stakedym); the
admin gates on `UpdateValidatorSharesExchRate` and `CalibrateDelegation`; the calibration cap
lifted; every rate limit, blacklisted denom and whitelisted pair removed; the three ICA oracles
deactivated; the four admin txs `MsgUndelegateFromValidators`, `MsgTransferFromIca`,
`MsgTransferStaketiaClaimBalance` and `MsgSweepTokensOffStride` with the same constants
(`HostToOsmosisTransferChannel`, `SweepUnwindChannels`, channel-5, the sweep operator, the
Osmosis vault).

Different from both:

- `ClaimUndelegatedTokens` is kept, with its ICA host allow-list entry. It is what pays the
  redemptions that finish after the upgrade. It stays permissionless (it always was) and becomes
  a no-op once the last record is claimed. The ICA host allow-list loses only `MsgLiquidStake`
  and `MsgRedeemStake`.
- No zone is halted. `Halted` stays false on the eleven in-scope zones so that
  `InitiateAllHostZoneUnbondings`, `SweepUnbondedTokensAllHostZones` and the claim keep
  running; they all iterate `GetAllActiveHostZone` or check the flag. Stakedym is not halted
  either: its messages are removed and its operator flushes its records after the upgrade; with
  no liquid stake or redeem left, its remaining flows have nothing to act on. The four
  deprecated zones are untouched, as before.
- The freeze is done in code (§4).
- The handler purges every pending withdrawal-balance ICQ (the reinvest query,
  `ICQCallbackID_WithdrawalHostBalance`) alongside the haqq slash queries, so no callback can
  delegate rewards after the reinvest call is gone.

## §4. Freeze by code instead of by halt

The halt in the two-upgrade spec did two jobs at once: stop the flows that must not run once the
rate is frozen, and stop the flows we now want to keep. So the single upgrade deletes the first
group's calls from `x/stakeibc/keeper/hooks.go` and leaves the second group in place. This is a
deletion of call sites, not a new flag: there is nothing to configure and nothing that can be
toggled back.

Removed from `BeforeEpochStart` (the keeper functions themselves may stay until the follow-up
cleanup; only the calls go):

| Call | Why it must stop |
|---|---|
| `UpdateRedemptionRates` (stride epoch) | The rate must not move once the drain starts: after an admin undelegation, `TotalDelegations` falls with no record behind it and the formula would cut the rate. Deleting the call freezes `HostZone.RedemptionRate` at its last pre-upgrade value. |
| `ReinvestRewards` (stride epoch) | It delegates the withdrawal ICA's rewards back to validators, which would create fresh delegations after the drain and need a second unbonding period. It is also the root of the whole fee machinery: it submits the withdrawal-balance ICQ, whose callback splits the balance, sends the fee cut to the fee ICA and delegates the rest; the reinvest ack then queues the fee-balance ICQ, whose callback moves the fee ICA to the reward collector. Removing this one call stops all of it: no ICQ, no fee split, no delegation, no reward-collector inflow. Rewards simply accumulate in the withdrawal ICA and leave with `MsgTransferFromIca WITHDRAWAL`; what the fee ICA holds at the upgrade leaves with `MsgTransferFromIca FEE`. Cutting the split inside the callback instead would be more code for the same result. |
| `StakeExistingDepositsOnHostZones` (stride epoch) | Same: new delegations after the drain. Deposits that still reach the delegation ICA simply stay there and leave with the ICA balance. |
| `RebalanceAllHostZones` (stride epoch) | Redelegations flag `DelegationChangesInProgress` and would collide with the drain. |
| `TransferAllRewardTokens` (stride epoch) | Already a no-op with the trade route deleted; removed for clarity. |
| `AuctionOffRewardCollectorBalance` (mint epoch) | It liquid stakes the validators' fee share, which mints stTokens against a frozen rate. Validators are paid in STRD only from the upgrade on (§9a of the two-upgrade spec already says the stToken stream ends). Native tokens reaching the reward collector are swept later like any other balance. |

Kept in `BeforeEpochStart`:

| Call | Why it stays |
|---|---|
| `InitiateAllHostZoneUnbondings`, `SubmitPendingUndelegations`, `CleanupEpochUnbondingRecords`, `CreateEpochUnbondingRecord` (day epoch) | Submit and track the redemptions that were open at the upgrade. |
| `SweepUnbondedTokensAllHostZones` (stride epoch) | Moves each completed record's amount from the delegation ICA to the redemption ICA, where the claim pays it. |
| `ClaimAccruedStakingRewards`, `SetWithdrawalAddress` (stride epoch) | Harmless; rewards withdrawn to the withdrawal ICA are swept to Osmosis. |
| `CreateDepositRecordsForEpoch`, `TransferExistingDepositsToHostZones` (stride epoch) | Harmless, and the transfer is useful: it carries any native voucher still in a deposit address to the delegation ICA, which is what the two-upgrade spec's handler step 4 did by hand. |

Also removed: the `UpdateRedemptionRateForHostZone` call at the end of the delegator-shares slash
callback (`icqcallbacks_delegator_shares.go`). The callback keeps correcting the validator's and
zone's delegation on a slash; it no longer rewrites the rate. A slash detected after the upgrade
therefore lowers the backing but not the rate, exactly as a slash during window 2 did in the
two-upgrade design, and the coverage check is where it shows up.

What "frozen" means here: the rate stops moving at the last stride-epoch update before the
upgrade height, instead of at the halt some 35 days later. Between the upgrade and the drain,
delegations keep earning for a few days; those rewards are withdrawn by the undelegation and end
up as pool surplus, which is the same direction as before. Nothing about the transmuter changes:
each pool's factors are still read from `HostZone.RedemptionRate` at pool creation.

## §5. The one ops window

Day 0 is the upgrade height. The redemptions open at that moment (99 user records on
2026-09-18; the number on the day is whatever it is) are in one of three states: queued and not
yet submitted, unbonding on the host, or unbonded and waiting for a sweep or a claim. None of
them needs anything new; the pipeline finishes all three.

1. Day 0: slash refresh on every in-scope zone with `UpdateValidatorSharesExchRate` (admin) and
   the drift measurement (§8a of the two-upgrade spec), exactly as window 2 step 1 did. The
   staketia operator undelegates the entire multisig delegation on Celestia via authz (the
   queued records' amounts and everything else, in one go). The stakedym operator sweeps and
   confirms the unbonded records and undelegates the queued ones.
2. Day 0 to 4: the next day epoch. `InitiateAllHostZoneUnbondings` submits the queued
   redemptions on every zone and flags their validators. Wait for those acks
   (`DelegationChangesInProgress` back to zero on every validator). Doing the drain after this
   point rather than before is the whole sequencing rule: the pipeline then computes its
   amounts from delegations the admin has not touched, and the admin drains what is left with
   no reservation arithmetic. A record whose submission fails (a slash between the refresh and
   the epoch) goes to `RETRY_QUEUE` and is retried at the next day epoch; the drain of that zone
   waits for it.
3. Then, per zone: `MsgUndelegateFromValidators` for one small validator as the live test, then
   with an empty list for the rest. Every validator's remaining delegation is drained. From here
   on no delegation exists that any record needs.
4. Day 0 onward: `MsgTransferFromIca` for the withdrawal and fee ICAs, the live test of the
   transfer tx on small amounts. Not yet the redemption ICA: it holds the tokens of claimable
   records until they are claimed.
5. Day 14 to 34, as each zone's unbondings complete: the pipeline sweeps each record's amount to
   the redemption ICA at the next stride epoch and the record goes `CLAIMABLE`; ops run
   `ClaimUndelegatedTokens` for every record (permissionless, as today). Once a zone has no
   record outside `CLAIMED` and no claim ICA in flight, `MsgTransferFromIca DELEGATION` for the
   full remaining balance, `WITHDRAWAL` again for the auto-withdrawn rewards, and `REDEMPTION`
   for whatever dust the claims left.
6. Staketia, day 21: the operator IBCs the whole unbonded balance via authz to the claim
   address and confirms the sweep for each open record; the hour epoch pays the redeemers from
   the claim address. Then `MsgTransferStaketiaClaimBalance` moves the remainder to the celestia
   delegation ICA, which leaves with that zone's balance in step 5.
7. As each denom's balance is complete on the Osmosis vault: coverage check, pool creation and
   funding, foreign-route denoms (§7 and §9 of the two-upgrade spec, unchanged).
8. Last days: the token sweep in batches, then the halt checklist and the halt (§8 of the
   two-upgrade spec, unchanged).

Roughly: day 0 upgrade, day 4 every undelegation submitted, day 34 the last one complete, day
40 halt.

Checklist to propose the upgrade (there is no "nothing in flight" condition any more):

- The haqq delegation delta table, the drift measurement and the mainnet-export tests match the
  chain at one recent height. A haqq redemption unbonding in flight at the upgrade would change
  the drift; measure right before the proposal and expect the delta helper to skip (never
  error) if it no longer matches.
- The wasm contract list, the ICA channel map, `SweepUnwindChannels`, the sweep operator and
  the Osmosis vault verified as in the two-upgrade spec's checklists.
- The two operators (staketia, stakedym) ready to act on day 0.

Checklist to halt: as in the two-upgrade spec, plus: zero user redemption records, zero
`HostZoneUnbonding` records outside `CLAIMED` with a non-zero amount, no claim ICA in flight,
staketia and stakedym with zero unbonding records outside `CLAIMED`, the redemption ICAs
drained.

## §6. What gets harder, honestly

- One larger upgrade. The review surface is the sum of both plans plus the hook surgery. The
  plans are already written as independent tasks and merge without much rework (§8).
- The drain and the record pipeline share validators and the delegation ICA channel for a few
  days. The sequencing rule in §5 step 2 keeps them apart; the admin tx also refuses any
  validator with a change in progress, so a mistimed drain fails loudly rather than racing.
- The redemption ICA cannot be drained until claims finish, so the transfer tx has one more
  "wait for X" in its runbook (§5 step 5).
- The frozen rate is a few days older than it would have been, because nothing compounds after
  the upgrade. The rewards of those days become surplus; the coverage check is unchanged.
- `ClaimUndelegatedTokens` stays live to the end. It only pays records that exist, so the risk
  is nil, but it is one more message in the final binary.
- If a haqq redemption is unbonding at the upgrade, the delta table measured earlier may not
  match and the reconciliation skips; haqq would then be trued up with the `offset` on the drain
  tx instead. Same tool as before, just more likely to be needed.

## §7. What gets simpler

- No window 1, no second proposal, no second binary, no second mainnet-export fixture.
- No halt handler step, no stakedym halt, no "resume" concerns, no upgrade checklist about
  in-flight records.
- The stranded-voucher bank move is gone: `TransferExistingDepositsToHostZones` keeps doing it
  every epoch.
- The end state is identical to the two-upgrade design: same pools, same sweep, same halt.

## §8. Plan merge

The two existing plans become one v35 plan with these edits, and no new tasks beyond the hook
surgery:

- Upgrade 1 plan: unchanged, except the ICA host allow-list helper keeps
  `MsgClaimUndelegatedTokens`.
- Upgrade 2 plan: the package is `app/upgrades/v35` and its handler steps join the v35
  handler; drop `HaltInScopeHostZones` and `HaltStakedym`; drop the removal of
  `ClaimUndelegatedTokens` and its allow-list step; keep everything else (oracles, rate limits,
  the six... now five handler removals, the ICQ gates, the calibration cap, the four txs, the
  ops scripts, the release gate).
- New task, "freeze by code": delete the six calls in `hooks.go` and the rate rewrite in the
  slash callback; purge pending withdrawal-balance ICQs in the handler; tests that each removed
  call is gone from the hook and each kept call still runs on a non-halted zone; a keeper test
  that `RedemptionRate` is unchanged across a stride epoch with deposit records and rewards
  present; a slash-callback test that the delegation is corrected and the rate is not.
- The undelegate tx no longer requires `Halted`; it requires that no listed validator has a
  change in progress (already the case) and that the zone is not deprecated.
- Localstride dry run: exercise the sequencing rule (submit a redemption, let the day epoch
  unbond it, then drain, then let the sweep and claim complete).

## §9. Decisions taken (brainstorm, 2026-09-28)

1. In-flight redemptions finish through the existing pipeline; no zone is halted, and the
   compounding flows are stopped by deleting their hook calls rather than by a new per-zone
   flag. Deleting is less code, cannot be toggled back, and reads as what it is.
2. The drain runs after the first post-upgrade day epoch has submitted the open redemptions
   and their acks have landed, never before it, so the pipeline and the drain never compete
   and no reservation arithmetic is needed.
3. `AuctionOffRewardCollectorBalance` is deleted at the upgrade: validators are paid in STRD
   only from day 0.
4. The fee cut is not carved out separately: deleting `ReinvestRewards` already stops the fee
   split, the fee ICQ and the reward-collector inflow (§4). Rewards stay in the withdrawal ICA
   until the transfer tx moves them.
5. Stakedym is never halted; its operator flushes after the upgrade and the module idles.
