#!/usr/bin/env python3
"""Checks for the wind-down transmuter pools on Osmosis, before and after funding.

Everything the script needs is in the CONSTANTS block below: the vault (admin) and moderator
addresses, and one entry per pool with its Stride host zone, Osmosis pool id, the redemption
rate it was created with, and which route the pool serves (None for the canonical denom, or
the two-hop trace for a foreign route). Each pool must hold exactly that one stToken denom
plus the native token, and no limiters. Run it with no arguments after
`MsgCreateCosmWasmPool`, again after `add_new_assets`, and again after registering limiters:

    python3 scripts/wind-down/check_transmuter_pool.py

Once a pool is funded and its native token is marked corrupted (the one-way mark, spec §8), run
it with --funded: the corrupted set must then be exactly the native token rather than empty.

    python3 scripts/wind-down/check_transmuter_pool.py --funded

For each pool it reads the on-chain state and the matching host zone and checks the asset
set, factor orientation, the encoded rate, roles, the absence of limiters, denom traces and
overflow headroom, one line per check, unit-test style. Exit code is 1 when any check on any pool fails.
"""

import argparse
import dataclasses
import enum
import hashlib
import json
import math
import sys
import time
import urllib.error
import urllib.parse
import urllib.request

# Stride Labs' private Polkachu endpoints (public fallbacks: osmosis-api / stride-api.polkachu.com)
OSMOSIS_REST_DEFAULT = "https://osmosis-strd-api.polkachu.com"
STRIDE_REST_DEFAULT = "https://stride-strd-api.polkachu.com"
USER_AGENT = "curl/8.0"

TRANSMUTER_CODE_ID = "996"
TRANSMUTER_VERSION = "3.2.0"
COSMWASMPOOL_MODULE = "osmo1rxjakgd8yhks2j7hc7pt6a22z3zd64grexpyf7"
STRIDE_TO_OSMOSIS_CHANNEL_ON_OSMOSIS = "channel-326"
UINT128_MAX = 2**128 - 1
RATE_DECIMALS = 10**18


@dataclasses.dataclass(frozen=True)
class PoolSpec:
    chain_id: str  # Stride host zone
    pool_id: str  # Osmosis pool id, filled in after MsgCreateCosmWasmPool
    rate_at_creation: str | None  # the redemption rate the factors encode; None = use the live rate (zone must be halted)
    route_trace: str | None  # None = canonical pool; else the two-hop trace, e.g. "transfer/channel-0/transfer/channel-391/stuatom"


# ----------------------------------------------------------------------------------------------
# CONSTANTS: edit these, nothing else takes input.
# ----------------------------------------------------------------------------------------------

# The Osmosis vault (spec §4) is both admin and moderator of every pool; it equals OsmosisVaultAddress
# in x/stakeibc/types/wind_down.go (test_check_transmuter_pool.py asserts that).
ADMIN = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"  # Osmosis vault (transmuter admin)
MODERATOR = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"  # Osmosis vault (freeze / corrupted-asset key)

POOLS = [
    # 2026-09-25 per-route test pools. Replace with the real pools as they are created: one entry per route.
    # The test pools are administered by the test key osmo1v0694qqq6ztzxvzl807dgq7h3e857hdxvpmdlc, so every
    # entry here fails the admin and moderator checks against the vault until the list is replaced.
    PoolSpec(chain_id="cosmoshub-4", pool_id="3595", rate_at_creation="2.002036647211047463", route_trace=None),
    PoolSpec(
        chain_id="cosmoshub-4",
        pool_id="3596",
        rate_at_creation="2.002036647211047463",
        route_trace="transfer/channel-0/transfer/channel-391/stuatom",  # Hub route
    ),
    PoolSpec(
        chain_id="cosmoshub-4",
        pool_id="3597",
        rate_at_creation="2.002036647211047463",
        route_trace="transfer/channel-88/transfer/channel-37/stuatom",  # Secret route
    ),
    # Deliberately inverted factors: this entry must FAIL, proving the script catches it.
    PoolSpec(chain_id="cosmoshub-4", pool_id="3598", rate_at_creation="2.002036647211047463", route_trace=None),
]

