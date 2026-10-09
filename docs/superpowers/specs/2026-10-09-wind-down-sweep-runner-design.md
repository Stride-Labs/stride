# Wind-down sweep runner and dashboard tab

Design for the last ops tool of the v35 wind-down: the holder sweep (`MsgSweepTokensOffStride`, spec
§7 of `2026-09-18-protocol-wind-down-design.md`). It replaces `scripts/wind-down/build_sweep_batches.py`
(export in, batch files out, everything after that by hand) with a package that plans from live chain
state, signs and submits each batch itself, keeps a ledger of what happened, and feeds a Sweep tab on
the dashboard.

## Decisions

Settled in the brainstorm on 2026-10-09:

- The sweep operator key is `stride-sweeper` in the local **test** keyring (never `os`: it prompts for
  the keychain password on every call). It resolves to `config.SWEEP_OPERATOR`
  (`stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9`). The runner signs and broadcasts itself.
- Holder state is read **live over REST**, not from an export. "Lower the floor and run again" is one
  command, and the `--as-of` footgun goes away.
- Order: a **test** sweep of `stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg` (an address we control)
  first, between the 10-12 upgrade and the 11-10 sweep; then on sweep day a **canary** of a few small
  real holders; then everyone above the floor **by value descending**; then re-plan at a lower floor
  and repeat, finishing within a couple of days.
- **Keyless** accounts (no pubkey, sequence 0: possibly hashed autopilot or forwarding addresses) are
  swept, but in a **final tier** so their owners get the most time to move first. The two unknown
  2-of-3 multisigs sweep in normal order: a multisig derives the same address on Osmosis.
- Exclusions move out of the code into a **reviewed file with a reason per address**. It holds only
  accounts the team moves by hand: the F5 multisig and the relayer keys. The five treasury accounts
  are swept (decided 2026-10-02).
- Prices are a **hard-coded rough proxy**: one USD figure per native token; an stToken is worth its
  native price times the zone's redemption rate (live, frozen after v35).
- The dashboard's **remaining** is live: on refresh it pulls every holder of every swept denom in
  bulk, so it also sees a transfer that timed out and refunded. The refresh is on demand (button, or
  a stale snapshot when the tab opens), never on a short interval. Batch status comes from the ledger
  on every request.
- The dashboard stays read-only. The runner is a CLI.

## Layout

A new stdlib-only package `scripts/wind-down/sweep/`:

| File | Role |
| --- | --- |
| `config.py` | Denoms, prices, operator key and address, node, chain id, gas settings, batch limits, state paths. |
| `exclusions.json` | The reviewed exclusion file. |
| `holders.py` | Live reads and holder classification (the on-chain skip rules). |
| `planner.py` | Tiers, batch packing, `plan.json` and the batch files. |
| `ledger.py` | The append-only `ledger.jsonl`: write, read, derive each batch's state. |
| `sweep.py` | The CLI: `plan`, `run`, `status`, `resolve`. |
| `chainio.py` | REST and `strided` subprocess wrappers (the seams the tests stub). |
| `state/` | `plan.json`, `batch-NNN.txt`, `ledger.jsonl`, `accounts.json`. Committed like `ops/status.json`. |
| `test_*.py` | Unit tests, no network, `python3 -m unittest discover -s scripts/wind-down/sweep`. |

`build_sweep_batches.py` and `test_build_sweep_batches.py` are deleted; `bech32_ref.py` stays (the
sweep package and `coverage_check.py` use it). The dashboard gains `dashboard/sweep_tab.py`,
`dashboard/static/sweep.js`, a tab in `index.html`, a README section and `dashboard/test_sweep_tab.py`.
The modules are `cli.py` and `sweep_tab.py`, not `sweep.py`: a module named like its package shadows it.
The dashboard imports the sweep package by inserting `scripts/wind-down/sweep` on `sys.path` in
`server.py` (the sweep package never imports the dashboard).

Python conventions per `~/.agents/AGENTS.md`: module imports, typed signatures, dataclasses for
multi-field results, named parameters, guard clauses, `Decimal` for USD. Every integer in a dashboard
payload is a string (the page uses BigInt), as on the other tabs.

## config.py

