"""Unit test for coverage_check.evaluate on a synthetic export with an injected fetcher."""

import unittest
from decimal import Decimal

import coverage_check

STRIDE_ESCROW_CH0 = coverage_check.escrow_address(channel_id="channel-0")
STRIDE_ESCROW_CH5 = coverage_check.escrow_address(channel_id="channel-5")
HOLDER = "stride1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"
VAULT = "osmo1vault"
NATIVE = "ibc/NATIVE"


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
            "route_pools": [{"channel_id": "channel-0", "pool_id": "2"}],
        }
    }


class CoverageCheckTest(unittest.TestCase):
    def test_covered_when_every_pool_holds_its_share(self) -> None:
        # supply 1,000,000 × 1.5 = 1,500,000 native needed; route escrow 100,000 × 1.5 = 150,000
        # in the route pool; the canonical pool gets everything else; the vault holds a surplus
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_000}}
        results = coverage_check.evaluate(
            export=synthetic_export(),
            pools=pools(),
            vault_balances={NATIVE: 10_000},
            fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )
        self.assertEqual(len(results), 1)
        result = results[0]
        self.assertEqual(result.st_denom, "stuatom")
        self.assertEqual(result.required_native, 1_500_000)
        self.assertEqual(result.native_on_osmosis, 1_510_000)
        self.assertEqual(result.route_shortfalls, {})
        self.assertTrue(result.covered)

    def test_shortfall_in_total_and_in_a_route_pool(self) -> None:
        liquidity = {"1": {NATIVE: 1_000_000}, "2": {NATIVE: 149_999}}
        result = coverage_check.evaluate(
            export=synthetic_export(),
            pools=pools(),
            vault_balances={},
            fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )[0]
        self.assertFalse(result.covered)
        self.assertEqual(result.native_on_osmosis, 1_149_999)
        self.assertEqual(result.route_shortfalls, {"channel-0": 1})
        self.assertEqual(result.canonical_expected, 1_500_000 - 150_000)

    def test_rate_rounds_down_to_base_units(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"][0]["redemption_rate"] = "1.333333333333333333"
        result = coverage_check.evaluate(
            export=export,
            pools=pools(),
            vault_balances={NATIVE: 2_000_000},
            fetch_liquidity=lambda pool_id: {NATIVE: 0},
        )[0]
        self.assertEqual(result.required_native, int(Decimal("1000000") * Decimal("1.333333333333333333")))
        self.assertEqual(result.route_expected["channel-0"], 133_333)

    def test_missing_pool_entry_is_an_error(self) -> None:
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                export=synthetic_export(),
                pools={},
                vault_balances={},
                fetch_liquidity=lambda pool_id: {},
            )

    def test_sttoken_missing_from_supply_is_an_error(self) -> None:
        # A typo in the pools file must not turn into "required 0, covered"
        mistyped = {"stuatom-typo": pools()["stuatom"]}
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                export=synthetic_export(),
                pools=mistyped,
                vault_balances={},
                fetch_liquidity=lambda pool_id: {},
            )

    def test_in_scope_sttoken_missing_from_pools_is_an_error(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "juno-1", "host_denom": "ujuno", "redemption_rate": "1.1"}
        )
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                export=export, pools=pools(), vault_balances={}, fetch_liquidity=lambda pool_id: {},
            )

    def test_deprecated_zone_is_out_of_scope(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"].append(
            {"chain_id": "comdex-1", "host_denom": "ucmdx", "redemption_rate": "1.1", "deprecated": True}
        )
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_000}}
        results = coverage_check.evaluate(
            export=export, pools=pools(), vault_balances={}, fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )
        self.assertEqual([result.st_denom for result in results], ["stuatom"])


if __name__ == "__main__":
    unittest.main()
