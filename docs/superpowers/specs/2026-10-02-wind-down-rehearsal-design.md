# Wind-Down Rehearsal on the k8s Network

Status: design for review. Large tier; an implementation plan follows once this is approved.

Parent spec: `docs/superpowers/specs/2026-09-18-protocol-wind-down-design.md` (the "wind-down
spec" below; section references such as §9 point into it).

## 1. Goal

Rehearse the protocol wind-down end to end on the k8s integration network before it runs on
mainnet: the v35 upgrade, the last redemption cycle, the admin unbonding, the transfers to
Osmosis, the transmuter pools, the token sweep off Stride, the staketia exit and the halt.

It is an ops dress rehearsal, not a test suite. We walk the wind-down spec's §9 ops window in
order, with the same CLI commands and ops scripts the real run uses, and stop at an explicit
checkpoint after every phase. The point is to find what breaks in the procedure and its
sequencing guards, which unit tests cannot reach.

Ground rules:

- All work lives on the `wind-down-rehearsal` branch (worktree
  `stride-worktrees/wind-down-rehearsal`), cut from main after the wind-down PR (#1541) and
  the v35 version bump (#1543) merged. The branch is throwaway and is never merged to main.
- Implementation quality on the branch does not matter. Anything may be changed to make the
  rehearsal work.
- A bug or a runbook error found here is fixed on main in its own PR, not on this branch
  alone.

## 2. Network

| Chain | Stands in for | Change from today's k8s network |
| --- | --- | --- |
| Stride, 4 validators (up from 3) | mainnet Stride | Starts on the v34 release and upgrades to the rehearsal branch through a gov proposal. The fourth validator exists so one can be stopped for a `strided export` without halting the chain. |
| Hub, 8 validators (up from 1) | cosmoshub-4 and celestia | More validators than one undelegate ICA batch holds, so the drain splits into several batches. One validator has weight 0 and a delegation (the NodeGuardians shape from §3). |
| Osmosis, 3 validators (up from 1) | osmosis-1, both as a host zone and as the destination | Mainnet's transmuter bytecode (code id 996 on osmosis-1) uploaded and whitelisted for `cosmwasmpool` at genesis. |

- **Relayers.** The existing stride↔hub and stride↔osmosis paths, plus a new hub↔osmosis
  transfer path. `MsgTransferFromIca` sends host → Osmosis directly, and the foreign-route
  stATOM denom travels the same path.
- **Both transfer forms are covered.** The Hub zone exercises the ICA `MsgTransfer` form; the
  Osmosis zone exercises the ICA bank-send form.
- **Timing.** The network's existing short periods are the starting point (day epoch 180s,
  stride epoch 45s, unbonding 240s). The hour epoch must also be shortened, because staketia's
  `MarkFinishedUnbondings` and `DistributeClaims` run on it. The plan may lengthen the day
  epoch if the drain window proves too tight to work in.

### Keys

- **Protocol admin and Osmosis vault:** one 2-of-3 multisig, the same 20 bytes under the
  `stride` and `osmo` prefixes, as on mainnet. A multisig signs in amino-JSON, so this is the
  only way to catch a missing amino registration on the four new messages.
- **Sweep operator:** a single key, distinct from the admin.
- **Staketia on the Hub:** the delegation account is a 2-of-3 multisig that grants an operator
  key the mainnet authz set: delegate, undelegate, withdraw rewards, and a
  `TransferAuthorization` on the Hub → Stride channel allow-listed to the staketia claim
  address only, with empty `allowed_packet_data`.

## 3. What changes on the branch

Constants only, plus one timeout. The upgrade handler is not touched: every step is meant to
log and skip on missing state, so running it without haqq, comdex or the dYdX trade route
tests that claim.

| Constant | Mainnet value | Rehearsal value |
| --- | --- | --- |
| Protocol admin (`utils.Admins`) | F5 multisig | the test multisig |
| `OsmosisVaultAddress` | F5 with the `osmo` prefix | the test multisig with the `osmo` prefix |
| `SweepOperatorAddress` | mainnet sweep key | a test key |
| `OsmosisChainId`, `StrideToOsmosisTransferChannelId` | `osmosis-1`, `channel-5` | the k8s Osmosis chain id and channel |
| `HostToOsmosisTransferChannel` | ten mainnet zones | the Hub zone → its channel to Osmosis; the Osmosis zone → `""` |
| `SweepUnwindChannels` | seven mainnet channels | Stride's Hub channel → `cosmos`, its Osmosis channel → `osmo` |
| Staketia `CelestiaChainId`, `ClaimAddress`, `CelestiaNativeTokenIBCDenom` (and the copies in `x/stakeibc/types/wind_down.go`) | celestia values | the Hub zone, a test claim account, the ATOM voucher on Stride |
| `WindDownTransferTimeout` | 24h | about 15 minutes, so a timed-out transfer and its refund can be observed |

### Ops scripts

None of the scripts in `scripts/wind-down/` runs against the k8s network as written; each is
wired to mainnet through module-level tables. On the branch, four get their data tables
swapped and their logic left alone:

| Script | Table swapped | What the rehearsal tests that its unit tests cannot |
| --- | --- | --- |
| `measure_delegation_drift.py` | Stride REST URL, zone table, host endpoint discovery | It must find a slash we inject, then show zero over-recorded validators after the refresh. |
| `build_sweep_batches.py` | unwind-channel map, protocol-address deny-list | It claims to mirror the on-chain skip rules, so a batch it emits must produce `num_skipped = 0` on chain. |
| `coverage_check.py` | `REQUIRED_ROUTES` (stATOM → the Hub channel, stOSMO → none) | It runs against pools as they move through test deposit, funding, swaps and an outsider's join. |
| `check_transmuter_pool.py` | its constants block (REST URLs, code id, module address, vault, pool specs, channel map) | Least new coverage; it already ran against the mainnet test pools. |

Not used: `pick_live_test_validators.py` (the live-test validator is picked by hand),
`gen_delta_table.py` (haqq only) and `build_relayer_scope.py` (mainnet relayer scope). The
mainnet tables themselves are not tested by the rehearsal; they can only be checked against
mainnet.

## 4. State at the upgrade height

Seeded on v34 by a re-runnable script, so the network can be rebuilt after a restart. The
two stakeibc zones are registered, liquid stakes fund them, and redemptions are spaced over
several day epochs so that every record status is present at the upgrade height.

| What is open at the upgrade | What must happen after |
| --- | --- |
| Redemption swept but unclaimed (`CLAIMABLE`) | The claim pays on the host. |
| Redemption unbonded, not yet swept (`EXIT_TRANSFER_QUEUE`) | The stride-epoch sweep moves it, then the claim pays. |
| Redemption mid-unbonding (`UNBONDING_IN_PROGRESS`) | It matures, is swept and is claimed. |
| Redemption queued (`UNBONDING_QUEUE`) | The first post-upgrade day epoch submits it. |
| Redemption stuck in `UNBONDING_RETRY_QUEUE` behind a slashed, unrefreshed Hub validator | The day-0 refresh fixes the drift and the retry clears: the mainnet Hub incident replayed. |
| Liquid stake whose deposit record is still in flight | It reaches the delegation ICA, is never staked, and leaves with the ICA balance. |
| Rewards in the withdrawal ICA and a balance in the fee ICA | Never reinvested; both leave through `MsgTransferFromIca`. |
| Staketia unbonding records queued, in progress and accumulating, plus one redemption sized past `RemainingDelegatedBalance` so it spills into stakeibc | The operator finishes them and the redeemers are paid on Stride. |
| A rate limit on stATOM, autopilot stakeibc on, the ICA host allow-list at its default | The handler removes the limit, turns autopilot off and trims the allow-list. |
| stATOM held on Stride by a base account, a vesting account, a module account and an interchain account; stATOM on Osmosis, on the Hub, and sent Hub → Osmosis as a two-hop denom; STRD and ATOM vouchers on Stride | Sweep targets and skip cases; the canonical and foreign-route pools. |

The redemption rate of both zones is recorded just before the upgrade and asserted unchanged
at every later checkpoint.

## 5. Sequence

Each phase ends in a checkpoint that must pass before the next phase starts. Phases 3 to 6
overlap in time, as they do in §9; the order below is the order in which each one starts.

### Phase 0: pre-flight

The §9 "checklist to propose the upgrade", on the k8s network.

- Run the drift measurement on both zones.
- On each host, confirm the delegation ICA's withdraw address is the withdrawal ICA, and that
  the ICA host `allow_messages` includes `MsgTransfer` (Hub) and bank `MsgSend` (Osmosis).
- Prove the vault and sweep-operator constants: a test transfer to each, then a signed spend
  from each.
- Record both redemption rates.

### Phase 1: the upgrade

Pass the gov proposal and let the binary swap.

Checkpoint:

- Every handler step logs applied or skipped; none errors.
- Liquid stake and redeem stake fail to route as a direct tx, through autopilot and through
  the ICA host.
- The rate limit is gone, autopilot stakeibc is off, and wasm code upload is gov-only.
- A liquid-stake tx from before the upgrade still decodes with `strided q tx`.

### Phase 2: day 0

- The multisig admin refreshes every validator on both zones with `update-delegation`. The
  same tx from a non-admin key is rejected.
- Attempt the drain and confirm it is refused, because records are still queued or retrying.
- The staketia operator undelegates the whole Hub multisig delegation through authz and
  confirms the queued staketia records.

Checkpoint: the slashed validator's recorded delegation is corrected, the redemption rate is
untouched, and the drift script shows zero over-recorded validators.

### Phase 3: the last redemption cycle

- The next day epoch submits the queued record and the retrying record; wait for the acks.
- Each record matures, is swept to the redemption ICA at a stride epoch, and is claimed with
  `ClaimUndelegatedTokens`.

Checkpoint:

- Zero user redemption records remain on either zone.
- Across several stride epochs there is no reinvestment, no new delegation, no rewards-claim
  ICA and no new epoch unbonding record.
- The stranded deposit sits unstaked in the delegation ICA.
- The redemption rate is unchanged.

### Phase 4: the admin drain

Starts once phase 3's undelegation acks have landed.

- Per zone: fully drain one small validator as the live test, then send
  `MsgUndelegateFromValidators` with an empty list for the rest. On the Hub this must split
  into more than one ICA batch.

Checkpoint: a success ack decrements the validator and zone delegations and burns nothing;
`TotalDelegations` ends at dust and no validator has a change in progress.

Failure injections, each followed by the cure the wind-down spec prescribes:

| Injection | Expected | Cure |
| --- | --- | --- |
| Slash a Hub validator after the refresh | Its batch fails with an error ack, no balance moves, the other batches succeed | Refresh that validator and resubmit it |
| Pause the relayer past a batch's ICA timeout | The ordered channel closes | `restore-interchain-account`, flags reset, resubmit |
| A batch executes on the host but its ack is lost to a channel close | Stride still records the full delegation on an emptied validator | `CalibrateDelegation`; the empty response corrects the record to zero |
| Submit in the last fifth of the day epoch, where the computed timeout is already past | The send fails and nothing is flagged (§9 says "usually just fails") | Resubmit after the next day epoch |
| Drain one validator with an `offset` | The offset stays delegated and recorded | Calibrate or drain the remainder |

### Phase 5: transfers to Osmosis

- `MsgTransferFromIca WITHDRAWAL` and `FEE` first, as the live test. The Hub's arrival on
  Osmosis proves the channel map and the denom the tokens land as; the Osmosis zone's proves
  the bank-send form.
- After the claims are done, and after running the §9 pre-transfer checklist by hand each
  time: `DELEGATION` for the full balance, `WITHDRAWAL` again for the auto-withdrawn rewards,
  and `REDEMPTION` for the dust the claims left.

Checkpoint: every balance arrives in the vault as the canonical denom, and all four ICAs on
both zones are at dust.

Failure injection: pause the hub↔osmosis relayer past the shortened timeout. The transfer
must refund to the ICA on the host, and a resubmission must succeed.

### Phase 6: staketia

- Once the multisig's unbonding matures, the operator sends the balance to the claim address
  through authz. A transfer with a memo, or to any other receiver, must be rejected by the
  grant.
- `MsgConfirmUnbondedTokenSweep` for each record; the hour epoch pays the redeemers on Stride.
- `MsgTransferStaketiaClaimBalance` with a small explicit amount, then with zero for the
  remainder. It lands in the Hub zone's delegation ICA as native ATOM and leaves with a
  `DELEGATION` transfer. If phase 5 already sent the Hub's delegation balance, a second
  `DELEGATION` transfer carries the staketia portion; the stATOM pools are funded only after
  it arrives.

Checkpoint: zero staketia redemption records, every unbonding record `CLAIMED`, the claim
address empty, and no key of the claim address signed anything.

### Phase 7: pools

- `coverage_check.py` gates the funding.
- The vault multisig creates three pools with `MsgCreateCosmWasmPool`, each with factors
  `1e18` for the stToken and `RedemptionRate × 1e18` for the native token and the alloyed
  asset: canonical stATOM, canonical stOSMO, and the Hub-route stATOM pool.
- Per pool: a small test deposit, the native-only `join_pool`, then `mark_corrupted_assets`
  on the native token, then `check_transmuter_pool.py`.

Checkpoint:

- A stToken → native swap through the poolmanager pays exactly the frozen rate, rounded down.
- A native → stToken swap is refused.
- The route pool holds exactly its channel's escrow × the rate, and the canonical pool the
  remainder.
- Unmark, join, re-mark completes an allocation after a test deposit.
- An outsider's `join_pool` with stTokens shows up in the coverage check as a remaining claim.

### Phase 8: the sweep off Stride

- Stop the fourth Stride validator, take a fresh export, and restart it.
- Build batches with `build_sweep_batches.py` and sign them with the sweep operator. The same
  tx from the admin multisig is rejected.
- Measure gas per batch at several batch sizes.

Checkpoint:

- stTokens and STRD land on Osmosis at the derived addresses in the canonical denom.
- ATOM vouchers land as native ATOM on the Hub at the derived addresses.
- A hand-built batch containing the module account and the interchain account emits a skip
  event for each and counts them in `num_skipped`; a builder-made batch skips nothing.
- An address whose transfer timed out is refunded on Stride and swept on resubmission.

### Phase 9: the halt

- Run the §9 halt checklist, including clearing every channel in both directions.
- Set a halt height on the Stride validators and let the chain stop.

Final invariant: with Stride dead, every stToken holder on Osmosis, canonical and
foreign-route, swaps their whole balance. Every swap must succeed, and every pool must end
with a non-negative native surplus.

## 6. Not covered

- The haqq close-and-restore sequence (§9c) and the haqq delta table.
- Stakedym claims on a halted zone.
- The LSM requeue, the trade-route deletion, the comdex flag and the wasm contract admin
  moves. Their handler steps run and must skip cleanly, but nothing is seeded for them.
- The stale `DelegationChangesInProgress` reset: not worth fabricating the state for.
- Celestia-specific host behaviour (its ICA host allow-list, its ibc-go version's authz
  semantics). The Hub stands in for Celestia; those stay as read-only mainnet checks.
- A sweep rejected for a voucher on a non-whitelisted channel. It needs a third Stride
  channel; the plan adds it only if that is cheap.
- Mainnet scale: one foreign-route pool rather than 28, two zones rather than eleven.
- Validator key destruction after the halt, and the mainnet values of every swapped table.

## 7. Output

- A rehearsal log in the style of `docs/wind-down/transmuter-test-log.md`: per phase, the
  commands run, the tx hashes, and each checkpoint's result.
- A findings list. Each finding is a code bug, a runbook error in the wind-down spec, or a
  script bug, and each is fixed on main in its own PR.

## 8. Open items for the plan

- **The v34 binary and staketia.** v34's staketia code looks up the stakeibc zone by the
  `CelestiaChainId` constant, so genesis config alone may not be enough on the old binary.
  Either the old binary is built from the v34 tag with that constant patched, or the Hub
  runs with the chain id `celestia`. Decide after reading the v34 code paths.
- **Which v34 tag.** `v34.0.0` and `v34.1.0` both exist; start from the one mainnet runs.
- **Transmuter on the test Osmosis.** The k8s Osmosis image is v28.0.0. Confirm it accepts
  mainnet's transmuter bytecode and that `cosmwasmpool` can whitelist the code id at genesis;
  bump the image if not. The code id will not be 996.
- **Hub image version.** `values.yaml` says gaia v22.1.0 and the Dockerfile says v25.1.0;
  settle which one runs, and confirm its authz `TransferAuthorization` enforces the empty
  `allowed_packet_data` rule.
- **Validator keys.** `keys.json` has five validator keys; the Hub needs eight.
- **Undelegate batch size.** Confirm the per-batch validator count for the Hub zone so the
  validator count is certain to exceed it.
- **The release gate (wind-down PR 6).** Find what it enforces at build or start time and
  whether the swapped constants trip it.
- **Seeding the retry record.** Getting a record into `UNBONDING_RETRY_QUEUE` needs a slash
  timed against the short epochs; if it cannot be made reliable before the upgrade, inject it
  right after instead.
- **Kube context.** Two clusters have an `integration` namespace; verify the context before
  starting or tearing anything down.
