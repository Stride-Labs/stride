"""Collector for the Pools tab: the wind-down transmuter pools on Osmosis, their gate checks and their funding math.

Absorbs scripts/wind-down/check_transmuter_pool.py (the per-pool gate) and does the coverage check live: every code-996
pool the vault administers is discovered from Osmosis, assigned to a zone by the stToken it holds, resolved to a Stride
channel when it serves a foreign route, and given its allocation (a route pool: its channel's escrow at the pool's own
rate; the canonical pool: everything else the zone holds on Osmosis, less the vault's OSMO fee reserve for osmosis-1).
A second pool of the same kind, or on the same route channel, is flagged and left unallocated. Every integer in the
payload is a string.

The planned pools are the counterpart: the canonical pool plus one per policy route a zone should have, each with the
denoms it will hold, whether the test wallet holds its stToken denom, the factors it will be created with and the
existing pool that already holds the denom, so the Multisig tab can write the creation txs and the Pools tab shows
what is still missing.
"""

import base64
import collections
import concurrent.futures
import dataclasses
import decimal
import functools
import json
import math
from collections.abc import Callable
from dataclasses import dataclass
from decimal import Decimal
from enum import StrEnum
from typing import Any

import chain
import config
import funds

TRANSMUTER_CODE_ID = "996"
TRANSMUTER_VERSION = "3.2.0"
COSMWASMPOOL_LISTING_PATH = "/osmosis/cosmwasmpool/v1beta1/pools"
CW2_INFO_KEY = b"contract_info"  # the cw2 raw storage key that holds the contract name and version
UINT128_LIMIT = 2**128  # the transmuter keeps every normalised balance in a Uint128
ROUTE_HOPS = 2  # a route pool's stToken travelled Stride -> holder chain -> Osmosis
ST_PREFIX = "st"
CANONICAL_SLOT = "canonical"  # the one slot every canonical pool of a zone claims; a route pool claims its channel
SIX_DECIMALS = 6
ST_FACTOR_SIX_DECIMALS = 10**18  # the stToken factor a six-decimal zone's pool is created with
ST_FACTOR_EIGHTEEN_DECIMALS = 10**6  # an 18-decimal zone's: 1e18-scaled factors overflow weights() on its supply
SEED_AMOUNT_PLACES = 2  # the seed transfer sends 0.01 stToken: any non-zero amount gives the denom supply
# Polkachu's REST serves the capitalised path and answers "Not Implemented" for the lower-case one; try both.
POOLMANAGER_PARAMS_PATHS = ("/osmosis/poolmanager/v1beta1/Params", "/osmosis/poolmanager/v1beta1/params")

ZONE_WORKERS = 16
POOL_WORKERS = 8
RATE_PLACES = Decimal("1.000000000000000000")
GAP_PLACES = Decimal("0.0001")
# Ceil(escrow x rate) on an 18-decimal supply needs ~46 exact digits; 80 leaves room for the ratio too.
DECIMAL_CONTEXT = decimal.Context(prec=80)


class PoolKind(StrEnum):
    CANONICAL = "canonical"
    ROUTE = "route"
    UNRECOGNISED = "unrecognised"  # a route pool whose stToken path matches no policy channel: reported, not allocated


class CheckName(StrEnum):
    CODE_ID = "code id 996"
    CW2_VERSION = "cw2 version 3.2.0"
    ASSETS = "assets: one stToken denom + the native, nothing else"
    FACTORS = "factors: native / stToken == rate, native == alloyed"
    RATE = "rate <= Stride's rate"
    ADMIN = "admin is the vault"
    MODERATOR = "moderator is the pool moderator"
    NO_ADMIN_TRANSFER = "no admin transfer in flight"
    ACTIVE = "active"
    NO_LIMITERS = "no limiters"
    ST_TRACE = "stToken trace"
    NATIVE_TRACE = "native trace"
    HEADROOM = "uint128 headroom"
    CORRUPTED = "corrupted set"
    UNIQUE = "only pool of its kind/channel"


# ---- payload


@dataclass(frozen=True)
class Check:
    name: CheckName
    ok: bool | None  # None when the lookup the check needs failed
    detail: str


@dataclass(frozen=True)
class Route:
    """The foreign route a pool serves: holders on `chain_id` whose stTokens are escrowed on Stride's `stride_channel`."""

    chain_id: str
    stride_channel: str
    counterparty_channel: str  # the holder chain's channel to Stride (the second hop of the pool's stToken trace)


@dataclass(frozen=True)
class PoolReport:
    contract: str
    pool_id: str | None
    kind: PoolKind
    route: Route | None
    st_denom: str  # the stToken as held on Osmosis (ibc/...)
    st_trace: str  # its full trace path
    st_balance: int
    native_balance: int
    alloyed_denom: str
    alloyed_supply: int | None
    vault_shares: int
    test_wallet_shares: int  # the test join's shares (config.POOL_SEED_ADDRESS); strangers' shares are not tracked
    rate: str | None  # native factor / stToken factor, fixed at creation
    rate_gap_pct: str | None  # (stride_rate - rate) / stride_rate x 100: the surplus that stays in the canonical pool
    escrow: int | None  # route only: Stride's escrow balance of the stToken on the route's channel
    allocation: int | None
    funded_exactly: (
        bool | None
    )  # vault_shares == allocation: shares are minted 1:1 to native joined and stay in the vault
    corrupted: list[str]
    native_marked: bool  # the native token is in the corrupted set (the one-way mark after funding)
    checks: list[Check]
    ready: bool  # every check ok


# ---- internal structures


@dataclass(frozen=True)
class ListedPool:
    pool_id: str
    code_id: str


@dataclass(frozen=True)
class AdminCandidate:
    """The answer to get_admin_candidate, kept apart from a failed lookup (which is None)."""

    address: str | None


@dataclass(frozen=True)
class DenomTrace:
    path: str  # e.g. transfer/channel-0/transfer/channel-391
    base_denom: str

    @property
    def full(self) -> str:
        return f"{self.path}/{self.base_denom}"


@dataclass(frozen=True)
class RawPool:
    """A transmuter contract as read from Osmosis, before it is assigned to a zone.

    The required reads (assets, share denom, liquidity, traces) fail the whole snapshot; each optional read is None
    when it failed, and the check that needs it reports n/a.
    """

    contract: str
    pool_id: str | None  # None for an EXTRA_POOL_CONTRACTS entry the cosmwasmpool listing does not hold
    code_id: str | None
    admin: str | None
    alloyed_denom: str
    assets: dict[str, int]  # asset denom -> normalization factor, including the alloyed denom
    liquidity: dict[str, int]
    traces: dict[str, DenomTrace]  # ibc asset denom -> its trace
    alloyed_supply: int | None
    cw2_version: str | None
    moderator: str | None
    admin_candidate: AdminCandidate | None
    is_active: bool | None
    limiters: list[str] | None  # "<denom>/<label>" per registered limiter
    corrupted: list[str] | None


@dataclass(frozen=True)
class ZoneDenoms:
    chain_id: str
    host_denom: str
    st_denom: str
    osmosis_denom: str  # the native token on Osmosis: the bare denom for osmosis-1, else the voucher over the host channel
    native_trace: str  # what the native asset's trace must read (the bare denom for osmosis-1)
    canonical_st_trace: str  # transfer/channel-326/<st_denom>


@dataclass(frozen=True)
class StAsset:
    """The stToken a pool holds for a zone."""

    denom: str
    trace: DenomTrace


@dataclass(frozen=True)
class RouteHops:
    osmosis_channel: str  # Osmosis's channel to the holder chain
    holder_channel: str  # the holder chain's channel to Stride


@dataclass(frozen=True)
class ChannelTarget:
    """What a Stride transfer channel points at: its counterparty channel and the chain its client tracks."""

    counterparty_channel: str | None
    chain_id: str


@dataclass(frozen=True)
class RouteLookup:
    """A route resolution that ran: the holder chain, and the policy channel it matched or None."""

    chain_id: str
    route: Route | None


