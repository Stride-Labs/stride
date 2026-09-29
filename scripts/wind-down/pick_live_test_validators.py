#!/usr/bin/env python3
"""Pick, per in-scope zone, the validator to use as the live test of MsgUndelegateFromValidators.

The live test drains one validator in full before the empty-list drain (spec §7, §9). The best
candidate is the validator with the smallest delegation of at least one whole token that has no unbonding entry
in flight from the delegation ICA: a full drain of it exercises the real path, costs nothing
if something is wrong, and spends one unbonding-entry slot on a validator that needs nothing
further. Writes the table into the spec's §9b, between the live-test-validators markers.

    python3 scripts/wind-down/pick_live_test_validators.py
"""

import dataclasses
import datetime
import json
import pathlib
import re
import time
import urllib.error
import urllib.request

STRIDE_REST = "https://stride-api.polkachu.com"
COINGECKO = "https://api.coingecko.com/api/v3/simple/price"
USER_AGENT = "curl/8.0"
REPO = pathlib.Path(__file__).resolve().parents[2]
SPEC = REPO / "docs" / "superpowers" / "specs" / "2026-09-18-protocol-wind-down-design.md"
LOCATIONS_DOC = REPO / "docs" / "wind-down" / "sttoken-locations.md"
SECTION_START = "<!-- live-test-validators:start -->"
SECTION_END = "<!-- live-test-validators:end -->"
TOKEN_HEADER = re.compile(r"^## st[A-Z]+ \(st\w+, host (?P<chain_id>[\w.-]+)\)")
TOKEN_STATS = re.compile(r"^Supply (?P<supply>[\d,.]+) · RR (?P<rate>[\d.]+) · \$(?P<usd>[\d,]+) total")

# chain id -> (polkachu REST host, coingecko id, decimals)
ZONES = {
    "celestia": ("celestia", "celestia", 6),
    "cosmoshub-4": ("cosmos", "cosmos", 6),
    "dydx-mainnet-1": ("dydx", "dydx-chain", 18),
    "haqq_11235-1": ("haqq", "islamic-coin", 18),
    "injective-1": ("injective", "injective-protocol", 18),
    "juno-1": ("juno", "juno-network", 6),
    "laozi-mainnet": ("band", "band-protocol", 6),
    "osmosis-1": ("osmosis", "osmosis", 6),
    "phoenix-1": ("terra", "terra-luna-2", 6),
    "sommelier-3": ("sommelier", "sommelier", 6),
    "ssc-1": ("saga", "saga-2", 6),
}


@dataclasses.dataclass
class Candidate:
    chain_id: str
    host_denom: str
    validator: str
    name: str
    delegation: int
    usd: float
    entries: int
    validators_total: int
    next_smallest_usd: float


def main() -> None:
    prices, price_source = load_prices()
    host_zones = get(f"{STRIDE_REST}/Stride-Labs/stride/stakeibc/host_zone")["host_zone"]
    candidates = [pick(zone, prices) for zone in host_zones if zone["chain_id"] in ZONES]
    splice(render(candidates, price_source))
    print(f"updated {SPEC.relative_to(REPO)} §9b")


def load_prices() -> tuple[dict[str, float], str]:
    """USD per native token keyed by chain id: CoinGecko when it answers, else the locations doc's snapshot."""
    try:
        quotes = get(f"{COINGECKO}?ids={','.join(z[1] for z in ZONES.values())}&vs_currencies=usd", attempts=2)
        return {chain_id: quotes[z[1]]["usd"] for chain_id, z in ZONES.items()}, "CoinGecko, live"
    except SystemExit:
        return prices_from_locations_doc(), f"the snapshot in {LOCATIONS_DOC.name} (total USD / (supply × rate) per token)"


def prices_from_locations_doc() -> dict[str, float]:
    prices: dict[str, float] = {}
    chain_id = None
    for line in LOCATIONS_DOC.read_text().splitlines():
        header = TOKEN_HEADER.match(line)
        if header:
            chain_id = header.group("chain_id")
            continue
        stats = TOKEN_STATS.match(line)
        if stats and chain_id:
            supply = float(stats.group("supply").replace(",", ""))
            prices[chain_id] = int(stats.group("usd").replace(",", "")) / (supply * float(stats.group("rate")))
            chain_id = None
    return prices


