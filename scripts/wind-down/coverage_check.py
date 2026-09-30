#!/usr/bin/env python3
"""The wind-down coverage check (spec §10), run before each stToken's pools are funded and
again before the halt.

Per in-scope stToken: the native tokens held on Osmosis for that denom (the vault's balance
plus every pool's native liquidity) must cover Stride's bank supply of the stToken times the
frozen HostZone.RedemptionRate; each route pool must hold its channel's escrow balance times
the rate and no more (spec §8/§10 require exact funding); the canonical pool the remainder.
Bank supply is the right reference because stTokens that left Stride over IBC are escrowed,
not burned. A pool that already served redemptions holds stTokens too: those stTokens are
already paid for, so the native requirement is (escrow - stTokens in the pool) x rate. Required
amounts round UP: rounding down could under-fund by a base unit.

Inputs: a trimmed Stride export (bank supply and balances, stakeibc host zones), a pools file
(see the PR 6 plan for the shape; every pool names the stToken denom it holds: the canonical
denom for the canonical pool, the route's two-hop denom for a route pool), the Osmosis vault
address. Reads Osmosis over REST.
Read-only; prints a table and exits 1 on any shortfall.

Usage:
  python3 scripts/wind-down/coverage_check.py --export export.json.gz --pools pools.json \
      --vault osmo1... [--osmosis-rest https://osmosis-api.polkachu.com]
"""

import argparse
import base64
import gzip
import hashlib
import json
import pathlib
import sys
import time
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from decimal import ROUND_CEILING, Decimal, localcontext
from typing import Callable

import bech32_ref

OSMOSIS_REST_DEFAULT = "https://osmosis-api.polkachu.com"
USER_AGENT = "curl/8.0"
TIMEOUT_SECONDS = 30
MAX_ATTEMPTS = 6
BACKOFF_SECONDS = 5
HTTP_TOO_MANY_REQUESTS = 429
DECIMAL_PRECISION = 80
ESCROW_HASH_BYTES = 20
TRANSFER_PORT = "transfer"
ESCROW_ADDRESS_VERSION = "ics20-1"
STRIDE_BECH32_PREFIX = "stride"
ST_DENOM_PREFIX = "st"

LiquidityFetcher = Callable[[str], dict[str, int]]


class CoverageInputError(Exception):
    """A pools entry or export section is missing or malformed; nothing was checked."""


@dataclass(frozen=True)
class RoutePool:
    channel_id: str
    pool_id: str
    st_denom: str  # the route's two-hop stToken denom as held in the pool


@dataclass(frozen=True)
class PoolSpec:
    chain_id: str
    native_denom_on_osmosis: str
    canonical_pool_id: str | None
    canonical_st_denom: str | None
    route_pools: list[RoutePool]


@dataclass
class CoverageResult:
    st_denom: str
    chain_id: str
    rate: Decimal
    supply: int
    required_native: int
    native_on_osmosis: int
    route_expected: dict[str, int] = field(default_factory=dict)
    route_shortfalls: dict[str, int] = field(default_factory=dict)
    route_overfunded: dict[str, int] = field(default_factory=dict)
    canonical_expected: int = 0
    canonical_actual: int = 0

    @property
    def covered(self) -> bool:
        total_ok = self.native_on_osmosis >= self.required_native
        canonical_ok = self.canonical_actual >= self.canonical_expected
        return total_ok and canonical_ok and not self.route_shortfalls and not self.route_overfunded


# ----------------------------------------------------------------------------------------------
# Public API
# ----------------------------------------------------------------------------------------------

def main() -> int:
    args = parse_args()
    export = load_export(path=args.export)
    pools = load_pools(path=args.pools)
    vault_balances = fetch_vault_balances(osmosis_rest=args.osmosis_rest, vault=args.vault)
    fetcher = make_liquidity_fetcher(osmosis_rest=args.osmosis_rest)

    results = evaluate(export=export, pools=pools, vault_balances=vault_balances, fetch_liquidity=fetcher)
    print_table(results)

    shortfalls = [result for result in results if not result.covered]
    if shortfalls:
        print(f"\n{len(shortfalls)} stToken(s) NOT covered: {[r.st_denom for r in shortfalls]}")
        return 1
    print("\nevery stToken is covered")
    return 0


