# Protocol Wind-Down: Withdrawal-Only Mode

Status: approved design, pre-plan. Large tier; the implementation plan follows in
`docs/superpowers/plans/`.

## §1. Goal

Wind Stride down into a withdrawal-only protocol. Stop every flow that moves funds on behalf of
users (liquid staking, reinvesting, rebalancing, redemption-rate updates, fee liquid staking),
unbond every delegation, bring the native tokens back to Stride, and let stToken holders swap
their stTokens for the backing native tokens directly on Stride at a fixed rate. The guiding
constraints are, in order: least risk of a bug that loses funds, least new code, most reuse of
code that already runs on mainnet, and a chain that is inert while it waits.

Dormant capital is acceptable. If rounding, slashing drift or the 99.9% unbond leave a residue
that no stToken can claim, it stays in the pool. What is not acceptable is a pool that cannot
cover a valid redemption or claim.

## §2. Scope

In scope:

- Every non-halted stakeibc host zone: celestia, comdex-1, cosmoshub-4, dydx-mainnet-1,
  haqq_11235-1, injective-1, juno-1, laozi-mainnet, osmosis-1, phoenix-1, sommelier-3, ssc-1.
- Staketia (stTIA redemptions and the 5-of-7 Celestia multisig), including its 7 in-flight
  unbonding records.
- The 99 unclaimed stakeibc user redemption records (as of 2026-09-18).

Out of scope, explicitly:

- The already-halted zones evmos_9001-2, stargaze-1 and umee-1. stEVMOS, stSTARS and stUMEE
  holders are unaffected by this work.
- Stakedym. stDYM holders are unaffected.
- The IBC rate limits. They stay as they are (they only cover stTokens; no native denom is
  limited).
- STRD-side modules (mint, auction, strdburner, airdrop, claim).

## §3. Facts the design rests on

Verified against mainnet on 2026-09-18 (REST `stride-api.polkachu.com`, Celestia REST
`celestia.rpc.uquad.org`).

- `HostZone.Halted` already gates every stakeibc flow we want stopped: `LiquidStake`,
  `RedeemStake`, `ClaimUndelegatedTokens`, LSM, autopilot (via `LiquidStake`), and every epoch
  hook loop (`ReinvestRewards`, delegation, `UpdateRedemptionRates`, rebalance,
  `InitiateAllHostZoneUnbondings`, `SweepUnbondedTokensAllHostZones`, reward-collector fee
  liquid staking). All iterate `GetAllActiveHostZone`.
- The v34 pending-undelegation pipeline (`x/stakeibc/keeper/pending_undelegation.go`) queues
  an amount per host zone, submits it at the day epoch through the normal validator-capacity
  logic (`GetUndelegateMessagesForAmount` + `BatchSubmitUndelegateICAMessages`), tracks
  in-flight batches, and retries any failed batch the next day. It reads the host zone with
  `GetHostZone`, so it runs on a halted zone. Its undelegate callback decrements
  `TotalDelegations` and validator balances and burns nothing.
- An ICA-wrapped IBC `MsgTransfer` from a host account back to Stride already exists in the
  trade-route code (`x/stakeibc/keeper/reward_converter.go`, `BuildHostToTradeTransferMsg` and
  `TransferConvertedTokensTradeToHost`). The redemption sweep is a host-internal bank `MsgSend`
  from the delegation ICA to the redemption ICA (`submitRedemptionSweep`), not a cross-chain
  transfer.
- Rate-limit whitelisting of a (sender, receiver) pair is done with
  `RatelimitKeeper.SetWhitelistedAddressPair`, as for the deposit → delegation ICA path today
  (`x/stakeibc/keeper/ibc.go`).
- `UserRedemptionRecord` stores `Receiver` (a host-chain address), `NativeTokenAmount`,
  `StTokenAmount`, `EpochNumber`, `HostZoneId`, `ClaimIsPending`. It does **not** store a Stride
  address. All 99 open records have a non-zero `StTokenAmount`; 3 have `ClaimIsPending`.
