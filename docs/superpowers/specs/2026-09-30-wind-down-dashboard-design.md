# Wind-Down Dashboard

Status: approved design (mock v1 approved 2026-09-30). Medium tier. Throwaway ops tool, run
locally during the v35 ops window (`2026-09-18-protocol-wind-down-design.md`).

## Goal

One local web page that answers three questions during the wind-down without anyone running
ad-hoc queries:

1. Are the channels the wind-down depends on open and clear? (Channels tab)
2. Does Stride's recorded delegation per validator match what the host chain has? (Validators tab)
3. Where is each zone's backing right now, from staked to the Osmosis pools? (Funds flow tab)

Read-only. It never signs or submits anything.

## Shape

`scripts/wind-down/dashboard/`, Python 3.12 stdlib only (no install step), plus one static page.

    python3 scripts/wind-down/dashboard/server.py        # then open http://localhost:8787

The server queries the chains (the browser never does: no CORS problems, and the private
endpoints stay out of the page), caches one snapshot per tab in memory, and refreshes each on
its own interval in a background thread. The page polls the cached snapshots.

| File | Purpose |
|---|---|
| `server.py` | `http.server` app: serves `static/`, `GET /api/<tab>`, `POST /api/refresh/<tab>`; owns the snapshot cache and the refresh threads. |
| `config.py` | Every constant: endpoints, in-scope zones, channel maps, operator addresses, holder routes, refresh intervals. |
| `chain.py` | HTTP and Cosmos query helpers shared by the collectors (REST, CometBFT RPC, pagination, IBC denom hash). |
| `channels.py` | Collector for the Channels tab. |
| `validators.py` | Collector for the Validators tab. |
| `funds.py` | Collector for the Funds flow tab. |
| `static/index.html`, `static/app.js`, `static/channels.js`, `static/validators.js`, `static/funds.js`, `static/style.css` | The page: shell plus one render module per tab. Vanilla JS, no build. |
| `test_*.py` | `unittest` for the pure logic, run with `python3 -m unittest discover -s scripts/wind-down/dashboard`. |
| `README.md` | How to run, what each tab reads, known gaps. |

The visual reference is the approved mock,
`.superpowers/brainstorm/45165-1790786005/content/dashboard-v1.html` (untracked; copy its CSS
and layout, replace its invented data). Python follows the conventions in
`~/.claude/AGENTS.md` (module-qualified imports, dataclasses for structured returns, typed
signatures, guard clauses).

### Collector contract

Each collector module exposes `collect() -> dict`, JSON-serialisable, shaped as
`{"zones": [...], ...tab-level fields}`. Per-zone work runs in a `ThreadPoolExecutor` and is
wrapped in one error boundary per zone that catches network and decode errors
(`urllib.error.URLError`, `TimeoutError`, `json.JSONDecodeError`, `KeyError`) and records
`{"chain_id": ..., "error": "<message>"}` for that zone, so one dead endpoint never blanks a
tab. Within a zone, an optional lookup that fails (a tx index that does not go back far
enough, a holder chain with no REST) yields `null` for that field, rendered as "n/a".

`server.py` wraps each snapshot as `{"fetched_at": <iso>, "duration_seconds": <float>, "refreshing": <bool>, "data": <collect()>}`.
Refresh intervals (config): channels 60s, funds 120s, validators 300s. `POST /api/refresh/<tab>`
triggers one now. The page shows the snapshot age per tab and a per-zone error badge.

### Endpoints

Private Polkachu endpoints for Stride and every in-scope zone, pattern
`https://<name>-strd-api.polkachu.com` and `https://<name>-strd-rpc.polkachu.com`, with
`<name>` ∈ stride, cosmos, celestia, dydx, haqq, injective, juno, band, osmosis, terra,
sommelier, saga. All requests send `User-Agent: curl/8.0` (Polkachu rejects the urllib default).
Holder-route chains (Secret, Agoric, Neutron, Carbon, Dymension, Axelar, Penumbra) use
`https://rest.cosmos.directory/<registry name>` and `https://rpc.cosmos.directory/<registry name>`;
Penumbra is not an SDK chain and is shown from the Osmosis side only. Each zone's endpoints
are one overridable entry in `config.py`.

### In-scope zones (config)

