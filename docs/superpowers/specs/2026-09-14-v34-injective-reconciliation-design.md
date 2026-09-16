# v34: Injective Delegation Reconciliation and Redemption Sweep Unblock

**Date:** 2026-09-14
**Status:** Approved
**Supersedes:** PR #1524 (`moonyandfriends/stride@bcfa27712`), whose reconciliation and callback
hardening are cherry-picked verbatim here; its record re-queue is replaced (§4).

## §1. Problem

Between mid-July and early September 2026 the `injective-1` DELEGATION ICA channel closed seven
times (no relayer on the `icacontroller` path → epoch ICA timeouts closed the ordered channel).
Each closure stranded the acks of `MsgDelegate` packets that had already executed on Injective.
Two consequences:

1. **Ledger drift.** The delegate callbacks never ran, so the host zone never booked those
   delegations. Actual on-chain delegation exceeds tracked delegation by **200.476671 INJ**
   (net; one whole redelegation is also misattributed between zellic/blackpanther/autostake).
2. **Redemption funds staked.** When the channel was restored, `RestoreInterchainAccount` reset
   the in-flight deposit records to `DELEGATION_QUEUE` (`x/stakeibc/keeper/msg_server.go:584-593`)
   and they were re-delegated. The second delegation drew on the delegation ICA's liquid balance —
   which is where completed unbondings sit waiting for the redemption sweep. ~200 INJ of redeemers'
   money is therefore staked instead of liquid.

The sweep (`x/stakeibc/keeper/redemption_sweep.go`) sums `NativeTokenAmount` over every
`EXIT_TRANSFER_QUEUE` record and sends one `MsgSend` for the total. Records total 1,298.908789 INJ;
the ICA holds 1,103.197979 INJ net of pending deposits. The single send fails every stride epoch
and zero records sweep — 11 host-zone unbonding records stuck since epoch 1369.

All figures verified live on 2026-09-14 (Stride 40280619 / Injective 183003315): all 32 per-
validator deltas in the PR's table match chain to 18 decimals; Σ deltas = 200.476671; no drift
since the PR's 2026-09-13 measurement. RR headroom: reconciliation moves RR 1.541302 → ≈1.555664
vs `max_inner_redemption_rate` 1.568710 — inside the band.

## §2. Design summary

Three independent mechanisms, each reusing an existing pipeline:

| Mechanism | Fixes | Where |
|---|---|---|
| Reconcile validator deltas (cherry-picked) | ledger drift | `app/upgrades/v34/injective.go` |
| One-shot undelegation of exactly the applied excess, via the normal undelegate pipeline | funds staked | handler writes a pending amount; day-epoch hook submits it |
| Per-record redemption sweep, scoped to `injective-1` | all-or-nothing sweep | `redemption_sweep.go` |

Outcome after the upgrade: at the first stride epoch, records sweep oldest-first and nine of the
eleven (1,044.65 INJ, everything up to and including 1434 except 1426/1429) succeed; 1426 and 1429
fail-and-retry each epoch. At the first day epoch, 200.476671 INJ is undelegated through the
standard validator-capacity logic. ~21 days later it lands liquid and 1426/1429 sweep. Ledger ends
at exactly the pre-incident `TotalDelegations` (21,514.925033 INJ); RR net change zero; no idle
leftover created; no unbonding record is hand-edited.

**Why not the PR's re-queue?** It re-undelegates three real records (204.21 INJ) to carry the
excess. That over-unbonds by 3.73 INJ, delays three hand-picked users, and creates a record state
(`NativeTokensToUnbond > 0`, `StTokensToBurn = 0`) the undelegate callback's partial-batch branch
(`icacallbacks_undelegate.go:257-270`) does not handle — it would back-compute a positive burn
from the implied rate, drive `StTokensToBurn` negative and burn other users' stINJ. Unreachable on
Injective today (one batch: 32 msgs/tx vs 28 validators with capacity) but only by luck.

**Why not a custom ICA from the upgrade handler?** Feasible but worse: the upgrade runs in
PreBlocker before `BeforeEpochStart` refreshes the epoch tracker, so `SubmitTxsDayEpoch`'s timeout
would be stale (underflows into `ErrTimeoutElapsed` if the halt outlasts the epoch); any error
must be swallowed or the chain halts; and a send in the upgrade block cannot be re-triggered if the
channel is closed or relayers are mid-restart. The day-epoch hook (v17 `DisableHubTokenization`
precedent, commit `4674eed7c`) has none of these problems.

## §3. Reconciliation (cherry-picked, trimmed)

Commit 1 of the branch is `git cherry-pick bcfa27712` — authorship preserved, delta table
byte-identical. Commit 2 trims it:

- Delete `RequeueInjectiveUnbondings`, `RequeuedUnbondingEpochs`, their tests, and the
  `recordskeeper.Keeper` parameter added to `CreateUpgradeHandler` / `app/upgrades.go`.