# ----------------------------------------------------------------------------------------------

# Osmosis's transfer channel to each in-scope host zone: this is where the native token's
# canonical Osmosis denom comes from (verified against the chain registry on 2026-09-23;
# sommelier-3 was verified on chain on 2026-10-02: transfer/channel-165 is STATE_OPEN, its client
# 07-tendermint-1745 tracks sommelier-3, counterparty channel-0).
OSMOSIS_CHANNEL_TO_HOST = {
    "cosmoshub-4": "channel-0",
    "celestia": "channel-6994",
    "dydx-mainnet-1": "channel-6787",
    "haqq_11235-1": "channel-1575",
    "injective-1": "channel-122",
    "juno-1": "channel-42",
    "laozi-mainnet": "channel-148",
    "phoenix-1": "channel-251",
    "sommelier-3": "channel-165",
    "ssc-1": "channel-38946",
}


class Outcome(enum.StrEnum):
    PASS = "PASS"
    FAIL = "FAIL"
    WARN = "WARN"
    SKIP = "SKIP"
    INFO = "INFO"


COLORS = {
    Outcome.PASS: "\033[32m",
    Outcome.FAIL: "\033[31m",
    Outcome.WARN: "\033[33m",
    Outcome.SKIP: "\033[36m",
    Outcome.INFO: "\033[90m",
}
RESET = "\033[0m"


@dataclasses.dataclass
class Report:
    use_color: bool
    counts: dict[Outcome, int] = dataclasses.field(
        default_factory=lambda: {o: 0 for o in Outcome}
    )

    def line(self, outcome: Outcome, name: str, detail: str = "") -> None:
        self.counts[outcome] += 1
        tag = (
            f"{COLORS[outcome]}{outcome:>4}{RESET}"
            if self.use_color
            else f"{outcome:>4}"
        )
        suffix = f"  {detail}" if detail else ""
        print(f"  {tag}  {name}{suffix}")

    def check(self, name: str, ok: bool, detail: str = "") -> None:
        self.line(
            outcome=Outcome.PASS if ok else Outcome.FAIL, name=name, detail=detail
        )

    def section(self, title: str) -> None:
        print(f"\n{title}")


@dataclasses.dataclass
class HostZone:
    chain_id: str
    host_denom: str
    st_denom: str
    redemption_rate: str
    rate_int: int
    halted: bool
    st_supply_on_stride: int


@dataclasses.dataclass
class Pool:
    pool_id: str
    contract: str
    code_id: str
    asset_configs: list[dict[str, str]]
    alloyed_denom: str
    admin: str
    moderator: str
    admin_candidate: str | None
    is_active: bool
    swap_fee: str
    liquidity: dict[str, int]
    limiters: list[tuple[tuple[str, str], dict]]
    corrupted: list[str]
    wasm_admin: str
    cw2_version: str


# --- HTTP -----------------------------------------------------------------------------------


def get_json(url: str) -> dict:
    for attempt in range(6):
        try:
            request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
            with urllib.request.urlopen(request, timeout=30) as response:
                return json.load(response)
        except urllib.error.HTTPError as error:
            if error.code == 429:
                time.sleep(2 * (attempt + 1))
                continue
            body = error.read().decode("utf-8", errors="replace")[:300]
            return {"_error": f"HTTP {error.code}: {body}"}
    return {"_error": "rate limited, gave up"}


def smart_query(rest: str, contract: str, query: dict) -> dict:
    encoded = urllib.parse.quote(_base64(json.dumps(query).encode()), safe="")
    result = get_json(f"{rest}/cosmwasm/wasm/v1/contract/{contract}/smart/{encoded}")
    if "_error" in result:
        return result
    return result.get("data", {})


def _base64(raw: bytes) -> str:
    import base64

    return base64.b64encode(raw).decode()


