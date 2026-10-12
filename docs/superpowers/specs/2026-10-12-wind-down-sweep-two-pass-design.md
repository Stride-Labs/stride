# Wind-down sweep: two passes, staked-STRD deferral, module-account re-check

Date: 2026-10-12. Builds on `2026-10-09-wind-down-sweep-runner-design.md`. Branch `wind-down-dashboard`.

## Why

The sweep runs Wed 2026-10-14 to Fri 2026-10-16 (pass 1), eleven days before the STRD the v35 upgrade undelegated
becomes liquid. Every unbonding entry the upgrade created matures at `2026-10-26T08:55:25Z` (read from mainnet on
10-12: 101 validators, all `BOND_STATUS_UNBONDING`, bonded pool 0, 10,410,527 STRD not bonded, validator tokens 0).
A holder with STRD still unbonding would be swept twice: its liquid balance in pass 1 and the matured STRD in pass 2.
The decision (user, 10-12): a holder with *any* STRD in unbonding is deferred out of pass 1, strictly, no threshold,
so each address is processed once. Pass 2 on Tue 2026-10-27 sweeps the deferred holders together with their STRD,
anyone new or refunded, and the keyless tier last.

The second concern is module accounts: the protocol is mid wind-down and module accounts still hold funds bound for
the Osmosis vault. The sweep never *sends to* a module account (each transfer goes to the holder's own bytes on the
destination chain) and a module account can never be a *sender*: the chain refuses it by the bank blocklist, by the
account-type rule (only `BaseAccount` and the four vesting types pass; a `ModuleAccount` or `InterchainAccount` is
refused by type), and the planner mirrors the same rules at plan time. Host-zone deposit addresses are 32-byte and fail
the first rule. This design adds one more layer at run time, re-read live, so the guard is not only at plan time.

## Decisions

1. **Deferral rule.** Strict: any unbonding entry, or any delegation the upgrade loop skipped, defers the holder.
   The `--test` plan never defers (the test address itself has STRD unbonding; the rehearsal must still run).
2. **Bulk read, not per address.** Staked STRD is read per validator, not per candidate: `/cosmos/staking/v1beta1/validators`
   (every status, all pages), then `/validators/{valoper}/unbonding_delegations` (all pages) and, only for a validator
   whose `tokens` is non-zero, `/validators/{valoper}/delegations`. About a hundred requests, under a minute, and the
   result covers every address so the ladder and the dashboard can use it below the floor too. The per-delegator
   `/delegations` route answers `Not Implemented` on v35, which rules out the per-address form anyway.
3. **Fail closed on an incomplete staking read.** The read reconciles against `/cosmos/staking/v1beta1/pool`:
   bonded + not bonded must equal the sum of every validator's `tokens` plus every unbonding entry's `balance`.
   A mismatch raises `StakingError` and `plan` stops (a missed entry would un-defer someone). The build verifies the
   identity holds on mainnet before the rule ships; if it does not hold exactly, the check becomes a warning line.
4. **No new tier.** Deferred holders are a list on the plan, not a tier, so `run` cannot sweep them by accident.
   After 10-26 the staking map is empty and pass 2's plan includes them with no flag or option.
5. **Run-time module re-check.** A new preflight row re-reads the module accounts (`/cosmos/auth/v1beta1/module_accounts`
   plus the `BLOCKED_MODULE_NAMES` hashes), the transfer escrows and the 20-byte wasm contracts live, and refuses any
   pending batch naming one. Eleven rows now, this one after "exclusions and protocol addresses".
6. **Ops plan.** `strd-unbonding-complete` moves to Mon 10-26 with the exact time. Pass 1 is a window on Wed 10-14
   titled for the three days; pass 2 a window on Tue 10-27. The 11-10 "The sweep" window goes away. Step ids
   `sweep-plan`, `sweep-submit`, `sweep-test` keep their pins; pass 2 adds `sweep-plan-final`, `sweep-submit-final`,
   `sweep-announce-final`.

## Design

