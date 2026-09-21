# Protocol Wind-Down: Withdrawal-Only Mode

Status: approved design, pre-plan. Large tier; the implementation plan follows in
`docs/superpowers/plans/`.

## §1. Goal

Wind Stride down into a withdrawal-only protocol. Stop every flow that moves funds on behalf of
users, unbond every delegation, bring the native tokens back to Stride into one pool, and let
stToken holders swap their stTokens for the backing native tokens directly on Stride at a fixed
rate. The guiding constraints, in order: least risk of a bug that loses funds, least new code,
most reuse of code that already runs on mainnet, and a chain that is inert while it waits.

The plan is three upgrades separated by two ops windows. It trades calendar time (roughly 70
days end to end) for simplicity: each upgrade lands on a chain with nothing in flight, so no
upgrade has to reason about partially-processed records.

Dormant capital is acceptable. If rounding or slashing drift leave a residue
that no stToken can claim, it stays in the pool. What is not acceptable is a pool that cannot
cover a valid redemption.

## §2. Scope

In scope:

- Every non-deprecated stakeibc host zone: celestia, cosmoshub-4, dydx-mainnet-1,
  haqq_11235-1, injective-1, juno-1, laozi-mainnet, osmosis-1, phoenix-1, sommelier-3, ssc-1.
- Staketia (stTIA redemptions and the 5-of-7 Celestia multisig), including its 7 open
  unbonding records. Its TIA joins the same pool as the stakeibc Celestia TIA.
- The stakeibc user redemption records open at upgrade 1 (99 on 2026-09-18) and any created
  before the redeem message is removed. All are flushed and claimed in window 1.
- Stakedym's 11 open unbonding records (flushed by its operator in window 1), and halting
  stakedym at upgrade 2.
- Removing every message the protocol no longer needs, moving wasm control to gov, and
  removing every IBC rate limit at upgrade 2.

Out of scope, explicitly:

- The deprecated zones comdex-1, evmos_9001-2, stargaze-1 and umee-1. Apart from upgrade 1
  setting the `Deprecated` flag on comdex-1 (the other three already carry it), no upgrade
  touches them, and none of them are unbonded, drained or redeemable.
  Their host chains are halted, so nothing could be done through their ICAs anyway. Stakedym
  is halted at upgrade 2 once its operator has flushed its records, and is likewise never
  unbonded, drained or redeemable. Holders of stCMDX, stEVMOS, stSTARS, stUMEE and stDYM have
  no on-chain redemption path after this work, which matches their status today but is now a
  deliberate decision.
- Removing whole modules or their state. Only messages are removed; stores stay.
- STRD-side modules beyond message removal (mint, strdburner).

## §3. Facts the design rests on

Verified against mainnet on 2026-09-18 and 2026-09-21 (REST `stride-api.polkachu.com`,
Celestia REST `celestia.rpc.uquad.org`).

Gating that already exists:

- `HostZone.Halted` gates every stakeibc flow we want stopped at upgrade 2: `LiquidStake`,
  `RedeemStake`, `ClaimUndelegatedTokens`, LSM, and every epoch-hook loop (`ReinvestRewards`,
  delegation, `UpdateRedemptionRates`, rebalance, `InitiateAllHostZoneUnbondings`,
  `SweepUnbondedTokensAllHostZones`, reward-collector fee liquid staking, LSM loops). All
  iterate `GetAllActiveHostZone`. Staketia and stakedym gate their flows on their own
  `Halted` via `GetUnhaltedHostZone`.
- Two things the halt does not stop. `TransferAllRewardTokens` iterates trade routes, not
  host zones; one route is live and actively converting: dYdX pays rewards in USDC to the
  withdrawal ICA, an epochly ICQ moves it via Noble to the trade ICA on Osmosis, an off-chain
  trade controller with an authz grant swaps it to DYDX, and a second ICQ moves the DYDX back
  to the withdrawal ICA for the normal reinvest split (`reward_converter.go`). The fee ICA is
  an ordinary ICA on each host that receives Stride's commission slice at reinvest and is
  swept to the reward collector by an epochly ICQ; nothing in the pipeline is a hot wallet
  except the trade controller on Osmosis. The delegator-shares
  ICQ callback (reached from `UpdateValidatorSharesExchRate` and `CalibrateDelegation`, both
  permissionless today) applies a detected slash and then calls
  `UpdateRedemptionRateForHostZone` with no halt check.