def ibc_hash(path: str) -> str:
    return "ibc/" + hashlib.sha256(path.encode()).hexdigest().upper()


# --- Loading --------------------------------------------------------------------------------


def load_host_zone(stride_rest: str, chain_id: str) -> HostZone:
    zone = get_json(f"{stride_rest}/Stride-Labs/stride/stakeibc/host_zone/{chain_id}")
    if "_error" in zone:
        sys.exit(f"cannot load host zone {chain_id}: {zone['_error']}")
    host = zone["host_zone"]
    rate = host["redemption_rate"]
    rate_int = rate_to_int(rate)
    st_denom = "st" + host["host_denom"]
    supply = get_json(
        f"{stride_rest}/cosmos/bank/v1beta1/supply/by_denom?denom={st_denom}"
    )
    st_supply = int(supply.get("amount", {}).get("amount", "0"))
    return HostZone(
        chain_id=chain_id,
        host_denom=host["host_denom"],
        st_denom=st_denom,
        redemption_rate=rate,
        rate_int=rate_int,
        halted=bool(host.get("halted", False)),
        st_supply_on_stride=st_supply,
    )


def rate_to_int(rate: str) -> int:
    whole, _, frac = rate.partition(".")
    return int(whole) * RATE_DECIMALS + int(frac.ljust(18, "0")[:18])


def load_pool(osmosis_rest: str, pool_id: str) -> Pool:
    pool_response = get_json(
        f"{osmosis_rest}/osmosis/poolmanager/v1beta1/pools/{pool_id}"
    )
    if "_error" in pool_response:
        sys.exit(f"cannot load pool {pool_id}: {pool_response['_error']}")
    pool = pool_response["pool"]
    if "contract_address" not in pool:
        sys.exit(f"pool {pool_id} is not a cosmwasm pool: {pool.get('@type')}")
    contract = pool["contract_address"]

    query = lambda q: smart_query(rest=osmosis_rest, contract=contract, query=q)  # noqa: E731
    configs = query({"list_asset_configs": {}}).get("asset_configs", [])
    liquidity = {
        c["denom"]: int(c["amount"])
        for c in query({"get_total_pool_liquidity": {}}).get("total_pool_liquidity", [])
    }
    limiters = [
        ((entry[0][0], entry[0][1]), entry[1])
        for entry in query({"list_limiters": {}}).get("limiters", [])
    ]
    contract_info = get_json(
        f"{osmosis_rest}/cosmwasm/wasm/v1/contract/{contract}"
    ).get("contract_info", {})
    cw2_raw = get_json(
        f"{osmosis_rest}/cosmwasm/wasm/v1/contract/{contract}/raw/{_base64(b'contract_info')}"
    )
    cw2_version = ""
    if "data" in cw2_raw:
        import base64

        cw2_version = json.loads(base64.b64decode(cw2_raw["data"])).get("version", "")

    return Pool(
        pool_id=pool_id,
        contract=contract,
        code_id=str(pool.get("code_id", "")),
        asset_configs=configs,
        alloyed_denom=query({"get_share_denom": {}}).get("share_denom", ""),
        admin=query({"get_admin": {}}).get("admin", ""),
        moderator=query({"get_moderator": {}}).get("moderator", ""),
        admin_candidate=query({"get_admin_candidate": {}}).get("admin_candidate"),
        is_active=bool(query({"is_active": {}}).get("is_active", False)),
        swap_fee=query({"get_swap_fee": {}}).get("swap_fee", ""),
        liquidity=liquidity,
        limiters=limiters,
        corrupted=query({"get_corrupted_denoms": {}}).get("corrupted_denoms", []),
        wasm_admin=contract_info.get("admin", ""),
        cw2_version=cw2_version,
    )


# --- Checks ---------------------------------------------------------------------------------