```python
SWEEP_OPERATOR_KEY = "stride-sweeper"
KEYRING_BACKEND = "test"
SWEEP_OPERATOR = "stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9"
TEST_ADDRESS = "stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg"
CHAIN_ID = "stride-1"; REST = "https://stride-strd-api.polkachu.com"; RPC = "https://stride-strd-rpc.polkachu.com:443"
STATE_DIR = <package dir>/state

# Native tokens: symbol -> (base denom on its chain, decimals, rough USD). Edit freely; a proxy, not a market feed.
NATIVE_PRICES_USD = {"ATOM": "4", "OSMO": "0.3", "TIA": "1.5", "DYDX": "0.5", "ISLM": "0.004", "INJ": "8",
                     "JUNO": "0.1", "BAND": "0.3", "LUNA": "0.1", "SOMM": "0.01", "SAGA": "0.1", "STRD": "0.05"}
# (rough figures set 2026-10-09; the floor is a packet-count knob, not accounting, so a factor of two is fine)
# Stride-native sweep denoms (to Osmosis over channel-5): the eleven stTokens and ustrd.
NATIVE_SWEEP_DENOMS = ("stuatom", "stuosmo", "stutia", "stinj", "stadydx", "staISLM", "stujuno", "stuband",
                       "stuluna", "stusomm", "stusaga", "ustrd")
# Vouchers that unwind one hop: (base denom, Stride channel it arrived on). Must stay a subset of
# types.SweepUnwindChannels in x/stakeibc/types/wind_down.go; the planner re-checks each against the live denom trace.
VOUCHER_SWEEP_DENOMS = (("uatom", "channel-0"), ("utia", "channel-162"), ("uosmo", "channel-5"), ("ujuno", "channel-24"),
                        ("usomm", "channel-150"), ("usaga", "channel-213"), ("adydx", "channel-160"))
UNWIND_CHANNELS = {"channel-0": "cosmos", "channel-162": "celestia", "channel-5": "osmo", "channel-24": "juno",
                   "channel-150": "somm", "channel-213": "saga", "channel-160": "dydx"}
PROTOCOL_ADDRESSES = {...the nine from the old builder...}

MAX_ADDRESSES_PER_BATCH = 100
GAS_BUDGET_PER_BATCH = 40_000_000     # the planner packs under this
BLOCK_GAS_LIMIT = 100_000_000         # the runner refuses a simulated batch over this
GAS_PER_TRANSFER_DEFAULT = 150_000    # until the ledger has a confirmed batch to calibrate from
GAS_ADJUSTMENT = 1.3; GAS_PRICE = "0.001ustrd"
MAX_PLAN_AGE_SECONDS = 6 * 3600; TX_POLL_SECONDS = 3; TX_WAIT_SECONDS = 180
```

An stToken's price is `NATIVE_PRICES_USD[symbol] × redemption_rate` of its host zone (`stakeibc/host_zone`;
stutia from `staketia/host_zone`). A voucher's and ustrd's price is the native figure.

## exclusions.json

```json
{
  "sections": [
    {"name": "team", "reason": "moved by hand; v35 also sends the community pool here",
     "addresses": [{"address": "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh", "label": "F5 team multisig"}]},
    {"name": "relayers", "reason": "their STRD pays for relaying until the halt",
     "addresses": [{"address": "stride1e6llcr7fkxvqdgyrcgzdlwll9tkvfh2rnfcpyd", "label": "main relayer"}, ...]}
  ]
}
```

Seeded with exactly today's `BUILDER_EXCLUDED_ADDRESSES` (F5 multisig plus the seventeen relayer keys,
labels from the code comments). Loading validates every address is bech32 `stride` and 20 bytes and
that no address repeats; a bad file fails `plan` and `run`.

## holders.py: live reads and classification

Reads, all through `chainio` (each a thin REST call the tests stub):

1. **Height**: the latest block, recorded in the plan.
2. **Denoms**: for each native sweep denom, the chain's rule (`isSweepableNativeDenom`): `ustrd`,
   `stutia`, or `st<host_denom>` of a host zone that is not deprecated. For each voucher, the
   `ibc/` hash of `transfer/<channel>/<base>` must have a denom trace on chain whose outermost hop is
   that transfer channel, and the channel must be in `UNWIND_CHANNELS`. A denom failing either rule
   fails `plan` naming it (the chain would reject the whole batch).
