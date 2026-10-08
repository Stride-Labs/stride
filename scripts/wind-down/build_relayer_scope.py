#!/usr/bin/env python3
"""Decide, per chain, which relayer legs the wind-down gets for free and which it must run.

Takes every chain from the per-token escrow tables in docs/wind-down/sttoken-locations.md, joins
the routes and light-client ages embedded in docs/wind-down/relayer-map.html where a route exists,
adds the last real packet seen on each leg (tx_search on the Stride and Osmosis RPCs, since a fresh
client header only proves someone updates the client, not that they relay our channel), and
applies the scope rule:

- Stride leg: hosts get it for free (our ICA relayers); nobody else gets one after the upgrade.
- Osmosis leg: "free" when a packet crossed it within FREE_PACKET_MAX_AGE_DAYS and the client is
  not expired (a recent packet outranks the map's stale label); otherwise "ops" when the chain's in-scope value is at least MIN_USD_FOR_OPS_RELAYER,
  else "none" (holders self-relay or move before the upgrade).
- Pool routes: every token worth at least SMALL_TOKEN_USD on a chain whose Osmosis leg is served.

The result is spliced into docs/wind-down/sttoken-locations.md as the "Relayer scope per chain"
section, above the existing per-token tables, between the relayer-scope markers.

    python3 scripts/wind-down/build_relayer_scope.py            # queries the RPCs, updates the doc
    python3 scripts/wind-down/build_relayer_scope.py --offline  # reuse the cached packet lookups
"""

import argparse
import dataclasses
import datetime
import json
import pathlib
import re
import sys
import time
import urllib.parse
import urllib.request

# Stride Labs' private Polkachu endpoints (public fallbacks: stride-rpc / osmosis-rpc.polkachu.com)
STRIDE_RPC = "https://stride-strd-rpc.polkachu.com"
OSMOSIS_RPC = "https://osmosis-strd-rpc.polkachu.com"
USER_AGENT = "curl/8.0"

FREE_PACKET_MAX_AGE_DAYS = 7.0
MIN_USD_FOR_OPS_RELAYER = 1_000
SMALL_TOKEN_USD = 500
REQUEST_PAUSE_SECONDS = 0.6
# Chains dropped by decision whatever their packet ages say: chain id -> the reason shown in the table
DROPPED_CHAINS: dict[str, str] = {
    "axelar-dojo-1": "unsupported, Stride <-> Axelar clients expired (decided 2026-10-08)",
}

REPO = pathlib.Path(__file__).resolve().parents[2]
RELAYER_MAP = REPO / "docs" / "wind-down" / "relayer-map.html"
LOCATIONS_DOC = REPO / "docs" / "wind-down" / "sttoken-locations.md"
CACHE = REPO / "scripts" / "wind-down" / "relayer_scope_cache.json"
SECTION_START = "<!-- relayer-scope:start -->"
SECTION_END = "<!-- relayer-scope:end -->"
INSERT_BEFORE = "## Summary"

STRIDE_CHAIN_ID = "stride-1"
OSMOSIS_CHAIN_ID = "osmosis-1"
STRIDE_PHASE = "p1"
OSMOSIS_PHASE = "p2"


@dataclasses.dataclass
class Leg:
    """One channel between a chain and Stride or Osmosis, with the youngest packet a relayer delivered on it."""

    label: str
    near_channel: str  # the channel id on the chain we query (Stride or Osmosis)
    far_channel: str
    client_status: str
    client_age_days: float | None
    last_recv_days: float | None = None  # youngest recv_packet delivered to the near chain (chain → near)
    last_ack_days: float | None = None  # youngest acknowledge_packet delivered to the near chain (near → chain landed)


@dataclasses.dataclass
class ChainScope:
    chain_id: str
    name: str
    host: bool
    usd: int  # every stToken on the chain, whatever its status in the locations doc
    tokens: list[dict]  # {"sym", "usd", "status"}
    statuses: set[str]
    stride_channels: str
    stride_leg: Leg | None
    osmosis_leg: Leg | None
    osmosis_status: str
    stride_decision: str = ""
    osmosis_decision: str = ""
    pool_routes: list[str] = dataclasses.field(default_factory=list)

    @property
    def dead(self) -> bool:
        return self.statuses <= {"ignored · unrecoverable"}

    @property
    def deprecated_only(self) -> bool:
        return self.statuses <= {"ignored · deprecated"}

    @property
    def unsupported(self) -> bool:
        return self.statuses <= {"ignored · unsupported"}