def corrupted_assets_check(corrupted: list[str], native_denom: str, funded: bool) -> tuple[str, bool]:
    """The check name and result for the pool's corrupted set.

    Before funding nothing may be corrupted. After funding the moderator marks the native token
    corrupted so the pool only moves stToken in and native out (spec §8), and nothing else.
    """
    if funded:
        return "only the native token is marked corrupted (one-way pool)", corrupted == [native_denom]
    return "no corrupted assets", not corrupted


def check_contract(report: Report, pool: Pool, native_denom: str, funded: bool) -> None:
    report.section(f"Contract (pool {pool.pool_id}, {pool.contract})")
    report.check(
        name="code id is the transmuter",
        ok=pool.code_id == TRANSMUTER_CODE_ID,
        detail=f"code_id={pool.code_id}",
    )
    report.check(
        name="cw2 version",
        ok=pool.cw2_version == TRANSMUTER_VERSION,
        detail=f"version={pool.cw2_version or '?'}",
    )
    report.check(
        name="wasm admin is the cosmwasmpool module",
        ok=pool.wasm_admin == COSMWASMPOOL_MODULE,
        detail=pool.wasm_admin,
    )
    report.check(
        name="swap fee is zero",
        ok=pool.swap_fee == "0",
        detail=f"swap_fee={pool.swap_fee}",
    )
    report.check(name="pool is active", ok=pool.is_active)
    corrupted_name, corrupted_ok = corrupted_assets_check(
        corrupted=pool.corrupted, native_denom=native_denom, funded=funded
    )
    report.check(name=corrupted_name, ok=corrupted_ok, detail=", ".join(pool.corrupted))
    report.check(
        name="alloyed denom belongs to this contract",
        ok=pool.alloyed_denom.startswith(f"factory/{pool.contract}/alloyed/"),
        detail=pool.alloyed_denom,
    )


def check_roles(
    report: Report, pool: Pool, admin: str | None, moderator: str | None
) -> None:
    report.section("Roles")
    if admin:
        report.check(name="admin", ok=pool.admin == admin, detail=pool.admin)
    else:
        report.line(
            outcome=Outcome.INFO,
            name="admin (pass --admin to assert)",
            detail=pool.admin,
        )
    if moderator:
        report.check(
            name="moderator", ok=pool.moderator == moderator, detail=pool.moderator
        )
    else:
        report.line(
            outcome=Outcome.INFO,
            name="moderator (pass --moderator to assert)",
            detail=pool.moderator,
        )
    report.check(
        name="no admin transfer in flight",
        ok=pool.admin_candidate is None,
        detail=str(pool.admin_candidate or ""),
    )


def classify_assets(
    pool: Pool, zone: HostZone, route_trace: str | None
) -> tuple[dict | None, dict | None, list[dict], dict | None]:
    """Split asset configs into (this pool's stToken, native, unexpected extras, alloyed)."""
    st_denom = expected_st_denom(zone=zone, route_trace=route_trace)
    native_denom = native_denom_on_osmosis(zone)
    canonical = native = alloyed = None
    routes: list[dict] = []
    for config in pool.asset_configs:
        denom = config["denom"]
        if denom == pool.alloyed_denom:
            alloyed = config
        elif denom == st_denom:
            canonical = config
        elif denom == native_denom:
            native = config
        else:
            routes.append(config)
    return canonical, native, routes, alloyed


def expected_st_denom(zone: HostZone, route_trace: str | None) -> str:
    if route_trace is None:
        return ibc_hash(f"transfer/{STRIDE_TO_OSMOSIS_CHANNEL_ON_OSMOSIS}/{zone.st_denom}")
    if not route_trace.endswith(f"/{zone.st_denom}"):
        sys.exit(f"route trace {route_trace} does not end in {zone.st_denom}")
    return ibc_hash(route_trace)


def native_denom_on_osmosis(zone: HostZone) -> str:
    if zone.chain_id == "osmosis-1":
        return zone.host_denom
    channel = OSMOSIS_CHANNEL_TO_HOST.get(zone.chain_id)
    if channel is None:
        sys.exit(
            f"no Osmosis channel known for {zone.chain_id}; add it to OSMOSIS_CHANNEL_TO_HOST"
        )
    return ibc_hash(f"transfer/{channel}/{zone.host_denom}")


