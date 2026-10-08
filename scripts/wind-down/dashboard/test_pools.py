import dataclasses
import importlib.util
import json
import pathlib
import re
import sys
import unittest
import urllib.error
from typing import Any
from unittest import mock

import chain
import config
import pools

WIND_DOWN_DIR = pathlib.Path(__file__).resolve().parent.parent
VAULT = config.OSMOSIS_VAULT
STRANGER = "osmo1stranger"
STRIDE_RATE = "1.5"
ST_FACTOR = 10**18  # the test pools use 1e18-scaled factors, as the six-decimal zones' real pools do
RATE_FACTOR = 1_500_000_000_000_000_000  # native factor encoding a 1.5 rate against the 1e18 stToken factor
NEEDED = 1_500_000  # the 1_000_000 stToken supply of the test zone at the 1.5 rate
# The haqq zone's order of magnitude: an 18-decimal supply at its rate, and the factors its pool would be created with.
HAQQ_NEEDED = 107_131_023 * 10**18
HAQQ_RATE_FACTOR = 1_060_496_560_022_296_837
HUB_NEEDED = 2_174_903 * 10**6
HUB_RATE_FACTOR = 2_013_525_450_106_978_250

ATOM_TRACE = pools.DenomTrace(path="transfer/channel-0", base_denom="uatom")
ATOM_ON_OSMOSIS = chain.ibc_denom(path=ATOM_TRACE.full)
CANONICAL_TRACE = pools.DenomTrace(path="transfer/channel-326", base_denom="stuatom")
CANONICAL_STATOM = chain.ibc_denom(path=CANONICAL_TRACE.full)
HUB_TRACE = pools.DenomTrace(
    path="transfer/channel-0/transfer/channel-391", base_denom="stuatom"
)
HUB_STATOM = chain.ibc_denom(path=HUB_TRACE.full)
SECRET_TRACE = pools.DenomTrace(
    path="transfer/channel-88/transfer/channel-37", base_denom="stuatom"
)
SECRET_STATOM = chain.ibc_denom(path=SECRET_TRACE.full)
INJ_TRACE = pools.DenomTrace(path="transfer/channel-326", base_denom="stinj")
CANONICAL_STINJ = chain.ibc_denom(path=INJ_TRACE.full)
OSMO_TRACE = pools.DenomTrace(path="transfer/channel-326", base_denom="stuosmo")
CANONICAL_STUOSMO = chain.ibc_denom(path=OSMO_TRACE.full)

ATOM_ZONE = pools.zone_denoms(chain_id="cosmoshub-4", host_denom="uatom")
HUB_ROUTE = pools.Route(
    chain_id="cosmoshub-4",
    stride_channel="channel-0",
    counterparty_channel="channel-391",
)
HUB_LOOKUP = pools.RouteLookup(chain_id="cosmoshub-4", route=HUB_ROUTE)


def raw_pool(
    contract: str = "osmo1canon",
    st_denom: str = CANONICAL_STATOM,
    st_trace: pools.DenomTrace = CANONICAL_TRACE,
    native: int = 0,
    st_balance: int = 0,
    alloyed_supply: int | None = 0,
    **overrides: Any,
) -> pools.RawPool:
    """A green canonical stATOM pool before funding; override fields to break one check at a time."""
    alloyed_denom = f"factory/{contract}/alloyed/x"
    fields: dict[str, Any] = {
        "contract": contract,
        "pool_id": "3595",
        "code_id": pools.TRANSMUTER_CODE_ID,
        "admin": VAULT,
        "alloyed_denom": alloyed_denom,
        "assets": {
            st_denom: ST_FACTOR,
            ATOM_ON_OSMOSIS: RATE_FACTOR,
            alloyed_denom: RATE_FACTOR,
        },
        "liquidity": {st_denom: st_balance, ATOM_ON_OSMOSIS: native},
        "traces": {st_denom: st_trace, ATOM_ON_OSMOSIS: ATOM_TRACE},
        "alloyed_supply": alloyed_supply,
        "cw2_version": pools.TRANSMUTER_VERSION,
        "moderator": VAULT,
        "admin_candidate": pools.AdminCandidate(address=None),
        "is_active": True,
        "limiters": [],
        "corrupted": [],
    }
    return pools.RawPool(**{**fields, **overrides})


def osmo_pool(contract: str = "osmo1osmopool", native: int = 0) -> pools.RawPool:
    """A green canonical stOSMO pool: the osmosis-1 zone, whose native token is the bare uosmo."""
    alloyed_denom = f"factory/{contract}/alloyed/x"
    return raw_pool(
        contract=contract,
        st_denom=CANONICAL_STUOSMO,
        st_trace=OSMO_TRACE,
        assets={CANONICAL_STUOSMO: ST_FACTOR, "uosmo": RATE_FACTOR, alloyed_denom: RATE_FACTOR},
        liquidity={CANONICAL_STUOSMO: 0, "uosmo": native},
        traces={CANONICAL_STUOSMO: OSMO_TRACE},
    )


def scaled_pool(st_factor: int, native_factor: int, alloyed_factor: int | None = None) -> pools.RawPool:
    """The green canonical pool with its factors replaced: the alloyed factor follows the native unless given."""
    raw = raw_pool()
    return dataclasses.replace(
        raw,
        assets={
            CANONICAL_STATOM: st_factor,
            ATOM_ON_OSMOSIS: native_factor,
            raw.alloyed_denom: native_factor if alloyed_factor is None else alloyed_factor,
        },
    )


def report(
    raw: pools.RawPool,
    vault_shares: int = 0,
    route_lookup: pools.RouteLookup | None = None,
    escrow: int | None = None,
    stride_rate: str = STRIDE_RATE,
    needed: int = NEEDED,
) -> pools.PoolReport:
    st = pools.st_asset_of(raw=raw, zone=ATOM_ZONE)
    assert st is not None
    return pools.build_pool_report(
        raw=raw,
        st=st,
        zone=ATOM_ZONE,
        stride_rate=stride_rate,
        needed=needed,
        vault_shares=vault_shares,
        route_lookup=route_lookup,
        escrow=escrow,
    )


def check(pool: pools.PoolReport, name: pools.CheckName) -> pools.Check:
    return next(entry for entry in pool.checks if entry.name == name)


def zone_pools(
    reports: list[pools.PoolReport],
    vault_native: int = 0,
    st_supply: int = 1_000_000,
    planned: list[pools.PlannedPool] | None = None,
    creation_fee: pools.CreationFee | None = None,
    vault_fee_balance: int | None = None,
) -> pools.ZonePools:
    return pools.build_zone_pools(
        zone=config.ZONES_BY_CHAIN_ID["cosmoshub-4"],
        host=pools.HostZoneInfo(
            host_denom="uatom", stride_rate=STRIDE_RATE, st_supply=st_supply
        ),
        denoms=ATOM_ZONE,
        vault_native=vault_native,
        fee_reserve=0,
        reports=reports,
        planned=planned or [],
        creation_fee=creation_fee,
        vault_fee_balance=vault_fee_balance,
    )


def load_script(name: str) -> Any:
    """Import a scripts/wind-down module by path; the dashboard never imports from its parent directory."""
    spec = importlib.util.spec_from_file_location(name, WIND_DOWN_DIR / f"{name}.py")
    module = importlib.util.module_from_spec(spec)
    sys.modules[name] = module
    spec.loader.exec_module(module)
    return module


class ConfigTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        load_script("bech32_ref")  # coverage_check imports it by bare name
        cls.coverage_check = load_script("coverage_check")

    def test_required_routes_equal_the_coverage_check_policy(self) -> None:
        self.assertEqual(config.REQUIRED_ROUTES, self.coverage_check.REQUIRED_ROUTES)

    def test_osmosis_channel_hosts_match_the_go_transfer_channel_map(self) -> None:
        go_source = (WIND_DOWN_DIR.parent.parent / "x" / "stakeibc" / "types" / "wind_down.go").read_text()
        go_map = re.search(r"HostToOsmosisTransferChannel = map\[string\]string\{(.*?)\}", go_source, re.DOTALL)
        self.assertIsNotNone(go_map)

        go_hosts = set(re.findall(r'"([^"]+)":', go_map.group(1))) - {"osmosis-1"}

        self.assertEqual(go_hosts, set(config.OSMOSIS_CHANNEL_TO_HOST))

    def test_escrow_address_matches_the_script(self) -> None:
        self.assertEqual(
            chain.escrow_address(channel_id="channel-0"),
            self.coverage_check.escrow_address(channel_id="channel-0"),
        )
        self.assertEqual(
            chain.escrow_address(channel_id="channel-0"),
            "stride1a53udazy8ayufvy0s434pfwjcedzqv3448zelz",
        )
        self.assertNotEqual(
            chain.escrow_address(channel_id="channel-1"),
            chain.escrow_address(channel_id="channel-0"),
        )

    def test_pools_refresh_interval_is_registered(self) -> None:
        self.assertEqual(config.REFRESH_INTERVAL_SECONDS["pools"], 300)


