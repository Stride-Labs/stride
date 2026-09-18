# v34: Celestia and Cosmos Hub Stranded-Ack Reconciliation, Staketia Balance Correction

Status: approved design, pre-build. Companion to
`2026-09-14-v34-injective-reconciliation-design.md`; extends the same v34 handler.

## §1. Problem

The Injective spec describes the root cause: an ordered DELEGATION ICA channel closes on an
ICA timeout, stranding the acknowledgements of packets that had already executed on the host.
Their success callbacks never run, so Stride's tracked validator delegations understate what
is actually staked. The 2026-09-17 audit (this session) found the same drift on two more hosts,
in two further variants, plus one unrelated staketia bookkeeping error.

**1a. Celestia, deposit-record variant.** The delegation ICA channel died ~90 times between
June and September 2026 (an automated restore at 01:12/14:12 UTC re-sent every open record
into a channel the ops relayer could not clear; see the memory note
`celestia-delegation-restore-loop`). Restoring a channel flips every DELEGATION_IN_PROGRESS
deposit record back to DELEGATION_QUEUE, so a record whose delegate had already executed was
re-sent and executed *again* with other records' liquid. The liquid pool drained into
untracked validator stake while every record stayed open. After the channel was cleared by
hand on 2026-09-17:

| | utia |
|---|---|
| ICA stake on Celestia (89 validators) | 587,357,974,464 |
| Stride tracked validator sum | 571,918,115,568 |
| Untracked ("phantom") stake | **15,439,858,896** |
| Open deposit records with no liquid behind them (296 records) | 15,460,354,745 |
| ICA liquid | 131,006 |

The phantom records can never complete: their coins are already staked. Every stride epoch
re-sends them, each fails with insufficient funds (error ack, record back to queue), and if
the relayer cannot land ~300 heavy packets before the ICA timeout the channel closes again.

**1c. Cosmos Hub, LSM-deposit variant.** One LSM detokenize for stakewithus executed on the
Hub (+10,999,999 uatom on-chain vs tracked, measured to the uatom) but Stride holds that
deposit in DETOKENIZATION_FAILED (the retry after a restore failed because the shares were
already redeemed). The delegation is real and unbooked; the record will be retried forever.

**4. Staketia `remaining_delegated_balance`.** staketia thinks the multisig still holds
196,907,943,420 utia delegated; on-chain it holds 156,831,200,577 (delta −40,076,742,843).
stakeibc `total_delegations` is *correct* (it implies the multisig figure within 41 TIA).
Redemptions route to the multisig until the staketia number reaches zero, so it would run
dry ~40k TIA early and stakeibc redemptions would never auto-enable. The existing
`MsgAdjustDelegatedBalance` moves both numbers, so it cannot be used.

Root cause (traced 2026-09-18): a v25 migration straddle. Before v25 (2025-02-06 15:00 UTC,
gov prop 260) `RedeemStake` did not touch the delegated balance and `ConfirmUndelegation`
decremented it. After v25, `RedeemStake` decrements staketia's `remaining_delegated_balance`
at redeem time and `ConfirmUndelegation` decrements stakeibc's `total_delegations` instead.
Unbonding record 884 accumulated redemptions from 2025-02-03 to 02-07 (mostly under the old
code, so no staketia decrement) and was confirmed on 2025-02-13 under the new code (stakeibc
decrement only). Its native amount, 39,829,976,251 utia, is 99.4% of the gap. The residual
246,766,592 utia is 0.02% of all volume undelegated since and matches a second, ongoing
effect: staketia decrements by the redeem-time estimate while stakeibc decrements by the
amount finalized up to four days later at a slightly higher rate. That drift is left as-is
(tabled with item 2); it is ~0.02-0.05% of each redemption.

The redemption rate is unaffected by all three today: stakeibc counts queued/in-progress
deposit records as undelegated backing and failed LSM deposits as tokenized delegations, so
the phantom amounts are counted once, in the wrong bucket. Every fix below is a bucket move
and must leave the rate unchanged to the utia.