def main() -> None:
    args = parse_args()
    phases = load_phases()
    chains = build_chains(phases=phases, escrow=load_escrow_tables())

    cache = json.loads(CACHE.read_text()) if CACHE.exists() else {}
    for chain in chains:
        fill_last_packets(chain=chain, cache=cache, offline=args.offline)
    CACHE.write_text(json.dumps(cache, indent=1, sort_keys=True))

    for chain in chains:
        decide(chain)
    splice_section(doc=LOCATIONS_DOC, section=render(chains))
    update_relayer_map(phases=phases, chains=chains)
    print(f"updated {LOCATIONS_DOC.relative_to(REPO)} and {RELAYER_MAP.relative_to(REPO)} for {len(chains)} chains")
    print(render_result(chain_count=len(chains)))


def render_result(chain_count: int) -> str:
    """The one-line verdict an operator reads last."""
    return (
        f"RESULT: DONE — docs/wind-down/sttoken-locations.md and relayer-map.html updated for {chain_count} chains; "
        "review the 'Relayer scope per chain' table"
    )


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--offline", action="store_true", help="do not query RPCs; use the cache only")
    return parser.parse_args()


# --- inputs -------------------------------------------------------------------------------------


def load_phases() -> dict[str, dict]:
    """The PHASES array embedded in the relayer map page, keyed by phase key."""
    for line in RELAYER_MAP.read_text().splitlines():
        match = re.match(r"\s*const PHASES\s*=\s*(\[.*\]);?\s*$", line)
        if match:
            return {phase["key"]: phase for phase in json.loads(match.group(1))}
    raise SystemExit(f"no PHASES array in {RELAYER_MAP}")


ROW_PATTERN = re.compile(r"^\| (?P<name>[^|]+?) \| (?P<chain_id>[^|]+?) \| (?P<channels>[^|]+?) \| (?P<amount>[\d,.]+) \| (?P<pct>[\d.]+)% \| \$(?P<usd>[\d,]+) \| (?P<status>[^|]+?) \|$")
TOKEN_HEADER = re.compile(r"^## (?P<sym>st[A-Z]+) \(")


def load_escrow_tables() -> dict[str, dict]:
    """Every chain row of every per-token table in the locations doc, keyed by chain id."""
    chains: dict[str, dict] = {}
    token = None
    for line in LOCATIONS_DOC.read_text().splitlines():
        header = TOKEN_HEADER.match(line)
        if header:
            token = header.group("sym")
            continue
        row = ROW_PATTERN.match(line)
        if token is None or row is None or row.group("chain_id") in ("Chain id", STRIDE_CHAIN_ID):
            continue
        chain = chains.setdefault(row.group("chain_id"), {"name": row.group("name"), "channels": row.group("channels"), "tokens": [], "statuses": set()})
        chain["tokens"].append({
            "sym": token,
            "amount": float(row.group("amount").replace(",", "")),
            "pct": float(row.group("pct")),
            "usd": int(row.group("usd").replace(",", "")),
            "status": row.group("status"),
        })
        chain["statuses"].add(row.group("status"))
    return chains


def build_chains(phases: dict[str, dict], escrow: dict[str, dict]) -> list[ChainScope]:
    """One ChainScope per chain that holds any stToken, largest first, with map legs where a route exists."""
    stride_routes = {route["id"]: route for route in phases[STRIDE_PHASE]["routes"]}
    osmosis_routes = {route["id"]: route for route in phases[OSMOSIS_PHASE]["routes"]}
    host_ids = {cid for cid, route in {**stride_routes, **osmosis_routes}.items() if route.get("host")}
    chains = []
    for chain_id, data in escrow.items():
        route = osmosis_routes.get(chain_id) or stride_routes.get(chain_id) or {}
        chains.append(ChainScope(
            chain_id=chain_id,
            name=data["name"],
            host=chain_id in host_ids,
            usd=sum(token["usd"] for token in data["tokens"]),
            tokens=data["tokens"],
            statuses=data["statuses"],
            stride_channels=data["channels"],
            stride_leg=leg_toward(stride_routes.get(chain_id), center="Stride"),
            osmosis_leg=leg_toward(osmosis_routes.get(chain_id), center="Osmosis"),
            osmosis_status=route.get("status", "unmapped"),
        ))
    return sorted(chains, key=lambda chain: -chain.usd)