@dataclass(frozen=True)
class HostZoneInfo:
    host_denom: str
    stride_rate: str
    st_supply: int

    @property
    def needed(self) -> int:
        """What the zone's pools must hold between them: the stToken supply at Stride's rate, rounded up."""
        return ceil_times_rate(amount=self.st_supply, rate=self.stride_rate)


@dataclass(frozen=True)
class CreationFee:
    """The poolmanager's pool_creation_fee: one coin, paid by the creator (the vault) per pool."""

    denom: str
    amount: int


@dataclass(frozen=True)
class Factors:
    """The normalization factors a pool is created with: native / stToken encodes the rate."""

    st_factor: int
    native_factor: int


@dataclass(frozen=True)
class PlannedRoute:
    """One policy channel resolved to what its stToken crosses: Stride -> the holder chain X -> Osmosis.

    Resolution stops at the first gap, leaving the later fields None and `error` saying where; what did resolve is
    kept so the Pools tab can show it.
    """

    stride_channel: str
    holder_chain_id: str | None
    holder: config.HolderChain | None
    counterparty_channel: str | None  # X -> Stride
    osmosis_channel: str | None  # Osmosis -> X
    holder_to_osmosis_channel: str | None  # X -> Osmosis
    error: str | None


@dataclass(frozen=True)
class RouteDenoms:
    """A route stToken's `ibc/` denoms on the holder chain (one hop) and on Osmosis (two hops), where resolvable."""

    on_holder: str | None
    on_osmosis: str | None


@dataclass(frozen=True)
class PlannedPool:
    """A pool the zone will have: the canonical one, or one per policy route (`stride_channel`).

    `seeded` is whether the test wallet (config.POOL_SEED_ADDRESS) holds its stToken denom on Osmosis, which proves the
    denom exists there (a cosmwasm pool cannot hold a denom with no supply) and funds the test join;
    `seed_command` is the single-signer IBC transfer that gives it some; `live_contract` is the existing pool that
    already holds the denom, which replaces the planned row once created. `error` says why a route could not be
    resolved; such a pool has no denoms, message or commands.
    """

    kind: PoolKind
    stride_channel: str | None  # route only
    holder_chain_id: str | None
    holder_name: str | None
    holder_binary: str | None
    holder_node: str | None
    counterparty_channel: str | None  # the holder chain's channel to Stride (None for canonical)
    osmosis_channel: str | None  # Osmosis's channel to the holder chain (channel-326 for canonical: to Stride)
    holder_to_osmosis_channel: str | None  # the holder chain's channel to Osmosis (channel-5 for canonical)
    denom_on_holder: str | None  # ibc/... of transfer/<counterparty_channel>/<st_denom>; None for canonical
    denom_on_osmosis: str | None  # the stToken as the pool will hold it
    seeded: bool | None
    test_wallet_balance: int | None
    st_factor: int
    native_factor: int
    alloyed_subdenom: str
    instantiate_msg: dict[str, Any] | None
    seed_command: str | None
    live_contract: str | None
    error: str | None
    blocked: str | None = None  # why the pool cannot be created yet (config.BLOCKED_HOLDER_CHAINS); None when it can


@dataclass(frozen=True)
class OsmosisSnapshot:
    vault_balances: dict[str, int]
    pools: list[RawPool]
    test_wallet_balances: dict[str, int] = dataclasses.field(default_factory=dict)  # config.POOL_SEED_ADDRESS
    creation_fee: CreationFee | None = None  # None when the poolmanager params could not be read
    # Osmosis's channel to each holder chain (config.HOLDER_ROUTES) -> what it points at; None where the lookup failed.
    holder_targets: dict[str, ChannelTarget | None] = dataclasses.field(default_factory=dict)


@dataclass(frozen=True)
class ZonePools:
    chain_id: str
    symbol: str
    decimals: int
    st_denom: str
    osmosis_denom: str
    host_denom: str
    stride_rate: str
    st_supply: int
    needed: int  # st_supply x stride_rate, rounded up
    vault_native: int
    fee_reserve: int  # uosmo the vault keeps for the funding txs (osmosis-1 only, 0 elsewhere): not allocated
    pools_native: int
    native_on_osmosis: int  # vault + pools
    coverage: str | None  # native_on_osmosis / needed, six places; None when nothing is needed
    missing_routes: list[str]  # policy channels with no route pool
    pools: list[PoolReport]  # canonical first, then routes by Stride channel, unrecognised last
    pools_ready: (
        bool | None
    )  # a canonical pool exists, no missing route, every pool ready; None while a check is unknown
    pools_funded: bool | None  # pools_ready and every pool funded exactly and marked; None while an input is unknown
    planned: list[PlannedPool]  # canonical first, then routes by Stride channel
    routes_seeded: bool | None  # the test wallet holds every route planned pool's stToken denom
    canonical_seeded: bool | None
    pools_created: bool | None  # every planned pool has a live contract
    creation_fee: CreationFee | None  # the poolmanager's pool_creation_fee; None when it could not be read
    vault_fee_balance: int | None  # the vault's balance of the fee denom
    creation_fee_short: bool | None  # vault_fee_balance < fee x planned pools without a live contract


# Contracts whose admin is not the vault: adminship of someone else's pool never changes, so a negative answer is
# memoised for the process lifetime. Our own pools are re-read every refresh so an admin transfer shows.
_NOT_OURS: set[str] = set()


def collect() -> dict[str, Any]:
    stride = chain.stride_chain()
    osmosis = chain.zone_chain(zone=config.ZONES_BY_CHAIN_ID[config.OSMOSIS_CHAIN_ID])
    host_zones = {
        host_zone["chain_id"]: host_zone
        for host_zone in chain.rest_get(
            chain=stride, path="/Stride-Labs/stride/stakeibc/host_zone"
        )["host_zone"]
    }

    # Every zone reads the same Osmosis snapshot, so an unreadable Osmosis is every zone's error.
    snapshot = chain.within_error_boundary(
        chain_id=config.OSMOSIS_CHAIN_ID,
        work=lambda: _osmosis_snapshot(osmosis=osmosis),
    )
    if isinstance(snapshot, chain.ZoneError):
        zones: list[ZonePools | chain.ZoneError] = [
            chain.ZoneError(chain_id=zone.chain_id, error=snapshot.error)
            for zone in config.ZONES
        ]
        return _payload(zones=zones)

    with concurrent.futures.ThreadPoolExecutor(max_workers=ZONE_WORKERS) as pool:
        zones = list(
            pool.map(
                lambda zone: chain.within_error_boundary(
                    chain_id=zone.chain_id,
                    work=lambda: _zone_pools(
                        zone=zone,
                        host_zone=host_zones[zone.chain_id],
                        stride=stride,
                        osmosis=osmosis,
                        snapshot=snapshot,
                    ),
                ),
                config.ZONES,
            )
        )
    return _payload(zones=zones)


# ---- pure logic: discovery and assignment


def transmuter_candidates(listed: dict[str, ListedPool], skip: set[str]) -> list[str]:
    """Contracts worth asking for their admin: the code-996 pools not already known to be someone else's."""
    return [
        contract
        for contract, pool in listed.items()
        if pool.code_id == TRANSMUTER_CODE_ID and contract not in skip
    ]


def select_our_pools(
    admins: dict[str, str | None], extra: tuple[str, ...]
) -> list[str]:
    """The pools the vault administers, plus the configured extras whatever their admin (or if it could not be read)."""
    return [
        contract
        for contract, admin in admins.items()
        if admin == config.OSMOSIS_VAULT or contract in extra
    ]


def zone_denoms(chain_id: str, host_denom: str) -> ZoneDenoms:
    st_denom = f"{ST_PREFIX}{host_denom}"
    osmosis_channel = config.OSMOSIS_CHANNEL_TO_HOST.get(chain_id)
    native_trace = (
        host_denom
        if osmosis_channel is None
        else f"{chain.TRANSFER_PORT}/{osmosis_channel}/{host_denom}"
    )
    return ZoneDenoms(
        chain_id=chain_id,
        host_denom=host_denom,
        st_denom=st_denom,
        osmosis_denom=host_denom
        if osmosis_channel is None
        else chain.ibc_denom(path=native_trace),
        native_trace=native_trace,
        canonical_st_trace=f"{chain.TRANSFER_PORT}/{funds.STRIDE_CHANNEL_ON_OSMOSIS}/{st_denom}",
    )