- Admin messages that move funds via ICA without a halt check: `RebalanceValidators`
  (redelegations) and `ClearBalance` (ICA transfer from the fee account). `ResumeHostZone`
  un-halts a zone.
- Autopilot has an existing `StakeibcActive` param. The ICA host on Stride allow-lists
  `MsgLiquidStake`, `MsgRedeemStake` and `MsgClaimUndelegatedTokens` alongside bank, staking,
  distribution, transfer and gov-vote messages.
- Wasm code upload is restricted to two addresses; all 44 codes were uploaded by one Stride
  address. Deployed contracts are executable by anyone.

Pipelines that already exist:

- Since v34 the undelegate ICA callback accepts a submission with no epoch unbonding record
  ids: on success it only decrements the validator and host zone delegation balances and burns
  nothing (`icacallbacks_undelegate.go`). `BatchSubmitUndelegateICAMessages` takes
  per-validator `MsgUndelegate`s directly and flags `DelegationChangesInProgress` on each, and
  `applySharesRoundingSafety` trims a full-drain amount so share truncation cannot make it
  exceed the on-chain delegation. Neither depends on the zone being active.
- Every host zone has eight ICA types: delegation, fee, withdrawal, redemption, community-pool
  deposit and return, and the two trade-route converter accounts (`ICAAccountType`). The dYdX
  withdrawal ICA holds ~3.8 USDC beside its DYDX.
- `HostZone.Deprecated` is true on exactly evmos_9001-2, stargaze-1 and umee-1, which are
  also halted. comdex-1 is neither halted nor deprecated on chain; its host chain is halted.
  The existing `DeprecateHostZone` message sets both `Halted` and `Deprecated`.
- v34 has a chain-agnostic delegation delta helper (`app/upgrades/v34/delegation_deltas.go`):
  a per-validator table of on-chain minus tracked delegation, applied all-or-nothing to the
  validators and `TotalDelegations`, skipped with a log (never an upgrade error) if any constant
  no longer matches state, with a mainnet-export test. Injective and Celestia use it. The
  delegator-shares slash callback computes a slash as tracked delegation minus on-chain shares
  × the stored exchange rate, so a delegation-only delta is not re-applied when the rate is
  later refreshed.
- An ICA-wrapped IBC `MsgTransfer` from a host account back to Stride exists in the
  trade-route code (`x/stakeibc/keeper/reward_converter.go`, `BuildHostToTradeTransferMsg`).
- The rate limiter covers only stTokens (stATOM, stOSMO, stTIA, stJUNO, stEVMOS); no native
  denom is limited. Its keeper exposes removal of limits, blacklisted denoms and whitelisted
  address pairs.
- Claim today: unbonded tokens sit in the zone's redemption ICA on the host;
  `ClaimUndelegatedTokens` is permissionless, submits an ICA bank `MsgSend` from the
  redemption ICA to the record's host-chain `Receiver`, sets `ClaimIsPending`, and the callback
  deletes the record on ack (the timeout callback resets a stuck one). `UserRedemptionRecord`
  stores no Stride address; autopilot redemptions set `Creator` to a derived address, so claims
  can only ever be paid on the host.
- Staketia's flow is operator-driven with automatic payout: redeem escrows stTIA → epoch
  `PrepareUndelegation` freezes the record → operator undelegates on Celestia and calls
  `MsgConfirmUndelegation` (burns stTIA) → after 21 days `MarkFinishedUnbondings` → operator
  IBC-transfers to the claim address and calls `MsgConfirmUnbondedTokenSweep` (checks the
  claim-address balance covers the record) → next hour epoch `DistributeClaims` pays each
  redeemer on Stride. Stakedym is the same module shape.
