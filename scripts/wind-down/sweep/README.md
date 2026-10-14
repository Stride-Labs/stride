# Sweep runner

The holder sweep of the v35 wind-down (`MsgSweepTokensOffStride`, design in
`docs/superpowers/specs/2026-10-09-wind-down-sweep-runner-design.md`). Stdlib only.

    python3 scripts/wind-down/sweep/cli.py plan --floor-usd 100 --canary 3   # live state → state/plan.json + batch files
    python3 scripts/wind-down/sweep/cli.py run                               # preflight, then sign and submit each batch
    python3 scripts/wind-down/sweep/cli.py status                            # no network
    python3 scripts/wind-down/sweep/cli.py resolve                           # finish a submission the poll gave up on

The operator key is `stride-sweeper` in the **test** keyring. `config.py` holds the denoms, the rough prices, gas
settings and limits; `exclusions.json` the accounts the team moves by hand (a section, a reason, a label per address).
`exclusions.json` is enforced at `plan` and again at `run`.

A `plan` reads every holder of the nineteen sweep denoms and takes about six minutes against Polkachu (measured 2026-10-09: 5m45s).

## The test (after the 10-12 upgrade, before the sweep)

    python3 scripts/wind-down/sweep/cli.py plan --test     # one batch: stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg
    python3 scripts/wind-down/sweep/cli.py run --dry-run   # gas estimates only
    python3 scripts/wind-down/sweep/cli.py run

The address already holds what the test needs: STRD and stTIA (Stride-native, to Osmosis over channel-5) and a TIA
voucher (back to Celestia over channel-162), plus dust of most other stTokens; every sweepable denom it holds moves.
Then check the same bytes under `osmo1nwyvkxm89yg8e3fyxgruyct4zp90mg4n5x5jak` hold the STRD and stTokens and under
`celestia1nwyvkxm89yg8e3fyxgruyct4zp90mg4ndhkj3f` the TIA, and that the Sweep tab shows the batch confirmed. The test
batch also calibrates gas per transfer for the real run. `status` reprints the plan's tables any time. The exact command list, with the balance checks, is the
`sweep-test` step on the Ops tab.

## Sweep day

1. `plan --floor-usd <announced floor> --canary 3`, read the tables (excluded, skipped, ladder), `run`.
   The canary tier (the three smallest holders above the floor) goes first, then everyone by value descending.
   A plain `run` walks the canary and main tiers only and prints how many keyless batches it left behind.
2. Lower the floor: `plan --floor-usd 25`, `run`; repeat. Swept holders have no balance, so a re-plan never lists them;
   a transfer that timed out (24 h) refunds the holder, who reappears in the next plan.
3. The `keyless` tier (never-signed accounts) is last in every plan and a plain `run` skips it, so those owners get the
   most time. Run it once, at the very end, with `run --tier keyless` (it runs only that tier).

`warm [--floor-usd 1] [--workers 8]` answers the per-address questions (auth account, STRD unbonding) for every holder
at or above the floor and fills the caches, so later plans only read balances. Run it once in the background after the
first plan; it is idempotent and resumable, writes no plan or batch, and is safe alongside `plan` and `run` because
cache saves merge with the file instead of overwriting it.

`plan` records the operator's account sequence in `plan.json`. Preflight and, again right before each broadcast, `run`
require the chain's sequence to equal that plus the run's `submitted` lines: a `strided` crash or an unreadable broadcast
reply can leave a tx in the mempool with no ledger line, and the sequence is how the next run finds out. On a mismatch
find the stray tx (`strided q tx`, the operator's history), record or discard it, then `plan` again.

`run` stops on any `sweep_skipped` event (the planner and the chain disagree: inspect, fix, re-plan), on a failed tx,
on a batch over the block gas limit, and when a tx is not found within three minutes (`resolve`). `--dry-run` prints
every command and gas estimate without broadcasting; `--yes` skips the per-batch prompt (`a` at the prompt does the
same for the rest of the run); `--batches N` stops after N.

## Two passes: STRD still unbonding

The v35 upgrade undelegated everything; the STRD becomes liquid at `2026-10-26T08:55:25Z`. A holder with any STRD still
unbonding would be swept twice (its liquid balance now, the matured STRD later), so `plan` defers it: strictly, any
unbonding entry, no threshold. Deferred holders sit in `plan.deferred` (and `render_plan` prints their count, USD, the
latest completion and a table of the first 20: address, USD, staked STRD, completion), in no tier and no batch, so `run` cannot sweep them. They still count in the ladder.

- **Per-delegator read.** Each candidate above the floor that passes every skip rule gets one read of
  `/cosmos/staking/v1beta1/delegators/<address>/unbonding_delegations` (all pages): the sum of the entries' `balance` and the
  latest `completion_time`. The per-validator index misses removed validators, so it is not used. Delegations are not read
  (nothing is bonded after v35). `plan` also prints the chain-wide figure from `/cosmos/staking/v1beta1/pool`.
- **Cache.** `state/staking.json` (`{address: {"ustrd": "N", "completes_at": "..." | null}}`, sorted) keeps every answer.
  A zero answer is final; a positive one is served until its `completes_at`, after which the address is re-read. Re-plans
  at a lower floor read only the new addresses.
- **`--test` never defers**: the test address itself has STRD unbonding.
- **Pass 2** (after 2026-10-26 08:55 UTC) needs no flag: the staking map is empty, so the same `plan` includes the
  deferred holders and their matured STRD.

## Preflight rows

`run` prints eleven rows and stops on any failure: chain id, binary version, operator key, plan age, batch files,
unresolved submissions, operator sequence, exclusions and protocol addresses, module accounts / escrows / contracts,
channels open, operator fee balance. The module row re-reads the module accounts (plus the blocked module hashes), the
transfer escrows and the 20-byte wasm contracts live and fails if any pending batch names one; it fails closed on a
read error.

State files under `state/` are committed like `dashboard/ops/status.json`: `plan.json` and the batch files are
overwritten by every `plan`; `ledger.jsonl` is append-only.
