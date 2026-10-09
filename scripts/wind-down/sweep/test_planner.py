"""Tier membership and order, batch packing by addresses and by gas, the plan file round trip, the ladder."""

import json
import pathlib
import tempfile
import unittest
from decimal import Decimal

from sweep import config, holders, ledger, planner

DENOMS = [
    holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5"),
    holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5"),
]


def holder(index: int, usd: int, keyless: bool = False, denoms: int = 1) -> holders.Holder:
    balances = {"stuatom": usd * 1_000_000 // 6} if denoms == 1 else {"stuatom": usd * 1_000_000 // 12, "ustrd": usd * 10_000_000}
    return holders.Holder(address=f"stride1holder{index:04d}", balances=balances, usd=Decimal(usd), keyless=keyless)


def holder_set(holder_list: list[holders.Holder], below: list[holders.Holder] | None = None) -> holders.HolderSet:
    return holders.HolderSet(height=100, denoms=DENOMS, holders=sorted(holder_list, key=lambda h: h.usd, reverse=True),
                             excluded=[holders.Excluded(address="stride1excluded", reason="excluded: team: F5", usd=Decimal(9))],
                             skipped=[holders.Excluded(address="stride1escrow", reason="transfer escrow address", usd=Decimal(99))],
                             below_floor=below or [])


class TierTests(unittest.TestCase):
    def test_canary_is_the_smallest_then_main_by_value_then_keyless_last(self) -> None:
        hs = holder_set([holder(1, 500), holder(2, 50), holder(3, 5000), holder(4, 20), holder(5, 900, keyless=True), holder(6, 30)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=2, canary=2, test=False, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        tiers = {tier.name: [a.address for b in tier.batches for a in b.addresses] for tier in plan.tiers}
        self.assertEqual([tier.name for tier in plan.tiers], [planner.TierName.CANARY, planner.TierName.MAIN, planner.TierName.KEYLESS])
        self.assertEqual(tiers[planner.TierName.CANARY], ["stride1holder0004", "stride1holder0006"])
        self.assertEqual(tiers[planner.TierName.MAIN], ["stride1holder0003", "stride1holder0001", "stride1holder0002"])
        self.assertEqual(tiers[planner.TierName.KEYLESS], ["stride1holder0005"])
        self.assertEqual([b.id for tier in plan.tiers for b in tier.batches], ["002-001", "002-002", "002-003"])

    def test_test_plan_has_one_tier_with_one_address(self) -> None:
        hs = holder_set([holder(1, 0)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(100), run_id=1, canary=3, test=True, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        self.assertEqual([tier.name for tier in plan.tiers], [planner.TierName.TEST])
        self.assertEqual(plan.tiers[0].batches[0].id, "001-001")
        self.assertTrue(plan.test)

    def test_empty_tiers_are_omitted(self) -> None:
        plan = planner.build_plan(holder_set=holder_set([holder(1, 50)]), floor_usd=Decimal(10), run_id=1, canary=0, test=False,
                                  gas_per_transfer=100_000, max_addresses=100, gas_budget=40_000_000, created_at="t")
        self.assertEqual([tier.name for tier in plan.tiers], [planner.TierName.MAIN])


class PackingTests(unittest.TestCase):
    def test_splits_at_max_addresses(self) -> None:
        batches = planner.pack(holder_list=[holder(i, 10) for i in range(7)], tier=planner.TierName.MAIN, run_id=1, first_index=1,
                               gas_per_transfer=100_000, max_addresses=3, gas_budget=40_000_000)
        self.assertEqual([len(b.addresses) for b in batches], [3, 3, 1])
        self.assertEqual([b.id for b in batches], ["001-001", "001-002", "001-003"])
        self.assertEqual(batches[0].estimated_gas, 300_000)
        self.assertEqual(batches[0].usd, Decimal(30))

    def test_splits_at_the_gas_budget_counting_transfers_per_holder(self) -> None:
        # each holder has two denoms = two transfers = 200,000 gas; a 500,000 budget fits two holders
        batches = planner.pack(holder_list=[holder(i, 12, denoms=2) for i in range(5)], tier=planner.TierName.MAIN, run_id=1, first_index=4,
                               gas_per_transfer=100_000, max_addresses=100, gas_budget=500_000)
        self.assertEqual([len(b.addresses) for b in batches], [2, 2, 1])
        self.assertEqual(batches[0].transfers, 4)
        self.assertEqual(batches[0].id, "001-004")

    def test_a_single_holder_over_budget_still_gets_a_batch(self) -> None:
        batches = planner.pack(holder_list=[holder(1, 12, denoms=2)], tier=planner.TierName.MAIN, run_id=1, first_index=1,
                               gas_per_transfer=100_000, max_addresses=100, gas_budget=50_000)
        self.assertEqual(len(batches), 1)


class LadderAndCalibrationTests(unittest.TestCase):
    def test_ladder_counts_holders_at_each_floor(self) -> None:
        rungs = planner.ladder(holder_list=[holder(1, 50), holder(2, 7), holder(3, 3), holder(4, 0)], floor_usd=Decimal(25))
        self.assertEqual([(r.floor, r.holders, r.usd) for r in rungs], [
            (Decimal(25), 1, Decimal(50)), (Decimal(10), 1, Decimal(50)), (Decimal(5), 2, Decimal(57)),
            (Decimal(1), 3, Decimal(60)), (Decimal(0), 4, Decimal(60)),
        ])

    def test_gas_per_transfer_falls_back_to_the_default(self) -> None:
        self.assertEqual(planner.gas_per_transfer(events=[]), config.GAS_PER_TRANSFER_DEFAULT)
        t = ledger.Transfer(address="a", denom="d", amount=1, channel="c", receiver="r")
        events = [ledger.confirmed(batch_id="001-001", tx_hash="A", at="t", height=1, gas_used=100_000, transfers=[t], skipped=[])]
        self.assertEqual(planner.gas_per_transfer(events=events), 120_000)
        self.assertEqual(planner.next_run_id(events=[ledger.submitted(run_id=4, batch_id="004-001", tx_hash="A", at="t", addresses=1, transfers_estimate=1, gas_wanted=1)]), 5)


class PlanFileTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_write_then_load_round_trips_and_replaces_old_batch_files(self) -> None:
        (self.state / "batch-000-009.txt").write_text("stale\n")
        hs = holder_set([holder(1, 500), holder(2, 50)], below=[holder(9, 3)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=1, canary=0, test=False, gas_per_transfer=100_000,
                                  max_addresses=1, gas_budget=40_000_000, created_at="2026-11-10T09:00:00+00:00")
        planner.write_plan(plan=plan, state_dir=self.state)

        self.assertFalse((self.state / "batch-000-009.txt").exists())
        self.assertEqual((self.state / "batch-001-001.txt").read_text(), "stride1holder0001\n")
        loaded = planner.load_plan(path=self.state / "plan.json")
        self.assertEqual(loaded, plan)
        self.assertEqual(loaded.batch(batch_id="001-002").sha256, planner.batch_sha256(content="stride1holder0002\n"))
        self.assertEqual(loaded.below_floor_count, 1)
        self.assertEqual(loaded.below_floor_usd, Decimal(3))
        raw = json.loads((self.state / "plan.json").read_text())
        self.assertEqual(raw["tiers"][0]["batches"][0]["addresses"][0]["balances"]["stuatom"], str(500 * 1_000_000 // 6))
        self.assertIsNone(planner.load_plan(path=self.state / "missing.json"))

    def test_load_plan_wraps_malformed_files_in_plan_error(self) -> None:
        hs = holder_set([holder(1, 500)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=1, canary=0, test=False, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        good = planner.plan_to_dict(plan=plan)
        missing_key = {key: value for key, value in good.items() if key != "height"}
        cases = {
            "bad decimal": json.dumps({**good, "floor_usd": "not-a-number"}),
            "list top level": json.dumps([1, 2]),
            "missing key": json.dumps(missing_key),
            "invalid json": "{not json",
        }
        for name, text in cases.items():
            with self.subTest(name):
                path = self.state / "plan.json"
                path.write_text(text)
                with self.assertRaises(planner.PlanError):
                    planner.load_plan(path=path)

    def test_pending_batches_follow_tier_order(self) -> None:
        hs = holder_set([holder(1, 500), holder(2, 50, keyless=True), holder(3, 20)])
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=1, canary=1, test=False, gas_per_transfer=100_000,
                                  max_addresses=100, gas_budget=40_000_000, created_at="t")
        self.assertEqual([(tier.name, batch.id) for tier, batch in plan.pending_batches()],
                         [(planner.TierName.CANARY, "001-001"), (planner.TierName.MAIN, "001-002"), (planner.TierName.KEYLESS, "001-003")])


if __name__ == "__main__":
    unittest.main()