- `ReconcileInjectiveDelegations(ctx, sk) (appliedDelta sdkmath.Int)` — returns the total delta
  applied, and applies the table all-or-nothing, never as an upgrade error. A missing host zone
  (non-mainnet) returns zero. A validator missing from the host zone, or a delta that would drive
  a delegation negative, means the constants no longer describe chain state: nothing is written,
  zero is returned, an Error is logged, and the upgrade proceeds with the reconciliation deferred
  to a later upgrade. A partial application would return a partial sum — dropping blackpanther's
  −666 INJ entry alone turns the ~200 INJ excess into ~867 INJ — and that sum is what gets
  undelegated. Skipping rather than erroring is deliberate: an error would halt the chain, which
  is disproportionate for an accounting fix, and the per-record sweep still pays the records that
  fit. With the validator list and weights unchanged, the negative check can only trip if acks
  are lost again (tracked and on-chain delegations otherwise move together, and
  `tracked + delta` is the on-chain amount, which is ≥ 0).
- Keep: `InjectiveDelegationDeltas`, `DelegationDelta`, `mustInt`, the redemption-callback
  hardening in `icacallbacks_redemption.go` (only resets records still `EXIT_TRANSFER_IN_PROGRESS`)
  and its test.

Handler (`app/upgrades/v34/upgrades.go`), after `DeleteStuckQueries` and before `UpdateGovParams`:

```
applied := ReconcileInjectiveDelegations(ctx, stakeibcKeeper)
if applied.IsPositive() { stakeibcKeeper.SetPendingUndelegation(ctx, InjectiveChainId, applied) }
```

A zero applied delta (host zone absent, or table not applied) queues nothing.

## §4. Pending undelegation

### State

One new store key in `x/stakeibc/types/keys.go`:
`PendingUndelegationKeyPrefix = "PendingUndelegation-value-"`, keyed by chain id, value =
`sdkmath.Int` (marshalled via `Int.Marshal`). The amount stays stored until each submitted batch
acks successfully, when that batch's amount is subtracted (the key is removed at zero). A second
key `PendingUndelegationInFlight-value-<chainId>` → `uint64` counts the ICA batches awaiting an
ack; while it is non-zero the hook does not resubmit. This is the one-shot marker *and* the channel
that gets the amount from `app/upgrades/v34` into `x/stakeibc/keeper` (which cannot import it —
cycle) without duplicating the number.

Why "delete on success" rather than "delete on submit": the k8s rehearsal showed the delegation
channel dying with the undelegate ICA in flight behind a packet whose timeout can never be
processed once `RestoreInterchainAccount` has zeroed the in-progress counters. A re-queue that
lives only in the callback's failure/timeout path is then unreachable and the amount is lost.
Keeping the amount until success makes recovery a one-line restore change (clear the in-flight
count) with no scanning of the 21k-entry icacallbacks store.

Keeper (`x/stakeibc/keeper/pending_undelegation.go`):
`SetPendingUndelegation(ctx, chainId, amount)`, `GetPendingUndelegation(ctx, chainId) (Int, bool)`,
`RemovePendingUndelegation(ctx, chainId)`, `GetAllPendingUndelegations(ctx) []PendingUndelegation`
(a small struct `{ChainId string; Amount sdkmath.Int}` — no proto; never exported over gRPC or
genesis. If the chain is exported with a key still present, it is simply lost, which is acceptable:
the key exists for at most one day epoch on a healthy channel).

### Hook

`SubmitPendingUndelegations(ctx, epochNumber)` is called in `BeforeEpochStart` inside the
`DAY_EPOCH` block, immediately after `InitiateAllHostZoneUnbondings` (`hooks.go:29`). It is
permanent and store-driven — a no-op when no keys exist — so nothing to remove in v35. For each
pending entry:

1. Load the host zone; if missing, log and delete the key.
2. If `epochNumber % hostZone.GetUnbondingFrequency() == 0` (the host zone's own unbonding
   epoch, the same check as `InitiateAllHostZoneUnbondings`), log and `continue` with the key
   kept: both flows compute capacity from `validator.Delegation`, which is not decremented
   until the ack, so submitting both in one epoch could overshoot the on-chain delegation.
3. Build messages with the normal capacity logic. Refactor `UnbondFromHostZone`
   (`unbonding.go:416-525`): extract the segment from "Determine the total eligible unbond
   amount" through `GetUnbondingICAMessages` into
   `GetUndelegateMessagesForAmount(ctx, hostZone, amount) (msgs, splits, err)`; `UnbondFromHostZone`
   calls it. No behavior change for the normal path.