class DiscoveryTest(unittest.TestCase):
    def setUp(self) -> None:
        pools._NOT_OURS.clear()

    def test_candidates_are_the_code_996_pools_not_already_ruled_out(self) -> None:
        listed = {
            "osmo1ours": pools.ListedPool(pool_id="1", code_id="996"),
            "osmo1other": pools.ListedPool(pool_id="2", code_id="148"),
            "osmo1known": pools.ListedPool(pool_id="3", code_id="996"),
        }

        self.assertEqual(
            pools.transmuter_candidates(listed=listed, skip={"osmo1known"}),
            ["osmo1ours"],
        )

    def test_keeps_vault_admin_pools_and_extras_only(self) -> None:
        admins = {
            "osmo1ours": VAULT,
            "osmo1theirs": STRANGER,
            "osmo1unknown": None,
            "osmo1extra": STRANGER,
        }

        ours = pools.select_our_pools(admins=admins, extra=("osmo1extra",))

        self.assertEqual(ours, ["osmo1ours", "osmo1extra"])

    def test_snapshot_memoises_only_the_pools_that_are_not_ours(self) -> None:
        listed = {
            "osmo1ours": pools.ListedPool(pool_id="1", code_id="996"),
            "osmo1theirs": pools.ListedPool(pool_id="2", code_id="996"),
            "osmo1legacy": pools.ListedPool(pool_id="3", code_id="148"),
        }
        admin_calls: list[str] = []

        def fake_admin(osmosis: chain.Chain, contract: str) -> str:
            admin_calls.append(contract)
            return VAULT if contract == "osmo1ours" else STRANGER

        def fake_raw_pool(
            osmosis: chain.Chain,
            contract: str,
            listed: pools.ListedPool | None,
            admin: str | None,
        ) -> pools.RawPool:
            return raw_pool(
                contract=contract,
                pool_id=listed.pool_id if listed else None,
                admin=admin,
            )

        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")
        with (
            mock.patch.object(pools, "_balances", return_value={ATOM_ON_OSMOSIS: 7}),
            mock.patch.object(pools, "_listed_pools", return_value=listed),
            mock.patch.object(pools, "_admin", fake_admin),
            mock.patch.object(pools, "_raw_pool", fake_raw_pool),
            mock.patch.object(pools, "_creation_fee", return_value=FEE),
            mock.patch.object(pools, "_holder_targets", return_value=HOLDER_TARGETS),
        ):
            first = pools._osmosis_snapshot(osmosis=osmosis)
            second = pools._osmosis_snapshot(osmosis=osmosis)

        self.assertEqual([pool.contract for pool in first.pools], ["osmo1ours"])
        self.assertEqual(first.pools[0].pool_id, "1")
        self.assertEqual(first.vault_balances, {ATOM_ON_OSMOSIS: 7})
        self.assertEqual((first.creation_fee, first.holder_targets), (FEE, HOLDER_TARGETS))
        # The stranger's pool is asked once; ours is asked again so an admin transfer shows on the next refresh.
        self.assertEqual(sorted(admin_calls), ["osmo1ours", "osmo1ours", "osmo1theirs"])
        self.assertEqual([pool.contract for pool in second.pools], ["osmo1ours"])

    def test_an_extra_is_read_whatever_its_admin(self) -> None:
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")
        with (
            mock.patch.object(pools, "_balances", return_value={}),
            mock.patch.object(pools, "_listed_pools", return_value={}),
            mock.patch.object(pools, "_admin", side_effect=TimeoutError("slow")),
            mock.patch.object(
                pools,
                "_raw_pool",
                side_effect=lambda osmosis, contract, listed, admin: raw_pool(
                    contract=contract, pool_id=None, admin=admin
                ),
            ),
            mock.patch.object(config, "EXTRA_POOL_CONTRACTS", ("osmo1extra",)),
            mock.patch.object(pools, "_creation_fee", return_value=None),
            mock.patch.object(pools, "_holder_targets", return_value={}),
        ):
            snapshot = pools._osmosis_snapshot(osmosis=osmosis)

        self.assertEqual(
            [(pool.contract, pool.admin, pool.pool_id) for pool in snapshot.pools],
            [("osmo1extra", None, None)],
        )
        self.assertEqual(pools._NOT_OURS, set())


class ClassificationTest(unittest.TestCase):
    def test_zone_denoms_follow_the_osmosis_channel_to_the_host(self) -> None:
        self.assertEqual(ATOM_ZONE.osmosis_denom, ATOM_ON_OSMOSIS)
        self.assertEqual(ATOM_ZONE.native_trace, "transfer/channel-0/uatom")
        self.assertEqual(ATOM_ZONE.st_denom, "stuatom")
        self.assertEqual(ATOM_ZONE.canonical_st_trace, "transfer/channel-326/stuatom")

    def test_osmosis_itself_uses_the_bare_denom(self) -> None:
        osmo = pools.zone_denoms(chain_id="osmosis-1", host_denom="uosmo")

        self.assertEqual((osmo.osmosis_denom, osmo.native_trace), ("uosmo", "uosmo"))

    def test_the_pool_belongs_to_the_zone_whose_sttoken_it_holds(self) -> None:
        canonical = pools.st_asset_of(raw=raw_pool(), zone=ATOM_ZONE)
        route = pools.st_asset_of(
            raw=raw_pool(st_denom=HUB_STATOM, st_trace=HUB_TRACE), zone=ATOM_ZONE
        )
        other_zone = pools.st_asset_of(
            raw=raw_pool(st_denom=CANONICAL_STINJ, st_trace=INJ_TRACE), zone=ATOM_ZONE
        )

        self.assertEqual(
            canonical, pools.StAsset(denom=CANONICAL_STATOM, trace=CANONICAL_TRACE)
        )
        self.assertEqual(route, pools.StAsset(denom=HUB_STATOM, trace=HUB_TRACE))
        self.assertIsNone(other_zone)

    def test_a_pool_with_two_sttokens_is_classified_by_the_canonical_one(self) -> None:
        raw = raw_pool()
        raw = dataclasses.replace(
            raw,
            assets={**raw.assets, HUB_STATOM: ST_FACTOR},
            traces={**raw.traces, HUB_STATOM: HUB_TRACE},
        )

        self.assertEqual(
            pools.st_asset_of(raw=raw, zone=ATOM_ZONE).denom, CANONICAL_STATOM
        )

    def test_route_hops_need_exactly_two_transfer_hops(self) -> None:
        self.assertEqual(
            pools.route_hops(trace=HUB_TRACE),
            pools.RouteHops(osmosis_channel="channel-0", holder_channel="channel-391"),
        )
        self.assertIsNone(pools.route_hops(trace=CANONICAL_TRACE))
        self.assertIsNone(
            pools.route_hops(
                trace=pools.DenomTrace(
                    path="transfer/channel-0/transfer/channel-1/transfer/channel-2",
                    base_denom="stuatom",
                )
            )
        )
        self.assertIsNone(
            pools.route_hops(
                trace=pools.DenomTrace(
                    path="transfer/channel-0/wasm/channel-1", base_denom="stuatom"
                )
            )
        )


class RouteMatchTest(unittest.TestCase):
    TARGETS = {
        "channel-0": pools.ChannelTarget(
            counterparty_channel="channel-391", chain_id="cosmoshub-4"
        ),
        "channel-6": pools.ChannelTarget(
            counterparty_channel="channel-391", chain_id="injective-1"
        ),
        "channel-40": pools.ChannelTarget(
            counterparty_channel="channel-37", chain_id="secret-4"
        ),
    }

    def match(self, holder_chain_id: str, holder_channel: str) -> pools.Route | None:
        return pools.match_route(
            holder_chain_id=holder_chain_id,
            holder_channel=holder_channel,
            policy_channels=frozenset({"channel-0", "channel-6", "channel-40"}),
            targets=self.TARGETS,
        )

    def test_picks_the_policy_channel_whose_counterparty_and_client_both_match(
        self,
    ) -> None:
        self.assertEqual(
            self.match(holder_chain_id="cosmoshub-4", holder_channel="channel-391"),
            HUB_ROUTE,
        )
        self.assertEqual(
            self.match(holder_chain_id="injective-1", holder_channel="channel-391"),
            pools.Route(
                chain_id="injective-1",
                stride_channel="channel-6",
                counterparty_channel="channel-391",
            ),
        )

    def test_no_match_when_the_counterparty_or_the_chain_differs(self) -> None:
        self.assertIsNone(
            self.match(holder_chain_id="secret-4", holder_channel="channel-391")
        )
        self.assertIsNone(
            self.match(holder_chain_id="cosmoshub-4", holder_channel="channel-37")
        )

    def test_a_policy_channel_outside_the_targets_is_skipped(self) -> None:
        route = pools.match_route(
            holder_chain_id="cosmoshub-4",
            holder_channel="channel-391",
            policy_channels=frozenset({"channel-0", "channel-999"}),
            targets=self.TARGETS,
        )

        self.assertEqual(route.stride_channel, "channel-0")


class PoolReportTest(unittest.TestCase):
    def test_a_green_canonical_pool_before_funding(self) -> None:
        pool = report(raw_pool())

        self.assertEqual(pool.kind, pools.PoolKind.CANONICAL)
        self.assertIsNone(pool.route)
        self.assertEqual(pool.rate, "1.500000000000000000")
        self.assertEqual(pool.rate_gap_pct, "0.0000")
        self.assertEqual(pool.outside_shares, 0)
        self.assertIsNone(pool.allocation)  # filled in by allocate()
        self.assertIsNone(pool.funded_exactly)
        self.assertFalse(pool.native_marked)
        self.assertTrue(pool.ready)
        self.assertEqual(
            [entry.ok for entry in pool.checks], [True] * len(pools.CheckName)
        )
        self.assertEqual([entry.name for entry in pool.checks], list(pools.CheckName))

    def test_a_route_pool_is_allocated_its_escrow_at_its_own_rate_rounded_up(
        self,
    ) -> None:
        # The pool's factors encode 1.5 while Stride's frozen rate is 1.6: the gap stays in the canonical pool.
        pool = report(
            raw_pool(st_denom=HUB_STATOM, st_trace=HUB_TRACE),
            route_lookup=HUB_LOOKUP,
            escrow=1_000_001,
            stride_rate="1.6",
        )

        self.assertEqual(pool.kind, pools.PoolKind.ROUTE)
        self.assertEqual(pool.route, HUB_ROUTE)
        self.assertEqual(pool.escrow, 1_000_001)
        self.assertEqual(pool.allocation, 1_500_002)
        self.assertFalse(pool.funded_exactly)
        self.assertEqual(pool.rate_gap_pct, "6.2500")
        self.assertEqual(check(pool, pools.CheckName.ST_TRACE).ok, True)
        self.assertIn(
            "cosmoshub-4 via Stride channel-0",
            check(pool, pools.CheckName.ST_TRACE).detail,
        )

    def test_funded_exactly_compares_the_vault_shares_with_the_allocation(self) -> None:
        raw = raw_pool(
            st_denom=HUB_STATOM,
            st_trace=HUB_TRACE,
            native=1_500_002,
            alloyed_supply=1_500_002,
            corrupted=[ATOM_ON_OSMOSIS],
        )

        funded = report(
            raw, vault_shares=1_500_002, route_lookup=HUB_LOOKUP, escrow=1_000_001
        )
        short = report(
            raw, vault_shares=1_500_001, route_lookup=HUB_LOOKUP, escrow=1_000_001
        )

        self.assertTrue(funded.funded_exactly)
        self.assertTrue(funded.native_marked)
        self.assertTrue(funded.ready)
        self.assertFalse(short.funded_exactly)
        self.assertEqual(short.outside_shares, 1)
        self.assertEqual(check(short, pools.CheckName.NO_OUTSIDE_SHARES).ok, False)

    def test_allocation_is_null_when_the_escrow_is_unknown(self) -> None:
        pool = report(
            raw_pool(st_denom=HUB_STATOM, st_trace=HUB_TRACE),
            route_lookup=HUB_LOOKUP,
            escrow=None,
        )

        self.assertIsNone(pool.allocation)
        self.assertIsNone(pool.funded_exactly)
        self.assertIsNone(pool.escrow)

    def test_a_route_matching_no_policy_channel_is_unrecognised(self) -> None:
        lookup = pools.RouteLookup(chain_id="secret-4", route=None)

        pool = report(
            raw_pool(st_denom=SECRET_STATOM, st_trace=SECRET_TRACE),
            route_lookup=lookup,
            escrow=None,
        )

        self.assertEqual(pool.kind, pools.PoolKind.UNRECOGNISED)
        self.assertIsNone(pool.route)
        self.assertIsNone(pool.allocation)
        self.assertEqual(check(pool, pools.CheckName.ST_TRACE).ok, False)
        self.assertIn(
            "secret-4 via channel-37", check(pool, pools.CheckName.ST_TRACE).detail
        )
        self.assertFalse(pool.ready)

    def test_a_failed_route_lookup_keeps_the_pool_a_route_with_its_trace_check_unknown(
        self,
    ) -> None:
        pool = report(
            raw_pool(st_denom=HUB_STATOM, st_trace=HUB_TRACE), route_lookup=None
        )

        self.assertEqual(pool.kind, pools.PoolKind.ROUTE)
        self.assertIsNone(pool.route)
        self.assertIsNone(check(pool, pools.CheckName.ST_TRACE).ok)
        self.assertFalse(pool.ready)

    def test_a_three_hop_sttoken_is_unrecognised(self) -> None:
        trace = pools.DenomTrace(
            path="transfer/channel-0/transfer/channel-1/transfer/channel-2",
            base_denom="stuatom",
        )
        denom = chain.ibc_denom(path=trace.full)

        pool = report(raw_pool(st_denom=denom, st_trace=trace), route_lookup=None)

        self.assertEqual(pool.kind, pools.PoolKind.UNRECOGNISED)
        self.assertEqual(check(pool, pools.CheckName.ST_TRACE).ok, False)


