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

Dormant capital is acceptable. If rounding, slashing drift or the 99.9% unbond leave a residue
that no stToken can claim, it stays in the pool. What is not acceptable is a pool that cannot
cover a valid redemption.

## §2. Scope

In scope:

- Every non-halted stakeibc host zone: celestia, comdex-1, cosmoshub-4, dydx-mainnet-1,
  haqq_11235-1, injective-1, juno-1, laozi-mainnet, osmosis-1, phoenix-1, sommelier-3, ssc-1.
- Staketia (stTIA redemptions and the 5-of-7 Celestia multisig), including its 7 open
  unbonding records. Its TIA joins the same pool as the stakeibc Celestia TIA.
- The stakeibc user redemption records open at upgrade 1 (99 on 2026-09-18) and any created
  before the redeem message is removed. All are flushed and claimed in window 1.
- Stakedym's 11 open unbonding records (flushed by its operator in window 1), and halting
  stakedym at upgrade 2.
- Removing every message the protocol no longer needs, and moving wasm control to gov.

Out of scope, explicitly:

- Redemption of stEVMOS, stSTARS, stUMEE (zones already halted) and stDYM (halted at upgrade
  2). Holders of those four tokens have no on-chain redemption path after this work, which
  matches their status today but is now a deliberate decision.
- The IBC rate limits. They stay (they only cover stTokens; no native denom is limited).
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
  host zones; one route is live (dYdX rewards → Noble USDC → Osmosis). The delegator-shares
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

- The v34 pending-undelegation pipeline (`x/stakeibc/keeper/pending_undelegation.go`) queues
  an amount per host zone, submits it at the day epoch through the normal validator-capacity
  logic, tracks in-flight batches, and retries failed batches daily. It reads the host zone
  with `GetHostZone`, so it runs on a halted zone. Its callback decrements `TotalDelegations`
  and validator balances and burns nothing.
- An ICA-wrapped IBC `MsgTransfer` from a host account back to Stride exists in the
  trade-route code (`x/stakeibc/keeper/reward_converter.go`, `BuildHostToTradeTransferMsg`).
- Rate-limit whitelisting of a (sender, receiver) pair:
  `RatelimitKeeper.SetWhitelistedAddressPair`, as for deposit → delegation ICA today.
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
| Upgrade 1: close the doors | none | liquid stake, redeem, and every create-things message; wasm to gov | ~35 days: flush unbondings, claim for everyone, operators finish staketia/stakedym |
| Upgrade 2: halt and unbond | two admin txs, one ValidateBasic gate | claim, rebalance, clear-balance, resume, trade route, oracles | ~32 days: undelegate 99.9%, drain every ICA into the pool |
| Upgrade 3: withdrawal mode | pool account, pinned rates, one redeem tx | nothing | permanent |

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

Everything else keeps running on purpose: reinvest, rate updates, unbonding, sweep, claim, the
reward-collector fee liquid stake, the trade route, the oracles, and the staketia/stakedym
operator flows. The fee liquid stake mints stTokens for validators during the window; it is
accounting-consistent and stops at upgrade 2, so it is not worth a switch.

## §6. Upgrade 2: halt and unbond

Handler:

1. `Halted = true` on every in-scope stakeibc zone and on stakedym's host zone.
2. Delete the dYdX trade route. Deactivate the three ICA oracles (existing toggle logic).
   Remove `MsgClaimUndelegatedTokens` from the ICA host allow-list.
3. Whitelist in the rate limiter, per zone, (delegation ICA → pool), (withdrawal ICA → pool),
   (fee ICA → pool), (redemption ICA → pool). The pool is the module account defined in §7; its
   address is deterministic, so the pair can be set before the account exists.

Remove message handlers: stakeibc `ClaimUndelegatedTokens`, `RebalanceValidators`,
`ClearBalance`, `ResumeHostZone`; staketia and stakedym `ResumeHostZone`. Add
`utils.ValidateAdminAddress` to the ValidateBasic of `UpdateValidatorSharesExchRate` and
`CalibrateDelegation`; ops keep them to diagnose a failing undelegation, but nobody else can
trigger the slash path that rewrites the rate.