3. **Holders**: `bank/v1beta1/denom_owners_by_query?denom=<d>` paginated (1,000 per page), per sweep
   denom. Aggregated into `address -> {denom: amount}`.
4. **Candidates**: addresses whose USD total across sweep denoms is at or above the floor, plus
   `--test`'s address whatever its value. Only candidates get the per-address reads below.
5. **Accounts**: `auth/v1beta1/accounts/<address>` for each candidate not in `state/accounts.json`.
   Cached fields: `type`, `has_pubkey`, `sequence`. A cached entry with `has_pubkey: false` is
   re-read on every plan (it may have signed since). The cache is written back after each plan.
6. **Spendable**: for a candidate whose type is a vesting account, `bank/v1beta1/spendable_balances`
   replaces its balances for the sweep denoms. That is exactly what the chain moves, so the old
   vesting mirror is gone.
7. **Skip inputs**, one call each: module accounts (`auth/v1beta1/module_accounts`), transfer
   escrows (every `ibc/core/channel/v1/channels` with port `transfer`, via the escrow derivation
   already in `dashboard/chain.py`, copied into the package), wasm contracts (`cosmwasm/wasm/v1/code`
   then `code/<id>/contracts` per code; only 20-byte ones can collide with a holder).

Classification, in the chain's order (`sweepSkipReason`), then the builder's own rules:

| Reason | Rule |
| --- | --- |
| `address is not 20 bytes` | bech32 decode |
| `protocol address` | `PROTOCOL_ADDRESSES` or the operator |
| `blocked module address` | in the module-account set or `sha256(name)[:20]` of the blocked names |
| `transfer escrow address` | escrow set |
| `account not found` | auth lookup 404 |
| `interchain account` / `account type … is not sweepable` | type not BaseAccount or a vesting type |
| `wasm contract address` | contract set (builder-only) |
| `excluded: <section>: <label>` | exclusions file (builder-only) |

Output: `HolderSet(height, denoms: [SweepDenom], holders: [Holder(address, balances, usd, keyless, locked)],
excluded: [Excluded(address, reason, usd)], below_floor: [Holder])`. `below_floor` carries every
sweepable holder under the floor, so the planner can print the floor ladder and the dashboard can show
"below floor". `locked` is a vesting holder's total minus spendable per denom (empty otherwise); the planner
carries it into each planned address so the dashboard does not read the remainder as a refund.

## planner.py: tiers and batches

Tiers, in the order the runner walks them:

| Tier | Members | When |
| --- | --- | --- |
| `test` | the `--test` address only | the plan has nothing else when `--test` is given |
| `canary` | the `--canary N` smallest holders at or above the floor (default 0) | sweep day, first run |
| `main` | every other non-keyless holder, by USD descending | |
| `keyless` | keyless holders, by USD descending | last |

Batch packing within a tier: walk the holders in order, start a new batch when the next holder would
push the batch over `MAX_ADDRESSES_PER_BATCH` or over `GAS_BUDGET_PER_BATCH` of estimated gas. A
holder's transfers are its sweep denoms with a positive spendable amount; estimated gas is transfers ×
`gas_per_transfer`. `gas_per_transfer` is calibrated from the ledger: the largest `gas_used /
transfers` among confirmed batches, times 1.2, else `GAS_PER_TRANSFER_DEFAULT`. The test batch is
what calibrates the real run.

`state/plan.json`:

```json
{"run_id": 3, "created_at": "...", "height": "14201338", "floor_usd": "25", "test": false, "canary": 3,
 "gas_per_transfer": "118000", "prices": {"stuatom": "4.71", ...},
 "denoms": [{"denom": "stuatom", "symbol": "stATOM", "decimals": 6, "destination": "osmosis-1", "channel": "channel-5"}, ...],
 "tiers": [{"name": "main", "batches": [{"id": "003-001", "file": "batch-003-001.txt", "sha256": "...",
            "addresses": [{"address": "...", "balances": {"stuatom": "1200"}, "usd": "5.65", "transfers": 2}],
            "usd": "...", "transfers": 190, "estimated_gas": "22420000"}]}],
 "excluded": [{"address": "...", "reason": "excluded: team: F5 team multisig", "usd": "..."}],
 "skipped": [{"address": "...", "reason": "wasm contract address", "usd": "..."}],
 "below_floor": {"count": 38412, "usd": "42760.11"},
 "ladder": [{"floor": "25", "holders": 1790, "usd": "..."}, {"floor": "10", ...}, {"floor": "5", ...}, {"floor": "1", ...}, {"floor": "0", ...}]}
```

