# Wind-down sweep: warm caches, parallel reads, plan progress

Date: 2026-10-14. Branch: `wind-down-dashboard`. Builds on
`2026-10-09-wind-down-sweep-runner-design.md` and `2026-10-12-wind-down-sweep-two-pass-design.md`.

## Problem

A `plan` at a $100 floor spends its time in two sequential phases: the bulk balance read of nineteen
denoms (ustrd alone is 428 pages of 1000 holders, ~0.7 s each) and then one auth read plus one
unbonding read per candidate above the floor (~0.2–0.4 s each, one at a time). The second phase
answers questions that never change for an address (account type, module/contract, STRD unbonding;
a zero unbonding answer is final after v35) and should be answered once, ahead of time, for every
address we could ever plan. The first phase should fan out. The Sweep tab shows totals but not
"with this floor there are X holders and we are through Y of them".

Measured 2026-10-14 against Polkachu: `denom_owners_by_query` with `pagination.offset=300000` and
`count_total=true` works and takes 0.8 s, the same as the first page.

## Decisions

1. **Backoff in `chainio`.** `rest_get` retries on HTTP 429 and 5xx and on network errors with
   sleeps of 1, 2, 4, 8 s (five attempts), then raises `ChainError`. 404 stays `NotFound` at once.
   `urllib` is thread-safe; nothing else in `chainio` changes.

2. **Parallel balance reads, pinned to one height.** `read_balances` fetches every denom's first
   page with `pagination.count_total=true`, computes the page count, and fetches the remaining
   pages by `pagination.offset` from one `ThreadPoolExecutor` of `config.READ_WORKERS` (8) shared
   across denoms. Every request carries the `x-cosmos-block-height` header set to the height read
   at the start, so the pages form one snapshot. If the node refuses the pinned height (an error
   on the first pinned page), the read falls back to unpinned pages and prints one line saying so.
   Zero amounts are dropped and the address dict dedupes as today.

3. **Prefetch, shared by `plan` and `warm`.** New `holders.prefetch(addresses, accounts, staking,
   now, workers)` reads, in one thread pool, the auth account and the unbonding entries of every
   address whose cache entry is missing or not final (keyless-looking accounts and positive
   unbonding past its completion stay as they are: `lookup` re-reads those). It fills both caches
   and saves them every `config.PREFETCH_SAVE_EVERY` (200) addresses and at the end. The addresses
   handed to it are those that pass the skip rules that need no account (`skip_reason` with
   `account=None` returning `ACCOUNT_NOT_FOUND`) and are not excluded. `read_holder_set` calls
   `prefetch` for the candidates at or above the floor before `classify`, so `classify` is
   unchanged and its lookups are cache hits. Unbonding is read for every prefetched address, even
   one whose account turns out unsweepable; that costs one extra read on a few addresses and
   keeps the function simple.

4. **Merge-on-save caches.** `AccountCache.save` and `StakingCache.save` re-read the file, overlay
   the in-memory entries on top, and write atomically. Two processes (a `warm` and a `plan`) then
   never clobber each other; the later writer only adds. A `load` after a `save` by another process
   sees the union.

5. **`cli.py warm [--floor-usd 1] [--workers 8]`.** Reads balances (decision 2), prints the ladder
   counts, then prefetches (decision 3) every holder at or above the floor, printing progress every
   500 addresses and a final `RESULT: WARMED — N addresses read, M already cached`. Idempotent and
   resumable. Default floor $1 because that is the lowest ladder rung anyone would plan at; `--floor-usd 0`
   warms every holder. It never writes `plan.json` or batches and may run while a `plan` or `run`
   is in progress (decision 4 makes that safe). Documented in the README and the ops `sweep-plan`
   step as the thing to run after the first plan: "warm the caches in the background; later plans
   only read balances".

6. **Plan progress on the Sweep tab.** `compose` adds `progress`: `{run_id, floor_usd, test,
   tiers: [{name, batches, batches_confirmed, addresses, addresses_confirmed, usd, usd_confirmed}],
   batches, batches_confirmed, addresses, addresses_confirmed, usd, usd_confirmed}` from the plan
   and the ledger alone (a batch is confirmed when its latest ledger event is `confirmed`; its
   addresses and USD count then). The tab renders one line under the run line, e.g.
   `Run 2 · floor $1,000 · 3 of 12 batches · 214 of 1,140 holders · $412,000 of $1,900,000`,
   and a small per-tier table (canary, main, keyless) with the same columns. Hidden when there is
   no plan.

7. **A snapshot cannot judge a newer transfer.** `_address_states` and `_refunds` treat a confirmed
   transfer whose event `at` is later than `live_fetched_at` as swept, not refunded: the snapshot
   predates it. The refund check applies once a snapshot taken after the transfer exists. This
   removes the false "refunded" the tab shows for up to 30 minutes after every batch (seen on the
   test batch today). Flagged to the user as an addition to the ask.

Out of scope: pre-reading vesting spendable balances (few addresses); changing the sweep collector's
30-minute interval; any planner or chain-rule change.

## Build plan

### Chunk 1: chainio backoff and parallel balances
Files: `scripts/wind-down/sweep/chainio.py`, `holders.py` (`read_balances` only), `config.py`
(`READ_WORKERS = 8`, `REST_BACKOFF_SECONDS = (1, 2, 4, 8)`), `test_chainio.py`, `test_holders.py`.
Produces: `chainio.rest_get(path, params, height=None)` adding the pinned-height header when given;
`rest_get_all_pages` unchanged; `holders.read_balances(denoms, workers=config.READ_WORKERS)` returning
the same shape. Tests: backoff on 429 then success, give-up after five; first page + offset pages
assembled, dedupe, pinned header present, fallback path. Depends on: none.

### Chunk 2: prefetch, merge-on-save, `warm`
Files: `holders.py` (caches, `prefetch`, `read_holder_set`), `cli.py` (`cmd_warm`, parser),
`config.py` (`PREFETCH_SAVE_EVERY = 200`), `README.md`, `dashboard/ops/plan.json` (`sweep-plan`
command block and text mention `warm`), `dashboard/test_ops.py` if a pin needs it, `test_holders.py`,
`test_cli.py`. Produces: `holders.prefetch(...)`, merge-on-save, `warm`. Consumes nothing from chunk 1
(it calls `read_balances` by its existing signature). Tests: merge-on-save keeps the other writer's
entries; prefetch skips final entries, reads the rest, saves periodically; `read_holder_set` prefetches
only candidates at or above the floor; `warm` prints the RESULT line and leaves `plan.json` alone.
Depends on: none.

### Chunk 3: Sweep tab progress and the snapshot rule
Files: `scripts/wind-down/dashboard/sweep_tab.py`, `static/sweep.js`, `static/style.css`,
`test_sweep_tab.py`, `dashboard/README.md`. Produces: `progress` in `compose`; the snapshot rule in
`_address_states` / `_refunds`. Tests: progress counts per tier and overall with a mixed ledger; a
confirmed transfer newer than `live_fetched_at` is swept even when the snapshot still shows the balance;
one older than it is refunded. Depends on: none.