celestia, cosmoshub-4, dydx-mainnet-1, haqq_11235-1, injective-1, juno-1, laozi-mainnet,
osmosis-1, phoenix-1, sommelier-3, ssc-1. Deprecated zones are not shown. Per zone the config
holds: endpoint name, token symbol, decimals (18 for dydx, haqq, injective; 6 otherwise), and
the host-side transfer channel to Osmosis, copied from `HostToOsmosisTransferChannel`
(`x/stakeibc/types/wind_down.go` on the PR 4 branch): celestia channel-2, cosmoshub-4
channel-141, dydx-mainnet-1 channel-3, haqq_11235-1 channel-2, injective-1 channel-8, juno-1
channel-0, laozi-mainnet channel-83, phoenix-1 channel-1, sommelier-3 channel-0, ssc-1
channel-1, osmosis-1 none. The Osmosis-side id of each is read from the host channel's
counterparty. Everything else about a zone (ICA addresses, validators, denom, rate, Stride
transfer channel) is read live from `stakeibc/host_zone`.

Operator addresses: protocol admin `stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh`, gov
`stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl`, sweep operator
`stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9`, Osmosis vault
`osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af`.

## Tab 1: Channels

Three groups, as in the mock.

**Stride ↔ host zones.** Source of the channel list: `https://channels.main.stridenet.co/api/data`
(client, connection, transfer channel and ICA channels with both ends' state), filtered to the
in-scope zones, plus Stride ↔ Osmosis `channel-5` as the sweep channel (it is osmosis-1's
transfer channel in that feed). Rows: the transfer channel and each ICA channel.

Per channel:

- **State** on both ends (from the feed for ICAs; `ibc/core/channel/v1/channels/{id}/ports/{port}`
  for transfer channels).
- **Pending packets →** (Stride → host): Stride's `packet_commitments` for the channel, passed
  to the host's `unreceived_packets`. Unreceived = pending packets. Commitments that are
  received = **pending acks**. ICA channels only have this direction (host port `icahost`).
- **Pending packets ←** and their pending acks (transfer channels only): the same with the
  roles swapped.
- **Oldest pending**: lowest pending sequence; its age when `tx_search` on the sending chain
  finds the `send_packet` for that sequence, otherwise the sequence alone.
- **Last sent / last received / last ack**: newest tx from `tx_search` (desc, 1 result) for
  `send_packet.packet_src_channel` on the sender, `recv_packet.packet_dst_channel` on the
  receiver, and `acknowledge_packet.packet_src_channel` on the sender; block time from the
  tx's height. ICA packets sent by epoch hooks are not in the tx index (checked:
  `block_search` returns nothing on Polkachu), so for ICA channels "last sent" only reflects
  tx-initiated sends, which is what the admin txs are; last received and last ack are what
  show relayer activity. `null` when the index holds none.
- **Status** pill: `closed` / `handshake stuck` (state not OPEN on either end), `stuck`
  (a pending packet or ack known to be older than 30 minutes), `pending` (any pending,
  younger or of unknown age), else `ok`.

Busy transfer channels have many foreign commitments; commitments are paginated with a cap
of 1,000 per channel per direction and the count is shown as "1000+" beyond it.

Zone header row: client status and time to expiry on both chains (`client_status`, plus
latest consensus state timestamp + trusting period), connection state, and the worst status
of its channels.

**Host → Osmosis legs.** One row per in-scope zone except osmosis-1: host channel, Osmosis
channel, both clients' status and expiry, last received on Osmosis, last ack on the host. No
commitment counts (shared with all traffic); our own in-flight transfers are on the Funds tab.

**Holder routes into Osmosis.** A config list of `(chain, Osmosis channel, relayed by)` taken
from the relayer scope table in `docs/wind-down/sttoken-locations.md`: Cosmos Hub channel-0,
HAQQ channel-1575, Injective channel-122, Secret channel-88, Penumbra channel-79703, Agoric
channel-320, Neutron channel-874, Carbon channel-188, Terra channel-251, Dymension
channel-19774, Axelar channel-208, Celestia channel-6994, Saga channel-38946, Juno channel-42,
dYdX channel-6787, Band channel-148. Queried on the Osmosis side only: channel state, client
status and expiry, last received, last ack. Collapsed by default.

