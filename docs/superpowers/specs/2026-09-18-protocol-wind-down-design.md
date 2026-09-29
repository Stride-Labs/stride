# Protocol Wind-Down: Migration to Osmosis

Status: approved design, pre-plan. Large tier; the implementation plan is
`docs/superpowers/plans/` (the two existing plans merge into one v35 plan, §12).

## §1. Goal

Wind Stride down and shut the chain off. Stop every flow that moves funds on behalf of users,
unbond every delegation, send the native tokens from the host chains straight to Osmosis, and
put them into transmuter pools where holders swap stTokens for the backing native tokens at a
fixed rate. stTokens and STRD still sitting in Stride accounts are sent to their owners on
Osmosis before the halt, and IBC vouchers go back to the chain they came from, so nothing a
user owns is left on a dead chain. The guiding constraints, in order: least risk of a bug that
loses funds, least new code, most reuse of code that already runs on mainnet, and a chain that
stops rather than one that idles for a year.

The shape is one upgrade, v35, followed by one ops window of roughly 40 days that ends with a
coordinated halt: the longest unbonding period plus a few days of ops on either side. The
upgrade lands on a live chain and does not need it quiet. The redemptions open at that moment
finish through the pipeline that already exists (undelegate at the day epoch, sweep on
completion, claim), in parallel with the admin drain of everything else, and the redemption
rate is frozen by deleting the code that moves it rather than by halting the zones (§6). No
on-chain rate is ever computed; the frozen rate becomes each pool's exchange ratio.

Dormant capital is acceptable. Backing that no stToken ever claims stays in the pools and comes
back to Stride Labs by exiting them later. What is not acceptable is a pool that cannot cover
a valid swap.

## §2. Scope

In scope:

- Every non-deprecated stakeibc host zone: celestia, cosmoshub-4, dydx-mainnet-1,
  haqq_11235-1, injective-1, juno-1, laozi-mainnet, osmosis-1, phoenix-1, sommelier-3, ssc-1.
- Staketia (stTIA redemptions and the 5-of-7 Celestia multisig), including its 7 open
  unbonding records. Its TIA joins the same stTIA pools as the stakeibc Celestia TIA.
- The stakeibc user redemption records open at the upgrade (99 on 2026-09-18) and any created
  before the redeem message is removed. All finish through the existing pipeline in the ops
  window and are claimed for everyone.
- Stakedym's 11 open unbonding records, flushed by its operator after the upgrade. Stakedym
  is never halted; its messages are removed and the module idles.
- Removing every message the protocol no longer needs, moving wasm control to gov, removing
  every IBC rate limit, and deleting the epoch-hook calls that compound or move stake.
- Four admin txs: per-validator undelegation, ICA-to-Osmosis transfer, the staketia
  claim-address transfer, and the batched token sweep off Stride.
- The Osmosis side as an ops procedure (§8): one transmuter pool per stToken route, its
  funding, and the coverage check that gates it. No Stride code is involved.

Out of scope, explicitly:

- The deprecated zones comdex-1, evmos_9001-2, stargaze-1 and umee-1. Apart from setting the
  `Deprecated` flag on comdex-1 (the other three already carry it), the upgrade does not touch
  them, and none of them are unbonded, drained or given a pool. Their host chains have stopped
  producing blocks, so nothing could be done through their ICAs anyway. Stakedym is likewise
  never unbonded, drained or given a pool. Holders of stCMDX, stEVMOS, stSTARS, stUMEE and
  stDYM have no redemption path after this work, which matches their status today but is now
  a deliberate decision.
- stTokens held by Stride accounts that no key controls (contracts, interchain accounts owned
  by other chains, module accounts including the community pool). They are not swept; their
  owners, where there is one, move them out themselves before the halt (§9). Their backing
  stays in the pools.
- Removing whole modules or their state. Only messages and hook calls are removed; stores
  stay.
- STRD-side modules beyond message removal (mint, strdburner). What happens to STRD, the
  validator set and the chain's history after the halt is a separate decision.

## §3. Facts the design rests on

Verified against mainnet between 2026-09-18 and 2026-09-28 (REST `stride-api.polkachu.com`,
Celestia REST `celestia.rpc.uquad.org`) and against the code at v34.

What the epoch hook drives (`x/stakeibc/keeper/hooks.go`, `BeforeEpochStart`):

- Day epoch: `InitiateAllHostZoneUnbondings` (submits every queued redemption as undelegate
  ICAs), `SubmitPendingUndelegations` (the v34 one-shot pipeline, a no-op with nothing
  pending), `CleanupEpochUnbondingRecords`, `CreateEpochUnbondingRecord`.
- Stride epoch: `ClaimAccruedStakingRewards` (withdraws rewards to the withdrawal ICA),
  `CreateDepositRecordsForEpoch`, `SetWithdrawalAddress` (an ICA `MsgSetWithdrawAddress` from
  the delegation ICA, every epoch, every zone), `UpdateRedemptionRates`,
  `TransferExistingDepositsToHostZones`, `StakeExistingDepositsOnHostZones`,
  `ReinvestRewards`, `RebalanceAllHostZones`, `SweepUnbondedTokensAllHostZones`,
  `TransferAllRewardTokens`.
- Mint epoch: `AuctionOffRewardCollectorBalance` (liquid stakes 15% of the reward collector's
  host fees for the POA validators, sends the rest to the auction module).
- `UpdateEpochTracker` runs first on every epoch; the day and stride trackers feed the ICA
  timeouts of undelegate batches and sweep ICAs.
- Every loop iterates `GetAllActiveHostZone` or checks `HostZone.Halted`, so the halt flag
  stops all of them at once: the ones that compound (rate update, reinvest, delegate,
  rebalance) and the ones that pay redemptions (unbond, sweep, claim). Staketia and stakedym
  gate their flows on their own `Halted` via `GetUnhaltedHostZone`.
- `ReinvestRewards` is the root of the fee machinery: it submits the withdrawal-balance ICQ;
  that callback splits the withdrawal ICA balance, sends the fee cut to the fee ICA by ICA
  bank send and delegates the rest; the reinvest ack queues the fee-balance ICQ, whose
  callback IBC-transfers the fee ICA to the reward collector on Stride. Nothing in the
  pipeline is a hot wallet.
- `TransferAllRewardTokens` iterates trade routes, not host zones. One route is live: dYdX
  pays rewards in USDC to the withdrawal ICA, an epochly ICQ moves it via Noble to the trade
  ICA on Osmosis, an off-chain trade controller with an authz grant swaps it to DYDX, and a
  second ICQ moves the DYDX back to the withdrawal ICA (`reward_converter.go`). The amounts
  are immaterial: 0.05 to 0.1 USDC per hour on 406k DYDX (sampled 2026-09-21).
- The delegator-shares ICQ callback, reached from `UpdateValidatorSharesExchRate` and
  `CalibrateDelegation` (both permissionless today), applies a detected slash to the
  validator's and zone's delegation and then calls `UpdateRedemptionRateForHostZone`. It
  computes the slash as tracked delegation minus on-chain shares × the stored exchange rate,
  so a delegation-only correction is not re-applied when the rate is later refreshed. The
  calibration callback refuses any change above `CalibrationThreshold` (5,000 base units).
- Admin messages that move funds via ICA: `RebalanceValidators` (redelegations) and
  `ClearBalance` (ICA transfer from the fee account). `ResumeHostZone` un-halts a zone.
- Autopilot has a `StakeibcActive` param. The ICA host on Stride allow-lists
  `MsgLiquidStake`, `MsgRedeemStake` and `MsgClaimUndelegatedTokens` alongside bank, staking,
  distribution, transfer and gov-vote messages.
- Wasm code upload is restricted to two addresses; all 44 codes were uploaded by one Stride
  address. 21 contracts are instantiated (all Hyperlane): 17 have no admin and 4 have the
  Stride deploy key `stride159smvptpq6evq0x6jmca6t8y7j8xmwj6kxapyh` as admin. Deployed
  contracts are executable by anyone.
