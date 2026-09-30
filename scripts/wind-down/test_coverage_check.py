"""Unit test for coverage_check.evaluate on a synthetic export with an injected fetcher."""

import unittest
from decimal import Decimal

import coverage_check

STRIDE_ESCROW_CH0 = coverage_check.escrow_address(channel_id="channel-0")
STRIDE_ESCROW_CH5 = coverage_check.escrow_address(channel_id="channel-5")
HOLDER = "stride1qqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqqq"
VAULT = "osmo1vault"
NATIVE = "ibc/NATIVE"
CANONICAL_ST = "ibc/STUATOM"
ROUTE_ST = "ibc/STUATOM_TWO_HOP"


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

    def test_rate_rounds_up_to_base_units(self) -> None:
        export = synthetic_export()
        export["app_state"]["stakeibc"]["host_zone_list"][0]["redemption_rate"] = "1.333333333333333333"
        result = coverage_check.evaluate(
            export=export,
            pools=pools(),
            vault_balances={NATIVE: 2_000_000},
            fetch_liquidity=lambda pool_id: {NATIVE: 0},
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
            export=synthetic_export(),
            pools=pools(),
            vault_balances={},
            fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )[0]
        self.assertEqual(result.required_native, 1_140_000)
        self.assertEqual(result.route_expected["channel-0"], 90_000)
        self.assertEqual(result.canonical_expected, 1_050_000)
        self.assertEqual(result.route_shortfalls, {})
        self.assertTrue(result.covered)

    def test_partly_redeemed_route_pool_short_of_the_reduced_requirement_fails(self) -> None:
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 89_999, ROUTE_ST: 40_000}}
        result = coverage_check.evaluate(
            export=synthetic_export(), pools=pools(), vault_balances={}, fetch_liquidity=lambda pool_id: liquidity[pool_id],
        )[0]
        self.assertEqual(result.route_shortfalls, {"channel-0": 1})
        self.assertFalse(result.covered)

    def test_route_pool_holding_more_than_escrow_times_rate_fails(self) -> None:
        # escrow 100,000 x 1.5 = 150,000; 150,001 is over-funded even though every total is met
        liquidity = {"1": {NATIVE: 1_350_000}, "2": {NATIVE: 150_001}}
        result = coverage_check.evaluate(
            export=synthetic_export(), pools=pools(), vault_balances={}, fetch_liquidity=lambda pool_id: liquidity[pool_id],
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
                    export=synthetic_export(), pools=bad, vault_balances={}, fetch_liquidity=lambda pool_id: {},
                )

    def test_route_pool_without_st_denom_is_an_error(self) -> None:
        bad = pools()
        del bad["stuatom"]["route_pools"][0]["st_denom"]
        with self.assertRaises(coverage_check.CoverageInputError):
            coverage_check.evaluate(
                export=synthetic_export(), pools=bad, vault_balances={}, fetch_liquidity=lambda pool_id: {},
            )

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