def pick(zone: dict, prices: dict[str, float]) -> Candidate:
    host, _, decimals = ZONES[zone["chain_id"]]
    price = prices.get(zone["chain_id"], 0.0)  # a zone without a quote (stSOMM is not tabulated) shows $0.00
    entries = unbonding_entries(host, zone["delegation_ica_address"])
    # At least one whole token: a dust delegation (1 ujuno) would not exercise the real conversion path
    funded = sorted((v for v in zone["validators"] if int(v["delegation"]) >= 10**decimals), key=lambda v: int(v["delegation"]))

    # Prefer a validator with no entry in flight; fall back to the smallest overall
    clean = [v for v in funded if entries.get(v["address"], 0) == 0]
    chosen = (clean or funded)[0]
    others = [v for v in funded if v["address"] != chosen["address"]]
    to_usd = lambda v: int(v["delegation"]) / 10**decimals * price  # noqa: E731
    return Candidate(
        chain_id=zone["chain_id"],
        host_denom=zone["host_denom"],
        validator=chosen["address"],
        name=chosen.get("name", ""),
        delegation=int(chosen["delegation"]),
        usd=to_usd(chosen),
        entries=entries.get(chosen["address"], 0),
        validators_total=len(funded),
        next_smallest_usd=to_usd(others[0]) if others else 0.0,
    )


def unbonding_entries(host: str, delegator: str) -> dict[str, int]:
    """Unbonding entries the delegation ICA holds per validator (MaxEntries is 7 per pair)."""
    data = get(f"https://{host}-api.polkachu.com/cosmos/staking/v1beta1/delegators/{delegator}/unbonding_delegations?pagination.limit=500")
    return {u["validator_address"]: len(u["entries"]) for u in data.get("unbonding_responses", [])}


def get(url: str, attempts: int = 5) -> dict:
    # polkachu wants a curl-like agent; coingecko refuses one
    agent = "Mozilla/5.0" if url.startswith(COINGECKO) else USER_AGENT
    for attempt in range(attempts):
        try:
            request = urllib.request.Request(url, headers={"User-Agent": agent})
            with urllib.request.urlopen(request, timeout=60) as response:
                return json.loads(response.read())
        except Exception as error:  # noqa: BLE001 - public endpoints rate limit; back off
            print(f"retry {attempt + 1} for {url[:80]}: {error}")
            time.sleep(8 * (attempt + 1))
    raise SystemExit(f"gave up on {url}")


def render(candidates: list[Candidate], price_source: str) -> str:
    today = datetime.date.today().isoformat()
    lines = [
        SECTION_START,
        f"Generated {today} by `scripts/wind-down/pick_live_test_validators.py`; rerun on the day, after the day-0 refresh.",
        f"Prices: {price_source}.",
        "",
        "The first `MsgUndelegateFromValidators` on each zone drains exactly one validator in full, as the live test of the tx",
        "and its callback, before the empty-list drain of the rest (spec §7, §9 step 3). The pick is the validator with the",
        "smallest recorded delegation of at least one whole token that has no unbonding entry in flight from the delegation ICA (the SDK allows 7",
        "concurrent entries per delegator-validator pair; a validator drained in full never needs a second one). \"Next\" is the",
        "second-smallest delegation, to show how much the pick matters.",
        "",
        "| Zone | Validator | Recorded delegation | USD | Entries in flight | Funded validators | Next smallest (USD) |",
        "|---|---|---:|---:|---:|---:|---:|",
    ]
    for c in candidates:
        lines.append(f"| {c.chain_id} | {c.name or c.validator} (`{c.validator}`) | {c.delegation:,} {c.host_denom} | ${c.usd:,.2f} | {c.entries} | {c.validators_total} | ${c.next_smallest_usd:,.2f} |")
    lines += ["", f"Total value put at risk by the eleven live tests: ${sum(c.usd for c in candidates):,.2f}.", SECTION_END]
    return "\n".join(lines)


def splice(section: str) -> None:
    text = SPEC.read_text()
    start = text.index(SECTION_START)
    end = text.index(SECTION_END) + len(SECTION_END)
    SPEC.write_text(text[:start] + section + text[end:])


if __name__ == "__main__":
    main()