def evaluate(
    export: dict,
    pools: dict,
    vault_balances: dict[str, int],
    fetch_liquidity: LiquidityFetcher,
) -> list[CoverageResult]:
    """Pure: the §10 arithmetic for every stToken in the pools file."""
    supply_by_denom = {entry["denom"]: int(entry["amount"]) for entry in export["app_state"]["bank"]["supply"]}
    rates = {zone["chain_id"]: Decimal(zone["redemption_rate"]) for zone in export["app_state"]["stakeibc"]["host_zone_list"]}
    transfer_channels = transfer_channel_ids(export=export)
    escrow_balances = escrow_balances_by_channel(export=export)
    if not pools:
        raise CoverageInputError("pools file has no entries")

    # The pools file must cover exactly the in-scope stTokens (the non-deprecated stakeibc zones,
    # the same set the on-chain sweep allow-list uses): a forgotten token would otherwise pass
    # unchecked, and a deprecated one has no pools to check
    in_scope = in_scope_st_denoms(export=export)
    if set(pools) != in_scope:
        raise CoverageInputError(
            f"pools file and in-scope stTokens differ: missing {sorted(in_scope - set(pools))}, "
            f"unexpected {sorted(set(pools) - in_scope)}"
        )

    results = []
    for st_denom, raw_spec in sorted(pools.items()):
        spec = parse_pool_spec(raw=raw_spec)
        if spec.chain_id not in rates:
            raise CoverageInputError(f"{st_denom}: host zone {spec.chain_id} not in the export")
        rate = rates[spec.chain_id]
        for route in spec.route_pools:
            if route.channel_id not in transfer_channels:
                # A typo'd channel would read a zero escrow and pass with a zero requirement
                raise CoverageInputError(f"{st_denom}: route channel {route.channel_id} is not a transfer channel in the export")
        if st_denom not in supply_by_denom:
            # A mistyped stToken denom would otherwise make the requirement zero and pass
            raise CoverageInputError(f"{st_denom}: not in the export's bank supply")
        supply = supply_by_denom[st_denom]
        results.append(evaluate_one(
            st_denom=st_denom, spec=spec, rate=rate, supply=supply,
            escrow_balances=escrow_balances, vault_balances=vault_balances, fetch_liquidity=fetch_liquidity,
        ))
    return results


# ----------------------------------------------------------------------------------------------
# Per-token arithmetic
# ----------------------------------------------------------------------------------------------

def evaluate_one(
    st_denom: str,
    spec: PoolSpec,
    rate: Decimal,
    supply: int,
    escrow_balances: dict[str, dict[str, int]],
    vault_balances: dict[str, int],
    fetch_liquidity: LiquidityFetcher,
) -> CoverageResult:
    native = spec.native_denom_on_osmosis
    result = CoverageResult(
        st_denom=st_denom, chain_id=spec.chain_id, rate=rate, supply=supply,
        required_native=0, native_on_osmosis=vault_balances.get(native, 0),
    )
    st_in_pools = 0

    # Each route pool must hold its channel's escrow share, less what it already paid out: a pool
    # below that cannot pay every holder on that chain, a pool above the full escrow x rate is
    # over-funded (spec §8/§10 fund exactly)
    route_escrow_total = 0
    for route in spec.route_pools:
        escrow = escrow_balances.get(route.channel_id, {}).get(st_denom, 0)
        route_escrow_total += escrow
        liquidity = fetch_liquidity(route.pool_id)
        actual = liquidity.get(native, 0)
        st_in_pool = liquidity.get(route.st_denom, 0)
        st_in_pools += st_in_pool
        expected = native_for(st_amount=max(escrow - st_in_pool, 0), rate=rate)
        result.route_expected[route.channel_id] = expected
        result.native_on_osmosis += actual
        if actual < expected:
            result.route_shortfalls[route.channel_id] = expected - actual
        if actual > native_for(st_amount=escrow, rate=rate):
            result.route_overfunded[route.channel_id] = actual - native_for(st_amount=escrow, rate=rate)

    # The canonical pool covers every other holder: everything not in a route escrow, less the
    # stTokens it already took in
    canonical_st_in_pool = 0
    if spec.canonical_pool_id is not None:
        liquidity = fetch_liquidity(spec.canonical_pool_id)
        result.canonical_actual = liquidity.get(native, 0)
        canonical_st_in_pool = liquidity.get(spec.canonical_st_denom, 0)
        st_in_pools += canonical_st_in_pool
        result.native_on_osmosis += result.canonical_actual
    result.canonical_expected = native_for(
        st_amount=max(supply - route_escrow_total - canonical_st_in_pool, 0), rate=rate,
    )
    result.required_native = native_for(st_amount=max(supply - st_in_pools, 0), rate=rate)
    return result