class ChecksTest(unittest.TestCase):
    def assert_fails(
        self, raw: pools.RawPool, name: pools.CheckName, vault_shares: int = 0
    ) -> pools.Check:
        pool = report(raw, vault_shares=vault_shares)
        failing = [entry.name for entry in pool.checks if entry.ok is False]
        self.assertEqual(failing, [name])
        self.assertFalse(pool.ready)
        return check(pool, name)

    def test_wrong_code_id(self) -> None:
        self.assert_fails(raw_pool(code_id="148"), pools.CheckName.CODE_ID)

    def test_wrong_cw2_version(self) -> None:
        self.assert_fails(raw_pool(cw2_version="3.1.0"), pools.CheckName.CW2_VERSION)

    def test_a_second_sttoken_fails_the_assets_check(self) -> None:
        raw = raw_pool()
        raw = dataclasses.replace(
            raw,
            assets={**raw.assets, HUB_STATOM: ST_FACTOR},
            traces={**raw.traces, HUB_STATOM: HUB_TRACE},
        )

        failing = self.assert_fails(raw, pools.CheckName.ASSETS)

        self.assertIn(HUB_STATOM, failing.detail)

    def test_inverted_factors_encode_no_exact_rate(self) -> None:
        pool = report(scaled_pool(st_factor=RATE_FACTOR, native_factor=ST_FACTOR))

        self.assertEqual(check(pool, pools.CheckName.FACTORS).ok, False)
        self.assertEqual(pool.rate, "0.666666666666666667")
        self.assertEqual(
            check(pool, pools.CheckName.RATE).ok, True
        )  # below Stride's rate: not a payout risk, but wrong

    def test_alloyed_factor_must_equal_the_native_factor(self) -> None:
        failing = self.assert_fails(
            scaled_pool(st_factor=ST_FACTOR, native_factor=RATE_FACTOR, alloyed_factor=ST_FACTOR),
            pools.CheckName.FACTORS,
        )

        self.assertEqual(failing.detail, f"stToken={ST_FACTOR} native={RATE_FACTOR} alloyed={ST_FACTOR}: rate 1.500000000000000000")

    def test_factors_pass_at_the_1e18_and_the_1e6_scale(self) -> None:
        scaled_1e18 = report(scaled_pool(st_factor=10**18, native_factor=RATE_FACTOR))
        scaled_1e6 = report(scaled_pool(st_factor=1_000_000, native_factor=1_500_000))

        for pool in (scaled_1e18, scaled_1e6):
            self.assertEqual(check(pool, pools.CheckName.FACTORS).ok, True)
            self.assertEqual(pool.rate, "1.500000000000000000")
            self.assertEqual(pool.rate_gap_pct, "0.0000")
            self.assertEqual(check(pool, pools.CheckName.RATE).ok, True)
        self.assertIn("stToken=1000000 native=1500000 alloyed=1500000", check(scaled_1e6, pools.CheckName.FACTORS).detail)

    def test_rate_gap_and_rate_check_work_at_the_1e6_scale(self) -> None:
        raw = scaled_pool(st_factor=1_000_000, native_factor=1_500_000)

        below = report(raw, stride_rate="1.6")
        above = report(raw, stride_rate="1.499999")

        self.assertEqual(below.rate_gap_pct, "6.2500")
        self.assertEqual(check(below, pools.CheckName.RATE).ok, True)
        self.assertEqual(check(above, pools.CheckName.RATE).ok, False)
        self.assertEqual(above.rate_gap_pct, "-0.0001")

    def test_mismatched_alloyed_factor_fails_at_the_1e6_scale(self) -> None:
        self.assert_fails(
            scaled_pool(st_factor=1_000_000, native_factor=1_500_000, alloyed_factor=1_000_000),
            pools.CheckName.FACTORS,
        )

    def test_a_rate_above_strides_pays_out_more_than_its_backing(self) -> None:
        pool = report(raw_pool(), stride_rate="1.499999999999999999")

        self.assertEqual(check(pool, pools.CheckName.RATE).ok, False)
        self.assertTrue(pool.rate_gap_pct.startswith("-0.0000"))

    def test_foreign_admin(self) -> None:
        failing = self.assert_fails(raw_pool(admin=STRANGER), pools.CheckName.ADMIN)

        self.assertEqual(failing.detail, STRANGER)

    def test_foreign_moderator(self) -> None:
        self.assert_fails(raw_pool(moderator=STRANGER), pools.CheckName.MODERATOR)

    def test_an_admin_transfer_in_flight(self) -> None:
        failing = self.assert_fails(
            raw_pool(admin_candidate=pools.AdminCandidate(address=STRANGER)),
            pools.CheckName.NO_ADMIN_TRANSFER,
        )

        self.assertEqual(failing.detail, f"candidate {STRANGER}")

    def test_inactive(self) -> None:
        self.assert_fails(raw_pool(is_active=False), pools.CheckName.ACTIVE)

    def test_a_limiter(self) -> None:
        failing = self.assert_fails(
            raw_pool(limiters=[f"{ATOM_ON_OSMOSIS}/1h"]), pools.CheckName.NO_LIMITERS
        )

        self.assertEqual(failing.detail, f"{ATOM_ON_OSMOSIS}/1h")

    def test_native_missing_fails_native_trace_assets_and_factors(self) -> None:
        raw = raw_pool()
        raw = dataclasses.replace(
            raw,
            assets={CANONICAL_STATOM: ST_FACTOR, raw.alloyed_denom: RATE_FACTOR},
        )

        pool = report(raw)

        self.assertEqual(check(pool, pools.CheckName.NATIVE_TRACE).ok, False)
        self.assertEqual(
            check(pool, pools.CheckName.NATIVE_TRACE).detail,
            "expects transfer/channel-0/uatom",
        )
        self.assertEqual(check(pool, pools.CheckName.ASSETS).ok, False)
        self.assertEqual(check(pool, pools.CheckName.FACTORS).ok, False)
        self.assertIsNone(pool.rate)
        self.assertIsNone(check(pool, pools.CheckName.RATE).ok)
        self.assertIsNone(check(pool, pools.CheckName.HEADROOM).ok)

    def test_uint128_headroom_is_sized_by_needed_on_an_empty_pool(self) -> None:
        # Every pool is empty before funding, so the balance says nothing: the zone's `needed` is what must fit once the
        # transmuter scales it by lcm(factors) / native factor.
        haqq_1e18 = report(scaled_pool(st_factor=10**18, native_factor=HAQQ_RATE_FACTOR), needed=HAQQ_NEEDED)
        hub_1e18 = report(scaled_pool(st_factor=10**18, native_factor=HUB_RATE_FACTOR), needed=HUB_NEEDED)
        haqq_1e6 = report(scaled_pool(st_factor=1_000_000, native_factor=1_060_497), needed=HAQQ_NEEDED)

        self.assertEqual(check(haqq_1e18, pools.CheckName.HEADROOM).ok, False)
        self.assertEqual(check(hub_1e18, pools.CheckName.HEADROOM).ok, True)
        self.assertEqual(check(haqq_1e6, pools.CheckName.HEADROOM).ok, True)
        self.assertEqual(
            check(haqq_1e18, pools.CheckName.HEADROOM).detail,
            f"needed 1.071e+26 x (lcm 1.060e+36 / native factor {HAQQ_RATE_FACTOR}) = 1.071e+44 vs 2^128 3.403e+38: "
            "use factors scaled to 1e6 (stToken 1000000, native floor(rate x 1e6), never rounded up) at creation",
        )
        self.assertEqual(
            check(haqq_1e6, pools.CheckName.HEADROOM).detail,
            "needed 1.071e+26 x (lcm 1.060e+12 / native factor 1060497) = 1.071e+32 vs 2^128 3.403e+38",
        )
        self.assertFalse(haqq_1e18.ready)

    def test_uint128_headroom_uses_the_lcm_when_the_factors_share_a_divisor(self) -> None:
        # The Hub-like native factor shares 250 with 1e18, so the lcm is 1e18 x native / 250 and the multiplier 4e15.
        pool = report(scaled_pool(st_factor=10**18, native_factor=HUB_RATE_FACTOR), needed=HUB_NEEDED)

        self.assertEqual(check(pool, pools.CheckName.HEADROOM).detail, "needed 2.175e+12 x (lcm 8.054e+33 / native factor 2013525450106978250) = 8.700e+27 vs 2^128 3.403e+38")

    def test_corrupted_set_must_match_the_funding_phase(self) -> None:
        marked_before_funding = report(
            raw_pool(corrupted=[ATOM_ON_OSMOSIS]), vault_shares=0
        )
        unmarked_after_funding = report(
            raw_pool(alloyed_supply=5, corrupted=[]), vault_shares=5
        )
        marked_after_funding = report(
            raw_pool(alloyed_supply=5, corrupted=[ATOM_ON_OSMOSIS]), vault_shares=5
        )
        wrong_denom = report(
            raw_pool(alloyed_supply=5, corrupted=[CANONICAL_STATOM]), vault_shares=5
        )

        self.assertEqual(
            check(marked_before_funding, pools.CheckName.CORRUPTED).ok, False
        )
        self.assertTrue(marked_before_funding.native_marked)
        self.assertEqual(
            check(unmarked_after_funding, pools.CheckName.CORRUPTED).ok, False
        )
        self.assertEqual(
            check(marked_after_funding, pools.CheckName.CORRUPTED).ok, True
        )
        self.assertEqual(check(wrong_denom, pools.CheckName.CORRUPTED).ok, False)
        self.assertFalse(wrong_denom.native_marked)

    def test_shares_outside_the_vault(self) -> None:
        pool = report(raw_pool(alloyed_supply=12), vault_shares=10)

        self.assertEqual(pool.outside_shares, 2)
        self.assertEqual(check(pool, pools.CheckName.NO_OUTSIDE_SHARES).ok, False)
        self.assertEqual(
            check(pool, pools.CheckName.NO_OUTSIDE_SHARES).detail,
            "2 shares outside the vault",
        )

    def test_a_failed_optional_lookup_makes_its_check_unknown_and_the_pool_not_ready(
        self,
    ) -> None:
        raw = raw_pool(
            code_id=None,
            cw2_version=None,
            admin=None,
            moderator=None,
            admin_candidate=None,
            is_active=None,
            limiters=None,
            corrupted=None,
            alloyed_supply=None,
        )

        pool = report(raw)

        unknown = {entry.name for entry in pool.checks if entry.ok is None}
        self.assertEqual(
            unknown,
            {
                pools.CheckName.CODE_ID,
                pools.CheckName.CW2_VERSION,
                pools.CheckName.ADMIN,
                pools.CheckName.MODERATOR,
                pools.CheckName.NO_ADMIN_TRANSFER,
                pools.CheckName.ACTIVE,
                pools.CheckName.NO_LIMITERS,
                pools.CheckName.CORRUPTED,
                pools.CheckName.NO_OUTSIDE_SHARES,
            },
        )
        self.assertFalse(any(entry.ok is False for entry in pool.checks))
        self.assertFalse(pool.ready)
        self.assertIsNone(pool.outside_shares)
        self.assertEqual(pool.corrupted, [])