4. `BatchSubmitUndelegateICAMessages(ctx, hostZone, nil, msgs, splits, batchSize)` — unchanged
   function. With `EpochUnbondingRecordIds = nil` it submits via `SubmitTxsDayEpoch` (timeout is
   correct: the tracker was refreshed at the top of the hook), sets `ICACallbackID_Undelegate`,
   and increments `DelegationChangesInProgress` on each validator. That increment is required:
   `MarkUndelegationAckReceived` decrements it and errors at zero (`validator.go:226-229`), which
   would fail the ack tx and wedge the ordered delegation channel.
5. Success → `SetPendingUndelegationInFlight(numTxs)`, `EmitUndelegationEvent`; the amount stays.
   A host zone with batches in flight is skipped. Any error (channel closed, insufficient capacity,
   ICA submit failure) → `Logger.Error`, key kept, retried next day epoch. Never returns an error to
   the hook; never panics.

### Callback

Existing `UndelegateCallback` handles nil record ids with no change (verified against
`icacallbacks_undelegate.go`): success → `UpdateDelegationBalances` decrements each validator and
`TotalDelegations`; `UpdateHostZoneUnbondingsAfterUndelegation` loops zero times and returns zero;
`BurnStTokensAfterUndelegation(0)` is a clean no-op (`sdk.NewCoins` drops zero coins);
`MarkUndelegationAckReceived` decrements in-progress; failure/timeout paths touch no records.
Completion time is parsed from the ack's `MsgUndelegateResponse`, not from records.

With no record ids: success → `ConsumePendingUndelegation(chainId, batchAmount)` (subtract, remove
at zero) and `DecrementPendingUndelegationInFlight`; failure / timeout → only
`DecrementPendingUndelegationInFlight`, leaving the amount to be resubmitted at the next eligible
day epoch. `RestoreInterchainAccount` (delegation branch) calls `RemovePendingUndelegationInFlight`
so a batch stranded on the dead channel — whose ack can never arrive and whose timeout may never be
processable — is released for resubmission on the new channel.

Inherent limit (all designs): a batch that executed on the host but whose success ack was
stranded is resubmitted and undelegates twice; the slash query then sees on-chain < tracked and
corrects tracked. Window: minutes, same as the incident's own mechanism.

Interaction with the normal flow on the same day epoch: the hook skips a host zone on its own
unbonding epoch (step 2 above), so the two flows never cascade onto the same validators within
one epoch.

## §5. Per-record redemption sweep (scoped)

`x/stakeibc/types`: `const PerRecordSweepChainId = "injective-1"` with
`// TODO [cleanup]: remove after v34 — revert injective-1 to the bundled sweep`.

`SweepUnbondedTokensForHostZone` (`redemption_sweep.go:58-123`): after
`GetTotalRedemptionSweepAmountAndRecordIds`, extract the submit into
`submitRedemptionSweep(ctx, hostZone, amount, epochUnbondingRecordIds) error` (builds the
`MsgSend`, marshals the callback, `SubmitTxsStrideEpoch`, sets those records
`EXIT_TRANSFER_IN_PROGRESS`, emits the sweep event). Then:

- `chainId != PerRecordSweepChainId` → one call with the total and all ids (bundled, unchanged).
- `chainId == PerRecordSweepChainId` → sort ids ascending; for each id, one call with that
  record's `NativeTokenAmount` and `[]uint64{id}`. A submit error aborts the loop and is returned
  (records already submitted stay IN_PROGRESS with their own callbacks; the rest remain queued).

Per-record ordering is oldest-first so the fairness of who waits is deterministic. On the ordered
channel a bank failure returns an error ack and the next packet still executes; only timeouts
close the channel, so this adds no closure risk. The cherry-picked callback hardening is kept: each
callback now carries one id, but a record could still be moved by another path while in flight.

## §6. Tests

- **Cherry-picked**: `injective_test.go` reconciliation cases (adjusted for the new return value —
  assert the returned applied delta equals the sum over seeded validators, and equals zero when
  the host zone is missing); `TestRedemptionCallback_FailureOnlyResetsInProgressRecords`. Requeue
  tests deleted.
- **Upgrade** (`upgrades_test.go`): with the injective host zone seeded, the pending key holds the
  applied delta after the handler; without it, no key.
- **Pending undelegation** (`pending_undelegation_test.go`): set/get/remove round-trip and
  `GetAll`; hook with an ICA channel (`CreateICAChannel`, `CheckICATxSubmitted`) submits, removes
  the key, increments `DelegationChangesInProgress` on the chosen validator(s); with no channel the
  submit fails, key kept, no panic; no keys → no ICA (`CheckICATxNotSubmitted`); missing host zone
  → key removed; on the host zone's unbonding epoch → no ICA, key kept, and the next epoch
  submits.
