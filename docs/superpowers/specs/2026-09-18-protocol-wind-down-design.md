# Protocol Wind-Down: Migration to Osmosis

Status: approved design, planned. Large tier; the v35 work is written from this spec as six
stacked PRs (§12) with one implementation plan per PR
(`docs/superpowers/plans/2026-09-29-wind-down-pr1-remove-handlers.md` through
`...-pr6-release-gate.md`), and the notes carried over from the earlier plans in §13.

## §1. Goal

Wind Stride down and shut the chain off. Stop every flow that moves funds on behalf of users,
unbond every delegation, send the native tokens from the host chains straight to Osmosis, and
put them into transmuter pools where holders swap stTokens for the backing native tokens at a
fixed rate. stTokens and STRD still sitting in Stride accounts are sent to their owners on
Osmosis before the halt, and IBC vouchers go back to the chain they came from, so nothing a
user owns is left on a dead chain. The guiding constraints, in order: least risk of a bug that
loses funds, least new code, most reuse of code that already runs on mainnet, and a chain that
stops rather than one that idles for a year.

The shape is one upgrade, v35, followed by one ops window of roughly 35 days that ends with a
coordinated halt: the longest unbonding period (28 days on Juno and Sommelier) plus a few days
of ops on either side. The
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
  producing blocks, so nothing could be done through their ICAs anyway. Stakedym is halted
  on mainnet (its rate crossed its 1.1 max bound), stays halted, and is treated as deprecated:
  it is never unbonded further, drained or given a pool. The one exception is the
  redemptions already in flight when it halted: the operator finishes those cycles and the
  binary lets the claims pay out on the halted zone (§6, §9). Holders of stCMDX, stEVMOS, stSTARS, stUMEE and
  stDYM have no redemption path after this work, which matches their status today but is now
  a deliberate decision (for stakedym, a deliberate choice not to recover it).
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
- The record-driven unbonding path (`UnbondFromHostZone` → `GetUndelegateMessagesForAmount`
  → `GetTargetValAmtsForHostZone`) computes balanced targets from the zone total and errors
  when the delegation left after the unbond is not positive (`host_zone.go`), so a record
  whose amount is at or above the zone's remaining `TotalDelegations` fails on every day
  epoch and stays in `UNBONDING_RETRY_QUEUE` with its stTokens escrowed (audit finding STRIDE-07). The
  admin drain does not use that path; it matters only for a record left queued on a zone
  the drain has already emptied (§6, §7).
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
  validators and `TotalDelegations`, skipped with a log (never an upgrade error) when a listed
  validator is missing from the zone or a delta would drive a delegation negative (it does not
  detect a tracked amount that merely changed; an unbond in the gap moves tracked and on-chain
  by the same amount, so the delta survives it, and a slash in the gap is what the day-0
  refresh catches), with a mainnet-export test. Injective and Celestia use it.
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

State on mainnet (refreshed 2026-09-29 unless dated otherwise):

- Real unbonding times (host staking params, 2026-09-30): haqq_11235-1 7 days, osmosis-1
  and celestia 14, juno-1 and sommelier-3 28, every other in-scope zone 21, dydx-mainnet-1
  included. Stride's stored `HostZone.UnbondingPeriod` is stale on three zones (dydx 30,
  celestia 21, haqq 21); it only sets the submission cadence, `period / 7 + 1` day epochs
  (3 for osmosis-1, 5 for dydx, juno and sommelier, 4 for the rest), while the sweep waits
  for the completion time the host returns in the undelegate ack, so the stale values cost
  nothing. Day epochs roll over at 19:00 UTC, and a zone submits on the day epochs divisible
  by its cadence, so all three cadences coincide every 60 epochs, next on epoch 1500,
  Monday 12 October 2026 19:00 UTC.
- 127 open stakeibc user redemption records, 3 with a claim pending. `HostZoneUnbonding`
  records with a non-zero amount: 17 `UNBONDING_QUEUE` (cosmoshub-4 4, phoenix-1 3, ssc-1 3,
  juno-1 2, osmosis-1 2, dydx-mainnet-1 1, laozi-mainnet 1, sommelier-3 1), 4
  `UNBONDING_RETRY_QUEUE` (all cosmoshub-4), 49 `EXIT_TRANSFER_QUEUE` (unbonded, awaiting
  the sweep), and 3,661 `CLAIMABLE` (swept; the terminal status, kept until the cleanup
  deletes the epoch record).
- **The Cosmos Hub pipeline is stuck, and the cause is the exact failure this design's
  day-0 refresh exists to prevent.** Four records are in `UNBONDING_RETRY_QUEUE` (epochs
  1479 to 1483, ~61.6k ATOM) and four more are queued behind them (1484 to 1487, ~128k ATOM,
  34 user records): roughly $340k of redemptions nobody is receiving. The
  `undelegation_failed` events (five since 2026-09-06, one per four-day unbonding epoch) and
  the acknowledgements behind them show what happens: each epoch the Hub's undelegation is
  split into ICA batches of five validators; the batch that holds NodeGuardians
  (`cosmosvaloper1jst8q8…`, decommissioned, weight 0, recorded 32,766.32 ATOM against
  32,763.08 on chain, stored rate exactly 1.0) is rejected by the Hub with ABCI code 18,
  "invalid shares amount", and the other batches succeed. A zero-weight validator is drained
  in full first by the cascade, the full-drain amount is 3.25 ATOM more than exists, and
  `applySharesRoundingSafety` only buffers a validator whose stored rate is below one, which
  Stride still believes this one is not. Records satisfied by the successful batches move
  on; whatever the failing batch was to cover stays in retry, and every new epoch's records
  join it, so the stuck amount grows. The fix needs no code and is available today:
  `UpdateValidatorSharesExchRate` (CLI `update-delegation`) for NodeGuardians and for
  Forbole (58.22 ATOM, over by 0.006, same shape), which detects the 0.01% downtime slash,
  lowers the recorded delegation to the chain's, and stores the sub-one rate so the next
  full drain both fits and gets the safety buffer. Done on 2026-09-29 from the hot wallet
  (txs `591DFD17…` and `63238D95…`); both callbacks applied within minutes and the Hub host
  zone now records NodeGuardians at 32,763.075057 ATOM and Forbole at 58.214178, rates
  0.9999. The backlog should clear at the next Hub unbonding epoch, 2026-09-30 19:00 UTC;
  verify with the `undelegation_failed` event search and the retry records. Separately, six Hub validators carry a stale
  `DelegationChangesInProgress` (keplr 12, stakewithus 2, four small ones 1), Juno has 21
  and Haqq 30, all with open ICA channels and nothing unacked; those flags do not affect
  the pipeline (the capacity calculation ignores them) but they do make the slash callback
  skip a validator and would make the drain refuse it, which is why the handler resets them
  (§5).
- Staketia: 6 `UNBONDING_IN_PROGRESS`, 2 `UNBONDING_QUEUE`, 1 accumulating, 0 redemption
  records. Stakedym (re-measured 2026-10-01): 6 `UNBONDING_IN_PROGRESS` (records 1444-1464,
  about 3,615 DYM owed to 23 redemption records, all completing 2026-10-12 between 20:05 and
  20:11 UTC) and 1 empty accumulating record; nothing queued, and no new redemption is
  possible on the halted zone. These six are finished by the operator (§9).
- One open LSM deposit, stuck in `DETOKENIZATION_FAILED` since 2026-06-05: 67,850,952
  `cosmosvaloper1xwazl8…j02r/116327` (jabbey, about 67.8 ATOM) on the cosmoshub-4 delegation
  ICA, its 35.6 stATOM already minted and its tokens already in the redemption rate but not in
  `TotalDelegations`. The redeem failed with `not enough delegation shares` because the
  tokenize-share record holds 67,850,951.998 shares, the rounding failure v23, v25 and v32
  fixed by retrying with one token less (§5).
- Zero auctions, zero ICQ-oracle price queries; both airdrop-module
  airdrops ended December 2024. Three ICA oracles active (injective-1, neutron-1, osmosis-1).
  Pending ICQs: haqq_11235-1 has 15 queries open, the oldest submitted on 2026-09-21, none
  answered; withdrawal and fee balance queries are open on haqq, juno-1, comdex-1 and
  laozi-mainnet.
- **Haqq is barely producing blocks.** Its height advanced 2,382 blocks between 2026-09-21
  and 2026-09-30, about one block every five minutes against a normal six seconds, and no
  ICQ response for haqq has landed on Stride in that time: the seven slash refreshes sent on
  2026-09-29 (§9a) applied on Terra and Injective within minutes, but on Haqq only the ones
  whose second query happened to be answered did. Nothing on Haqq can be fixed by a
  refresh while this lasts, which is the reason the haqq delta table at the upgrade covers
  every drifted validator rather than only the dust cases. Slow blocks do not make relaying
  impossible, only late; what breaks Haqq is our own ICA timeouts. A stride-epoch ICA times
  out about 4.8 hours after submission (next epoch start minus a fifth of the epoch,
  `GetICATimeoutNanos`), a day-epoch one after about 19 hours, and a timeout closes the
  ordered channel. On 2026-09-27 the delegation ICA channel (channel-869, the eleventh on
  that connection) closed on a timed-out `MsgWithdrawDelegatorReward` from
  `ClaimAccruedStakingRewards` and the withdrawal ICA channel (channel-668, the sixth) on a
  timed-out fee-split `MsgSend`; neither has been restored, so Haqq's delegation and
  withdrawal ICAs are unusable today and the zone's 30 stale in-progress flags are the
  packets that were in flight on those channels when they closed (their timeouts are never
  relayed on a closed channel; `RestoreInterchainAccount` is what resets them). Any ICA the
  epoch hook keeps sending from these accounts will close a restored channel again the
  first time delivery takes longer than 4.8 hours, which on Haqq now is routine. If Haqq
  stops for good before the drain, stISLM (about $417k) joins the unrecoverable set with the
  four dead zones (§12).
- Native vouchers stranded on Stride are dust: the eleven deposit addresses, the reward
  collector and the auction module together hold about $3 of in-scope native denoms. Nothing
  on Stride except the staketia claim address will hold a native balance worth moving. User
  accounts hold more of the host tokens, but of the four hosts whose vouchers the sweep cannot
  unwind (their wallets are not coin type 118, §7) the whole holding is about $3.2k, nearly all
  INJ dust (measured 2026-09-29, figures in §7).
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

