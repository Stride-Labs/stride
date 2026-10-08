"""Unit test for coverage_check.evaluate on a synthetic export with an injected fetcher."""

import base64
import contextlib
import io
import json
import pathlib
import re
import unittest
import urllib.parse
from decimal import Decimal
from unittest import mock

import coverage_check

STRIDE_ESCROW_CH0 = coverage_check.escrow_address(channel_id="channel-0")
STRIDE_ESCROW_CH5 = coverage_check.escrow_address(channel_id="channel-5")
HOLDER = "stride1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"
VAULT = "osmo1vault"
NATIVE = "ibc/NATIVE"
CANONICAL_ST = "ibc/STUATOM"
ROUTE_ST = "ibc/STUATOM_TWO_HOP"
REQUIRED_ROUTES = {"stuatom": frozenset({"channel-0"})}
ALLOYED_CANONICAL = "factory/osmo1canonical/alloyed/stuatom"
ALLOYED_ROUTE = "factory/osmo1route/alloyed/stuatom"


def pool_state(
    liquidity: dict[str, int],
    alloyed_denom: str = ALLOYED_CANONICAL,
    alloyed_supply: int = 0,
) -> coverage_check.PoolState:
    return coverage_check.PoolState(liquidity=liquidity, alloyed_denom=alloyed_denom, alloyed_supply=alloyed_supply)


def synthetic_export() -> dict:
    return {
        "app_state": {
            "bank": {
                "supply": [{"denom": "stuatom", "amount": "1000000"}],
                "balances": [
                    {"address": STRIDE_ESCROW_CH0, "coins": [{"denom": "stuatom", "amount": "100000"}]},
                    {"address": STRIDE_ESCROW_CH5, "coins": [{"denom": "stuatom", "amount": "300000"}]},
                    {"address": HOLDER, "coins": [{"denom": "stuatom", "amount": "600000"}]},
                ],
            },
            "stakeibc": {
                "host_zone_list": [
                    {"chain_id": "cosmoshub-4", "host_denom": "uatom", "redemption_rate": "1.500000000000000000"}
                ]
            },
            "ibc": {
                "channel_genesis": {
                    "channels": [
                        {"port_id": "transfer", "channel_id": "channel-0"},
                        {"port_id": "icacontroller-GAIA.DELEGATION", "channel_id": "channel-1"},
                        {"port_id": "transfer", "channel_id": "channel-5"},
                    ]
                }
            },
        }
    }


def pools() -> dict:
    return {
        "stuatom": {
            "chain_id": "cosmoshub-4",
            "native_denom_on_osmosis": NATIVE,
            "canonical_pool_id": "1",
            "canonical_st_denom": CANONICAL_ST,
            "route_pools": [{"channel_id": "channel-0", "pool_id": "2", "st_denom": ROUTE_ST}],
        }
    }