### `sweep/holders.py`

- `Staked(ustrd: int, completes_at: str | None)`: STRD still in x/staking for one delegator: unbonding balances plus
  any delegation left bonded; `completes_at` is the latest unbonding completion as the chain prints it, `None` when a
  delegation is still bonded (no completion exists yet).
- `read_staked() -> dict[str, Staked]`: decision 2 and 3. Completion times are compared after `parse_chain_time()`
  (RFC 3339 with up to nine fractional digits, truncated to six for `datetime.fromisoformat`).
- `Deferred(holder: Holder, staked: Staked)`; `HolderSet.deferred: list[Deferred]` (above the floor, passed every skip
  rule, in the staking map) and `HolderSet.staked: dict[str, Staked]` (the whole map, for the ladder and the summary).
- `classify(..., staked=...)`: after the vesting step, when the address is in `staked` and this is not a test plan,
  the holder goes to `deferred` instead of `holders`. Below-floor holders are not deferred (the floor already rules
  them out); keyless holders can be deferred like any other.
- `read_holder_set()` calls `read_staked()` after `read_skip_inputs()`.
- `StakingError(Exception)` for the reconciliation failure, caught by `cli.main` like `DenomError`.

### `sweep/planner.py`

- `DeferredAddress(address, usd, transfers, staked_ustrd, completes_at)`; `Plan.deferred: list[DeferredAddress]`,
  `Plan.staked_addresses: int`, `Plan.staked_ustrd: int` (the whole map's size and total, every floor).
- `build_plan`: deferred sorted by USD descending; the ladder is built over `holders + below_floor` minus any address
  in `holder_set.staked`; the `locked` map also covers deferred vesting holders.
- `plan_to_dict` / `plan_from_dict`: `"deferred": [...]` (USD to cents, integers as strings),
  `"staked": {"addresses": "N", "ustrd": "X"}`; reading tolerates their absence like `locked`.

### `sweep/cli.py`

- `render_plan`: after the tiers, `deferred N addresses $X (STRD still unbonding until <date>; pass 2)` and
  `staked STRD: N delegators, X STRD`; the ladder label says deferred holders are not counted.
- `preflight`: new `_check_module_addresses(pending)` → `Check(name="module accounts, escrows and contracts", ...)`,
  inserted after `_check_exclusions`. It calls `holders.read_skip_inputs()`; detail names the counts read and any hit
  with its batch id.
- `main` catches `holders.StakingError`.

### Dashboard (`dashboard/sweep_tab.py`, `static/sweep.js`, `static/style.css`)

- `collect()` adds `"staked": {address: {"ustrd": str, "completes_at": str | null}}` from `holders.read_staked()`.
- `compose()`:
  - `totals.deferred = {addresses, usd}`: the plan's deferred addresses, USD from live balances at the plan's prices.
  - top-level `deferred = {addresses, usd, staked_addresses, staked_strd, staked_usd, completes_at, top: [...]}`
    where `staked_*` summarise the live map (all floors), `completes_at` its latest completion, `top` the 25 largest
    deferred by USD as `{address, usd, staked_strd}`. Without a live snapshot `staked_*` and `usd` are `null`.
  - `_known_addresses` includes the deferred, so they are never "unplanned above floor".
  - `_ladder` excludes every address in the live staking map.
  - by-token rows gain `deferred_usd` (live balance of the deferred addresses in that denom).
- `sweep.js`: a `Deferred` tile (sub-line "STRD unbonding until <date>: N delegators, $X"), a `deferred` segment
  in the progress bar and legend, a `Deferred USD` column after `Below floor USD`, and a collapsed `Deferred` panel
  with the top rows and a note that pass 2 sweeps them. `style.css`: `.sw-deferred` on the existing palette.
- READMEs: a paragraph each (sweep: the rule, the bulk read, the reconciliation; dashboard: the tile and panel).

### Ops plan (`dashboard/ops/plan.json`, `test_ops.py`)

- `strd-unbonding-complete` → 2026-10-26, window title "STRD · the upgrade's mass undelegation completes (14 days
  after the upgrade block: 08:55 UTC, 4:55am ET)". Text updated; command unchanged.
