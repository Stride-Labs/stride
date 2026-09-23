# Protocol Wind-Down: Migration to Osmosis

Status: approved design, pre-plan. Large tier; the implementation plan follows in
`docs/superpowers/plans/`.

## §1. Goal

Wind Stride down and shut the chain off. Stop every flow that moves funds on behalf of users,
unbond every delegation, send the native tokens from the host chains straight to Osmosis, and
put them into one transmuter pool per stToken where holders swap stTokens for the backing
native tokens at a fixed rate. stTokens still sitting in Stride accounts are sent to their
owners on Osmosis before the halt, so nothing a user owns is left on a dead chain. The guiding
constraints, in order: least risk of a bug that loses funds, least new code, most reuse of code
that already runs on mainnet, and a chain that stops rather than one that idles for a year.

The plan is two upgrades separated by one ops window, then a second ops window that ends with
a coordinated halt. It trades calendar time (roughly 70 days end to end) for simplicity: each
upgrade lands on a chain with nothing in flight, so no upgrade has to reason about
partially-processed records, and no on-chain rate is ever computed.

Dormant capital is acceptable. Backing that no stToken ever claims stays in the pools and comes
back to Stride Labs by exiting them later. What is not acceptable is a pool that cannot cover
a valid swap.

## §2. Scope

In scope:

- Every non-deprecated stakeibc host zone: celestia, cosmoshub-4, dydx-mainnet-1,
  haqq_11235-1, injective-1, juno-1, laozi-mainnet, osmosis-1, phoenix-1, sommelier-3, ssc-1.
- Staketia (stTIA redemptions and the 5-of-7 Celestia multisig), including its 7 open
  unbonding records. Its TIA joins the same stTIA pool as the stakeibc Celestia TIA.
- The stakeibc user redemption records open at upgrade 1 (99 on 2026-09-18) and any created
  before the redeem message is removed. All are flushed and claimed in window 1.
- Stakedym's 11 open unbonding records (flushed by its operator in window 1), and halting
  stakedym at upgrade 2.
- Removing every message the protocol no longer needs, moving wasm control to gov, and
  removing every IBC rate limit at upgrade 2.
- Four admin txs added at upgrade 2: per-validator undelegation, ICA-to-Osmosis transfer,
  the staketia claim-address transfer,
  and the batched stToken sweep to Osmosis.
- The Osmosis side as an ops procedure (§7): one transmuter per stToken, its assets, its
  funding, and the coverage check that gates it. No Stride code is involved.

Out of scope, explicitly:

- The deprecated zones comdex-1, evmos_9001-2, stargaze-1 and umee-1. Apart from upgrade 1
  setting the `Deprecated` flag on comdex-1 (the other three already carry it), no upgrade
  touches them, and none of them are unbonded, drained or given a pool.
  Their host chains are halted, so nothing could be done through their ICAs anyway. Stakedym
  is halted at upgrade 2 once its operator has flushed its records, and is likewise never
  unbonded, drained or given a pool. Holders of stCMDX, stEVMOS, stSTARS, stUMEE and stDYM have
  no redemption path after this work, which matches their status today but is now a
  deliberate decision.
- stTokens held by Stride accounts that no key controls (contracts, interchain accounts owned
  by other chains, module accounts including the community pool). They are not swept; their
  owners, where there is one, move them out themselves before the halt (§8). Their backing
  stays in the pools.
- Removing whole modules or their state. Only messages are removed; stores stay.
- STRD-side modules beyond message removal (mint, strdburner). What happens to STRD, the
  validator set and the chain's history after the halt is a separate decision.

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
  except the trade controller on Osmosis. The amounts are immaterial: sampled 2026-09-21, the
  ICA accrues 0.05 to 0.1 USDC per hour (1 to 3 USDC per day on 406k DYDX, 0.14% of dYdX's
  bonded stake), so the whole wind-down is on the order of 100 USDC. The delegator-shares
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
  address. 21 contracts are instantiated (all Hyperlane): 17 have no admin and 4 have the
  Stride deploy key `stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh` as admin. Deployed
  contracts are executable by anyone.

Pipelines that already exist:

- Since v34 the undelegate ICA callback accepts a submission with no epoch unbonding record
  ids: on success it only decrements the validator and host zone delegation balances and burns
  nothing (`icacallbacks_undelegate.go`). `BatchSubmitUndelegateICAMessages` takes
  per-validator `MsgUndelegate`s directly and flags `DelegationChangesInProgress` on each, and
  `applySharesRoundingSafety` trims a full-drain amount so share truncation cannot make it
  exceed the on-chain delegation. Neither depends on the zone being active.
- Every host zone has eight ICA types: delegation, fee, withdrawal, redemption, community-pool
  deposit and return, and the two trade-route converter accounts (`ICAAccountType`). The dYdX
  withdrawal ICA holds ~3.8 USDC beside its DYDX. The community-pool ICAs (the feature was
  disabled in v28) hold single-digit base units on every zone, plus 50 USDC in dYdX's deposit
  ICA (checked 2026-09-22); they are written off along with the converter ICAs.
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
- An ICA-wrapped IBC `MsgTransfer` from a host account to a third chain exists in the
  trade-route code (`x/stakeibc/keeper/reward_converter.go`, `BuildHostToTradeTransferMsg`
  sends from the withdrawal ICA on the host to Osmosis via Noble). Autopilot and staketia
  already call `transferKeeper.Transfer` from keeper code with a sender that is not the tx
  signer, so a module can move a user's balance over IBC without a signature from them.