- The Celestia multisig (`celestia1d6ntc7s8gs86tpdyn422vsqc6uaz9cejnxz5p5`, 5-of-7) has
  granted the operator key (`celestia1ghhu67ttgmxrsyxljfl2tysyayswklvxzls400`) unexpiring
  authz for delegate, undelegate, withdraw rewards, cancel unbonding, and an IBC
  `TransferAuthorization` on `transfer/channel-4` with an effectively unlimited utia limit and
  exactly one allow-listed receiver: staketia's claim address
  `stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd`. Bank `MsgSend` is not granted. So the
  operator alone can move the whole multisig balance to Stride, but only onto the claim
  address; upgrade 3 bank-moves it from there into the pool. Staketia's IBC middleware only
  hooks outbound acks and timeouts, so an unrecorded inbound transfer to the claim address is
  inert.

State on mainnet (2026-09-18/21):

- Unbonding periods: osmosis-1 14 days, dydx-mainnet-1 30, juno-1 and sommelier-3 28, all
  others 21. Zones submit undelegations every 4 day-epochs.
- 99 open stakeibc user redemption records, 3 with a claim pending. Open unbondings:
  queued-not-submitted amounts on osmosis-1, phoenix-1, ssc-1, laozi-mainnet and a
  cosmoshub-4 retry (~54.7k ATOM); unbonded-not-swept on cosmoshub-4, dydx-mainnet-1,
  injective-1, juno-1, osmosis-1, phoenix-1, ssc-1.
- Staketia: 6 `UNBONDING_IN_PROGRESS`, 1 `UNBONDING_QUEUE` (~717 stTIA), 1 accumulating.
  Stakedym: 5 `UNBONDED` (~1,067 DYM awaiting sweep), 6 `UNBONDING_QUEUE` (~3,294 stDYM
  escrowed), 30 redemption records; 284,578 DYM recorded delegation vs 260,352 stDYM supply.
- Zero auctions, zero ICQ-oracle price queries; both airdrop-module airdrops ended December
  2024. Three ICA oracles active (injective-1, neutron-1, osmosis-1).

## §4. Overview

| | New logic | Removes | Ops window after |
|---|---|---|---|
| Upgrade 1: close the doors | none (a Haqq delegation delta table, v34 pattern) | liquid stake, redeem, and every create-things message; wasm to gov | ~35 days: flush unbondings, claim for everyone, operators finish staketia/stakedym |
| Upgrade 2: halt and unbond | two admin txs, one ValidateBasic gate | claim, rebalance, clear-balance, resume, trade route, oracles, all rate limits, calibration cap | ~32 days: refresh slashes, undelegate every validator, drain every ICA into the pool |
| Upgrade 3: withdrawal mode | pool account, one redeem tx | nothing | permanent |

Each upgrade lands on a chain with nothing in flight, verified by the checklist that gates its
proposal (§8).

## §5. Upgrade 1: close the doors

No new logic. Only message removals and parameter changes.

Remove message handlers (proto `rpc`, msg server, CLI, tests; keeper functions stay because the
flush and the unbond pipeline use them):

- stakeibc: `LiquidStake`, `LSMLiquidStake`, `RedeemStake`, `RegisterHostZone`,
  `CreateTradeRoute`, `UpdateTradeRoute`, `DeleteTradeRoute`, `SetCommunityPoolRebate`,
  `ToggleTradeController`.
- staketia and stakedym: `RedeemStake` (their `LiquidStake` is already a disabled stub; remove
  the stub too).
- icaoracle: `AddOracle`, `InstantiateOracle`.
- icqoracle: `RegisterTokenPriceQuery`, `RemoveTokenPriceQuery`.
- auction: `PlaceBid`, `CreateAuction`, `UpdateAuction`.
- airdrop: all seven. claim (legacy): all four.

Removing a message rejects it at tx decode for every entry path that goes through the msg
router with a registered type, and is a stronger block than a flag. Two entry points do not
decode a message and are closed separately:

- Autopilot: handler sets `StakeibcActive = false`.
- ICA host: handler removes `MsgLiquidStake` and `MsgRedeemStake` from the allow-list.
  `MsgClaimUndelegatedTokens` stays until upgrade 2 so ICA-originated claims work in window 1.

Wasm: handler sets `code_upload_access` to the gov module address only, and for every deployed
contract whose admin is a Stride-controlled key, sets the admin to the gov module address
(`ContractKeeper.UpdateContractAdmin`). The plan lists the contracts from a per-contract query.

Comdex: handler sets `Deprecated = true` on comdex-1 so it carries the same flag as the other
three deprecated zones. `Halted` is not touched, per the decision to leave deprecated zones as
they are; the flag is documentation and the upgrade 3 redeem gate.

Haqq delegation reconciliation: apply a per-validator delta table to haqq_11235-1 with the v34
helper, exactly as v34 did for Injective. The 2026-09-21 measurement (§8a) has 14 validators
over-recorded, all undetected downtime slashes plus two sub-token dust cases, the largest
853.8 ISLM. Every delta is negative, so `TotalDelegations` drops by the table's sum and the
next epoch's rate update, which still runs in window 1, lowers the stISLM redemption rate to
match. That is the correct outcome for a slash that was never detected. The stored
`SharesToTokensRate` is deliberately left as is: the window-2 refresh will update it, and the
slash callback then finds tracked delegation equal to on-chain shares × the refreshed rate, so
nothing is applied twice. The table is generated from `measure_delegation_drift.py`,
re-measured right before the proposal, and covered by a mainnet-export test. This is the
Injective shape (real loss, no stranded liquid), not the v33 Osmosis shape (phantom stake
credited back as a deposit record).

Before applying the table, the handler throws out every slash-path ICQ open for haqq_11235-1
at that moment: it iterates the pending interchain queries, deletes those for that chain whose
callback is the validator exchange rate, delegator shares or calibration callback, and clears
`SlashQueryInProgress` on every haqq validator. Unlike v34's `DeleteStuckQueries` and
`ResetStuckSlashQueries`, which pin query ids and validator addresses, this is dynamic, so
the constants cannot go stale between measurement and execution (haqq had 8 such queries
open and one flagged validator on 2026-09-21). A stale response landing after the delta could
not double-apply a slash (§3), but a query submitted against the pre-delta state has no
reason to exist and the reinvest path resubmits fresh ones each epoch. The withdrawal-balance
query is not a slash query and is left alone.

Everything else keeps running on purpose: reinvest, rate updates, unbonding, sweep, claim, the
reward-collector fee liquid stake, the dYdX trade route, the oracles, and the
staketia/stakedym operator flows. The fee liquid stake mints stTokens for validators during the window; it is
accounting-consistent and stops at upgrade 2, so it is not worth a switch.

## §6. Upgrade 2: halt and unbond

Pool: a new stakeibc module account (name in the plan, e.g. `withdrawal_pool`) is registered in
this binary, added to the bank blocked-address list so nothing can be sent to it from outside
the module, and is the destination of every window-2 transfer. One account for all denoms; the
denom identifies the zone.

Handler:

1. `Halted = true` on every in-scope stakeibc zone and on stakedym's host zone. The four
   deprecated zones are not touched.
2. Delete the dYdX trade route (its converter ICAs are written off; the addresses are noted in
   the plan). Deactivate the three ICA oracles (existing toggle logic). Remove
   `MsgClaimUndelegatedTokens` from the ICA host allow-list.
3. Remove every rate limit, every blacklisted denom and every whitelisted address pair from
   the rate limiter. After upgrade 3 the only outbound flow that matters is users taking native
   tokens off the chain, which must never be throttled, and there is no longer a mint path that
   an stToken limit would protect. The module and middleware stay in the stack with empty state.