def st_asset_of(raw: RawPool, zone: ZoneDenoms) -> StAsset | None:
    """The zone's stToken among the pool's assets (the canonical one when several), or None: the pool is not the zone's."""
    candidates = sorted(
        (
            StAsset(denom=denom, trace=trace)
            for denom, trace in raw.traces.items()
            if trace.base_denom == zone.st_denom
        ),
        key=lambda asset: (asset.trace.full != zone.canonical_st_trace, asset.denom),
    )
    return candidates[0] if candidates else None


def route_hops(trace: DenomTrace) -> RouteHops | None:
    """The two channels of a route pool's stToken path, or None when the path is not exactly two transfer hops."""
    parts = trace.path.split("/")
    ports, channels = parts[0::2], parts[1::2]
    if len(channels) != ROUTE_HOPS or any(
        port != chain.TRANSFER_PORT for port in ports
    ):
        return None
    return RouteHops(osmosis_channel=channels[0], holder_channel=channels[1])


def match_route(
    holder_chain_id: str,
    holder_channel: str,
    policy_channels: frozenset[str],
    targets: dict[str, ChannelTarget],
) -> Route | None:
    """The policy channel whose counterparty is the holder's channel and whose client tracks the holder chain."""
    matches = [
        channel
        for channel in sorted(policy_channels, key=_channel_number)
        if channel in targets
        and targets[channel].counterparty_channel == holder_channel
        and targets[channel].chain_id == holder_chain_id
    ]
    if not matches:
        return None
    return Route(
        chain_id=holder_chain_id,
        stride_channel=matches[0],
        counterparty_channel=holder_channel,
    )


# ---- pure logic: per pool


def build_pool_report(
    raw: RawPool,
    st: StAsset,
    zone: ZoneDenoms,
    stride_rate: str,
    needed: int,
    vault_shares: int,
    route_lookup: RouteLookup | None,
    escrow: int | None,
    test_wallet_shares: int = 0,
) -> PoolReport:
    """A route pool's allocation is known here; a canonical pool's is filled in by `allocate` once every route is.

    `needed` is what the zone's pools must hold between them: the headroom check sizes the pool by it.
    """
    kind = _kind(st=st, zone=zone, route_lookup=route_lookup)
    route = route_lookup.route if route_lookup else None
    rate = pool_rate(raw=raw, st_denom=st.denom, native_denom=zone.osmosis_denom)
    allocation = (
        route_allocation(escrow=escrow, rate=rate) if kind == PoolKind.ROUTE else None
    )
    corrupted = raw.corrupted or []
    checks = build_checks(
        raw=raw,
        st=st,
        zone=zone,
        kind=kind,
        route_lookup=route_lookup,
        rate=rate,
        stride_rate=stride_rate,
        needed=needed,
        vault_shares=vault_shares,
    )
    return PoolReport(
        contract=raw.contract,
        pool_id=raw.pool_id,
        kind=kind,
        route=route,
        st_denom=st.denom,
        st_trace=st.trace.full,
        st_balance=raw.liquidity.get(st.denom, 0),
        native_balance=raw.liquidity.get(zone.osmosis_denom, 0),
        alloyed_denom=raw.alloyed_denom,
        alloyed_supply=raw.alloyed_supply,
        vault_shares=vault_shares,
        test_wallet_shares=test_wallet_shares,
        rate=rate,
        rate_gap_pct=rate_gap_pct(stride_rate=stride_rate, rate=rate),
        escrow=escrow if kind == PoolKind.ROUTE else None,
        allocation=allocation,
        funded_exactly=None if allocation is None else vault_shares == allocation,
        corrupted=corrupted,
        native_marked=zone.osmosis_denom in corrupted,
        checks=checks,
        ready=all(check.ok is True for check in checks),
    )


def allocate(reports: list[PoolReport], vault_native: int, fee_reserve: int = 0) -> list[PoolReport]:
    """Flag the pools that share a slot (a second canonical pool, a second pool on one route channel) and leave them
    unallocated, then give the canonical pool the remainder: what the zone holds on Osmosis less what the routes are
    owed."""
    duplicates = duplicate_slots(reports=reports)
    flagged = [_flag_duplicate(report=report, duplicates=duplicates) for report in reports]

    # Every canonical pool claims the same slot, so either none is a duplicate or all are.
    allocation = canonical_allocation(reports=flagged, vault_native=vault_native, fee_reserve=fee_reserve)
    return [
        dataclasses.replace(
            report,
            allocation=allocation,
            funded_exactly=None
            if allocation is None
            else report.vault_shares == allocation,
        )
        if report.kind == PoolKind.CANONICAL and CANONICAL_SLOT not in duplicates
        else report
        for report in flagged
    ]


def canonical_allocation(reports: list[PoolReport], vault_native: int, fee_reserve: int = 0) -> int | None:
    """vault (less the fee reserve, never below zero) + every pool's native - every route's allocation; None while any
    non-canonical pool's share is unknown."""
    others = [report for report in reports if report.kind != PoolKind.CANONICAL]
    if any(report.allocation is None for report in others):
        return None
    return (
        max(vault_native - fee_reserve, 0)
        + sum(report.native_balance for report in reports)
        - sum(report.allocation for report in others)
    )


def fee_reserve(chain_id: str) -> int:
    """The uosmo the vault keeps for the funding txs: only osmosis-1's native token is also the gas token."""
    return config.OSMO_FEE_RESERVE if chain_id == config.OSMOSIS_CHAIN_ID else 0


def pool_slot(kind: PoolKind, route: Route | None) -> str | None:
    """What a pool must be the only one of: the canonical slot or its route channel; None when it claims neither."""
    if kind == PoolKind.CANONICAL:
        return CANONICAL_SLOT
    if kind == PoolKind.ROUTE and route is not None:
        return f"route on {route.stride_channel}"
    return None


def duplicate_slots(reports: list[PoolReport]) -> dict[str, list[str]]:
    """Slot -> the contracts claiming it, for every slot more than one pool claims."""
    slots = {report.contract: pool_slot(kind=report.kind, route=report.route) for report in reports}
    claims = {
        slot: sorted(contract for contract, claimed in slots.items() if claimed == slot)
        for slot in set(slots.values())
        if slot is not None
    }
    return {slot: contracts for slot, contracts in claims.items() if len(contracts) > 1}


def _flag_duplicate(report: PoolReport, duplicates: dict[str, list[str]]) -> PoolReport:
    """A pool sharing its slot gets its uniqueness check failed and nothing allocated until one of them goes."""
    slot = pool_slot(kind=report.kind, route=report.route)
    if slot not in duplicates:
        return report

    failed = Check(
        name=CheckName.UNIQUE,
        ok=False,
        detail=f"{len(duplicates[slot])} pools claim {slot}: {', '.join(duplicates[slot])}",
    )
    checks = [failed if entry.name == CheckName.UNIQUE else entry for entry in report.checks]
    return dataclasses.replace(report, allocation=None, funded_exactly=None, checks=checks, ready=False)


def route_allocation(escrow: int | None, rate: str | None) -> int | None:
    """ceil(escrow x the pool's own rate): what the pool will actually pay out for every escrowed stToken."""
    if escrow is None or rate is None:
        return None
    return ceil_times_rate(amount=escrow, rate=rate)


def ceil_times_rate(amount: int, rate: str) -> int:
    """amount x rate rounded UP to base units: rounding down could under-fund by one."""
    product = DECIMAL_CONTEXT.multiply(Decimal(amount), Decimal(rate))
    return int(product.to_integral_value(rounding=decimal.ROUND_CEILING))


def pool_rate(raw: RawPool, st_denom: str, native_denom: str) -> str | None:
    """native factor / stToken factor, 18 places, as the Funds tab computes it; None when either asset is missing."""
    if st_denom not in raw.assets or native_denom not in raw.assets:
        return None
    rate = factor_ratio(native_factor=raw.assets[native_denom], st_factor=raw.assets[st_denom])
    return str(rate.quantize(RATE_PLACES))