- Claim today: the unbonded tokens sit in the host zone's redemption ICA on the host;
  `ClaimUndelegatedTokens` submits an ICA bank `MsgSend` from the redemption ICA to
  `Receiver` for `NativeTokenAmount`, sets `ClaimIsPending`, and the ICA callback deletes the
  record on ack. It requires the host zone to be active and the epoch's `HostZoneUnbonding`
  to be `CLAIMABLE`, and decrements `ClaimableNativeTokens`.
- Autopilot redemptions set the stakeibc `Creator` to the IBC packet's Stride receiver, which
  is usually a derived address the user cannot sign for. This rules out paying claims to a
  reconstructed Stride sender.
- Staketia records store the Stride `Redeemer`; staketia claims pay on Stride from the claim
  address, automatically, in the epoch hook (`DistributeClaims`). The full flow is: redeem
  escrows stTIA → epoch `PrepareUndelegation` freezes the record → operator undelegates on
  Celestia and calls `MsgConfirmUndelegation` (burns stTIA) → after 21 days
  `MarkFinishedUnbondings` → operator IBC-transfers to the claim address and calls
  `MsgConfirmUnbondedTokenSweep` (checks the claim-address balance covers the record) →
  next epoch `DistributeClaims`. `RemainingDelegatedBalance` is read only by staketia
  `RedeemStake`, so zeroing it blocks new redemptions without affecting this flow.
- The Celestia multisig (`celestia1d6ntc7s8gs86tpdyn422vsqc6uaz9cejnxz5p5`, 5-of-7 legacy
  amino multisig) has granted the operator key
  (`celestia1ghhu67ttgmxrsyxljfl2tysyayswklvxzls400`) five unexpiring authz grants:
  `StakeAuthorization` delegate, `StakeAuthorization` undelegate, generic
  `MsgWithdrawDelegatorReward`, generic `MsgCancelUnbondingDelegation`, and an IBC
  `TransferAuthorization` on `transfer/channel-4` with an effectively unlimited utia spend
  limit and a single allow-listed receiver: staketia's claim address
  `stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd`. A bank `MsgSend` is **not** granted, so the
  multisig cannot fund the Celestia delegation ICA without a 5-of-7 signing; it can move
  everything to Stride's claim address with the operator key alone. Staketia's IBC middleware
  only hooks outbound acks and timeouts, so an unrecorded inbound transfer to the claim address
  is inert.
- Unbonding periods: osmosis-1 14 days, dydx-mainnet-1 30 days, juno-1 and sommelier-3 28
  days, everything else 21 days. Upgrade 2 cannot land before ~day 32.
- In-flight stakeibc records on 2026-09-18 (recent 40 epochs): queued-but-not-unbonded
  amounts on osmosis-1 (~8.7k OSMO), phoenix-1 (~12.3k LUNA), ssc-1, laozi-mainnet, and a
  cosmoshub-4 retry (~54.7k ATOM); `EXIT_TRANSFER_QUEUE` (unbonded, not yet swept) on
  cosmoshub-4, dydx-mainnet-1, injective-1, juno-1, osmosis-1, phoenix-1, ssc-1.
- Staketia: 6 `UNBONDING_IN_PROGRESS` records, 1 `UNBONDING_QUEUE` record (~717 stTIA /
  ~843 TIA), 1 accumulating. `remaining_delegated_balance` is 196,064 TIA, known to overstate
  the real multisig delegation (~157k TIA); v34 corrects it, and this design never relies on it.

## §4. Design overview

Two upgrades, with an ops window of about 32 days between them.

| | Upgrade 1 (freeze + unbond) | Ops window | Upgrade 2 (withdrawal mode) |
|---|---|---|---|
| New code | Two admin txs, rate-limit whitelist | none | Pool account, one new redeem tx, one relaxed check in claim, handler with pinned constants |
| Moves funds | No (queues undelegations) | Admin ICA transfers, staketia operator | Bank moves on Stride only |
| Reused | `Halted`, pending-undelegation pipeline | trade-route ICA transfer, sweep `MsgSend`, staketia flow | existing claim path, burn/send helpers |

The split is deliberate: upgrade 1 is almost entirely existing code so it can ship quickly and
start the unbonding clock, and the ICA transfer tx it introduces gets exercised on small real
balances during the window before it is used for the full amounts.

## §5. Upgrade 1

Handler:

1. Set `Halted = true` on every in-scope stakeibc host zone. Do not touch the three
   already-halted zones.