Tiles: channels open / total, pending packets, pending acks, oldest pending, clients live and
soonest expiry.

## Tab 2: Validators

Zone chips (each with a badge: count of validators recorded over the host amount, or `ok`),
summary tiles for the selected zone, and one row per validator sorted by |diff| descending.

Per zone: Stride's `host_zone.validators` (delegation, weight, `shares_to_tokens_rate`,
`slash_query_in_progress`, `delegation_changes_in_progress`) joined on validator address with
the host's `delegations/{delegation_ica}` (all pages), `validators` (moniker, status, jailed,
tokens / delegator_shares as the chain's rate) and `delegators/{ica}/unbonding_delegations`
(entry count per validator, of 7).

Row: moniker, operator address, weight %, recorded, actual, diff (recorded − actual), diff %,
rate difference, unbonding entries, in-progress flag, bond status. A validator on the host
with a delegation but missing from Stride is a row with recorded 0 and a `not registered`
badge. Severity: over (diff > 0) by any amount is amber, over by at least 0.0001% of the
validator's recorded delegation is red, since an undelegate for the recorded amount fails
when the host has less; under is neutral.

Tiles: validator count (and how many hold a delegation on the host), recorded sum, actual
sum, diff, count of in-progress flags.

A `staketia` chip shows the 5-of-7 multisig (`delegation_address` from `staketia/host_zone`):
its per-validator delegations on Celestia with no recorded column, and the total against
`remaining_delegated_balance`.

Amounts are exact integers in the payload (strings) plus the zone's decimals; the page formats.

## Tab 3: Funds flow

**Stage table**, one row per zone, native units, with a stacked bar over six stages:

| Stage | Definition |
|---|---|
| Staked | Delegation ICA delegations (+ staketia multisig delegations for celestia). |
| Unbonding | Delegation ICA unbonding entries (+ multisig's). Column beside it: earliest and latest completion time. |
| Liquid, pre-transfer | Host-denom bank balance of the delegation, withdrawal and fee ICAs (+ multisig liquid balance and the staketia claim address's TIA voucher on Stride for celestia). |
| In flight | Our ICA transfers to the vault whose packet is still committed on the host (below). |
| Osmosis vault | Vault balance of the zone's native denom on Osmosis (`ibc/` + sha256 of `transfer/<Osmosis-side channel>/<host denom>`; `uosmo` for osmosis-1). |
| In pools | Native balance of the zone's transmuter pools. |

Further columns: redemption ICA balance (owed to open user claims, not part of the bar),
needed = stToken bank supply on Stride × `redemption_rate`, and coverage =
(vault + pools native + stTokens held by the pools × rate) ÷ needed, since a swapped-in
stToken has already received its backing.

**In flight.** On the host RPC, `tx_search` for `ibc_transfer.sender='<ica>'` for each of the
four ICAs; each hit gives sequence, amount, denom and time from the tx events. A transfer is
in flight while the host still has a packet commitment for its sequence on the host → Osmosis
channel. The list (time, ICA, amount, status in flight / settled) is shown under the diagram.
osmosis-1 has no leg (bank send). If the index lookup fails the stage is `null`.

**Pools.** Discovered, not configured: every vault balance whose denom is
`factory/<contract>/alloyed/...` names a transmuter contract the vault joined; its bank
balances give the native and stToken amounts. A pool belongs to the zone whose native denom it
holds; it is canonical when its stToken denom is `ibc/` + sha256 of
`transfer/channel-326/<st denom>`, otherwise a route pool. `config.EXTRA_POOL_CONTRACTS`
adds pools the vault has not joined yet.

**Diagram** for the selected zone: inline SVG built from the payload, the mock's layout.
Host lane: Staked → Unbonding → the four ICAs; an arrow to the Osmosis lane (vault → canonical
pool, route pools) labelled with the host → Osmosis channel and the in-flight amount; the
redemption ICA's arrow exits to "user claims" with the zone's open user redemption record
count. Celestia adds the staketia lane: multisig staked → multisig unbonding + liquid → claim
address on Stride → delegation ICA. Arrows carry the admin tx that moves funds across them.

**Accounts table** for the selected zone: every address read (four ICAs, deposit address,
staketia addresses for celestia, vault, each pool, protocol admin, gov, sweep operator) with
chain, full address (click to copy), liquid, staked, unbonding, and a note. Non-host denoms
with a non-zero balance in an ICA (the USDC in the dYdX withdrawal ICA) are listed in the note.

## Error handling

Per-zone boundary and `null` fields as in the collector contract. A refresh that raises
anything else leaves the previous snapshot in place and logs the traceback to stderr. HTTP
timeout 20s per request, one retry. The page keeps rendering the last good snapshot and marks
it stale when older than three intervals.

## Testing

Unit tests on the pure functions, with small literal fixtures: IBC denom hash, pending
packet / pending ack split, channel status rule, drift row construction (including the
unregistered validator and the 18-decimal case), stage aggregation and coverage, pool
classification, client expiry. No test hits the network.

Acceptance is a live run: the server starts, all three `/api/*` snapshots complete with no
zone-level error for the eleven zones (or each error is explained by the host itself being
unreachable), haqq's validator diffs match `scripts/wind-down/drift/report.md` to the order of
magnitude, the closed celestia and haqq ICA channels show as closed, and each tab renders in
the browser.

## Out of scope

Deprecated zones, stakedym, alerting, history or charts over time, USD values, the token
sweep's progress, authentication, deployment.

## Build plan

Three chunks, run one after another in the `wind-down-dashboard` worktree (they share
`config.py`, `chain.py` and the page shell, so they are not parallel).

### Chunk 1: foundation and the Channels tab

Depends on: none.

- Files: `server.py`, `config.py`, `chain.py`, `channels.py`, `static/index.html`,
  `static/style.css`, `static/app.js`, `static/channels.js`, `test_chain.py`,
  `test_channels.py`, `README.md`.
- `config.py`: endpoints, the zone table, operator addresses, holder routes, intervals, port 8787.
- `chain.py` produces, for later chunks: `get_json(url)`, `rest_get(chain, path, params)`,
  `rest_get_all_pages(chain, path, key)`, `rpc_tx_search_latest(chain, query)` returning
  height and block time, `rpc_tx_search(chain, query)`, `ibc_denom(path)`, plus typed IBC
  helpers (channel end, packet commitments, unreceived packets, client status and expiry).
  A `Chain` dataclass (chain id, rest, rpc) is the handle passed around.
- `server.py`: the snapshot cache, a refresh thread per registered collector, static files,
  the two API routes. Collectors register in one dict so chunks 2 and 3 add one line each.
- Page shell: header with snapshot age and refresh button, tabs, and `app.js` exposing a
  small registry (`registerTab(name, render)`) plus shared formatters (amount with decimals,
  relative time, pill, truncated address with copy). `validators.js` and `funds.js` do not
  exist yet; their tabs show "not built".
- Key decisions: status rule and the 1,000-commitment cap as specified; the zone header's
  client expiry computed from consensus state timestamp + trusting period.
- Tests: denom hash against a known Osmosis denom, pending/ack split, status rule, expiry.

### Chunk 2: Validators tab

Depends on: chunk 1.

- Files: `validators.py`, `static/validators.js`, `test_validators.py`; one line each in
  `server.py` and `static/index.html`.
- Consumes `chain.py` helpers and `config.ZONES`. Produces `collect()` with, per zone,
  totals and a row list of exact-integer strings.
- Key decisions: join on validator address; unregistered host delegations become rows; the
  staketia pseudo-zone; severity thresholds as specified. `measure_delegation_drift.py` is
  the reference for the queries, not imported.
- Tests: row construction, totals, severity, unregistered validator, 18 decimals.

### Chunk 3: Funds flow tab

Depends on: chunk 1.

- Files: `funds.py`, `static/funds.js`, `test_funds.py`; one line each in `server.py` and
  `static/index.html`.
- Consumes `chain.py` helpers, `config.ZONES`, operator addresses. Produces `collect()` with,
  per zone, the account list, the six stage amounts, redemption ICA balance, needed, coverage,
  in-flight transfer list and pools; plus the operator accounts.
- Key decisions: pool discovery from alloyed denoms; in-flight by commitment lookup; coverage
  formula as specified; the diagram is an SVG string built in JS from the zone payload with
  fixed node coordinates from the mock, the staketia lane drawn only for celestia.
- Tests: stage aggregation, coverage, pool classification, in-flight status from a fixture.