- The rate limiter covers only stTokens (stATOM, stOSMO, stTIA, stJUNO, stEVMOS); no native
  denom is limited. Its keeper exposes removal of limits, blacklisted denoms and whitelisted
  address pairs.

Pipelines the design reuses:

- Since v34 the undelegate ICA callback accepts a submission with no epoch unbonding record
  ids: on success it only decrements the validator and host zone delegation balances and burns
  nothing; on an ack error or timeout it unflags the batch's validators and changes no
  balance; an ack arriving on a channel that has since been restored is ignored
  (`icacallbacks_undelegate.go`, with tests for all three). `BatchSubmitUndelegateICAMessages`
  takes per-validator `MsgUndelegate`s directly and flags `DelegationChangesInProgress` on
  each; `applySharesRoundingSafety` trims a full-drain amount so share truncation cannot make
  it exceed the on-chain delegation. `RestoreInterchainAccount` reopens an ICA channel that
  an ordered-channel timeout closed and resets every validator's in-progress counter. None of
  this depends on the zone being active.
- Claim today: unbonded tokens sit in the zone's redemption ICA on the host;
  `ClaimUndelegatedTokens` is permissionless, submits an ICA bank `MsgSend` from the
  redemption ICA to the record's host-chain `Receiver`, sets `ClaimIsPending`, and the callback
  deletes the record on ack (the timeout callback resets a stuck one). `UserRedemptionRecord`
  stores no Stride address; autopilot redemptions set `Creator` to a derived address, so claims
  can only ever be paid on the host.
- Every host zone has eight ICA types: delegation, fee, withdrawal, redemption, community-pool
  deposit and return, and the two trade-route converter accounts (`ICAAccountType`). The dYdX
  withdrawal ICA holds ~3.8 USDC beside its DYDX. The community-pool ICAs (the feature was
  disabled in v28) hold single-digit base units on every zone plus 50 USDC in dYdX's deposit
  ICA (checked 2026-09-22); they are written off along with the converter ICAs.
- `HostZone.Deprecated` is true on exactly evmos_9001-2, stargaze-1 and umee-1, which are
  also halted. comdex-1 is neither halted nor deprecated on chain; its host chain is dead.
- v34 has a chain-agnostic delegation delta helper (`app/upgrades/v34/delegation_deltas.go`):
  a per-validator table of on-chain minus tracked delegation, applied all-or-nothing to the
  validators and `TotalDelegations`, skipped with a log (never an upgrade error) if any constant
  no longer matches state, with a mainnet-export test. Injective and Celestia use it.
- An ICA-wrapped IBC `MsgTransfer` from a host account to a third chain exists in the
  trade-route code (`BuildHostToTradeTransferMsg` sends from the withdrawal ICA on the host to
  Osmosis via Noble). Autopilot and staketia already call `transferKeeper.Transfer` from
  keeper code with a sender that is not the tx signer, so a module can move an account's
  balance over IBC without a signature from it.
- Every admin tx checks its signer against the hard-coded admin set (`utils.Admins`: the F5
  key and the gov module) in ValidateBasic (`utils.ValidateAdminAddress`).
- Every IBC voucher on Stride carries a denom trace in the transfer module (`GetDenom` by
  hash): its base denom and the hops it took, outermost first. Sending a voucher back over
  its outermost channel unwinds exactly one hop: a single-hop voucher becomes the native
  token on its source chain (how Stride already returns host tokens to their host), and a
  multi-hop voucher becomes the shorter voucher on the chain it last came from. Either way
  the token goes back to the chain it arrived from and never grows a new hop.

Staketia:

- Its flow is operator-driven with automatic payout: redeem escrows stTIA → epoch
  `PrepareUndelegation` freezes the record → operator undelegates on Celestia and calls
  `MsgConfirmUndelegation` (burns stTIA) → after 21 days `MarkFinishedUnbondings` → operator
  IBC-transfers to the claim address and calls `MsgConfirmUnbondedTokenSweep` (checks the
  claim-address balance covers the record) → next hour epoch `DistributeClaims` pays each
  redeemer on Stride. Stakedym is the same module shape.
- The Celestia multisig (`celestia1d6ntc7s8gs86tpdyn422vsqc6uaz9cejnxz5p5`, 5-of-7) has
  granted the operator key (`celestia1ghhu67ttgmxrsyxljfl2tysyayswklvxzls400`) unexpiring
  authz for delegate, undelegate, withdraw rewards, cancel unbonding, and an IBC
  `TransferAuthorization` on `transfer/channel-4` with an effectively unlimited utia limit,
  exactly one allow-listed receiver (staketia's claim address
  `stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd`), and an empty `allowed_packet_data`, which
  ibc-go enforces as "memo must be empty". Bank `MsgSend` is not granted, and the 5-of-7
  signers are no longer reachable, so the operator's authz route is the only way to move the
  multisig balance, and it can only land on the claim address as a TIA voucher on Stride.
- The claim address (labelled S2 in `x/staketia/types/celestia.go`, from the January 2024
  launch) is itself a 5-of-7 multisig `BaseAccount` with a signer set that shares no key with
  the Celestia multisig; it has signed exactly one tx in its life. Whether five of its signers
  can be gathered is unknown; the design does not depend on it. Staketia's keeper already
  moves funds out of a multisig `BaseAccount` without a signature: its deposit address (S0)
  is IBC-transferred from keeper code with the deposit address as sender
  (`x/staketia/keeper/delegation.go`). Staketia's IBC middleware only acts on packets it
  recorded by sequence, so an inbound transfer to the claim address and an outbound one from
  it are both inert to it.

State on mainnet (2026-09-18 to 2026-09-22):

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
- Native vouchers stranded on Stride are dust: the eleven deposit addresses, the reward
  collector and the auction module together hold about $3 of in-scope native denoms. Nothing
  on Stride except the staketia claim address will hold a native balance worth moving.
- There is no direct transfer channel between Stride and noble-1. The only single-hop `uusdc`
  vouchers on Stride are Axelar's (channel-69 and channel-11, about 2 USDC in total on
  2026-09-28), so "USDC on Stride" is negligible; the largest non-host vouchers are stTokens
  that came back through Osmosis without unwinding.

Where the stTokens are (bank `denom_owners` and per-channel escrows on 2026-09-22; the full
tables and the relayer scope per chain are in `docs/wind-down/sttoken-locations.md`):

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
  escrow accounts; the zones' deposit addresses (in-flight redemptions, 30k stATOM, all of
  which finish in the ops window); the distribution module (15.5k stATOM, mostly unwithdrawn
  validator commission and STRD staking rewards; the community pool proper holds about $17k
  of stTokens across every denom); interchain accounts owned by other chains (2.3k stATOM in
  one); and a handful of 32-byte contract-style accounts (7k stTIA in one).
- Stride and Osmosis both derive addresses from secp256k1 keys on coin type 118, so the
  Osmosis address of a 20-byte Stride account is the same bytes with the `osmo` prefix, and
  the same key controls both. A multisig address is the hash of its member keys and threshold
  and derives the same way. 32-byte addresses (contracts, interchain accounts, module-derived
  accounts) have no controlling key and no counterpart. The same holds for every chain whose
  wallets use secp256k1 on coin type 118 (chain registry `slip44`): cosmoshub-4, celestia,
  osmosis-1, juno-1, sommelier-3, ssc-1, dydx-mainnet-1, noble-1. It does not hold where the
  wallet convention differs: phoenix-1 (coin type 330), laozi-mainnet (494), injective-1 and
  haqq_11235-1 (60, Ethereum-style keys and hashing). On those chains the address with the
  same bytes belongs to nobody the holder knows, so nothing may be sent there by derivation.