def factor_ratio(native_factor: int, st_factor: int) -> Decimal:
    """native factor / stToken factor, exact to DECIMAL_CONTEXT: the rate the factors encode at any scale."""
    return DECIMAL_CONTEXT.divide(Decimal(native_factor), Decimal(st_factor))


def rate_gap_pct(stride_rate: str, rate: str | None) -> str | None:
    """How far the pool's rate sits below Stride's, as a percentage of Stride's rate, four places."""
    if rate is None:
        return None
    gap = (
        DECIMAL_CONTEXT.divide(
            Decimal(stride_rate) - Decimal(rate), Decimal(stride_rate)
        )
        * 100
    )
    return str(gap.quantize(GAP_PLACES))


def build_checks(
    raw: RawPool,
    st: StAsset,
    zone: ZoneDenoms,
    kind: PoolKind,
    route_lookup: RouteLookup | None,
    rate: str | None,
    stride_rate: str,
    needed: int,
    vault_shares: int,
) -> list[Check]:
    """The gate checks of check_transmuter_pool.py, each n/a (ok None) when the lookup it needs failed.

    The uniqueness check passes here, where only one pool is in view; `allocate` fails it on the duplicates.
    """
    st_factor = raw.assets.get(st.denom)
    native_factor = raw.assets.get(zone.osmosis_denom)
    alloyed_factor = raw.assets.get(raw.alloyed_denom)
    assets = set(raw.assets) - {raw.alloyed_denom}
    route = route_lookup.route if route_lookup else None
    corrupted = raw.corrupted
    limiters = raw.limiters
    candidate = raw.admin_candidate

    return [
        Check(
            name=CheckName.CODE_ID,
            ok=_known(raw.code_id, lambda: raw.code_id == TRANSMUTER_CODE_ID),
            detail=f"code_id={raw.code_id or '?'}",
        ),
        Check(
            name=CheckName.CW2_VERSION,
            ok=_known(raw.cw2_version, lambda: raw.cw2_version == TRANSMUTER_VERSION),
            detail=f"version={raw.cw2_version or '?'}",
        ),
        Check(
            name=CheckName.ASSETS,
            ok=assets == {st.denom, zone.osmosis_denom},
            detail=f"{len(assets)} assets; unexpected: {sorted(assets - {st.denom, zone.osmosis_denom})}",
        ),
        Check(
            name=CheckName.FACTORS,
            ok=_factors_ok(
                st_factor=st_factor,
                native_factor=native_factor,
                alloyed_factor=alloyed_factor,
                rate=rate,
            ),
            detail=f"stToken={st_factor} native={native_factor} alloyed={alloyed_factor}: rate {rate or '?'}",
        ),
        Check(
            name=CheckName.RATE,
            ok=None if rate is None else Decimal(rate) <= Decimal(stride_rate),
            detail=f"pool {rate or '?'} vs Stride {stride_rate} (gap {rate_gap_pct(stride_rate=stride_rate, rate=rate) or '?'}%)",
        ),
        Check(
            name=CheckName.ADMIN,
            ok=_known(raw.admin, lambda: raw.admin == config.OSMOSIS_VAULT),
            detail=raw.admin or "?",
        ),
        Check(
            name=CheckName.MODERATOR,
            ok=_known(raw.moderator, lambda: raw.moderator == config.POOL_MODERATOR),
            detail=raw.moderator or "?",
        ),
        Check(
            name=CheckName.NO_ADMIN_TRANSFER,
            ok=None if candidate is None else candidate.address is None,
            detail=""
            if candidate is None or candidate.address is None
            else f"candidate {candidate.address}",
        ),
        Check(name=CheckName.ACTIVE, ok=raw.is_active, detail=""),
        Check(
            name=CheckName.NO_LIMITERS,
            ok=None if limiters is None else not limiters,
            detail="" if limiters is None else ", ".join(limiters),
        ),
        _st_trace_check(st=st, zone=zone, kind=kind, route_lookup=route_lookup),
        Check(
            name=CheckName.NATIVE_TRACE,
            ok=zone.osmosis_denom in raw.assets,
            detail=f"expects {zone.native_trace}",
        ),
        _headroom_check(
            factors=[raw.assets[denom] for denom in assets],
            native_factor=native_factor,
            needed=needed,
        ),
        Check(
            name=CheckName.CORRUPTED,
            ok=None
            if corrupted is None
            else _corrupted_set_ok(
                corrupted=corrupted,
                native_denom=zone.osmosis_denom,
                vault_shares=vault_shares,
            ),
            detail=f"corrupted={corrupted if corrupted is not None else '?'} vault_shares={vault_shares}",
        ),
        Check(
            name=CheckName.UNIQUE,
            ok=True,
            detail=pool_slot(kind=kind, route=route) or "claims no slot",
        ),
    ]


def _factors_ok(
    st_factor: int | None, native_factor: int | None, alloyed_factor: int | None, rate: str | None
) -> bool:
    """The factors encode the reported rate exactly, at whatever scale, and the alloyed asset tracks the native."""
    if st_factor is None or native_factor is None or rate is None:
        return False
    exact = factor_ratio(native_factor=native_factor, st_factor=st_factor) == Decimal(rate)
    return exact and native_factor == alloyed_factor