- Every admin tx checks its signer against the hard-coded admin set (`utils.Admins`: the F5
  key and the gov module) in ValidateBasic (`utils.ValidateAdminAddress`).
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
  `stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd`, and an empty `allowed_packet_data`, which
  ibc-go enforces as "memo must be empty", so the operator cannot attach a forwarding memo.
  Bank `MsgSend` is not granted, and the 5-of-7 signers are no longer reachable, so the
  operator's authz route is the only way to move the multisig balance, and it can only land
  on the claim address as a TIA voucher on Stride. The claim address (labelled S2 in
  `x/staketia/types/celestia.go`, from the January 2024 launch) is itself a 5-of-7 multisig
  `BaseAccount`, with a signer set that shares no key with the Celestia multisig, and it has
  signed exactly one tx in its life (sequence 1). Whether five of its signers can be gathered
  is unknown; the design does not depend on it. Staketia's keeper already moves funds out of
  a multisig `BaseAccount` without a signature: its deposit address (S0, the same signer set as
  the Celestia multisig) is IBC-transferred from keeper code via `transferKeeper.Transfer`
  with the deposit address as sender (`x/staketia/keeper/delegation.go`), which is the
  pattern the claim-address tx reuses. Staketia's IBC middleware only acts on packets it
  recorded by sequence (its own delegation transfers), so an inbound transfer to the claim
  address and an outbound one from it are both inert to it.
- Native vouchers stranded on Stride are dust (2026-09-22, with liquid staking still live):
  the eleven deposit addresses, the reward collector and the auction module together hold
  about $3 of in-scope native denoms. Nothing on Stride except the staketia claim address
  will hold a native balance worth moving.

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

Where the stTokens are (bank `denom_owners` on 2026-09-22, in-scope denoms):

- Between 15% and 85% of each stToken's supply is held on Stride outside IBC escrow (55% of
  stATOM, 719k of 1.30M). The rest sits in the escrow accounts of 3 to 30 transfer channels
  per denom, i.e. it lives on other chains as vouchers; for stATOM channel-5 (Osmosis) holds
  456k, channel-0 (Cosmos Hub) 62k, channel-6 (Injective) 31k, and 27 channels hold the rest.
  The per-channel escrow balance is exactly how much of that stToken is on that counterparty
  chain.
- Holder counts are dominated by dust: 567k (account, denom) pairs hold a non-zero balance,
  5.7k hold at least $10 worth and 1.1k at least $100 (prices of 2026-09-22). The $10 floor
  strands about $58k of $3.09M; the $100 floor about $193k.
- The largest holders are ordinary 20-byte accounts, several of them Stride vesting accounts
  (`StridePeriodicVestingAccount`, `ContinuousVestingAccount`; stTokens are not their vesting
  denom, so they are freely transferable). Holders with no owner on Osmosis: the transfer
  escrow accounts; the zones' deposit addresses (in-flight redemptions, 30k stATOM, all
  flushed in window 1); the distribution module (15.5k stATOM, mostly unwithdrawn validator
  commission and STRD staking rewards; the community pool proper holds about $17k of
  stTokens across every denom); interchain accounts owned by other chains (2.3k stATOM in
  one); and a handful of 32-byte contract-style accounts (7k stTIA in one).
- Stride and Osmosis both derive addresses from secp256k1 keys on coin type 118, so the
  Osmosis address of a 20-byte Stride account is the same bytes with the `osmo` prefix, and
  the same key controls both. A multisig address is the hash of its member keys and threshold
  and derives the same way. 32-byte addresses (contracts, interchain accounts, module-derived
  accounts) have no controlling key and no counterpart.

Transmuter (osmosis-labs/transmuter v3.2.0, code id 996 on osmosis-1; source checked
2026-09-22):

- A transmuter is a `cosmwasmpool` that swaps its assets at a fixed ratio. v3 ("alloyed")
  gives each asset a `normalization_factor` (any `Uint128`); the rate between two assets is
  the ratio of their factors. A larger factor makes a unit of that token worth
  less (`out = in × out_factor / in_factor`), so setting the stToken factor to `1e18` and
  the native factor to `RedemptionRate × 1e18` encodes the 18-decimal frozen rate exactly;
  the alloyed asset takes the native factor so one alloyed unit is one native base unit.
  (Corrected 2026-09-23; an earlier draft had the two factors swapped.) Swap fee is hard-coded
  to zero; exact-in output rounds down, exact-out input rounds up.
- Relative factors cannot be changed after instantiation. The admin's only factor operation,
  `rescale_normalization_factor`, multiplies every factor by the same ratio. The admin can
  `add_new_assets` later, each with its own factor, provided the denom already has bank
  supply on Osmosis. The v1/v2 non-alloyed transmuter has no factors and swaps 1:1 only.
- Instantiation is permissionless via `MsgCreateCosmWasmPool { code_id, instantiate_msg,
  sender }` with a whitelisted code id and the normal pool-creation fee. The instantiate
  message is `{ pool_asset_configs: [{denom, normalization_factor}], alloyed_asset_subdenom,
  alloyed_asset_normalization_factor, admin, moderator }`; admin and moderator are plain
  addresses and can be multisigs. The pool starts empty. `join_pool` accepts funds in any
  subset of the assets and mints the alloyed asset (a tokenfactory denom) to the depositor as
  the LP receipt; `exit_pool` burns it and returns a pro-rata share of whatever the pool
  holds. Swaps go through the normal poolmanager routes.