def check_factors(
    report: Report, pool: Pool, zone: HostZone, route_trace: str | None
) -> tuple[dict | None, dict | None, list[dict]]:
    report.section(
        f"Factors against {zone.chain_id} (RR {zone.redemption_rate}{', halted' if zone.halted else ', NOT halted'})"
    )
    if not zone.halted:
        report.line(
            outcome=Outcome.WARN,
            name="host zone is not halted: the live rate still moves, so the factor check "
            "is exact only against rate_at_creation in the CONSTANTS block",
        )
    canonical, native, routes, alloyed = classify_assets(
        pool=pool, zone=zone, route_trace=route_trace
    )
    which = "canonical" if route_trace is None else f"route {route_trace}"
    report.check(
        name=f"{which} {zone.st_denom} is a pool asset",
        ok=canonical is not None,
        detail=canonical["denom"] if canonical else "missing",
    )
    report.check(
        name="pool holds exactly one stToken denom and the native token",
        ok=len(pool.asset_configs) == 3 and not routes,
        detail=f"{len(pool.asset_configs) - 1} assets; unexpected: {[r['denom'][:22] for r in routes]}",
    )
    report.check(
        name=f"native {zone.host_denom} is a pool asset",
        ok=native is not None,
        detail=native["denom"] if native else "missing",
    )
    report.check(name="alloyed asset is listed", ok=alloyed is not None)
    if canonical is None or native is None or alloyed is None:
        return canonical, native, routes

    f_st = int(canonical["normalization_factor"])
    f_native = int(native["normalization_factor"])
    f_alloyed = int(alloyed["normalization_factor"])

    # 1 stToken must be worth RR native: out = in × f_native / f_st, so f_native / f_st == RR exactly.
    ratio_ok = f_native * RATE_DECIMALS == zone.rate_int * f_st
    report.check(
        name="native factor / stToken factor == redemption rate, exactly",
        ok=ratio_ok,
        detail=f"{f_native} / {f_st} vs {zone.redemption_rate}",
    )
    report.check(
        name="factors are not inverted (native factor > stToken factor when RR > 1)",
        ok=(f_native > f_st) == (zone.rate_int > RATE_DECIMALS),
    )
    report.check(
        name="alloyed factor equals the native factor (1 share = 1 native base unit)",
        ok=f_alloyed == f_native,
        detail=str(f_alloyed),
    )

    return canonical, native, routes


def check_prices(
    report: Report,
    osmosis_rest: str,
    pool: Pool,
    zone: HostZone,
    canonical: dict | None,
    native: dict | None,
    routes: list[dict],
) -> None:
    report.section("Live quotes from the contract")
    if canonical is None or native is None:
        report.line(
            outcome=Outcome.SKIP,
            name="quotes",
            detail="canonical or native asset missing",
        )
        return
    query = lambda q: smart_query(rest=osmosis_rest, contract=pool.contract, query=q)  # noqa: E731

    spot = query(
        {
            "spot_price": {
                "base_asset_denom": canonical["denom"],
                "quote_asset_denom": native["denom"],
            }
        }
    ).get("spot_price", "")
    report.check(
        name="spot price stToken/native equals the redemption rate string",
        ok=spot == zone.redemption_rate,
        detail=spot,
    )
    for route in routes:
        route_spot = query(
            {
                "spot_price": {
                    "base_asset_denom": route["denom"],
                    "quote_asset_denom": canonical["denom"],
                }
            }
        ).get("spot_price", "")
        report.check(
            name=f"spot price {route['denom'][:22]}…/stToken is 1",
            ok=route_spot == "1",
            detail=route_spot,
        )

    # calc_out simulates against liquidity, so on an unfunded pool it errors with the required amount;
    # either way the number it produces must be floor(1e6 × RR).
    expected_out = 10**6 * zone.rate_int // RATE_DECIMALS
    quote = get_json(
        f"{osmosis_rest}/cosmwasm/wasm/v1/contract/{pool.contract}/smart/"
        + urllib.parse.quote(
            _base64(
                json.dumps(
                    {
                        "calc_out_amt_given_in": {
                            "token_in": {
                                "denom": canonical["denom"],
                                "amount": "1000000",
                            },
                            "token_out_denom": native["denom"],
                            "swap_fee": "0",
                        }
                    }
                ).encode()
            ),
            safe="",
        )
    )
    if "_error" in quote:
        required = _parse_required(quote["_error"])
        report.check(
            name="1,000,000 stToken exact-in quotes floor(1e6 × RR) (from the empty-pool error)",
            ok=required == expected_out,
            detail=f"required={required} expected={expected_out}",
        )
    else:
        out = int(quote.get("data", {}).get("token_out", {}).get("amount", "0"))
        report.check(
            name="1,000,000 stToken exact-in quotes floor(1e6 × RR)",
            ok=out == expected_out,
            detail=f"out={out} expected={expected_out}",
        )


