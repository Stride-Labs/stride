"""compose(): address states from the plan, the ledger and live balances; totals; by-denom; ladder; batches."""

import dataclasses
import pathlib
import sys
import unittest
from decimal import Decimal

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package, before any import of it

import sweep_tab  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402

STATOM = holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5")
STRD = holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5")
A_SWEPT, A_REFUNDED, A_PENDING, A_EXCLUDED, A_DUST, A_NEW, A_ESCROW = (f"stride1{name}" for name in ("swept", "refund", "pend", "excl", "dust", "new", "escrow"))


def make_plan() -> planner.Plan:
    def h(address: str, statom: int, keyless: bool = False) -> holders.Holder:
        return holders.Holder(address=address, balances={"stuatom": statom, "ustrd": 1_000_000}, usd=Decimal(statom) / 1_000_000 * 6 + Decimal("0.05"), keyless=keyless)
    hs = holders.HolderSet(
        height=100, denoms=[STATOM, STRD],
        holders=[h(A_SWEPT, 10_000_000), h(A_REFUNDED, 5_000_000), h(A_PENDING, 2_000_000, keyless=True)],
        excluded=[holders.Excluded(address=A_EXCLUDED, reason="excluded: team: F5", usd=Decimal(99))],
        skipped=[holders.Excluded(address=A_ESCROW, reason="transfer escrow address", usd=Decimal(1000))],
        below_floor=[h(A_DUST, 100_000)],
    )
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=2, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=1, gas_budget=40_000_000, operator_sequence=5, created_at="2026-11-10T09:00:00+00:00")


def t(address: str, denom: str, amount: int) -> ledger.Transfer:
    return ledger.Transfer(address=address, denom=denom, amount=amount, channel="channel-5", receiver="osmo1x")


EVENTS = [
    ledger.submitted(run_id=2, batch_id="002-001", tx_hash="A", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
    ledger.confirmed(batch_id="002-001", tx_hash="A", at="t", height=5, gas_used=1, transfers=[t(A_SWEPT, "stuatom", 10_000_000), t(A_SWEPT, "ustrd", 1_000_000)], skipped=[]),
    ledger.submitted(run_id=2, batch_id="002-002", tx_hash="B", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
    ledger.confirmed(batch_id="002-002", tx_hash="B", at="t", height=6, gas_used=1, transfers=[t(A_REFUNDED, "stuatom", 5_000_000), t(A_REFUNDED, "ustrd", 1_000_000)],
                     skipped=[ledger.Skip(address="stride1other", reason="interchain account")]),
]
LIVE = {
    "height": "200",
    "denoms": [{"denom": "stuatom", "symbol": "stATOM", "decimals": "6", "price_usd": "6", "destination": "osmosis-1", "channel": "channel-5"},
               {"denom": "ustrd", "symbol": "STRD", "decimals": "6", "price_usd": "0.05", "destination": "osmosis-1", "channel": "channel-5"}],
    "balances": {
        A_REFUNDED: {"stuatom": "5000000"},  # the stATOM transfer timed out and came back; the STRD landed
        A_PENDING: {"stuatom": "2000000", "ustrd": "1000000"},
        A_EXCLUDED: {"stuatom": "16500000"},
        A_DUST: {"stuatom": "100000"},
        A_NEW: {"stuatom": "3000000"},  # above the floor but not in the plan: appeared since
        A_ESCROW: {"stuatom": "999000000"},
    },
    "operator_strd": "212400000",
}


class ComposeTests(unittest.TestCase):
    def setUp(self) -> None:
        self.data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=LIVE, exclusions={}, live_fetched_at="2026-11-10T14:00:00+00:00")

    def test_run_and_batches(self) -> None:
        run = self.data["run"]
        self.assertEqual((run["run_id"], run["floor_usd"], run["height"]), ("2", "10", "100"))
        self.assertEqual(run["batches"], {"total": "3", "confirmed": "2", "submitted": "0", "failed": "0", "lost": "0", "pending": "1"})
        self.assertEqual(run["operator_strd"], "212400000")
        batches = {b["batch_id"]: b for b in self.data["batches"]}
        self.assertEqual(batches["002-001"]["state"], "confirmed")
        self.assertEqual(batches["002-002"]["skipped"], [{"address": "stride1other", "reason": "interchain account"}])
        self.assertEqual(batches["002-002"]["refunded"], "1")
        self.assertEqual((batches["002-003"]["state"], batches["002-003"]["tier"], batches["002-003"]["tx_hash"]), ("pending", "keyless", None))

    def test_address_states_and_totals(self) -> None:
        totals = self.data["totals"]
        self.assertEqual(totals["swept"], {"addresses": "1", "usd": "60.05"})
        self.assertEqual(totals["refunded"], {"addresses": "1", "usd": "30.00"})  # the stATOM that came back, at plan prices
        self.assertEqual(totals["remaining"], {"addresses": "1", "usd": "12.05"})
        self.assertEqual(totals["excluded"], {"addresses": "1", "usd": "99.00"})
        self.assertEqual(totals["below_floor"], {"addresses": "1", "usd": "0.60"})
        self.assertEqual(totals["unplanned"], {"addresses": "1", "usd": "18.00"})
        self.assertEqual(self.data["refunded"], [{"address": A_REFUNDED, "denom": "stuatom", "amount": "5000000", "channel": "channel-5", "batch_id": "002-002"}])

    def test_by_denom(self) -> None:
        by_denom = {d["denom"]: d for d in self.data["by_denom"]}
        self.assertEqual(by_denom["stuatom"]["swept"], {"addresses": "2", "amount": "15000000", "usd": "90.00"})
        self.assertEqual(by_denom["stuatom"]["remaining"], {"addresses": "1", "amount": "2000000", "usd": "12.00"})
        self.assertEqual(by_denom["stuatom"]["refunded_addresses"], "1")
        self.assertEqual(by_denom["stuatom"]["excluded_usd"], "99.00")
        self.assertEqual(by_denom["stuatom"]["below_floor_usd"], "0.60")
        self.assertEqual(by_denom["ustrd"]["destination"], "osmosis-1")

    def test_ladder_counts_live_holders_still_to_sweep(self) -> None:
        # not yet swept and not excluded/skipped: refunded (30), pending (12.05), new (18), dust (0.60)
        self.assertEqual([(r["floor"], r["holders"], r["usd"]) for r in self.data["ladder"]],
                         [("10", "3", "60.05"), ("5", "3", "60.05"), ("1", "3", "60.05"), ("0", "4", "60.65")])

    def test_keyless_and_exclusions_sections(self) -> None:
        self.assertEqual(self.data["keyless"], {"addresses": "1", "usd": "12.05", "swept": "0"})
        data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=LIVE,
                                 exclusions={A_EXCLUDED: holders.Exclusion(address=A_EXCLUDED, section="team", label="F5", reason="by hand")},
                                 live_fetched_at="x")
        self.assertEqual(data["exclusions"], [{"section": "team", "address": A_EXCLUDED, "label": "F5", "reason": "by hand", "live_usd": "99.00"}])