Batch ids are `<run>-<n>` so ledger lines from different runs never collide. Batch files hold one
address per line (the CLI's input). `plan` overwrites the previous plan and its batch files; the
ledger is never rewritten. `plan` refuses while the ledger has a batch submitted but unresolved.

## ledger.py

`state/ledger.jsonl`, one JSON object per line, appended with a single `write` of the whole line and
`flush` + `fsync`, so a crash never leaves a partial record:

| kind | fields |
| --- | --- |
| `submitted` | `run_id, batch_id, tx_hash, at, addresses (count), transfers (estimate), gas_wanted` |
| `confirmed` | `batch_id, tx_hash, height, gas_used, at, transfers: [{address, denom, amount, channel, receiver}], skipped: [{address, reason}]` |
| `failed` | `batch_id, tx_hash, code, codespace, raw_log, at` (the broadcast was rejected or the tx failed on chain) |
| `lost` | `batch_id, tx_hash, at` (`resolve` gave up waiting; the next `plan` reads the truth from balances) |

`ledger.read() -> [Event]`, `ledger.batch_states(events) -> {batch_id: pending | submitted | confirmed |
failed | lost}`, `ledger.unresolved(events) -> [submitted events with no terminal event]`,
`ledger.calibration(events) -> gas_per_transfer | None`.

## sweep.py: the CLI

```
python3 scripts/wind-down/sweep/cli.py plan --floor-usd 100 [--canary 3] [--test] [--max-addresses 100] [--gas-budget 40000000]
python3 scripts/wind-down/sweep/cli.py run [--tier test|canary|main|keyless] [--batches N] [--yes] [--dry-run] [--continue-on-skip]
python3 scripts/wind-down/sweep/cli.py status
python3 scripts/wind-down/sweep/cli.py resolve [--wait-seconds 600]
```

**plan** reads live state, classifies, packs, writes the state files and prints: the floor, height,
each tier's batch count, address count and USD; the excluded table with live USD; the skipped table;
the keyless count; the floor ladder; and the one-line `RESULT: PLANNED …`. `--test` plans only the
test address (floor ignored for it) and the tier is `test`.

**run** does, in order:

1. Preflight, every check printed as a row and any failure stops before anything is signed:
   - the node's chain id is `stride-1` and `strided version` starts with `v35`;
   - `strided keys show stride-sweeper --keyring-backend test -a` prints `SWEEP_OPERATOR`;
   - the plan is younger than `MAX_PLAN_AGE_SECONDS` and every batch file's sha256 matches the plan;
   - the ledger has no unresolved submitted batch (else: run `resolve`);
   - the operator's on-chain account sequence equals `plan.operator_sequence` (read at `plan`) plus the number of
     `submitted` lines of this run (lost ones excepted): a tx signed without a ledger line fails the run, naming both
     numbers. The same equality is re-checked immediately before every broadcast;
   - no address in any pending batch is in the exclusions file (re-read now), the protocol set, or
     is the operator: the exclusion file is enforced at signing time, not only at planning;
   - every destination channel the plan uses is OPEN on Stride's side;
   - the operator's `ustrd` balance covers the fees of every pending batch (estimated gas × price).
2. For each pending batch in tier order (a plain `run` skips the `keyless` tier and prints how many batches it left
   for `run --tier keyless`; `--tier` runs only that tier; optionally at most N batches): print
   the batch (id, tier, addresses, transfers, USD, the three largest holders); unless `--yes`, ask
   `sweep batch 003-001? [y/N/a]` (`a` = yes to all for this run).
3. Simulate: the CLI command with `--dry-run`, parsing the SDK's `gas estimate: N`. Refuse the batch
   and stop if `N × GAS_ADJUSTMENT > BLOCK_GAS_LIMIT` (re-plan with a smaller `--gas-budget`).
   `--dry-run` on `run` stops here for every batch and prints the commands and estimates. The simulation passes
   `--from <SWEEP_OPERATOR address>` (the SDK requires a bech32 address when simulating) and no `--gas`; the broadcast
   passes `--from stride-sweeper`.
4. Broadcast with `--gas <N × GAS_ADJUSTMENT> --gas-prices GAS_PRICE --broadcast-mode sync -y
   --output json`. A non-zero `code` in the response is a `failed` event and stops the run. Else
   append `submitted`. A non-zero `strided` exit or an unparseable reply stops the run with nothing in the ledger
   (the tx may have been broadcast); the next run's operator-sequence check finds out.
5. Poll `tx/v1beta1/txs/<hash>` every `TX_POLL_SECONDS` up to `TX_WAIT_SECONDS`. Found with code 0:
   parse `sweep_transfer` and `sweep_skipped` events, append `confirmed`, print `transfers=… skipped=…
   gas_used=…`. Found with a non-zero code: `failed`, stop. Not found in time: leave it `submitted`
   and stop (run `resolve`).
6. Any skip event stops the run after recording it, because it means the planner and the chain
   disagree; `--continue-on-skip` overrides. Any `failed` always stops.

Each `strided` invocation is one line of this shape (the test suite asserts it):

```
strided tx stakeibc sweep-tokens-off-stride <denoms> <file> --from stride-sweeper --keyring-backend test
  --chain-id stride-1 --node <RPC> --gas <n> --gas-prices 0.001ustrd --broadcast-mode sync -y --output json
strided tx stakeibc sweep-tokens-off-stride <denoms> <file> --from <SWEEP_OPERATOR> --keyring-backend test
  --chain-id stride-1 --node <RPC> --gas-prices 0.001ustrd --dry-run
```

**status** needs no network: per tier, batches by state, confirmed addresses and USD, skipped and
failed batches with reasons, unresolved submissions.

**resolve** polls each unresolved submission until found or `--wait-seconds` after its `at`, then
appends `confirmed`/`failed`/`lost`.

The test flow is therefore: `plan --test`, `run` (one batch, prompts once), then check the Osmosis
address holds the stToken and STRD and the Hub address holds the ATOM, and the Sweep tab shows the
batch confirmed. The test address should hold a little of an stToken, some ustrd, and an ATOM voucher
so all three destination rules are exercised.

## Dashboard Sweep tab

`dashboard/sweep_tab.py` has two parts:

- `collect() -> dict`: the expensive live read, the Sweep tab's collector in `COLLECTORS` with a
  1,800 s interval. It calls `holders.read_balances(denoms)` (the same bulk `denom_owners` reads the
  planner uses) and returns `{height, prices, balances: {address: {denom: amount}}}` for every holder
  of every sweep denom.
- `compose(plan, ledger_events, live) -> dict`: pure, the route body. `GET /api/sweep` reads
  `plan.json` and `ledger.jsonl` from disk on every request (like `/api/ops`), takes the live
  snapshot from the cache, and returns `{fetched_at, refreshing, data}` where `fetched_at` is the
  live snapshot's. Before the first live snapshot `data` still carries the plan and ledger parts with
  the live-dependent fields null.

Per planned address, with `transferred` = the denoms its confirmed `sweep_transfer` events moved and
`live` = its current balances of the sweep denoms:

| state | rule |
| --- | --- |
| `swept` | transferred non-empty (any run) and live holds none of the transferred denoms beyond the address's planned `locked` remainder |
| `refunded` | transferred non-empty (any run) and live holds a transferred denom above its `locked` remainder (timeout refund) |
| `remaining` | no confirmed transfer yet (batch pending, submitted, failed or lost) |
| `excluded` | in the plan's excluded list |

Payload `data`:

- `run`: `run_id, floor_usd, created_at, height, test, batches {total, confirmed, submitted, failed, lost, pending}, operator_strd, fee_estimate_ustrd`.
- `totals`: for `swept, remaining, refunded, excluded, below_floor`: `{addresses, usd}` at the plan's prices on live balances (swept USD is the transferred amounts at plan prices).
- `by_denom`: per sweep denom: `symbol, destination, channel, swept {addresses, amount, usd}, remaining {…}, refunded_addresses, excluded_usd, below_floor_usd`.
- `batches`: every ledger-known and plan batch, grouped by run: `run_id, batch_id, tier, addresses, usd, state, tx_hash, at, skipped: [{address, reason}], refunded: count`.
- `ladder`: live sweepable holders not yet swept and not excluded, at the plan's floor and at 10, 5, 1, 0 below it: `{floor, holders, usd}` (a per-rung `batches` count was designed but is not implemented).
- `refunded`: `[{address, denom, amount, channel, batch_id}]`.
- `exclusions`: the file's sections with each address's live USD.
- `keyless`: `{addresses, usd}` of the keyless tier and how many of them are swept.

`sweep.js` renders what the signed-off mock shows: the run line, the five tiles, the progress bar,
the by-token table, the runs-and-batches table beside the ladder and refunded panels, the collapsed
exclusions and keyless panels. A batch with skips gets a red `N skipped` badge with the reasons in
its title. The tab's refresh button posts `/api/refresh/sweep`; on selecting the tab the page posts
one when the live snapshot is older than ten minutes or missing. With no `plan.json` the tab says so
and shows nothing else.

## Hardening (the full list)

- Exclusions enforced twice: at `plan` and again at `run`, re-read from the file.
- The protocol set and the operator can never be in a batch, checked at `run`.
- Batch files are content-hashed in the plan; a mismatch stops `run`.
- A stale plan (older than six hours) stops `run`; re-plan.
- Gas: packed under a budget at plan time, simulated per batch at run time, refused over the block
  limit. Fees pre-checked against the operator's balance.
- Channels checked OPEN before signing (a closed channel rejects the whole tx on chain anyway, but
  the check names it first).
- Every batch id is submitted at most once; resume is a ledger read; nothing in the ledger is ever
  rewritten.
- Any skip event or failed tx stops the loop.
- `--dry-run` prints every command without broadcasting.
- The ledger and plan are plain files committed to the repo, so the team can read the same state.

## Testing

No network anywhere; `chainio` is the stub seam (REST calls and the `strided` subprocess).

- `sweep/test_holders.py`: a synthetic chain (denom owners, accounts of each type, a vesting account
  with a smaller spendable balance, module accounts, an escrow, a 20-byte contract, the test address)
  → every skip reason, the floor, keyless detection, the exclusions file, and a denom the chain would
  refuse (a deprecated zone's stToken; a voucher whose outer hop is not whitelisted).
- `sweep/test_planner.py`: tier order and membership, canary picks the smallest, keyless last,
  packing by addresses and by gas, calibration from a ledger, `--test` plan shape, batch file hashes,
  ladder counts, refusal while a submission is unresolved.
- `sweep/test_ledger.py`: append, read, batch states, unresolved, calibration; a truncated last line
  is reported, not swallowed.
- `sweep/test_cli.py`: the preflight rows (each failure mode), the exact `strided` command line,
  gas estimate parsing, the broadcast and poll loop against a fixture `tx_responses` with
  `sweep_transfer` and `sweep_skipped` events, stop on skip, stop on failure, `--yes`/`--batches`,
  `resolve` outcomes.
- `dashboard/test_sweep_tab.py`: `compose` over a plan, ledger and live snapshot → each address state,
  totals, by-denom, ladder, refunded, batches; the no-plan and no-live cases.
- Unit tests of the existing dashboard keep passing (`python3 -m unittest discover -s scripts/wind-down/dashboard`).

## Docs and plan updates

- `scripts/wind-down/dashboard/README.md`: a "Sweep tab" section in the style of the others.
- `scripts/wind-down/sweep/README.md`: the four commands, the test flow, the sweep-day loop
  (plan at the announced floor with `--canary 3`, run, re-plan lower, run, …, the keyless tier last),
  and how to resolve a stuck submission.
- `dashboard/ops/plan.json`: `sweep-export` becomes `sweep-plan` (the `plan` command at the announced
  floor with the canary), `sweep-submit` becomes the `run` command with the ledger and the Sweep tab
  as what to watch, `vote-finalize-sweep` names `sweep/config.py` and `sweep/exclusions.json`, and a
  new `sweep-test` step in the first block after the upgrade: `plan --test` then `run`, then check
  the destination balances. `test_ops.py` is updated for the renamed ids.
- The design spec's §7/§9 mentions of `build_sweep_batches.py` get a one-line pointer to this spec.

## Out of scope

- Sweeping denoms outside the configured list (comdex's stToken, Axelar USDC): add to config if
  decided.
- A UI to edit exclusions or trigger a run from the dashboard.
- Osmosis-side verification that the swept tokens arrived (the Funds and Pools tabs cover the
  vault; holders' own balances are theirs to check).
