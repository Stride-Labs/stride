"""Route split: each policy route's share of its stToken supply, recorded through the wind-down so a move shows.

Each route pool is funded with its route's escrow x the pool's rate and the canonical pool with the rest, so the split
is fixed by Stride's escrow balances. Someone sending a route's stTokens back through Stride to Osmosis would move
escrow from the route to the canonical side after the pools are funded: the route pool would then hold more than it
can pay out and the canonical pool less. Every Pools refresh reads the stToken supply and every policy route's escrow
on Stride and appends any amount that changed to `ops/route_split.jsonl`. The Pools tab shows each zone's split now,
at the baseline (the first live reading) and at funding (from the vault's shares), and turns red once anything moved
by more than one whole stToken: against the funding once the pools are funded (a move before then was funded for), and
against the baseline until then. The split at the v35 upgrade is shown beside it for reference only:
the flushed redemptions burned stTokens and holders moved escrow before any pool was funded, so it is never red.

Backfill the upgrade reading once, from a node that still holds the height (Polkachu prunes it; the local v35
log-capture node keeps it):
    python3 scripts/wind-down/dashboard/route_split.py backfill --rest http://localhost:1317 --rpc http://localhost:26657
"""

import argparse
import collections
import concurrent.futures
import dataclasses
import datetime
import decimal
import json
import pathlib
import sys
import urllib.parse
import urllib.request
from dataclasses import dataclass
from decimal import Decimal
from typing import Any

import chain
import config
import pools

HISTORY_PATH = pathlib.Path(__file__).parent / "ops" / "route_split.jsonl"
SUPPLY_KEY = "supply"
CANONICAL_LABEL = "canonical"
LIVE_SOURCE = "live"
UPGRADE_SOURCE = "v35 upgrade"
V35_UPGRADE_HEIGHT = 41522514  # cosmos/upgrade applied_plan v35
HEIGHT_HEADER = "x-cosmos-block-height"
READ_WORKERS = 8
SHARE_PLACES = Decimal(
    "0.000001"
)  # a share as a fraction of the supply: four decimals of a percent


@dataclass(frozen=True)
class Reading:
    """One recorded amount: a stToken's supply, or its escrow on one Stride transfer channel."""

    at: str
    source: str  # LIVE_SOURCE, or UPGRADE_SOURCE for the backfilled reference
    height: int | None
    st_denom: str
    key: str  # SUPPLY_KEY or a Stride channel id
    amount: int


@dataclass(frozen=True)
class Supplies:
    """A stToken's supply now, at the baseline and at the upgrade: the denominators of every share."""

    now: int
    baseline: int
    at_upgrade: int | None


@dataclass(frozen=True)
class SplitRow:
    """One side of a zone's split: a policy route (its escrow) or the canonical remainder (supply less every route)."""

    label: str  # the holder chain and channel, or CANONICAL_LABEL
    channel: str | None
    pool_id: str | None
    now: int
    baseline: int  # the first live reading
    at_upgrade: (
        int | None
    )  # the v35 upgrade reading, for reference; None until backfilled
    at_funding: (
        int | None
    )  # the stTokens the pool was funded for; None until it is funded and marked
    share_now: str | None
    share_baseline: str | None
    share_at_upgrade: str | None
    share_at_funding: str | None
    moved: int  # the move since funding, or since the baseline before then, in stTokens (the upgrade does not count)
    drifted: bool
    matches_funding: bool | None  # now within one stToken of what the pool was funded for; None until funded


@dataclass(frozen=True)
class ZoneSplit:
    baseline_at: str
    supply_now: int
    supply_baseline: int
    supply_at_upgrade: int | None
    rows: list[SplitRow]
    drifted: bool  # a row or the supply moved by more than one whole stToken


# ---- public API