Transmuter (osmosis-labs/transmuter v3.2.0, code id 996 on osmosis-1; source checked
2026-09-22, behaviour tested on mainnet 2026-09-23 to 25, see `docs/wind-down/transmuter.md`
and the test log beside it):

- A transmuter is a `cosmwasmpool` that swaps its assets at a fixed ratio. v3 ("alloyed")
  gives each asset a `normalization_factor` (any `Uint128`); the rate between two assets is
  the ratio of their factors, and a larger factor makes a unit of that token worth less
  (`out = in × out_factor / in_factor`). So the stToken factor is `1e18` and the native
  factor is `RedemptionRate × 1e18`, which encodes the 18-decimal frozen rate exactly; the
  alloyed asset takes the native factor so one alloyed unit is one native base unit. Swap fee
  is hard-coded to zero; exact-in output rounds down, exact-out input rounds up.
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
  `mark_corrupted_assets`, after which the contract refuses any action that raises that
  asset's balance; when a corrupted asset's balance reaches zero the contract removes it from
  the pool, and `add_new_assets` at the same factor plus a join restores it with the rate
  exact (tested). Adminship transfers in two steps and cannot be renounced (a transfer only
  completes when the candidate claims it). Contract code migration is governance-only (the
  wasm admin is the cosmwasmpool module).

## §4. Operator addresses

Three addresses run the wind-down. Each has one name used everywhere in this spec, the plan
and the code, and each is a hard-coded constant in the upgrade binary (the protocol admin
already is). Nothing on Stride receives native tokens: every native balance leaves from a
host chain, so there is no Stride-side vault.

| Name | Chain | Type | Status | Constant | Address | Role |
|---|---|---|---|---|---|---|
| Protocol admin | Stride | key (F5) and the gov module | exists, `utils.Admins` | `utils.Admins` | `stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh` (F5), `stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl` (gov) | Signs `MsgUndelegateFromValidators`, `MsgTransferFromIca`, `MsgTransferStaketiaClaimBalance`, and the two admin-gated ICQ messages. |
| Sweep operator | Stride | new key | to create | `SweepOperatorAddress` | `stride1...` (paste here) | The only address that can sign `MsgSweepTokensOffStride`. Separate from the protocol admin so the sweep, the one tx that moves user balances, has its own key and its own blast radius. Holds STRD for fees only. |
| Osmosis vault | Osmosis | new multisig | to create | `OsmosisVaultAddress` | `osmo1...` (paste here) | Receives every ICA transfer, instantiates and funds the pools, holds the alloyed assets, and is each pool's admin and moderator. |

The sweep operator and the Osmosis vault are created and proven before the upgrade PR is cut
(a signed spend from each), and their addresses go into the binary as constants (§9). The
Osmosis vault's moderator role on the pools can be split to a second multisig later with
`assign_moderator`; that is an ops choice, not a design point.

## §5. The upgrade: what it removes and reconfigures

One SDK upgrade, v35. Everything in this section is either a deletion from the binary or a
one-time state change in the handler; the new code is confined to the four admin txs (§7).
Every handler step logs and skips on missing state rather than erroring, with one exception
noted below.

**Message handlers removed** (the `rpc` line in the proto `Msg` service, the msg-server
method, the CLI command, their tests). The message types and both of their registrations stay:
the interface-registry entry, so that every historical transaction containing them still
decodes (`strided q tx`, the tx REST endpoints, explorers; dropping it would break that for
most of the chain's history, and since the SDK registers Msg types from the service
descriptor, the explicit `RegisterImplementations` entries must be kept once the `rpc` is
gone), and the amino name, so that legacy-amino renderings of those transactions keep
working. Removing the `rpc` is what removes the handler: with nothing in the msg service
router, a submitted message fails with "can't route message" for every entry path that goes
through the router, which is a stronger block than a flag. Decoding needs the type, not the
handler, so old transactions are unaffected; a decode test proves it (§11). Keeper functions stay wherever something still calls them, so stakeibc's
`LiquidStake` and `RedeemStake` keeper paths survive for autopilot and the community pool,
while staketia's and stakedym's redeem keeper paths go with their handlers.

- stakeibc: `LiquidStake`, `LSMLiquidStake`, `RedeemStake`, `RegisterHostZone`,
  `CreateTradeRoute`, `UpdateTradeRoute`, `DeleteTradeRoute`, `SetCommunityPoolRebate`,
  `ToggleTradeController`, `RebalanceValidators`, `ClearBalance`, `ResumeHostZone`.
- staketia and stakedym: `RedeemStake` (their `LiquidStake` is already a disabled stub; remove
  the stub too), `ResumeHostZone`.
- icaoracle: `AddOracle`, `InstantiateOracle`. icqoracle: `RegisterTokenPriceQuery`,
  `RemoveTokenPriceQuery`. auction: `PlaceBid`, `CreateAuction`, `UpdateAuction`. airdrop:
  all seven. claim (legacy): all four.

`ClaimUndelegatedTokens` is kept: it pays the redemptions that finish after the upgrade, it
always was permissionless, and it is a no-op once the last record is claimed.
`RestoreInterchainAccount` is kept for the drain's timeout path (§7).

**Messages gated.** `UpdateValidatorSharesExchRate` and `CalibrateDelegation` gain
`utils.ValidateAdminAddress` in ValidateBasic: ops use them to refresh slashes before the
drain (§9), and nobody else can reach the slash path. The 5,000 base-unit
`CalibrationThreshold` check is removed from the calibration callback: it existed to bound
what a permissionless caller could move.

**Entry points that bypass the router.** Autopilot: the handler sets `StakeibcActive =
false`. ICA host: the handler removes `MsgLiquidStake` and `MsgRedeemStake` from the
allow-list; `MsgClaimUndelegatedTokens` stays, so ICA-originated claims keep working while
the open redemptions finish.

