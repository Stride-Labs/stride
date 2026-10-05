# Wind-down dashboard

A local, read-only page for the v35 wind-down ops window. It never signs or submits anything.

    python3 scripts/wind-down/dashboard/server.py        # then open http://localhost:8787

Python 3.12 standard library only; the page is vanilla JS with no build step. Tests (no network):

    python3 -m unittest discover -s scripts/wind-down/dashboard

## How it works

`server.py` owns one snapshot per registered collector (the `COLLECTORS` dict) and refreshes each in a background thread on
its interval from `config.REFRESH_INTERVAL_SECONDS` (channels 60s, funds 120s, validators 300s). The page polls
`GET /api/<tab>`, which returns `{fetched_at, duration_seconds, refreshing, data}` (HTTP 503 `{loading: true}` until the
first snapshot). `POST /api/refresh/<tab>` refreshes now; `GET /api/config` returns the intervals.

A refresh that raises anything other than a per-zone network or decode error keeps the previous snapshot and logs the
traceback to stderr. The page marks a snapshot stale after three intervals.

Adding a tab (the Ops tab is the exception, see below): write `<tab>.py` with `collect() -> dict` (`{"zones": [...], ...}`), add one line to `COLLECTORS`, write
`static/<tab>.js` calling `registerTab(name, render)`, and uncomment its `<script>` line in `static/index.html`.

## Ops tab

The first tab and the default (`#ops`): the dated wind-down checklist, one collapsible section per block, with the current
block and the next one open. It is not a collector: it has no snapshot, stale badge or refresh button.

- `ops/plan.json` is the plan (anchors, then `days` of blocks with `windows` of `steps`; steps with `zones` get one
  sub-tick per zone, ids `<id>:<zone>`). It is reviewed like the spec and read from disk on every request, so an edit
  shows on reload.
- `ops/status.json` holds the ticks: `{"<id>": {"done": true, "at": "<iso utc>", "by": "<name>"}}`.
- Routes: `GET /api/ops` returns `{plan, status, today}` (today is the UTC date); `POST /api/ops/check` with
  `{"id", "done", "by"}` records or (when `done` is false) deletes a tick and returns the full status. 400 for a bad body
  or an id that is not in the plan.

To tick, type your name in the "you are" field (kept in your browser) and click the checkbox. The server rewrites
`status.json` atomically with sorted keys, so each tick is a one-line diff. Ticks are committed like any other file:
commit and push `ops/status.json` so the team sees the same state. Logic lives in `ops.py`; `server.py` only routes.

## Channels tab

- Stride <-> host zones: from `https://channels.main.stridenet.co/api/data` (in-scope zones, including osmosis-1's transfer
  channel-5 and its ICAs), with the transfer channel's ends and every channel's packet commitments read from Stride and
  the host. Every commitment is a pending packet (the receiver's `unreceived_packets` lists it) or a pending ack (it does
  not). Commitments are read up to 1,000 per channel per direction; the page shows `N+` beyond the cap.
- Status: `closed` / `handshake stuck` (an end not OPEN) / `stuck` (oldest pending known to be over 30 minutes old) /
  `pending` / `ok`. The age comes from `tx_search` for the `send_packet` of the lowest pending sequence; when that is
  not indexed and the sequence is a pending ack, it comes from the `recv_packet` on the receiver instead (the cell reads
  `received 5h ago`, otherwise `sent 45m ago`).