def leg_toward(route: dict | None, center: str) -> Leg | None:
    """The chain → center transfer leg; the near channel is the one on the center chain (the destination)."""
    if route is None:
        return None
    for leg in route["legs"]:
        if leg["dir"].endswith(f"→ {center}") and "channel-" in leg["channel"]:
            far, near = [part.strip() for part in leg["channel"].split("→")]
            return Leg(label=leg["dir"], near_channel=near, far_channel=far, client_status=leg["status"], client_age_days=leg.get("age"))
    return None


# --- last packet per leg ------------------------------------------------------------------------


def fill_last_packets(chain: ChainScope, cache: dict, offline: bool) -> None:
    for leg, rpc in ((chain.stride_leg, STRIDE_RPC), (chain.osmosis_leg, OSMOSIS_RPC)):
        if leg is None:
            continue
        leg.last_recv_days = last_packet_age_days(rpc, "recv_packet.packet_dst_channel", leg.near_channel, cache, offline)
        leg.last_ack_days = last_packet_age_days(rpc, "acknowledge_packet.packet_src_channel", leg.near_channel, cache, offline)


def last_packet_age_days(rpc: str, event_key: str, channel: str, cache: dict, offline: bool) -> float | None:
    """Age in days of the youngest tx on `rpc` carrying `event_key = channel`, or None if there is none."""
    cache_key = f"{rpc}|{event_key}|{channel}"
    if cache_key in cache:
        return age_days(cache[cache_key])
    if offline:
        return None

    query = urllib.parse.quote(f'"{event_key}=\'{channel}\'"')
    result = rpc_get(f"{rpc}/tx_search?query={query}&order_by=%22desc%22&per_page=1")
    txs = result.get("result", {}).get("txs", [])
    if not txs:
        cache[cache_key] = None
        return None
    block = rpc_get(f"{rpc}/block?height={txs[0]['height']}")
    cache[cache_key] = block["result"]["block"]["header"]["time"]
    return age_days(cache[cache_key])


def rpc_get(url: str) -> dict:
    time.sleep(REQUEST_PAUSE_SECONDS)
    for attempt in range(5):
        try:
            request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
            with urllib.request.urlopen(request, timeout=60) as response:
                return json.loads(response.read())
        except Exception as error:  # noqa: BLE001 - public RPCs rate limit; back off and retry
            print(f"retry {attempt + 1} for {url[:90]}: {error}")
            time.sleep(10 * (attempt + 1))
    raise SystemExit(f"gave up on {url}")


def age_days(timestamp: str | None) -> float | None:
    if timestamp is None:
        return None
    # CometBFT emits RFC3339 with nanoseconds; trim to microseconds for fromisoformat
    trimmed = re.sub(r"(\.\d{6})\d*", r"\1", timestamp).replace("Z", "+00:00")
    then = datetime.datetime.fromisoformat(trimmed)
    return (datetime.datetime.now(datetime.timezone.utc) - then).total_seconds() / 86_400


# --- the rule ----------------------------------------------------------------------------------


def decide(chain: ChainScope) -> None:
    if chain.chain_id == OSMOSIS_CHAIN_ID:
        chain.stride_decision = "sweep channel, we relay it"
    else:
        chain.stride_decision = "ICA channel, we relay it" if chain.host else "not needed after the upgrade"

    leg = chain.osmosis_leg
    if chain.chain_id == OSMOSIS_CHAIN_ID:
        chain.osmosis_decision = "destination, the pools live here"
    elif chain.dead:
        chain.osmosis_decision = "not served: chain dead"
    elif chain.chain_id in DROPPED_CHAINS:
        chain.osmosis_decision = f"not served: {DROPPED_CHAINS[chain.chain_id]}"
    elif chain.unsupported:
        chain.osmosis_decision = "not served: unsupported (decided 2026-09-30)"
    elif chain.deprecated_only:
        chain.osmosis_decision = "not served: deprecated zone"
    elif leg is None and chain.usd < MIN_USD_FOR_OPS_RELAYER:
        chain.osmosis_decision = "not served: no relayer"
    elif leg is None:
        chain.osmosis_decision = "we relay it (leg not mapped yet)"
    elif chain.osmosis_status == "blocked":
        # Served: the Osmosis team is fixing the rate limiter that rejects these packets
        chain.osmosis_decision = "free once Osmosis fixes its rate limiter (spec §12)"
    elif leg.client_status != "expired" and leg.last_recv_days is not None and leg.last_recv_days <= FREE_PACKET_MAX_AGE_DAYS:
        # A recent packet beats the map's stale label: it proves a relayer is working the channel
        chain.osmosis_decision = "free, someone else relays it"
    elif chain.usd >= MIN_USD_FOR_OPS_RELAYER:
        chain.osmosis_decision = "we relay it, after client recovery" if leg.client_status == "expired" else "we relay it"
    else:
        chain.osmosis_decision = "not served: no relayer"

    served = not chain.osmosis_decision.startswith("not served")
    chain.pool_routes = [token["sym"] for token in chain.tokens if served and token["usd"] >= SMALL_TOKEN_USD and token["status"] != "ignored · deprecated"]


