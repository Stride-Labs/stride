# Wind-down dashboard: Pools tab, transfer and pool-funding multisig sets, uniform transfer-day checklist

Date: 2026-10-07. Extends `2026-10-07-wind-down-multisig-tab-design.md` (Multisig tab, `auto` checks,
`start` gating). The dashboard lives in `scripts/wind-down/dashboard/`.

## Why

The transfer days (10/20 haqq, 10/27 osmosis + celestia, 11/3 six zones, 11/10 juno + sommelier) run
the same sequence for every zone, but the plan currently spells it out five times with templated
commands and two scripts (`coverage_check.py`, `check_transmuter_pool.py`) whose inputs are hand-typed
constants and an export. Decisions (2026-10-07):

- A **Pools tab** absorbs `check_transmuter_pool.py` (which is retired) and does the coverage math live;
  `coverage_check.py` stays only for the published export-based audit at the halt.
- The Multisig tab gains the **ICA transfer** set (per zone, per ICA: a 1-token test tx then the real
  one), the **staketia claim balance** set and the **pool funding** set (per pool: a 1-token test join,
  the remaining allocation, then `mark_corrupted_assets`). Osmosis txs use `osmosisd` with the same key
  names (`FS5`/`FA5`/`FR5`, multisig `F5`) and `https://osmosis-strd-rpc.polkachu.com:443`; the vault
  `osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af` is a 2-of-3 legacy-amino multisig on chain (account
  818208).
- Every transfer-day block becomes the same eight bullets (celestia with its staketia bullets in
  addition), each live-checked where a check exists.
- Pools are created in the voting week at the then-current rate and only joined on the transfer days;
  the frozen Stride rate may end up above a pool's rate, and that surplus stays in the canonical pool.

## 1. Pools tab (`pools.py`, `GET /api/pools`, `static/pools.js`)

A collector like Channels/Validators/Funds (`COLLECTORS["pools"]`, refresh 300 s), fourth tab button
after Funds flow and before Multisig. Payload `{zones: [ZonePools]}`, one entry per `config.ZONES` zone
(plus `error` entries on failure, as the other tabs do).

### Discovery