def collect() -> dict[str, Any]:
    """The Pools collector: the pools payload, each zone carrying its route split, after recording what changed."""
    payload = pools.collect()
    history = load_history()
    read = chain.within_error_boundary(
        chain_id=config.STRIDE_CHAIN_ID,
        work=lambda: read_amounts(rest=chain.stride_chain().rest, height=None),
    )
    if isinstance(read, chain.ZoneError):
        return annotate(pools_data=payload, history=history, read_error=read.error)

    changed = changes(history=history, readings=read)
    append_history(readings=changed)
    return annotate(pools_data=payload, history=history + changed, read_error=None)


def read_amounts(rest: str, height: int | None) -> list[Reading]:
    """Every policy stToken's supply and its escrow on each of its route channels, at `height` (None: the latest)."""
    targets = [
        (st_denom, key)
        for st_denom, channels in config.REQUIRED_ROUTES.items()
        for key in (SUPPLY_KEY, *sorted(channels, key=_channel_number))
    ]
    with concurrent.futures.ThreadPoolExecutor(max_workers=READ_WORKERS) as executor:
        amounts = list(
            executor.map(
                lambda target: _read_amount(
                    rest=rest, height=height, st_denom=target[0], key=target[1]
                ),
                targets,
            )
        )

    # A past read is stamped with its block's time by the caller; a live one with now.
    at = datetime.datetime.now(datetime.UTC).isoformat() if height is None else ""
    source = LIVE_SOURCE if height is None else UPGRADE_SOURCE
    return [
        Reading(
            at=at,
            source=source,
            height=height,
            st_denom=st_denom,
            key=key,
            amount=amount,
        )
        for (st_denom, key), amount in zip(targets, amounts, strict=True)
    ]


def changes(history: list[Reading], readings: list[Reading]) -> list[Reading]:
    """The readings whose amount differs from the last live one for the same stToken and key, or that are the first
    live one (the backfilled upgrade reading is never a live baseline)."""
    latest = {
        (reading.st_denom, reading.key): reading.amount
        for reading in history
        if reading.source == LIVE_SOURCE
    }
    return [
        reading
        for reading in readings
        if latest.get((reading.st_denom, reading.key)) != reading.amount
    ]


def annotate(
    pools_data: dict[str, Any], history: list[Reading], read_error: str | None
) -> dict[str, Any]:
    """The pools payload with `route_split` (a ZoneSplit, or None while nothing is recorded) and `route_split_error`
    on every readable zone."""
    by_denom: dict[str, list[Reading]] = collections.defaultdict(list)
    for reading in history:
        by_denom[reading.st_denom].append(reading)

    zones = [
        zone
        if "error" in zone
        else {
            **zone,
            "route_split": _stringified(
                split=zone_split(zone=zone, readings=by_denom.get(zone["st_denom"], []))
            ),
            "route_split_error": read_error,
        }
        for zone in pools_data["zones"]
    ]
    return {**pools_data, "zones": zones}


def zone_split(zone: dict[str, Any], readings: list[Reading]) -> ZoneSplit | None:
    """A zone's split from its recorded readings: the first live one per key is the baseline, the last is now."""
    first: dict[str, Reading] = {}
    last: dict[str, Reading] = {}
    upgrade: dict[str, Reading] = {}
    for reading in readings:
        if reading.source == UPGRADE_SOURCE:
            upgrade[reading.key] = reading
            continue
        first.setdefault(reading.key, reading)
        last[reading.key] = reading
    if SUPPLY_KEY not in last:
        return None

    supplies = Supplies(
        now=last[SUPPLY_KEY].amount,
        baseline=first[SUPPLY_KEY].amount,
        at_upgrade=upgrade[SUPPLY_KEY].amount if SUPPLY_KEY in upgrade else None,
    )
    threshold = 10 ** int(zone["decimals"])  # one whole stToken

    # Each route, then the canonical remainder: what the supply holds beyond every route's escrow.
    routes = [
        _route_row(
            zone=zone,
            channel=channel,
            first=first,
            last=last,
            upgrade=upgrade,
            supplies=supplies,
            threshold=threshold,
        )
        for channel in sorted(
            config.REQUIRED_ROUTES[zone["st_denom"]], key=_channel_number
        )
    ]
    rows = [
        *routes,
        _canonical_row(
            zone=zone, routes=routes, supplies=supplies, threshold=threshold
        ),
    ]
    funded = any(row.at_funding is not None for row in rows)
    return ZoneSplit(
        baseline_at=first[SUPPLY_KEY].at,
        supply_now=supplies.now,
        supply_baseline=supplies.baseline,
        supply_at_upgrade=supplies.at_upgrade,
        rows=rows,
        # A supply change moves the canonical row, which compares against funding once funded.
        drifted=any(row.drifted for row in rows)
        or (not funded and abs(supplies.now - supplies.baseline) > threshold),
    )