class AllocationTest(unittest.TestCase):
    def test_the_canonical_pool_gets_everything_the_routes_are_not_owed(self) -> None:
        canonical = report(raw_pool(native=100), vault_shares=100)
        hub = report(
            raw_pool(
                contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE, native=30
            ),
            route_lookup=HUB_LOOKUP,
            escrow=20,
        )

        allocated = pools.allocate(reports=[hub, canonical], vault_native=1_000)

        self.assertEqual(allocated[1].allocation, 1_000 + 100 + 30 - 30)
        self.assertFalse(allocated[1].funded_exactly)
        self.assertEqual(allocated[0], hub)  # route reports are untouched

    def test_the_osmosis_zone_keeps_the_fee_reserve_out_of_the_allocation(self) -> None:
        canonical = report(raw_pool(native=100))
        reserve = pools.fee_reserve(chain_id="osmosis-1")

        osmosis = pools.allocate(reports=[canonical], vault_native=40_000_000, fee_reserve=reserve)
        short = pools.allocate(reports=[canonical], vault_native=5_000_000, fee_reserve=reserve)
        hub = pools.allocate(reports=[canonical], vault_native=40_000_000, fee_reserve=pools.fee_reserve(chain_id="cosmoshub-4"))

        self.assertEqual(reserve, config.OSMO_FEE_RESERVE)
        self.assertEqual(osmosis[0].allocation, 40_000_000 - config.OSMO_FEE_RESERVE + 100)
        self.assertEqual(short[0].allocation, 100)  # the reserve never pushes the vault below zero
        self.assertEqual(hub[0].allocation, 40_000_000 + 100)

    def test_two_canonical_pools_are_duplicates_and_neither_is_allocated(self) -> None:
        first = report(raw_pool(contract="osmo1canonb", native=100), vault_shares=100)
        second = report(raw_pool(contract="osmo1canona", native=50))

        allocated = pools.allocate(reports=[first, second], vault_native=1_000)

        self.assertEqual([pool.allocation for pool in allocated], [None, None])
        self.assertEqual([pool.funded_exactly for pool in allocated], [None, None])
        self.assertEqual([check(pool, pools.CheckName.UNIQUE).ok for pool in allocated], [False, False])
        self.assertEqual(check(allocated[0], pools.CheckName.UNIQUE).detail, "2 pools claim canonical: osmo1canona, osmo1canonb")
        self.assertEqual([pool.ready for pool in allocated], [False, False])

    def test_two_route_pools_on_one_channel_are_duplicates_and_block_the_canonical_share(self) -> None:
        canonical = report(raw_pool(native=100))
        hub = report(
            raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE, native=30),
            route_lookup=HUB_LOOKUP,
            escrow=20,
        )
        twin = report(
            raw_pool(contract="osmo1hubtwin", st_denom=HUB_STATOM, st_trace=HUB_TRACE, native=30),
            route_lookup=HUB_LOOKUP,
            escrow=20,
        )

        allocated = pools.allocate(reports=[canonical, hub, twin], vault_native=1_000)

        self.assertEqual([pool.allocation for pool in allocated], [None, None, None])
        self.assertEqual([check(pool, pools.CheckName.UNIQUE).ok for pool in allocated], [True, False, False])
        self.assertEqual(check(allocated[1], pools.CheckName.UNIQUE).detail, "2 pools claim route on channel-0: osmo1hub, osmo1hubtwin")
        self.assertEqual(check(allocated[0], pools.CheckName.UNIQUE).detail, "canonical")
        self.assertTrue(allocated[0].ready)
        self.assertEqual([pool.ready for pool in allocated[1:]], [False, False])

    def test_unrecognised_and_unresolved_pools_claim_no_slot(self) -> None:
        unrecognised = report(
            raw_pool(contract="osmo1secret", st_denom=SECRET_STATOM, st_trace=SECRET_TRACE),
            route_lookup=pools.RouteLookup(chain_id="secret-4", route=None),
        )
        unresolved = report(raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE), route_lookup=None)

        self.assertEqual(pools.duplicate_slots(reports=[unrecognised, unrecognised, unresolved, unresolved]), {})
        self.assertEqual(check(unrecognised, pools.CheckName.UNIQUE).detail, "claims no slot")

    def test_canonical_is_funded_exactly_when_the_vault_holds_that_many_shares(
        self,
    ) -> None:
        canonical = report(
            raw_pool(native=1_000, alloyed_supply=1_000, corrupted=[ATOM_ON_OSMOSIS]),
            vault_shares=1_000,
        )

        allocated = pools.allocate(reports=[canonical], vault_native=0)

        self.assertEqual(allocated[0].allocation, 1_000)
        self.assertTrue(allocated[0].funded_exactly)

    def test_canonical_allocation_is_null_while_a_route_or_unrecognised_pool_is_unknown(
        self,
    ) -> None:
        canonical = report(raw_pool())
        escrow_unknown = report(
            raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE),
            route_lookup=HUB_LOOKUP,
            escrow=None,
        )
        unrecognised = report(
            raw_pool(
                contract="osmo1secret", st_denom=SECRET_STATOM, st_trace=SECRET_TRACE
            ),
            route_lookup=pools.RouteLookup(chain_id="secret-4", route=None),
        )

        self.assertIsNone(
            pools.canonical_allocation(
                reports=[canonical, escrow_unknown], vault_native=5
            )
        )
        self.assertIsNone(
            pools.canonical_allocation(
                reports=[canonical, unrecognised], vault_native=5
            )
        )
        self.assertEqual(
            pools.canonical_allocation(reports=[canonical], vault_native=5), 5
        )

    def test_ceil_times_rate_rounds_up_and_is_exact_at_eighteen_decimals(self) -> None:
        supply = 107_130_613_496_528_606_863_394_583
        rate = "1.060496560022296837"

        self.assertEqual(
            pools.ceil_times_rate(amount=1_000_000, rate="1.0000001"), 1_000_001
        )
        self.assertEqual(pools.ceil_times_rate(amount=2, rate="1.5"), 3)
        self.assertEqual(
            pools.ceil_times_rate(amount=supply, rate=rate),
            -(-supply * 1_060_496_560_022_296_837 // 10**18),
        )


class ZoneTest(unittest.TestCase):
    def route_report(
        self,
        contract: str,
        stride_channel: str,
        counterparty: str,
        native: int = 0,
        escrow: int | None = 0,
    ) -> pools.PoolReport:
        trace = pools.DenomTrace(
            path=f"transfer/channel-0/transfer/{counterparty}", base_denom="stuatom"
        )
        route = pools.Route(
            chain_id="cosmoshub-4",
            stride_channel=stride_channel,
            counterparty_channel=counterparty,
        )
        return report(
            raw_pool(
                contract=contract,
                st_denom=chain.ibc_denom(path=trace.full),
                st_trace=trace,
                native=native,
            ),
            route_lookup=pools.RouteLookup(chain_id="cosmoshub-4", route=route),
            escrow=escrow,
        )

    def test_needed_coverage_and_pool_order(self) -> None:
        canonical = report(raw_pool(native=400))
        late = self.route_report(
            contract="osmo1late",
            stride_channel="channel-148",
            counterparty="channel-9",
            native=50,
        )
        early = self.route_report(
            contract="osmo1early",
            stride_channel="channel-6",
            counterparty="channel-8",
            native=50,
        )

        zone = zone_pools(
            reports=[late, canonical, early], vault_native=1_000, st_supply=1_000_001
        )

        self.assertEqual(zone.needed, 1_500_002)  # 1_000_001 x 1.5 rounded up
        self.assertEqual(zone.pools_native, 500)
        self.assertEqual(zone.native_on_osmosis, 1_500)
        self.assertEqual(zone.coverage, "0.001000")
        self.assertEqual(
            [pool.contract for pool in zone.pools],
            ["osmo1canon", "osmo1early", "osmo1late"],
        )
        self.assertEqual(
            (zone.symbol, zone.decimals, zone.st_denom, zone.osmosis_denom),
            ("ATOM", 6, "stuatom", ATOM_ON_OSMOSIS),
        )

    def test_coverage_is_null_when_nothing_is_needed(self) -> None:
        self.assertIsNone(zone_pools(reports=[], st_supply=0).coverage)

    def test_missing_routes_are_the_policy_channels_without_a_pool(self) -> None:
        served = self.route_report(
            contract="osmo1hub", stride_channel="channel-0", counterparty="channel-391"
        )

        zone = zone_pools(reports=[report(raw_pool()), served])

        self.assertEqual(
            zone.missing_routes,
            [
                "channel-6",
                "channel-40",
                "channel-52",
                "channel-123",
                "channel-148",
            ],
        )
        self.assertFalse(zone.pools_ready)

    def test_pools_ready_needs_a_canonical_pool_every_route_and_every_check(
        self,
    ) -> None:
        full_policy = frozenset({"channel-0"})
        canonical = report(raw_pool())
        hub = self.route_report(
            contract="osmo1hub", stride_channel="channel-0", counterparty="channel-391"
        )
        broken = report(raw_pool(admin=STRANGER))
        unknown = report(raw_pool(is_active=None))

        with mock.patch.dict(config.REQUIRED_ROUTES, {"stuatom": full_policy}):
            ready = zone_pools(reports=[canonical, hub])
            no_canonical = zone_pools(reports=[hub])
            failing = zone_pools(reports=[broken, hub])
            unknown_check = zone_pools(reports=[unknown, hub])
            failing_and_unknown = zone_pools(reports=[unknown, hub, broken])

        self.assertTrue(ready.pools_ready)
        self.assertFalse(no_canonical.pools_ready)
        self.assertFalse(failing.pools_ready)
        self.assertIsNone(unknown_check.pools_ready)
        self.assertFalse(failing_and_unknown.pools_ready)

    def test_duplicate_pools_keep_the_zone_not_ready(self) -> None:
        twins = pools.allocate(
            reports=[report(raw_pool(contract="osmo1a")), report(raw_pool(contract="osmo1b"))], vault_native=0
        )
        hub = self.route_report(contract="osmo1hub", stride_channel="channel-0", counterparty="channel-391")
        hub_twin = self.route_report(contract="osmo1hubtwin", stride_channel="channel-0", counterparty="channel-391")
        route_twins = pools.allocate(reports=[report(raw_pool()), hub, hub_twin], vault_native=0)

        with mock.patch.dict(config.REQUIRED_ROUTES, {"stuatom": frozenset({"channel-0"})}):
            two_canonical = zone_pools(reports=twins + [hub])
            two_routes = zone_pools(reports=route_twins)

        self.assertFalse(two_canonical.pools_ready)
        self.assertFalse(two_canonical.pools_funded)
        self.assertFalse(two_routes.pools_ready)
        self.assertEqual(two_routes.missing_routes, [])
        self.assertEqual([pool.allocation for pool in two_routes.pools], [None, None, None])

    def test_pools_funded_needs_ready_funded_exactly_and_marked_everywhere(
        self,
    ) -> None:
        funded_canonical = pools.allocate(
            reports=[
                report(
                    raw_pool(
                        native=1_000, alloyed_supply=1_000, corrupted=[ATOM_ON_OSMOSIS]
                    ),
                    vault_shares=1_000,
                )
            ],
            vault_native=0,
        )[0]
        unfunded_canonical = pools.allocate(
            reports=[report(raw_pool())], vault_native=1_000
        )[0]
        # Both pools funded and marked, but the route's escrow could not be read: the canonical share is unknown.
        marked_canonical = report(
            raw_pool(native=1_000, alloyed_supply=1_000, corrupted=[ATOM_ON_OSMOSIS]), vault_shares=1_000
        )
        marked_route = report(
            raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE, native=30, alloyed_supply=30, corrupted=[ATOM_ON_OSMOSIS]),
            vault_shares=30,
            route_lookup=HUB_LOOKUP,
            escrow=None,
        )
        unknown_allocation = pools.allocate(reports=[marked_canonical, marked_route], vault_native=0)

        with mock.patch.dict(config.REQUIRED_ROUTES, {"stuatom": frozenset()}):
            funded = zone_pools(reports=[funded_canonical])
            unfunded = zone_pools(reports=[unfunded_canonical])
            unknown = zone_pools(reports=unknown_allocation)

        self.assertTrue(funded.pools_ready)
        self.assertTrue(funded.pools_funded)
        self.assertTrue(unfunded.pools_ready)
        self.assertFalse(unfunded.pools_funded)
        self.assertFalse(unfunded.pools[0].native_marked)
        self.assertTrue(unknown.pools_ready)
        self.assertEqual([pool.allocation for pool in unknown.pools], [None, None])
        self.assertIsNone(unknown.pools_funded)