# --- output -------------------------------------------------------------------------------------


def render(chains: list[ChainScope]) -> str:
    today = datetime.date.today().isoformat()
    is_served = lambda chain: not chain.osmosis_decision.startswith("not served")  # noqa: E731
    served = [chain for chain in chains if is_served(chain)]
    ops = [chain for chain in chains if chain.osmosis_decision.startswith("we relay it")]
    dropped = [chain for chain in chains if not is_served(chain)]
    lines = [
        SECTION_START,
        "## Relayer scope per chain",
        "",
        f"Generated {today} by `scripts/wind-down/build_relayer_scope.py` from `relayer-map.html` (client ages) plus the youngest",
        "packet a relayer actually delivered on each leg (`tx_search` on the Stride and Osmosis RPCs). Edit the constants at the",
        "top of the script to change the rule, then rerun; `--offline` reuses the cached lookups in `relayer_scope_cache.json`.",
        "",
        "**Reading the two decision columns.** *Stride leg* says whether anything still has to cross between Stride and the",
        "chain after the upgrade: only host ICAs (the drain and the ICA transfers) and the sweep channel to Osmosis do, and we",
        "already relay those; for every other chain the Stride leg is simply not needed, whatever its state. *Osmosis leg*",
        "says how holders on the chain reach the pools: someone else already relays it (free), we will relay it, or the chain",
        "is not served and why.",
        "",
        "**The rule.** Relayers cost per chain and pool routes cost per token, so the minimum applies to a chain's total.",
        f"A leg is *free* when a packet crossed it within {FREE_PACKET_MAX_AGE_DAYS:.0f} days and its client is not expired (a fresh client header",
        "alone proves someone updates the client, not that they relay our channel; a recent packet outranks the map's stale",
        "label). A leg that is not free is run by *ops* when the",
        f"chain's in-scope value is at least ${MIN_USD_FOR_OPS_RELAYER:,}, otherwise *none*: holders there move before the upgrade or",
        "relay their own hop. The Stride leg only matters after the upgrade for hosts (our ICA relayers); nobody else gets one.",
        f"On a served chain every token worth at least ${SMALL_TOKEN_USD:,} gets a pool route. \"Last in / out\" is the age of the",
        "youngest packet received on the leg and the youngest acknowledgement delivered for the opposite direction.",
        "",
        f"Served: ${sum(c.usd for c in served):,} across {len(served)} chains, of which ${sum(c.usd for c in ops):,} needs a relayer from us"
        + (f" ({', '.join(c.name for c in ops)})" if ops else "") + "."
        + (" Not served: " + ", ".join(f"{c.name} (${c.usd:,})" for c in dropped) + "." if dropped else ""),
        "",
        "Every chain that holds any stToken is listed, largest first, whatever its status below; the USD column is the chain's",
        "total across tokens. Legs are shown where the relayer map has a route for the chain.",
        "",
        "| Chain | Total USD | Stride channel(s) | Host | Stride leg: client · last in / out | Stride leg after the upgrade | Osmosis leg: client · last in / out | Osmosis leg for holders | Pool routes |",
        "|---|---:|---|---|---|---|---|---|---|",
    ]
    for chain in chains:
        lines.append(
            f"| {chain.name} (`{chain.chain_id}`) | ${chain.usd:,} | {chain.stride_channels} | {'yes' if chain.host else ''} | {leg_cell(chain.stride_leg)} | {chain.stride_decision} "
            f"| {leg_cell(chain.osmosis_leg)} | {chain.osmosis_decision} | {', '.join(chain.pool_routes) or '–'} |"
        )
    lines += ["", "Per-token value on each chain, with the status from the tables below:", ""]
    for chain in chains:
        tokens = ", ".join(f"{t['sym']} ${t['usd']:,} ({t['status']})" for t in sorted(chain.tokens, key=lambda t: -t["usd"]))
        lines.append(f"- {chain.name}: {tokens}")
    lines += ["", SECTION_END, ""]
    return "\n".join(lines)