2. Set staketia `HostZone.RemainingDelegatedBalance = 0`. Staketia `RedeemStake` now returns
   `ErrRedemptionsDisabled`; its confirm/distribute flow is unaffected.
3. For each in-scope zone, whitelist in the rate limiter the pairs
   (delegation ICA → pool), (withdrawal ICA → pool), (fee ICA → pool). The pool address is the
   new module account from §7 (its address is deterministic, so the whitelist can be set before
   the account exists). Native denoms are not rate-limited today, so this is belt-and-braces
   in case a native limit is ever added.

Note on `Halted`: the existing `MsgResumeHostZone` admin message can un-halt a zone, and the
redemption-rate safety check in `BeginBlocker` keeps running against the frozen rate (a no-op
since the rate no longer changes). No new flag is introduced; reusing `Halted` is the least code
and matches the "chain is inert" goal.

New admin tx `MsgSetPendingUndelegation { creator, chain_id, amount }`, gated by
`utils.ValidateAdminAddress`:

- Calls `SetPendingUndelegation(chainId, amount)`; a zero amount removes the key. Nothing else.
  The v34 pipeline does the submission, retry and in-flight tracking at the day epoch.
- Ops pass ~99.9% of `TotalDelegations` per zone. ICA txs are atomic, so a single
  `MsgUndelegate` that exceeds a validator's true on-chain delegation fails the whole batch;
  recorded per-validator delegations drift from the chain (the v34 reconciliations are the
  proof), and the 0.1% margin absorbs that. The residue is dormant capital under §1. The tx
  is also the override lever if a zone's batch keeps failing (lower the amount, retry).
- The pipeline's "defer if the zone unbonds this epoch" check keys on the epoch number only,
  so a halted zone may wait one extra day on some epochs. Acceptable.

New admin tx `MsgTransferFromIca { creator, chain_id, ica_type, amount, destination }`,
gated by `utils.ValidateAdminAddress`:

- `ica_type ∈ {DELEGATION, WITHDRAWAL, FEE}`; `destination ∈ {POOL, REDEMPTION_ICA}`.
- `POOL`: submits one ICA containing an IBC `MsgTransfer` of `amount` of the host denom from
  the chosen ICA over the zone's transfer channel to the pool address, built the same way as
  `BuildHostToTradeTransferMsg`. Timeout per the existing ICA timeout param. No callback state
  is needed: a timed-out or failed transfer leaves the funds in the ICA and ops resubmit.
- `REDEMPTION_ICA`: submits the existing sweep bank `MsgSend` (delegation ICA → redemption
  ICA) for `amount`. Only valid with `ica_type = DELEGATION`. No record bookkeeping.
- The tx does not read or write host zone accounting. Amounts are an ops input, verified
  off-chain against ICA balances.

Not in upgrade 1: any change to redemption records, the rate, or the pool.

## §6. Ops window

Stakeibc, per zone:

1. Day 0: `MsgSetPendingUndelegation` for each zone at 99.9% of `TotalDelegations`. Read
   `TotalDelegations` only after any undelegate ICA that was in flight at the upgrade has
   acked (its callback decrements the total); an overshoot is caught by the capacity check and
   simply never submits until the amount is lowered. Watch the day-epoch submissions and acks. Delegation ICA channels and their relayers must stay healthy
   until every batch has acked; `RestoreInterchainAccount` clears the in-flight count if a
   channel dies, and the pipeline resubmits.
2. Day 0+: `MsgTransferFromIca` for residual withdrawal and fee ICA balances to `POOL`. This
   is the live test of the transfer tx on small amounts.
3. As each zone's unbonding completes (day 14 to day 30): compute the zone's reserve (§8),
   `MsgTransferFromIca DELEGATION → REDEMPTION_ICA` for the reserve, then
   `MsgTransferFromIca DELEGATION → POOL` for the rest. Then the withdrawal ICA again, since
   undelegation auto-withdraws accrued rewards there.
4. Relayers for the transfer channels stay up permanently (users need them to leave), and the
   redemption ICA channels stay up for the claim trickle (§9).

Staketia (operator key only, no multisig signing):

1. Day 0: undelegate the entire real multisig delegation on Celestia via authz. Call
   `MsgConfirmUndelegation` for the one queued record with that tx hash.