## §2. Design summary

Three independent additions to `app/upgrades/v34`, each following the v34 convention
(**apply all-or-nothing; never return an error from an accounting fix; log loudly and skip
when constants no longer describe chain state**):

| Piece | File | Effect |
|---|---|---|
| 1a Celestia | `celestia.go` | true up validators by a per-address delta table, then remove exactly the same total from Celestia deposit records |
| 1c Cosmos Hub | `cosmoshub.go` | replay the detokenize success path for one LSM deposit |
| 4 Staketia | `celestia.go` | apply a signed delta to staketia `remaining_delegated_balance` only |

The shared "apply a delta table to a host zone" loop is lifted out of
`ReconcileInjectiveDelegations` into `reconcileHostZoneDelegations(ctx, sk, chainId, deltas)`
so Celestia reuses it byte-for-byte; the Injective wrapper keeps its signature and tests.

## §3. Celestia (1a)

### Constants (`celestia.go`)

- `CelestiaChainId = "celestia"`.
- `CelestiaDelegationDeltas []DelegationDelta`: measured 2026-09-18 at Stride height 40362843
  / Celestia height 14149694 by diffing `x/staking` delegations of the delegation ICA
  `celestia1rdfn69mf3xey2jlqyp70vljtx6df4lkydndplz9drdueuhqh8geqk9gjkj` against the host zone
  validators (89 non-zero entries, all positive, sum 15,439,858,896 utia). **Re-measure right
  before the proposal**; the file comment carries the measurement commands.
- The phantom amount is **not** a separate constant: `phantomAmount = Σ deltas`. Using one
  number for both halves is what guarantees the rate does not move.

Why the record set is not a constant: phantom records churn. Each epoch the new daily deposit
brings real liquid, the oldest small phantom is sent first and succeeds against that liquid,
and the legitimate record then fails and becomes the phantom. The *amount* is invariant
(it equals the on-chain minus tracked gap); the IDs are not. Queue records are fungible claims
on one liquid pool, so removing any set summing to the phantom amount leaves the survivors
equal to the real liquid.

### `ReconcileCelestia(ctx, sk, rk, ick) (applied bool)`

Order matters; nothing is written until every check passes.

1. Host zone `celestia` present, else log + return false (non-mainnet).
2. Validate the delta table against the host zone exactly as Injective does (every address
   present, no delegation driven negative). Failure → log + return false, nothing written.
3. Collect removable records: Celestia deposit records in `DELEGATION_QUEUE` **and**
   `DELEGATION_IN_PROGRESS`, sorted by id ascending, queue records first, then in-progress.
   `TRANSFER_QUEUE` records are never touched (their coins are still on Stride).
   If `Σ amounts < phantomAmount` → log + return false, nothing written.
4. Apply the validator deltas via `reconcileHostZoneDelegations` (this also raises
   `TotalDelegations` by the sum).
5. Remove `phantomAmount` from the collected records in order: delete whole records while
   `remaining ≥ record.Amount`; for the first record that would overshoot, set
   `record.Amount -= remaining` and stop. The loop therefore removes exactly `phantomAmount`.
6. For every **deleted** record that was `DELEGATION_IN_PROGRESS`: its delegate packet is in
   flight on the current DELEGATION channel and its ack (error, or success if liquid arrived)
   will call `DelegateCallback`, which errors on a missing record and would fail the relayer's
   ack tx and leave the packet commitment forever. So, for each icacallbacks entry on port
   `icacontroller-celestia.DELEGATION` whose `DelegateCallback.DepositRecordId` is a deleted
   record: (a) if the entry is on the host zone's active DELEGATION channel
   (`ICAControllerKeeper.GetActiveChannelID`), decrement `DelegationChangesInProgress` for each
   validator in its `SplitDelegations` (guarding at zero); (b) remove the entry regardless of
   channel (orphans on dead channels are just deleted). The ack then hits "callback data not
   found", which icacallbacks treats as a no-op. An in-progress record is never shrunk in
   place: its in-flight callback carries the original split amounts, and a success ack would
   subtract more than the shrunk record holds. So when the overshoot record is in-progress,
   delete it (with the callback cleanup above) and append a fresh `DELEGATION_QUEUE` record
   for the leftover `record.Amount - remaining` (same host zone, denom, source, epoch), the
   way v33 credited Osmosis. Removal therefore stays exact in every case.