New admin tx `MsgSetPendingUndelegation { creator, chain_id, amount }`, admin-gated in
ValidateBasic: calls `SetPendingUndelegation`; zero removes the key. Nothing else. The v34
pipeline submits, tracks and retries at the day epoch. Ops pass ~99.9% of `TotalDelegations`.
ICA txs are atomic, so one `MsgUndelegate` above a validator's true on-chain delegation fails
the whole batch; recorded delegations drift from the chain and the margin absorbs it. The tx is
also the override if a zone keeps failing (lower the amount).

New admin tx `MsgTransferFromIca { creator, chain_id, ica_type, amount }`, admin-gated:
`ica_type ∈ {DELEGATION, WITHDRAWAL, FEE, REDEMPTION}`. Submits one ICA containing an IBC
`MsgTransfer` of `amount` of the host denom from that ICA over the zone's transfer channel to
the pool address, built like `BuildHostToTradeTransferMsg`, with the existing ICA timeout. No
callback state: a failed or timed-out transfer leaves the funds in the ICA and ops resubmit.
The tx never reads or writes host zone accounting; amounts are an ops input checked against
ICA balances.

## §7. Upgrade 3: withdrawal mode

Pool: a new stakeibc module account (name in the plan, e.g. `withdrawal_pool`) holding the
native IBC denom of every in-scope zone. One account; the denom identifies the zone.

Handler, in order:

1. Bank-move the full balance of staketia's claim address to the pool.
2. For each in-scope zone, bank-move the native IBC-denom balance of the zone's deposit address
   to the pool (liquid stakes that never transferred out; they count in the rate today).
3. Write the pinned rate per zone into a new store entry `WithdrawalRate{chain_id → rate}`
   from the handler's constants (§9). This entry is the only thing the redeem tx reads and the
   only marker that a zone is withdrawal-enabled. It is deliberately not
   `HostZone.RedemptionRate`, so no legacy path (the slash callback, bounds checks, the oracle)
   can touch it. Evmos, stargaze, umee and stakedym get no entry.
4. Assert per zone that pool balance ≥ stToken supply × pinned rate. Log on failure and
   continue; the fix is an additive `MsgTransferFromIca`, never a state edit.

New tx `MsgRedeemFromPool { creator, amount (stToken coin) }`:

- Resolve the zone from the stToken denom; require a `WithdrawalRate` entry.
- `native = amount × rate`, truncated; reject zero.
- Burn `amount` from the creator (send to module, burn), then send `native` of the zone's
  `IbcDenom` from the pool to the creator. Two bank operations, one multiplication, no records,
  no ICA. Insufficient pool balance fails the tx cleanly.
- Emit an event with creator, denom, stToken amount, native amount.

Nothing else changes. `MsgTransferFromIca` and `MsgSetPendingUndelegation` stay for stragglers.

## §8. Ops windows and proposal checklists

Window 1 (after upgrade 1, ~35 days):

1. Each zone's queued redemptions are submitted at its next unbonding epoch (≤ 4 days), unbond
   (≤ 30 days), sweep at the next stride epoch, and go `CLAIMABLE`. Watch the cosmoshub-4
   retry record until it succeeds.
2. As records go claimable, run `ClaimUndelegatedTokens` for every open record (it is
   permissionless; ops already do this). A record whose host receiver rejects the bank send is
   handled by hand.
3. Staketia operator, day 0: undelegate the entire real multisig delegation on Celestia via
   authz; `MsgConfirmUndelegation` for the queued record with that tx hash. Day 21: IBC the
   whole liquid balance via authz to the claim address; `MsgConfirmUnbondedTokenSweep` for
   each unbonded record. The hour-epoch hook pays them. The remainder sits in the claim address
   until upgrade 3.
4. Stakedym operator: sweep and confirm the 5 unbonded records; undelegate and confirm the 6
   queued ones; after 21 days sweep and confirm those. The hook pays them.

Checklist to propose upgrade 2:

- Zero stakeibc user redemption records; zero `HostZoneUnbonding` records outside `CLAIMABLE`
  with a non-zero amount; no claim ICA in flight.
- Staketia: zero unbonding records outside `CLAIMED`; zero redemption records; multisig
  delegation zero and unbondings complete; the claim address holds the full multisig balance.
- Stakedym: same, for its records.
- No trade-route transfer in flight (Noble and Osmosis accounts checked; residual USDC or DYDX
  is either swept by hand or written off as dormant).
- No open LSM token deposit on cosmoshub-4.

Window 2 (after upgrade 2, ~32 days):

1. Day 0: `MsgSetPendingUndelegation` per zone at 99.9% of `TotalDelegations`. Delegation ICA
   channels and relayers stay healthy until every batch acks; a dead channel is restored with
   the existing flow and the pipeline resubmits.
2. Day 0+: `MsgTransferFromIca` for withdrawal, fee and redemption ICA balances. This is the
   live test of the transfer tx on small real amounts.
3. As each zone's unbonding completes (day 14 to day 30): `MsgTransferFromIca DELEGATION` for
   the full balance, then `WITHDRAWAL` again (undelegation auto-withdraws accrued rewards
   there).
4. Transfer-channel relayers stay up permanently; users need them to leave. ICA channels can be
   left to close once every balance is drained.

Checklist to propose upgrade 3:

- Every pending undelegation key gone, no batch in flight, `TotalDelegations` ≈ 0.1% of its
  pre-unbond value per zone.
- Delegation, withdrawal, fee and redemption ICA balances at dust on every zone.
- Constants file and verifier (§9) regenerated from the final balances, and the pool assertion
  passes for every zone.

## §9. Accounting

The only rate math in the design happens once, off-chain, in a script that produces the
upgrade 3 constants. On-chain code never computes a rate again. Per zone:

```
R = min( HostZone.RedemptionRate at upgrade 2,
         (pool balance + deposit-address native + [staketia claim address, for celestia])
         / bank supply of the stToken )
```

Bank supply is the right denominator: stTokens that left Stride over IBC are escrowed here,
not burned, so they are counted; stTokens burned by flushed redemptions are gone from both
sides. Nothing else needs a term, because upgrade 3 lands with zero open records, empty
redemption ICAs, and empty unbonding queues.

The `min` is because the frozen rate is a promise made under the old accounting, and the
observed ratio is the truth after the 99.9% unbond, slashing drift and rewards up to the
undelegation. In practice they are within a fraction of a percent; the last redeemers absorb
the difference, which §1 accepts.

The script reads bank supply and balances on Stride and emits the constants file plus a
`verify_constants` check, v34-style. Both are regenerated right before the upgrade 3
proposal, when every balance is static.

## §10. Testing

- Upgrade 1: handler tests against a mainnet export (`app/upgrades/vN/testdata/`, v34-style)
  for the autopilot param, ICA host allow-list, wasm params and contract admins; a compile-time
  guarantee that the removed messages no longer exist; existing keeper tests for the flows
  that keep running stay green.
- Upgrade 2: unit tests for both admin txs (gating, validation, message construction, no
  accounting mutation) and for the two ValidateBasic gates; handler tests for the halt flags,
  trade route removal, oracle deactivation and whitelist pairs. Localstride run through a day
  epoch to see a pending undelegation submit.
- Upgrade 3: `MsgRedeemFromPool` is the highest-review item: table-driven tests covering
  rounding to zero, unknown denoms, denoms without a `WithdrawalRate`, insufficient pool, an
  exact drain of the pool, and a second redeem after the drain. Handler tests for the bank
  moves, the pinned entries and the assertion. Localstride run: upgrade, then a redeem.
- The constants script is tested against the same export.

## §11. Open items for the plan

- Module account name; exact proto shapes and enum names for the two admin txs, the redeem tx
  and `WithdrawalRate`.
- Per-contract wasm admin listing for upgrade 1.
- Version numbers for the three upgrades.
- Whether the legacy claim module's 2022 airdrops are already expired (its REST query is not
  served; confirm from a full node). Its messages are removed either way.