**Wasm.** The handler sets `code_upload_access` to the gov module address only, and for each
of the four contracts whose admin is the Stride deploy key, sets the admin to the gov module
address through the gov permission keeper (whose authorization policy allows the change
without the current admin's signature). The upload-access write is the one handler step that
fails the upgrade on error rather than logging: it can only fail on invalid params, and
leaving upload open to the two keys with nothing but a log line is the worse outcome. A
contract that no longer exists is logged and skipped. The list is re-checked right before the
proposal.

**Trade route.** The handler deletes the dYdX trade route. Its conversions are worth 1 to 3
USDC a day (§3), deleting it removes the off-chain trade controller from the picture, and the
USDC already in the dYdX withdrawal ICA is swept as surplus (§7). The converter ICA addresses
are noted in the plan; the stale authz grant on the Osmosis trade ICA is harmless.

**Oracles and rate limits.** The handler deactivates the three ICA oracles (existing toggle
logic) and removes every rate limit, blacklisted denom and whitelisted address pair from the
rate limiter. The token sweep (§7) sends most of each stToken's on-Stride supply out over one
channel in a few days, which no limit would allow, and there is no longer a mint path that a
limit would protect. The module and middleware stay in the stack with empty state.

**Comdex.** The handler sets `Deprecated = true` on comdex-1 so it carries the same flag as
the other three deprecated zones. `Halted` is not touched; the flag is documentation.

**Pending ICQs.** The handler throws out two kinds of pending interchain query. For
haqq_11235-1 it deletes every slash-path query (validator exchange rate, delegator shares,
calibration callbacks) and clears `SlashQueryInProgress` on every haqq validator, before the
delta table below is applied: a query submitted against the pre-delta state has no reason to
exist, and unlike v34's `DeleteStuckQueries` and `ResetStuckSlashQueries`, which pin query ids
and validator addresses, this is dynamic and cannot go stale between measurement and
execution (haqq had 8 such queries open and one flagged validator on 2026-09-21). For every
zone it deletes the pending withdrawal-balance queries, because their callback delegates and
the call that submits them is gone (§6).

**Haqq delegation reconciliation.** The handler applies a per-validator delta table to
haqq_11235-1 with the v34 helper, exactly as v34 did for Injective. The 2026-09-21
measurement (§9a) has 17 validators off: 14 over-recorded (undetected downtime slashes and
sub-token rounding, the largest 853.8 ISLM) and 3 under-recorded by sub-token dust. Both
signs are applied so every tracked delegation equals the chain's; the net is a decrease of
about 1,758 ISLM, so `TotalDelegations` drops. The rate update is deleted in the same
upgrade (§6), so this no longer flows into the stISLM redemption rate; the rate stays about
0.002% above the backing (1,758 ISLM against 101M stISLM, roughly $7), which the rewards
accrued until the drain cover many times over and the coverage check (§10) reports either
way. The table matched live state on three measurements over two days. The stored
`SharesToTokensRate` is deliberately left as is: the day-0 refresh updates it, and the slash
callback then finds tracked delegation equal to on-chain shares × the refreshed rate, so
nothing is applied twice. The table is generated from `measure_delegation_drift.py`,
re-measured right before the proposal, and covered by a mainnet-export test. This is the
Injective shape (real loss, no stranded liquid), not the v33 Osmosis shape (phantom stake
credited back as a deposit record).

## §6. What keeps running and what stops

The upgrade does not halt any zone. `Halted` would stop the flows the open redemptions need
along with the ones that must not run once the rate is frozen, so instead the binary deletes
the second group's call sites from `BeforeEpochStart` and leaves the first group in place.
Nothing is configurable and nothing can be toggled back; `Halted` stays false on every
in-scope zone, stakedym included. The keeper functions behind the deleted calls may stay
until the follow-up cleanup (§12).

Deleted from `BeforeEpochStart`:

| Call | Why it must stop |
|---|---|
| `UpdateRedemptionRates` (stride epoch) | The rate must not move once the drain starts: an admin undelegation lowers `TotalDelegations` with no record behind it and the formula would cut the rate. Deleting the call freezes `HostZone.RedemptionRate` at its last pre-upgrade value. |
| `ReinvestRewards` (stride epoch) | It delegates the withdrawal ICA's rewards back to validators, which would create fresh delegations after the drain and need a second unbonding period. It is also the root of the fee machinery (§3): removing this one call stops the withdrawal-balance ICQ, the fee split, the fee-balance ICQ and the reward-collector inflow. Rewards simply accumulate in the withdrawal ICA and leave with `MsgTransferFromIca WITHDRAWAL`; what the fee ICA holds at the upgrade leaves with `MsgTransferFromIca FEE`. |
| `StakeExistingDepositsOnHostZones` (stride epoch) | New delegations after the drain. Deposits that reach the delegation ICA stay there and leave with the ICA balance. |
| `RebalanceAllHostZones` (stride epoch) | Redelegations flag `DelegationChangesInProgress` and would collide with the drain. |
| `TransferAllRewardTokens` (stride epoch) | A no-op with the trade route deleted; removed for clarity. |
| `SetWithdrawalAddress` (stride epoch) | Redundant (the withdraw address persists on the host, set years ago) and traffic on the delegation ICA channel the drain uses. A checklist line replaces it (§9). |
| `CreateDepositRecordsForEpoch` (stride epoch) | Nothing creates deposit amounts once liquid staking is gone. |
| `CreateEpochUnbondingRecord` (day epoch) | Only `RedeemStake` appended to the current record; new ones would stay empty. |
| `AuctionOffRewardCollectorBalance` (mint epoch) | It liquid stakes the validators' fee share, minting stTokens against a frozen rate from tokens that never get delegated. Validators are paid in STRD only from day 0 (§10a). |

Kept in `BeforeEpochStart`:

| Call | Why it stays |
|---|---|
| `UpdateEpochTracker` (every epoch) | The day-epoch tracker feeds the ICA timeout of every undelegate batch, including the drain's; the stride tracker feeds the sweep ICA. |
| `InitiateAllHostZoneUnbondings`, `SubmitPendingUndelegations`, `CleanupEpochUnbondingRecords` (day epoch) | Submit the redemptions open at the upgrade, retry a failed batch, and delete records once every zone's unbonding is claimed. |
| `SweepUnbondedTokensAllHostZones` (stride epoch) | Moves each completed record's amount from the delegation ICA to the redemption ICA, where the claim pays it. |
| `ClaimAccruedStakingRewards` (stride epoch) | Harmless; rewards withdrawn to the withdrawal ICA are swept to Osmosis. |
| `TransferExistingDepositsToHostZones` (stride epoch, deposit interval) | Carries any native voucher still in a deposit address to the delegation ICA, so it leaves with the ICA balance instead of being stranded. |

Also deleted: the `UpdateRedemptionRateForHostZone` call at the end of the delegator-shares
slash callback. The callback keeps correcting the validator's and zone's delegation on a
slash; it no longer rewrites the rate. A slash detected after the upgrade lowers the backing
but not the rate, and the coverage check (§10) is where it shows up.

What "frozen" means: the rate stops at the last stride-epoch update before the upgrade
height. Between the upgrade and the drain, delegations keep earning for a few days; those
rewards are withdrawn by the undelegation and end up as pool surplus. Each pool's factors are
read from `HostZone.RedemptionRate` at pool creation (§8).

Sequencing with the open redemptions: the first post-upgrade day epoch submits the queued
records from each validator's stored delegation, exactly as today. The drain runs only after
that epoch has submitted them and their acks have landed (`DelegationChangesInProgress` back
to zero), so the pipeline computes from delegations the admin has not touched and the admin
drains what is left with an empty list and no reservation arithmetic. The undelegate tx
refuses any validator with a change in progress, so a mistimed drain fails rather than races.
The redemption ICA is drained only after a zone's claims are done, since it holds the tokens
of claimable records.

## §7. The four admin txs

All four are gated in ValidateBasic: the undelegate, ICA transfer and claim-address txs on the
protocol admin (`utils.ValidateAdminAddress`), the sweep tx on the sweep operator (§4). None
reads or writes host zone accounting; amounts and address lists are ops inputs, and every
destination is a constant reviewed in the upgrade PR.

**`MsgUndelegateFromValidators { creator, chain_id, validators: [{address, offset}] }`.** An
empty `validators` list means every validator on the zone with a non-zero stored delegation;
otherwise only the listed ones. Per validator the amount is the stored delegation minus
`offset` (default zero), passed through `applySharesRoundingSafety`, and rejected if it is not
positive. No blanket offset is needed for the real run: an unslashed validator has an
exchange rate of exactly one, so a full-drain amount converts to shares with no truncation,
and the helper already buffers full drains of validators whose rate is below one. Both rely on
the stored rate matching the chain, which the day-0 refresh guarantees; `offset` is the lever
for a validator that has drifted since. A listed validator with `DelegationChangesInProgress`
set is rejected, and a deprecated zone is rejected. The messages go through
`BatchSubmitUndelegateICAMessages` with no epoch unbonding record ids, in the zone's usual
batch size, and the tx registers the batches as in flight so the callback's record-less
accounting stays clean.

Submitting by validator rather than cascading a zone amount by weight is what lets ops test on
a single validator first, skip or shave a validator that is failing or drifted by dust, and
keep every submitted amount equal to what the host will accept. Retry is resubmission by ops;
there is no queue and no epoch hook. ICA txs are atomic, so one over-recorded validator still
fails its whole batch; the slash refresh (§9) is what prevents that. The failure path is the
existing record-less path of the undelegate callback (§3): an ack error unflags the batch's
validators and changes no balance, and other batches of the same drain succeed
independently; a timeout does the same and, because ICA channels are ordered, closes the
channel, which `RestoreInterchainAccount` reopens while resetting the counters, the late ack
then being ignored; a success decrements the delegations by the split amounts and burns
nothing.

**`MsgTransferFromIca { creator, chain_id, ica_type, amount (Coin) }`.** `ica_type ∈
{DELEGATION, WITHDRAWAL, FEE, REDEMPTION}`, the four ICAs that hold anything (§3). The
destination is not a tx input: the receiver is the Osmosis vault (§4), one hard-coded
constant, and the host-side transfer channel to Osmosis comes from a hard-coded map
`chain_id → channel_id` covering the ten non-Osmosis in-scope zones; a `chain_id` absent from
the map is rejected. Both are reviewed and tested in the upgrade PR (§11) and verified
against the hosts before the proposal (§9), so a fat-fingered channel or receiver on the day
is not possible; the only per-tx inputs are the zone, the ICA and the amount. `amount`
carries its denom as it exists on the host, so foreign balances such as the USDC in the dYdX
withdrawal ICA can be sent too. The tx submits one ICA containing an ICS-20 `MsgTransfer` of
`amount` from that ICA over the mapped channel to the Osmosis vault, built like
`BuildHostToTradeTransferMsg` without the forwarding memo, with a one-day timeout. No callback
state: a failed or timed-out transfer refunds to the ICA on the host and ops resubmit.

The tokens go from the host straight to Osmosis, never through Stride, because that is the
only route that lands them as the canonical denom on Osmosis: a voucher that reaches Stride
first and is forwarded from there arrives as a two-hop denom that no pool would use. For the
osmosis-1 zone the ICA is already on Osmosis and the tx is an ICA bank `MsgSend` to the
Osmosis vault instead (osmosis-1 is in the map with an empty channel, which selects this
form).

**`MsgTransferStaketiaClaimBalance { creator }`**, with no other field. It moves the staketia
claim address's whole TIA voucher balance to the stakeibc celestia zone's delegation ICA on
Celestia, where it unwinds to native TIA and is then sent on to Osmosis by `MsgTransferFromIca
DELEGATION` with the rest of the zone's balance. Everything is a constant or read from state:
sender is staketia's `ClaimAddress`, denom is `CelestiaNativeTokenIBCDenom`, amount is the
full balance (reject if zero), channel is the celestia host zone's `TransferChannelId`,
receiver is its `DelegationIcaAddress`, one-day timeout, no memo. It calls
`transferKeeper.Transfer` with the claim address as sender, the pattern staketia's keeper
already uses for its deposit address (§3), so no multisig coordination is on the critical
path. A timeout refunds the claim address and ops resubmit. The same trick, sending a
Stride-side voucher to the zone's delegation ICA so it leaves with the ICA balance, is why
nothing else needs a Stride-side account; the claim address is the only Stride-side balance
worth it.

**`MsgSweepTokensOffStride { creator, denom, addresses: [string] }`**, the batched token
sweep. `denom` is any bank denom; `addresses` is non-empty and at most 100 entries (the batch
bound that keeps a tx inside the block gas limit; the plan measures the real cost and can
raise it). The destination is decided once per tx from the denom, before any address is
looked at, and is a (channel, bech32 prefix) pair:

- A Stride-native denom (anything that is not an `ibc/` voucher: every stToken, `ustrd`)
  goes to Osmosis over channel-5 with the `osmo` prefix. It lands as its canonical Osmosis
  denom, which is what the pools use.
- An `ibc/` voucher goes back over the channel it arrived on, the outermost hop of its denom
  trace, so it unwinds one hop: a host token lands native on its host, and a voucher that came
  through Osmosis lands on Osmosis as whatever it was there. That channel must be in
  `SweepUnwindChannels`, a hard-coded map from Stride transfer channel to the counterparty's
  bech32 prefix, which is exactly the channels to chains whose wallets derive the same address
  bytes as Stride (§3): cosmoshub-4 (channel-0, `cosmos`), celestia (channel-162, `celestia`),
  osmosis-1 (channel-5, `osmo`), juno-1 (channel-24, `juno`), sommelier-3 (channel-150,
  `somm`), ssc-1 (channel-213, `saga`) and dydx-mainnet-1 (channel-160, `dydx`). A voucher
  whose outermost channel is anything else rejects the whole tx: a derived address on
  phoenix-1, laozi-mainnet, injective-1 or haqq_11235-1 is not the holder's, and the Axelar
  channels carry about 2 USDC. Holders of those vouchers move them themselves before the halt.

Then, for every listed address, in order:

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
  the destination prefix, over the destination channel, with a one-day timeout and no memo. A
  timeout or a rejected receive refunds the holder on Stride through the normal ICS-20 path
  and ops resubmit that address.

Every destination is therefore the canonical form on Osmosis, the native form on the source
chain, or one hop closer to it; the sweep never adds a hop to any denom and never needs a
forwarding memo. One tx sweeps up to 100 holders of one denom; the whole sweep at the chosen
floor is a few thousand packets over a few days (§3). There is no floor on chain: only the
sweep operator can sign the tx, so nobody can spam it, and the floor is an ops choice made
from prices on the day and stated in the announcement. Every account that clears the floor is
swept; holders below it, and holders on other chains, move themselves (§8).

## §8. Osmosis side: pools, funding, foreign-route denoms, shutdown

Nothing in this section is Stride code. It is the procedure ops follow in the ops window as
the native tokens arrive.

**One pool per stToken route**, instantiated from code id 996 (v3.2.0) by the Osmosis vault
(§4): for each in-scope stToken, a canonical pool holding the canonical denom on Osmosis (the
one minted by transfers over Stride's channel-5) plus the native token, and one more two-asset
pool for every foreign route the locations table marks in scope
(`docs/wind-down/sttoken-locations.md`; about 28 pools in all, 7 of them for stATOM). A pool
never holds more than one stToken denom, so the only thing a holder can do with a flavour is
turn it into the native token at the rate, and a compromised source chain can reach nothing
but its own pool. Every pool uses normalization factors `1e18` for the stToken and
`HostZone.RedemptionRate × 1e18` for the native token and the alloyed asset (§3), the rate
read from Stride at instantiation and frozen since the upgrade (§6). Admin and moderator are
the Osmosis vault; the moderator's freeze is the incident lever, and since adminship cannot be
renounced the vault keeps it for the life of the pools. The alloyed asset each pool mints is
the LP receipt and stays in the vault; it is the withdrawal key to the pool's backing and is
custodied like the backing itself.

**No limiters.** Isolation comes from the pool layout: a route pool is funded with exactly
that route's escrow share, which bounds what a counterfeit route could ever drain to the same
amount a cap would have, without a permanent limiter to configure, check or widen. The levers
that remain are the moderator's freeze per pool and `mark_corrupted_assets`.

**One-way pools.** Immediately after each funding join the moderator marks the native token
corrupted (`mark_corrupted_assets [native]`). The contract then refuses any action that raises
the native balance, so the only possible movement is stToken in, native out: nobody can buy
stTokens from a pool, and de-hopping a foreign route through two pools (route → native →
canonical) is impossible. Redemptions by router swap and by join-then-exit are unaffected
(tested 2026-09-25). A top-up therefore means unmark, join, re-mark; and when a pool's native
balance reaches zero the contract removes the native asset, after which `add_new_assets` at
the same factor plus a join restores it with the rate exact (tested). The canonical denom's
counterfeit risk is a forged Stride header, handled by the key destruction after the halt
(§9).

**Funding** is one `join_pool` per pool with native tokens only, once every source for that
native denom has arrived: the delegation ICA transfer and the withdrawal, fee and redemption
ICA sweeps (for stTIA the delegation ICA transfer includes the former multisig balance,
routed through the claim address). Each route pool receives exactly `escrow_route ×
RedemptionRate` native tokens, where `escrow_route` is the balance of Stride's escrow account
for that route's channel in the halt export; the canonical pool receives everything else
(§10). A stToken's pools can be created and funded as soon as its native denom is complete;
zones finish unbonding on different days and nothing couples them.

**Foreign-route denoms.** A stToken that left Stride to chain X and is sent from X to Osmosis
arrives as a two-hop denom (`transfer/<osmosis-X channel>/transfer/<X-stride channel>/st...`),
distinct from the canonical one. Rather than route it back through Stride, which is
impossible after the halt, each such denom gets its own pool with the same factors, so it
swaps to the native token at the same rate. The per-channel escrow balances on Stride (§3)
list exactly which chains hold which stToken and how much, so the set of pools is known
before the halt. The contract requires every pool asset to have supply on Osmosis at
instantiation, so ops seed each route denom with a small transfer from that chain first (or
create the pool after the first user's transfer lands). A third-hop denom gets a pool on
request the same way, funded by moving native tokens out of the canonical pool. Holders in
DeFi protocols on other chains withdraw there and transfer to Osmosis directly.
`add_new_assets` is not used in normal operation.

**Shutdown.** Once every pool is funded, the sweep is complete and the halt checklist (§9)
passes, validators set a `halt-height` and the chain stops. Stride's IBC clients on other
chains expire after their trusting periods; stTokens on those chains stay ordinary vouchers
and keep working in the pools, and any packet toward Stride that was never relayed times out
and refunds on its source chain. The pools outlive the chain. Backing that is never claimed is
reclaimed later by exiting the pools with the alloyed asset; that timing is a policy decision
outside this design.

## §9. The ops window and its checklists

Before the upgrade proposal:

1. Create the sweep operator key and the Osmosis vault (§4) before the upgrade PR is cut,
   because the binary hard-codes both. Gather the host-side channel id to Osmosis for each of
   the ten non-Osmosis zones, each verified by querying the host's channel and confirming its
   client's counterparty chain id is osmosis-1 and the channel is open, and the list of
   foreign-route denoms per stToken from the escrow balances.
2. Decide which chains get a relayer from us and which are served for free, from the "Relayer
   scope per chain" table in `docs/wind-down/sttoken-locations.md` (regenerated by
   `scripts/wind-down/build_relayer_scope.py`); the relayer map beside it shows the same per
   phase.
3. Announce the timeline, the sweep floor and date, and that holders on other chains transfer
   to Osmosis directly.

Checklist to propose the upgrade (there is no "nothing in flight" condition):

- The haqq delegation delta table, the drift measurement (§9a) and the mainnet-export tests
  match the chain at one recent height. A haqq redemption unbonding at the upgrade would
  change the drift; measure right before the proposal and expect the delta helper to skip
  (never error) if it no longer matches, with the `offset` on the drain tx as the fallback.
- On every in-scope host, the delegation ICA's withdraw address (distribution module query) is
  the zone's withdrawal ICA; the epoch call that re-set it every epoch is deleted (§6).
- The two new address constants, the channel map and `SweepUnwindChannels` in the binary
  re-verified: the maps against the hosts, each address by a test transfer of a few tokens to
  it from any wallet followed by a signed spend from it, so every constant is proven to be an
  address we control before anything is sent there.
- The staketia and stakedym operators ready to act on day 0.

The ops window (after the upgrade, ~40 days). The redemptions open at the upgrade are queued,
unbonding, or unbonded and waiting for a sweep or a claim; the pipeline finishes all three
(§6) while the rest proceeds:

1. Day 0: refresh every validator's exchange rate with `UpdateValidatorSharesExchRate` (CLI
   `update-delegation`) on every in-scope zone and wait for the callbacks. Where the rate
   moved, the callback applies the slash to the recorded delegation (the redemption rate is no
   longer rewritten, §6). Use `CalibrateDelegation` for a validator whose rate is unchanged
   but whose recorded balance is off. Then rerun the drift measurement (§9a) and require zero
   over-recorded validators on the zone. Same day: the staketia operator undelegates the
   entire multisig delegation on Celestia via authz (the queued records' amounts and
   everything else, in one go); the stakedym operator sweeps and confirms the unbonded records
   and undelegates the queued ones.
2. Day 0 to 4: the next day epoch submits the queued redemptions on every zone and flags their
   validators. Wait for those acks. A record whose submission fails (a slash between the
   refresh and the epoch) goes to `RETRY_QUEUE` and is retried at the next day epoch; that
   zone's drain waits for it.
3. Then `MsgUndelegateFromValidators` per zone: first for a single small validator as a live
   test of the tx and the callback, then with an empty list for the rest. An ICA tx is atomic,
   so one over-recorded validator fails its whole batch; after the refresh there are none. If
   a slash lands between the refresh and the submission, that batch fails, ops rerun the
   refresh (or set that validator's `offset`) and resubmit for the affected validators.
   Delegation ICA channels and relayers stay healthy until every batch acks; a dead channel
   is restored with the existing flow and the affected validators are resubmitted. From here
   on no delegation exists that any record needs.
4. Day 0+: `MsgTransferFromIca` for the withdrawal and fee ICA balances, including foreign
   denoms such as the dYdX USDC. This is the live test of the transfer tx on small real
   amounts, and the first arrival on Osmosis confirms the mapped channel end to end and the
   denom each zone lands as. Not yet the redemption ICA: it holds the tokens of claimable
   records until they are claimed.
5. Day 14 to 34, as each zone's unbondings complete: the pipeline sweeps each record's amount
   to the redemption ICA at the next stride epoch and the record goes `CLAIMABLE`; ops run
   `ClaimUndelegatedTokens` for every record (permissionless, as today; a record whose host
   receiver rejects the bank send is handled by hand). Once a zone has no record outside
   `CLAIMED` and no claim ICA in flight: `MsgTransferFromIca DELEGATION` for the full
   remaining balance, `WITHDRAWAL` again for the auto-withdrawn rewards, and `REDEMPTION` for
   whatever dust the claims left. Then create and fund that stToken's pools (§8), the
   canonical one and one per in-scope route, once the coverage check passes (§10).
6. Staketia, day 21: the operator IBCs the whole unbonded balance via authz to the claim
   address (no memo; the grant forbids one) and `MsgConfirmUnbondedTokenSweep` for each open
   record; the hour-epoch hook pays the redeemers from the claim address. Then
   `MsgTransferStaketiaClaimBalance` moves the remainder to the celestia delegation ICA, which
   leaves with that zone's balance in step 5. No key of the claim address signs anything.
   Stakedym: after 21 days the operator sweeps and confirms the last records.
7. Last days: `MsgSweepTokensOffStride` in batches of up to 100, per denom, for every holder
   at or above the floor, built from a fresh export: the eleven stTokens and `ustrd` (to
   Osmosis), then every voucher whose outermost channel is whitelisted and that is worth
   sweeping (ATOM to the Hub, TIA to Celestia, and so on, each back one hop). Vouchers over
   the four host channels outside the whitelist are announced as self-service. Resubmit any
   address whose transfer timed out (its balance is back on Stride). Relayers on channel-5 and
   on every whitelisted channel stay up until the last packet acks.
8. Transfer-channel relayers stay up until the halt. ICA channels can be left to close once
   every balance is sent.
9. After the halt: every validator rotates or destroys its consensus key, and Stride Labs
   confirms it in writing from each. Osmosis's light client of Stride (`07-tendermint-2119`,
   12-day trusting period) accepts any header signed by two thirds of the last trusted
   validator set until it expires; a forged header could mint canonical stToken vouchers on
   Osmosis, which the pools would honour. Keys gone means the window is closed on day 0 rather
   than day 12. No relayer is asked to update that client after the halt.

Checklist to halt the chain:

- Zero stakeibc user redemption records; zero `HostZoneUnbonding` records outside `CLAIMED`
  with a non-zero amount; no claim ICA in flight. Staketia and stakedym: zero unbonding
  records outside `CLAIMED`, zero redemption records, claim addresses drained.
- No undelegate batch in flight (no validator with `DelegationChangesInProgress`) and
  `TotalDelegations` at dust on every zone except celestia, where it still carries the
  multisig portion (the per-validator unbond only touches ICA validators).
- All four ICA balances at dust on every zone. Celestia multisig delegation zero, its
  unbondings complete, its balance on Osmosis.
- Every pool, canonical and per route, created, funded and passing the coverage check (§10)
  against a fresh export.
- The sweep complete: no sweepable account at or above the floor holds any in-scope stToken,
  `ustrd`, or a whitelisted voucher on the sweep list, and no sweep packet outstanding on
  channel-5 or a whitelisted channel.
- Validators and STRD delegators have withdrawn their rewards; interchain-account holders
  have been notified and given time to move out.

### §9a. Drift measurement

Measured 2026-09-21 for every in-scope zone except cosmoshub-4 and injective-1 (handled by
v34): recorded per-validator delegations versus the delegation ICA's on-chain balances.
dydx-mainnet-1, ssc-1 and sommelier-3 are exact. celestia is under-recorded by ~15,440 TIA
(the v34 phantom stake; harmless for unbonding). osmosis-1 has one validator over by 3,026
uosmo with an unchanged rate. juno-1 (1 validator, 1.94 JUNO), laozi-mainnet (2, 4.53 BAND),
phoenix-1 (10, max 1.20 LUNA) and haqq_11235-1 (14 over plus 3 under by dust, max 853.8
ISLM, 0.01% of that validator) are over-recorded, and in every case Stride's stored exchange
rate is above the chain's, i.e. undetected downtime slashes. Recomputing each validator as
on-chain shares × the chain's current rate reproduces the on-chain balance exactly for every
validator on every zone, so the refresh leaves no rounding gap and no per-validator buffer is
needed. The one wrinkle is haqq_11235-1's 18-decimal denom: two validators (SureStake,
Islamic Staking) are over by 203,557 and 3,216,141 aISLM with an unchanged rate, which is
dust in ISLM but above the calibration cap the upgrade removes, and even one base unit of
over-recording fails the host's share check. Haqq is therefore trued up by the delta table at
the upgrade (§5), which covers the dust cases too, and the `offset` on
`MsgUndelegateFromValidators` is the fallback for anything that drifts afterwards. The
measurement script is `scripts/wind-down/measure_delegation_drift.py` and is rerun as the gate
before the drain; haqq_11235-1 is expected to measure clean by then.