| Name           | Chain   | Type                        | Status                 | Constant               | Address                                                                                                     | Role                                                                                                                                                                                                                          |
| -------------- | ------- | --------------------------- | ---------------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Protocol admin | Stride  | key (F5) and the gov module | exists, `utils.Admins` | `utils.Admins`         | `stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh` (F5), `stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl` (gov) | Signs `MsgUndelegateFromValidators`, `MsgTransferFromIca`, `MsgTransferStaketiaClaimBalance`, and the two admin-gated ICQ messages.                                                                                           |
| Sweep operator | Stride  | key                         | exists                 | `SweepOperatorAddress` | `stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9`                                                             | The only address that can sign `MsgSweepTokensOffStride`. Separate from the protocol admin so the sweep, the one tx that moves user balances, has its own key and its own blast radius. Holds STRD for fees only.             |
| Osmosis vault  | Osmosis | the protocol-admin multisig | exists (same key set)  | `OsmosisVaultAddress`  | `osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af`                                                               | The protocol-admin multisig re-encoded with the `osmo` prefix (same 20 bytes, same signers). Receives every ICA transfer, instantiates and funds the pools, holds the alloyed assets, and is each pool's admin and moderator. |

The Osmosis vault is deliberately the same multisig as the protocol admin, so no new key set
is created for it; both constants are still proven before the release-gate PR merges (a
signed spend from each on its own chain) and double-checked against this table (§9). The
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
  `ToggleTradeController`, `RebalanceValidators`, `ClearBalance`.
- staketia and stakedym: `RedeemStake` (their `LiquidStake` is already a disabled stub; remove
  the stub too), `ResumeHostZone`.
- icaoracle: `AddOracle`, `InstantiateOracle`. icqoracle: `RegisterTokenPriceQuery`,
  `RemoveTokenPriceQuery`. auction: `PlaceBid`, `CreateAuction`, `UpdateAuction`. airdrop:
  all seven. claim (legacy): all four.

`ClaimUndelegatedTokens` is kept: it pays the redemptions that finish after the upgrade, it
always was permissionless, and it is a no-op once the last record is claimed.
`RestoreInterchainAccount` is kept for the drain's timeout path (§7). `ResumeHostZone` is kept for stakeibc: a zone can still halt during the ops window through the safety-bounds check (a mis-set `UpdateInnerRedemptionRateBounds` is enough), a halted zone drops out of every kept epoch flow, and without the resume tx only an upgrade could bring it back.

**Messages gated.** `UpdateValidatorSharesExchRate` and `CalibrateDelegation` gain
`utils.ValidateAdminAddress` in ValidateBasic: ops use them to refresh slashes before the
drain (§9), and nobody else can reach the slash path. The 5,000 base-unit
`CalibrationThreshold` check is removed from the calibration callback: it existed to bound
what a permissionless caller could move. New calibration queries snapshot the validator's
recorded delegation in `DelegatorSharesQueryCallback`. The callback discards a response if
that delegation changed while the query was in flight or a delegation-changing ICA is
still active; this also catches a completed undelegation whose in-progress counter is zero.
Missing or malformed snapshots and nil or non-positive stored rates are successful no-ops.
An admin may submit a fresh calibration after a stale response is discarded; the callback
does not start a retry loop. `MsgCalibrateDelegation` also takes an optional `reset_delegation_changes_in_progress` (default false) that zeroes the validator's flag before the query is submitted, for a flag known to be stale. The tx rejects the reset unless the zone's delegation channel is open with no packet commitment outstanding, the same condition the upgrade handler's reset uses: with a packet in flight the flag is not stale, and its ack would fail on the zeroed counter and wedge the ordered channel; with the channel closed, `RestoreInterchainAccount` resets every flag anyway.

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
limit would protect. The module and middleware stay in the stack with empty state. Stakedym, whose rate stays
above its max bound, re-blacklists `stadym` every block, which is left as is (stakedym is deprecated, §2), so
the post-upgrade blacklist on mainnet holds exactly `stadym`.

**Comdex.** The handler sets `Deprecated = true` on comdex-1 so it carries the same flag as
the other three deprecated zones. `Halted` is not touched; the flag is documentation.

**Stale in-progress flags.** For every in-scope zone whose delegation ICA channel is open
with no unacked packet, the handler resets `DelegationChangesInProgress` to zero on every
validator, exactly what `RestoreInterchainAccount` does after a channel restore, and logs
each reset. A flag with no ICA behind it is stale by definition (the callback that clears it
can never fire). Stale flags do not stop the record pipeline, but they make the slash
callback skip that validator (so a day-0 refresh would silently not apply) and make the
drain refuse it, and today the Hub, Juno and Haqq carry dozens of them (§3, §9a). A zone
with an unacked packet is skipped and logged; ops clear it with the restore flow after the
upgrade.

**Pending ICQs.** The handler throws out three kinds of pending stakeibc interchain query. For
haqq_11235-1 it deletes every slash-path query (validator exchange rate, delegator shares,
calibration callbacks) and clears `SlashQueryInProgress` on every haqq validator, before the
delta table below is applied: a query submitted against the pre-delta state has no reason to
exist, and unlike v34's `DeleteStuckQueries` and `ResetStuckSlashQueries`, which pin query ids
and validator addresses, this is dynamic and cannot go stale between measurement and
execution (haqq had 10 such queries open and one flagged validator on 2026-09-29). For every
zone it deletes the pending withdrawal-balance queries, because their callback delegates and
the call that submits them is gone (§6), and all pending calibration queries, because old
permissionless requests lack the delegation snapshots now required by admin calibration.
Queries belonging to other modules are preserved even when their callback IDs collide.
Late responses to deleted queries remain successful no-ops in the ICQ message handler.

**Haqq delegation reconciliation.** The handler applies a per-validator delta table to
haqq_11235-1 with the v34 helper, exactly as v34 did for Injective. The 2026-09-29
measurement (§9a) has 16 validators off: 12 over-recorded (undetected downtime slashes and
sub-token rounding, the largest 853.8 ISLM) and 4 under-recorded by sub-token dust. Both
signs are applied so every tracked delegation equals the chain's; the regenerated table's net
is a decrease of 1,393.47 ISLM (2026-09-29; gmocoin's slash was booked on chain between
measurements), so `TotalDelegations` drops. The rate update is deleted in the same
upgrade (§6), so this no longer flows into the stISLM redemption rate; the rate stays about
0.0014% above the backing (1,393 ISLM against 101M stISLM, roughly $6), which the rewards
accrued until the drain cover many times over and the coverage check (§10) reports either
way. The table also pins each row's tracked delegation
(`HaqqExpectedTrackedDelegations`, each row's `recorded` value from the same `drift.json` as
its delta; a host zone file passed to the generator only cross-checks them) and skips,
with an error log, each row whose tracked delegation no longer matches its pin, so a slash
booked between generation and the upgrade cannot be applied twice; the other rows still
apply. The pin compares Stride's tracked value only: a slash on haqq that no query has booked
on Stride leaves its pin matching, the row applies, and the day-0 refresh books that slash
(§9c). The stored
`SharesToTokensRate` is deliberately left as is: the day-0 refresh updates it, and the slash
callback then finds tracked delegation equal to on-chain shares × the refreshed rate, so
nothing is applied twice. The table is generated from `measure_delegation_drift.py`,
re-measured right before the proposal, and covered by a mainnet-export test. This is the
Injective shape (real loss, no stranded liquid), not the v33 Osmosis shape (phantom stake
credited back as a deposit record).

**Failed LSM deposit.** The handler requeues the one `DETOKENIZATION_FAILED` deposit (§3) as
`DETOKENIZATION_QUEUE` with its amount reduced by one (67,850,951), as v32 did for the last
one, and only if the record is still exactly that status and amount. The EndBlocker retries
it on the next block; the success callback books about 67.8 ATOM to jabbey (unbonded and
jailed on the Hub, so its undelegation completes at once) and the drain unbonds it with the
rest. The rate is frozen, so nothing moves; one token of dust stays in the ICA. Ops drain
jabbey only after the retry's ack has landed, which the drain enforces anyway (it refuses a
validator with a change in progress).

## §6. What keeps running and what stops

The upgrade does not halt any zone. `Halted` would stop the flows the open redemptions need
along with the ones that must not run once the rate is frozen, so instead the binary deletes
the second group's call sites from `BeforeEpochStart` and leaves the first group in place.
Nothing is configurable and nothing can be toggled back; `Halted` stays false on every
in-scope zone. The keeper functions behind the deleted calls may stay
until the follow-up cleanup (§12).

Deleted from `BeforeEpochStart`:

| Call                                              | Why it must stop                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UpdateRedemptionRates` (stride epoch)            | The rate must not move once the drain starts: an admin undelegation lowers `TotalDelegations` with no record behind it and the formula would cut the rate. Deleting the call freezes `HostZone.RedemptionRate` at its last pre-upgrade value.                                                                                                                                                                                                                                                                  |
| `ReinvestRewards` (stride epoch)                  | It delegates the withdrawal ICA's rewards back to validators, which would create fresh delegations after the drain and need a second unbonding period. It is also the root of the fee machinery (§3): removing this one call stops the withdrawal-balance ICQ, the fee split, the fee-balance ICQ and the reward-collector inflow. Rewards simply accumulate in the withdrawal ICA and leave with `MsgTransferFromIca WITHDRAWAL`; what the fee ICA holds at the upgrade leaves with `MsgTransferFromIca FEE`. |
| `StakeExistingDepositsOnHostZones` (stride epoch) | New delegations after the drain. Deposits that reach the delegation ICA stay there and leave with the ICA balance.                                                                                                                                                                                                                                                                                                                                                                                             |
| `RebalanceAllHostZones` (stride epoch)            | Redelegations flag `DelegationChangesInProgress` and would collide with the drain.                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `TransferAllRewardTokens` (stride epoch)          | A no-op with the trade route deleted; removed for clarity.                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| `SetWithdrawalAddress` (stride epoch)             | Redundant (the withdraw address persists on the host, set years ago) and traffic on the delegation ICA channel the drain uses. A checklist line replaces it (§9).                                                                                                                                                                                                                                                                                                                                              |
| `CreateDepositRecordsForEpoch` (stride epoch)     | Nothing creates deposit amounts once liquid staking is gone.                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| `CreateEpochUnbondingRecord` (day epoch)          | Only `RedeemStake` appended to the current record; new ones would stay empty.                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| `AuctionOffRewardCollectorBalance` (mint epoch)   | It liquid stakes the validators' fee share, minting stTokens against a frozen rate from tokens that never get delegated. Validators are paid in STRD only from day 0 (§10a).                                                                                                                                                                                                                                                                                                                                   |
| `ClaimAccruedStakingRewards` (stride epoch) | An ICA `MsgWithdrawDelegatorReward` from the delegation ICA on every zone every stride epoch, with a 4.8-hour timeout. It is unnecessary (the undelegation withdraws rewards anyway) and it is what closed Haqq's delegation channel on 2026-09-27 (§3): one slow delivery kills the channel the drain needs. Nothing but the drain and the transfer tx should touch the delegation ICA after the upgrade. |

Kept in `BeforeEpochStart`:

| Call                                                                                                      | Why it stays                                                                                                                              |
| --------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `UpdateEpochTracker` (every epoch)                                                                        | The day-epoch tracker feeds the ICA timeout of every undelegate batch, including the drain's; the stride tracker feeds the sweep ICA.     |
| `InitiateAllHostZoneUnbondings`, `SubmitPendingUndelegations`, `CleanupEpochUnbondingRecords` (day epoch) | Submit the redemptions open at the upgrade, retry a failed batch, and delete records once every zone's unbonding is claimed.              |
| `SweepUnbondedTokensAllHostZones` (stride epoch)                                                          | Moves each completed record's amount from the delegation ICA to the redemption ICA, where the claim pays it.                              |
| `TransferExistingDepositsToHostZones` (stride epoch, deposit interval)                                    | Carries any native voucher still in a deposit address to the delegation ICA, so it leaves with the ICA balance instead of being stranded. |

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
drains what is left with an empty list and no reservation arithmetic. The order matters for
a second reason: the record-driven path computes balanced targets from the zone total and
rejects a non-positive remainder, so a record still queued after the drain could never be
submitted and its holder would never be paid. The undelegate tx therefore refuses a zone
with any queued or retrying record, and any validator with a change in progress, so a
mistimed drain fails rather than races or strands anyone (§7).
The redemption ICA is drained only after a zone's claims are done, since it holds the tokens
of claimable records.

Stakedym is the one halted zone with redemptions in flight (§3). Its hooks are untouched and
it stays halted, which already stops everything that mints, redeems or moves the rate.
Of the steps that finish a redemption, marking a record `UNBONDED` (hour epoch) and the
operator's `MsgConfirmUnbondedTokenSweep` ignore the halt; only `DistributeClaims` (hour
epoch) refused a halted zone. The binary drops that check: a claim pays a fixed native
amount the operator has already swept to the claim address and never reads the rate, so
paying it on a halted zone is safe, and it lets the six open records finish (§9).

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
for a validator that has drifted since. Three things reject the tx before anything is
submitted: a listed validator with `DelegationChangesInProgress` set, a deprecated zone, and
a zone with any `HostZoneUnbonding` record in `UNBONDING_QUEUE` or `UNBONDING_RETRY_QUEUE` with a
non-zero amount. The last one exists because of a known bug in the record-driven path
(`GetTargetValAmtsForHostZone` errors when the delegation left after an unbond is not
positive, so a record can never be submitted on a drained zone and would retry forever with
its stTokens escrowed and its backing already on Osmosis); the guard turns "drained too
early" into a rejected transaction instead of a stranded holder. The messages go through
`BatchSubmitUndelegateICAMessages` with no epoch unbonding record ids, in the zone's usual
batch size, and the tx registers the batches as in flight so the callback's record-less
accounting stays clean.

Submitting by validator rather than cascading a zone amount by weight is what lets ops test on
a single validator first, skip or shave a validator that is failing or drifted by dust, and
keep every submitted amount equal to what the host will accept. The live test on each zone is
a full drain of one validator with a small delegation and no unbonding entry in flight (the
pick per zone is §9b, about $1,400 across all eleven on the 2026-09-22 prices), not a partial
drain with an offset: a full drain exercises the exact
path the real run takes, and it spends one unbonding entry on a validator that then needs
nothing further. The SDK caps concurrent unbonding entries at 7 per delegator-validator pair
(the delegation ICA is the delegator); the day epoch adds at most one per validator and the
drain adds one, and today's worst case is 4 (ssc-1), so no check is needed: a batch that
hits the cap fails whole, and the cure is waiting up to four days for the oldest entry to
mature and resubmitting. Retry is resubmission by ops;
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

**`MsgTransferStaketiaClaimBalance { creator, amount }`**, where `amount` is an optional
integer of utia and zero means the whole balance. It moves that much of the staketia claim
address's TIA voucher balance to the stakeibc celestia zone's delegation ICA on Celestia,
where it unwinds to native TIA and is then sent on to Osmosis by `MsgTransferFromIca
DELEGATION` with the rest of the zone's balance. The optional amount exists for one reason:
a small test transfer first, since the balance is the whole staketia stake and the sender is
a multisig account the keeper signs for. Everything else is a constant or read from state:
sender is staketia's `ClaimAddress`, denom is `CelestiaNativeTokenIBCDenom`, the amount is
capped at the balance (reject if the balance is zero or below the amount), channel is the
celestia host zone's `TransferChannelId`, receiver is its `DelegationIcaAddress`, one-day
timeout, no memo. It calls
`transferKeeper.Transfer` with the claim address as sender, the pattern staketia's keeper
already uses for its deposit address (§3), so no multisig coordination is on the critical
path. A timeout refunds the claim address and ops resubmit. The same trick, sending a
Stride-side voucher to the zone's delegation ICA so it leaves with the ICA balance, is why
nothing else needs a Stride-side account; the claim address is the only Stride-side balance
worth it.

**`MsgSweepTokensOffStride { creator, denoms: [string], addresses: [string] }`**, the
batched token sweep. `denoms` is a non-empty list of bank denoms and `addresses` a non-empty
list with no on-chain bound (the tx is operator-gated and atomic, so a batch that exceeds the
block gas limit simply fails atomically; the batch size is an ops choice set from the localstride gas
measurement, and the builder defaults to 100). Taking a list of denoms lets the off-chain
builder walk holders once, by value, and sweep everything each holder has in one tx instead
of one pass per denom. Each denom's destination is decided once per tx, before any address is
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
  channels carry about 2 USDC. A denom with no destination rejects the whole tx: that is a
  bad batch, not a bad holder. Holders of those vouchers move them themselves before the halt
  (an ordinary IBC transfer to an address they type in works; only derivation does not). What
  that leaves behind is small: on 2026-09-29 the four non-118 hosts' native vouchers on Stride
  were 396.4 INJ across 47,821 holders (about $3,080, the largest 25 INJ), 12,531 ISLM across
  31 (about $49), 135.4 BAND across 10 (about $29) and 120.3 LUNA across 78 (about $6), about
  $3.2k in all at the §9b prices, so nothing is worth a per-chain exception.

Then, for every listed address, and for every listed denom it holds:

- Skip the address, with an event naming it and the reason, if it is not sweepable: it must
  decode to 20 bytes and its account must be a `BaseAccount` or one of the vesting account
  types; transfer escrow addresses, module accounts (including the deposit addresses, the
  distribution module and the reward collector), interchain accounts and unknown accounts are
  skipped. Skipping moves nothing, so a wrong address in a batch can never send funds
  anywhere; the event and the `num_skipped` count in the response are how the off-chain
  builder learns it disagreed with the chain, and the rest of the batch goes through. The
  on-chain rule stays the safety net; it just does not hold up the good addresses.
- Skip silently if the balance of that denom is zero (with a list of denoms this is the
  common case, not an anomaly).
- Otherwise submit an ICS-20 `MsgTransfer` through `transferKeeper.Transfer` of the full
  balance, sender the holder's Stride address, receiver the same 20 bytes bech32-encoded with
  the destination prefix, over the destination channel, with a one-day timeout and no memo. A
  transfer error here (a closed channel, a send disabled) is not a per-holder condition and
  rejects the whole tx. A timeout or a rejected receive refunds the holder on Stride through
  the normal ICS-20 path and ops resubmit that address.

Every destination is therefore the canonical form on Osmosis, the native form on the source
chain, or one hop closer to it; the sweep never adds a hop to any denom and never needs a
forwarding memo. One tx sweeps a batch of holders (size set from the gas measurement, no on-chain bound) across the listed denoms; the whole sweep at
the chosen floor is a few thousand packets over a few days (§3). There is no floor on chain: only the
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
(`docs/wind-down/sttoken-locations.md`; 28 route pools plus one canonical per stToken,
8 of the routes for stATOM). Each distinct Stride channel is a separate voucher route:
Axelar's stATOM channels 11 and 69 and Terra's stLUNA channels 13 and 52 each need separate
pools. A pool
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
(tested 2026-09-25). Completing an initial allocation after a test deposit therefore means
unmark, join, re-mark. The mark also closes the window in which an outsider can join a route
pool with native (§10), so the funding join and the mark go out back to back, and a route pool
is checked for outside alloyed shares before funding. If the test deposit's native balance
reaches zero, the contract removes
the native asset; `add_new_assets` at the same factor before the remaining funding join
restores it with the rate exact (tested). This completes the allocation, not a replenishment
of redeemed funds. The canonical denom's
counterfeit risk is a forged Stride header, handled by the key destruction after the halt
(§9).

**Funding** uses `join_pool` with native tokens only, once every source for that
native denom has arrived: the delegation ICA transfer and the withdrawal, fee and redemption
ICA sweeps (for stTIA the delegation ICA transfer includes the former multisig balance,
routed through the claim address). Each route pool receives exactly `escrow_route ×
RedemptionRate` native tokens, where `escrow_route` is the balance of Stride's escrow account
for that route's channel in the halt export; the canonical pool receives everything else
(§10). A stToken's pools can be created and funded as soon as its native denom is complete;
zones finish unbonding on different days and nothing couples them.

A small test deposit counts toward the pool's total allocation. The remaining deposit is
the allocation minus all prior confirmed native funding deposits, **not** the allocation
minus the pool's current native balance. For example, a 100-ATOM allocation funded with a
1-ATOM test receives another 99 ATOM, even if the test ATOM has already been redeemed.
Once a foreign-route pool has received its full allocation, do not fund it again: native
tokens paid out through redemptions are never replenished. The coverage check is not a
cumulative funding audit (§10).

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
and keep working in the pools. Clear channels in both directions before halting (§9): a late
packet toward Stride is not guaranteed to time out and refund on its source chain, since the
required non-receipt proof may need a Stride height or timestamp the halted chain never reaches.
The pools outlive the chain. Backing that is never claimed is
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

- A full accounting check: for every validator on every in-scope zone, Stride's recorded
  delegation equals the delegation ICA's on-chain delegation (the drift measurement, §9a),
  with any difference either in the haqq delta table or explained. The haqq delegation delta
  table, the drift measurement and the mainnet-export tests match the chain at one recent
  height. A haqq redemption unbonding at the upgrade would change the drift; measure right
  before the proposal. Regenerate in this order, immediately before the proposal: the drift
  measurement, then the delta table from that same `drift.json` (the deltas and the pins both
  come from it, so they cannot disagree), then `app/upgrades/v35/testdata/verify_constants.py`
  against the chain. Haqq's ICA channel stays closed until after the upgrade (§9c). A row
  whose validator's tracked delegation on Stride no longer equals its pin is skipped on its
  own (§5); otherwise the helper skips the table (never errors) only on a missing validator or
  a negative result, so a row whose on-chain side has moved (a slash on haqq not yet booked on
  Stride) is applied as is and the day-0 refresh books the rest; the `offset` on the drain tx
  is the last fallback.
- On every in-scope host, the delegation ICA's withdraw address (distribution module query) is
  the zone's withdrawal ICA; the epoch call that re-set it every epoch is deleted (§6).
- The two address constants, the channel map and `SweepUnwindChannels` in the binary
  re-verified: the maps against the hosts, each address by a test transfer of a few tokens to
  it from any wallet followed by a signed spend from it, so every constant is proven to be an
  address we control before anything is sent there.
- On every in-scope host, the ICA host module's `allow_messages` (REST
  `/ibc/apps/interchain_accounts/host/v1/params`) includes
  `/ibc.applications.transfer.v1.MsgTransfer`, and on osmosis-1 `/cosmos.bank.v1beta1.MsgSend`,
  since `MsgTransferFromIca` executes exactly those through the ICA (§7). A host that rejects
  the message leaves the funds in the ICA, so this is a delay, not a loss, but it belongs
  beside the channel check.
- The staketia operator ready to act on day 0, and the stakedym operator ready for
  2026-10-12, when stakedym's last six records finish unbonding (§9 step 7).
- A relayer on Stride `channel-197` ↔ Dymension from 2026-10-12 until the halt: it carries
  the stakedym operator's sweep to the claim address and the redeemers' DYM home (DYM is not
  on the sweep whitelist, since Dymension addresses do not share Stride's bytes, so redeemers
  move it themselves).

The ops window (after the upgrade, ~35 days). The redemptions open at the upgrade are queued,
unbonding, or unbonded and waiting for a sweep or a claim; the pipeline finishes all three
(§6) while the rest proceeds:

**Accepted LSM tail risk.** New LSM liquid stakes are disabled at the upgrade and none are
expected to remain, but an earlier request may still have callbacks or a deposit in flight.
Any such negligible remainder may finish through the existing pipeline while the chain is
running; if it does not finish in time, it is ignored. Do not add an LSM-specific drain or
halt prerequisite, or extend the ops window for it. Existing ICA-safety and coverage checks
remain unchanged; this is an accepted residual risk, not a guarantee that every LSM request
will settle.

1. Day 0, before anything else (including every other zone's refresh and calibration): the
   haqq sequence. Haqq goes first because its blocks are slow, its handshakes take hours, and
   it has a hard deadline: everything through 1f must finish before haqq's first unbonding
   epoch after the upgrade (2026-10-12 19:00 UTC if the upgrade lands between 10-08 and
   10-12). Missing that epoch moves the queued records, and the haqq drain behind them, back
   four days, to 10-16. Both light clients must stay alive through the sequence: haqq's
   client of Stride expires 2026-10-13 06:27 UTC and Stride's client of haqq 2026-10-17
   21:30 UTC unless relaying updates them. The haqq relayer comes back on at the upgrade
   (§9c). Background and the pre-upgrade rules are in §9c.
   a. Confirm the upgrade's haqq work: the `v35:` logs show the delta table applied, not
      skipped, and which rows (if any) were skipped on a stale pin. No haqq slash-path query
      remains, and no haqq validator has `SlashQueryInProgress`. A skipped row is one whose
      slash was already booked on Stride; the refresh in step 1f confirms it. If the whole
      table was skipped (a missing validator), stop and decide by hand before restoring.
   b. Close haqq's channel-29. The relayer submits `MsgChannelCloseConfirm` on haqq for
      `icahost`/channel-29, proving channel-869 CLOSED on Stride. This needs haqq's client of
      Stride Active. Verify channel-29 is CLOSED on haqq.
   c. Do not relay timeouts for packets 85-98. The restore zeroes their flags; a timeout
      processed after that hits the zero check in `DecrementValidatorDelegationChangesInProgress`
      and errors. Whether that is harmless for every packet type in 85-98 is unverified, so it
      is not left to chance.
   d. `strided tx stakeibc restore-interchain-account haqq_11235-1 connection-143
      haqq_11235-1.DELEGATION`. The four-step handshake runs across haqq's slow blocks (allow
      hours). It also updates Stride's client of haqq. Verify: a new delegation channel OPEN on
      both ends, `HostZone.DelegationIcaAddress` unchanged, every haqq validator at
      `DelegationChangesInProgress = 0`, and the queued records still in `UNBONDING_QUEUE`.
      Then the same for `haqq_11235-1.WITHDRAWAL`: its channel-668 is also CLOSED, and
      `MsgTransferFromIca WITHDRAWAL` (steps 5 and 6) needs it. It touches no delegation, so it
      can wait until after 1f if the handshakes are slow.
   e. Prove an ICQ round trip: `update-delegation` for one haqq validator. Wait for both the
      rate callback and the delegator-shares callback on Stride. If the second one keeps
      timing out, fix the relayer (frequent client updates toward Stride) before step 1f.
   f. The day-0 refresh on haqq for every validator, then the drift measurement (§9a). Require
      zero over-recorded haqq validators, because one fails the whole undelegate batch on the
      host.
   g. The haqq unbonding epoch submits the queued records and flags their validators. Wait for
      the acks, then drain haqq as in step 4.
2. Day 0: refresh every validator's exchange rate with `UpdateValidatorSharesExchRate` (CLI
   `update-delegation`) on every in-scope zone other than haqq (done in step 1) and wait for
   the callbacks. Where the rate
   moved, the callback applies the slash to the recorded delegation (the redemption rate is no
   longer rewritten, §6). Use `CalibrateDelegation` for a validator whose rate is unchanged
   but whose recorded balance is off. Then rerun the drift measurement (§9a) and require zero
   over-recorded validators on the zone. Same day: the staketia operator undelegates the
   entire multisig delegation on Celestia via authz (the queued records' amounts and
   everything else, in one go).
3. Day 0 to 4: the next day epoch submits the queued redemptions on every zone and flags their
   validators. Wait for those acks. A record whose submission fails (a slash between the
   refresh and the epoch) goes to `UNBONDING_RETRY_QUEUE` and is retried at the next day epoch; that
   zone's drain waits for it.
4. Then `MsgUndelegateFromValidators` per zone: first for a single small validator as a live
   test of the tx and the callback, then with an empty list for the rest. The tx refuses the
   zone while any record is still queued or retrying (§7), so step 3 cannot be skipped by
   accident. An ICA tx is atomic,
   so one over-recorded validator fails its whole batch; after the refresh there are none. If
   a slash lands between the refresh and the submission, that batch fails, ops rerun the
   refresh (or set that validator's `offset`) and resubmit for the affected validators.
   Delegation ICA channels and relayers stay healthy until every batch acks; a dead channel
   is restored with the existing flow and the affected validators are resubmitted. From here
   on no delegation exists that any record needs.
   **If a batch executes on the host but its ack never reaches Stride.** Acks cannot be
   delivered on a closed channel, so if a later packet's timeout closes the ordered channel
   first, the batches that did execute stay unacknowledged and the restore zeroes their
   flags: Stride still records the full delegation on validators the host has already
   unbonded. Calibration cannot repair a validator the drain emptied. The delegation no
   longer exists on the host, the query comes back empty, and the interchain-query module
   drops an empty response before the callback runs (a calibration that corrects to zero was
   written and reverted, 4150e47ba and c32c06c26, as too invasive a change for a
   contingency); where the drain left a remainder on the host, the query is not empty and a
   calibration does correct the record. So prevent it: when clearing a
   drain's packets, relay acknowledgements before timeouts, and do not relay a timeout while
   an earlier sequence's ack is outstanding. If it happens anyway, read the delegation ICA's
   delegations and unbonding entries on the host before resubmitting, and resubmit with an
   explicit validator list that leaves out every validator the host has already unbonded; an
   empty list would include them and fail their batches whole. Those validators' recorded
   delegations and the zone's `TotalDelegations` then stay overstated by the stranded amount.
   Nothing reads either once the rate is frozen and no record is queued, and the tokens still
   reach the delegation ICA when the unbonding completes and leave with its balance, so the
   backing is unaffected; the halt checklist takes it as a documented exception.
   **Timing: submit drain batches only between 19:00 UTC and about 08:00 UTC.** The drain
   reuses the epoch unbonding submitter, so every batch's ICA timeout is the next day-epoch
   start minus a buffer (a fifth of the epoch at the default `buffer_size` of 5), not a fixed
   duration. The mainnet day epoch rolls over at 19:00 UTC (3pm Eastern), so the timeout
   lands at about 14:12 UTC the next day: a batch sent just after 19:00 UTC has about 19
   hours, one sent at 13:00 UTC about an hour. Between about 14:12 and 19:00 UTC the
   computed timeout is already in the past; the send usually just fails, but if the
   counterparty light client lags the packet can go out, time out at once and close the
   ordered delegation channel, forcing a restore. Submitting after the 19:00 UTC day epoch
   fits the sequence anyway: that epoch submits the queued redemptions, their acks land,
   then the drain goes out with most of the window ahead. The ICA-transfer and
   claim-balance txs use fixed 24h and 48h timeouts and have no window. Confirm the
   mainnet `buffer_size` before day 0, since the REST params endpoint does not serve it.
5. Day 0+: `MsgTransferFromIca` for the withdrawal and fee ICA balances, including foreign
   denoms such as the dYdX USDC. This is the live test of the transfer tx on small real
   amounts, and the first arrival on Osmosis confirms the mapped channel end to end and the
   denom each zone lands as. Not yet the redemption ICA: it holds the tokens of claimable
   records until they are claimed.
6. Day 7 to 29, as each zone's unbondings complete (Haqq after 7 days, Osmosis and Celestia
   after 14, most zones after 21, Juno and Sommelier after 28): the pipeline sweeps each record's amount
   to the redemption ICA at the next stride epoch and the record goes `CLAIMABLE`; ops run
   `ClaimUndelegatedTokens` for every record (permissionless, as today; a record whose host
   receiver rejects the bank send is handled by hand). Once a zone has no record outside
   `CLAIMABLE` with a non-zero amount, no user redemption record, and no claim ICA in
   flight (confirmed by hand with the transfer checklist below, since the tx checks none of
   it): `MsgTransferFromIca DELEGATION` for the full
   remaining balance, `WITHDRAWAL` again for the auto-withdrawn rewards, and `REDEMPTION` for
   whatever dust the claims left. Then create and fund that stToken's pools (§8), the
   canonical one and one per in-scope route, once the coverage check passes (§10).
7. Staketia, day 21: the operator IBCs the whole unbonded balance via authz to the claim
   address (no memo; the grant forbids one) and `MsgConfirmUnbondedTokenSweep` for each open
   record; the hour-epoch hook pays the redeemers from the claim address. Then, once the
   transfer checklist below confirms every redeemer is paid,
   `MsgTransferStaketiaClaimBalance` with a small `amount` as the live test, and once it has
   landed on the delegation ICA, again with zero for the remainder, which leaves with that
   zone's balance in step 6. No key of the claim address signs anything.
   Stakedym, 2026-10-12 after about 20:11 UTC: the hour epoch marks records 1444-1464
   `UNBONDED`; the stakedym operator sends the unbonded DYM from the Dymension delegation
   account to the stakedym claim address over `channel-197`, exactly as in its normal cycle,
   and `MsgConfirmUnbondedTokenSweep` for each record (the claim address must hold at least
   the record's native amount). The next hour epoch pays the 23 redeemers from the claim
   address even though the zone is halted (§6), and the records archive as `CLAIMED`.
   Announce that redeemers must move their DYM to Dymension themselves before the halt.
   Nothing else on stakedym runs: no delegation, no new unbonding, no pool.
8. Last days: `MsgSweepTokensOffStride` in batches (size set from the gas measurement; the builder defaults to 100), each tx listing every
   denom on the sweep list, for every holder at or above the floor, built from a fresh
   export: the eleven stTokens and `ustrd` (to Osmosis) and every voucher whose outermost
   channel is whitelisted and that is worth sweeping (ATOM to the Hub, TIA to Celestia, and so
   on, each back one hop). Vouchers over
   the four host channels outside the whitelist are announced as self-service. Resubmit any
   address whose transfer timed out (its balance is back on Stride). Relayers on channel-5 and
   on every whitelisted channel stay up until the last packet acks.
9. Transfer-channel relayers stay up until the halt. Before stopping, check and clear the
   channels in both directions, including pending receives and acknowledgements, not just
   the outbound sweep packets. ICA channels can be left to close once every balance is sent
   and its acknowledgements are cleared. Announce that nobody should send funds to Stride
   near or after the halt: late inbound packets may be unrecoverable, and that sender risk
   is accepted rather than adding a new shutdown mechanism.
10. After the halt: every validator rotates or destroys its consensus key, and Stride Labs
   confirms it in writing from each. Osmosis's light client of Stride (`07-tendermint-2119`,
   12-day trusting period) accepts any header signed by two thirds of the last trusted
   validator set until it expires; a forged header could mint canonical stToken vouchers on
   Osmosis, which the pools would honour. Keys gone means the window is closed on day 0 rather
   than day 12. No relayer is asked to update that client after the halt.

Checklist before each balance transfer (steps 5 to 7). Neither `MsgTransferFromIca` nor
`MsgTransferStaketiaClaimBalance` checks anything against open records, by choice: mainnet
has had records sit for weeks behind a dead channel or a pending claim, and an on-chain guard
would turn one such record into a lock on the zone's whole balance with no way around it. A
transfer sent early is recoverable (the Osmosis vault sends the tokens back to the ICA on the
host and the pipeline retries by itself), but it stops every redeemer on that zone until
then, so confirm by hand before signing:

- `DELEGATION`: the zone has no `HostZoneUnbonding` record with a non-zero amount in any
  status before `CLAIMABLE` (`UNBONDING_QUEUE`, `UNBONDING_IN_PROGRESS`,
  `UNBONDING_RETRY_QUEUE`, `EXIT_TRANSFER_QUEUE`, `EXIT_TRANSFER_IN_PROGRESS`); REST
  `/Stride-Labs/stride/records/epoch_unbonding_record`, filtered on the zone. The sweep moves
  all of a zone's unbonded records out of the delegation ICA in one bank send, so a balance
  short by one record fails the sweep for every record, every stride epoch (injective-1
  sweeps one record per send, oldest first, so there a short balance fails only the records
  it cannot cover).
- `REDEMPTION`: the zone has no user redemption record left, and so no claim in flight
  (`claim_is_pending`); REST `/Stride-Labs/stride/records/user_redemption_record`. A record
  that cannot be claimed is settled by hand first. On 2026-10-02 one in-scope record had a
  claim pending, `dydx-mainnet-1` epoch 1436 for 5,752 DYDX, and the count of pending claims
  had not moved since 2026-09-29 (§3), which suggests it is stuck rather than in transit. It
  blocks the halt checklist too, so resolve it early.
- `WITHDRAWAL` and `FEE`: nothing to confirm; no redemption is paid from either.
- `MsgTransferStaketiaClaimBalance`: staketia has no redemption record left (REST
  `/Stride-Labs/stride/staketia/redemption_records`), meaning the hour epoch has paid every
  confirmed unbonding record and archived it as `CLAIMED`. An omitted or zero `amount` means
  the whole balance, so the small live test passes an explicit amount.

Checklist to halt the chain:

- Zero stakeibc user redemption records; zero `HostZoneUnbonding` records outside `CLAIMABLE`
  with a non-zero amount (`CLAIMABLE` is the terminal status of a swept record; the user
  records under it are what the claim deletes); no claim ICA in flight. Staketia:
  zero unbonding records outside `CLAIMED`, zero redemption records, claim
  addresses drained.
- No undelegate batch in flight (no validator with `DelegationChangesInProgress`) and
  `TotalDelegations` at dust on every zone except celestia, where it still carries the
  multisig portion (the per-validator unbond only touches ICA validators), and any zone with
  a stranded drain ack, which stays overstated by exactly that amount (step 4).
- All four ICA balances at dust on every zone. Celestia multisig delegation zero, its
  unbondings complete, its balance on Osmosis.
- Every pool, canonical and per route, created, funded and passing the coverage check (§10)
  against a fresh export.
- For each foreign-route pool, confirmed native funding deposits, including any test deposit,
  total exactly its allocation (§8); no redemption payouts have been replenished. Check the
  deposits, not just the remaining pool balance.
- The sweep complete: no sweepable account at or above the floor holds any in-scope stToken,
  `ustrd`, or a whitelisted voucher on the sweep list, and no sweep packet outstanding on
  channel-5 or a whitelisted channel.
- Channel queues checked and cleared in both directions: no outstanding receives or
  acknowledgements on the channels used during the ops window, including inbound transfers
  to Stride. Do not assume a packet left toward the halted chain will automatically refund.
- Validators and STRD delegators have withdrawn their rewards; interchain-account holders
  have been notified and given time to move out.

### §9c. Haqq: keep the channel closed, then close and restore after the upgrade

State on 2026-10-01:

- Stride's haqq delegation channel-869 is CLOSED (packet 84 timed out). Packets 85-98 are
  still committed on Stride and were never delivered; haqq's end of the channel (icahost
  channel-29, connection-8) is still OPEN. The earlier haqq delegation channels 647, 667
  and 837 are closed on both ends.
- 30 haqq validators carry `DelegationChangesInProgress = 1` for those packets. The v35
  reset skips haqq because its channel is not open (§5), so these flags survive the
  upgrade. Only `restore-interchain-account` clears them.
- 3 validators (Neuler, Kioqq, StakingCabin) are flagged `SlashQueryInProgress` with a
  delegation query open, plus about a dozen other haqq queries. The upgrade deletes all of
  them and clears the flags (§5).
- Two haqq redemption records are queued: epochs 1487 and 1488, 3,722.03 stISLM for
  3,947.22 ISLM. Haqq unbonds on day epochs divisible by 4 (unbonding period 21, frequency
  4): 19:00 UTC on 2026-10-04, 10-08, 10-12, 10-16 and so on. The 2026-09-30 epoch left
  them queued, because a closed channel fails the submission without changing state.
- Light clients. Stride's client of haqq (`07-tendermint-143`, connection-143) is Active
  with a 17.8-day trusting period, but was last updated 2026-09-30 01:30 UTC, so it expires
  2026-10-17 21:30 UTC unless something updates it. Every relay toward Stride does (an ICQ
  answer, an ack, a handshake step). Haqq's client of Stride (`07-tendermint-6`, 11.9-day
  trusting period against Stride's 14-day unbonding) was updated 2026-10-01 09:27 UTC and
  expires 2026-10-13 06:27 UTC if nothing relays toward haqq; closing channel-29 and the
  restore handshake both need it. The stale Stride side is consistent with no haqq ICQ
  being answered since 2026-09-30.
- Haqq has since cut its unbonding to 7 days (`unbonding_time` 604800s). Stride's client of
  haqq was created for the old 21 days (`unbonding_period` 1814400s) and expiry follows its
  own 17.8-day trusting period, so the dates above stand, but that trusting period is now
  longer than haqq's unbonding: a validator set that unbonded more than 7 days before a
  header could sign a forged one without being slashable. The tail risk is a forged ICQ
  proof or ack on Stride; keeping the client fresh shrinks the window. Stride's host zone
  still records `unbonding_period` 21, so the unbonding frequency stays 4, and haqq
  unbondings complete in 7 days rather than the 21 Stride expects.
- Haqq blocks are slow and bursty: about 400 a day (one every ~3.5 minutes on average),
  with stretches of one every 6 minutes and bursts of one every 25 seconds. Handshakes and
  ICA packets only take longer. A delegation ICQ has a 1-hour timeout and is resubmitted
  (`interchainqueries.go:135`), so it lands only when the relayer gets a fresh haqq header
  and proof in within the hour.

Before the upgrade: leave the channel closed, and do not restore it. Restoring reopens the
v34 epoch flows on haqq: the 2026-10-04 and 10-08 epochs would submit the queued
unbondings, and delegation, reinvestment and rebalancing would follow. Each changes
tracked delegations, and every row whose tracked value moves is skipped (§5). With the channel closed, a
slash on haqq is the only way the ICA's delegations can change. That moves the on-chain
side only, so the table still applies and the day-0 refresh books the slash. What does
skip a row is a Stride-side change to it before the upgrade: a haqq delegation query or
calibration answered, or anyone sending `update-delegation` or `calibrate-delegation` for
a haqq validator. Ops send none. On the last day of the vote, re-read the haqq host zone
against `HaqqExpectedTrackedDelegations`. Both `restore-interchain-account` and closing a
channel are permissionless. Restoring early is blocked in practice while channel-29 is
OPEN, because haqq's ICA host refuses a second active channel for the same owner. A third
party could still close channel-29 and then restore, so watch for a new haqq channel during
the vote.

The haqq relayer (packets and ICQ responses) is off until the upgrade, so no haqq query can
be answered and no row can move. It comes back on at the upgrade, not at the deadline: step 1
of the ops window needs hours of handshakes before 2026-10-12 19:00 UTC. While it is off:

- Neither haqq light client is updated, so the upgrade and the relayer's return must both
  come before 2026-10-13 06:27 UTC (haqq's client of Stride). If the vote slips, send a bare
  client update in both directions a few days before that. It moves no packet and answers
  no query.
- Transfers on channel-240 time out and refund, and haqq liquid stakes wait in their
  deposit records.
- Nothing may be sent on haqq's open ICA channels (FEE channel-614, REDEMPTION channel-244,
  COMMUNITY_POOL_DEPOSIT channel-245, COMMUNITY_POOL_RETURN channel-246). An unrelayed packet
  times out once the relayer is back, and the timeout closes that ordered channel. Those
  ICAs are sent from query callbacks or for completed unbondings, so with no query answered
  and no haqq unbonding in progress, none should be sent; check their packet commitments
  before turning the relayer back on.
- Anyone can run a haqq relayer, so turning ours off does not guarantee nothing is relayed.
  The per-row pin check is the backstop.

After the upgrade, the haqq sequence is step 1 of the ops window (§9).

### §9b. Live-test validator per zone

<!-- live-test-validators:start -->

Generated 2026-09-29 by `scripts/wind-down/pick_live_test_validators.py`; rerun on the day, after the day-0 refresh.
Prices: the snapshot in sttoken-locations.md (total USD / (supply × rate) per token).

The first `MsgUndelegateFromValidators` on each zone drains exactly one validator in full, as the live test of the tx
and its callback, before the empty-list drain of the rest (spec §7, §9 step 4). The pick is the validator with the
smallest recorded delegation of at least one whole token that has no unbonding entry in flight from the delegation ICA (the SDK allows 7
concurrent entries per delegator-validator pair; a validator drained in full never needs a second one). "Next" is the
second-smallest delegation, to show how much the pick matters.

| Zone           | Validator                                                             |                 Recorded delegation |     USD | Entries in flight | Funded validators | Next smallest (USD) |
| -------------- | --------------------------------------------------------------------- | ----------------------------------: | ------: | ----------------: | ----------------: | ------------------: |
| celestia       | mhventures (`celestiavaloper1q2kaajedxm0r5xc0twdqz6atap96502d67yjyj`) |                     15,861,063 utia |   $7.07 |                 0 |                92 |               $8.48 |
| cosmoshub-4    | icycro (`cosmosvaloper1ukpah0340rx7k3x2njnavwyjv6pfpvn632df9q`)       |                    14,207,897 uatom |  $24.86 |                 0 |                64 |              $26.51 |
| dydx-mainnet-1 | luganodes (`dydxvaloper1fs0t34g628xdqc8alfefnadq2x3qawt8g88mav`)      |    14,328,832,501,947,858,372 adydx |   $1.94 |                 0 |                25 |               $2.25 |
| haqq_11235-1   | digiser2 (`haqqvaloper1nekpsmetpxx2crsuzznuy4epv9eqvj03rtmxae`)       | 3,877,996,874,608,234,702,464 aISLM |  $15.10 |                 0 |                41 |              $15.10 |
| injective-1    | autostake (`injvaloper1acgud5qpn3frwzjrayqcdsdr9vkl3p6hrz34ts`)       |       9,813,332,455,629,722,717 inj |  $76.25 |                 0 |                38 |              $76.79 |
| juno-1         | cosmosspaces (`junovaloper1836fhsg6yqpu98vezfc7caakchqe8pvske7t8q`)   |                 9,007,060,654 ujuno |  $84.44 |                 0 |                21 |              $95.05 |
| laozi-mainnet  | meria (`bandvaloper1plau7keptn9qdt7nmhphltakv5t054f8lgwjdn`)          |                 4,158,959,467 uband | $897.36 |                 0 |                32 |             $897.36 |
| osmosis-1      | haannode (`osmovaloper1hqqzynrdqxzky82mw92ugwsrry0ntrse84g5nr`)       |                 7,791,194,655 uosmo | $285.37 |                 0 |                23 |             $485.13 |
| phoenix-1      | coinhall (`terravaloper1ge3vqn6cjkk2xkfwpg5ussjwxvahs2f6at87yp`)      |                   411,068,185 uluna |  $21.96 |                 0 |                40 |              $27.49 |
| sommelier-3    | ztakeorg (`sommvaloper13ul4wf2gwuwfrqrx70h2j9evje05vtglpc44sc`)       |                 2,450,134,223 usomm |   $0.00 |                 0 |                19 |               $0.00 |
| ssc-1          | solva (`sagavaloper13pcp0cstupahzz3n36x0dlhpa9fr8m9vlcy78y`)          |                   117,083,715 usaga |   $4.41 |                 4 |                16 |              $26.46 |

Total value put at risk by the eleven live tests: $1,418.75.

<!-- live-test-validators:end -->

### §9a. Drift measurement

Measured 2026-09-29, after v34, for every in-scope zone: recorded per-validator delegations
versus the delegation ICA's on-chain balances (`scripts/wind-down/measure_delegation_drift.py`,
rerun as the gate before the drain).

- Exact: celestia (the v34 phantom stake is gone), dydx-mainnet-1, osmosis-1, sommelier-3,
  ssc-1.
- cosmoshub-4: 2 over-recorded (NodeGuardians 3.25 ATOM, Forbole 0.006 ATOM), 5 under by at
  most 9 uatom.
- injective-1: sub-token dust both ways (7 over, 6 under, all below 1e-15 INJ), plus 606 wei
  on chain against a validator Stride does not track.
- juno-1: 1 over by 9.39 JUNO, 20 under (largest 6.96 JUNO; under-recording is harmless for
  unbonding, the ICA holds more than Stride will ask for).
- laozi-mainnet: 4 under by dust. phoenix-1: 2 over (0.20 LUNA largest, 0.40 in total).
- haqq_11235-1: 13 over (largest Neuler at 853.8 ISLM, 0.01% of that validator), 3 under by
  sub-token dust, net over by 1,758.26 ISLM, unchanged since the 2026-09-21 measurement
  (one validator has since moved to exact).

Refreshes run on 2026-09-29 from the hot wallet, after this measurement: the Hub's
NodeGuardians and Forbole (§3), Terra's Mario and Y, Injective's Allnodes, and Haqq's kioqq,
gmocoin, StakingCabin and "shut down". Every non-Haqq one applied within minutes and the
record now equals the chain (Y was rewritten to the live number, which had grown since the
measurement). On Haqq only gmocoin completed; kioqq and StakingCabin got their rate but their
shares query is unanswered, and "shut down" got nothing, because Haqq's chain is crawling
(§3). The over-recorded validators a refresh cannot reach, because their stored rate already
equals the chain's, are Haqq's Neuler, SureStake and Islamic Staking and Injective's rounding
dust; the ones it cannot reach because of a stale in-progress flag are Juno's SmartNodes
(9.39 JUNO) and Haqq's TakamulFi (91 ISLM) and AlxVoy. The delta table and the drain's
`offset` cover all of these; the measurement is rerun before the proposal.

In every over-recorded case Stride's stored exchange rate is above the chain's, i.e. an
undetected downtime slash, and recomputing each validator as on-chain shares × the chain's
current rate reproduces the on-chain balance exactly, so the day-0 refresh leaves no rounding
gap and no per-validator buffer is needed. The one wrinkle is haqq_11235-1's 18-decimal
denom: two validators are over by 203,557 and 3,216,141 aISLM with an unchanged rate, dust in
ISLM but above the calibration cap the upgrade removes, and even one base unit of
over-recording fails the host's share check. Haqq is therefore trued up by the delta table at
the upgrade (§5), which covers the dust cases too, and the `offset` on
`MsgUndelegateFromValidators` is the fallback for anything that drifts afterwards.

The same measurement shows how widespread stale `DelegationChangesInProgress` flags are:
cosmoshub-4 has 6 flagged validators (keplr at 12), juno-1 has 21 (every one at 2) and
haqq_11235-1 has 30, plus one haqq validator flagged `SlashQueryInProgress`; no zone has an
unacked ICA packet. On the Hub the flags have already stopped the pipeline (§3); on Juno and
Haqq they would stop the next submission or the drain the same way, which is why the
handler resets them (§5).

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

Before querying or summing Osmosis backing, `scripts/wind-down/coverage_check.py` validates
the complete pool topology. Its checked-in `REQUIRED_ROUTES` policy comes from the per-token
tables' exact `in scope` rows in `docs/wind-down/sttoken-locations.md`, excluding Stride and
Osmosis (canonical), and retaining every channel in multi-channel rows. The relayer scope
table is not the authority: temporarily blocked Injective and host-dust routes still require
pools, while unsupported Penumbra and all ignored foreign rows do not. A scope change must
update the reviewed tables and the policy together; an automated consistency test fails on
drift. The CLI always uses the trusted policy, with no operator override. Every non-deprecated
stakeibc export zone still requires canonical coverage, including stSOMM with an empty foreign
route set; a newly eligible token without a reviewed policy fails closed.

Each pool entry must name that token's own export host zone. Route channels must match the
approved set exactly and cannot repeat within a token; Osmosis pool IDs cannot repeat anywhere,
including between canonical and route pools or across tokens. Pool IDs must be canonical
positive decimal strings, and channels must use canonical `channel-N` spelling, so aliases
cannot evade the checks. A missing route fails validation even if canonical holds enough
native tokens for the entire supply. These checks precede the existing per-pool arithmetic:
the route upper bound remains its original escrow balance times the frozen rate. This checks
the current native balance, not cumulative funding: extra deposits after redemptions can
remain below that bound and pass. Ops therefore verify the total confirmed funding deposits
against the allocation and follow the no-replenishment rule (§8/§9); the checker does not
replace that funding check.

`join_pool` is permissionless, so the stTokens a pool holds are not all redeemed ones: a holder
who joined with stTokens and has not exited holds alloyed shares that still redeem for native.
The checker therefore adds each pool's alloyed supply not held by the vault to that pool's
native requirement, one share being one native base unit (the pool gate asserts the alloyed and
native factors are equal), so a join with stTokens can neither hide a shortfall nor block the
gate. A join with the native token into a route pool is possible only while native is not
marked corrupted (§8). The over-funded bound is unchanged, so the checker reports that route as
over-funded by the joined amount until the joiner exits or redemptions pay out that much, and
the vault cannot clear it by holding fewer shares, because the same amount then shows as a
shortfall.

Why a frozen rate is covered. Confirmed against cosmos-sdk v0.54.3: `Unbond` calls the
distribution hook `BeforeDelegationSharesModified`, which withdraws the accrued rewards, and
then removes the shares; rewards are computed from delegation shares only, so an unbonding
entry earns nothing during its 7 to 28 days. Between the upgrade and the pool being funded,
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
  a row skipped on a stale pin), the two ICQ purges (the haqq purge deletes only that chain's
  slash-path queries and clears the validator flags; the withdrawal-balance purge leaves every
  other query), the autopilot param, the ICA host allow-list, wasm params and contract admins,
  oracle deactivation, rate-limit removal, the stale-flag reset (cleared on a zone whose
  channel has no unacked packet, left alone on one that does), the two ValidateBasic gates
  and the lifted calibration cap; a compile-time guarantee that the removed messages no longer exist and a
  decode test proving historical txs still parse; existing keeper tests for the flows that
  keep running stay green.
- Freeze by code (§6): a test that each deleted call is absent from the hook and each kept
  call still runs on a non-halted zone; a keeper test that `RedemptionRate` is unchanged
  across a stride epoch with deposit records and rewards present; a slash-callback test that
  the delegation is corrected and the rate is not.
- Admin txs (§7): unit tests for gating on all four. Undelegate: per-validator message
  construction with and without offsets, empty versus explicit validator lists, rounding
  safety on a full drain, rejection of a validator with a change in progress, of a
  deprecated zone, and of a zone with a record in `UNBONDING_QUEUE` or `UNBONDING_RETRY_QUEUE`
  (accepted once every record is `UNBONDING_IN_PROGRESS` or later), acceptance on a
  non-halted zone, no accounting mutation, the in-flight registration. Transfer: every ICA type, a foreign denom, the osmosis-1 bank-send form, a
  chain id absent from the map, that the map has an entry for every in-scope zone and none
  for a deprecated one, that the Osmosis vault constant parses as an `osmo` bech32 address
  and the sweep operator constant as a `stride` one, and the built `MsgTransfer` fields
  (mapped channel, Osmosis vault, timeout, empty memo). Sweep, the highest-review item:
  table-driven tests for a base account and each vesting type (swept), an escrow address, a
  module account, an interchain account, a 32-byte address and an unknown account (each
  skipped with its reason in the event while the rest of the batch is sent and `num_skipped`
  counts it), a zero balance (skipped silently), a batch of more than 100 addresses (accepted: no bound), an invalid denom
  string, an empty denom list, a holder with two of three listed denoms (two transfers), a
  transfer error rejecting the whole tx,
  an stToken and `ustrd` landing on channel-5 with the `osmo` prefix, a single-hop voucher on
  a whitelisted channel landing on that channel with that chain's prefix, a two-hop voucher
  unwinding one hop over its whitelisted outer channel, a voucher whose outer channel is not
  whitelisted rejecting the batch, the derived address bytes, the full balance and only that
  denom being sent, and the ICS-20 refund on timeout returning the balance to the holder.
  Claim address: a zero `amount` sending the full balance and a positive one sending exactly
  that much, only the TIA denom moving, the built `MsgTransfer` fields (celestia channel,
  delegation ICA receiver, timeout, empty memo), a zero balance and an amount above the
  balance rejected, and the refund on timeout landing back on the claim address.
- Localstride run: upgrade with a redemption submitted beforehand, let the day epoch unbond
  it, drain the rest, let the sweep and claim complete, then one ICA transfer to a second
  local chain and one sweep batch whose packets are relayed and land at the derived
  addresses.
- Ops scripts: the coverage check and the batch builder are tested against a mainnet export
  and their output for the export is checked in beside the plan.

## §12. Open items for the plan

- **The Cosmos Hub unbonding pipeline is stuck today** (§3): ~190k ATOM of redemptions
  across eight epochs are retrying or queued because the batch that drains NodeGuardians in
  full is rejected every epoch (recorded 3.25 ATOM above the chain after an undetected
  slash). The permissionless slash refresh was run on NodeGuardians and Forbole on
  2026-09-29 and applied; confirm on 2026-09-30 after 19:00 UTC that the epoch's batches all
  acked with results, the retry records moved on, and no `undelegation_failed` event fired. Until it is fixed, Hub holders who
  redeemed in September are not being paid, and the drain of the Hub would be refused by the
  queued-record guard (§7). The stale in-progress flags on the Hub, Juno and Haqq are a
  second cleanup that the v35 handler does (§5). They accumulate when an ordered ICA
  channel closes on a timeout with other packets still in flight: those packets' timeouts
  are never relayed on the closed channel, so their callbacks never run and their flags
  stay until a restore resets them (§3). The Hub's 28 closed delegation channels and Haqq's
  11 are where its flags came from.
- **Stale flags the handler cannot reset** (measured 2026-09-29): haqq_11235-1's delegation channel-869 is CLOSED with 14 packet commitments (sequences 85-98) and 30 flagged validators, so haqq needs channel-29 closed and `restore-interchain-account` after the upgrade, before the day-0 refresh and its first unbonding epoch (§9c); juno-1's open channel-491 has pending commitments from sequence 5729, so its 21 flags clear only once those packets are relayed. The v35 reset skips both zones by design.
- **Haqq's chain health** (§3): with one block every few minutes since at least
  2026-09-21, no ICQ answered in that time, and both the delegation and withdrawal ICA
  channels closed since 2026-09-27, decide before the proposal whether haqq stays in scope.
  If it does, every haqq correction rides on the delta table (no refresh is
  possible before the upgrade) and the drain's ICA relaying must be watched by hand (§9c);
  Stride's client of haqq expires around 2026-10-17 if nothing relays toward Stride; if Haqq halts for good,
  stISLM moves to the unrecoverable set and its holders get no pool.
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
- Stranded stToken holders behind expired clients: Penumbra (Osmosis's and Stride's clients
  both expired; ~6.2k stATOM and stOSMO, ~$29k) will not be supported (decided 2026-09-30;
  status `ignored · unsupported` in the locations doc, left off the relayer map); Kujira (~5.8k stATOM, ~$20k) has no reachable RPC and
  looks stopped, so it is ignored (decided 2026-09-23). Agoric is fine for holders
  (Agoric→Osmosis is active) though its Stride hop is expired. The five deprecated zones and
  their stTokens (~$16.8k in total) are not touched by the migration: Evmos, Stargaze, Umee
  and Comdex have stopped producing blocks (unrecoverable, like Kujira), and Dymension is
  alive but left alone by choice. `docs/wind-down/sttoken-locations.md` separates ignored
  balances into small, unrecoverable, deprecated and unsupported.
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
  size after measuring gas (an ops choice, not an on-chain bound).
- The plans: one per PR, written from this spec on 2026-09-29 (the two plans written
  for the earlier two-upgrade sequencing were deleted the same day; what they had learned is
  in §13). The work is delivered as six stacked PRs, each reviewable on its own, in the order
  below. **Branching and order:** PR 1 branches off `wind-down-design-consolidation`; every
  later PR branches off the previous PR's branch (PR 2 off PR 1, PR 3 off PR 2, and so on),
  so each diff shows only its own change; and the PRs are implemented and merged strictly in
  sequence, never in parallel. Branch names: `wind-down-pr1-remove-handlers`,
  `wind-down-pr2-freeze-by-code`, `wind-down-pr3-upgrade-handler`, `wind-down-pr4-admin-txs`,
  `wind-down-pr5-sweep-tx`, `wind-down-pr6-release-gate`.
  1. Remove tx handlers: pure deletions across stakeibc, staketia, stakedym, icaoracle,
     icqoracle, auction, airdrop and claim, plus rebalance and clear-balance (staketia and stakedym also drop resume; stakeibc keeps it); types
     and registrations stay; the "no handler" guard test and the historical-tx decode test.
     Large diff, no logic.
  2. Freeze by code: the ten hook-call deletions, the slash callback's rate rewrite removed,
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
     the changelog, the two address constants once the accounts exist, and stakedym's
     `DistributeClaims` paying on a halted zone (§6).
     Its plan is written like the others, but it is the one PR that cannot merge until the
     two accounts exist and are proven (§9), since it fills their constants.
     The module-path bump to `/v35` stays outside all six as a manual step after they land:
     it touches every file and would make the stacked diffs unreviewable.
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
  part of the upgrade. Epoch unbonding records whose host zone entries never reach CLAIMABLE (zero-stToken entries
  are never submitted) are not deleted by `CleanupEpochUnbondingRecords`, so the record for the
  upgrade's day epoch stays in state permanently; harmless, and no new records are created after
  the upgrade.

## §13. Notes for the plan, carried over from the earlier plans and their dry run

The two plans written for the two-upgrade sequencing were executed once end to end on a
scratch branch (2026-09-21) and reverted, then extended for the second upgrade and finally
deleted when the design became one upgrade. Everything below was learned there and is not
derivable from the spec or the code at a glance.

Removals (PR 1):

- `RegisterHostZone`'s tests are keeper tests behind a pure-delegate handler: rewrite the 18
  `GetMsgServer().RegisterHostZone(` calls to the keeper, do not delete the file. The keeper
  method survives for `x/staketia/keeper/migration.go`.
- `community_pool.go`, `handler.go` (the legacy switch, dead since SDK 47) and
  `client/cli/tx_test.go` also reference the removed stakeibc handlers; `redeem_stake_test.go`
  goes through `GetMsgServer().RedeemStake` in 18 places and is rewritten to the keeper.
- `msgServer.LiquidStake` has to become `Keeper.LiquidStake(ctx, msg)`: autopilot and
  `community_pool.go` call it through `NewMsgServerImpl`. Keep `message_liquid_stake.go` and
  `message_redeem_stake.go` (their constructors and ValidateBasic are still used); the other
  removed messages' `message_*.go` files can go once `grep` shows only the deleted handler,
  CLI and tests referenced them. CLI flag constants that lose their only user
  (`FlagMinRedemptionRate`, `FlagMaxRedemptionRate`, `FlagCommunityPoolTreasuryAddress`,
  `FlagMaxMessagesPerIcaTx`, `FlagLegacy`) go with them.
- `ClaimUndelegatedTokens`'s handler lives in `keeper/claim.go`, not `msg_server.go`. It is
  kept in this design, but that is where to look.
- After a `.proto` edit `make proto-gen` (docker) may rewrite descriptor bytes in unrelated
  `*.pb.go`; commit only the `tx.pb.go` of the module whose proto changed and revert the
  rest. With the module path untouched the churn should be nil.
- The first dry-run pass dropped `RegisterImplementations` and broke `strided q tx` on old
  hashes; the decode test (§11) exists because of that.
- Unreferenced after the removals and left in place (the §12 cleanup): `StartLSMLiquidStake`,
  `SubmitValidatorSlashQuery`, `ShouldCheckIfValidatorWasSlashed`,
  `EmitPendingLSMLiquidStakeEvent`, `BuildTradeAuthzMsg`, `GetTradeRouteFromTradeAccountChainId`,
  `RegisterTradeRouteICAAccount`, `EnableRedemptions`.
- The full suite passes except `utils` `TestCreateModuleAccount`, which fails on main too.

Upgrade handler (PR 3):

- The trade-route store key is `RewardDenomOnRewardZone + "-" + HostDenomOnHostZone`;
  mainnet's route is `uusdc` / `adydx`, not the IBC hashes.
- Wasm admins move through `wasmkeeper.NewGovPermissionKeeper(...)`'s `UpdateContractAdmin`;
  the unit test instantiates a real contract (`hackatom.wasm` from the wasmd testdata) to
  prove it. The upload-access param write is the one step whose error fails the upgrade.
- ICA host params are `icahostkeeper.Keeper.GetParams/SetParams` with `AllowMessages`
  filtered by `sdk.MsgTypeURL`; `app/upgrades/v7` is the exemplar. The keeper is a pointer on
  the app (`app.ICAHostKeeper`).
- The rate limiter is ibc-go v11's rate-limiting middleware, not a local module:
  `GetAllRateLimits/RemoveRateLimit(denom, channelOrClientId)`,
  `GetAllBlacklistedDenoms/RemoveDenomFromBlacklist`,
  `GetAllWhitelistedAddressPairs/RemoveWhitelistedAddressPair(sender, receiver)`; `RateLimit`
  has `Path.Denom` and `Path.ChannelOrClientId`; methods have pointer receivers, so the
  handler takes `&app.RatelimitKeeper`.
- `icaoraclekeeper.ToggleOracle(ctx, chainId, false)` skips the channel validation that the
  `true` case does.
- The haqq delta table is generated by `scripts/wind-down/gen_delta_table.py` from
  `drift.json`, and the v34 `delegation_deltas.go` helper is copied verbatim into the v35
  package (`package` and log prefix renamed). The measurement gave 17 deltas, not the 14 a
  first look suggested: 3 are positive sub-token dust and both signs apply.
- The localstride upgrade dry run needs the binary trick recorded in the team memory notes
  (a handler-bearing binary panics at the plan height on a chain that never scheduled it).

Admin txs (PRs 4 and 5):

- All four messages go into `proto/stride/stakeibc/tx.proto` in one proto-gen; the keeper
  logic of each lives in its own `wind_down_*.go` file with a thin delegate in
  `msg_server_wind_down.go`, so the sweep PR touches only its own files.
- The two operator addresses and channel-5 are package `var`s so tests can set them; the
  addresses start empty and every use fails closed (the sweep gate rejects everyone, the
  transfer tx errors "osmosis vault is not configured"), and a release-gate test skips while
  they are empty and passes once filled.
- Host-side channels to Osmosis from the chain registry on 2026-09-24, to re-verify against
  each host before the PR: celestia channel-2, cosmoshub-4 channel-141, dydx-mainnet-1
  channel-3, haqq_11235-1 channel-2, injective-1 channel-8, juno-1 channel-0, laozi-mainnet
  channel-83, phoenix-1 channel-1 (terra2 lists four preferred channels; channel-1 is the
  original), sommelier-3 channel-0, ssc-1 channel-1; osmosis-1 maps to an empty channel.
- A record-less undelegate batch acks through the pending-undelegation branch of the callback,
  which decrements an in-flight counter and logs an error if none is registered; the tx sets
  `SetPendingUndelegationInFlight(current + batches)` after submitting. `SubmitTxsDayEpoch`
  is the submit path, so the day-epoch tracker must exist (it does; `UpdateEpochTracker`
  stays, §6).
- `applySharesRoundingSafety` divides by `UndelegationSharesSafetyDivisor = 1e17`, so on a
  full drain of a slashed validator the buffer floors to zero and becomes one base unit.
- `GetHostZoneFromHostDenom` returns `(*HostZone, error)`, not `(HostZone, bool)`.
- `x/stakeibc` importing `x/staketia/types` (for `ClaimAddress`,
  `CelestiaNativeTokenIBCDenom`, `CelestiaChainId`) is cycle-free; staketia only imports
  stakeibc types. The claim-address tx therefore lives in stakeibc with the other three.
- The ICA transfer is built exactly like `BuildHostToTradeTransferMsg` minus the memo and
  submitted with `SubmitICATxWithoutCallback(ctx, connectionId, owner, msgs, timeout)`,
  `owner = types.FormatHostZoneICAOwner(chainId, icaType)`; the four ICA addresses are
  separate `HostZone` fields (`DelegationIcaAddress`, `WithdrawalIcaAddress`,
  `FeeIcaAddress`, `RedemptionIcaAddress`) resolved by a switch.
- Sweep skip rules: escrow addresses are computed per transfer channel with
  `k.IBCKeeper.ChannelKeeper.GetAllChannelsWithPortPrefix(ctx, "transfer")` and
  `transfertypes.GetEscrowAddress`; the allowed account types are `*authtypes.BaseAccount`,
  the SDK's `Continuous/Delayed/PeriodicVestingAccount` and
  `*claimvestingtypes.StridePeriodicVestingAccount` (`x/claim/vesting/types`); interchain
  accounts are `*icatypes.InterchainAccount`. Denom traces come from
  `transfertypes.ParseHexHash` + `k.RecordsKeeper.TransferKeeper.GetDenom(ctx, hash)`, the
  pattern in `lsm.go`, and `Denom.Trace[0]` is the outermost hop.
- Test conventions: `s.CreateICAChannel(owner)` and `s.CheckICATxSubmitted(port, channel,
fn)` for ICA submissions; `s.CreateTransferChannel(chainId)` plus
  `s.MustGetNextSequenceNumber(transfertypes.PortID, ibctesting.FirstChannelID)` for ICS-20
  transfers (the real transfer keeper runs; sequence delta and burned/escrowed balances are
  the assertions); `s.App.TransferKeeper.SetDenom(ctx, transfertypes.NewDenom(base, hops...))`
  to register a voucher trace; `apptesting.GetAdminAddress()` for an admin signer;
  `sdk.MustBech32ifyAddressBytes(prefix, addr)` for the derived receiver; the batch-submit
  test in `unbonding_test.go` shows the day-epoch tracker mock the drain needs.

Ops scripts (PRs 5 and 6):

- `build_sweep_batches.py` applies the on-chain skip rules plus the dollar floor to an
  export and writes one address file per batch that the CLI reads (`sweep-tokens-off-stride
DENOMS FILE`, one batch of holders sweeping every listed denom they hold); extra denoms are passed as `--extra-denom DENOM=USD_PER_TOKEN:DECIMALS`, and an
  `ibc/` extra denom is refused unless its outermost hop is in `SweepUnwindChannels`.
  `coverage_check.py` is the §10 check over an export and the vault's Osmosis balances. Both
  are unit-tested against a synthetic export and their real output is checked in.
- The per-chain relayer scope is `build_relayer_scope.py` (§9); the pre-funding pool check is
  `check_transmuter_pool.py` (§8).