def _headroom_check(factors: list[int], native_factor: int | None, needed: int) -> Check:
    """The transmuter normalises every balance to the lcm of the pool's factors (amount x lcm / factor) in a Uint128 on
    each join, swap and exit, so the amount the pool is meant to hold (at most the zone's `needed`) must fit once
    scaled. An 18-decimal supply with 1e18-scaled factors does not; 1e6-scaled factors leave room."""
    if native_factor is None:
        return Check(name=CheckName.HEADROOM, ok=None, detail=f"needed {needed:.3e}; native factor unknown")

    lcm = math.lcm(*factors)
    normalised = needed * (lcm // native_factor)
    detail = (
        f"needed {needed:.3e} x (lcm {lcm:.3e} / native factor {native_factor}) = {normalised:.3e} "
        f"vs 2^128 {UINT128_LIMIT:.3e}"
    )
    if normalised < UINT128_LIMIT:
        return Check(name=CheckName.HEADROOM, ok=True, detail=detail)
    advice = "use factors scaled to 1e6 (stToken 1000000, native floor(rate x 1e6), never rounded up) at creation"
    return Check(name=CheckName.HEADROOM, ok=False, detail=f"{detail}: {advice}")


# ---- pure logic: per zone


def build_zone_pools(
    zone: config.ZoneConfig,
    host: HostZoneInfo,
    denoms: ZoneDenoms,
    vault_native: int,
    fee_reserve: int,
    reports: list[PoolReport],
    planned: list[PlannedPool],
    creation_fee: CreationFee | None,
    vault_fee_balance: int | None,
) -> ZonePools:
    needed = host.needed
    pools_native = sum(report.native_balance for report in reports)
    native_on_osmosis = vault_native + pools_native

    served = {report.route.stride_channel for report in reports if report.route}
    blocked_channels = {plan.stride_channel for plan in planned if plan.blocked and plan.stride_channel}
    missing_routes = sorted(
        config.REQUIRED_ROUTES[denoms.st_denom] - served - blocked_channels, key=_channel_number
    )
    ordered = sorted(reports, key=_pool_order)
    pools_ready = _pools_ready(reports=ordered, missing_routes=missing_routes)

    # The voting-week gates: every route denom seeded, the canonical denom seeded, every planned pool created, and
    # the vault able to pay the creation fee for what is still to create. An unresolved route keeps them unknown.
    ordered_planned = sorted(planned, key=_planned_order)
    # A blocked pool (a holder chain we cannot seed from yet) waits outside every gate until the block is lifted.
    open_planned = [plan for plan in ordered_planned if not plan.blocked]
    route_plans = [plan for plan in open_planned if plan.kind == PoolKind.ROUTE]
    canonical_plans = [plan for plan in open_planned if plan.kind == PoolKind.CANONICAL]
    to_create = sum(1 for plan in open_planned if plan.live_contract is None)
    return ZonePools(
        chain_id=zone.chain_id,
        symbol=zone.symbol,
        decimals=zone.decimals,
        st_denom=denoms.st_denom,
        osmosis_denom=denoms.osmosis_denom,
        host_denom=host.host_denom,
        stride_rate=host.stride_rate,
        st_supply=host.st_supply,
        needed=needed,
        vault_native=vault_native,
        fee_reserve=fee_reserve,
        pools_native=pools_native,
        native_on_osmosis=native_on_osmosis,
        coverage=funds.coverage_ratio(covered=native_on_osmosis, needed=needed),
        missing_routes=missing_routes,
        pools=ordered,
        pools_ready=pools_ready,
        pools_funded=_pools_funded(reports=ordered, pools_ready=pools_ready),
        planned=ordered_planned,
        routes_seeded=tri_state(outcomes=[plan.seeded for plan in route_plans]),
        canonical_seeded=tri_state(outcomes=[plan.seeded for plan in canonical_plans]),
        pools_created=tri_state(
            outcomes=[None if plan.error else plan.live_contract is not None for plan in open_planned]
        ),
        creation_fee=creation_fee,
        vault_fee_balance=vault_fee_balance,
        creation_fee_short=creation_fee_short(fee=creation_fee, vault_balance=vault_fee_balance, to_create=to_create),
    )


def _pools_ready(reports: list[PoolReport], missing_routes: list[str]) -> bool | None:
    """False on a structural gap or a failing check; None while every failure is only an unknown; else True."""
    has_canonical = any(report.kind == PoolKind.CANONICAL for report in reports)
    outcomes = [check.ok for report in reports for check in report.checks]
    if (
        not has_canonical
        or missing_routes
        or any(outcome is False for outcome in outcomes)
    ):
        return False
    return None if any(outcome is None for outcome in outcomes) else True


def _pools_funded(reports: list[PoolReport], pools_ready: bool | None) -> bool | None:
    outcomes = (
        [pools_ready]
        + [report.funded_exactly for report in reports]
        + [report.native_marked for report in reports]
    )
    return tri_state(outcomes=outcomes)


def _pool_order(report: PoolReport) -> tuple[int, int, str]:
    rank = {PoolKind.CANONICAL: 0, PoolKind.ROUTE: 1, PoolKind.UNRECOGNISED: 2}[
        report.kind
    ]
    channel = _channel_number(report.route.stride_channel) if report.route else 0
    return (rank, channel, report.contract)


def _planned_order(plan: PlannedPool) -> tuple[int, int]:
    channel = _channel_number(plan.stride_channel) if plan.stride_channel else 0
    return (plan.kind != PoolKind.CANONICAL, channel)


def _kind(st: StAsset, zone: ZoneDenoms, route_lookup: RouteLookup | None) -> PoolKind:
    if st.trace.full == zone.canonical_st_trace:
        return PoolKind.CANONICAL
    if route_lookup is None:
        # Either the path is not two hops (unrecognised) or the lookup failed (still a route, its check n/a).
        return (
            PoolKind.UNRECOGNISED
            if route_hops(trace=st.trace) is None
            else PoolKind.ROUTE
        )
    return PoolKind.ROUTE if route_lookup.route else PoolKind.UNRECOGNISED


def _st_trace_check(
    st: StAsset, zone: ZoneDenoms, kind: PoolKind, route_lookup: RouteLookup | None
) -> Check:
    if kind == PoolKind.CANONICAL:
        return Check(
            name=CheckName.ST_TRACE, ok=True, detail=f"canonical: {st.trace.full}"
        )

    hops = route_hops(trace=st.trace)
    if hops is None:
        return Check(
            name=CheckName.ST_TRACE,
            ok=False,
            detail=f"not a two-hop path: {st.trace.full}",
        )
    if route_lookup is None:
        return Check(
            name=CheckName.ST_TRACE,
            ok=None,
            detail=f"route lookup failed: {st.trace.full}",
        )
    if route_lookup.route is None:
        return Check(
            name=CheckName.ST_TRACE,
            ok=False,
            detail=f"no policy channel of {zone.st_denom} reaches {route_lookup.chain_id} via {hops.holder_channel}",
        )
    return Check(
        name=CheckName.ST_TRACE,
        ok=True,
        detail=f"{route_lookup.chain_id} via Stride {route_lookup.route.stride_channel}: {st.trace.full}",
    )


def _corrupted_set_ok(
    corrupted: list[str], native_denom: str, vault_shares: int
) -> bool:
    """Nothing corrupted before funding; exactly the native token once the vault holds shares (the one-way mark)."""
    if vault_shares == 0:
        return not corrupted
    return corrupted == [native_denom]


def _known(value: Any, verdict: Callable[[], bool]) -> bool | None:
    """A check outcome that is n/a when the value it judges could not be read."""
    return None if value is None else verdict()


def _channel_number(channel_id: str) -> int:
    return int(channel_id.rsplit("-", 1)[1])


# ---- pure logic: planned pools


def plan_route(
    stride_channel: str, target: ChannelTarget | None, holder_targets: dict[str, ChannelTarget | None]
) -> PlannedRoute:
    """Resolve a policy channel: the holder chain X behind Stride's channel, then Osmosis's channel to X, which is the
    `config.HOLDER_ROUTES` entry whose channel's client tracks X (`holder_targets`), and its counterparty back."""
    if target is None:
        return _unresolved_route(
            stride_channel=stride_channel,
            error=f"Stride {stride_channel} could not be resolved (channel, connection or client lookup failed)",
        )
    if target.counterparty_channel is None:
        return _unresolved_route(
            stride_channel=stride_channel, error=f"Stride {stride_channel} has no counterparty channel yet"
        )

    holder = config.HOLDER_CHAINS.get(target.chain_id)
    osmosis_channel = next(
        (
            channel
            for channel, hop in holder_targets.items()
            if hop is not None and hop.chain_id == target.chain_id and hop.counterparty_channel is not None
        ),
        None,
    )
    error = None
    if holder is None:
        error = f"{target.chain_id} is not in config.HOLDER_CHAINS"
    elif osmosis_channel is None:
        error = f"no channel in config.HOLDER_ROUTES tracks {target.chain_id} on Osmosis"
    return PlannedRoute(
        stride_channel=stride_channel,
        holder_chain_id=target.chain_id,
        holder=holder,
        counterparty_channel=target.counterparty_channel,
        osmosis_channel=osmosis_channel,
        holder_to_osmosis_channel=holder_targets[osmosis_channel].counterparty_channel if osmosis_channel else None,
        error=error,
    )


def _unresolved_route(stride_channel: str, error: str) -> PlannedRoute:
    return PlannedRoute(
        stride_channel=stride_channel,
        holder_chain_id=None,
        holder=None,
        counterparty_channel=None,
        osmosis_channel=None,
        holder_to_osmosis_channel=None,
        error=error,
    )


def route_denoms(route: PlannedRoute, st_denom: str) -> RouteDenoms:
    """The stToken's denom on the holder chain (`transfer/<X -> Stride>/<st_denom>`) and on Osmosis (that path behind
    `transfer/<Osmosis -> X>/`), each None while its channels are unresolved."""
    if route.counterparty_channel is None:
        return RouteDenoms(on_holder=None, on_osmosis=None)
    holder_trace = f"{chain.TRANSFER_PORT}/{route.counterparty_channel}/{st_denom}"
    on_osmosis = (
        None
        if route.osmosis_channel is None
        else chain.ibc_denom(path=f"{chain.TRANSFER_PORT}/{route.osmosis_channel}/{holder_trace}")
    )
    return RouteDenoms(on_holder=chain.ibc_denom(path=holder_trace), on_osmosis=on_osmosis)


def creation_factors(decimals: int, stride_rate: str) -> Factors:
    """1e18-scaled for a six-decimal zone (exact: the rate has 18 places); 1e6-scaled for an 18-decimal zone, whose
    supply overflows the transmuter's Uint128 at 1e18. The native factor is rounded down either way, so the pool
    never prices a stToken above Stride's rate."""
    scale = ST_FACTOR_SIX_DECIMALS if decimals == SIX_DECIMALS else ST_FACTOR_EIGHTEEN_DECIMALS
    product = DECIMAL_CONTEXT.multiply(Decimal(stride_rate), Decimal(scale))
    return Factors(st_factor=scale, native_factor=int(product.to_integral_value(rounding=decimal.ROUND_FLOOR)))


def alloyed_subdenom(st_symbol: str, route: PlannedRoute | None, shared_chain: bool) -> str:
    """`stATOM` for the canonical pool, `stATOM.cosmoshub` for a route, `stATOM.axelar.channel11` when two policy
    channels reach the same chain (`shared_chain`); an unresolved route is named by its Stride channel alone."""
    if route is None:
        return st_symbol
    if route.holder is None:
        return f"{st_symbol}.{_dashless(route.stride_channel)}"
    if shared_chain:
        return f"{st_symbol}.{route.holder.name}.{_dashless(route.stride_channel)}"
    return f"{st_symbol}.{route.holder.name}"


def instantiate_message(denom_on_osmosis: str, native_denom: str, factors: Factors, subdenom: str) -> dict[str, Any]:
    """The transmuter instantiate message (docs/wind-down/transmuter.md): the vault is admin, config.POOL_MODERATOR the
    moderator, and the alloyed asset tracks the native factor so one alloyed unit is one native base unit of pool value."""
    return {
        "pool_asset_configs": [
            {"denom": denom_on_osmosis, "normalization_factor": str(factors.st_factor)},
            {"denom": native_denom, "normalization_factor": str(factors.native_factor)},
        ],
        "alloyed_asset_subdenom": subdenom,
        "alloyed_asset_normalization_factor": str(factors.native_factor),
        "admin": config.OSMOSIS_VAULT,
        "moderator": config.POOL_MODERATOR,
    }


def seed_command(holder: config.HolderChain, holder_chain_id: str, channel: str, denom: str, decimals: int) -> str:
    """A single-signer IBC transfer of 0.01 stToken from the holder chain to the test wallet on Osmosis: any non-zero
    amount gives the denom supply there and funds the test join of that route's pool."""
    amount = 10 ** (decimals - SEED_AMOUNT_PLACES)
    return (
        f"{holder.binary} tx ibc-transfer transfer {chain.TRANSFER_PORT} {channel} {config.POOL_SEED_ADDRESS} "
        f"{amount}{denom} "
        f"--from <KEY_ON_{holder.name.upper()}> --chain-id {holder_chain_id} --node {holder.node} "
        f"--gas auto --gas-adjustment 1.5 --fees <FEES>"
    )


def build_canonical_plan(
    zone: config.ZoneConfig, denoms: ZoneDenoms, stride_rate: str, supply: int | None, reports: list[PoolReport]
) -> PlannedPool:
    """The canonical pool: its stToken comes straight from Stride over channel-5, so the seed (shown only while the
    denom has no supply) is a `strided` transfer from any Stride key."""
    denom_on_osmosis = chain.ibc_denom(path=denoms.canonical_st_trace)
    factors = creation_factors(decimals=zone.decimals, stride_rate=stride_rate)
    subdenom = alloyed_subdenom(st_symbol=st_symbol(zone=zone), route=None, shared_chain=False)
    seeded = None if supply is None else supply > 0
    seed = (
        seed_command(
            holder=config.STRIDE_HOLDER,
            holder_chain_id=config.STRIDE_CHAIN_ID,
            channel=config.STRIDE_CHANNEL_TO_OSMOSIS,
            denom=denoms.st_denom,
            decimals=zone.decimals,
        )
        if seeded is False
        else None
    )
    return PlannedPool(
        kind=PoolKind.CANONICAL,
        stride_channel=None,
        holder_chain_id=config.STRIDE_CHAIN_ID,
        holder_name=config.STRIDE_HOLDER.name,
        holder_binary=config.STRIDE_HOLDER.binary,
        holder_node=config.STRIDE_HOLDER.node,
        counterparty_channel=None,
        osmosis_channel=funds.STRIDE_CHANNEL_ON_OSMOSIS,
        holder_to_osmosis_channel=config.STRIDE_CHANNEL_TO_OSMOSIS,
        denom_on_holder=None,
        denom_on_osmosis=denom_on_osmosis,
        seeded=seeded,
        test_wallet_balance=supply,
        st_factor=factors.st_factor,
        native_factor=factors.native_factor,
        alloyed_subdenom=subdenom,
        instantiate_msg=instantiate_message(
            denom_on_osmosis=denom_on_osmosis, native_denom=denoms.osmosis_denom, factors=factors, subdenom=subdenom
        ),
        seed_command=seed,
        live_contract=live_contract_of(denom=denom_on_osmosis, reports=reports),
        error=None,
    )


def build_route_plan(
    zone: config.ZoneConfig,
    denoms: ZoneDenoms,
    stride_rate: str,
    route: PlannedRoute,
    subdenom: str,
    supply: int | None,
    reports: list[PoolReport],
) -> PlannedPool:
    """A route pool; an unresolved route keeps whatever did resolve and carries the error instead of commands."""
    route_denom = route_denoms(route=route, st_denom=denoms.st_denom)
    factors = creation_factors(decimals=zone.decimals, stride_rate=stride_rate)
    holder = route.holder
    resolved = route.error is None and holder is not None and route_denom.on_osmosis is not None
    seeded = None if supply is None else supply > 0
    return PlannedPool(
        kind=PoolKind.ROUTE,
        stride_channel=route.stride_channel,
        holder_chain_id=route.holder_chain_id,
        holder_name=holder.name if holder else None,
        holder_binary=holder.binary if holder else None,
        holder_node=holder.node if holder else None,
        counterparty_channel=route.counterparty_channel,
        osmosis_channel=route.osmosis_channel,
        holder_to_osmosis_channel=route.holder_to_osmosis_channel,
        denom_on_holder=route_denom.on_holder,
        denom_on_osmosis=route_denom.on_osmosis,
        seeded=seeded,
        test_wallet_balance=supply,
        st_factor=factors.st_factor,
        native_factor=factors.native_factor,
        alloyed_subdenom=subdenom,
        instantiate_msg=instantiate_message(
            denom_on_osmosis=route_denom.on_osmosis,
            native_denom=denoms.osmosis_denom,
            factors=factors,
            subdenom=subdenom,
        )
        if resolved
        else None,
        seed_command=seed_command(
            holder=holder,
            holder_chain_id=route.holder_chain_id,
            channel=route.holder_to_osmosis_channel,
            denom=route_denom.on_holder,
            decimals=zone.decimals,
        )
        if resolved
        else None,
        live_contract=live_contract_of(denom=route_denom.on_osmosis, reports=reports),
        error=route.error,
        blocked=config.BLOCKED_HOLDER_CHAINS.get(route.holder_chain_id) if route.holder_chain_id else None,
    )


def live_contract_of(denom: str | None, reports: list[PoolReport]) -> str | None:
    """The existing pool whose stToken is `denom`: the planned pool is created once one is."""
    if denom is None:
        return None
    return next((report.contract for report in reports if report.st_denom == denom), None)


def st_symbol(zone: config.ZoneConfig) -> str:
    """The stToken's symbol as the Funds tab shows it: `st` + the zone's symbol (stATOM)."""
    return f"{ST_PREFIX}{zone.symbol}"


def tri_state(outcomes: list[bool | None]) -> bool | None:
    """False when any outcome is False, None while any is unknown, else True (vacuously on nothing)."""
    if any(outcome is False for outcome in outcomes):
        return False
    return None if any(outcome is None for outcome in outcomes) else True


def creation_fee_short(fee: CreationFee | None, vault_balance: int | None, to_create: int) -> bool | None:
    """Whether the vault cannot pay the creation fee for every pool still to create; None while either is unknown."""
    if fee is None or vault_balance is None:
        return None
    return vault_balance < fee.amount * to_create


def _dashless(channel_id: str) -> str:
    # The subdenoms use dots as their only separator (`stATOM.axelar.channel11`), so the channel loses its dash.
    return channel_id.replace("-", "")


# ---- per zone: network


def _zone_pools(
    zone: config.ZoneConfig,
    host_zone: dict[str, Any],
    stride: chain.Chain,
    osmosis: chain.Chain,
    snapshot: OsmosisSnapshot,
) -> ZonePools:
    denoms = zone_denoms(chain_id=zone.chain_id, host_denom=host_zone["host_denom"])
    host = HostZoneInfo(
        host_denom=host_zone["host_denom"],
        stride_rate=host_zone["redemption_rate"],
        st_supply=_supply(chain_handle=stride, denom=denoms.st_denom),
    )
    assigned = [
        (raw, st) for raw in snapshot.pools if (st := st_asset_of(raw=raw, zone=denoms))
    ]
    reports = [
        _pool_report(
            raw=raw,
            st=st,
            denoms=denoms,
            host=host,
            stride=stride,
            osmosis=osmosis,
            snapshot=snapshot,
        )
        for raw, st in assigned
    ]
    vault_native = snapshot.vault_balances.get(denoms.osmosis_denom, 0)
    reserve = fee_reserve(chain_id=zone.chain_id)
    allocated = allocate(reports=reports, vault_native=vault_native, fee_reserve=reserve)

    # What the zone should have, matched against what it has: a planned pool with a live contract is created.
    planned = _planned_pools(
        zone=zone, denoms=denoms, host=host, stride=stride, osmosis=osmosis, snapshot=snapshot, reports=allocated
    )
    fee = snapshot.creation_fee
    return build_zone_pools(
        zone=zone,
        host=host,
        denoms=denoms,
        vault_native=vault_native,
        fee_reserve=reserve,
        reports=allocated,
        planned=planned,
        creation_fee=fee,
        vault_fee_balance=None if fee is None else snapshot.vault_balances.get(fee.denom, 0),
    )


def _pool_report(
    raw: RawPool,
    st: StAsset,
    denoms: ZoneDenoms,
    host: HostZoneInfo,
    stride: chain.Chain,
    osmosis: chain.Chain,
    snapshot: OsmosisSnapshot,
) -> PoolReport:
    # The route and its escrow are optional lookups: a failure leaves the pool's trace check and allocation n/a.
    hops = (
        None
        if st.trace.full == denoms.canonical_st_trace
        else route_hops(trace=st.trace)
    )
    route_lookup = (
        chain.optional(
            lambda: _resolve_route(
                stride=stride, osmosis=osmosis, hops=hops, st_denom=denoms.st_denom
            )
        )
        if hops
        else None
    )
    route = route_lookup.route if route_lookup else None
    escrow = (
        chain.optional(
            lambda: _escrow_balance(
                stride=stride, channel_id=route.stride_channel, denom=denoms.st_denom
            )
        )
        if route
        else None
    )
    return build_pool_report(
        raw=raw,
        st=st,
        zone=denoms,
        stride_rate=host.stride_rate,
        needed=host.needed,
        vault_shares=snapshot.vault_balances.get(raw.alloyed_denom, 0),
        test_wallet_shares=snapshot.test_wallet_balances.get(raw.alloyed_denom, 0),
        route_lookup=route_lookup,
        escrow=escrow,
    )


def _resolve_route(
    stride: chain.Chain, osmosis: chain.Chain, hops: RouteHops, st_denom: str
) -> RouteLookup:
    holder_chain_id = _osmosis_channel_chain_id(osmosis=osmosis, channel_id=hops.osmosis_channel)
    policy = config.REQUIRED_ROUTES[st_denom]
    targets = {
        channel: _channel_target(chain_handle=stride, channel_id=channel)
        for channel in policy
    }
    return RouteLookup(
        chain_id=holder_chain_id,
        route=match_route(
            holder_chain_id=holder_chain_id,
            holder_channel=hops.holder_channel,
            policy_channels=policy,
            targets=targets,
        ),
    )


@functools.lru_cache(maxsize=256)
def _osmosis_channel_chain_id(osmosis: chain.Chain, channel_id: str) -> str:
    """The chain an Osmosis transfer channel's client tracks; fixed once the channel is open, so cached for the process."""
    response = chain.rest_get(
        chain=osmosis,
        path=f"/ibc/core/channel/v1/channels/{channel_id}/ports/{chain.TRANSFER_PORT}/client_state",
    )
    return response["identified_client_state"]["client_state"]["chain_id"]


@functools.lru_cache(maxsize=256)
def _channel_target(chain_handle: chain.Chain, channel_id: str) -> ChannelTarget:
    """A transfer channel's counterparty and the chain its connection's client tracks (Stride's policy channels and
    Osmosis's holder channels alike); fixed once open, so cached for the process."""
    end = chain.ibc_channel_end(chain=chain_handle, channel_id=channel_id, port_id=chain.TRANSFER_PORT)
    connection = chain.ibc_connection_end(chain=chain_handle, connection_id=end.connection_id)
    return ChannelTarget(
        counterparty_channel=end.counterparty_channel,
        chain_id=chain.ibc_client_chain_id(chain=chain_handle, client_id=connection.client_id),
    )


def _escrow_balance(stride: chain.Chain, channel_id: str, denom: str) -> int:
    """Stride's bank balance of `denom` on the channel's ICS-20 escrow account: the stTokens held on that route."""
    response = chain.rest_get(
        chain=stride,
        path=f"/cosmos/bank/v1beta1/balances/{chain.escrow_address(channel_id=channel_id)}/by_denom",
        params={"denom": denom},
    )
    return int(response["balance"]["amount"])


def _planned_pools(
    zone: config.ZoneConfig,
    denoms: ZoneDenoms,
    host: HostZoneInfo,
    stride: chain.Chain,
    osmosis: chain.Chain,
    snapshot: OsmosisSnapshot,
    reports: list[PoolReport],
) -> list[PlannedPool]:
    """The canonical pool, then one per policy channel by channel number; every lookup here is optional (a failure
    leaves that pool's field n/a or carries its error) and the channel resolutions are cached for the process."""
    routes = [
        plan_route(
            stride_channel=channel,
            target=_optional_target(chain_handle=stride, channel_id=channel),
            holder_targets=snapshot.holder_targets,
        )
        for channel in sorted(config.REQUIRED_ROUTES[denoms.st_denom], key=_channel_number)
    ]

    # Two policy channels to one chain (Terra) need the channel in their subdenoms to tell the pools apart.
    chains = collections.Counter(route.holder_chain_id for route in routes if route.holder_chain_id)
    symbol = st_symbol(zone=zone)
    canonical = build_canonical_plan(
        zone=zone,
        denoms=denoms,
        stride_rate=host.stride_rate,
        supply=_optional_supply(osmosis=osmosis, denom=chain.ibc_denom(path=denoms.canonical_st_trace)),
        reports=reports,
    )
    planned_routes = [
        build_route_plan(
            zone=zone,
            denoms=denoms,
            stride_rate=host.stride_rate,
            route=route,
            subdenom=alloyed_subdenom(st_symbol=symbol, route=route, shared_chain=chains[route.holder_chain_id] > 1),
            supply=_optional_supply(
                osmosis=osmosis, denom=route_denoms(route=route, st_denom=denoms.st_denom).on_osmosis
            ),
            reports=reports,
        )
        for route in routes
    ]
    return [canonical, *planned_routes]


def _optional_target(chain_handle: chain.Chain, channel_id: str) -> ChannelTarget | None:
    return chain.optional(lambda: _channel_target(chain_handle=chain_handle, channel_id=channel_id))


def _optional_supply(osmosis: chain.Chain, denom: str | None) -> int | None:
    """The test wallet's balance of a denom on Osmosis: whether it is seeded (and can fund the test join); None for an
    unresolved denom or a failed lookup."""
    if denom is None:
        return None
    return chain.optional(lambda: _balance_of(chain_handle=osmosis, address=config.POOL_SEED_ADDRESS, denom=denom))


# ---- once per collect: Osmosis


def _osmosis_snapshot(osmosis: chain.Chain) -> OsmosisSnapshot:
    vault_balances = _balances(chain_handle=osmosis, address=config.OSMOSIS_VAULT)
    listed = _listed_pools(osmosis=osmosis)
    candidates = transmuter_candidates(listed=listed, skip=_NOT_OURS)
    contracts = list(dict.fromkeys(candidates + list(config.EXTRA_POOL_CONTRACTS)))

    # One get_admin per candidate; a failed answer is unknown (not memoised, and an extra still shows with admin n/a).
    with concurrent.futures.ThreadPoolExecutor(max_workers=POOL_WORKERS) as pool:
        answers = pool.map(
            lambda contract: chain.optional(
                lambda: _admin(osmosis=osmosis, contract=contract)
            ),
            contracts,
        )
        admins = dict(zip(contracts, answers))
    _NOT_OURS.update(
        contract
        for contract, admin in admins.items()
        if admin is not None
        and admin != config.OSMOSIS_VAULT
        and contract not in config.EXTRA_POOL_CONTRACTS
    )

    ours = select_our_pools(admins=admins, extra=config.EXTRA_POOL_CONTRACTS)
    with concurrent.futures.ThreadPoolExecutor(max_workers=POOL_WORKERS) as pool:
        pools = list(
            pool.map(
                lambda contract: _raw_pool(
                    osmosis=osmosis,
                    contract=contract,
                    listed=listed.get(contract),
                    admin=admins[contract],
                ),
                ours,
            )
        )
    return OsmosisSnapshot(
        vault_balances=vault_balances,
        test_wallet_balances=_balances(chain_handle=osmosis, address=config.POOL_SEED_ADDRESS),
        pools=pools,
        creation_fee=_creation_fee(osmosis=osmosis),
        holder_targets=_holder_targets(osmosis=osmosis),
    )


def _holder_targets(osmosis: chain.Chain) -> dict[str, ChannelTarget | None]:
    """What Osmosis's channel to each holder chain points at; a failed lookup is None (that route gets an error)."""
    return {
        route.osmosis_channel: _optional_target(chain_handle=osmosis, channel_id=route.osmosis_channel)
        for route in config.HOLDER_ROUTES
    }


def _creation_fee(osmosis: chain.Chain) -> CreationFee | None:
    """The poolmanager's pool_creation_fee. Polkachu's REST answers "Not Implemented" for the lower-case `params`
    path and serves the capitalised one, so both are tried; None when neither answers."""
    for path in POOLMANAGER_PARAMS_PATHS:
        fee = chain.optional(lambda: _pool_creation_fee(osmosis=osmosis, path=path))
        if fee is not None:
            return fee
    return None


def _pool_creation_fee(osmosis: chain.Chain, path: str) -> CreationFee | None:
    coins = chain.rest_get(chain=osmosis, path=path)["params"]["pool_creation_fee"]
    if not coins:
        return None
    return CreationFee(denom=coins[0]["denom"], amount=int(coins[0]["amount"]))


def _listed_pools(osmosis: chain.Chain) -> dict[str, ListedPool]:
    entries = chain.rest_get_all_pages(
        chain=osmosis, path=COSMWASMPOOL_LISTING_PATH, key="pools"
    )
    return {
        entry["contract_address"]: ListedPool(
            pool_id=entry["pool_id"], code_id=entry["code_id"]
        )
        for entry in entries
    }


def _admin(osmosis: chain.Chain, contract: str) -> str:
    return _smart_query(osmosis=osmosis, contract=contract, query={"get_admin": {}})[
        "admin"
    ]


def _raw_pool(
    osmosis: chain.Chain, contract: str, listed: ListedPool | None, admin: str | None
) -> RawPool:
    def query(message: dict[str, Any]) -> dict[str, Any]:
        return _smart_query(osmosis=osmosis, contract=contract, query=message)

    configs = query({"list_asset_configs": {}})["asset_configs"]
    assets = {entry["denom"]: int(entry["normalization_factor"]) for entry in configs}
    alloyed_denom = query({"get_share_denom": {}})["share_denom"]
    liquidity = {
        coin["denom"]: int(coin["amount"])
        for coin in query({"get_total_pool_liquidity": {}})["total_pool_liquidity"]
    }
    traces = {
        denom: _denom_trace(rest=osmosis.rest, denom=denom)
        for denom in assets
        if denom.startswith(funds.IBC_PREFIX)
    }

    # An extra the listing does not hold still gets its code id from the contract info.
    code_id = (
        listed.code_id
        if listed
        else chain.optional(
            lambda: _contract_code_id(osmosis=osmosis, contract=contract)
        )
    )
    return RawPool(
        contract=contract,
        pool_id=listed.pool_id if listed else None,
        code_id=code_id,
        admin=admin,
        alloyed_denom=alloyed_denom,
        assets=assets,
        liquidity=liquidity,
        traces=traces,
        alloyed_supply=chain.optional(
            lambda: _supply(chain_handle=osmosis, denom=alloyed_denom)
        ),
        cw2_version=chain.optional(
            lambda: _cw2_version(osmosis=osmosis, contract=contract)
        ),
        moderator=chain.optional(lambda: query({"get_moderator": {}})["moderator"]),
        admin_candidate=chain.optional(
            lambda: AdminCandidate(
                address=query({"get_admin_candidate": {}})["admin_candidate"]
            )
        ),
        is_active=chain.optional(lambda: bool(query({"is_active": {}})["is_active"])),
        limiters=chain.optional(
            lambda: [
                f"{entry[0][0]}/{entry[0][1]}"
                for entry in query({"list_limiters": {}})["limiters"]
            ]
        ),
        corrupted=chain.optional(
            lambda: list(query({"get_corrupted_denoms": {}})["corrupted_denoms"])
        ),
    )


def _contract_code_id(osmosis: chain.Chain, contract: str) -> str:
    response = chain.rest_get(
        chain=osmosis, path=f"/cosmwasm/wasm/v1/contract/{contract}"
    )
    return str(response["contract_info"]["code_id"])


def _cw2_version(osmosis: chain.Chain, contract: str) -> str:
    key = base64.b64encode(CW2_INFO_KEY).decode()
    response = chain.rest_get(
        chain=osmosis, path=f"/cosmwasm/wasm/v1/contract/{contract}/raw/{key}"
    )
    return json.loads(base64.b64decode(response["data"]))["version"]


def _smart_query(
    osmosis: chain.Chain, contract: str, query: dict[str, Any]
) -> dict[str, Any]:
    encoded = base64.b64encode(json.dumps(query).encode()).decode()
    return chain.rest_get(
        chain=osmosis, path=f"/cosmwasm/wasm/v1/contract/{contract}/smart/{encoded}"
    )["data"]


@functools.lru_cache(maxsize=1024)
def _denom_trace(rest: str, denom: str) -> DenomTrace:
    """The trace of an `ibc/` voucher; a trace never changes, so a hit is cached for the life of the process."""
    trace = chain.get_json(
        url=f"{rest}/ibc/apps/transfer/v1/denom_traces/{denom.removeprefix(funds.IBC_PREFIX)}"
    )
    return DenomTrace(
        path=trace["denom_trace"]["path"], base_denom=trace["denom_trace"]["base_denom"]
    )


# ---- helpers


def _balances(chain_handle: chain.Chain, address: str) -> dict[str, int]:
    balances = chain.rest_get_all_pages(
        chain=chain_handle,
        path=f"/cosmos/bank/v1beta1/balances/{address}",
        key="balances",
    )
    return {balance["denom"]: int(balance["amount"]) for balance in balances}


def _balance_of(chain_handle: chain.Chain, address: str, denom: str) -> int:
    response = chain.rest_get(
        chain=chain_handle,
        path=f"/cosmos/bank/v1beta1/balances/{address}/by_denom",
        params={"denom": denom},
    )
    return int(response["balance"]["amount"])


def _supply(chain_handle: chain.Chain, denom: str) -> int:
    response = chain.rest_get(
        chain=chain_handle,
        path="/cosmos/bank/v1beta1/supply/by_denom",
        params={"denom": denom},
    )
    return int(response["amount"]["amount"])


def _payload(zones: list[ZonePools | chain.ZoneError]) -> dict[str, Any]:
    return chain.stringify_ints({"zones": [dataclasses.asdict(zone) for zone in zones]})