## §10. Accounting

There is no rate calculation anywhere in the design. The rate for a zone is the
`HostZone.RedemptionRate` that was on chain at the upgrade height; it is read from Stride
when the pool is instantiated and becomes the pool's normalization factors, which cannot be
changed relative to each other afterwards (§3). Nothing on Stride recomputes it: the epoch
update is deleted (§6) and the slash callback's rate rewrite with it, so a slash found by the
day-0 refresh lowers the recorded delegation, not the rate, and the coverage check reports the
difference.

Coverage check, per stToken, run from a fresh Stride export before the pools are funded and
again before the halt: native tokens held on Osmosis for that denom ≥ Stride bank supply of
the stToken × `HostZone.RedemptionRate`, and, per pool, each route pool holds exactly its
channel's escrow balance × the rate while the canonical pool holds the remainder. Bank supply
is the right reference: stTokens that left Stride over IBC are escrowed here, not burned, so
they are counted, and the per-channel escrows are exactly the route pools' funding; stTokens
burned by flushed redemptions are gone from both sides. The check is a script over two
queries and has no on-chain counterpart, which is the one place this design is weaker than an
on-chain assertion; the mitigation is that funding is a deliberate `join_pool` for the
computed amount, so an under-funded pool can only come from a wrong number in a script that
is run twice and published.

