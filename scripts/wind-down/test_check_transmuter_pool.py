"""Checks that the pool gate's constants agree with the Go wind-down constants.

cd scripts/wind-down && python3 -m unittest test_check_transmuter_pool
"""

import contextlib
import io
import pathlib
import re
import sys
import unittest
from unittest import mock

import check_transmuter_pool

WIND_DOWN_GO = (
    pathlib.Path(__file__).resolve().parents[2]
    / "x"
    / "stakeibc"
    / "types"
    / "wind_down.go"
)
OSMOSIS_CHAIN_ID = "osmosis-1"  # has no transfer channel to itself ("" in the Go map)


def go_source() -> str:
    return WIND_DOWN_GO.read_text()


def go_string_constant(name: str) -> str:
    match = re.search(rf'\b{name}\s*=\s*"([^"]*)"', go_source())
    assert match, f"{name} not found in {WIND_DOWN_GO}"
    return match.group(1)


def go_string_map(name: str) -> dict[str, str]:
    match = re.search(
        rf"\b{name}\s*=\s*map\[string\]string\{{(.*?)\n\}}",
        go_source(),
        flags=re.DOTALL,
    )
    assert match, f"{name} not found in {WIND_DOWN_GO}"
    return dict(re.findall(r'"([^"]+)":\s*"([^"]*)"', match.group(1)))


class CheckTransmuterPoolConstantsTest(unittest.TestCase):
    def test_admin_and_moderator_are_the_osmosis_vault(self) -> None:
        vault = go_string_constant(name="OsmosisVaultAddress")

        self.assertTrue(vault.startswith("osmo1"))
        self.assertEqual(vault, check_transmuter_pool.ADMIN)
        self.assertEqual(vault, check_transmuter_pool.MODERATOR)

    def test_every_in_scope_zone_has_an_osmosis_channel(self) -> None:
        zones = set(go_string_map(name="HostToOsmosisTransferChannel")) - {
            OSMOSIS_CHAIN_ID
        }

        self.assertIn("sommelier-3", zones)
        self.assertEqual(
            set(), zones - set(check_transmuter_pool.OSMOSIS_CHANNEL_TO_HOST)
        )

    def test_sommelier_channel_is_the_verified_one(self) -> None:
        self.assertEqual(
            "channel-165", check_transmuter_pool.OSMOSIS_CHANNEL_TO_HOST["sommelier-3"]
        )


class CorruptedAssetsCheckTest(unittest.TestCase):
    NATIVE = "ibc/NATIVE"
    ST_TOKEN = "ibc/STTOKEN"

    def check(self, corrupted: list[str], funded: bool) -> bool:
        _, ok = check_transmuter_pool.corrupted_assets_check(
            corrupted=corrupted, native_denom=self.NATIVE, funded=funded
        )
        return ok

    def test_before_funding_nothing_may_be_corrupted(self) -> None:
        self.assertTrue(self.check(corrupted=[], funded=False))
        self.assertFalse(self.check(corrupted=[self.NATIVE], funded=False))

    def test_after_funding_exactly_the_native_token_is_corrupted(self) -> None:
        self.assertTrue(self.check(corrupted=[self.NATIVE], funded=True))
        self.assertFalse(self.check(corrupted=[], funded=True))
        self.assertFalse(self.check(corrupted=[self.ST_TOKEN], funded=True))
        self.assertFalse(self.check(corrupted=[self.NATIVE, self.ST_TOKEN], funded=True))


class ResultLineTest(unittest.TestCase):
    def run_main(self, failing_pool_ids: set[str]) -> tuple[int, list[str]]:
        def fake_check_pool(
            spec: check_transmuter_pool.PoolSpec, use_color: bool, funded: bool
        ) -> check_transmuter_pool.Report:
            report = check_transmuter_pool.Report(use_color=use_color)
            if spec.pool_id in failing_pool_ids:
                report.counts[check_transmuter_pool.Outcome.FAIL] += 1
            return report

        stdout = io.StringIO()
        with (
            mock.patch.object(check_transmuter_pool, "check_pool", fake_check_pool),
            mock.patch.object(sys, "argv", ["check_transmuter_pool.py"]),
            contextlib.redirect_stdout(stdout),
        ):
            status = check_transmuter_pool.main()
        return status, stdout.getvalue().splitlines()

    def test_all_pools_passing_ends_with_a_pass_line(self) -> None:
        status, lines = self.run_main(failing_pool_ids=set())

        self.assertEqual(0, status)
        self.assertEqual(f"RESULT: PASS — all {len(check_transmuter_pool.POOLS)} pool(s) passed", lines[-1])

    def test_a_failing_pool_ends_with_a_fail_line_naming_it(self) -> None:
        failing = check_transmuter_pool.POOLS[0].pool_id

        status, lines = self.run_main(failing_pool_ids={failing})

        self.assertEqual(1, status)
        self.assertEqual(
            f"RESULT: FAIL — pools {failing} have failing checks; fix them (or the CONSTANTS block) and rerun",
            lines[-1],
        )


if __name__ == "__main__":
    unittest.main()