- Host -> Osmosis legs: one row per zone except osmosis-1, from the host channel in `config.ZONES`.
- Holder routes into Osmosis: `config.HOLDER_ROUTES`, queried on the Osmosis side only. Collapsed by default.
- Pre-flight checks (last panel, `zone.preflight`): one row per zone with an `ok` / `fail` / `n/a` pill per check, and
  "N of 11 zones pass every check" in the header. Withdraw address: the host's
  `distribution/delegators/{delegation ICA}/withdraw_address` equals the withdrawal ICA (a mismatch shows both
  addresses). ICA host allow list: `interchain_accounts/host/v1/params` allows `MsgTransfer` (or `*`); osmosis-1 also
  needs `bank MsgSend`. Host -> Osmosis leg: the host's Osmosis channel is OPEN and its client tracks `osmosis-1`
  (osmosis-1 itself has no leg: `n/a`, "bank send", and that does not fail the row). Host client of Stride: the
  counterparty client of the ICA connection is Active (Band's is expected to be Expired). Stride's host zones are read
  once per refresh; a failed lookup makes only its own pill `n/a`, which does not count as a pass.

## Validators tab

- Per zone, Stride's recorded delegation per validator (`host_zone.validators`, joined on validator address) against what
  the host chain holds: the `balance.amount` of the host's `delegations/{delegation_ica}`. `validators` supplies moniker,
  status, jailed and the rate (tokens / delegator_shares); `unbonding_delegations` supplies entry counts per validator,
  out of a maximum of 7.
- A host delegation Stride has no entry for gets an unregistered row: recorded 0 and a `not registered` badge.
- Severity: the host holding less than recorded is neutral; over by any amount is amber; over by at least 0.0001% of the
  recorded delegation is red. Over by no more than the drain's rounding buffer (recorded ÷ 1e17, at least one base unit)
  on a validator whose stored rate is below 1 is `buffered`: shown muted and not counted as over, because
  `applySharesRoundingSafety` shaves that much off a full drain, so the overage cannot fail it. The buffer does not
  exist at a rate of exactly 1, where any overage counts.
- The staketia chip compares the multisig's delegations on Celestia with Stride's `remaining_delegated_balance`.
- Live test: each row carries a `live test` pill on the zone's pick and a muted `next` pill on the runner-up, the zone
  summary and the single-zone tiles name the pick. Same rule as `pick_live_test_validators.py`: the smallest recorded
  delegation of at least one whole token (10^decimals) with no unbonding entry in flight, excluding unregistered
  validators and any with a delegation change or slash query in progress. When the unbonding lookup failed, or no
  validator qualifies (for example, every funded validator already has an entry), the pick is null with
  `live_test_reason`. The script falls back to the smallest overall in that case; the dashboard does not.
- Delegations are the one required lookup (a failure gives the zone an error row). The validator list and the unbonding
  entries are optional: if they fail, those cells show `n/a` and the row falls back to Stride's name.

## Funds flow tab

- Stage table, one row per zone: staked (delegation ICA delegations), unbonding (its unbonding entries, with the
  earliest maturity), liquid (delegation + withdrawal + fee ICA balances in the host denom), in flight (our ICA
  transfers to the Osmosis vault whose packet commitment the host still holds), Osmosis vault balance, and the
  native balance of the zone's transmuter pools. Celestia adds the staketia multisig's delegations, unbonding and
  liquid balance and the claim address's TIA voucher on Stride. The redemption ICA is shown beside the bar, not in it.
- Needed = stToken bank supply on Stride x `redemption_rate`; coverage = (vault + pools' native + stTokens swapped
  into the pools x rate) / needed. The pill is `ok` at 100%, `bad` when nothing is left upstream and it is short.
- In flight: `tx_search` on the host for `ibc_transfer.sender='<ica>' AND ibc_transfer.receiver='<vault>'` (the 100
  newest hits per ICA), sequence / amount / denom decoded from each `send_packet`, status from one `unreceived_acks`
  query for the sequences. `n/a` when the index lookup fails.
- Pools: every `factory/<contract>/alloyed/...` denom in the vault names a transmuter the vault joined, plus
  `config.EXTRA_POOL_CONTRACTS`. A pool's asset list (`list_asset_configs`) says which stToken it holds even before a
  swap; its bank balances give the amounts. It belongs to the zone whose native Osmosis denom it holds, or, once
  drained of native, whose stToken it holds (by denom trace); canonical when the stToken came over channel-326.
- Click a row for the zone's diagram, its transfer list and its accounts (ICAs, deposit address, staketia addresses
  for celestia, vault, pools, operators). Every integer in the payload is a string; the page uses BigInt.
- Records column: `epoch unbonding entries not yet CLAIMABLE / user redemption records` for the zone (zero is muted).
  Both come from Stride, read once per refresh across all pages: `records/epoch_unbonding_record` (entries with
  `native_token_amount > 0`) and `records/user_redemption_record`. Celestia also reads the staketia redemption and
  unbonding records. The payload's `records` object carries the counts and the `delegation_transfer_ready`,
  `redemption_transfer_ready` and (celestia) `staketia_claim_ready` booleans.
- Transfer checklist panel (selected zone, above the ICA transfers): DELEGATION transfer is ready when nothing is
  outside CLAIMABLE (otherwise the count by status), REDEMPTION transfer when there are no user redemption records and
  no pending claims, and for celestia the staketia claim balance when no staketia redemption record and no unbonding
  record outside CLAIMED remains. `n/a` when a record table could not be read.
- Click the Staked or Unbonding node in the diagram for the per-validator breakdown (`validator_positions`): every
  validator the delegation ICA, and for celestia the multisig, has stake or unbonding entries on, with a bar per
  validator split into its staked amount and one segment per unbonding entry (hover for the amount and completion).

## Known gaps

- "Last sent" only reflects txs: ICA packets sent by epoch hooks are not in the tx index (`block_search` returns nothing
  on Polkachu). Last received and last ack are what show relayer activity.
- A null (`n/a`) means the lookup failed or the index holds nothing. `tx_search` is tried over the last 100,000 blocks
  first because it times out on Osmosis for busy channels, then over the whole index.
- A send time is usually not in the index for old commitments (Cosmos Hub channel-0 holds ~1,000 very old unreceived
  packets) or for epoch-hook ICA sends. A pending ack then ages from its receive time and can become `stuck`; a pending
  packet has no receive tx, so one sent by a hook stays unknown-age and shows as `pending` with the sequence only.
- If a host's REST is down the zone still renders from Stride and the feed (host-side cells `n/a`, a "host REST
  unreachable" badge); its host -> Osmosis leg shows as an error row.
- On a channel with more than 1,000 commitments the "oldest pending" sequence is the lowest among the first commitments
  the store returns (keyed by decimal string, so lexicographic), not guaranteed to be the true minimum.
- Funds: if Osmosis cannot be read, vault, pools and coverage are `n/a` for every zone; a timed-out ICA transfer
  shows as `settled` (its commitment is gone) although the tokens were refunded to the ICA, where they show as liquid.