Why a frozen rate is covered. Confirmed against cosmos-sdk v0.54.3: `Unbond` calls the
distribution hook `BeforeDelegationSharesModified`, which withdraws the accrued rewards, and
then removes the shares; rewards are computed from delegation shares only, so an unbonding
entry earns nothing during its 21 to 30 days. Between the upgrade and the pool being funded,
the backing therefore moves as follows:

- Up: staking rewards accrue only from the upgrade until the drain executes (up to four days
  for the day epoch plus the drain itself, well under 0.1%), are auto-withdrawn to the
  withdrawal ICA by the undelegation, and are sent to Osmosis. Rewards withdrawn but not yet
  reinvested at the upgrade are also sent.
- Down: any slash during the ops window, and the haqq delta that no longer reaches the rate
  (about $7, §5).

Coverage then follows by construction. The refresh (§9, step 1) makes the recorded
delegations reflect every slash that has happened, the full recorded amount unbonds, and
Osmosis receives it plus the rewards swept on undelegation. Supply × frozen rate is the
delegated amount, so the rewards are the buffer. Only a slash between the refresh and the
undelegation, on a zone whose retried batch never lands, can leave a zone under, and the
coverage check catches that before funding; the shortfall is then covered from the surplus of
the other denoms or by Stride Labs, or the last swappers absorb it, which §1 accepts.

