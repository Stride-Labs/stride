import unittest

import route_split

ST = "staISLM"
CHANNEL = "channel-240"
TOKEN = 10**18


def reading(key: str, amount: int, source: str = route_split.LIVE_SOURCE, at: str = "2026-10-20T21:00:00+00:00") -> route_split.Reading:
    return route_split.Reading(at=at, source=source, height=None, st_denom=ST, key=key, amount=amount)


def zone(route_pool: dict[str, object] | None = None, canonical: dict[str, object] | None = None) -> dict[str, object]:
    pools = [
        {"kind": "canonical", "pool_id": "3641", "native_marked": False, "rate": "1.0", "vault_shares": "0", **(canonical or {})},
        {
            "kind": "route",
            "pool_id": "3655",
            "route": {"chain_id": "haqq_11235-1", "stride_channel": CHANNEL},
            "native_marked": False,
            "rate": "1.0",
            "vault_shares": "0",
            **(route_pool or {}),
        },
    ]
    return {"chain_id": "haqq_11235-1", "st_denom": ST, "decimals": "18", "pools": pools, "planned": []}


class ChangesTest(unittest.TestCase):
    def test_only_amounts_that_moved_since_the_last_live_reading_are_recorded(self) -> None:
        history = [reading("supply", 100 * TOKEN), reading(CHANNEL, 40 * TOKEN)]
        now = [reading("supply", 100 * TOKEN), reading(CHANNEL, 39 * TOKEN)]

        self.assertEqual(route_split.changes(history=history, readings=now), [now[1]])

    def test_an_upgrade_reading_is_never_the_live_baseline(self) -> None:
        history = [reading("supply", 100 * TOKEN, source=route_split.UPGRADE_SOURCE)]
        now = [reading("supply", 100 * TOKEN)]

        self.assertEqual(route_split.changes(history=history, readings=now), now)


class ZoneSplitTest(unittest.TestCase):
    def test_shares_against_the_first_live_reading_with_the_upgrade_for_reference(self) -> None:
        readings = [
            reading("supply", 120 * TOKEN, source=route_split.UPGRADE_SOURCE),
            reading(CHANNEL, 30 * TOKEN, source=route_split.UPGRADE_SOURCE),
            reading("supply", 100 * TOKEN),
            reading(CHANNEL, 40 * TOKEN),
        ]

        split = route_split.zone_split(zone=zone(), readings=readings)

        route, canonical = split.rows
        self.assertEqual((route.label, route.pool_id, route.now, route.share_now, route.share_at_upgrade), ("haqq_11235-1 · channel-240", "3655", 40 * TOKEN, "0.400000", "0.250000"))
        self.assertEqual((canonical.now, canonical.share_now, canonical.at_upgrade), (60 * TOKEN, "0.600000", 90 * TOKEN))
        self.assertFalse(split.drifted)  # far from the upgrade, but that is reference only
        self.assertEqual([row.matches_funding for row in split.rows], [None, None])  # nothing funded yet

    def test_a_move_of_more_than_one_stToken_since_the_baseline_turns_the_route_and_canonical_red(self) -> None:
        readings = [reading("supply", 100 * TOKEN), reading(CHANNEL, 40 * TOKEN), reading(CHANNEL, 38 * TOKEN)]

        split = route_split.zone_split(zone=zone(), readings=readings)

        route, canonical = split.rows
        self.assertEqual((route.moved, route.drifted, canonical.moved, canonical.drifted, split.drifted), (2 * TOKEN, True, 2 * TOKEN, True, True))

    def test_one_stToken_or_less_is_steady(self) -> None:
        readings = [reading("supply", 100 * TOKEN), reading(CHANNEL, 40 * TOKEN), reading(CHANNEL, 40 * TOKEN - TOKEN)]

        self.assertFalse(route_split.zone_split(zone=zone(), readings=readings).drifted)

    def test_after_funding_a_move_against_the_funded_amount_is_red_and_the_canonical_surplus_is_not(self) -> None:
        # The route was funded for 40 stTokens; the canonical pool also holds a surplus beyond its 60.
        funded_route = {"native_marked": True, "vault_shares": str(int(40 * TOKEN * 1.06)), "rate": "1.06"}
        surplus_canonical = {"native_marked": True, "vault_shares": str(70 * TOKEN)}
        readings = [reading("supply", 100 * TOKEN), reading(CHANNEL, 40 * TOKEN)]

        steady = route_split.zone_split(zone=zone(route_pool=funded_route, canonical=surplus_canonical), readings=readings)
        self.assertEqual([row.at_funding for row in steady.rows], [40 * TOKEN, 60 * TOKEN])
        self.assertEqual([row.matches_funding for row in steady.rows], [True, True])
        self.assertFalse(steady.drifted)

        moved = route_split.zone_split(zone=zone(route_pool=funded_route, canonical=surplus_canonical), readings=[*readings, reading(CHANNEL, 35 * TOKEN)])
        self.assertEqual([row.moved for row in moved.rows], [5 * TOKEN, 5 * TOKEN])
        self.assertEqual([row.matches_funding for row in moved.rows], [False, False])
        self.assertTrue(moved.drifted)

    def test_nothing_recorded_is_no_split(self) -> None:
        self.assertIsNone(route_split.zone_split(zone=zone(), readings=[]))


if __name__ == "__main__":
    unittest.main()