- **Undelegate refactor**: existing `UnbondFromHostZone` tests pass unchanged.
- **Callback** (`icacallbacks_undelegate_test.go`): `TestUndelegateCallback_NoRecords` — success
  decrements validator delegation and `TotalDelegations`, burns nothing, decrements in-progress,
  leaves no pending key; failure and timeout return nil and re-queue the batch amount into the
  pending store (added to any existing pending amount).
- **Sweep** (`redemption_sweep_test.go`): existing bundled tests unchanged; new per-record case
  using `PerRecordSweepChainId` — N records → sequence advances by N, every record IN_PROGRESS,
  and an ordering assertion that submissions are ascending by epoch.
- **Post-build check** (not a unit test): `git diff bcfa27712 -- app/upgrades/v34/injective.go`
  shows only the deletions from §3; the delta table is untouched.

No dockernet rehearsal (decided 2026-09-14).

## §7. Release notes for the proposal

- Re-measure the delta table against `injective-1` immediately before the upgrade proposal
  (`x/staking` delegations of `inj16ujqtje2en9ns59hrcjfu2885epsrvus9czdw8ljtd45whxucs5srrejpa`
  vs host zone validators). It has been stable across 2026-09-08 / 09-13 / 09-14.
- The delegation ICA channel must be open for the hook to submit; if it is closed at the first
  day epoch, the hook retries every day epoch until it succeeds — no action needed beyond the
  usual channel restoration. The hook also defers the submission on Injective's own unbonding
  epoch (every 4th day epoch), so the reconciliation may land one day later than the upgrade.
  If the submitted ICA fails or times out, the callback re-queues the amount and the hook
  resubmits it at the next eligible day epoch.
- **Upgrade timing.** Between the upgrade block and the undelegate ack, `TotalDelegations`
  carries the +200.48 INJ and the RR reads ~0.93% high (≈ +0.0144). The RR updates every stride
  epoch (01/07/13/19 UTC) and the undelegation goes out at the first non-unbonding day boundary
  (19:00 UTC, day epochs with `epoch % 4 != 0`; day 1472 started 2026-09-14 19:00 and is an
  unbonding day). Scheduling the upgrade 1–3 hours before such a boundary limits the inflated RR
  to a single epoch. Target: 2026-09-23 15:00–16:00 UTC (day 1481 boundary at 19:00; the
  Sep 22 boundary, day 1480, is an unbonding day). The RR is deliberately not adjusted for the
  pending amount — the exposure is one epoch and the value stays inside the bounds.
- **Inner bounds.** A bounds breach halts the host zone (`abci.go`), which would also stop the
  sweep and unbondings. Projected RR on 2026-09-23 ≈ 1.5432 → ≈ 1.5576 after reconciliation vs
  `max_inner_redemption_rate` 1.56871 today. The inner bounds are set manually; confirm
  `max_inner_redemption_rate ≥ 1.575` on upgrade day (or raise it to 1.60 the day before and
  restore it after the ack).
- v35 cleanup: remove `PerRecordSweepChainId` and the per-record branch.

---

## Build plan

### Chunk A — cherry-pick and trim
**Depends on:** B (the handler calls `SetPendingUndelegation`)

- `git cherry-pick bcfa27712` (commit 1). Resolve any conflict in `app/upgrades.go` /
  `app/upgrades/v34/upgrades.go` against main's current handler signature (stakeibc, icq, gov
  keepers); the delta table must not be touched.
- Commit 2: delete `RequeueInjectiveUnbondings`, `RequeuedUnbondingEpochs`, their tests, the
  `recordskeeper` param and import; `ReconcileInjectiveDelegations` returns `(sdkmath.Int, error)`;
  handler stores the pending amount (§3). Update `CHANGELOG.md` line to describe the new mechanism.
- Interfaces consumed: `SetPendingUndelegation` / `GetPendingUndelegation` from chunk B.
- Tests: §6 cherry-picked + upgrade cases.

### Chunk B — pending undelegation (state, hook, refactor)
**Depends on:** none

- `x/stakeibc/types/keys.go`: prefix. `x/stakeibc/keeper/pending_undelegation.go`: keeper
  functions + `SubmitPendingUndelegations`. `unbonding.go`: extract
  `GetUndelegateMessagesForAmount`. `hooks.go`: one call in the DAY_EPOCH block.
- Key decisions: nil record ids into the existing batch submitter; errors logged and retried;
  key deleted only after a successful submit.
- Tests: §6 pending-undelegation, refactor, and callback cases.

### Chunk C — per-record sweep
**Depends on:** none

- `x/stakeibc/types`: `PerRecordSweepChainId`. `redemption_sweep.go`: `submitRedemptionSweep`
  helper + scoped per-record loop with explicit ascending sort.
- Tests: §6 sweep cases.

Order: B and C in parallel, then A (A's handler change calls B's keeper function).
Reviewer note: money-moving diff — review B and the callback interaction at the most capable tier.