2. Day 21: IBC-transfer the entire liquid multisig balance via authz to the claim address.
   Call `MsgConfirmUnbondedTokenSweep` for each of the 7 records. The epoch hook pays them.
3. Whatever remains in the claim address after the last distribution is moved to the pool by
   the upgrade 2 handler.

## §7. Upgrade 2

Pool: a new stakeibc module account (name in the plan, e.g. `withdrawal_pool`) that holds the
native IBC denoms of every in-scope zone. One account for all denoms; the denom identifies the
zone.

Handler, in order:

1. Move the full balance of staketia's claim address to the pool (bank send; the account is
   module-controlled). Skip if the balance is zero.
2. For each in-scope zone, move the native IBC-denom balance of the host zone's deposit address
   to the pool (liquid stakes that never transferred out; they count in the rate today).
3. For each unclaimed user redemption record whose stTokens are still escrowed (its
   `HostZoneUnbonding` is `UNBONDING_QUEUE` or `UNBONDING_RETRY_QUEUE`): burn
   `StTokenAmount` from the deposit address and set `NativeTokenAmount = StTokenAmount ×
   pinned rate`. Then set that `HostZoneUnbonding` to `CLAIMABLE` with
   `ClaimableNativeTokens` = the sum of its records' native amounts.
4. For each unclaimed record whose stTokens were already burned (`UNBONDING_IN_PROGRESS`,
   `EXIT_TRANSFER_QUEUE`): leave `NativeTokenAmount` as recorded and set its
   `HostZoneUnbonding` to `CLAIMABLE` likewise. Records already `CLAIMABLE` are untouched.
5. Pin the rate: set `HostZone.RedemptionRate` (and `LastRedemptionRate`) per zone to a
   constant from the handler's constants file (§8). Set `MinRedemptionRate`/`Max…` and the
   inner bounds so the pinned rate is inside them, so `BeginBlocker`'s safety check stays
   quiet.
6. Assert, per zone, that the pool's balance of the zone's IBC denom is ≥ circulating supply
   of the stToken × pinned rate, and that each redemption ICA balance is ≥ that zone's reserve
   as pinned. On failure, log and continue (an assertion is a check on the constants, not a
   reason to fail the upgrade; ops top up with `MsgTransferFromIca`).

New tx `MsgRedeemFromPool { creator, amount (stToken coin) }`:

- Resolve the host zone from the stToken denom (`HostDenomFromStAssetDenom`); the zone must
  be marked withdrawal-enabled. The upgrade 2 handler sets a per-zone store flag (key name in
  the plan) for exactly the in-scope zones when it pins their rates; `Halted` alone is not the
  marker, since evmos, stargaze and umee are halted without a pinned rate. Reject stEVMOS,
  stSTARS, stUMEE and unknown denoms.
- `native = amount × RedemptionRate`, truncated. Reject if zero.
- Burn `amount` from the creator (send to module, burn), then send `native` of the zone's
  `IbcDenom` from the pool to the creator. Both through the bank keeper; no records, no ICA,
  no rate math beyond the one multiplication. Insufficient pool balance fails the tx cleanly.
- Emit an event with creator, denom, stToken amount, native amount.

Claim: `ClaimUndelegatedTokens` unchanged except that the `GetActiveHostZone` check becomes
`GetHostZone`. Everything else (status check, `ClaimIsPending`, ICA `MsgSend` from the
redemption ICA, callback deleting the record) is the existing tested path. Claims keep paying
the record's `NativeTokenAmount` on the host chain to `Receiver`.