def in_scope_st_denoms(export: dict) -> set[str]:
    """stTokens of non-deprecated stakeibc host zones, as the sweep's allow-list defines them.

    stutia is in scope through the stakeibc celestia zone, and the whole stutia supply is checked
    at that zone's rate: staketia and stakeibc share one denom, so there is no separate staketia
    rate to apply to a subset of the supply.
    """
    zones = export["app_state"]["stakeibc"]["host_zone_list"]
    return {f"{ST_DENOM_PREFIX}{zone['host_denom']}" for zone in zones if not zone.get("deprecated", False)}


def native_for(st_amount: int, rate: Decimal) -> int:
    """stTokens × rate, rounded UP to base units: a requirement rounded down could under-fund by one."""
    with localcontext() as context:
        # The default 28 digits would round a 1e15 supply times an 18-decimal rate
        context.prec = DECIMAL_PRECISION
        return int((Decimal(st_amount) * rate).to_integral_value(rounding=ROUND_CEILING))


# ----------------------------------------------------------------------------------------------
# Export reading
# ----------------------------------------------------------------------------------------------

def load_export(path: pathlib.Path) -> dict:
    opener = gzip.open if path.suffix == ".gz" else open
    with opener(path, "rt", encoding="utf-8") as handle:
        return json.load(handle)


def load_pools(path: pathlib.Path) -> dict:
    return json.loads(path.read_text())


def parse_pool_spec(raw: dict) -> PoolSpec:
    try:
        spec = PoolSpec(
            chain_id=raw["chain_id"],
            native_denom_on_osmosis=raw["native_denom_on_osmosis"],
            canonical_pool_id=raw.get("canonical_pool_id"),
            canonical_st_denom=raw.get("canonical_st_denom"),
            route_pools=[
                RoutePool(channel_id=r["channel_id"], pool_id=r["pool_id"], st_denom=r["st_denom"])
                for r in raw.get("route_pools", [])
            ],
        )
    except KeyError as missing:
        raise CoverageInputError(f"pools entry missing {missing}") from missing
    if spec.canonical_pool_id is not None and not spec.canonical_st_denom:
        raise CoverageInputError(f"pools entry for {spec.chain_id} has a canonical pool but no canonical_st_denom")
    return spec


def escrow_address(channel_id: str) -> str:
    """ibc-go's GetEscrowAddress: ADR-028 hash of "ics20-1" NUL "transfer/<channel>", 20 bytes."""
    pre_image = ESCROW_ADDRESS_VERSION.encode() + b"\x00" + f"{TRANSFER_PORT}/{channel_id}".encode()
    return bech32_ref.encode(hrp=STRIDE_BECH32_PREFIX, address=hashlib.sha256(pre_image).digest()[:ESCROW_HASH_BYTES])


def transfer_channel_ids(export: dict) -> list[str]:
    """Every transfer-port channel in the export's IBC state."""
    channels = export["app_state"]["ibc"]["channel_genesis"]["channels"]
    channel_ids = [channel["channel_id"] for channel in channels if channel["port_id"] == TRANSFER_PORT]
    if not channel_ids:
        raise CoverageInputError("export has no transfer channels under app_state.ibc.channel_genesis.channels")
    return channel_ids


