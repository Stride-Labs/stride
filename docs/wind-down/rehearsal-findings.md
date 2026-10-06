# Wind-down rehearsal findings (2026-10-03, run 4 added 2026-10-05)

The v35 wind-down was rehearsed end to end on the k8s integration network from the throwaway
branch `wind-down-rehearsal`, which is never merged. The command-by-command log, with every tx
hash, is `docs/wind-down/rehearsal-log.md` on that branch. This file records what the rehearsal
found and what, if anything, each finding means for mainnet.

## Setup

| Test chain | Stands in for | Notes |
| --- | --- | --- |
| stride-test-1, 4 validators | mainnet Stride | Started on v34.1.0 and upgraded to v35 by governance |
| cosmoshub-test-1, 8 validators | cosmoshub-4 and celestia | A drain splits into 3 ICA batches; hosts the staketia multisig |
| osmosis-test-1, 3 validators | osmosis-1, as host zone and destination | Runs mainnet's transmuter bytecode (code id 996) |

Only addresses, channel ids and `WindDownTransferTimeout` differed from the mainnet binary. The
protocol admin and the Osmosis vault were one 2-of-3 multisig signing in amino-json.

Three runs were needed. Run 1 ended after the upgrade when the test network stalled. Run 2 ended
in the transfer phase when the test Hub ran out of memory and was then wiped by an operator
mistake. Run 3 completed every phase. Run 4 (2026-10-05) repeated everything on main after #1544
and #1548 merged, with the authority hand-off, the STRD mass undelegation and the community pool
transfer in scope; see "Run 4" below.

## What passed

Every step of the ops window (spec §9) passed in run 3:

- Upgrade with redemptions open in every status; handler clean; absent-chain steps skipped.
- Redemption rate identical from the upgrade to the halt.
- Day-0 refresh by the multisig; drain refused while a record was queued or retrying.
- Last redemption cycle: every record unbonded, swept and claimed.
- Admin drain of both zones, single validator first, then the whole zone.
- `MsgTransferFromIca` for all four ICA types on both zones, both transfer forms.
- Staketia exit through `MsgTransferStaketiaClaimBalance` to the vault.
- Three transmuter pools created, funded and marked one-way; coverage check passed.
- `MsgSweepTokensOffStride` from a builder batch that skipped nothing on chain.
- Halt, followed by every holder swapping their full balance with the pool left in surplus.

Recovery procedures exercised: a failed batch after a slash, an ICA timeout and
`restore-interchain-account`, the lost-ack path (`close-delegation-channel`, restore,
`calibrate-delegation`), a timed-out transfer refund, and a governance client recovery.

## Findings that cannot occur on mainnet

These looked serious during the rehearsal. Each came from a mistake in the rehearsal's setup, so
no chain code changes.

### 1. A drain's last ack was rejected and the delegation channel wedged

The undelegate callback refuses a change larger than `HostZone.TotalDelegations`. If the zone
total is below the sum of `Validator.Delegation`, the last batch of a full drain hits that check
while processing a success ack. The error reverts the ack, and on an ordered channel no later
ack can be delivered.

- **Rehearsal cause:** the staketia stake was booked incorrectly, so the total started below the
  validator sum.
- **Mainnet:** every stakeibc path changes both numbers together, so they are equal on every
  zone. Celestia's total is higher by design, because it includes the staketia stake: 156.87B utia
  of headroom on 2026-10-03, against 14.3B of pending staketia confirmations, and no new staketia
  records can be created after v35.
- **Recovery if it ever happened:** `close-delegation-channel`, `restore-interchain-account`, then
  `calibrate-delegation` on each validator the lost batch emptied.

### 2. `TotalDelegations` went negative after calibration

The calibration callback subtracts a validator's over-recorded amount from the zone total with no
lower bound. It went negative only because the total was already below the validator sum, the
same setup mistake as finding 1.

### 3. A queued redemption retried forever and blocked the drain

The record-driven unbonding path (not the admin drain) computes each validator's spare capacity
and requires the capacities to cover the whole record. When fully draining a validator with a
stored rate below one, `applySharesRoundingSafety` holds back a few units. If every other
validator is already at full capacity, nothing absorbs those units, the path returns "unable to
unbond full amount", and it does so again every day epoch. The drain refuses a zone with a queued
record, so the zone is stuck.