The token sweep does not change any of this: it moves stTokens from Stride accounts into
escrow, so bank supply is unchanged and the swept tokens become canonical-denom holders on
Osmosis. Sweeping `ustrd` and unwinding vouchers to their source chains has no bearing on the
pools at all. Unsweepable holdings (§2) stay in supply, are covered by the funding, and their
backing is what the pools hold at the end.

### §10a. Validator compensation

Today POA validators receive two streams: STRD from mint provisions (the staking share of
~58 STRD per hour epoch, about 16%) plus tx fees, and stTokens from the reward collector (15%
of Stride's commission on host staking rewards, liquid staked at the mint epoch).

The stToken stream ends at the upgrade: the reward-collector liquid stake and the fee split
behind it are deleted (§6), so validators are paid in STRD only from day 0. The rewards that
accrue during the ops window go to Osmosis with the rest and become pool surplus rather than
being split. The STRD stream continues unchanged until the halt. If validator pay needs to
rise to compensate, that is a mint distribution-proportion parameter change by governance and
needs no code. Validators withdraw their accumulated commission and rewards (including
stTokens, which they then move to Osmosis themselves) before the halt (§9).

## §11. Testing

- Handler (§5): tests against a mainnet export (`app/upgrades/v35/testdata/`, v34-style) for
  the trade route deletion, the comdex-1 `Deprecated` flag, the Haqq delta table (applied, and
  skipped on a stale constant), the two ICQ purges (the haqq purge deletes only that chain's
  slash-path queries and clears the validator flags; the withdrawal-balance purge leaves every
  other query), the autopilot param, the ICA host allow-list, wasm params and contract admins,
  oracle deactivation, rate-limit removal, the two ValidateBasic gates and the lifted
  calibration cap; a compile-time guarantee that the removed messages no longer exist and a
  decode test proving historical txs still parse; existing keeper tests for the flows that
  keep running stay green.
- Freeze by code (§6): a test that each deleted call is absent from the hook and each kept
  call still runs on a non-halted zone; a keeper test that `RedemptionRate` is unchanged
  across a stride epoch with deposit records and rewards present; a slash-callback test that
  the delegation is corrected and the rate is not.
- Admin txs (§7): unit tests for gating on all four. Undelegate: per-validator message
  construction with and without offsets, empty versus explicit validator lists, rounding
  safety on a full drain, rejection of a validator with a change in progress and of a
  deprecated zone, acceptance on a non-halted zone, no accounting mutation, the in-flight
  registration. Transfer: every ICA type, a foreign denom, the osmosis-1 bank-send form, a
  chain id absent from the map, that the map has an entry for every in-scope zone and none
  for a deprecated one, that the Osmosis vault constant parses as an `osmo` bech32 address
  and the sweep operator constant as a `stride` one, and the built `MsgTransfer` fields
  (mapped channel, Osmosis vault, timeout, empty memo). Sweep, the highest-review item:
  table-driven tests for a base account, each vesting type, an escrow address, a module
  account, an interchain account, a 32-byte address, an unknown account, a zero balance
  (skipped, others in the batch still sent), a batch over the bound, an invalid denom string,
  an stToken and `ustrd` landing on channel-5 with the `osmo` prefix, a single-hop voucher on
  a whitelisted channel landing on that channel with that chain's prefix, a two-hop voucher
  unwinding one hop over its whitelisted outer channel, a voucher whose outer channel is not
  whitelisted rejecting the batch, the derived address bytes, the full balance and only that
  denom being sent, and the ICS-20 refund on timeout returning the balance to the holder.
  Claim address: the full balance and only the TIA denom being sent, the built `MsgTransfer`
  fields (celestia channel, delegation ICA receiver, timeout, empty memo), a zero balance
  rejected, and the refund on timeout landing back on the claim address.
- Localstride run: upgrade with a redemption submitted beforehand, let the day epoch unbond
  it, drain the rest, let the sweep and claim complete, then one ICA transfer to a second
  local chain and one sweep batch whose packets are relayed and land at the derived
  addresses.
- Ops scripts: the coverage check and the batch builder are tested against a mainnet export
  and their output for the export is checked in beside the plan.

## §12. Open items for the plan

- **Band's light client of Stride is expired** (laozi-mainnet `07-tendermint-169` on the ICA
  connection `connection-146`, last header 2026-08-05; the delegation ICA restore is stuck in
  `STATE_INIT` on channel-768). No ICA tx, and no Stride→Band transfer, can be delivered
  until Band governance recovers it with `MsgRecoverClient` and a fresh substitute client, so
  the Band zone (~$110k stBAND) cannot be unbonded without that proposal. Raise it with the
  Band team now; it has to pass before the upgrade. Every other in-scope zone's host-side
  client of Stride is active (checked 2026-09-23).