def vesting_plan(run_id: int, addresses: list[str], locked: dict[str, int]) -> planner.Plan:
    holder_list = [holders.Holder(address=address, balances={"stuatom": 1_000_000}, usd=Decimal(6), keyless=False, locked=dict(locked)) for address in addresses]
    hs = holders.HolderSet(height=100, denoms=[STATOM, STRD], holders=holder_list, excluded=[], skipped=[], below_floor=[])
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(1), run_id=run_id, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=10, gas_budget=40_000_000, operator_sequence=5, created_at="2026-11-10T09:00:00+00:00")


def live_with(balances: dict[str, dict[str, str]]) -> dict:
    return {**LIVE, "balances": balances}


def confirmed_run(run_id: int, address: str, amount: int = 1_000_000) -> list[ledger.Event]:
    batch_id = f"{run_id:03d}-001"
    return [ledger.submitted(run_id=run_id, batch_id=batch_id, tx_hash="V", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.confirmed(batch_id=batch_id, tx_hash="V", at="t", height=5, gas_used=1, transfers=[t(address, "stuatom", amount)], skipped=[])]


class VestingTests(unittest.TestCase):
    """A vesting holder keeps its locked remainder after the sweep; only a balance above that is a refund."""

    def compose(self, live_amount: str) -> dict:
        plan = vesting_plan(run_id=1, addresses=["stride1vest"], locked={"stuatom": 400_000})
        return sweep_tab.compose(plan=plan, events=confirmed_run(run_id=1, address="stride1vest"),
                                 live=live_with({"stride1vest": {"stuatom": live_amount}}), exclusions={}, live_fetched_at="x")

    def test_holding_exactly_the_locked_remainder_is_swept(self) -> None:
        data = self.compose("400000")
        self.assertEqual(data["totals"]["swept"], {"addresses": "1", "usd": "6.00"})
        self.assertEqual(data["totals"]["refunded"], {"addresses": "0", "usd": "0.00"})
        self.assertEqual(data["refunded"], [])
        self.assertEqual(data["batches"][0]["refunded"], "0")
        self.assertEqual({d["denom"]: d["refunded_addresses"] for d in data["by_denom"]}["stuatom"], "0")

    def test_holding_more_than_the_locked_remainder_is_refunded_for_the_excess(self) -> None:
        data = self.compose("1400000")
        self.assertEqual(data["totals"]["swept"], {"addresses": "0", "usd": "0.00"})
        self.assertEqual(data["totals"]["refunded"], {"addresses": "1", "usd": "6.00"})  # 1.4 held - 0.4 locked = 1 stATOM back
        self.assertEqual(len(data["refunded"]), 1)
        self.assertEqual(data["batches"][0]["refunded"], "1")
        self.assertEqual({d["denom"]: d["refunded_addresses"] for d in data["by_denom"]}["stuatom"], "1")


class EarlierRunLockedTests(unittest.TestCase):
    def test_a_vesting_holder_swept_in_run_one_and_only_in_run_twos_locked_map_is_not_refunded(self) -> None:
        plan = vesting_plan(run_id=2, addresses=["stride1pending"], locked={})
        plan = dataclasses.replace(plan, locked={"stride1vest": {"stuatom": 400_000}})
        live = live_with({"stride1vest": {"stuatom": "400000"}, "stride1pending": {"stuatom": "1000000"}})
        data = sweep_tab.compose(plan=plan, events=confirmed_run(run_id=1, address="stride1vest"), live=live, exclusions={}, live_fetched_at="x")

        self.assertEqual(data["totals"]["swept"], {"addresses": "1", "usd": "6.00"})
        self.assertEqual(data["totals"]["refunded"], {"addresses": "0", "usd": "0.00"})
        self.assertEqual(data["refunded"], [])


class EarlierRunTests(unittest.TestCase):
    def test_addresses_swept_in_an_earlier_run_stay_in_swept_after_a_replan(self) -> None:
        events = confirmed_run(run_id=1, address="stride1runone", amount=3_000_000) + confirmed_run(run_id=2, address="stride1runtwo", amount=2_000_000)
        plan = vesting_plan(run_id=2, addresses=["stride1runtwo", "stride1pending"], locked={})  # run 1's holder is not in this plan
        live = live_with({"stride1pending": {"stuatom": "1000000"}})
        data = sweep_tab.compose(plan=plan, events=events, live=live, exclusions={}, live_fetched_at="x")

        self.assertEqual(data["totals"]["swept"], {"addresses": "2", "usd": "30.00"})  # 3 + 2 stATOM at $6
        self.assertEqual(data["totals"]["remaining"], {"addresses": "1", "usd": "6.00"})  # only the current plan's unswept address
        stuatom = {d["denom"]: d for d in data["by_denom"]}["stuatom"]
        self.assertEqual(stuatom["swept"], {"addresses": "2", "amount": "5000000", "usd": "30.00"})
        self.assertEqual(stuatom["remaining"]["addresses"], "1")

    def test_an_earlier_runs_refunded_holder_is_listed_and_not_called_unplanned(self) -> None:
        events = confirmed_run(run_id=1, address="stride1runone", amount=3_000_000)
        plan = vesting_plan(run_id=2, addresses=["stride1pending"], locked={})
        live = live_with({"stride1runone": {"stuatom": "3000000"}, "stride1pending": {"stuatom": "1000000"}})
        data = sweep_tab.compose(plan=plan, events=events, live=live, exclusions={}, live_fetched_at="x")

        self.assertEqual(data["totals"]["refunded"], {"addresses": "1", "usd": "18.00"})
        self.assertEqual(data["totals"]["unplanned"], {"addresses": "0", "usd": "0.00"})
        self.assertEqual([entry["address"] for entry in data["refunded"]], ["stride1runone"])


class EdgeTests(unittest.TestCase):
    def test_no_plan(self) -> None:
        data = sweep_tab.compose(plan=None, events=[], live=None, exclusions={}, live_fetched_at=None)
        self.assertIsNone(data["run"])
        self.assertEqual(data["batches"], [])

    def test_no_live_snapshot_keeps_plan_and_ledger_parts(self) -> None:
        data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=None, exclusions={}, live_fetched_at=None)
        self.assertEqual(data["run"]["batches"]["confirmed"], "2")
        self.assertIsNone(data["totals"])
        self.assertIsNone(data["ladder"])
        self.assertEqual(len(data["batches"]), 3)


if __name__ == "__main__":
    unittest.main()