7. Log per-record removals and the totals; return true.

Idempotence: a second run finds the delta table would apply again (validators exist) but the
records would no longer cover the amount, so step 3 skips before anything is written.
`TestUpgrade` runs the handler once; an explicit idempotence test runs `ReconcileCelestia`
twice and asserts the second call is a no-op.

## §4. Staketia (4)

`AdjustStaketiaRemainingDelegatedBalance(ctx, stk, delta)` with
`StaketiaRemainingDelegatedBalanceDelta = -40,076,742,843` (utia; measured 2026-09-18).
Measurement rule, to re-run right before the proposal:
`delta = (multisig on-chain delegation − native amount owed by the current
ACCUMULATING_REDEMPTIONS and UNBONDING_QUEUE records) − remaining_delegated_balance`. The
subtraction matters because `remaining_delegated_balance` is already reduced for redemptions
the multisig has not undelegated yet; on 2026-09-18 that term was zero. Applies `hostZone.RemainingDelegatedBalance += delta` on the **staketia** host zone
only. Skips with a log if the staketia host zone is absent or the result would be negative.
It deliberately does not mirror to stakeibc (contrast `MsgAdjustDelegatedBalance`).

A delta rather than an absolute set because user redemptions decrement the live number
between measurement and upgrade; the delta stays valid as long as no `MsgAdjustDelegatedBalance`
or `ConfirmDelegation` lands in between (both would be visible in the pre-proposal re-measure).

## §5. Cosmos Hub (1c)

Constants (`cosmoshub.go`): `CosmosHubChainId = "cosmoshub-4"` and one
`CosmosHubStrandedLsmDeposit{Denom, ValidatorAddress, Amount}` for the stakewithus deposit
(`cosmosvaloper1jlr62guqwrwkdt4m3y00zh2rrsamhjf9num5xr`, 10,999,999 uatom; denom copied from
the live record at proposal time).

`CloseCosmosHubLsmDeposit(ctx, sk, rk) (applied bool)`: requires the host zone, the LSM
deposit (by chain id + denom) in status `DETOKENIZATION_FAILED` with exactly the constant
amount and validator address, and the validator present in the host zone. Then replays the
detokenize success path: `RemoveLSMTokenDeposit` and `AddDelegationToValidator(amount)`
(which also raises `TotalDelegations`), `SetHostZone`. Any mismatch → log + skip. No pending
undelegation: the coins belong in the validator.

## §6. Wiring

`CreateUpgradeHandler` gains `recordsKeeper recordskeeper.Keeper`,
`icacallbacksKeeper icacallbackskeeper.Keeper` and `staketiaKeeper staketiakeeper.Keeper`;
`app/upgrades.go` passes them. The three calls slot in after the Injective step, in the order
Celestia → Staketia → Cosmos Hub; each is independent. The opening log line lists the new
steps.

## §7. Tests