def _parse_required(error_text: str) -> int | None:
    marker = "required: "
    if marker not in error_text:
        return None
    digits = ""
    for char in error_text.split(marker, 1)[1]:
        if not char.isdigit():
            break
        digits += char
    return int(digits) if digits else None


def check_traces(
    report: Report,
    osmosis_rest: str,
    zone: HostZone,
    canonical: dict | None,
    native: dict | None,
    routes: list[dict],
    route_trace: str | None,
) -> None:
    report.section("Denom traces on Osmosis")
    st_suffix = (
        f"transfer/{STRIDE_TO_OSMOSIS_CHANNEL_ON_OSMOSIS}/{zone.st_denom}"
        if route_trace is None
        else route_trace
    )
    for label, config, expected_suffix in [
        ("this pool's stToken", canonical, st_suffix),
        ("native token", native, f"/{zone.host_denom}"),
    ] + [(f"route {r['denom'][:22]}…", r, f"/{zone.st_denom}") for r in routes]:
        if config is None or not config["denom"].startswith("ibc/"):
            if config is not None:
                report.line(
                    outcome=Outcome.INFO,
                    name=f"{label} is a native Osmosis denom",
                    detail=config["denom"],
                )
            continue
        trace = get_json(
            f"{osmosis_rest}/ibc/apps/transfer/v1/denom_traces/{config['denom'][4:]}"
        )
        if "_error" in trace:
            report.line(
                outcome=Outcome.SKIP, name=f"{label} trace", detail=trace["_error"][:80]
            )
            continue
        path = (
            trace.get("denom_trace", {}).get("path", "")
            + "/"
            + trace.get("denom_trace", {}).get("base_denom", "")
        )
        report.check(
            name=f"{label} trace ends in {expected_suffix}",
            ok=path.endswith(expected_suffix),
            detail=path,
        )
        if label.startswith("this pool") and route_trace is not None:
            report.check(
                name="this pool's stToken is a two-hop denom",
                ok=path.count("transfer/") == 2,
                detail=path,
            )


def check_limiters(report: Report, pool: Pool) -> None:
    report.section("Limiters")
    labels = [f"{denom[:22]}…/{label}" for (denom, label), _ in pool.limiters]
    report.check(name="no limiters registered on any asset", ok=not pool.limiters, detail=str(labels))