- 2026-10-14, new window "The sweep, pass 1 · Wed 10-14 to Fri 10-16": `sweep-plan` (floor, canaries, the deferred
  table and the staked line to read, 11 preflight rows), `sweep-submit` (the day loop: dry-run, canary, main, lower
  the floor; no keyless in pass 1), `sweep-relayers` (through pass 2), `sweep-announce` (pass 1 done; holders with
  STRD unbonding are swept in pass 2 on 10-27; below-floor and non-whitelisted vouchers move themselves).
- 2026-10-27, new window "The sweep, pass 2 · the deferred holders and their matured STRD": `sweep-plan-final`
  (`plan` must print `deferred 0` and `staked STRD: 0 delegators`; if not, unbonding has not completed: wait),
  `sweep-submit-final` (the same loop, ending with `run --tier keyless`), `sweep-rewards`, `sweep-community-pool`,
  `sweep-announce-final`.
- `sweep-test` expect text: eleven preflight rows in the new order.
- `test_ops.py` pins: `sweep-plan` on 2026-10-14 and `sweep-plan-final` on 2026-10-27 both contain
  `cli.py plan --floor-usd`; `sweep-submit-final` contains `run --tier keyless` and `sweep-submit` does not;
  `strd-unbonding-complete` on 2026-10-26; no step on 2026-11-10 has an id starting with `sweep-`.

### Error handling

- `read_staked` raises `ChainError` on a failed read (as every read does) and `StakingError` on the reconciliation
  mismatch; `plan` prints the message and exits 1.
- The new preflight row fails closed like the others: any hit, or a read error, stops the run before signing.

### Tests

- `test_holders.py`: a holder with an unbonding entry is deferred, keeps its balances and USD; the same holder in a
  `--test` plan is planned; a below-floor staker stays below floor; `read_staked` sums entries across validators,
  keeps the latest completion, reads delegations only for a validator with tokens, and raises on a pool mismatch;
  `parse_chain_time` on nine fractional digits.
- `test_planner.py`: deferred holders appear in `plan.deferred` sorted by USD and in no tier; the ladder excludes
  staked addresses at every rung; the plan dict round-trips `deferred` and `staked`; an old plan without them loads.
- `test_cli.py`: the new preflight row passes on a clean setup, fails naming the address and batch when a pending
  address is a module, escrow or contract address; eleven rows in order; `render_plan` prints the deferred line.
- `test_sweep_tab.py`: `totals.deferred`, the `deferred` object, `deferred_usd` per denom, the ladder excluding
  staked addresses, a deferred address not counted as unplanned.
- `test_ops.py`: the pins above.

## Build plan

Chunk 1, sweep package (files: `holders.py`, `planner.py`, `cli.py`, their tests, `sweep/README.md`). Produces
`Staked`, `read_staked`, `parse_chain_time`, `StakingError`, `Deferred`, `HolderSet.deferred/staked`,
`DeferredAddress`, `Plan.deferred/staked_addresses/staked_ustrd`, the preflight row. First step: a throwaway script
verifying the pool identity on mainnet (decision 3). Depends on: none.

Chunk 2, dashboard (files: `sweep_tab.py`, `static/sweep.js`, `static/style.css`, `test_sweep_tab.py`,
`dashboard/README.md`). Consumes chunk 1's `Plan.deferred`, `holders.read_staked`. Depends on: chunk 1.

Chunk 3, ops plan (files: `ops/plan.json`, `test_ops.py`). Edited by a script that round-trips the file with
`json.dumps(plan, indent=2, ensure_ascii=False) + "\n"`; command blocks keep notes on `#` lines, no trailing comments.
Depends on: none (the preflight row names come from this spec).

Out of scope: the Go CLI help text naming `build_sweep_batches.py`; a deferral threshold; caching the staking map.