class CollectTest(unittest.TestCase):
    def test_zone_pools_reads_the_supply_route_and_escrow_and_serialises_ints_as_strings(
        self,
    ) -> None:
        canonical = raw_pool(
            native=1_000, alloyed_supply=1_000, corrupted=[ATOM_ON_OSMOSIS]
        )
        hub = raw_pool(
            contract="osmo1hub",
            st_denom=HUB_STATOM,
            st_trace=HUB_TRACE,
            native=30,
            alloyed_supply=30,
            corrupted=[ATOM_ON_OSMOSIS],
        )
        inj = raw_pool(
            contract="osmo1inj", st_denom=CANONICAL_STINJ, st_trace=INJ_TRACE
        )
        snapshot = pools.OsmosisSnapshot(
            vault_balances={
                ATOM_ON_OSMOSIS: 5,
                canonical.alloyed_denom: 1_000,
                hub.alloyed_denom: 30,
                ALLUSDC: 25_000_000,
            },
            pools=[canonical, hub, inj],
            creation_fee=FEE,
            holder_targets=HOLDER_TARGETS,
        )
        stride = chain.stride_chain()
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")
        escrow_calls: list[tuple[str, str]] = []

        def fake_escrow(stride: chain.Chain, channel_id: str, denom: str) -> int:
            escrow_calls.append((channel_id, denom))
            return 20

        with (
            mock.patch.object(pools, "_supply", return_value=10**6),
            mock.patch.object(pools, "_balance_of", return_value=10**6),
            mock.patch.object(
                pools, "_resolve_route", return_value=HUB_LOOKUP
            ) as resolve,
            mock.patch.object(pools, "_escrow_balance", fake_escrow),
            mock.patch.object(pools, "_channel_target", return_value=HUB_TARGET),
            mock.patch.dict(
                config.REQUIRED_ROUTES, {"stuatom": frozenset({"channel-0"})}
            ),
        ):
            zone = pools._zone_pools(
                zone=config.ZONES_BY_CHAIN_ID["cosmoshub-4"],
                host_zone={"host_denom": "uatom", "redemption_rate": STRIDE_RATE},
                stride=stride,
                osmosis=osmosis,
                snapshot=snapshot,
            )
            payload = pools._payload(
                zones=[
                    zone,
                    chain.ZoneError(chain_id="juno-1", error="TimeoutError: slow"),
                ]
            )

        self.assertEqual(
            resolve.call_args.kwargs["hops"],
            pools.RouteHops(osmosis_channel="channel-0", holder_channel="channel-391"),
        )
        self.assertEqual(escrow_calls, [("channel-0", "stuatom")])
        self.assertEqual(
            [pool.contract for pool in zone.pools], ["osmo1canon", "osmo1hub"]
        )  # the stINJ pool is not this zone's
        self.assertEqual(zone.pools[1].allocation, 30)
        self.assertEqual(zone.pools[0].allocation, 5 + 1_000 + 30 - 30)
        self.assertEqual(zone.vault_native, 5)
        self.assertEqual(zone.fee_reserve, 0)
        self.assertTrue(zone.pools_ready)
        self.assertFalse(
            zone.pools_funded
        )  # the canonical pool holds 1_000 shares against a 1_005 allocation

        # The planned pools: the canonical one and the Hub route, both created (the hub pool holds the route denom).
        self.assertEqual([plan.alloyed_subdenom for plan in zone.planned], ["stATOM", "stATOM.cosmoshub"])
        self.assertEqual([plan.live_contract for plan in zone.planned], ["osmo1canon", "osmo1hub"])
        self.assertEqual([plan.seeded for plan in zone.planned], [True, True])  # the test wallet answers 10**6 for every denom
        self.assertEqual((zone.routes_seeded, zone.canonical_seeded, zone.pools_created), (True, True, True))
        self.assertEqual((zone.creation_fee, zone.vault_fee_balance, zone.creation_fee_short), (FEE, 25_000_000, False))

        encoded = json.loads(json.dumps(payload))
        self.assertEqual(encoded["zones"][0]["needed"], "1500000")
        self.assertEqual(encoded["zones"][0]["creation_fee"], {"denom": ALLUSDC, "amount": "20000000"})
        self.assertEqual(encoded["zones"][0]["vault_fee_balance"], "25000000")
        self.assertEqual(encoded["zones"][0]["planned"][1]["kind"], "route")
        self.assertEqual(encoded["zones"][0]["planned"][1]["st_factor"], "1000000000000000000")
        self.assertEqual(encoded["zones"][0]["planned"][1]["test_wallet_balance"], "1000000")
        self.assertEqual(encoded["zones"][0]["planned"][1]["instantiate_msg"]["alloyed_asset_subdenom"], "stATOM.cosmoshub")
        self.assertEqual(
            set(encoded["zones"][0]["planned"][0]),
            {
                "kind", "stride_channel", "holder_chain_id", "holder_name", "holder_binary", "holder_node",
                "counterparty_channel", "osmosis_channel", "holder_to_osmosis_channel", "denom_on_holder",
                "denom_on_osmosis", "seeded", "test_wallet_balance", "st_factor", "native_factor", "alloyed_subdenom",
                "instantiate_msg", "seed_command", "live_contract", "error", "blocked",
            },
        )
        self.assertEqual(encoded["zones"][0]["fee_reserve"], "0")
        self.assertEqual(encoded["zones"][0]["pools"][0]["kind"], "canonical")
        self.assertEqual(
            encoded["zones"][0]["pools"][1]["route"], dataclasses.asdict(HUB_ROUTE)
        )
        self.assertEqual(encoded["zones"][0]["pools"][1]["escrow"], "20")
        self.assertIs(encoded["zones"][0]["pools"][0]["funded_exactly"], False)
        self.assertEqual(
            encoded["zones"][0]["pools"][0]["checks"][0],
            {"name": "code id 996", "ok": True, "detail": "code_id=996"},
        )
        self.assertEqual(
            encoded["zones"][1], {"chain_id": "juno-1", "error": "TimeoutError: slow"}
        )

    def test_the_osmosis_zone_sizes_its_pool_by_needed_and_keeps_the_fee_reserve(self) -> None:
        pool = osmo_pool()
        snapshot = pools.OsmosisSnapshot(vault_balances={"uosmo": 40_000_000}, pools=[pool])
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")

        with (
            mock.patch.object(pools, "_supply", return_value=10**6),
            mock.patch.object(pools, "_balance_of", return_value=10**6),
            mock.patch.object(pools, "_channel_target", side_effect=TimeoutError("slow")),
        ):
            zone = pools._zone_pools(
                zone=config.ZONES_BY_CHAIN_ID["osmosis-1"],
                host_zone={"host_denom": "uosmo", "redemption_rate": STRIDE_RATE},
                stride=chain.stride_chain(),
                osmosis=osmosis,
                snapshot=snapshot,
            )

        self.assertEqual((zone.vault_native, zone.fee_reserve), (40_000_000, config.OSMO_FEE_RESERVE))
        self.assertEqual(zone.pools[0].allocation, 40_000_000 - config.OSMO_FEE_RESERVE)
        self.assertEqual(zone.needed, 1_500_000)
        self.assertIn("needed 1.500e+06", check(zone.pools[0], pools.CheckName.HEADROOM).detail)
        # No fee read and every Stride channel lookup failed: the gates are unknown, the routes carry their error.
        self.assertEqual((zone.creation_fee, zone.vault_fee_balance, zone.creation_fee_short), (None, None, None))
        self.assertEqual([plan.error is not None for plan in zone.planned], [False, True, True])
        self.assertEqual((zone.routes_seeded, zone.pools_created), (None, None))

    def test_a_zone_failure_becomes_an_error_entry_and_an_unreadable_osmosis_fails_every_zone(
        self,
    ) -> None:
        host_zones = {
            "host_zone": [
                {"chain_id": zone.chain_id, "host_denom": "u", "redemption_rate": "1"}
                for zone in config.ZONES
            ]
        }

        with (
            mock.patch.object(chain, "rest_get", return_value=host_zones),
            mock.patch.object(
                pools, "_osmosis_snapshot", side_effect=TimeoutError("osmosis down")
            ),
        ):
            payload = pools.collect()

        self.assertEqual(len(payload["zones"]), len(config.ZONES))
        self.assertTrue(
            all(
                zone["error"] == "TimeoutError: osmosis down"
                for zone in payload["zones"]
            )
        )

        with (
            mock.patch.object(chain, "rest_get", return_value=host_zones),
            mock.patch.object(
                pools,
                "_osmosis_snapshot",
                return_value=pools.OsmosisSnapshot(vault_balances={}, pools=[]),
            ),
            mock.patch.object(
                pools, "_supply", side_effect=TimeoutError("stride slow")
            ),
        ):
            payload = pools.collect()

        self.assertTrue(
            all(
                zone["error"] == "TimeoutError: stride slow"
                for zone in payload["zones"]
            )
        )