Remove message handlers: stakeibc `ClaimUndelegatedTokens`, `RebalanceValidators`,
`ClearBalance`, `ResumeHostZone`; staketia and stakedym `ResumeHostZone`. Add
`utils.ValidateAdminAddress` to the ValidateBasic of `UpdateValidatorSharesExchRate` and
`CalibrateDelegation`; ops keep them to refresh slashes before the unbond (§8), but nobody
else can trigger the slash path that rewrites the rate. Remove the 5,000 base-unit
`CalibrationThreshold` check from the calibration callback: it existed to bound what a
permissionless caller could move, and the message is admin-gated from here on.

New admin tx `MsgUndelegateFromValidators { creator, chain_id, offset, validators: [{address,
offset}] }`, admin-gated in ValidateBasic. An empty `validators` list means every validator on
the zone with a non-zero stored delegation; otherwise only the listed ones. The top-level
`offset` applies to every validator and a per-validator `offset` overrides it for that one.
Per validator the amount is the stored delegation minus its effective offset, passed through
`applySharesRoundingSafety`, and rejected if it is not positive. The real run uses a global
offset of one base unit so no validator can fail the host's share check on a full drain;
that is immaterial and is on top of the helper, which already buffers full drains of
validators whose exchange rate is below one. A listed validator with `DelegationChangesInProgress` set
is rejected, so a batch cannot be double-submitted while its ack is outstanding. The messages
go through `BatchSubmitUndelegateICAMessages` with no epoch unbonding record ids, so the
existing callback decrements the balances and nothing is burned.

Submitting by validator rather than cascading a zone amount by weight is what lets ops test on
a single validator first, skip or shave a validator that is failing or drifted by dust, and
keep every submitted amount equal to what the host will accept. Retry is resubmission by ops;
there is no queue and no epoch hook. ICA txs are atomic, so one over-recorded validator still
fails its whole batch; the slash refresh (§8, step 1) is what prevents that, and `offset` is
the manual lever if a validator is still off by dust.

New admin tx `MsgTransferFromIca { creator, chain_id, ica_type, amount (Coin) }`, admin-gated:
`ica_type ∈ {DELEGATION, WITHDRAWAL, FEE, REDEMPTION, COMMUNITY_POOL_DEPOSIT,
COMMUNITY_POOL_RETURN}`, the six ICAs a host zone owns (the two converter ICAs belong to the
trade route deleted in this same upgrade and are written off). `amount` carries its denom as it exists on the host,
so foreign balances such as the USDC in the dYdX withdrawal ICA can be swept too; anything
that is not the zone's host denom becomes pool surplus. Submits one ICA containing an IBC
`MsgTransfer` of `amount` from that ICA over the zone's transfer channel to the pool address,
built like `BuildHostToTradeTransferMsg`, with the existing ICA timeout. No
callback state: a failed or timed-out transfer leaves the funds in the ICA and ops resubmit.
The tx never reads or writes host zone accounting; amounts are an ops input checked against
ICA balances.

## §7. Upgrade 3: withdrawal mode

The pool account exists since upgrade 2 (§6) and is full by now.

Handler, in order:

1. Bank-move the full balance of staketia's claim address to the pool.
2. For each in-scope zone, bank-move the native IBC-denom balance of the zone's deposit address
   to the pool (liquid stakes that never transferred out; they count in the rate today).
3. Assert per zone that pool balance ≥ stToken supply × `HostZone.RedemptionRate`, both read
   from state, and return an error if any zone fails, which fails the upgrade. This is the one
   place a hard error is right: a shortfall must be fixed before redemptions open anyway, and a
   halted chain during the wind-down costs nothing because nobody can use it. Validators get a
   corrected binary. The fix is an additive `MsgTransferFromIca`, never a state edit.

There is no pinned-rate store and no constants file. The redeem tx reads the host zone's
`RedemptionRate`, which has been frozen since the upgrade 2 halt (§9).

New tx `MsgRedeemFromPool { creator, chain_id, amount (stToken coin) }`:

- ValidateBasic: `amount` is positive and at least the chain's minimum, from a hard-coded set
  of 18-decimal chains (dydx-mainnet-1, injective-1, haqq_11235-1, evmos_9001-2): 1e17 base
  units there, 1e5 elsewhere. No state is needed for the check.