- **Rehearsal cause:** the seed built an exact fit: two perfectly balanced validators plus a
  slashed zero-weight validator, and a redemption sized past the zero-weight stake.
- **Mainnet:** zones with slashed zero-weight validators have large spare capacity (cosmoshub-4
  35.7B uatom beyond its queued records on 2026-10-03). osmosis-1 and ssc-1 have almost none but
  have no such validators.
- **Status:** a latent bug in code that is already live (#1510). Left unpatched.
- **Escape hatch:** `change-validator-weight` on the zero-weight validator.

## Findings that apply to mainnet

| Finding | Fix |
| --- | --- |
| `MsgTransferFromIca` is rejected when the ICA's channel is closed. An old fee-ICA timeout had closed one. Nothing is sent. | Runbook check before each transfer (spec §9) |
| `check_transmuter_pool.py` always fails once the native token is marked corrupted | `--funded` mode |
| `strided export` writes nothing to stdout on SDK 0.54 | Runbook note: use `--output-document` |

## Test harness

Fixed in `integration-tests/`: osmosisd prints its node id and validator pubkey on stderr, gaia's
`create-validator` needs fees, and validators need more than 700M of memory.

Not fixed: after the upgrade-height halt, validators whose daemon cosmovisor restarted in-process
came back with no peers and consensus stalled until they were restarted by hand. Validator state
is an `emptyDir`, so deleting a pod wipes its chain.

## Open question

On mainnet the celestia zone's `TotalDelegations` exceeds its validator sum by 156.87B utia, while
staketia's own `remaining_delegated_balance` is 69.59B. The 87B difference is part of celestia's
frozen redemption rate and is worth explaining before the upgrade.

## Drain gas on mainnet-sized state

The rehearsal's state was too small to say anything about gas: every drain there fit in 600,000.
`TestDrainGasFromMainnetExport` (`app/upgrades/v35/drain_gas_test.go`) measures
`MsgUndelegateFromValidators` against the mainnet export instead, with every host zone and all
793 epoch unbonding records at their real size. The block gas limit on stride-1 is 100,000,000.

| Zone | Funded validators | ICA batches | One-validator drain | Full drain | One ack |
| --- | --- | --- | --- | --- | --- |
| cosmoshub-4 | 74 | 15 | 9.59M | 20.03M | 1.48M |
| osmosis-1 | 23 | 5 | 9.43M | 11.83M | 1.17M |
| ssc-1 | 16 | 16 | 8.96M | 11.41M | 0.27M |
| laozi-mainnet | 32 | 7 | 9.01M | 10.34M | 0.37M |
| celestia | 92 | 3 | 9.22M | 10.22M | 0.77M |
| sommelier-3 | 19 | 4 | 9.01M | 9.67M | 0.37M |
| injective-1 | 40 | 2 | 9.09M | 9.46M | 0.50M |
| haqq_11235-1 | 41 | 2 | 9.08M | 9.45M | 0.48M |
| phoenix-1 | 41 | 2 | 9.06M | 9.40M | 0.46M |
| juno-1 | 24 | 1 | 9.08M | 9.13M | 0.49M |
| dydx-mainnet-1 | 25 | 1 | 9.03M | 9.09M | 0.38M |

- **About 9M is a fixed cost.** The queued-record guard reads every epoch unbonding record, so
  even a one-validator live test costs 9M. Set `--gas` explicitly on the multisig tx.
- **Each ICA batch adds about 0.7M on the Hub**, because it rewrites the whole host zone. The
  Hub's full drain is the largest at 20% of the block limit.
- **Acks are separate relayer txs.** A Hub ack costs about 1.5M, so a relayer that packs all 15
  into one tx needs about 22M of gas for it.

The measurement excludes the ante handler (signature checks and tx size), which is small next
to these numbers. The test fails if a full drain or an ack exceeds half the block limit.

## Run 4: authority hand-off, STRD undelegation, community pool (2026-10-05)

Rehearsed from the same branch rebased onto main at `5b7fbe92b`. Seeded on v34: four stakers with 12
delegations across val1-3 (4,283 STRD), a fifth staking validator jailed below its minimum
self-delegation and already unbonded at the upgrade with delegations still on it, one pair holding
the 7-entry cap, a redelegation and a fresh unbonding entry in flight, a proposal in its deposit
period and one in its voting period, and a community pool holding stATOM, stOSMO, two vouchers and
STRD. State was read at the block before and the upgrade block through height-pinned REST.

Passed, handler steps 10-19:

- Consensus authority nil before, the multisig after; POA admin moved to the multisig; gov deposits
  1e18 / 2e18 ustrd with every other gov param unchanged; staking `max_entries` 7 -> 100; the ICA
  host allow-list lost the four staking and the two stakeibc messages and kept the rest.
- Both pending proposals rejected with "Governance closed by v35 wind-down", deposits and votes
  removed, depositors refunded in the upgrade block.
- "undelegated 12 delegations totaling 4283000000 ustrd, skipped 0": no delegation left, bonded pool
  0, val1-4 jailed and unbonding at zero tokens, the already-unbonded val5 removed in the upgrade
  block. The 7-entry pair got its 8th entry with the first 7 untouched; every new entry equals the
  delegation and completes exactly one unbonding period after the upgrade block; the pre-existing
  entry and the redelegation untouched. After the unbonding period: every staker's STRD liquid
  again, the redelegation completed, all staking validators removed with no panic on any node,
  blocks still produced, community pool refilling from fees with no bonded validators.
- Community pool: every denom's whole units reached the multisig, STRD at least the pool balance,
  0.72 ustrd of dust left.
- Ante: delegate, redelegate, create-validator and cancel-unbond refused with "is disabled: the chain
  is winding down". Stride has no authz module, so there is no MsgExec path to close.
- A proposal with the old full deposit never reaches voting. `MsgSoftwareUpgrade` from an ordinary
  key refused; the multisig scheduled an upgrade directly, cancelled it, and recovered an expired
  light client with `MsgRecoverClient` as the chain authority (no gov).
- An interchain account on Stride controlled from the Hub: `MsgLiquidStake` and `MsgDelegate`
  packets both left it untouched (this route did not open in run 3).
- Ops step "sweep-community-pool": after the holder sweep (which also swept the stakers' formerly
  staked STRD), the multisig spent the refilled pool to itself (an ordinary key is refused) and
  transferred stTokens, STRD and the OSMO voucher to the Osmosis vault and the ATOM voucher back to
  the Hub, keeping a fee reserve, before the halt checklist.
- Run-3 gaps closed: the ICA timeout, transfer refund and sweep refund injections all triggered
  (the relayer is frozen in place instead of scaled down), the dead-window drain was refused with
  the timeout reason, the pre-transfer ICA channel check ran, and the pool gate passes with
  `--funded`.
- Halt-height: proven on a fresh v35 Stride chain, all four validators stopping at the configured
  height ("halt per configuration height 722"). Runs 3 and 4 had never actually restarted the nodes
  with the new `app.toml`, so the earlier "halt" was a peer loss, not the mechanism.

Findings:

| Finding | Meaning for mainnet |
| --- | --- |
| After the undelegation, the staking `redelegations` query for a delegator whose destination validator has zero shares fails with "division by zero" (the response computes balance = shares x tokens/shares) until the entry matures. The record is intact and completes on time. | Query only. Dashboards or users reading `/cosmos/staking/v1beta1/delegators/{addr}/redelegations` during the 14 days after v35 may get an error instead of the entries. |
| Undelegating from an already-unbonded validator still books an unbonding entry with the full unbonding time (`Undelegate` does not special-case it); only the validator is removed on the spot. | Delegators to mainnet's 125 unbonded validators wait the 14 days like everyone else. The spec's wording ("returned immediately") was wrong; the handler is right. |
| Halt-height needs a node restart to take effect, and on this network restarted nodes came back peerless twice (the same cause as the post-upgrade stall). | Runbook: after setting `halt-height`, confirm every validator restarted with peers and is advancing before the height. |

## Not covered

- The haqq sequence, stakedym claims, the LSM requeue, the trade route and contract admin moves:
  no state for them on the test network. Covered by the mainnet-export unit tests.
- Celestia-specific host behaviour: the Hub stood in for Celestia.
- (covered in run 4) A drain submitted in the last fifth of the day epoch; liquid staking and
  delegating through an interchain account; the sweep's timeout refund.