Unit (`celestia_test.go`, `cosmoshub_test.go`, extending `upgrades_test.go`'s suite):

- Celestia applies in full: every validator moved by its delta, `TotalDelegations` up by the
  sum, exactly `Σ deltas` removed from records (whole deletes + one shrink), queue records
  consumed before in-progress, redemption-rate components unchanged
  (`GetUndelegatedBalance + TotalDelegations` before == after).
- Celestia skips entirely (no writes) when: host zone missing; a table validator missing; a
  delta would drive a delegation negative; records cannot cover the amount.
- Celestia deleted in-progress record: its callback entries removed on all channels,
  `DelegationChangesInProgress` decremented only for the active-channel entry, never below 0.
- Celestia shrink never targets an in-progress record.
- Celestia idempotent: second call is a no-op.
- Staketia: applies; skips on missing host zone; skips when result negative; stakeibc host
  zone untouched.
- Cosmos Hub: applies (record gone, validator and total up); skips on wrong status, wrong
  amount, wrong validator, missing record, missing validator, missing host zone.
- `TestUpgrade`: seeds all three states alongside the existing POA/Injective/gov seeds and
  asserts the combined result after `ConfirmUpgradeSucceeded`.

Mainnet export gate (`mainnet_export_test.go`, `testdata/README.md`): extend the trimmed
fixture with `stakeibc.host_zone_list` entries for `celestia` and `cosmoshub-4`,
`records.deposit_record_list` (celestia records only), `records.lsm_token_deposit_list`
(cosmoshub-4 only) and `staketia.host_zone`; assert each new piece applied in full with the
real constants (so a stale table skips loudly instead of silently). Fixture regeneration
commands added to the README; the fixture is refreshed at release prep together with the
constants.

## §8. Release notes for the proposal

- Books 15,440 TIA of Celestia stake that executed on-chain without acknowledgement and
  retires the equivalent phantom deposit records; stTIA redemption rate unchanged.
- Books one 11 ATOM LSM detokenization on Cosmos Hub and clears its failed record; rate
  unchanged.
- Corrects staketia's multisig delegated-balance bookkeeping by −40,077 TIA so redemptions
  hand over to the ICA at the right time; no user-facing effect until then.
- Operational prerequisites called out separately: relayer `max_msg_num` for Celestia,
  disabling the blind auto-restore job.

## Build plan

### Chunk A — shared helper, Celestia, Staketia
Depends on: none.
- Files: `app/upgrades/v34/injective.go` (extract `reconcileHostZoneDelegations`; keep
  `ReconcileInjectiveDelegations` behavior and tests green), new `app/upgrades/v34/celestia.go`
  (constants, `ReconcileCelestia`, `AdjustStaketiaRemainingDelegatedBalance`), new
  `celestia_test.go`, `upgrades.go` + `app/upgrades.go` wiring for records/icacallbacks/
  staketia keepers, `upgrades_test.go` `TestUpgrade` seeds and asserts for Celestia + staketia.
- Interfaces produced: `reconcileHostZoneDelegations(ctx, sk, chainId, []DelegationDelta) (sdkmath.Int, bool)`;
  `ReconcileCelestia(ctx, sk, rk, ick) bool`; `AdjustStaketiaRemainingDelegatedBalance(ctx, stk, delta) bool`.
- Key decisions: single constant for both halves; queue-before-in-progress ordering; shrink
  only queue records; callback cleanup only for deleted in-progress records.
- Tests: the Celestia and Staketia bullets in §7.

### Chunk B — Cosmos Hub
Depends on: A (wiring of `recordsKeeper` into the handler; the `TestUpgrade` seeding pattern).
- Files: new `cosmoshub.go`, `cosmoshub_test.go`, one call + log line in `upgrades.go`,
  `TestUpgrade` seed/assert.
- Interfaces produced: `CloseCosmosHubLsmDeposit(ctx, sk, rk) bool`.
- Tests: the Cosmos Hub bullets in §7.

### Chunk C — mainnet export gate
Depends on: A, B.
- Files: `testdata/README.md`, `testdata/mainnet_export.json.gz` (regenerated with the extra
  sections), `mainnet_export_test.go` (populate + assert for celestia, cosmoshub-4, staketia).
- Key decision: the fixture is assembled from the public REST API at a recorded height, same
  provenance model as today; constants and fixture are refreshed together at release prep.
- Tests: `TestUpgradeFromMainnetExport` asserts all five reconciliations (POA, Injective,
  Celestia, Staketia, Cosmos Hub) applied in full.