- Load the host zone by `chain_id`; require `Halted && !Deprecated`, and that `amount.Denom`
  is that zone's stToken denom. The gate admits exactly the eleven in-scope zones: all four
  deprecated zones carry the flag from upgrade 1 on, and comdex-1 is additionally never halted. A redeem of a deprecated stToken would in any case fail atomically at
  the send step because the pool holds none of that denom and cannot be funded from outside.
- `native = amount × HostZone.RedemptionRate`, truncated; reject zero.
- Burn `amount` from the creator (send to module, burn), then send `native` of the zone's
  `IbcDenom` from the pool to the creator. Two bank operations, one multiplication, no records,
  no ICA. Insufficient pool balance fails the tx cleanly.
- Emit an event with creator, denom, stToken amount, native amount.

Nothing else changes. `MsgTransferFromIca` and `MsgUndelegateFromValidators` stay for stragglers.

## §8. Ops windows and proposal checklists

Window 1 (after upgrade 1, ~35 days):

1. Each zone's queued redemptions are submitted at its next unbonding epoch (≤ 4 days), unbond
   (≤ 30 days), sweep at the next stride epoch, and go `CLAIMABLE`. Watch the cosmoshub-4
   retry record until it succeeds.
2. As records go claimable, run `ClaimUndelegatedTokens` for every open record (it is
   permissionless; ops already do this). A record whose host receiver rejects the bank send is
   handled by hand.
3. Staketia operator (flush only; the multisig stays staked until window 2): day 0,
   undelegate the queued record's native amount on Celestia via authz and
   `MsgConfirmUndelegation` for that record with the tx hash. Day 21, IBC the unbonded amount
   via authz to the claim address and `MsgConfirmUnbondedTokenSweep` for each of the 7
   records. The hour-epoch hook pays them.
4. Stakedym operator: sweep and confirm the 5 unbonded records; undelegate and confirm the 6
   queued ones; after 21 days sweep and confirm those. The hook pays them.

Checklist to propose upgrade 2:

- Zero stakeibc user redemption records; zero `HostZoneUnbonding` records outside `CLAIMABLE`
  with a non-zero amount; no claim ICA in flight.
- Staketia: zero unbonding records outside `CLAIMED`; zero redemption records; claim address
  empty.
- Stakedym: same, for its records.
- No open LSM token deposit on cosmoshub-4.
- The dYdX trade route is quiet: no pending withdrawal-reward or trade-converted-balance ICQ,
  no packet commitment on the dYdX→Noble, Noble→Osmosis or Osmosis→dYdX transfer channels used
  by the route, and zero USDC and DYDX on the Osmosis trade ICA. Any USDC left in the dYdX
  withdrawal ICA is swept in window 2 as surplus.

Window 2 (after upgrade 2, ~32 days):

1. Day 0: refresh every validator's exchange rate with `UpdateValidatorSharesExchRate`
   (CLI `update-delegation`) on every in-scope zone and wait for the callbacks. Where the rate
   moved, the callback applies the slash: recorded delegation becomes on-chain shares × the
   new rate, and the redemption rate is lowered to match. Use `CalibrateDelegation` only for a
   validator whose rate is unchanged but whose recorded balance is off by under 5,000 base
   units (its hard cap). Then rerun the drift measurement (§8a) and require zero
   over-recorded validators on the zone.