Staketia: no code change. The 7 records complete in the window (§6). If any is still open at
upgrade 2 the handler leaves the claim address alone for that amount (step 1 subtracts the sum
of open records' native amounts before moving the remainder).

## §8. Accounting rules

The only rate math in the whole design happens once, off-chain, in a script that produces the
upgrade 2 constants. On-chain code never computes a rate again.

Per zone:

```
circulating   = bank supply of stToken − stTokens escrowed in the deposit address
reserve       = Σ NativeTokenAmount over unclaimed records with burned stTokens
              + Σ (StTokenAmount × R) over unclaimed records with escrowed stTokens
total_native  = delegation ICA + withdrawal ICA + fee ICA + redemption ICA
              + deposit address native + pool balance already received
R             = min(frozen rate at upgrade 1, (total_native − reserve) / circulating)
```

Why the records are outside the rate: today a user whose stTokens were burned has their native
earmarked outside `TotalDelegations` (in an unbonding entry or the redemption ICA), and neither
side is in the rate. Pooling that native while their stTokens are absent from supply would
oversubscribe the pool by exactly their share. Keeping their money in the redemption ICA and
their claim at the recorded amount preserves today's invariant. Escrowed-not-burned records
are the mirror case: their stTokens are still in supply and their native is still delegated,
so they are consistent today, but once the escrow is burned they must be priced at R and moved
into the reserve.

`R` uses the `min` because the frozen rate is a promise made under the old accounting and the
observed ratio is the truth after the 99.9% unbond, slashing drift and rewards up to the
undelegation. In practice the two are within a fraction of a percent; the last redeemers absorb
the difference, which §1 accepts.

Because `R` depends on the reserve and the reserve for escrowed records depends on `R`, the
script solves it in two passes (burned-record reserve first, then `R`, then the escrowed
reserve at that `R`, then re-check the pool assertion). Escrowed records are a tiny fraction
of supply, so this converges immediately.

The script reads: bank supply and balances on Stride, ICA balances on each host (delegation,
withdrawal, fee, redemption), and the records. It emits the constants file and a
`verify_constants` check the same way v34 does, and both are regenerated right before the
upgrade 2 proposal, when all balances are static (nothing accrues after the undelegations
complete).

## §9. Risks and mitigations

- **A `MsgUndelegate` over the true on-chain amount fails the whole batch.** 99.9% margin,
  daily retry, admin override of the amount. Validator `DelegationChangesInProgress` flags
  left by stranded acks reduce capacity; v34 resets the stuck ones, and the admin tx can lower
  the amount for a zone if capacity is still short.
- **Delegation ICA channel dies mid-window.** Existing restore flow; the pipeline clears the
  in-flight count and resubmits. Relayer attention is required for the window, as today.
- **Transfer tx timeout.** Funds stay in the ICA; resubmit. No on-chain state to repair.
- **Wrong constants.** The handler assertions (§7 step 6) log the shortfall; the fix is an
  additive `MsgTransferFromIca`, never a state edit. The script and verifier are regenerated at
  proposal time from static balances.
- **Redemption ICA channel closes during the claim trickle.** Existing restore flow. Ops cost
  accepted for 99 records.
- **New redeem tx bug.** It is one burn and one send on a fixed rate, with no record state. The
  plan treats it as the highest-review item: table-driven tests including rounding to zero,
  unknown denoms, out-of-scope denoms, insufficient pool, and an exact-drain of the pool.
- **Un-halt by mistake.** `MsgResumeHostZone` still exists; it is admin-gated. The plan may
  add a guard that refuses to resume a zone with a pinned rate. Optional.
- **Autopilot.** Liquid stakes via memo fail on the halted zone and fall back to the packet
  error, refunding on the host. No change needed.

## §10. Testing

- Unit tests for both admin txs (gating, validation, message construction, no accounting
  mutation), for `MsgRedeemFromPool` (as above), and for the relaxed claim check.
- Upgrade handler tests against a mainnet export, v34-style
  (`app/upgrades/vN/testdata/mainnet_export.json.gz`), for both upgrades: halt flags, pending
  undelegation keys, whitelist pairs; then burns, record repricing, status flips, pinned rates,
  pool balances and the assertions.
- Localstride dry run of upgrade 1 through a full day epoch to see the pending undelegations
  submit, and of upgrade 2 followed by a redeem and a claim.
- The constants script is tested against the same export.

## §11. Open items for the plan

- Module account name and whether the pool lives in stakeibc (recommended) or a tiny new
  module.
- Exact proto shapes and enum names for the two admin txs and the redeem tx.
- Whether to add the resume guard.
- Version numbers for the two upgrades.
- Confirm there are no open LSM token deposits on cosmoshub-4 at upgrade 1 (v34 closed the
  last known stranded one). The LSM loops are halt-gated, so an open one would simply freeze;
  it should be closed or refunded first.