def check_liquidity_and_headroom(
    report: Report,
    pool: Pool,
    zone: HostZone,
    canonical: dict | None,
    native: dict | None,
) -> None:
    report.section("Liquidity and overflow headroom")
    total = sum(pool.liquidity.values())
    if total == 0:
        report.line(
            outcome=Outcome.INFO,
            name="pool is empty (expected before the funding join)",
        )
    else:
        report.line(
            outcome=Outcome.INFO,
            name="pool already holds liquidity",
            detail=str({d[:14]: a for d, a in pool.liquidity.items() if a}),
        )
    if canonical is None or native is None:
        return
    factors = [
        int(c["normalization_factor"])
        for c in pool.asset_configs
        if c["denom"] != pool.alloyed_denom
    ]
    lcm = 1
    for factor in factors:
        lcm = lcm * factor // math.gcd(lcm, factor)
    report.check(
        name="lcm of factors fits Uint128",
        ok=lcm <= UINT128_MAX,
        detail=f"lcm≈{lcm:.3e}",
    )
    f_st = int(canonical["normalization_factor"])
    f_native = int(native["normalization_factor"])
    # Worst case: every stToken in existence lands in the pool alongside the native backing for it.
    st_units = zone.st_supply_on_stride
    native_units = st_units * zone.rate_int // RATE_DECIMALS
    normalized_total = st_units * lcm // f_st + native_units * lcm // f_native
    report.check(
        name=f"normalized value of the whole supply ({st_units} {zone.st_denom}) fits Uint128",
        ok=normalized_total <= UINT128_MAX,
        detail=f"≈{normalized_total:.3e}",
    )
    report.check(
        name="alloyed minted for the whole supply fits Uint128",
        ok=st_units * f_native // f_st <= UINT128_MAX,
    )


# --- Main -----------------------------------------------------------------------------------


def check_pool(spec: PoolSpec, use_color: bool, funded: bool) -> Report:
    report = Report(use_color=use_color)
    zone = load_host_zone(stride_rest=STRIDE_REST_DEFAULT, chain_id=spec.chain_id)
    if spec.rate_at_creation:
        zone = dataclasses.replace(
            zone,
            redemption_rate=spec.rate_at_creation,
            rate_int=rate_to_int(spec.rate_at_creation),
        )
    pool = load_pool(osmosis_rest=OSMOSIS_REST_DEFAULT, pool_id=spec.pool_id)
    route = "canonical" if spec.route_trace is None else spec.route_trace
    print(
        f"\n=== Pool {pool.pool_id}: {zone.st_denom} ({route}) → {zone.host_denom}, host zone {zone.chain_id} ==="
    )

    check_contract(report=report, pool=pool, native_denom=native_denom_on_osmosis(zone), funded=funded)
    check_roles(report=report, pool=pool, admin=ADMIN, moderator=MODERATOR)
    canonical, native, routes = check_factors(
        report=report, pool=pool, zone=zone, route_trace=spec.route_trace
    )
    check_prices(
        report=report,
        osmosis_rest=OSMOSIS_REST_DEFAULT,
        pool=pool,
        zone=zone,
        canonical=canonical,
        native=native,
        routes=routes,
    )
    check_traces(
        report=report,
        osmosis_rest=OSMOSIS_REST_DEFAULT,
        zone=zone,
        canonical=canonical,
        native=native,
        routes=routes,
        route_trace=spec.route_trace,
    )
    check_limiters(report=report, pool=pool)
    check_liquidity_and_headroom(
        report=report, pool=pool, zone=zone, canonical=canonical, native=native
    )
    counts = report.counts
    print(
        f"  → {counts[Outcome.PASS]} passed, {counts[Outcome.FAIL]} failed, "
        f"{counts[Outcome.WARN]} warnings, {counts[Outcome.SKIP]} skipped"
    )
    return report


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument(
        "--funded",
        action="store_true",
        help="the pools are funded and one-way: expect exactly the native token marked corrupted",
    )
    funded = parser.parse_args().funded

    use_color = sys.stdout.isatty()
    failed_pools = [
        spec.pool_id
        for spec in POOLS
        if check_pool(spec=spec, use_color=use_color, funded=funded).counts[Outcome.FAIL]
    ]
    print()
    if failed_pools:
        print(f"FAILED: pools {', '.join(failed_pools)} have failing checks")
        return 1
    print(f"OK: all {len(POOLS)} pool(s) passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
