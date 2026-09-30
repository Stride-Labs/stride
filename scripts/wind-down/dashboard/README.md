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

Adding a tab: write `<tab>.py` with `collect() -> dict` (`{"zones": [...], ...}`), add one line to `COLLECTORS`, write
`static/<tab>.js` calling `registerTab(name, render)`, and uncomment its `<script>` line in `static/index.html`.

## Channels tab

- Stride <-> host zones: from `https://channels.main.stridenet.co/api/data` (in-scope zones, including osmosis-1's transfer
  channel-5 and its ICAs), with the transfer channel's ends and every channel's packet commitments read from Stride and
  the host. Every commitment is a pending packet (the receiver's `unreceived_packets` lists it) or a pending ack (it does
  not). Commitments are read up to 1,000 per channel per direction; the page shows `N+` beyond the cap.
- Status: `closed` / `handshake stuck` (an end not OPEN) / `stuck` (oldest pending known to be over 30 minutes old) /
  `pending` / `ok`. The age comes from `tx_search` for the `send_packet` of the lowest pending sequence.
- Host -> Osmosis legs: one row per zone except osmosis-1, from the host channel in `config.ZONES`.
- Holder routes into Osmosis: `config.HOLDER_ROUTES`, queried on the Osmosis side only. Collapsed by default.

## Known gaps

- "Last sent" only reflects txs: ICA packets sent by epoch hooks are not in the tx index (`block_search` returns nothing
  on Polkachu). Last received and last ack are what show relayer activity.
- A null (`n/a`) means the lookup failed or the index holds nothing. `tx_search` is tried over the last 100,000 blocks
  first because it times out on Osmosis for busy channels, then over the whole index.
- A send time is usually not in the index for old commitments (Cosmos Hub channel-0 holds ~1,000 very old unreceived
  packets), so those show as `pending` with the sequence only, never `stuck`.
- If a host's REST is down the zone still renders from Stride and the feed (host-side cells `n/a`, a "host REST
  unreachable" badge); its host -> Osmosis leg shows as an error row.