Transmuter pools are cosmwasm pools: `/osmosis/cosmwasmpool/v1beta1/pools` (paginated; entries carry
`contract_address` and `code_id`; the pool id comes from `/osmosis/poolmanager/v1beta1/pools/{id}` or
the cosmwasmpool listing's `pool_id` field — use whichever the listing exposes). A pool belongs to us
when `code_id == "996"` and its `get_admin` is the vault; `get_admin` answers are memoised per contract
for the process lifetime (adminship never changes for pools that are not ours). `config.EXTRA_POOL_CONTRACTS`
is added unconditionally. A pool is assigned to a zone by the stToken denom it holds (base denom from
the Osmosis denom trace == the zone's `st_denom`); its kind is `canonical` when that trace is exactly
`transfer/channel-326/<st_denom>` (Stride → Osmosis), else `route`.

### Route resolution (route pools)

The pool's stToken trace is `transfer/<c1>/transfer/<c2>/<st_denom>`: `c1` is Osmosis's channel to the
holder chain X, `c2` is X's channel to Stride. X's chain id comes from Osmosis
`/ibc/core/channel/v1/channels/<c1>/ports/transfer/client_state`. Stride's channel to X is the one
among `config.REQUIRED_ROUTES[st_denom]` whose `/ibc/core/channel/v1/channels/<ch>/ports/transfer`
has `counterparty.channel_id == c2` and whose client tracks X; none → the pool is `unrecognised`
(reported, not allocated). The escrow balance is Stride's bank balance of
`chain.escrow_address(stride_channel)` (port ibc-go's `GetEscrowAddress`: sha256 of
`"ics20-1" + NUL + "transfer/<channel>"`, first 20 bytes, bech32 `stride`; the bech32 encoder from
`scripts/wind-down/bech32_ref.py` is copied into `chain.py`, since the dashboard does not import from
its parent directory) in `st_denom`.

`config.REQUIRED_ROUTES` is a copy of `coverage_check.REQUIRED_ROUTES` (stToken denom → frozenset of
Stride channel ids); a test loads `scripts/wind-down/coverage_check.py` by path with `importlib` and
asserts the two are equal. `config.OSMOSIS_CHANNEL_TO_HOST` is copied from the retired script (zone →
Osmosis's channel to the host, the source of the native token's canonical Osmosis denom).

### Per pool: `PoolReport`

```
contract, pool_id (str|null), kind ("canonical"|"route"|"unrecognised"),
route: {chain_id, stride_channel, counterparty_channel} | null,
st_denom (ibc/… on Osmosis), st_trace (full path), st_balance, native_balance,
alloyed_denom, alloyed_supply, vault_shares, outside_shares (= alloyed_supply - vault_shares),
rate (from the factors, as the Funds tab computes it), rate_gap_pct (stride_rate - rate, % of stride_rate),
escrow (str|null, route only), allocation (str|null), funded_exactly (bool|null), corrupted: [denoms],
native_marked (bool), checks: [{name, ok: bool|null, detail}], ready (bool: every check ok)
```

Integers are strings. `allocation`: route pool = `ceil(escrow × rate)` using the **pool's own** rate
(what it will actually pay out; the gap to the frozen rate stays in the canonical pool); canonical =
`vault_native + Σ pools' native − Σ route allocations` (everything else the zone has on Osmosis),
null when any input is unknown. `funded_exactly` = `vault_shares == allocation` (shares are minted
1:1 to native joined and stay in the vault, so this is the cumulative funding audit the spec says the
coverage check is not; redemptions later lower `native_balance` but not `vault_shares`).

Checks (ported from `check_transmuter_pool.py`, each `ok` null when its lookup failed):
`code id 996`; `cw2 version 3.2.0` (raw key `contract_info`); `assets: one stToken denom + the native,
nothing else`; `factors: stToken 1e18, native == alloyed`; `rate ≤ Stride's rate` (detail shows the
gap; a pool priced above the frozen rate would pay out more than its backing); `admin is the vault`;
`moderator is the vault`; `no admin transfer in flight` (`get_admin_candidate` null); `active`;
`no limiters`; `stToken trace` (canonical: equals `transfer/channel-326/<st_denom>`; route: two hops
ending in `/<st_denom>` and resolved to a policy channel); `native trace equals
transfer/<OSMOSIS_CHANNEL_TO_HOST[zone]>/<host_denom>`; `uint128 headroom` (native_balance × native
factor < 2^128); `corrupted set` (ok when empty and `vault_shares == 0`, or exactly `[native]` and
`vault_shares > 0`); `no alloyed shares outside the vault`.

### Per zone: `ZonePools`

```
chain_id, symbol, decimals, st_denom, osmosis_denom, host_denom, stride_rate, st_supply, needed (supply × stride_rate, rounded up),
vault_native, pools_native, native_on_osmosis (vault + pools), coverage (ratio str|null),
missing_routes: [stride channel ids in REQUIRED_ROUTES[st_denom] with no pool],
pools: [PoolReport] (canonical first, then routes by stride_channel),
pools_ready: bool|null (a canonical pool exists, no missing route, every pool ready),
pools_funded: bool|null (pools_ready and every pool funded_exactly and native_marked)
```

`static/pools.js`: per zone a panel: header (needed vs native on Osmosis, coverage pill, rate and gap),
a missing-routes warning, then one row per pool: kind/route (chain id + Stride channel), pool id,
native, stToken, vault shares / outside shares, allocation and funded-exactly mark, corrupted mark, and
the checks as a compact list of ✓/✗/n/a with the detail on hover; a failing check turns the row red.
Zone panels collapsed by default except those with a failing check or a missing route.

## 2. Funds and Channels additions

Funds zone payload: `ica_balances` (`{ICA: {denom: amount}}`, every denom each ICA holds — the
foreign-denom transfers read it), `funds_settled` (staked < one whole token, unbonding == 0,
in_flight == 0; null when unknown), `transfers_landed` (every ICA's host-denom balance < one whole
token and in_flight == 0; null when unknown). Channels zone payload: `ica_channels_open` (every ICA
row — all rows except the transfer one — `STATE_OPEN` on both ends; null when any state is unknown).

## 3. Multisig sets

`multisig.tx_sets(validators_data, funds_data, pools_data)` (all `dict | None`). `MultisigTx` keeps
its shape; sets may hold several txs per zone, and `static/multisig.js` groups a set's txs by
`chain_id` under a zone heading with anchor `set-<set-id>-<chain_id>`; the hash `#multisig/<set>/<zone>`
scrolls to that heading (`#multisig/<set>` to the set, as today). The Ops step field `multisig` may be
`"<set-id>"` or `"<set-id>/<zone>"`; a step with `zones` and a bare set id renders one link per zone.
Commands follow the existing generate / sign ×3 / multisign+broadcast shape with the keyring note.

- **`ica-transfers`** (step suffix `transfers`), Stride, per zone, in order FEE, WITHDRAWAL,
  DELEGATION, REDEMPTION; for each ICA a `test` tx of one whole token (`10**decimals` + host denom) and a
  `rest` tx of the ICA's current host-denom balance from the Funds snapshot (title says
  "live balance, copy after the test has settled"); `ready=False` with a reason when the balance is
  below the test amount (nothing to send) or the snapshot is missing. Each foreign denom an ICA holds
  (e.g. dYdX's USDC voucher) gets one `rest`-style tx for its full balance.
  `strided tx stakeibc transfer-from-ica <chain_id> <ICA> <amount><denom> --from <MULTISIG_ADDRESS> --generate-only --chain-id stride-1 --node <NODE> --gas 600000 --fees 3000ustrd > /tmp/wind-down/transfer-<chain_id>-<ica lower>-<test|rest|denom-suffix>.unsigned.json`.
- **`staketia-claim-balance`** (celestia only): `test` = `1000000` utia, `rest` = `0` (the whole
  remainder): `strided tx stakeibc transfer-staketia-claim-balance <amount> …` same flags.
- **`pool-funding`** (step suffix `join-pools`), Osmosis, per zone, per pool from the Pools snapshot
  (canonical first): `test join` (one whole native token), `join rest` (`allocation − 10**decimals`,
  `ready=False` when the allocation is unknown or already funded exactly), `mark corrupted`.
  Generate: `osmosisd tx wasm execute <contract> '<msg>' [--amount <n><osmosis_denom>] --from osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af --generate-only --chain-id osmosis-1 --node https://osmosis-strd-rpc.polkachu.com:443 --gas 1500000 --fees 15000uosmo > /tmp/wind-down/pool-<chain_id>-<pool_id or contract[-6:]>-<test|rest|mark>.unsigned.json`
  with `{"join_pool":{}}` for joins and `{"mark_corrupted_assets":{"denoms":["<osmosis_denom>"]}}` for
  the mark (transmuter v3.2.0 execute variants). Sign / multisign / broadcast with `osmosisd`, the same
  key names and the Osmosis node. The set description states the order per zone: every pool's test
  join, then verify on the Pools tab (native one token, vault shares one token, no outside shares, rate
  exact), then every rest join, then every mark — the mark closes the outsider-join window, so rest
  join and mark go out back to back per pool. It also notes the vault must hold OSMO for about 3 txs
  per pool at 0.015 OSMO each.

Server: `/api/multisig` composes from the validators, funds and pools caches' current views;
`fetched_at` is the oldest of the three that exist.

## 4. Plan (`ops/plan.json`)

Each transfer-day zone group (prefixes `haqq`, `osmo`, `tia`, `d21`, `d28`) becomes exactly these steps,
ids `<prefix>-<suffix>`, every step with the group's `zones`:

| suffix | text (abridged) | live check / link |
|---|---|---|
| `claimable` | as today | funds `records.pending_before_claimable == 0` |
| `claims` | as today (claims bot) | funds `records.user_redemption_records == 0` |
| `settled` | Funds settled for the zone: nothing staked beyond dust, nothing unbonding, nothing in flight. | funds `funds_settled` |
| `channels` | All four ICA channels OPEN on both ends (a FEE or WITHDRAWAL ICA whose packet timed out before the upgrade left its ordered channel closed and transfer-from-ica is rejected with 'no active channel': restore-interchain-account it first). | channels `ica_channels_open` |
| `transfers` | Submit the four admin transfers to the Osmosis vault, FEE, WITHDRAWAL, DELEGATION, REDEMPTION in that order, each as a 1-token test tx first and the real one once the test shows settled (Funds tab → ICA transfers); a foreign denom in an ICA (dYdX's USDC) is its own tx. 24h timeout: a refund means resubmit. | `multisig: "ica-transfers"`; expect: code 0 each, the Funds tab lists each transfer in flight then settled |
| `landed` | Every transfer landed in the Osmosis vault: the four ICAs at dust, nothing in flight. | funds `transfers_landed` |
| `join-pools` | Join the pools: per pool a 1-token test join, check the Pools tab, then the remaining allocation, then mark the native token corrupted; every pool of the zone funded exactly and marked. | `multisig: "pool-funding"`, pools `pools_funded` |
| `announce` | as today | — |

Celestia keeps, between `claims` and `settled`: `tia-staketia-sweep` (as today) and `tia-staketia-paid`
(gains `auto: funds records.staketia_claim_ready equals true`); and between `channels` and
`transfers`: `tia-claim-balance` ("MsgTransferStaketiaClaimBalance: 1 TIA as the test, then amount 0
for the remainder; the TIA lands on the celestia delegation ICA and leaves with the DELEGATION transfer",
`multisig: "staketia-claim-balance"`). `tia-tranche1-sweep` and `strd-unbonding-complete` stay where
they are. Removed: every `*-done-transfer-checklist`, `*-done-arrived`, `*-done-coverage`,
`*-done-fund-pools`, `post-haqq-withdrawal` (folded into `channels`), `tia-claim-balance-test/rest`.
Day notes change "once the coverage check passes" to "once the Pools tab shows the zone covered".

`vote-pools-create`: drop the `command`/`expect`; text ends "The Pools tab lists every code-996 pool the
vault administers with its checks: one canonical pool plus one per in-scope route per zone, no missing
route, every check green (the rate gap to Stride is expected and shown); nothing is funded yet. Top the
vault up with about 10 OSMO for the funding txs." with `zones` (all eleven) and
`auto: pools pools_ready`. Any other step whose `command` names `check_transmuter_pool.py` or
`coverage_check.py` before the halt loses that command and points at the Pools tab instead; the halt
block's coverage step keeps `coverage_check.py` (the published audit).

## 5. Retirement

`git rm scripts/wind-down/check_transmuter_pool.py scripts/wind-down/test_check_transmuter_pool.py`;
fix any import of it (`grep -rn check_transmuter scripts docs/wind-down`); README gets a Pools tab
section and the new Funds/Channels/Multisig fields; `docs/wind-down/*.md` mentions of the script get a
one-line "replaced by the dashboard's Pools tab" note where they tell the operator to run it.

## Testing

- `test_pools.py`: with fakes for the Osmosis and Stride lookups: discovery keeps code-996 vault-admin
  pools and extras only; canonical vs route classification from the trace; route resolution picks the
  policy channel whose counterparty matches; allocation = ceil(escrow × pool rate) for a route, the
  remainder for canonical, null when the escrow is unknown; `funded_exactly` from vault shares; every
  check's pass and fail case (inverted factors, foreign admin, a limiter, a second stToken, corrupted
  set wrong for the phase, outside shares); `missing_routes`; `pools_ready` / `pools_funded`;
  `chain.escrow_address("channel-0")` equals the known Stride escrow address for channel-0 (take it
  from `coverage_check.escrow_address` in the test); `config.REQUIRED_ROUTES` equals the script's.
- `test_funds.py` / `test_channels.py`: the three new fields, pass / fail / null.
- `test_multisig.py`: full-string equality for one ICA transfer test + rest pair, a foreign-denom tx,
  the staketia pair, and one pool's test join / rest join / mark; not-ready reasons; the per-zone
  grouping order.
- `test_ops.py`: every transfer-day group has exactly the step suffixes above in order (celestia with its
  extras); `multisig` references (with or without `/zone`) resolve; no step command names the retired
  script.
- Frontend: headless Chrome screenshots of the Pools tab (today: no real pools, the test pools from 09-25
  are not vault-administered so the tab shows zones with a missing canonical pool), the Multisig tab's
  ICA-transfers set grouped by zone, and one transfer-day Ops block.

## Build plan

**Chunk 1 — Pools collector** (`pools.py`, `test_pools.py`, `chain.py` (bech32 + escrow address),
`config.py` (`REQUIRED_ROUTES`, `OSMOSIS_CHANNEL_TO_HOST`), `server.py` (collector + interval), the
Pools tab section of `README.md`). Produces the `/api/pools` payload exactly as §1. Depends on: none.

**Chunk 2 — Funds/Channels fields and multisig sets** (`funds.py`, `channels.py`, `multisig.py`,
`server.py` (`/api/multisig` composition from three caches), their tests). Consumes the §1 payload shape
(build against the spec; a fixture, not chunk 1's code). Depends on: none (merge after chunk 1).

**Chunk 3 — Frontend, plan, retirement** (`static/pools.js`, `static/index.html`, `static/app.js`,
`static/multisig.js` (zone grouping + anchors), `static/ops.js` (`<set>/<zone>` links), `static/style.css`,
`ops/plan.json`, `test_ops.py`, `git rm` of the script and its test, README/docs notes). Consumes the
§1 and §3 shapes from the spec. Depends on: none (merge last).