# --- the map page's data ------------------------------------------------------------------------


def update_relayer_map(phases: dict[str, dict], chains: list[ChainScope]) -> None:
    """Rewrite the PHASES array in the map page: every chain in every phase, with its decision."""
    for phase in phases.values():
        center_is_stride = phase["center"] == "Stride"
        existing = {route["id"]: route for route in phase["routes"]}
        routes = []
        for chain in chains:
            center = STRIDE_CHAIN_ID if center_is_stride else OSMOSIS_CHAIN_ID
            if chain.chain_id == center or chain.unsupported or chain.chain_id in DROPPED_CHAINS:
                continue  # unsupported and dropped chains are left off the map by decision
            route = existing.get(chain.chain_id) or unmapped_route(chain)
            if center_is_stride and chain.chain_id == OSMOSIS_CHAIN_ID:
                route["legs"] = sweep_channel_legs(phases)
                route["status"] = route["legs"][0]["status"] if route["legs"] else "unknown"
            route["decision"] = chain.stride_decision if center_is_stride else chain.osmosis_decision
            route["scope_status"] = map_status(chain, route)
            routes.append(route)
        if not center_is_stride and STRIDE_CHAIN_ID in existing:
            stride = existing[STRIDE_CHAIN_ID]
            stride["decision"] = "sweep channel, we relay it"
            routes.append(stride)
        for route in routes:
            route["status"] = route.pop("scope_status", route["status"])
        phase["routes"] = sorted(routes, key=lambda route: -route["usd"])

    text = RELAYER_MAP.read_text()
    new_line = "const PHASES=" + json.dumps(list(phases.values()), ensure_ascii=False, separators=(",", ": ")) + ";"
    lines = text.splitlines()
    for index, line in enumerate(lines):
        if re.match(r"\s*const PHASES\s*=", line):
            lines[index] = new_line
            RELAYER_MAP.write_text("\n".join(lines) + "\n")
            return
    raise SystemExit("PHASES line not found while writing the map")


def sweep_channel_legs(phases: dict[str, dict]) -> list[dict]:
    """The Stride → Osmosis leg (channel-5) as the map recorded it on the Osmosis-centred phase's Stride spoke."""
    for route in phases[OSMOSIS_PHASE]["routes"]:
        if route["id"] == STRIDE_CHAIN_ID:
            return [dict(leg) for leg in route["legs"]]
    return []


def unmapped_route(chain: ChainScope) -> dict:
    """A route for a chain the map had no client data for: tokens only, no legs."""
    return {
        "id": chain.chain_id,
        "name": chain.name,
        "usd": chain.usd,
        "tokens": [{"sym": t["sym"], "amount": t["amount"], "usd": t["usd"], "pct": t["pct"]} for t in chain.tokens],
        "native": [],
        "host": chain.host,
        "deprecated": chain.deprecated_only,
        "status": "unknown",
        "legs": [],
    }


def map_status(chain: ChainScope, route: dict) -> str:
    """The colour a chain gets on the map: the map's own leg status where a route exists, else why it is not served."""
    if chain.dead:
        return "dead"
    if route["legs"]:
        return route["status"]
    return "small"


def leg_cell(leg: Leg | None) -> str:
    if leg is None:
        return "not mapped"
    client = leg.client_status + ("" if leg.client_age_days is None else f" {leg.client_age_days:.1f}d")
    return f"{leg.near_channel}: {client} · {days(leg.last_recv_days)} / {days(leg.last_ack_days)}"


def days(value: float | None) -> str:
    return "never" if value is None else f"{value:.1f}d"


def splice_section(doc: pathlib.Path, section: str) -> None:
    """Replace the marked section in the doc, or insert it above the Summary table the first time."""
    text = doc.read_text()
    if SECTION_START in text and SECTION_END in text:
        start = text.index(SECTION_START)
        end = text.index(SECTION_END) + len(SECTION_END) + 1
        doc.write_text(text[:start] + section + text[end:])
        return
    anchor = text.index(INSERT_BEFORE)
    doc.write_text(text[:anchor] + section + "\n" + text[anchor:])


def run() -> None:
    try:
        main()
    except SystemExit as error:
        if error.code is None or isinstance(error.code, int):
            raise
        print(error.code, file=sys.stderr)
        print(f"RESULT: FAIL — {error.code}")
        sys.exit(1)


if __name__ == "__main__":
    run()