# ---- planned pools

HUB_RATE = "2.013525450106978250"
HAQQ_RATE = "1.060496560022296837"
HUB_TARGET = pools.ChannelTarget(counterparty_channel="channel-391", chain_id="cosmoshub-4")
# Osmosis's holder channels (config.HOLDER_ROUTES) -> what they point at, as resolved on 2026-10-08.
HOLDER_TARGETS = {
    "channel-0": pools.ChannelTarget(counterparty_channel="channel-141", chain_id="cosmoshub-4"),
    "channel-208": pools.ChannelTarget(counterparty_channel="channel-3", chain_id="axelar-dojo-1"),
    "channel-88": pools.ChannelTarget(counterparty_channel="channel-1", chain_id="secret-4"),
    "channel-874": None,  # the lookup failed
}
# Every policy channel in config.REQUIRED_ROUTES -> (holder chain, its channel to Stride), resolved live on 2026-10-08.
POLICY_TARGETS = {
    "channel-0": ("cosmoshub-4", "channel-391"),
    "channel-6": ("injective-1", "channel-89"),
    "channel-40": ("secret-4", "channel-37"),
    "channel-52": ("phoenix-1", "channel-46"),
    "channel-123": ("neutron-1", "channel-8"),
    "channel-148": ("agoric-3", "channel-59"),
    "channel-162": ("celestia", "channel-4"),
    "channel-197": ("dymension_1100-1", "channel-0"),
    "channel-240": ("haqq_11235-1", "channel-7"),
}
HUB_PLANNED_ROUTE = pools.plan_route(stride_channel="channel-0", target=HUB_TARGET, holder_targets=HOLDER_TARGETS)
ATOM_ZONE_CONFIG = config.ZONES_BY_CHAIN_ID["cosmoshub-4"]
HAQQ_ZONE_CONFIG = config.ZONES_BY_CHAIN_ID["haqq_11235-1"]
ALLUSDC = "factory/osmo147h5x9pcj7lm0cttlaefx6sqq5vdfnmwfcqxkmjd7exqm9gc7grqhr75m0/alloyed/allUSDC"
FEE = pools.CreationFee(denom=ALLUSDC, amount=20_000_000)


def route_plan(
    route: pools.PlannedRoute = HUB_PLANNED_ROUTE,
    subdenom: str = "stATOM.cosmoshub",
    supply: int | None = 5,
    reports: list[pools.PoolReport] | None = None,
    stride_rate: str = HUB_RATE,
) -> pools.PlannedPool:
    return pools.build_route_plan(
        zone=ATOM_ZONE_CONFIG,
        denoms=ATOM_ZONE,
        stride_rate=stride_rate,
        route=route,
        subdenom=subdenom,
        supply=supply,
        reports=reports or [],
    )


INJECTIVE_TARGET = pools.ChannelTarget(counterparty_channel="channel-89", chain_id="injective-1")
INJECTIVE_PLANNED_ROUTE = pools.plan_route(
    stride_channel="channel-6",
    target=INJECTIVE_TARGET,
    holder_targets={**HOLDER_TARGETS, "channel-122": pools.ChannelTarget(counterparty_channel="channel-8", chain_id="injective-1")},
)


class BlockedRouteTest(unittest.TestCase):
    def test_a_route_from_a_blocked_holder_chain_carries_the_reason_and_the_hub_route_does_not(self) -> None:
        with mock.patch.dict(config.BLOCKED_HOLDER_CHAINS, {"injective-1": "rate limiter"}, clear=True):
            blocked = route_plan(route=INJECTIVE_PLANNED_ROUTE, subdenom="stATOM.injective")
            open_route = route_plan()

        self.assertEqual((blocked.blocked, blocked.holder_chain_id), ("rate limiter", "injective-1"))
        self.assertIsNone(open_route.blocked)
        self.assertIsNone(canonical_plan().blocked)

    def test_blocked_pools_stay_out_of_the_zone_gates_and_the_missing_routes(self) -> None:
        with mock.patch.dict(config.BLOCKED_HOLDER_CHAINS, {"injective-1": "rate limiter"}, clear=True):
            blocked = route_plan(route=INJECTIVE_PLANNED_ROUTE, subdenom="stATOM.injective", supply=0)
        created_canonical = canonical_plan(reports=[report(raw_pool(native=1_000, alloyed_supply=1_000, corrupted=[ATOM_ON_OSMOSIS]), vault_shares=1_000)])
        with mock.patch.dict(config.REQUIRED_ROUTES, {"stuatom": frozenset({"channel-6"})}):
            zone = zone_pools(reports=[], planned=[created_canonical, blocked])

        # Unseeded and uncreated, yet the zone reads created and seeded: the blocked route is deferred, not missing.
        self.assertEqual(zone.missing_routes, [])
        self.assertEqual((zone.routes_seeded, zone.canonical_seeded), (True, True))  # no open route left to seed
        self.assertEqual(zone.creation_fee_short, False)


def canonical_plan(supply: int | None = 5, reports: list[pools.PoolReport] | None = None) -> pools.PlannedPool:
    return pools.build_canonical_plan(
        zone=ATOM_ZONE_CONFIG, denoms=ATOM_ZONE, stride_rate=HUB_RATE, supply=supply, reports=reports or []
    )


class PlannedRouteTest(unittest.TestCase):
    def test_resolves_the_holder_chain_and_both_osmosis_hops(self) -> None:
        self.assertEqual(
            HUB_PLANNED_ROUTE,
            pools.PlannedRoute(
                stride_channel="channel-0",
                holder_chain_id="cosmoshub-4",
                holder=config.HOLDER_CHAINS["cosmoshub-4"],
                counterparty_channel="channel-391",
                osmosis_channel="channel-0",
                holder_to_osmosis_channel="channel-141",
                error=None,
            ),
        )

    def test_a_chain_outside_holder_chains_keeps_what_resolved_and_carries_the_error(self) -> None:
        route = pools.plan_route(
            stride_channel="channel-8",
            target=pools.ChannelTarget(counterparty_channel="channel-32", chain_id="kaiyo-1"),
            holder_targets=HOLDER_TARGETS,
        )

        self.assertEqual(route.error, "kaiyo-1 is not in config.HOLDER_CHAINS")
        self.assertEqual((route.holder, route.holder_chain_id, route.counterparty_channel), (None, "kaiyo-1", "channel-32"))
        self.assertEqual((route.osmosis_channel, route.holder_to_osmosis_channel), (None, None))

    def test_no_osmosis_channel_tracking_the_chain_is_an_error(self) -> None:
        unlisted = pools.plan_route(
            stride_channel="channel-6",
            target=pools.ChannelTarget(counterparty_channel="channel-89", chain_id="injective-1"),
            holder_targets=HOLDER_TARGETS,
        )
        failed_lookup = pools.plan_route(
            stride_channel="channel-123",
            target=pools.ChannelTarget(counterparty_channel="channel-8", chain_id="neutron-1"),
            holder_targets=HOLDER_TARGETS,
        )

        self.assertEqual(unlisted.error, "no channel in config.HOLDER_ROUTES tracks injective-1 on Osmosis")
        self.assertEqual(failed_lookup.error, "no channel in config.HOLDER_ROUTES tracks neutron-1 on Osmosis")
        self.assertEqual(unlisted.holder, config.HOLDER_CHAINS["injective-1"])
        self.assertIsNone(unlisted.osmosis_channel)

    def test_a_failed_or_half_open_stride_channel_is_an_error_with_nothing_resolved(self) -> None:
        failed = pools.plan_route(stride_channel="channel-0", target=None, holder_targets=HOLDER_TARGETS)
        half_open = pools.plan_route(
            stride_channel="channel-0",
            target=pools.ChannelTarget(counterparty_channel=None, chain_id="cosmoshub-4"),
            holder_targets=HOLDER_TARGETS,
        )

        self.assertEqual(failed.error, "Stride channel-0 could not be resolved (channel, connection or client lookup failed)")
        self.assertEqual(half_open.error, "Stride channel-0 has no counterparty channel yet")
        for route in (failed, half_open):
            self.assertEqual((route.holder_chain_id, route.holder, route.counterparty_channel, route.osmosis_channel), (None,) * 4)

    def test_holder_chains_cover_every_chain_a_policy_channel_resolves_to(self) -> None:
        policy_channels = set().union(*config.REQUIRED_ROUTES.values())

        self.assertEqual(policy_channels, set(POLICY_TARGETS))  # the fixture is the whole policy
        for channel, (chain_id, counterparty) in POLICY_TARGETS.items():
            with self.subTest(channel=channel):
                self.assertIn(chain_id, config.HOLDER_CHAINS)
                route = pools.plan_route(
                    stride_channel=channel,
                    target=pools.ChannelTarget(counterparty_channel=counterparty, chain_id=chain_id),
                    holder_targets={"osmo-ch": pools.ChannelTarget(counterparty_channel="x-ch", chain_id=chain_id)},
                )
                self.assertIsNone(route.error)

    def test_holder_nodes_are_private_where_the_zone_has_one_else_public_or_a_placeholder(self) -> None:
        self.assertEqual(config.HOLDER_CHAINS["cosmoshub-4"].node, "https://cosmos-strd-rpc.polkachu.com:443")
        self.assertEqual(config.HOLDER_CHAINS["axelar-dojo-1"].node, "https://axelar-rpc.polkachu.com:443")
        self.assertEqual(config.HOLDER_CHAINS["secret-4"].node, "<RPC>")
        self.assertEqual(config.HOLDER_CHAINS["cosmoshub-4"].binary, "gaiad")
        self.assertEqual(config.STRIDE_HOLDER, config.HolderChain(name="stride", binary="strided", node="https://stride-strd-rpc.polkachu.com:443"))