class CoverageCheckTest(unittest.TestCase):
    def test_duplicate_route_cannot_fake_total_coverage(self) -> None:
        bad = pools()
        bad["stuatom"]["route_pools"] *= 2

        # Only 1,350,000 exists: counting the route twice used to report 1,500,000
        # and assign just 1,200,000 to canonical, so every coverage check passed.
        liquidity = {"1": {NATIVE: 1_200_000}, "2": {NATIVE: 150_000}}
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                required_routes=REQUIRED_ROUTES,
                export=synthetic_export(), pools=bad, vault_balances={},
                fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
            )

    def test_missing_route_cannot_allocate_its_liability_to_canonical(self) -> None:
        bad = pools()
        bad["stuatom"]["route_pools"] = []

        # All backing in canonical cannot redeem the foreign voucher.
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                required_routes=REQUIRED_ROUTES,
                export=synthetic_export(), pools=bad, vault_balances={},
                fetch_pool_state=lambda pool_id: pool_state(liquidity={NATIVE: 1_500_000}),
            )

    def test_covered_when_every_pool_holds_its_share(self) -> None:
        # supply 1,000,000 × 1.5 = 1,500,000 native needed; route escrow 100,000 × 1.5 = 150,000
        # in the route pool; the canonical pool gets everything else; the vault holds a surplus
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_000}}
        results = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=synthetic_export(),
            pools=pools(),
            vault_balances={NATIVE: 10_000},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
        )
        self.assertEqual(len(results), 1)
        result = results[0]
        self.assertEqual(result.st_denom, "stuatom")
        self.assertEqual(result.required_native, 1_500_000)
        self.assertEqual(result.native_on_osmosis, 1_510_000)
        self.assertEqual(result.route_shortfalls, {})
        self.assertTrue(result.covered)

    def test_duplicate_channel_with_different_pools_is_rejected(self) -> None:
        bad = pools()
        bad["stuatom"]["route_pools"].append(
            {"channel_id": "channel-0", "pool_id": "3", "st_denom": ROUTE_ST}
        )
        self.assert_topology_rejected(pool_config=bad, message="duplicate route channel")

    def test_pool_reused_on_different_channels_is_rejected(self) -> None:
        bad = pools()
        bad["stuatom"]["route_pools"].append(
            {"channel_id": "channel-5", "pool_id": "2", "st_denom": "ibc/OTHER_ROUTE"}
        )
        self.assert_topology_rejected(
            pool_config=bad, required_routes={"stuatom": frozenset({"channel-0", "channel-5"})},
            message="pool ID 2 reused",
        )

    def test_canonical_pool_cannot_also_be_a_route_pool(self) -> None:
        bad = pools()
        bad["stuatom"]["route_pools"][0]["pool_id"] = "1"
        self.assert_topology_rejected(pool_config=bad, message="pool ID 1 reused")

    def test_pool_ids_cannot_be_reused_across_tokens(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "juno-1", "host_denom": "ujuno", "redemption_rate": "1.1"}
        )
        export["app_state"]["bank"]["supply"].append({"denom": "stujuno", "amount": "10"})
        for reused_id in ("1", "2"):
            with self.subTest(pool_id=reused_id):
                bad = pools()
                bad["stujuno"] = {
                    "chain_id": "juno-1", "native_denom_on_osmosis": "ibc/JUNO",
                    "canonical_pool_id": reused_id, "canonical_st_denom": "ibc/STJUNO", "route_pools": [],
                }
                self.assert_topology_rejected(
                    pool_config=bad, export=export,
                    required_routes={"stuatom": frozenset({"channel-0"}), "stujuno": frozenset()},
                    message=f"pool ID {reused_id} reused",
                )

    def test_pool_ids_must_be_canonical_positive_decimal_strings(self) -> None:
        for pool_id in ("01", 1, "0", "-1", "+1", "1.0", " 1", "1 ", "١", True):
            for field in ("canonical_pool_id", "route_pool_id"):
                with self.subTest(pool_id=pool_id, field=field):
                    bad = pools()
                    if field == "canonical_pool_id":
                        bad["stuatom"][field] = pool_id
                    else:
                        bad["stuatom"]["route_pools"][0]["pool_id"] = pool_id
                    self.assert_topology_rejected(pool_config=bad, message="positive decimal string")

    def test_channel_aliases_are_rejected(self) -> None:
        for channel_id in ("channel-00", "channel-01", 0, "channel-+0", "channel-٠"):
            with self.subTest(channel_id=channel_id):
                bad = pools()
                bad["stuatom"]["route_pools"][0]["channel_id"] = channel_id
                self.assert_topology_rejected(pool_config=bad, message="canonical channel ID")

    def test_incomplete_or_unexpected_routes_are_rejected(self) -> None:
        for channel_id in (None, "channel-5"):
            with self.subTest(channel_id=channel_id):
                bad = pools()
                bad["stuatom"]["route_pools"] = [] if channel_id is None else [
                    {"channel_id": channel_id, "pool_id": "2", "st_denom": ROUTE_ST}
                ]
                self.assert_topology_rejected(pool_config=bad, message="approved routes differ")

    def test_canonical_pool_is_required(self) -> None:
        for pool_id in (None, "missing"):
            with self.subTest(pool_id=pool_id):
                bad = pools()
                if pool_id is None:
                    bad["stuatom"]["canonical_pool_id"] = None
                else:
                    del bad["stuatom"]["canonical_pool_id"]
                self.assert_topology_rejected(pool_config=bad, message="canonical pool is required")

    def test_token_cannot_use_another_host_zones_rate(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "other-1", "host_denom": "other", "redemption_rate": "0.01", "deprecated": True}
        )
        bad = pools()
        bad["stuatom"]["chain_id"] = "other-1"
        self.assert_topology_rejected(pool_config=bad, export=export, message="does not match export host zone")

    def test_new_export_token_requires_reviewed_policy(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"][0]["host_denom"] = "new"
        export["app_state"]["bank"]["supply"][0]["denom"] = "stnew"
        bad = {"stnew": pools()["stuatom"]}
        with self.assertRaisesRegex(coverage_check.CoverageInputError, "no reviewed route policy"):
            coverage_check.evaluate(
                export=export, pools=bad, vault_balances={}, fetch_pool_state=unexpected_fetch,
            )

    def test_default_policy_cannot_be_derived_from_pool_config(self) -> None:
        with self.assertRaisesRegex(coverage_check.CoverageInputError, "approved routes differ"):
            coverage_check.evaluate(
                export=synthetic_export(), pools=pools(), vault_balances={}, fetch_pool_state=unexpected_fetch,
            )

    def test_somm_keeps_canonical_coverage_without_foreign_routes(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"][0]["host_denom"] = "usomm"
        export["app_state"]["stakeibc"]["host_zone_list"][0]["chain_id"] = "sommelier-3"
        export["app_state"]["bank"]["supply"][0]["denom"] = "stusomm"
        config = pools()["stuatom"]
        config["chain_id"] = "sommelier-3"
        config["route_pools"] = []
        result = coverage_check.evaluate(
            export=export, pools={"stusomm": config}, vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity={NATIVE: 1_500_000}),
        )[0]
        self.assertTrue(result.covered)
        self.assertEqual(result.canonical_expected, 1_500_000)

    def test_same_channel_for_different_tokens_with_unique_pools_is_valid(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "juno-1", "host_denom": "ujuno", "redemption_rate": "1.1"}
        )
        export["app_state"]["bank"]["supply"].append({"denom": "stujuno", "amount": "10"})
        config = pools()
        config["stujuno"] = {
            "chain_id": "juno-1", "native_denom_on_osmosis": "ibc/JUNO",
            "canonical_pool_id": "3", "canonical_st_denom": "ibc/STJUNO",
            "route_pools": [{"channel_id": "channel-0", "pool_id": "4", "st_denom": "ibc/STJUNO_ROUTE"}],
        }
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_000}, "3": {"ibc/JUNO": 11}, "4": {}}
        results = coverage_check.evaluate(
            export=export, pools=config, vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
            required_routes={"stuatom": frozenset({"channel-0"}), "stujuno": frozenset({"channel-0"})},
        )
        self.assertTrue(all(result.covered for result in results))

    def assert_topology_rejected(
        self,
        pool_config: dict,
        message: str,
        export: dict | None = None,
        required_routes: dict[str, frozenset[str]] | None = None,
    ) -> None:
        with self.assertRaisesRegex(coverage_check.CoverageInputError, message):
            coverage_check.evaluate(
                export=synthetic_export() if export is None else export, pools=pool_config, vault_balances={},
                fetch_pool_state=unexpected_fetch,
                required_routes=REQUIRED_ROUTES if required_routes is None else required_routes,
            )

    def test_shortfall_in_total_and_in_a_route_pool(self) -> None:
        liquidity = {"1": {NATIVE: 1_000_000}, "2": {NATIVE: 149_999}}
        result = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=synthetic_export(),
            pools=pools(),
            vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
        )[0]
        self.assertFalse(result.covered)
        self.assertEqual(result.native_on_osmosis, 1_149_999)
        self.assertEqual(result.route_shortfalls, {"channel-0": 1})
        self.assertEqual(result.canonical_expected, 1_500_000 - 150_000)

    def test_rate_rounds_up_to_base_units(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"][0]["redemption_rate"] = "1.333333333333333333"
        result = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=export,
            pools=pools(),
            vault_balances={NATIVE: 2_000_000},
            fetch_pool_state=lambda pool_id: pool_state(liquidity={NATIVE: 0}),
        )[0]
        self.assertEqual(result.required_native, 1_333_334)
        self.assertEqual(result.route_expected["channel-0"], 133_334)

    def test_partly_redeemed_pools_are_covered(self) -> None:
        # Redemptions moved 40,000 stTokens into the route pool (its native paid out to match), so
        # it needs (100,000 - 40,000) x 1.5 = 90,000 native; the canonical pool took in 200,000
        # stTokens; the total needs (1,000,000 - 40,000 - 200,000) x 1.5 = 1,140,000 native
        liquidity = {
            "1": {NATIVE: 1_050_000, CANONICAL_ST: 200_000},
            "2": {NATIVE: 90_000, ROUTE_ST: 40_000},
        }
        result = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=synthetic_export(),
            pools=pools(),
            vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
        )[0]
        self.assertEqual(result.required_native, 1_140_000)
        self.assertEqual(result.route_expected["channel-0"], 90_000)
        self.assertEqual(result.canonical_expected, 1_050_000)
        self.assertEqual(result.route_shortfalls, {})
        self.assertTrue(result.covered)

    def test_partly_redeemed_route_pool_short_of_the_reduced_requirement_fails(self) -> None:
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 89_999, ROUTE_ST: 40_000}}
        result = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=synthetic_export(), pools=pools(), vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
        )[0]
        self.assertEqual(result.route_shortfalls, {"channel-0": 1})
        self.assertFalse(result.covered)

    def test_route_pool_holding_more_than_escrow_times_rate_fails(self) -> None:
        # escrow 100,000 x 1.5 = 150,000; 150,001 is over-funded even though every total is met
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_001}}
        result = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=synthetic_export(), pools=pools(), vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
        )[0]
        self.assertEqual(result.route_overfunded, {"channel-0": 1})
        self.assertFalse(result.covered)

    def test_route_channel_not_a_transfer_channel_is_an_error(self) -> None:
        # channel-1 is the delegation ICA channel in the synthetic export, channel-99 does not exist
        for channel_id in ("channel-1", "channel-99"):
            bad = pools()
            bad["stuatom"]["route_pools"][0]["channel_id"] = channel_id
            with self.assertRaises(coverage_check.CoverageInputError):
                coverage_check.evaluate(
                    required_routes=REQUIRED_ROUTES,
                    export=synthetic_export(), pools=bad, vault_balances={},
                    fetch_pool_state=lambda pool_id: pool_state(liquidity={}),
                )

    def test_route_pool_without_st_denom_is_an_error(self) -> None:
        bad = pools()
        del bad["stuatom"]["route_pools"][0]["st_denom"]
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                required_routes=REQUIRED_ROUTES,
                export=synthetic_export(), pools=bad, vault_balances={},
                fetch_pool_state=lambda pool_id: pool_state(liquidity={}),
            )

    def test_missing_pool_entry_is_an_error(self) -> None:
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                required_routes=REQUIRED_ROUTES,
                export=synthetic_export(),
                pools={},
                vault_balances={},
                fetch_pool_state=lambda pool_id: pool_state(liquidity={}),
            )

    def test_sttoken_missing_from_supply_is_an_error(self) -> None:
        # A typo in the pools file must not turn into "required 0, covered"
        mistyped = {"stuatom-typo": pools()["stuatom"]}
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                required_routes=REQUIRED_ROUTES,
                export=synthetic_export(),
                pools=mistyped,
                vault_balances={},
                fetch_pool_state=lambda pool_id: pool_state(liquidity={}),
            )

    def test_in_scope_sttoken_missing_from_pools_is_an_error(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "juno-1", "host_denom": "ujuno", "redemption_rate": "1.1"}
        )
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                required_routes=REQUIRED_ROUTES,
                export=export, pools=pools(), vault_balances={},
                fetch_pool_state=lambda pool_id: pool_state(liquidity={}),
            )

    def test_deprecated_zone_is_out_of_scope(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "comdex-1", "host_denom": "ucmdx", "redemption_rate": "1.1", "deprecated": True}
        )
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_000}}
        results = coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES,
            export=export, pools=pools(), vault_balances={},
            fetch_pool_state=lambda pool_id: pool_state(liquidity=liquidity[pool_id]),
        )
        self.assertEqual([result.st_denom for result in results], ["stuatom"])