def escrow_balances_by_channel(export: dict) -> dict[str, dict[str, int]]:
    """channel id -> denom -> amount, for every transfer-port channel in the export's IBC state."""
    balances_by_address = {
        entry["address"]: {coin["denom"]: int(coin["amount"]) for coin in entry["coins"]}
        for entry in export["app_state"]["bank"]["balances"]
    }

    # Escrow accounts are not labelled in an export, but every channel is: derive the escrow
    # address of each transfer-port channel, so no channel can be missed by a scan heuristic
    return {
        channel_id: balances_by_address[escrow_address(channel_id=channel_id)]
        for channel_id in transfer_channel_ids(export=export)
        if escrow_address(channel_id=channel_id) in balances_by_address
    }


# ----------------------------------------------------------------------------------------------
# Osmosis REST
# ----------------------------------------------------------------------------------------------

def get_json(url: str) -> dict:
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    for attempt in range(1, MAX_ATTEMPTS + 1):
        try:
            with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
                return json.load(response)
        except urllib.error.HTTPError as error:
            # Public endpoints rate-limit a script that makes many calls: back off and retry
            if error.code != HTTP_TOO_MANY_REQUESTS or attempt == MAX_ATTEMPTS:
                raise
            time.sleep(BACKOFF_SECONDS * attempt)
    raise AssertionError("unreachable")


def fetch_vault_balances(osmosis_rest: str, vault: str) -> dict[str, int]:
    page = get_json(f"{osmosis_rest}/cosmos/bank/v1beta1/balances/{vault}?pagination.limit=1000")
    return {coin["denom"]: int(coin["amount"]) for coin in page["balances"]}


def make_liquidity_fetcher(osmosis_rest: str) -> LiquidityFetcher:
    def fetch(pool_id: str) -> dict[str, int]:
        pool = get_json(f"{osmosis_rest}/osmosis/poolmanager/v1beta1/pools/{pool_id}")["pool"]
        contract = pool["contract_address"]
        query = base64.b64encode(json.dumps({"get_total_pool_liquidity": {}}).encode()).decode()
        response = get_json(f"{osmosis_rest}/cosmwasm/wasm/v1/contract/{contract}/smart/{query}")
        return {coin["denom"]: int(coin["amount"]) for coin in response["data"]["total_pool_liquidity"]}
    return fetch


# ----------------------------------------------------------------------------------------------
# Output and CLI
# ----------------------------------------------------------------------------------------------

def print_table(results: list[CoverageResult]) -> None:
    header = f"{'stToken':<14}{'zone':<16}{'supply':>20}{'rate':>22}{'required':>22}{'on osmosis':>22}  status"
    print(header)
    print("-" * len(header))
    for result in results:
        status = "OK" if result.covered else "SHORT"
        print(f"{result.st_denom:<14}{result.chain_id:<16}{result.supply:>20}{str(result.rate):>22}"
              f"{result.required_native:>22}{result.native_on_osmosis:>22}  {status}")
        for channel_id, shortfall in sorted(result.route_shortfalls.items()):
            print(f"    route {channel_id}: short by {shortfall} (expected {result.route_expected[channel_id]})")
        for channel_id, surplus in sorted(result.route_overfunded.items()):
            print(f"    route {channel_id}: over-funded by {surplus} (exact funding required)")
        if result.canonical_actual < result.canonical_expected:
            print(f"    canonical: {result.canonical_actual} < expected {result.canonical_expected}")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--export", type=pathlib.Path, required=True, help="trimmed Stride export (.json or .json.gz)")
    parser.add_argument("--pools", type=pathlib.Path, required=True, help="pools file, one entry per stToken")
    parser.add_argument("--vault", required=True, help="Osmosis vault address (spec §4)")
    parser.add_argument("--osmosis-rest", default=OSMOSIS_REST_DEFAULT)
    return parser.parse_args()


if __name__ == "__main__":
    sys.exit(main())