class PlannedDenomsAndFactorsTest(unittest.TestCase):
    def test_route_denoms_are_the_one_and_two_hop_hashes(self) -> None:
        denoms = pools.route_denoms(route=HUB_PLANNED_ROUTE, st_denom="stuatom")

        self.assertEqual(denoms.on_holder, chain.ibc_denom(path="transfer/channel-391/stuatom"))
        self.assertEqual(denoms.on_osmosis, chain.ibc_denom(path="transfer/channel-0/transfer/channel-391/stuatom"))
        # The hashes transmuter.md records for the Hub route.
        self.assertEqual(denoms.on_holder, "ibc/B05539B66B72E2739B986B86391E5D08F12B8D5D2C2A7F8F8CF9ADF674DFA231")
        self.assertEqual(denoms.on_osmosis, "ibc/7451074F46885686D3B47B12A6BF74F6D36847ED1891AC612FCFAEB7FB551E14")

    def test_route_denoms_stop_where_resolution_stopped(self) -> None:
        unknown_chain = pools.plan_route(
            stride_channel="channel-8",
            target=pools.ChannelTarget(counterparty_channel="channel-32", chain_id="kaiyo-1"),
            holder_targets=HOLDER_TARGETS,
        )
        unresolved = pools.plan_route(stride_channel="channel-0", target=None, holder_targets=HOLDER_TARGETS)

        self.assertEqual(
            pools.route_denoms(route=unknown_chain, st_denom="stuatom"),
            pools.RouteDenoms(on_holder=chain.ibc_denom(path="transfer/channel-32/stuatom"), on_osmosis=None),
        )
        self.assertEqual(pools.route_denoms(route=unresolved, st_denom="stuatom"), pools.RouteDenoms(on_holder=None, on_osmosis=None))

    def test_six_decimal_factors_are_1e18_scaled_and_exact(self) -> None:
        self.assertEqual(
            pools.creation_factors(decimals=6, stride_rate=HUB_RATE),
            pools.Factors(st_factor=10**18, native_factor=2_013_525_450_106_978_250),
        )

    def test_eighteen_decimal_factors_are_1e6_scaled_and_rounded_down(self) -> None:
        # 1.060496560022296837 x 1e6 = 1060496.56...: round would give 1060497, floor gives 1060496.
        self.assertEqual(
            pools.creation_factors(decimals=18, stride_rate=HAQQ_RATE),
            pools.Factors(st_factor=10**6, native_factor=1_060_496),
        )
        self.assertEqual(pools.creation_factors(decimals=18, stride_rate="1.0000009").native_factor, 1_000_000)

    def test_subdenoms(self) -> None:
        axelar = pools.plan_route(
            stride_channel="channel-11",
            target=pools.ChannelTarget(counterparty_channel="channel-33", chain_id="axelar-dojo-1"),
            holder_targets=HOLDER_TARGETS,
        )
        unresolved = pools.plan_route(stride_channel="channel-47", target=None, holder_targets=HOLDER_TARGETS)

        self.assertEqual(pools.alloyed_subdenom(st_symbol="stATOM", route=None, shared_chain=False), "stATOM")
        self.assertEqual(pools.alloyed_subdenom(st_symbol="stATOM", route=HUB_PLANNED_ROUTE, shared_chain=False), "stATOM.cosmoshub")
        self.assertEqual(pools.alloyed_subdenom(st_symbol="stATOM", route=axelar, shared_chain=True), "stATOM.axelar.channel11")
        self.assertEqual(pools.alloyed_subdenom(st_symbol="stATOM", route=unresolved, shared_chain=False), "stATOM.channel47")
        self.assertEqual(pools.st_symbol(zone=HAQQ_ZONE_CONFIG), "stISLM")


class PlannedPoolTest(unittest.TestCase):
    def test_a_route_plan_carries_the_denoms_factors_message_and_seed_command(self) -> None:
        plan = route_plan()
        on_holder = chain.ibc_denom(path="transfer/channel-391/stuatom")

        self.assertEqual(plan.kind, pools.PoolKind.ROUTE)
        self.assertEqual(
            (plan.stride_channel, plan.holder_chain_id, plan.holder_name, plan.holder_binary, plan.holder_node),
            ("channel-0", "cosmoshub-4", "cosmoshub", "gaiad", "https://cosmos-strd-rpc.polkachu.com:443"),
        )
        self.assertEqual(
            (plan.counterparty_channel, plan.osmosis_channel, plan.holder_to_osmosis_channel),
            ("channel-391", "channel-0", "channel-141"),
        )
        self.assertEqual((plan.denom_on_holder, plan.denom_on_osmosis), (on_holder, HUB_STATOM))
        self.assertEqual((plan.seeded, plan.test_wallet_balance), (True, 5))
        self.assertEqual((plan.st_factor, plan.native_factor), (10**18, 2_013_525_450_106_978_250))
        self.assertEqual(plan.alloyed_subdenom, "stATOM.cosmoshub")
        self.assertEqual(
            plan.instantiate_msg,
            {
                "pool_asset_configs": [
                    {"denom": HUB_STATOM, "normalization_factor": "1000000000000000000"},
                    {"denom": ATOM_ON_OSMOSIS, "normalization_factor": "2013525450106978250"},
                ],
                "alloyed_asset_subdenom": "stATOM.cosmoshub",
                "alloyed_asset_normalization_factor": "2013525450106978250",
                "admin": VAULT,
                "moderator": VAULT,
            },
        )
        self.assertEqual(
            plan.seed_command,
            f"gaiad tx ibc-transfer transfer transfer channel-141 {config.POOL_SEED_ADDRESS} 10000{on_holder} --from <KEY_ON_COSMOSHUB> "
            "--chain-id cosmoshub-4 --node https://cosmos-strd-rpc.polkachu.com:443 --gas auto --gas-adjustment 1.5 --fees <FEES>",
        )
        self.assertIsNone(plan.live_contract)
        self.assertIsNone(plan.error)

    def test_an_eighteen_decimal_route_seeds_a_hundredth_of_a_token_with_1e6_factors(self) -> None:
        haqq_denoms = pools.zone_denoms(chain_id="haqq_11235-1", host_denom="aISLM")
        route = pools.plan_route(
            stride_channel="channel-240",
            target=pools.ChannelTarget(counterparty_channel="channel-7", chain_id="haqq_11235-1"),
            holder_targets={"channel-1575": pools.ChannelTarget(counterparty_channel="channel-2", chain_id="haqq_11235-1")},
        )

        plan = pools.build_route_plan(
            zone=HAQQ_ZONE_CONFIG, denoms=haqq_denoms, stride_rate=HAQQ_RATE, route=route, subdenom="stISLM.haqq", supply=0, reports=[]
        )

        self.assertEqual((plan.st_factor, plan.native_factor), (10**6, 1_060_496))
        self.assertEqual(plan.instantiate_msg["alloyed_asset_normalization_factor"], "1060496")
        self.assertIn(f" {config.POOL_SEED_ADDRESS} {10**16}{plan.denom_on_holder} --from <KEY_ON_HAQQ> --chain-id haqq_11235-1 ", plan.seed_command)
        self.assertTrue(plan.seed_command.startswith("haqqd tx ibc-transfer transfer transfer channel-2 "))
        self.assertEqual((plan.seeded, plan.test_wallet_balance), (False, 0))

    def test_seeded_is_tri_state_and_the_route_seed_command_stays(self) -> None:
        self.assertEqual((route_plan(supply=0).seeded, route_plan(supply=None).seeded), (False, None))
        self.assertIsNotNone(route_plan(supply=0).seed_command)
        self.assertIsNotNone(route_plan(supply=None).seed_command)

    def test_the_live_contract_is_the_existing_pool_holding_the_denom(self) -> None:
        hub = report(raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE), route_lookup=HUB_LOOKUP, escrow=1)
        canonical = report(raw_pool(contract="osmo1canon"))

        self.assertEqual(route_plan(reports=[canonical, hub]).live_contract, "osmo1hub")
        self.assertEqual(canonical_plan(reports=[canonical, hub]).live_contract, "osmo1canon")
        self.assertIsNone(route_plan(reports=[canonical]).live_contract)

    def test_an_unresolved_route_has_no_message_or_commands_but_keeps_what_resolved(self) -> None:
        route = pools.plan_route(
            stride_channel="channel-8",
            target=pools.ChannelTarget(counterparty_channel="channel-32", chain_id="kaiyo-1"),
            holder_targets=HOLDER_TARGETS,
        )

        plan = route_plan(route=route, subdenom="stATOM.channel8", supply=None)

        self.assertEqual(plan.error, "kaiyo-1 is not in config.HOLDER_CHAINS")
        self.assertEqual((plan.holder_chain_id, plan.holder_name, plan.counterparty_channel), ("kaiyo-1", None, "channel-32"))
        self.assertEqual(plan.denom_on_holder, chain.ibc_denom(path="transfer/channel-32/stuatom"))
        self.assertEqual((plan.denom_on_osmosis, plan.instantiate_msg, plan.seed_command, plan.seeded), (None, None, None, None))
        self.assertEqual(plan.alloyed_subdenom, "stATOM.channel8")

    def test_the_canonical_plan_seeds_from_stride_only_while_unseeded(self) -> None:
        seeded = canonical_plan(supply=1)
        unseeded = canonical_plan(supply=0)
        unknown = canonical_plan(supply=None)

        self.assertEqual(seeded.kind, pools.PoolKind.CANONICAL)
        self.assertEqual(
            (seeded.stride_channel, seeded.holder_chain_id, seeded.holder_name, seeded.holder_binary, seeded.holder_node),
            (None, "stride-1", "stride", "strided", "https://stride-strd-rpc.polkachu.com:443"),
        )
        self.assertEqual((seeded.counterparty_channel, seeded.osmosis_channel, seeded.holder_to_osmosis_channel), (None, "channel-326", "channel-5"))
        self.assertEqual((seeded.denom_on_holder, seeded.denom_on_osmosis), (None, CANONICAL_STATOM))
        self.assertEqual(seeded.alloyed_subdenom, "stATOM")
        self.assertEqual(seeded.instantiate_msg["pool_asset_configs"][0]["denom"], CANONICAL_STATOM)
        self.assertEqual([plan.seed_command for plan in (seeded, unknown)], [None, None])
        self.assertEqual(
            unseeded.seed_command,
            f"strided tx ibc-transfer transfer transfer channel-5 {config.POOL_SEED_ADDRESS} 10000stuatom --from <KEY_ON_STRIDE> "
            "--chain-id stride-1 --node https://stride-strd-rpc.polkachu.com:443 --gas auto --gas-adjustment 1.5 --fees <FEES>",
        )