- Limiters are opt-in per denom (`register_limiter`); with none registered, the pool can move
  from 100% native to 100% stToken. The moderator can `set_active_status` (freeze) and
  `mark_corrupted_assets`; the admin can transfer adminship in two steps. Contract code
  migration is governance-only (the wasm admin is the cosmwasmpool module).

## §3a. Operator addresses

Three addresses run the wind-down. Each has one name used everywhere in this spec, the plan
and the code, and each is a hard-coded constant in the upgrade 2 binary (the protocol admin
already is). Nothing on Stride receives native tokens: every native balance leaves from a
host chain, so there is no Stride-side vault.

| Name | Chain | Type | Status | Constant | Role |
|---|---|---|---|---|---|
| Protocol admin | Stride | key (F5) and the gov module | exists, `utils.Admins` | `utils.Admins` | Signs `MsgUndelegateFromValidators`, `MsgTransferFromIca`, `MsgTransferStaketiaClaimBalance`, and the two admin-gated ICQ messages. |
| Sweep operator | Stride | new key | to create | `SweepOperatorAddress` | The only address that can sign `MsgSweepStTokens`. Separate from the protocol admin so the sweep, the one tx that moves user balances, has its own key and its own blast radius. Holds STRD for fees only. |
| Osmosis vault | Osmosis | new multisig | to create | `OsmosisVaultAddress` | Receives every ICA transfer, instantiates and funds the pools, holds the alloyed assets, and is each pool's admin and moderator. |

The sweep operator and the Osmosis vault are created in window 1, proven before the upgrade 2
PR is cut (a signed spend from each), and their addresses go into the binary as constants (§8).
The Osmosis vault's admin and moderator roles on the pools can be split to a second multisig
later with `assign_moderator`; that is an ops choice, not a design point.

## §4. Overview

| | New logic | Removes | Ops window after |
|---|---|---|---|
| Upgrade 1: close the doors | none (a Haqq delegation delta table, v34 pattern) | liquid stake, redeem, and every create-things message; the trade route; wasm to gov | ~35 days: flush unbondings, claim for everyone, operators finish staketia/stakedym |
| Upgrade 2: halt, unbond, migrate | four admin txs, one ValidateBasic gate | claim, rebalance, clear-balance, resume, oracles, all rate limits, calibration cap | ~35 days: refresh slashes, undelegate every validator, send every ICA balance to Osmosis, build and fund the pools, sweep stTokens, halt |

Each upgrade lands on a chain with nothing in flight, verified by the checklist that gates its
proposal (§8). The halt is a coordinated `halt-height`, not an upgrade.

## §5. Upgrade 1: close the doors

No new logic. Only message removals and parameter changes.

Remove message handlers (proto `rpc`, msg server, amino registration, CLI, tests; keeper
functions stay wherever something still calls them, so stakeibc's `LiquidStake` and
`RedeemStake` keeper paths survive for the reward collector, the community pool and autopilot,
while staketia's and stakedym's redeem keeper paths go with their handlers):

- stakeibc: `LiquidStake`, `LSMLiquidStake`, `RedeemStake`, `RegisterHostZone`,
  `CreateTradeRoute`, `UpdateTradeRoute`, `DeleteTradeRoute`, `SetCommunityPoolRebate`,
  `ToggleTradeController`.
- staketia and stakedym: `RedeemStake` (their `LiquidStake` is already a disabled stub; remove
  the stub too).
- icaoracle: `AddOracle`, `InstantiateOracle`.
- icqoracle: `RegisterTokenPriceQuery`, `RemoveTokenPriceQuery`.
- auction: `PlaceBid`, `CreateAuction`, `UpdateAuction`.
- airdrop: all seven. claim (legacy): all four.

The message types stay registered in the interface registry. A node must still be able to
decode every historical transaction that contains them (`strided q tx`, the tx REST endpoints);
dropping the registration would break that for most of the chain's history. Rejection happens
one step later: with no handler in the msg service router, a submitted message fails with
"can't route message" for every entry path that goes through the router, which is a stronger
block than a flag. Two entry points do not go through the router and are closed separately:

- Autopilot: handler sets `StakeibcActive = false`.
- ICA host: handler removes `MsgLiquidStake` and `MsgRedeemStake` from the allow-list.
  `MsgClaimUndelegatedTokens` stays until upgrade 2 so ICA-originated claims work in window 1.

