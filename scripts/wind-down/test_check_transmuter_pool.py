"""Checks that the pool gate's constants agree with the Go wind-down constants.

cd scripts/wind-down && python3 -m unittest test_check_transmuter_pool
"""

import pathlib
import re
import unittest

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

    @unittest.skip("mainnet table; rehearsal branch")
    def test_every_in_scope_zone_has_an_osmosis_channel(self) -> None:
        zones = set(go_string_map(name="HostToOsmosisTransferChannel")) - {
            OSMOSIS_CHAIN_ID
        }

        self.assertIn("sommelier-3", zones)
        self.assertEqual(
            set(), zones - set(check_transmuter_pool.OSMOSIS_CHANNEL_TO_HOST)
        )

    @unittest.skip("mainnet table; rehearsal branch")
    def test_sommelier_channel_is_the_verified_one(self) -> None:
        self.assertEqual(
            "channel-165", check_transmuter_pool.OSMOSIS_CHANNEL_TO_HOST["sommelier-3"]
        )


if __name__ == "__main__":
    unittest.main()