2. `MsgUndelegateFromValidators` per zone: first for a single small validator as a live test
   of the tx and the callback, then with an empty list for the rest. An ICA tx is atomic, so
   one over-recorded validator fails its whole batch; after the refresh there are none. If a
   slash lands between the refresh and the submission, that batch fails, ops rerun the refresh
   (or raise that validator's `offset`) and resubmit for the affected validators. The real run
   passes a global `offset` of one base unit. Delegation ICA channels and
   relayers stay healthy until every batch acks; a dead channel is restored with the existing
   flow and the affected validators are resubmitted.
3. Day 0+: `MsgTransferFromIca` for withdrawal, fee, redemption and community-pool ICA
   balances, including foreign denoms such as the dYdX USDC. This is the live test of the
   transfer tx on small real amounts.
4. As each zone's unbonding completes (day 14 to day 30): `MsgTransferFromIca DELEGATION` for
   the full balance, then `WITHDRAWAL` again (undelegation auto-withdraws accrued rewards
   there).
5. Staketia operator, no Stride txs needed: day 0, undelegate the entire multisig delegation
   on Celestia via authz alongside the stakeibc unbonds; day 21, IBC the whole liquid balance
   via authz to the claim address, where it waits for upgrade 3.
6. Transfer-channel relayers stay up permanently; users need them to leave. ICA channels can be
   left to close once every balance is drained.

### §8a. Drift measurement

Measured 2026-09-21 for every in-scope zone except cosmoshub-4 and injective-1 (handled by
v34): recorded per-validator delegations versus the delegation ICA's on-chain balances.
dydx-mainnet-1, ssc-1 and sommelier-3 are exact. celestia is under-recorded by ~15,440 TIA
(the v34 phantom stake; harmless for unbonding). osmosis-1 has one validator over by 3,026
uosmo with an unchanged rate (calibration range). juno-1 (1 validator, 1.94 JUNO),
laozi-mainnet (2, 4.53 BAND), phoenix-1 (10, max 1.20 LUNA) and haqq_11235-1 (14, max 853.8
ISLM, 0.01% of that validator) are over-recorded, and in every case Stride's stored exchange
rate is above the chain's, i.e. undetected downtime slashes. Recomputing each validator as
on-chain shares × the chain's current rate reproduces the on-chain balance exactly for every
validator on every zone, so the refresh leaves no rounding gap and no per-validator buffer is
needed. The one wrinkle is haqq_11235-1's 18-decimal denom: two validators (SureStake,
Islamic Staking) are over by 203,557 and 3,216,141 aISLM with an unchanged rate, which is
dust in ISLM but above calibration's 5,000-base-unit cap, and even one base unit of
over-recording fails the host's share check. Haqq is therefore trued up by a delta table at
upgrade 1 (§5), which covers the dust cases too, and the `offset` on
`MsgUndelegateFromValidators` is the fallback for anything that drifts afterwards. The
measurement script is `scripts/wind-down/measure_delegation_drift.py` and is rerun as the gate
before step 2; haqq_11235-1 is expected to measure clean by then.

Checklist to propose upgrade 3:

- No undelegate batch in flight (no validator with `DelegationChangesInProgress`) and
  `TotalDelegations` at dust on every zone except celestia, where it still carries the
  multisig portion (the per-validator unbond only touches ICA validators; that portion is the
  claim-address balance and is counted by the pool assertion).
- All six ICA balances at dust on every zone. Celestia multisig delegation zero, its
  unbondings complete, and the claim address holding its full balance.
- The upgrade 3 handler test is run against a fresh mainnet export and a localstride started
  from that export, and both pass the pool assertion (pool balance plus the staketia claim
  address, for celestia, ≥ stToken supply × `HostZone.RedemptionRate`) for every zone. The
  handler errors on a shortfall, so this is what keeps the upgrade from failing on mainnet;
  a shortfall found here is topped up with `MsgTransferFromIca` before the proposal.

## §9. Accounting

There is no rate calculation anywhere in the design. The withdrawal rate for a zone is the
`HostZone.RedemptionRate` that was on chain when upgrade 2 halted the zone, read live by the
redeem tx. Nothing recomputes it: the epoch update is
halt-gated, and the only other writer (the delegator-shares slash callback) is reachable only
through the two admin-gated ICQ messages. If ops deliberately run those in window 2 and a slash
is found, the frozen rate is lowered accordingly, which is the correct direction.

Why a frozen rate is covered. Confirmed against cosmos-sdk v0.54.3:
`Unbond` calls the distribution hook `BeforeDelegationSharesModified`, which withdraws the
accrued rewards, and then removes the shares; rewards are computed from delegation shares only,
so an unbonding entry earns nothing during its 21 to 30 days. Between the halt and the pool
being full, the backing therefore moves as follows:

- Up: staking rewards accrue only from the halt until the undelegate executes at the next day
  epoch (one or two days, roughly 0.03% to 0.05%), are auto-withdrawn to the withdrawal ICA by
  the undelegation, and are swept into the pool. Rewards withdrawn but not yet reinvested at
  the halt are also swept.
- Down: any slash during window 2.

Coverage then follows by construction. The refresh (§8, step 1) makes the recorded
delegations and the redemption rate reflect every slash that has happened, the full recorded
amount unbonds, and the pool receives it plus the rewards swept on undelegation. Supply ×
frozen rate is the delegated amount, so the rewards are the buffer. Only a slash between the
refresh and the undelegation, on a zone whose retried batch never lands, can leave a zone
under, and then the last redeemers absorb it, which §1 accepts. A computed rate would remove
that tail risk too, but adds a script on the one path where a bug is catastrophic; the frozen
rate is a number that has already been on chain and audited for weeks.

Bank supply is the right reference for the assertion: stTokens that left Stride over IBC are
escrowed here, not burned, so they are counted; stTokens burned by flushed redemptions are gone
from both sides. Nothing else needs a term, because upgrade 3 lands with zero open records,
empty redemption ICAs and empty unbonding queues.

## §9a. Validator compensation

Today POA validators receive two streams: STRD from mint provisions (the staking share of
~58 STRD per hour epoch, about 16%) plus tx fees, and stTokens from the reward collector (15%
of Stride's commission on host staking rewards, liquid staked at the mint epoch).

The stToken stream is unaffected in window 1 (everything keeps running) and ends at upgrade 2,
not because of the halt but because there are no staking rewards to share once the delegations
are gone. The rewards that accrue during window 2 are swept into the pool as protocol surplus
rather than split. The STRD stream continues unchanged. If validator pay needs to rise to
compensate, that is a mint distribution-proportion parameter change by governance and needs no
code. Releasing pool surplus (pool minus supply × rate) to validators or the community would
need a gov-gated withdrawal message and is out of scope for these three upgrades.

## §10. Testing

- Upgrade 1: handler tests against a mainnet export (`app/upgrades/vN/testdata/`, v34-style)
  for the comdex-1 `Deprecated` flag, the Haqq delta table (applied, and skipped on a stale
  constant), the haqq slash-query
  purge (deletes only that chain's slash-path queries, clears the validator flags, leaves other
  chains' and the withdrawal-balance queries), the autopilot param, ICA
  host allow-list, wasm params and contract admins; a compile-time
  guarantee that the removed messages no longer exist; existing keeper tests for the flows
  that keep running stay green.
- Upgrade 2: unit tests for both admin txs (gating, validation, per-validator message
  construction with and without offsets, empty versus explicit validator lists, rounding
  safety on a full drain, rejection of a validator with a change in progress, every ICA type
  and a foreign denom on the transfer, no accounting mutation), for the two ValidateBasic gates
  and for the lifted calibration cap; handler tests for the halt flags, trade route deletion,
  oracle deactivation, rate-limit removal and the pool account's existence and blocked status. Localstride run: upgrade, then
  `MsgUndelegateFromValidators` for one validator and its ack.
- Upgrade 3: `MsgRedeemFromPool` is the highest-review item: table-driven tests covering
  the per-chain minimum, rounding to zero, an unknown chain id, a denom that is not the zone's
  stToken, a zone that is not halted, a deprecated zone, insufficient pool, an exact drain of
  the pool, and a second redeem after the drain. Handler tests for the bank
  moves and the assertion. Localstride run: upgrade, then a redeem.

## §11. Open items for the plan

- Module account name; exact proto shapes and enum names for the two admin txs and the redeem
  tx.
- Per-contract wasm admin listing for upgrade 1.
- Version numbers for the three upgrades.
- Whether the legacy claim module's 2022 airdrops are already expired (its REST query is not
  served; confirm from a full node). Its messages are removed either way.