Wasm: handler sets `code_upload_access` to the gov module address only, and for each of the
four contracts whose admin is the Stride deploy key, sets the admin to the gov module address
through the gov permission keeper (whose authorization policy allows the change without the
current admin's signature). The upload-access write is the one upgrade 1 step that fails the
upgrade on error rather than logging: it can only fail on invalid params, and leaving upload
open to the two keys with nothing but a log line is the worse outcome. A contract that no
longer exists is logged and skipped. The list is re-checked right before the proposal.

Trade route: handler deletes the dYdX trade route. Its conversions are worth 1 to 3 USDC a
day (§3), so stopping them a month early costs nothing, and deleting it here removes the
off-chain trade controller from the picture and the need to prove the route quiet before
upgrade 2. The USDC that accumulates in the dYdX withdrawal ICA during window 1 is swept in
window 2 as surplus (§6, `denom`). The converter ICA addresses are noted in the plan; the
stale authz grant on the Osmosis trade ICA is harmless.

Comdex: handler sets `Deprecated = true` on comdex-1 so it carries the same flag as the other
three deprecated zones. `Halted` is not touched, per the decision to leave deprecated zones as
they are; the flag is documentation.

Haqq delegation reconciliation: apply a per-validator delta table to haqq_11235-1 with the v34
helper, exactly as v34 did for Injective. The 2026-09-21 measurement (§8a) has 17 validators
off: 14 over-recorded (undetected downtime slashes and sub-token rounding, the largest
853.8 ISLM) and 3 under-recorded by sub-token dust (the chain holds slightly more than
tracked). Both signs are applied so every tracked delegation equals the chain's; the net is a
decrease of about 1,758 ISLM, so `TotalDelegations` drops and the next epoch's rate update,
which still runs in window 1, lowers the stISLM redemption rate to match. That is the correct
outcome for a slash that was never detected, and the table matched live state on three
measurements over two days. The stored
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
reward-collector fee liquid stake, the oracles, and the staketia/stakedym operator flows. The fee liquid stake mints stTokens for validators during the window; it is
accounting-consistent and stops at upgrade 2, so it is not worth a switch.

## §6. Upgrade 2: halt, unbond, migrate

Handler:

1. `Halted = true` on every in-scope stakeibc zone and on stakedym's host zone. The four
   deprecated zones are not touched.
2. Deactivate the three ICA oracles (existing toggle logic). Remove
   `MsgClaimUndelegatedTokens` from the ICA host allow-list.
3. Remove every rate limit, every blacklisted denom and every whitelisted address pair from
   the rate limiter. The stToken sweep (below) sends most of each stToken's on-Stride supply
   out over one channel in a few days, which no limit would allow, and there is no longer a
   mint path that a limit would protect. The module and middleware stay in the stack with
   empty state.

The halt strands whatever native vouchers sit in the deposit addresses, the reward collector
and the auction module at that height. Liquid staking stopped at upgrade 1, so by then these
are the last mint epoch's fee liquid stake and rounding dust (about $3 today, §3); the
handler does not touch them and they are written off.

Remove message handlers: stakeibc `ClaimUndelegatedTokens`, `RebalanceValidators`,
`ClearBalance`, `ResumeHostZone`; staketia and stakedym `ResumeHostZone`. Add
`utils.ValidateAdminAddress` to the ValidateBasic of `UpdateValidatorSharesExchRate` and
`CalibrateDelegation`; ops keep them to refresh slashes before the unbond (§8), but nobody
else can trigger the slash path that rewrites the rate. Remove the 5,000 base-unit
`CalibrationThreshold` check from the calibration callback: it existed to bound what a
permissionless caller could move, and the message is admin-gated from here on.

All four new txs are gated in ValidateBasic: the undelegate, ICA transfer and claim-address
txs on the protocol admin (`utils.ValidateAdminAddress`), the sweep tx on the sweep operator
(§3a). None reads or writes host zone accounting; amounts and lists are ops inputs.

New admin tx `MsgUndelegateFromValidators { creator, chain_id, validators: [{address, offset}] }`.
An empty `validators` list means every validator on the zone with
a non-zero stored delegation; otherwise only the listed ones. Per validator the amount is the
stored delegation minus `offset` (default zero), passed through `applySharesRoundingSafety`,
and rejected if it is not positive. No blanket offset is needed for the real run: an unslashed
validator has an exchange rate of exactly one, so a full-drain amount converts to shares with
no truncation, and the helper already buffers full drains of validators whose rate is below
one. Both rely on the stored rate matching the chain, which the window-2 refresh guarantees.
The per-validator `offset` is only the lever for a validator that has drifted since. A listed
validator with `DelegationChangesInProgress` set is rejected, so a batch cannot be
double-submitted while its ack is outstanding. The messages go through
`BatchSubmitUndelegateICAMessages` with no epoch unbonding record ids, so the existing
callback decrements the balances and nothing is burned.

Submitting by validator rather than cascading a zone amount by weight is what lets ops test on
a single validator first, skip or shave a validator that is failing or drifted by dust, and
keep every submitted amount equal to what the host will accept. Retry is resubmission by ops;
there is no queue and no epoch hook. ICA txs are atomic, so one over-recorded validator still
fails its whole batch; the slash refresh (§8, step 1) is what prevents that, and `offset` is
the manual lever if a validator is still off by dust.

New admin tx `MsgTransferFromIca { creator, chain_id, ica_type, amount (Coin) }`:
`ica_type ∈ {DELEGATION, WITHDRAWAL, FEE, REDEMPTION}`, the four ICAs that hold anything (the
two community-pool ICAs hold dust and the two converter ICAs belonged to the trade route
deleted at upgrade 1; all four are written off). The destination is not a tx input. The
receiver is the Osmosis vault (§3a), one hard-coded constant in the upgrade 2 binary, and the
host-side transfer channel to Osmosis comes from a hard-coded map `chain_id → channel_id`
covering the ten non-Osmosis in-scope zones; a `chain_id` absent from the map is rejected.
Both are reviewed and tested in the upgrade PR (§10) and verified against the hosts before
the proposal (§8), so a fat-fingered channel or receiver on the day is not possible; the only
per-tx inputs are the zone, the ICA and the amount. `amount` carries its denom as it exists
on the host, so foreign balances such as the USDC in the dYdX withdrawal ICA can be sent too.
Submits one ICA containing an ICS-20 `MsgTransfer` of `amount` from that ICA over the mapped
channel to the Osmosis vault, built like `BuildHostToTradeTransferMsg` without the
forwarding memo, with the existing ICA timeout and a one-day transfer timeout. No callback
state: a failed or timed-out transfer refunds to the ICA on the host and ops resubmit.

The tokens go from the host straight to Osmosis, never through Stride, because that is the
only route that lands them as the canonical denom on Osmosis: a voucher that reaches Stride
first and is forwarded from there arrives as a two-hop denom that no pool would use. For the
osmosis-1 zone the ICA is already on Osmosis and the tx is an ICA bank `MsgSend` to the
Osmosis vault instead (osmosis-1 is in the map with an empty channel, which selects this
form).

New admin tx `MsgTransferStaketiaClaimBalance { creator }`, with no other field. It moves the
staketia claim address's whole TIA voucher balance to the stakeibc celestia zone's delegation
ICA on Celestia, where it unwinds to native TIA and is then sent on to Osmosis by
`MsgTransferFromIca DELEGATION` with the rest of the zone's balance. Everything is a constant
or read from state: sender is staketia's `ClaimAddress`, denom is
`CelestiaNativeTokenIBCDenom`, amount is the full balance (reject if zero), channel is the
celestia host zone's `TransferChannelId`, receiver is its `DelegationIcaAddress`, one-day
timeout, no memo. It calls `transferKeeper.Transfer` with the claim address as sender, the
pattern staketia's keeper already uses for its deposit address (§3). It is the route
regardless of whether the claim address's own 5-of-7 can sign, so no multisig coordination is
on the critical path. A timeout refunds the claim address and ops resubmit.

The same trick, sending a Stride-side voucher to the zone's delegation ICA so it leaves with
the ICA balance, is why nothing else needs a Stride-side account or a hop through a host
chain; the only Stride-side balance worth it is this one.

New admin tx `MsgSweepStTokens { creator, denom, addresses: [string] }`, the batched stToken
sweep. `denom` must be the stToken denom of a stakeibc host zone; `addresses` is non-empty
and at most 100 entries (the batch bound that keeps a tx inside the block gas limit; the plan
measures the real cost and can raise it). For every listed address, in order:

- Reject the whole tx if the address is not sweepable: it must decode to 20 bytes and its
  account must be a `BaseAccount` or one of the vesting account types; transfer escrow
  addresses, module accounts (including the deposit addresses, the distribution module and
  the reward collector) and interchain accounts are rejected. Batches are built off chain
  from an export, so a rejected address is an ops error worth surfacing, and the reject is
  what makes the on-chain rule the safety net rather than the script.
- Skip with an event if the balance of `denom` is zero (the holder moved it between the
  export and the tx).
- Otherwise submit an ICS-20 `MsgTransfer` through `transferKeeper.Transfer` of the full
  balance, sender the holder's Stride address, receiver the same 20 bytes bech32-encoded with
  the `osmo` prefix, over the Stride to Osmosis transfer channel (channel-5, a constant), with
  a one-day timeout and no memo. A timeout or a rejected receive refunds the holder on Stride
  through the normal ICS-20 path and ops resubmit that address.

One tx sweeps up to 100 holders of one denom; the whole sweep at the chosen floor is a few
thousand packets over a few days (§3). There is no floor on chain: only the sweep operator can
sign the tx, so nobody can spam it, and the floor is an ops choice made from prices on the day, stated in the
announcement. Every account that clears the floor is swept; holders below it, and holders on
other chains, move themselves (§7).

## §7. Osmosis side: pools, funding, foreign-route denoms, shutdown

Nothing in this section is Stride code. It is the procedure ops follow in window 2 as the
native tokens arrive, and it is what replaces the on-chain withdrawal mode.

One transmuter pool per in-scope stToken, eleven pools, instantiated from code id 996 (v3.2.0)
by the Osmosis vault (§3a). Each pool's initial assets are the
canonical stToken denom on Osmosis (the one minted by transfers over Stride's channel-5) and
the native token, with normalization factors `1e18` for the stToken and
`HostZone.RedemptionRate × 1e18` for the native token and the alloyed asset (§3: a larger
factor is a cheaper unit), the rate read from Stride at instantiation and frozen since
the upgrade 2 halt (§9). Admin and moderator are the Osmosis vault (§3a); the moderator's
freeze is the incident lever. Adminship cannot be renounced (a transfer only completes when
the candidate claims it), so the vault keeps it for the life of the pools. The alloyed asset
each pool mints is the LP receipt and stays in the Osmosis vault; it is the withdrawal key
to the pool's backing and is custodied like the backing itself.

Limiters: one `static_limiter` per foreign-route denom, registered right after the funding
join, with `upper_limit` = that route's share of the stToken's supply in the escrow snapshot
at the halt × 1.1, rounded up to the next 0.0005. None on the canonical stToken or the
native token. A limiter bounds a denom's share of pool value and only blocks moves that
raise it; because swaps leave the pool's value unchanged, the cap is a ceiling on how much
of that route the pool will ever absorb. Stride's per-channel escrow is a hard upper bound
on the genuine amount a route can deliver, so the cap costs honest holders nothing and
bounds the damage from a compromised source chain or light client minting counterfeit
two-hop vouchers to that route's cap. A denom's last limiter can never be deregistered,
only widened, so if the vault ever exits native tokens (reclaiming unclaimed backing) it
widens the caps first. The canonical denom is left uncapped because it is ~90% of supply;
its counterfeit risk is a forged Stride header, handled in the halt checklist (§8).

Funding is one `join_pool` per pool with native tokens only, for exactly the amount the
coverage check requires (§9), once every source for that denom has arrived: the delegation
ICA transfer and the withdrawal, fee and redemption ICA sweeps (for stTIA the delegation ICA
transfer includes the former multisig balance, routed through the claim address). A pool can be created and funded as soon as its
own denom is complete; zones finish unbonding on different days and nothing couples them.

Foreign-route denoms: a stToken that left Stride to chain X and is sent from X to Osmosis
arrives as a two-hop denom (`transfer/<osmosis-X channel>/transfer/<X-stride channel>/st...`),
distinct from the canonical one. Rather than route it back through Stride, which is
impossible after the halt, each such denom is added to the stToken's pool with the canonical stToken's
normalization factor (`1e18`) via `add_new_assets`, after which it swaps at the same rate
against the native token and 1:1 against every other route of the same stToken. Adding
denoms with an already-present factor leaves the factor lcm, and so the overflow headroom,
unchanged (§3). The
per-channel escrow balances on Stride (§3) list exactly which chains hold which stToken and
how much, so the set of denoms to add is known before the halt. The contract requires a denom
to have supply on Osmosis before it can be added, so ops seed each one with a small transfer
from that chain (or add it after the first user's transfer lands). Third-hop denoms are added
on request the same way. Osmosis users of DeFi protocols on other chains withdraw there and
transfer to Osmosis directly.

Shutdown: once every pool is funded, the sweep is complete and the halt checklist (§8) passes,
validators set a `halt-height` and the chain stops. Stride's IBC clients on other chains expire
after their trusting periods; stTokens on those chains stay ordinary vouchers and keep working
in the pools, and any packet toward Stride that was never relayed times out and refunds on its
source chain. The pools outlive the chain. Backing that is never claimed is reclaimed later by
exiting the pools with the alloyed asset; that timing is a policy decision outside this
design.

## §8. Ops windows and checklists

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
5. Create the sweep operator key and the Osmosis vault (§3a) before the upgrade 2 PR is cut,
   because the binary hard-codes both; gather the host-side
   channel id to Osmosis for each of the ten non-Osmosis zones, each verified by querying the
   host's channel and confirming its client's counterparty chain id is osmosis-1 and the
   channel is open; and the list of foreign-route denoms per stToken from the escrow
   balances.
6. Announce the timeline, the sweep floor and date, and that holders on other chains transfer
   to Osmosis directly.

Checklist to propose upgrade 2:

- Zero stakeibc user redemption records; zero `HostZoneUnbonding` records outside `CLAIMABLE`
  with a non-zero amount; no claim ICA in flight.
- Staketia: zero unbonding records outside `CLAIMED`; zero redemption records; claim address
  empty.
- Stakedym: same, for its records.
- No open LSM token deposit on cosmoshub-4.
- The two new address constants and the channel map in the binary re-verified: the map
  against the hosts, each address by a test transfer of a few tokens to it from any wallet
  followed by a signed spend from it, so every constant is proven to be an address we control
  before anything is sent there.

Window 2 (after upgrade 2, ~35 days):

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
   (or set that validator's `offset`) and resubmit for the affected validators. Delegation ICA
   channels and relayers stay healthy until every batch acks; a dead channel is restored with
   the existing flow and the affected validators are resubmitted.
3. Day 0+: `MsgTransferFromIca` for withdrawal, fee and redemption ICA balances, including
   foreign denoms such as the dYdX USDC. This is the live test of the
   transfer tx on small real amounts, and the first arrival on Osmosis confirms the mapped
   channel end to end and the denom each zone lands as.
4. Staketia (the Celestia multisig signers are no longer reachable, so everything goes through
   the operator's authz grant and Stride): day 0, the operator undelegates the entire multisig
   delegation on Celestia via authz alongside the stakeibc unbonds. Day 21, the operator IBCs
   the whole liquid balance via authz to the claim address (no memo; the grant forbids one).
   Then `MsgTransferStaketiaClaimBalance` moves it to the celestia delegation ICA, where it
   waits for step 5 with the zone's own unbonded balance. No key of the claim address signs
   anything.
5. As each zone's unbonding completes (day 14 to day 30): `MsgTransferFromIca DELEGATION` for
   the full balance, then `WITHDRAWAL` again (undelegation auto-withdraws accrued rewards
   there). Then create and fund that stToken's pool (§7) once the coverage check passes (§9),
   and add its foreign-route denoms.
6. Last days: `MsgSweepStTokens` in batches of up to 100, per denom, for every holder at or
   above the floor, built from a fresh export. Resubmit any address whose transfer timed out
   (its balance is back on Stride). Relayers on channel-5 stay up until the last packet acks.
7. Transfer-channel relayers stay up until the halt. ICA channels can be left to close once
   every balance is sent.
8. After the halt: every validator rotates or destroys its consensus key, and Stride Labs
   confirms it in writing from each. Osmosis's light client of Stride (`07-tendermint-2119`,
   12-day trusting period) accepts any header signed by two thirds of the last trusted
   validator set until it expires; a forged header could mint canonical stToken vouchers on
   Osmosis, which the pools would honour. Keys gone means the window is closed on day 0
   rather than day 12. No relayer is asked to update that client after the halt.

### §8a. Drift measurement

Measured 2026-09-21 for every in-scope zone except cosmoshub-4 and injective-1 (handled by
v34): recorded per-validator delegations versus the delegation ICA's on-chain balances.
dydx-mainnet-1, ssc-1 and sommelier-3 are exact. celestia is under-recorded by ~15,440 TIA
(the v34 phantom stake; harmless for unbonding). osmosis-1 has one validator over by 3,026
uosmo with an unchanged rate (calibration range). juno-1 (1 validator, 1.94 JUNO),
laozi-mainnet (2, 4.53 BAND), phoenix-1 (10, max 1.20 LUNA) and haqq_11235-1 (14 over plus 3
under by dust, max 853.8
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

Checklist to halt the chain:

- No undelegate batch in flight (no validator with `DelegationChangesInProgress`) and
  `TotalDelegations` at dust on every zone except celestia, where it still carries the
  multisig portion (the per-validator unbond only touches ICA validators).
- All four ICA balances at dust on every zone. Celestia multisig delegation zero, its
  unbondings complete, its balance on Osmosis.
- Every pool created, funded and passing the coverage check (§9) against a fresh export, with
  its known foreign-route denoms added.
- The sweep complete: no sweepable account at or above the floor holds any in-scope stToken,
  and no sweep packet outstanding on channel-5.
- Validators and STRD delegators have withdrawn their rewards; interchain-account holders
  have been notified and given time to move out.

## §9. Accounting

There is no rate calculation anywhere in the design. The rate for a zone is the
`HostZone.RedemptionRate` that was on chain when upgrade 2 halted the zone; it is read from
Stride when the pool is instantiated and becomes the pool's normalization factors, which
cannot be changed relative to each other afterwards (§3). Nothing on Stride recomputes it:
the epoch update is halt-gated, and the only other writer (the delegator-shares slash
callback) is reachable only through the two admin-gated ICQ messages. If ops deliberately run
those in window 2 and a slash is found, the frozen rate is lowered accordingly, which is the
correct direction, and the pool must be instantiated after that refresh.

Coverage check, per stToken, run from a fresh Stride export before the pool is funded and
again before the halt: native tokens held on Osmosis for that denom ≥ Stride bank supply of
the stToken × `HostZone.RedemptionRate`. Bank supply is the right reference: stTokens that
left Stride over IBC are escrowed here, not burned, so they are counted, and they are exactly
the foreign-route holders the pool must also serve; stTokens burned by flushed redemptions are
gone from both sides. The check is a script over two queries and has no on-chain counterpart,
which is the one place this design is weaker than an on-chain assertion; the mitigation is
that funding is a deliberate `join_pool` for the computed amount, so an under-funded pool can
only come from a wrong number in a script that is run twice and published.

Why a frozen rate is covered. Confirmed against cosmos-sdk v0.54.3:
`Unbond` calls the distribution hook `BeforeDelegationSharesModified`, which withdraws the
accrued rewards, and then removes the shares; rewards are computed from delegation shares only,
so an unbonding entry earns nothing during its 21 to 30 days. Between the halt and the pool
being funded, the backing therefore moves as follows:

- Up: staking rewards accrue only from the halt until the undelegate executes at the next day
  epoch (one or two days, roughly 0.03% to 0.05%), are auto-withdrawn to the withdrawal ICA by
  the undelegation, and are sent to Osmosis. Rewards withdrawn but not yet reinvested at the
  halt are also sent.
- Down: any slash during window 2.

Coverage then follows by construction. The refresh (§8, step 1) makes the recorded
delegations and the redemption rate reflect every slash that has happened, the full recorded
amount unbonds, and Osmosis receives it plus the rewards swept on undelegation. Supply ×
frozen rate is the delegated amount, so the rewards are the buffer. Only a slash between the
refresh and the undelegation, on a zone whose retried batch never lands, can leave a zone
under, and the coverage check catches that before funding; the shortfall is then covered from
the surplus of the other denoms or by Stride Labs, or the last swappers absorb it, which §1
accepts.

The stToken sweep does not change any of this: it moves stTokens from Stride accounts into
escrow, so bank supply is unchanged and the swept tokens become canonical-denom holders on
Osmosis. Unsweepable holdings (§2) stay in supply, are covered by the funding, and their
backing is what the pools hold at the end.

## §9a. Validator compensation

Today POA validators receive two streams: STRD from mint provisions (the staking share of
~58 STRD per hour epoch, about 16%) plus tx fees, and stTokens from the reward collector (15%
of Stride's commission on host staking rewards, liquid staked at the mint epoch).

The stToken stream is unaffected in window 1 (everything keeps running) and ends at upgrade 2,
not because of the halt but because there are no staking rewards to share once the delegations
are gone. The rewards that accrue during window 2 go to Osmosis with the rest and become
pool surplus rather than being split. The STRD stream continues unchanged until the halt. If
validator pay needs to rise to compensate, that is a mint distribution-proportion parameter
change by governance and needs no code. Validators withdraw their accumulated commission and
rewards (including stTokens, which they then move to Osmosis themselves) before the halt (§8).

## §10. Testing

- Upgrade 1: handler tests against a mainnet export (`app/upgrades/vN/testdata/`, v34-style)
  for the trade route deletion, the comdex-1 `Deprecated` flag, the Haqq delta table (applied,
  and skipped on a stale constant), the haqq slash-query purge (deletes only that chain's
  slash-path queries, clears the validator flags, leaves other chains' and the
  withdrawal-balance queries), the autopilot param, ICA host allow-list, wasm params and
  contract admins; a compile-time guarantee that the removed messages no longer exist;
  existing keeper tests for the flows that keep running stay green.
- Upgrade 2, admin txs: unit tests for gating on all three; for the undelegate tx, per-validator
  message construction with and without offsets, empty versus explicit validator lists,
  rounding safety on a full drain, rejection of a validator with a change in progress, no
  accounting mutation; for the transfer tx, every ICA type, a foreign denom, the osmosis-1
  bank-send form, a chain id absent from the map, that the map has an entry for every
  in-scope zone and none for a deprecated one, that the Osmosis vault constant parses as an
  `osmo` bech32 address and the sweep operator constant as a `stride` one, and the built
  `MsgTransfer` fields (mapped channel, Osmosis vault,
  timeout, empty memo); for the sweep tx, the highest-review item: table-driven tests for a
  base account, each vesting type, an escrow address, a module account, an interchain account,
  a 32-byte address, an unknown account, a zero balance (skipped, others in the batch still
  sent), a batch over the bound, a non-stToken denom, the derived `osmo` address bytes, the
  full balance and only that denom being sent, and the ICS-20 refund on timeout returning the
  balance to the holder; for the claim-address tx, the full balance and only the TIA denom
  being sent, the built `MsgTransfer` fields (celestia channel, delegation ICA receiver,
  timeout, empty memo), a zero balance rejected, and the refund on timeout landing back on
  the claim address.
- Upgrade 2, handler: the halt flags, oracle deactivation, rate-limit removal, the two
  ValidateBasic gates and the lifted calibration cap. Localstride run:
  upgrade, then one undelegate and its ack, one ICA transfer to a second local chain, and one
  sweep batch whose packets are relayed and land at the derived addresses.
- Ops scripts: the coverage check and the batch builder are tested against a mainnet export
  and their output for the export is checked in beside the plan.

## §11. Open items for the plan

- **Band's light client of Stride is expired** (laozi-mainnet `07-tendermint-169` on the ICA
  connection `connection-146`, last header 2026-08-05; the delegation ICA restore is stuck
  in `STATE_INIT` on channel-768). No ICA tx, and no Stride→Band transfer, can be delivered
  until Band governance recovers it with `MsgRecoverClient` and a fresh substitute client,
  so the Band zone (~$110k stBAND) cannot be unbonded in window 2 without that proposal.
  Raise it with the Band team now; it has to pass before upgrade 2. Every other in-scope
  zone's host-side client of Stride is active (checked 2026-09-23).
- Stranded stToken holders behind expired clients, to decide on: Penumbra (Osmosis's and
  Stride's clients both expired; ~6.2k stATOM and stOSMO, ~$29k) can be reopened by an
  Osmosis `MsgRecoverClient` or a new channel; Kujira (~5.8k stATOM, ~$20k) has no
  reachable RPC and looks stopped, so it is ignored (decided 2026-09-23). Agoric is
  fine for holders (Agoric→Osmosis is active) though its Stride hop is expired. The five
  deprecated zones and their stTokens (~$16.8k in total) are not touched by the migration:
  Evmos, Stargaze, Umee and Comdex have stopped producing blocks (unrecoverable, like Kujira),
  and Dymension is alive but left alone by choice. `docs/wind-down/sttoken-locations.md`
  separates ignored balances into small, unrecoverable and deprecated.
- Relayers: during window 1 users redeem through Stride, which needs both clients alive on
  every stToken chain ↔ Stride pair; Neutron's are active but nobody is updating them (35 h
  old on 2026-09-23), so ops relay that pair. In window 2 the host→Osmosis and
  Stride→Osmosis transfers are relayed by ops where public relayers are absent. After the
  halt, holders relay their own chain→Osmosis hop if nobody else does. The live map is at
  https://claude.ai/artifact/986V5LAXxFgzjq7jXPpE8r.
- Exact proto shapes and enum names for the four admin txs; the constants: the two new
  addresses in §3a once created, the channel-5 constant for the sweep and the
  `chain_id → host-side channel to Osmosis` map; the batch bound after measuring gas.
- Version numbers for the two upgrades.
- Identify the owners of the interchain accounts on Stride that hold stTokens (2.3k stATOM in
  one) and the 32-byte holders, and notify them.
- Whether the legacy claim module's 2022 airdrops are already expired (its REST query is not
  served; confirm from a full node). Its messages are removed either way.
- Keeper code left unreferenced by the upgrade 1 removals (stakeibc's LSM liquid-stake entry
  points, the trade-route authz and ICA registration helpers, `EnableRedemptions`) is dead but
  harmless; deleting it is a follow-up cleanup, not part of the two upgrades.