def load_history(path: pathlib.Path = HISTORY_PATH) -> list[Reading]:
    if not path.exists():
        return []
    return [
        Reading(**json.loads(line))
        for line in path.read_text().splitlines()
        if line.strip()
    ]


def append_history(readings: list[Reading], path: pathlib.Path = HISTORY_PATH) -> None:
    if not readings:
        return
    with path.open("a") as history:
        history.writelines(
            json.dumps(dataclasses.asdict(reading)) + "\n" for reading in readings
        )


# ---- rows


def _route_row(
    zone: dict[str, Any],
    channel: str,
    first: dict[str, Reading],
    last: dict[str, Reading],
    upgrade: dict[str, Reading],
    supplies: Supplies,
    threshold: int,
) -> SplitRow:
    pool = next(
        (
            pool
            for pool in zone["pools"]
            if pool.get("route") and pool["route"]["stride_channel"] == channel
        ),
        None,
    )
    return _row(
        label=f"{_holder_chain(zone=zone, channel=channel, pool=pool)} · {channel}",
        channel=channel,
        pool=pool,
        now=last[channel].amount if channel in last else 0,
        baseline=first[channel].amount if channel in first else 0,
        at_upgrade=upgrade[channel].amount if channel in upgrade else None,
        at_funding=_funded_escrow(pool=pool) if pool else None,
        supplies=supplies,
        threshold=threshold,
    )


def _canonical_row(
    zone: dict[str, Any], routes: list[SplitRow], supplies: Supplies, threshold: int
) -> SplitRow:
    """The supply beyond every route's escrow. Its at-funding side mirrors the routes' (the supply less what each route
    pool was funded for), not its own shares: the canonical pool also holds the zone's surplus, which is no move."""
    pool = next(
        (pool for pool in zone["pools"] if pool["kind"] == pools.PoolKind.CANONICAL),
        None,
    )
    routes_funded = bool(routes) and all(row.at_funding is not None for row in routes)
    upgrade_known = supplies.at_upgrade is not None and all(
        row.at_upgrade is not None for row in routes
    )
    return _row(
        label=CANONICAL_LABEL,
        channel=None,
        pool=pool,
        now=supplies.now - sum(row.now for row in routes),
        baseline=supplies.baseline - sum(row.baseline for row in routes),
        at_upgrade=supplies.at_upgrade - sum(row.at_upgrade for row in routes)
        if upgrade_known
        else None,
        at_funding=supplies.now - sum(row.at_funding for row in routes)
        if routes_funded
        else None,
        supplies=supplies,
        threshold=threshold,
    )


def _row(
    label: str,
    channel: str | None,
    pool: dict[str, Any] | None,
    now: int,
    baseline: int,
    at_upgrade: int | None,
    at_funding: int | None,
    supplies: Supplies,
    threshold: int,
) -> SplitRow:
    # Funding read the live split, so a move before it is already paid for: once funded, only a move since counts.
    moved = abs(now - baseline) if at_funding is None else abs(now - at_funding)
    return SplitRow(
        label=label,
        channel=channel,
        pool_id=pool["pool_id"] if pool else None,
        now=now,
        baseline=baseline,
        at_upgrade=at_upgrade,
        at_funding=at_funding,
        share_now=_share(amount=now, supply=supplies.now),
        share_baseline=_share(amount=baseline, supply=supplies.baseline),
        share_at_upgrade=None
        if at_upgrade is None
        else _share(amount=at_upgrade, supply=supplies.at_upgrade),
        share_at_funding=None
        if at_funding is None
        else _share(amount=at_funding, supply=supplies.now),
        moved=moved,
        drifted=moved > threshold,
        matches_funding=None if at_funding is None else abs(now - at_funding) <= threshold,
    )


