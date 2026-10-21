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
2. **Per-delegator read, cached.** Each candidate above the floor that passes every skip rule gets one read of
   `/cosmos/staking/v1beta1/delegators/{address}/unbonding_delegations` (all pages): the sum of every entry's
   `balance` and the latest `completion_time`. The per-validator index was measured on 10-12 and rejected: 16
   validators removed from the set still carry entries (1.4M of the 10.4M STRD not bonded; 31 of 301 sampled
   delegators had entries the index missed), so a bulk read would silently un-defer people. Delegations are not read:
   every validator's `tokens` is 0, nothing is bonded, and the per-delegator `/delegations` route answers
   `Not Implemented` on v35.
3. **The cache makes re-plans cheap.** A read costs about 0.2 s, so a plan at a low floor would add minutes.
   `StakingCache` (`state/staking.json`, the shape of `AccountCache`) keeps every answer: a zero answer is final (after
   v35 nothing can create an unbonding entry: `MsgDelegate`, `MsgBeginRedelegate`, `MsgCreateValidator` and
   `MsgCancelUnbondingDelegation` are refused by the ante decorator and dropped from the ICA allow-list, and no
   delegation is left to undelegate); a positive answer is final until its `completes_at`, after which the address is
   re-read. The whole-chain figure `plan` prints comes from `/cosmos/staking/v1beta1/pool` (`not_bonded_tokens`), one
   read, exact.
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

- `Staked(ustrd: int, completes_at: str | None)`: STRD still unbonding for one delegator: the sum of its entries'
  `balance` and the latest `completion_time` as the chain prints it; `Staked(ustrd=0, completes_at=None)` when none.
- `read_staked(address) -> Staked`: decision 2. Completion times are compared after `parse_chain_time()` (RFC 3339
  with up to nine fractional digits, truncated to six for `datetime.fromisoformat`, UTC).
- `read_staking_pool() -> int`: `not_bonded_tokens` plus `bonded_tokens` from `/cosmos/staking/v1beta1/pool`.
- `StakingCache` (`load`, `save`, `lookup(address, now)`): decision 3; the file is `config.STAKING_CACHE_PATH`
  (`state/staking.json`), written like `accounts.json` (tmp file, replace, sorted, integers as strings).
- `Deferred(holder: Holder, staked: Staked)`; `HolderSet.deferred: list[Deferred]` (above the floor, passed every skip
  rule, STRD unbonding), sorted by USD descending like `holders`.
- `classify(..., staking: StakingCache, now: datetime)`: after the vesting step, when this is not a test plan and
  the cache's answer has `ustrd > 0`, the holder goes to `deferred` instead of `holders`. Below-floor holders are not
  read (the floor already rules them out); keyless holders can be deferred like any other.
- `read_holder_set(..., staking)` passes the cache and `chainio.now()` through; `cmd_plan` loads and saves it beside
  the account cache.

### `sweep/planner.py`

- `DeferredAddress(address, usd, transfers, staked_ustrd, completes_at)`; `Plan.deferred: list[DeferredAddress]`,
  `Plan.staked_ustrd: int` (the chain-wide pool figure, passed into `build_plan` by `cmd_plan`).
- `build_plan`: deferred in `holder_set` order; the ladder is built over `holders + below_floor` as today (deferred
  holders are not yet swept, so they count; the label says so); the `locked` map also covers deferred vesting holders.
- `plan_to_dict` / `plan_from_dict`: `"deferred": [...]` (USD to cents, integers as strings),
  `"staked_ustrd": "X"`; reading tolerates their absence like `locked`.

### `sweep/cli.py`

- `render_plan`: after the tiers, `deferred N addresses $X (STRD still unbonding until <latest completes_at>; swept
  in pass 2)` and `STRD still unbonding chain-wide: X STRD`; the ladder label says deferred holders are counted.
- `preflight`: new `_check_module_addresses(pending)` → `Check(name="module accounts, escrows and contracts", ...)`,
  inserted after `_check_exclusions`. It calls `holders.read_skip_inputs()`; detail names the counts read and any hit
  with its batch id.

### Dashboard (`dashboard/sweep_tab.py`, `static/sweep.js`, `static/style.css`)

- `collect()` adds `"staked_ustrd": str` from `holders.read_staking_pool()`.
- `compose()`:
  - `totals.deferred = {addresses, usd}`: the plan's deferred addresses, USD from live balances at the plan's prices.
  - top-level `deferred = {addresses, usd, staked_strd, staked_usd, completes_at, top: [...]}` where `staked_strd`
    is the live pool figure (`null` without a snapshot), `staked_usd` prices it at the plan's STRD price,
    `completes_at` the latest completion among the deferred, `top` the 25 largest deferred by plan USD as
    `{address, usd, staked_strd, completes_at}`. Without a live snapshot `usd` is the plan's USD.
  - `_known_addresses` includes the deferred, so they are never "unplanned above floor".
  - `_ladder` is unchanged (deferred holders are still to sweep; its label says they are counted).
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

- `read_staked` raises `ChainError` on a failed read (as every read does); `plan` prints the message and exits 1.
- The new preflight row fails closed like the others: any hit, or a read error, stops the run before signing.

### Tests

- `test_holders.py`: a holder with an unbonding entry is deferred, keeps its balances, USD and locked part; the same
  holder in a `--test` plan is planned; a below-floor staker stays below floor and is not read; `read_staked` sums
  entries across validators and keeps the latest completion; `StakingCache` returns a zero answer without a read,
  re-reads a positive answer once `now` passes `completes_at`, round-trips through its file; `parse_chain_time` on
  nine fractional digits.
- `test_planner.py`: deferred holders appear in `plan.deferred` and in no tier and no batch; the plan dict
  round-trips `deferred` and `staked_ustrd`; an old plan without them loads.
- `test_cli.py`: the new preflight row passes on a clean setup, fails naming the address and batch when a pending
  address is a module, escrow or contract address; eleven rows in order; `render_plan` prints the deferred line;
  `plan` saves the staking cache.
- `test_sweep_tab.py`: `totals.deferred`, the `deferred` object with and without a live snapshot, `deferred_usd` per
  denom, a deferred address not counted as unplanned, the progress parts still summing to the whole.
- `test_ops.py`: the pins above.

## Build plan

Chunk 1, sweep package (files: `config.py`, `holders.py`, `planner.py`, `cli.py`, their tests, `sweep/README.md`).
Produces `config.STAKING_CACHE_PATH`, `Staked`, `read_staked`, `read_staking_pool`, `parse_chain_time`,
`StakingCache`, `Deferred`, `HolderSet.deferred`, `DeferredAddress`, `Plan.deferred`, `Plan.staked_ustrd`, the
preflight row. Depends on: none.

Chunk 2, dashboard (files: `sweep_tab.py`, `static/sweep.js`, `static/style.css`, `test_sweep_tab.py`,
`dashboard/README.md`). Consumes chunk 1's `Plan.deferred`, `Plan.staked_ustrd`, `holders.read_staking_pool`.
Depends on: chunk 1.

Chunk 3, ops plan (files: `ops/plan.json`, `test_ops.py`). Edited by a script that round-trips the file with
`json.dumps(plan, indent=2, ensure_ascii=False) + "\n"`; command blocks keep notes on `#` lines, no trailing comments.
Depends on: none (the preflight row names come from this spec).

Out of scope: the Go CLI help text naming `build_sweep_batches.py`; a deferral threshold; caching the staking map.