class ZonePlannedTest(unittest.TestCase):
    def test_planned_pools_are_ordered_canonical_then_by_stride_channel(self) -> None:
        late = route_plan(route=pools.plan_route(stride_channel="channel-148", target=HUB_TARGET, holder_targets=HOLDER_TARGETS))
        early = route_plan()

        zone = zone_pools(reports=[], planned=[late, canonical_plan(), early])

        self.assertEqual([plan.stride_channel for plan in zone.planned], [None, "channel-0", "channel-148"])

    def test_seeded_flags_are_tri_state_over_their_kind(self) -> None:
        seeded = route_plan(supply=3)
        unseeded = route_plan(supply=0)
        unknown = route_plan(supply=None)

        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(supply=0), seeded, seeded]).routes_seeded, True)
        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(), seeded, unknown]).routes_seeded, None)
        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(), unknown, unseeded]).routes_seeded, False)
        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(supply=0), seeded]).canonical_seeded, False)
        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(supply=None)]).canonical_seeded, None)
        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(supply=1)]).canonical_seeded, True)
        self.assertEqual(zone_pools(reports=[], planned=[canonical_plan(supply=0)]).routes_seeded, True)  # no route: vacuous

    def test_pools_created_needs_a_live_contract_everywhere_and_is_unknown_on_an_error(self) -> None:
        hub = report(raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE), route_lookup=HUB_LOOKUP, escrow=1)
        canonical = report(raw_pool(contract="osmo1canon"))
        errored = route_plan(route=pools.plan_route(stride_channel="channel-8", target=None, holder_targets={}), subdenom="x")

        created = zone_pools(reports=[canonical, hub], planned=[canonical_plan(reports=[canonical]), route_plan(reports=[hub])])
        missing = zone_pools(reports=[canonical], planned=[canonical_plan(reports=[canonical]), route_plan(reports=[canonical])])
        unknown = zone_pools(reports=[canonical, hub], planned=[canonical_plan(reports=[canonical]), route_plan(reports=[hub]), errored])
        missing_and_error = zone_pools(reports=[], planned=[canonical_plan(), errored])

        self.assertEqual((created.pools_created, missing.pools_created), (True, False))
        self.assertEqual((unknown.pools_created, missing_and_error.pools_created), (None, False))

    def test_creation_fee_shortfall_counts_the_pools_still_to_create(self) -> None:
        hub = report(raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE), route_lookup=HUB_LOOKUP, escrow=1)
        planned = [canonical_plan(), route_plan(reports=[hub]), route_plan()]  # two still to create

        short = zone_pools(reports=[hub], planned=planned, creation_fee=FEE, vault_fee_balance=39_999_999)
        enough = zone_pools(reports=[hub], planned=planned, creation_fee=FEE, vault_fee_balance=40_000_000)
        no_fee = zone_pools(reports=[hub], planned=planned, creation_fee=None, vault_fee_balance=None)
        no_balance = zone_pools(reports=[hub], planned=planned, creation_fee=FEE, vault_fee_balance=None)
        all_created = zone_pools(reports=[hub], planned=[route_plan(reports=[hub])], creation_fee=FEE, vault_fee_balance=0)

        self.assertEqual((short.creation_fee_short, enough.creation_fee_short), (True, False))
        self.assertEqual((no_fee.creation_fee_short, no_balance.creation_fee_short), (None, None))
        self.assertEqual((short.creation_fee, short.vault_fee_balance), (FEE, 39_999_999))
        self.assertEqual(all_created.creation_fee_short, False)

    def test_tri_state(self) -> None:
        self.assertEqual(pools.tri_state(outcomes=[]), True)
        self.assertEqual(pools.tri_state(outcomes=[True, None]), None)
        self.assertEqual(pools.tri_state(outcomes=[None, False]), False)
        self.assertEqual(pools.tri_state(outcomes=[True, True]), True)


class PlannedNetworkTest(unittest.TestCase):
    def test_planned_pools_resolve_every_policy_channel_and_read_each_denoms_supply(self) -> None:
        stride = chain.stride_chain()
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")
        targets = {
            "channel-0": HUB_TARGET,
            "channel-11": pools.ChannelTarget(counterparty_channel="channel-33", chain_id="axelar-dojo-1"),
            "channel-69": pools.ChannelTarget(counterparty_channel="channel-64", chain_id="axelar-dojo-1"),
        }
        supplies = {
            CANONICAL_STATOM: 10,
            HUB_STATOM: 0,
            chain.ibc_denom(path="transfer/channel-208/transfer/channel-33/stuatom"): 7,
            chain.ibc_denom(path="transfer/channel-208/transfer/channel-64/stuatom"): 0,
        }
        supply_calls: list[str] = []

        def fake_target(chain_handle: chain.Chain, channel_id: str) -> pools.ChannelTarget:
            self.assertEqual(chain_handle.chain_id, "stride-1")
            if channel_id == "channel-6":
                raise TimeoutError("slow")
            return targets[channel_id]

        def fake_supply(chain_handle: chain.Chain, address: str, denom: str) -> int:
            self.assertEqual(address, config.POOL_SEED_ADDRESS)
            supply_calls.append(denom)
            return supplies[denom]

        hub = report(raw_pool(contract="osmo1hub", st_denom=HUB_STATOM, st_trace=HUB_TRACE), route_lookup=HUB_LOOKUP, escrow=1)
        with (
            mock.patch.object(pools, "_channel_target", fake_target),
            mock.patch.object(pools, "_balance_of", fake_supply),
            mock.patch.dict(config.REQUIRED_ROUTES, {"stuatom": frozenset({"channel-69", "channel-0", "channel-11", "channel-6"})}),
        ):
            planned = pools._planned_pools(
                zone=ATOM_ZONE_CONFIG,
                denoms=ATOM_ZONE,
                host=pools.HostZoneInfo(host_denom="uatom", stride_rate=HUB_RATE, st_supply=1),
                stride=stride,
                osmosis=osmosis,
                snapshot=pools.OsmosisSnapshot(vault_balances={}, pools=[], holder_targets=HOLDER_TARGETS),
                reports=[hub],
            )

        self.assertEqual([plan.stride_channel for plan in planned], [None, "channel-0", "channel-6", "channel-11", "channel-69"])
        self.assertEqual(
            [plan.alloyed_subdenom for plan in planned],
            ["stATOM", "stATOM.cosmoshub", "stATOM.channel6", "stATOM.axelar.channel11", "stATOM.axelar.channel69"],
        )
        self.assertEqual([plan.seeded for plan in planned], [True, False, None, True, False])
        self.assertEqual([plan.live_contract for plan in planned], [None, "osmo1hub", None, None, None])
        self.assertEqual(planned[3].holder_to_osmosis_channel, "channel-3")
        self.assertEqual(planned[2].error, "Stride channel-6 could not be resolved (channel, connection or client lookup failed)")
        self.assertEqual(sorted(supply_calls), sorted(supplies))  # one supply read per resolved denom, none for the failed route

    def test_a_failed_supply_lookup_leaves_seeded_unknown(self) -> None:
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")

        with mock.patch.object(pools, "_balance_of", side_effect=TimeoutError("slow")):
            self.assertIsNone(pools._optional_supply(osmosis=osmosis, denom=HUB_STATOM))
        self.assertIsNone(pools._optional_supply(osmosis=osmosis, denom=None))

    def test_creation_fee_tries_the_capitalised_path_then_the_lower_case_one(self) -> None:
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")
        params = {"params": {"pool_creation_fee": [{"denom": ALLUSDC, "amount": "20000000"}]}}
        paths: list[str] = []

        def answers(chain: chain.Chain, path: str, params: dict[str, str] | None = None) -> dict[str, object]:
            paths.append(path)
            if path.endswith("/Params"):
                raise urllib.error.HTTPError(url=path, code=501, msg="Not Implemented", hdrs=None, fp=None)
            return {"params": {"pool_creation_fee": [{"denom": ALLUSDC, "amount": "20000000"}]}}

        with mock.patch.object(chain, "rest_get", return_value=params):
            self.assertEqual(pools._creation_fee(osmosis=osmosis), FEE)
        with mock.patch.object(chain, "rest_get", answers):
            self.assertEqual(pools._creation_fee(osmosis=osmosis), FEE)
        with mock.patch.object(chain, "rest_get", side_effect=TimeoutError("slow")):
            self.assertIsNone(pools._creation_fee(osmosis=osmosis))
        with mock.patch.object(chain, "rest_get", return_value={"params": {"pool_creation_fee": []}}):
            self.assertIsNone(pools._creation_fee(osmosis=osmosis))

        self.assertEqual(paths, ["/osmosis/poolmanager/v1beta1/Params", "/osmosis/poolmanager/v1beta1/params"])

    def test_holder_targets_are_read_per_holder_route_and_a_failure_is_none(self) -> None:
        osmosis = chain.Chain(chain_id="osmosis-1", rest="https://rest", rpc="")

        def fake_target(chain_handle: chain.Chain, channel_id: str) -> pools.ChannelTarget:
            if channel_id == "channel-88":
                raise TimeoutError("slow")
            return pools.ChannelTarget(counterparty_channel="x", chain_id=f"behind {channel_id}")

        with mock.patch.object(pools, "_channel_target", fake_target):
            targets = pools._holder_targets(osmosis=osmosis)

        self.assertEqual(set(targets), {route.osmosis_channel for route in config.HOLDER_ROUTES})
        self.assertIsNone(targets["channel-88"])
        self.assertEqual(targets["channel-0"].chain_id, "behind channel-0")


if __name__ == "__main__":
    unittest.main()
