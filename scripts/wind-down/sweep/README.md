# Sweep runner

The holder sweep of the v35 wind-down (`MsgSweepTokensOffStride`, design in
`docs/superpowers/specs/2026-10-09-wind-down-sweep-runner-design.md`). Stdlib only.

    python3 scripts/wind-down/sweep/cli.py plan --floor-usd 100 --canary 3   # live state → state/plan.json + batch files
    python3 scripts/wind-down/sweep/cli.py run                               # preflight, then sign and submit each batch
    python3 scripts/wind-down/sweep/cli.py status                            # no network
    python3 scripts/wind-down/sweep/cli.py resolve                           # finish a submission the poll gave up on

The operator key is `stride-sweeper` in the **test** keyring. `config.py` holds the denoms, the rough prices, gas
settings and limits; `exclusions.json` the accounts the team moves by hand (a section, a reason, a label per address).
Both are enforced at `plan` and again at `run`.

## The test (after the 10-12 upgrade, before the sweep)

    python3 scripts/wind-down/sweep/cli.py plan --test     # one batch: stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg
    python3 scripts/wind-down/sweep/cli.py run

Then check the same bytes under `osmo1…` hold the stToken and STRD and under `cosmos1…` the ATOM, and that the Sweep
tab shows the batch confirmed. The test batch also calibrates gas per transfer for the real run.

## Sweep day

1. `plan --floor-usd <announced floor> --canary 3`, read the tables (excluded, skipped, ladder), `run`.
   The canary tier (the three smallest holders above the floor) goes first, then everyone by value descending.
2. Lower the floor: `plan --floor-usd 25`, `run`; repeat. Swept holders have no balance, so a re-plan never lists them;
   a transfer that timed out (24 h) refunds the holder, who reappears in the next plan.
3. The `keyless` tier (never-signed accounts) is last in every plan; run it at the end with `run --tier keyless`.

`run` stops on any `sweep_skipped` event (the planner and the chain disagree: inspect, fix, re-plan), on a failed tx,
on a batch over the block gas limit, and when a tx is not found within three minutes (`resolve`). `--dry-run` prints
every command and gas estimate without broadcasting; `--yes` skips the per-batch prompt (`a` at the prompt does the
same for the rest of the run); `--batches N` stops after N.

State files under `state/` are committed like `dashboard/ops/status.json`: `plan.json` and the batch files are
overwritten by every `plan`; `ledger.jsonl` is append-only.