- **Injective-hop stATOM is rejected by Osmosis's IBC rate limiter** (verified 2026-09-23 with
  two transfers, see the test log): the contract's returning-token check compares channel ids
  without a trailing slash, so `transfer/channel-89/stuatom` arriving from Injective's
  channel-8 is mis-classified and every packet fails with `rate limit exceeded`. Until
  Osmosis migrates the contract, nothing Stride-issued on Injective can reach Osmosis
  directly: ~31k stATOM (~$108k) and the stINJ held there (~$53k). Checked against every other
  in-scope pair (Hub, Celestia, dYdX, Haqq, Juno, Band, Terra, Saga, Secret, Penumbra, Agoric,
  Neutron): only Injective's channel ids collide. Options: file the bug with Osmosis now and
  ask for a fix before the upgrade; tell Injective holders to redeem via Stride before the
  upgrade; or, after the halt, route Injective → Hub → Osmosis and add the resulting three-hop
  denom to the pool.
- Stranded stToken holders behind expired clients, to decide on: Penumbra (Osmosis's and
  Stride's clients both expired; ~6.2k stATOM and stOSMO, ~$29k) can be reopened by an Osmosis
  `MsgRecoverClient` or a new channel; Kujira (~5.8k stATOM, ~$20k) has no reachable RPC and
  looks stopped, so it is ignored (decided 2026-09-23). Agoric is fine for holders
  (Agoric→Osmosis is active) though its Stride hop is expired. The five deprecated zones and
  their stTokens (~$16.8k in total) are not touched by the migration: Evmos, Stargaze, Umee
  and Comdex have stopped producing blocks (unrecoverable, like Kujira), and Dymension is
  alive but left alone by choice. `docs/wind-down/sttoken-locations.md` separates ignored
  balances into small, unrecoverable and deprecated.
- Relayers: until the upgrade users redeem through Stride, which needs both clients alive on
  every stToken chain ↔ Stride pair; Neutron's and Carbon's are active but nobody is updating
  them (35 h and 297 h old on 2026-09-23/24), so ops relay those pairs. Axelar's Stride hop is
  expired on both sides (its Osmosis hop is fine), so Axelar holders wait for the pools. After
  the upgrade the Stride leg matters only for hosts (our ICA relayers) and the sweep channel;
  every chain's Osmosis leg is decided by the relayer scope table (a recent packet on the
  channel means someone relays it for free; otherwise we run it when the chain holds at least
  the minimum, currently $1k, else holders relay their own hop). The live map is at
  https://claude.ai/artifact/986V5LAXxFgzjq7jXPpE8r.
- Exact proto shapes and enum names for the four admin txs; the constants: the two new
  addresses in §4 once created, the channel-5 constant and the `SweepUnwindChannels`
  whitelist for the sweep, and the `chain_id → host-side channel to Osmosis` map; the batch
  bound after measuring gas.
- The plans: `2026-09-21-wind-down-upgrade-1.md` and `2026-09-24-wind-down-upgrade-2.md` were
  written for the two-upgrade sequencing and merge into one v35 plan. The second's package
  becomes `app/upgrades/v35`, its halt steps and the `ClaimUndelegatedTokens` removal are
  dropped, its undelegate tx no longer requires `Halted`, and one new task does the §6
  deletions, the ICQ purges and their tests. The merged plan is delivered as six stacked PRs,
  each reviewable on its own, in this order:
  1. Remove tx handlers: pure deletions across stakeibc, staketia, stakedym, icaoracle,
     icqoracle, auction, airdrop and claim, plus rebalance, clear-balance and resume; types
     and registrations stay; the "no handler" guard test and the historical-tx decode test.
     Large diff, no logic.
  2. Freeze by code: the nine hook-call deletions, the slash callback's rate rewrite removed,
     the admin gates on the two ICQ messages, the calibration cap lifted, and the tests that
     the rate does not move. Small diff, the one PR that needs the epoch machinery in the
     reviewer's head.
  3. Upgrade handler: the v35 package and wiring, and every handler helper with its unit
     test (autopilot param, ICA-host allow-list, wasm to gov, comdex flag, trade-route
     deletion, oracle deactivation, rate-limit removal, the two ICQ purges, the haqq delta
     table and its generator). The small state flips live here rather than in PRs of their
     own; the haqq reconciliation splits out as a seventh PR if it wants its own reviewer.
  4. Wind-down constants and the ICA-side admin txs: the constants file with the channel
     maps, the proto for all four messages (one proto-gen), and `MsgUndelegateFromValidators`,
     `MsgTransferFromIca` and `MsgTransferStaketiaClaimBalance` with types, keeper, CLI and
     tests.
  5. Sweep tx: `MsgSweepTokensOffStride` alone, with its skip rules, destination resolution
     and the batch-builder script that mirrors them. The one tx that moves user balances gets
     a review with nothing else in the diff.
  6. Release gate: the mainnet-export suite over the full handler, the coverage-check script,
     the changelog, and the two address constants once the accounts exist.
  The module-path bump to `/v35` stays outside all six as a manual step after they land.
- Which vouchers go on the sweep list (the whitelisted hosts' native tokens at least), sized
  from the export by value like the stTokens. Whether to whitelist the two Axelar channels for
  their 2 USDC (axelar uses coin type 118, so derivation would hold) is not worth a constant
  unless the export shows more.
- Identify the owners of the interchain accounts on Stride that hold stTokens (2.3k stATOM in
  one) and the 32-byte holders, and notify them.
- Whether the legacy claim module's 2022 airdrops are already expired (its REST query is not
  served; confirm from a full node). Its messages are removed either way.
- Keeper code left unreferenced by the removals (stakeibc's LSM liquid-stake entry points, the
  trade-route authz and ICA registration helpers, `EnableRedemptions`, the keeper functions
  behind the deleted hook calls) is dead but harmless; deleting it is a follow-up cleanup, not
  part of the upgrade.
