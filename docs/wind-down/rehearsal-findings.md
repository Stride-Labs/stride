# Wind-down rehearsal findings (2026-10-03)

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
mistake. Run 3 completed every phase.

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

## Open questions

On mainnet the celestia zone's `TotalDelegations` exceeds its validator sum by 156.87B utia, while
staketia's own `remaining_delegated_balance` is 69.59B. The 87B difference is part of celestia's
frozen redemption rate and is worth explaining before the upgrade.

The rehearsal did not measure the gas of `MsgUndelegateFromValidators`, and its state was too
small to extrapolate from: every drain fit in 600,000 gas. On mainnet two costs grow with state.
The queued-record guard reads every epoch unbonding record twice (796 records, about 4.3 MB as
JSON on 2026-10-03), and each ICA batch rewrites the whole host zone. A rough estimate for a full
cosmoshub-4 drain (90 funded validators, 18 batches of 5) is 25M to 35M gas against a 100M block
limit. Measure it against the mainnet export before the upgrade, and set `--gas` explicitly: even
a one-validator live test pays the guard's cost.

## Not covered

- The haqq sequence, stakedym claims, the LSM requeue, the trade route and contract admin moves:
  no state for them on the test network. Covered by the mainnet-export unit tests.
- Celestia-specific host behaviour: the Hub stood in for Celestia.
- The sweep's timeout refund: the relayer delivered the packet first. The same ICS-20 refund was
  proven on the Osmosis transfer.
- A drain submitted in the last fifth of the day epoch: the test tx landed just after the epoch
  rolled over.
- Liquid staking through an interchain account: the test channel never opened.