def unexpected_fetch(pool_id: str) -> coverage_check.PoolState:
    raise AssertionError(f"invalid topology fetched state for pool {pool_id}")


class AlloyedSharesOutsideVaultTest(unittest.TestCase):
    """join_pool is permissionless: shares held by anyone but the vault still redeem for native.

    Synthetic export: supply 1,000,000 stuatom at rate 1.5, route escrow 100,000 (channel-0).
    """

    def test_canonical_joiner_shares_are_owed_native_on_top_of_the_redeemed_requirement(self) -> None:
        # A joiner put 37,000 stTokens into the canonical pool and holds 55,500 shares (x 1.5). The
        # native for the stTokens not in the pool is (1,000,000 - 100,000 - 37,000) x 1.5 = 1,294,500
        result = self.evaluate_joiners(
            canonical_native=1_294_500, route_native=150_000, canonical_joiner=True, route_joiner=False,
        )

        self.assertEqual(result.canonical_outside_shares, 55_500)
        self.assertEqual(result.canonical_expected, 1_294_500 + 55_500)
        self.assertEqual(result.required_native, (1_000_000 - 37_000) * 3 // 2 + 55_500)
        self.assertEqual(result.route_outside_shares, {})
        self.assertLess(result.canonical_actual, result.canonical_expected)
        self.assertFalse(result.covered)

    def test_canonical_covered_once_native_is_raised_by_exactly_the_outside_shares(self) -> None:
        result = self.evaluate_joiners(
            canonical_native=1_294_500 + 55_500, route_native=150_000, canonical_joiner=True, route_joiner=False,
        )

        self.assertEqual(result.native_on_osmosis, result.required_native)
        self.assertEqual(result.canonical_actual, result.canonical_expected)
        self.assertTrue(result.covered)

    def test_route_joiner_shares_are_a_route_shortfall_of_exactly_the_outside_shares(self) -> None:
        # 20,000 stTokens joined the route pool for 30,000 shares: it owes (100,000 - 20,000) x 1.5 = 120,000
        # for the rest of the escrow plus those 30,000
        result = self.evaluate_joiners(
            canonical_native=1_350_000, route_native=120_000, canonical_joiner=False, route_joiner=True,
        )

        self.assertEqual(result.route_outside_shares, {"channel-0": 30_000})
        self.assertEqual(result.route_expected, {"channel-0": 150_000})
        self.assertEqual(result.route_shortfalls, {"channel-0": 30_000})
        self.assertEqual(result.required_native, (1_000_000 - 20_000) * 3 // 2 + 30_000)
        self.assertEqual(result.canonical_outside_shares, 0)
        self.assertFalse(result.covered)

    def test_route_covered_once_topped_up_and_not_over_funded(self) -> None:
        result = self.evaluate_joiners(
            canonical_native=1_350_000, route_native=150_000, canonical_joiner=False, route_joiner=True,
        )

        self.assertEqual(result.route_shortfalls, {})
        self.assertEqual(result.route_overfunded, {})
        self.assertTrue(result.covered)

    def test_route_pool_above_escrow_times_rate_is_still_over_funded(self) -> None:
        # The over-funded test ignores outside shares: 150,001 > escrow 100,000 x 1.5
        result = self.evaluate_joiners(
            canonical_native=1_350_000, route_native=150_001, canonical_joiner=False, route_joiner=True,
        )

        self.assertEqual(result.route_overfunded, {"channel-0": 1})
        self.assertFalse(result.covered)

    def test_route_native_join_reads_as_over_funded_by_the_joined_amount(self) -> None:
        # An outsider joined the route pool with 10 native and holds 10 shares, on top of the vault's full
        # funding: the shares are counted as owed, but the over-funded bound ignores them, so 150,010 is
        # over escrow 100,000 x 1.5 by the 10 joined
        states = {
            "1": pool_state(liquidity={NATIVE: 1_350_000}, alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=1_350_000),
            "2": pool_state(liquidity={NATIVE: 150_010}, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=150_010),
        }
        result = self.evaluate_states(
            states=states, vault_balances={ALLOYED_CANONICAL: 1_350_000, ALLOYED_ROUTE: 150_000},
        )

        self.assertEqual(result.route_outside_shares, {"channel-0": 10})
        self.assertEqual(result.route_expected, {"channel-0": 150_010})
        self.assertEqual(result.route_shortfalls, {})
        self.assertEqual(result.route_overfunded, {"channel-0": 10})
        self.assertFalse(result.covered)

    def test_vault_holding_fewer_shares_cannot_clear_a_route_native_join(self) -> None:
        # The same 10 as a shortfall instead: the vault holds 149,990 of 150,000 shares, so the 10 outside
        # shares are owed against a pool that holds exactly escrow x rate
        states = {
            "1": pool_state(liquidity={NATIVE: 1_350_000}, alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=1_350_000),
            "2": pool_state(liquidity={NATIVE: 150_000}, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=150_000),
        }
        result = self.evaluate_states(
            states=states, vault_balances={ALLOYED_CANONICAL: 1_350_000, ALLOYED_ROUTE: 149_990},
        )

        self.assertEqual(result.route_shortfalls, {"channel-0": 10})
        self.assertEqual(result.route_overfunded, {})
        self.assertFalse(result.covered)

    def test_shares_fully_held_by_the_vault_leave_the_requirement_unchanged(self) -> None:
        # The partly-redeemed case of CoverageCheckTest, with every share held by the vault
        states = {
            "1": pool_state(
                liquidity={NATIVE: 1_050_000, CANONICAL_ST: 200_000},
                alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=1_050_000,
            ),
            "2": pool_state(
                liquidity={NATIVE: 90_000, ROUTE_ST: 40_000}, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=90_000,
            ),
        }
        result = self.evaluate_states(
            states=states, vault_balances={ALLOYED_CANONICAL: 1_050_000, ALLOYED_ROUTE: 90_000},
        )

        self.assertEqual(result.required_native, 1_140_000)
        self.assertEqual(result.route_expected, {"channel-0": 90_000})
        self.assertEqual(result.canonical_expected, 1_050_000)
        self.assertEqual(result.route_outside_shares, {})
        self.assertEqual(result.canonical_outside_shares, 0)
        self.assertTrue(result.covered)

    def test_vault_balance_above_supply_clamps_to_zero_outside_shares(self) -> None:
        states = {
            "1": pool_state(liquidity={NATIVE: 1_350_000}, alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=1_000_000),
            "2": pool_state(liquidity={NATIVE: 150_000}, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=150_000),
        }
        result = self.evaluate_states(
            states=states, vault_balances={ALLOYED_CANONICAL: 2_000_000, ALLOYED_ROUTE: 500_000},
        )

        self.assertEqual(result.canonical_outside_shares, 0)
        self.assertEqual(result.route_outside_shares, {})
        self.assertEqual(result.canonical_expected, 1_350_000)
        self.assertEqual(result.route_expected, {"channel-0": 150_000})
        self.assertEqual(result.required_native, 1_500_000)
        self.assertTrue(result.covered)

    def test_vault_shares_of_another_pool_do_not_offset_outside_shares(self) -> None:
        # The vault holds plenty of the route pool's shares, none of the canonical pool's
        states = {
            "1": pool_state(liquidity={NATIVE: 1_350_000}, alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=1_350_000),
            "2": pool_state(liquidity={NATIVE: 150_000}, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=150_000),
        }
        result = self.evaluate_states(states=states, vault_balances={ALLOYED_ROUTE: 150_000})

        self.assertEqual(result.canonical_outside_shares, 1_350_000)
        self.assertEqual(result.route_outside_shares, {})
        self.assertEqual(result.required_native, 1_500_000 + 1_350_000)
        self.assertFalse(result.covered)

    def test_table_prints_outside_shares_for_a_covered_token(self) -> None:
        result = self.evaluate_joiners(
            canonical_native=1_294_500 + 55_500, route_native=150_000, canonical_joiner=True, route_joiner=True,
        )

        lines = self.printed_lines(results=[result])

        self.assertTrue(result.covered)
        self.assertIn("    route channel-0: 30000 alloyed shares held outside the vault (counted as owed)", lines)
        self.assertIn("    canonical: 55500 alloyed shares held outside the vault (counted as owed)", lines)

    def test_table_prints_outside_shares_for_an_uncovered_token(self) -> None:
        result = self.evaluate_joiners(
            canonical_native=1_294_500, route_native=120_000, canonical_joiner=True, route_joiner=True,
        )

        lines = self.printed_lines(results=[result])

        self.assertFalse(result.covered)
        self.assertIn("    route channel-0: 30000 alloyed shares held outside the vault (counted as owed)", lines)
        self.assertIn("    canonical: 55500 alloyed shares held outside the vault (counted as owed)", lines)

    def test_table_prints_no_outside_shares_line_when_there_are_none(self) -> None:
        states = {
            "1": pool_state(liquidity={NATIVE: 1_350_000}, alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=1_350_000),
            "2": pool_state(liquidity={NATIVE: 150_000}, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=150_000),
        }
        result = self.evaluate_states(
            states=states, vault_balances={ALLOYED_CANONICAL: 1_350_000, ALLOYED_ROUTE: 150_000},
        )

        self.assertNotIn("alloyed shares", "\n".join(self.printed_lines(results=[result])))

    def evaluate_joiners(
        self, canonical_native: int, route_native: int, canonical_joiner: bool, route_joiner: bool,
    ) -> coverage_check.CoverageResult:
        """The vault holds the shares its funding joins minted; a joiner's shares are on top of those:
        55,500 canonical shares (37,000 stTokens x 1.5) and 30,000 route shares (20,000 stTokens x 1.5)."""
        canonical_liquidity = {NATIVE: canonical_native}
        route_liquidity = {NATIVE: route_native}
        canonical_vault_shares = 1_350_000
        route_vault_shares = 150_000
        canonical_supply = canonical_vault_shares
        route_supply = route_vault_shares
        if canonical_joiner:
            canonical_liquidity[CANONICAL_ST] = 37_000
            canonical_supply += 55_500
        if route_joiner:
            route_liquidity[ROUTE_ST] = 20_000
            route_supply += 30_000
        states = {
            "1": pool_state(
                liquidity=canonical_liquidity, alloyed_denom=ALLOYED_CANONICAL, alloyed_supply=canonical_supply,
            ),
            "2": pool_state(liquidity=route_liquidity, alloyed_denom=ALLOYED_ROUTE, alloyed_supply=route_supply),
        }
        return self.evaluate_states(
            states=states,
            vault_balances={ALLOYED_CANONICAL: canonical_vault_shares, ALLOYED_ROUTE: route_vault_shares},
        )

    def evaluate_states(
        self, states: dict[str, coverage_check.PoolState], vault_balances: dict[str, int],
    ) -> coverage_check.CoverageResult:
        return coverage_check.evaluate(
            required_routes=REQUIRED_ROUTES, export=synthetic_export(), pools=pools(),
            vault_balances=vault_balances, fetch_pool_state=lambda pool_id: states[pool_id],
        )[0]

    def printed_lines(self, results: list[coverage_check.CoverageResult]) -> list[str]:
        output = io.StringIO()
        with contextlib.redirect_stdout(output):
            coverage_check.print_table(results)
        return output.getvalue().splitlines()


class PoolStateFetcherTest(unittest.TestCase):
    def test_reads_liquidity_share_denom_and_supply_from_osmosis(self) -> None:
        rest = "https://osmosis.example"
        share_denom = "factory/osmo1contract/alloyed/stuatom"
        responses = {
            "get_total_pool_liquidity": {"total_pool_liquidity": [
                {"denom": NATIVE, "amount": "1294500"}, {"denom": CANONICAL_ST, "amount": "37000"},
            ]},
            "get_share_denom": {"share_denom": share_denom},
        }

        def fake_get_json(url: str) -> dict:
            if url == f"{rest}/osmosis/poolmanager/v1beta1/pools/7":
                return {"pool": {"contract_address": "osmo1contract"}}
            smart_prefix = f"{rest}/cosmwasm/wasm/v1/contract/osmo1contract/smart/"
            if url.startswith(smart_prefix):
                (query_name,) = json.loads(base64.b64decode(urllib.parse.unquote(url.removeprefix(smart_prefix))))
                return {"data": responses[query_name]}
            # The share denom has slashes: only the URL-encoded query-parameter form addresses it
            if url == f"{rest}/cosmos/bank/v1beta1/supply/by_denom?denom=factory%2Fosmo1contract%2Falloyed%2Fstuatom":
                return {"amount": {"denom": share_denom, "amount": "1349500"}}
            raise AssertionError(f"unexpected request {url}")

        with mock.patch.object(coverage_check, "get_json", side_effect=fake_get_json):
            state = coverage_check.make_pool_state_fetcher(osmosis_rest=rest)("7")

        self.assertEqual(
            state,
            coverage_check.PoolState(
                liquidity={NATIVE: 1_294_500, CANONICAL_ST: 37_000},
                alloyed_denom=share_denom,
                alloyed_supply=1_349_500,
            ),
        )


class ApprovedRoutePolicyTest(unittest.TestCase):
    def test_policy_matches_per_token_location_tables(self) -> None:
        locations = pathlib.Path(__file__).resolve().parents[2] / "docs/wind-down/sttoken-locations.md"
        routes_by_token: dict[str, frozenset[str]] = {}
        for section in locations.read_text().split("\n## "):
            heading = section.splitlines()[0]
            token = re.fullmatch(r"st\w+ \((st\w+), host .+\)", heading)
            if token is None or ("ignored entirely" in heading and token.group(1) != "stusomm"):
                continue
            channels: set[str] = set()
            for line in section.splitlines()[1:]:
                if not line.startswith("|"):
                    continue
                cells = [cell.strip() for cell in line.strip("|").split("|")]
                if len(cells) != 7 or cells[-1] != "in scope" or cells[0] in {"Stride", "Osmosis"}:
                    continue
                self.assertRegex(cells[2], r"^channel-[0-9]+(?:, channel-[0-9]+)*$")
                channels.update(cells[2].split(", "))
            routes_by_token[token.group(1)] = frozenset(channels)
        self.assertEqual(coverage_check.REQUIRED_ROUTES, routes_by_token)
        self.assertEqual(sum(len(routes) for routes in routes_by_token.values()), 30)
        self.assertIn("channel-6", routes_by_token["stuatom"])
        self.assertTrue(all({"channel-11", "channel-69"}.isdisjoint(routes) for routes in routes_by_token.values()))
        self.assertTrue({"channel-13", "channel-52"} <= routes_by_token["stuluna"])
        self.assertIn("channel-52", routes_by_token["stuatom"])
        self.assertEqual(routes_by_token["stuband"], frozenset({"channel-0", "channel-258"}))
        self.assertEqual(routes_by_token["stusomm"], frozenset())
        self.assertIn("staISLM", routes_by_token)
        self.assertTrue(all("channel-307" not in routes for routes in routes_by_token.values()))
        self.assertEqual({token for token, routes in routes_by_token.items() if "channel-47" in routes},
                         {"stuatom", "stutia", "stinj", "stadydx", "stuluna"})


class ResultLineTest(unittest.TestCase):
    def test_pass_line(self) -> None:
        self.assertEqual("RESULT: PASS — every stToken is covered", coverage_check.render_result(uncovered=[]))

    def test_fail_line_names_the_uncovered_tokens_and_the_next_step(self) -> None:
        self.assertEqual(
            "RESULT: FAIL — 2 stToken(s) not covered: ['stuatom', 'stutia']; "
            "do not fund those pools until the shortfall is resolved",
            coverage_check.render_result(uncovered=["stuatom", "stutia"]),
        )


if __name__ == "__main__":
    unittest.main()