def _funded_escrow(pool: dict[str, Any]) -> int | None:
    """The stTokens a funded pool covers: its value (every share, whoever joined) over its rate. Only once marked,
    since a pool still being funded is expected to differ."""
    if not pool["native_marked"] or pool["rate"] is None or not pool["alloyed_supply"]:
        return None
    with decimal.localcontext(pools.DECIMAL_CONTEXT):
        return int(Decimal(int(pool["alloyed_supply"])) / Decimal(pool["rate"]))


def _holder_chain(
    zone: dict[str, Any], channel: str, pool: dict[str, Any] | None
) -> str:
    """The chain the route's holders are on: from its live pool, else its planned pool, else unknown."""
    if pool:
        return pool["route"]["chain_id"]
    planned = next(
        (
            plan
            for plan in zone.get("planned", [])
            if plan.get("stride_channel") == channel
        ),
        None,
    )
    return (planned or {}).get("holder_chain_id") or "unknown chain"


def _share(amount: int, supply: int | None) -> str | None:
    if not supply:
        return None
    return str((Decimal(amount) / Decimal(supply)).quantize(SHARE_PLACES))


def _stringified(split: ZoneSplit | None) -> dict[str, Any] | None:
    return None if split is None else chain.stringify_ints(dataclasses.asdict(split))


# ---- reads


def _read_amount(rest: str, height: int | None, st_denom: str, key: str) -> int:
    if key == SUPPLY_KEY:
        body = _get(
            rest=rest,
            path="/cosmos/bank/v1beta1/supply/by_denom",
            params={"denom": st_denom},
            height=height,
        )
        return int(body["amount"]["amount"])
    address = chain.escrow_address(channel_id=key)
    body = _get(
        rest=rest,
        path=f"/cosmos/bank/v1beta1/balances/{address}/by_denom",
        params={"denom": st_denom},
        height=height,
    )
    return int(body["balance"]["amount"])


def _get(rest: str, path: str, params: dict[str, str], height: int | None) -> Any:
    """chain.get_json, or a read at `height` with the height header (chain.get_json sends no headers)."""
    url = f"{rest}{path}?{urllib.parse.urlencode(params)}"
    if height is None:
        return chain.get_json(url=url)
    request = urllib.request.Request(
        url=url, headers={HEIGHT_HEADER: str(height), "User-Agent": "curl/8.0"}
    )
    with urllib.request.urlopen(request, timeout=30) as response:
        return json.loads(response.read())


def _channel_number(channel_id: str) -> int:
    return int(channel_id.removeprefix("channel-"))


# ---- backfill


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Backfill the route split's reference reading at the v35 upgrade."
    )
    subcommands = parser.add_subparsers(dest="command", required=True)
    backfill = subcommands.add_parser("backfill")
    backfill.add_argument(
        "--rest", required=True, help="a Stride REST that still holds the height"
    )
    backfill.add_argument(
        "--rpc", required=True, help="the same node's RPC, for the block's time"
    )
    backfill.add_argument("--height", type=int, default=V35_UPGRADE_HEIGHT)
    args = parser.parse_args()
    if any(reading.source == UPGRADE_SOURCE for reading in load_history()):
        sys.exit(f"{HISTORY_PATH} already has the upgrade readings")

    print(f"Reading every policy route at height {args.height}...", file=sys.stderr)
    block_time = chain.block_time(rpc=args.rpc, height=args.height)
    readings = [
        dataclasses.replace(reading, at=block_time)
        for reading in read_amounts(rest=args.rest, height=args.height)
    ]
    append_history(readings=readings)
    print(f"Wrote {len(readings)} readings to {HISTORY_PATH}", file=sys.stderr)


if __name__ == "__main__":
    main()
