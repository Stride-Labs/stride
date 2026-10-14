"""compose(): address states from the plan, the ledger and live balances; totals; by-denom; ladder; batches."""

import dataclasses
import datetime
import pathlib
import sys
import unittest
from unittest import mock
from decimal import Decimal

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package, before any import of it

import sweep_tab  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402

STATOM = holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5")
STRD = holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5")
A_SWEPT, A_REFUNDED, A_PENDING, A_EXCLUDED, A_DUST, A_NEW, A_ESCROW, A_DEFERRED = (f"stride1{name}" for name in ("swept", "refund", "pend", "excl", "dust", "new", "escrow", "defer"))
FETCHED_AT = "2026-11-10T14:00:00+00:00"
EARLIER = "2026-11-10T13:00:00+00:00"  # a confirmation the snapshot already covers
LATER = "2026-11-10T14:05:00+00:00"  # a confirmation newer than the snapshot
DEFERRED_COMPLETES = "2026-10-26T08:55:25.469915628Z"


def make_plan() -> planner.Plan:
    def h(address: str, statom: int, keyless: bool = False) -> holders.Holder:
        return holders.Holder(address=address, balances={"stuatom": statom, "ustrd": 1_000_000}, usd=Decimal(statom) / 1_000_000 * 6 + Decimal("0.05"), keyless=keyless)
    hs = holders.HolderSet(
        height=100, denoms=[STATOM, STRD],
        holders=[h(A_SWEPT, 10_000_000), h(A_REFUNDED, 5_000_000), h(A_PENDING, 2_000_000, keyless=True)],
        excluded=[holders.Excluded(address=A_EXCLUDED, reason="excluded: team: F5", usd=Decimal(99))],
        skipped=[holders.Excluded(address=A_ESCROW, reason="transfer escrow address", usd=Decimal(1000))],
        below_floor=[h(A_DUST, 100_000)],
        deferred=[holders.Deferred(holder=h(A_DEFERRED, 4_000_000), staked=holders.Staked(ustrd=7_500_000, completes_at=DEFERRED_COMPLETES))],
    )
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(10), run_id=2, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=1, gas_budget=40_000_000, operator_sequence=5, created_at="2026-11-10T09:00:00+00:00",
                              staked_ustrd=9_000_000)


def t(address: str, denom: str, amount: int) -> ledger.Transfer:
    return ledger.Transfer(address=address, denom=denom, amount=amount, channel="channel-5", receiver="osmo1x")


EVENTS = [
    ledger.submitted(run_id=2, batch_id="002-001", tx_hash="A", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
    ledger.confirmed(batch_id="002-001", tx_hash="A", at=EARLIER, height=5, gas_used=1, transfers=[t(A_SWEPT, "stuatom", 10_000_000), t(A_SWEPT, "ustrd", 1_000_000)], skipped=[]),
    ledger.submitted(run_id=2, batch_id="002-002", tx_hash="B", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
    ledger.confirmed(batch_id="002-002", tx_hash="B", at=EARLIER, height=6, gas_used=1, transfers=[t(A_REFUNDED, "stuatom", 5_000_000), t(A_REFUNDED, "ustrd", 1_000_000)],
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
        A_DEFERRED: {"stuatom": "3000000", "ustrd": "2000000"},  # the live balance, not the plan's
    },
    "operator_strd": "212400000",
    "staked_ustrd": "9000000",
}


class ComposeTests(unittest.TestCase):
    def setUp(self) -> None:
        self.data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=LIVE, exclusions={}, live_fetched_at=FETCHED_AT)

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
        self.assertEqual(totals["unplanned"], {"addresses": "1", "usd": "18.00"})  # the deferred holder is neither unplanned nor below floor
        self.assertEqual(totals["deferred"], {"addresses": "1", "usd": "18.10"})  # live: 3 stATOM at $6 + 2 STRD at $0.05
        self.assertEqual(self.data["refunded"], [{"address": A_REFUNDED, "denom": "stuatom", "amount": "5000000", "channel": "channel-5", "batch_id": "002-002"}])

    def test_by_denom(self) -> None:
        by_denom = {d["denom"]: d for d in self.data["by_denom"]}
        self.assertEqual(by_denom["stuatom"]["swept"], {"addresses": "2", "amount": "15000000", "usd": "90.00"})
        self.assertEqual(by_denom["stuatom"]["remaining"], {"addresses": "1", "amount": "2000000", "usd": "12.00"})
        self.assertEqual(by_denom["stuatom"]["refunded_addresses"], "1")
        self.assertEqual(by_denom["stuatom"]["excluded_usd"], "99.00")
        self.assertEqual(by_denom["stuatom"]["below_floor_usd"], "0.60")
        self.assertEqual(by_denom["stuatom"]["deferred_usd"], "18.00")
        self.assertEqual(by_denom["ustrd"]["deferred_usd"], "0.10")
        self.assertEqual(by_denom["ustrd"]["below_floor_usd"], "0.00")
        self.assertEqual(by_denom["ustrd"]["destination"], "osmosis-1")

    def test_ladder_counts_live_holders_still_to_sweep(self) -> None:
        # not yet swept and not excluded/skipped: refunded (30), pending (12.05), new (18), deferred (18.10), dust (0.60)
        self.assertEqual([(r["floor"], r["holders"], r["usd"]) for r in self.data["ladder"]],
                         [("10", "4", "78.15"), ("5", "4", "78.15"), ("1", "4", "78.15"), ("0", "5", "78.75")])

    def test_deferred_object_with_a_live_snapshot(self) -> None:
        deferred = self.data["deferred"]
        self.assertEqual((deferred["addresses"], deferred["usd"]), ("1", "18.10"))
        self.assertEqual((deferred["staked_strd"], deferred["staked_usd"]), ("9.00", "0.45"))  # 9 STRD chain-wide at $0.05
        self.assertEqual(deferred["completes_at"], DEFERRED_COMPLETES)
        self.assertEqual(deferred["top"], [{"address": A_DEFERRED, "usd": "24.05", "staked_strd": "7.50", "completes_at": DEFERRED_COMPLETES}])

    def test_progress_parts_sum_to_the_whole(self) -> None:
        totals = self.data["totals"]
        parts = ["swept", "refunded", "remaining", "deferred", "excluded", "below_floor", "unplanned"]
        # every live address lands in exactly one bucket: 60.05 + 30 + 12.05 + 18.10 + 99 + 0.60 + 18 (escrow is skipped, not counted)
        self.assertEqual(sum(Decimal(totals[part]["usd"]) for part in parts), Decimal("237.80"))

    def test_keyless_and_exclusions_sections(self) -> None:
        self.assertEqual(self.data["keyless"], {"addresses": "1", "usd": "12.05", "swept": "0"})
        data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=LIVE,
                                 exclusions={A_EXCLUDED: holders.Exclusion(address=A_EXCLUDED, section="team", label="F5", reason="by hand")},
                                 live_fetched_at=FETCHED_AT)
        self.assertEqual(data["exclusions"], [{"section": "team", "address": A_EXCLUDED, "label": "F5", "reason": "by hand", "live_usd": "99.00"}])


def vesting_plan(run_id: int, addresses: list[str], locked: dict[str, int]) -> planner.Plan:
    holder_list = [holders.Holder(address=address, balances={"stuatom": 1_000_000}, usd=Decimal(6), keyless=False, locked=dict(locked)) for address in addresses]
    hs = holders.HolderSet(height=100, denoms=[STATOM, STRD], holders=holder_list, excluded=[], skipped=[], below_floor=[])
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(1), run_id=run_id, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=10, gas_budget=40_000_000, operator_sequence=5, created_at="2026-11-10T09:00:00+00:00", staked_ustrd=0)


def live_with(balances: dict[str, dict[str, str]]) -> dict:
    return {**LIVE, "balances": balances}


def confirmed_run(run_id: int, address: str, amount: int = 1_000_000) -> list[ledger.Event]:
    batch_id = f"{run_id:03d}-001"
    return [ledger.submitted(run_id=run_id, batch_id=batch_id, tx_hash="V", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.confirmed(batch_id=batch_id, tx_hash="V", at=EARLIER, height=5, gas_used=1, transfers=[t(address, "stuatom", amount)], skipped=[])]


class VestingTests(unittest.TestCase):
    """A vesting holder keeps its locked remainder after the sweep; only a balance above that is a refund."""

    def compose(self, live_amount: str) -> dict:
        plan = vesting_plan(run_id=1, addresses=["stride1vest"], locked={"stuatom": 400_000})
        return sweep_tab.compose(plan=plan, events=confirmed_run(run_id=1, address="stride1vest"),
                                 live=live_with({"stride1vest": {"stuatom": live_amount}}), exclusions={}, live_fetched_at=FETCHED_AT)

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
        data = sweep_tab.compose(plan=plan, events=confirmed_run(run_id=1, address="stride1vest"), live=live, exclusions={}, live_fetched_at=FETCHED_AT)

        self.assertEqual(data["totals"]["swept"], {"addresses": "1", "usd": "6.00"})
        self.assertEqual(data["totals"]["refunded"], {"addresses": "0", "usd": "0.00"})
        self.assertEqual(data["refunded"], [])


class EarlierRunTests(unittest.TestCase):
    def test_addresses_swept_in_an_earlier_run_stay_in_swept_after_a_replan(self) -> None:
        events = confirmed_run(run_id=1, address="stride1runone", amount=3_000_000) + confirmed_run(run_id=2, address="stride1runtwo", amount=2_000_000)
        plan = vesting_plan(run_id=2, addresses=["stride1runtwo", "stride1pending"], locked={})  # run 1's holder is not in this plan
        live = live_with({"stride1pending": {"stuatom": "1000000"}})
        data = sweep_tab.compose(plan=plan, events=events, live=live, exclusions={}, live_fetched_at=FETCHED_AT)

        self.assertEqual(data["totals"]["swept"], {"addresses": "2", "usd": "30.00"})  # 3 + 2 stATOM at $6
        self.assertEqual(data["totals"]["remaining"], {"addresses": "1", "usd": "6.00"})  # only the current plan's unswept address
        stuatom = {d["denom"]: d for d in data["by_denom"]}["stuatom"]
        self.assertEqual(stuatom["swept"], {"addresses": "2", "amount": "5000000", "usd": "30.00"})
        self.assertEqual(stuatom["remaining"]["addresses"], "1")

    def test_an_earlier_runs_refunded_holder_is_listed_and_not_called_unplanned(self) -> None:
        events = confirmed_run(run_id=1, address="stride1runone", amount=3_000_000)
        plan = vesting_plan(run_id=2, addresses=["stride1pending"], locked={})
        live = live_with({"stride1runone": {"stuatom": "3000000"}, "stride1pending": {"stuatom": "1000000"}})
        data = sweep_tab.compose(plan=plan, events=events, live=live, exclusions={}, live_fetched_at=FETCHED_AT)

        self.assertEqual(data["totals"]["refunded"], {"addresses": "1", "usd": "18.00"})
        self.assertEqual(data["totals"]["unplanned"], {"addresses": "0", "usd": "0.00"})
        self.assertEqual([entry["address"] for entry in data["refunded"]], ["stride1runone"])


class DeferredTests(unittest.TestCase):
    def test_latest_completion_wins_and_top_is_ordered_and_capped(self) -> None:
        def deferred_holder(index: int, completes_at: str | None) -> holders.Deferred:
            holder = holders.Holder(address=f"stride1d{index:03d}", balances={"stuatom": (index + 1) * 1_000_000}, usd=Decimal(index + 1) * 6, keyless=False)
            return holders.Deferred(holder=holder, staked=holders.Staked(ustrd=index * 1_000_000, completes_at=completes_at))

        completions = {0: "2026-10-26T08:55:25.469915628Z", 1: "2026-10-26T08:55:25.5Z", 2: None}
        entries = [deferred_holder(index=index, completes_at=completions.get(index, "2026-10-20T00:00:00Z")) for index in range(30)]
        hs = holders.HolderSet(height=100, denoms=[STATOM, STRD], holders=[], excluded=[], skipped=[], below_floor=[], deferred=entries)
        plan = planner.build_plan(holder_set=hs, floor_usd=Decimal(1), run_id=1, canary=0, test=False, gas_per_transfer=100_000,
                                  max_addresses=10, gas_budget=40_000_000, operator_sequence=5, created_at="2026-11-10T09:00:00+00:00", staked_ustrd=0)
        data = sweep_tab.compose(plan=plan, events=[], live=None, exclusions={}, live_fetched_at=None)["deferred"]

        self.assertEqual(data["completes_at"], "2026-10-26T08:55:25.5Z")  # later by the nanosecond fraction, though sorted earlier as text
        self.assertEqual(len(data["top"]), 25)
        self.assertEqual([row["address"] for row in data["top"][:2]], ["stride1d029", "stride1d028"])

    def test_no_deferred_holders(self) -> None:
        plan = vesting_plan(run_id=1, addresses=["stride1a"], locked={})
        data = sweep_tab.compose(plan=plan, events=[], live={**live_with({"stride1a": {"stuatom": "1000000"}}), "staked_ustrd": "0"}, exclusions={}, live_fetched_at=FETCHED_AT)
        self.assertEqual(data["deferred"], {"addresses": "0", "usd": "0.00", "staked_strd": "0.00", "staked_usd": "0.00", "completes_at": None, "top": []})
        self.assertEqual(data["totals"]["deferred"], {"addresses": "0", "usd": "0.00"})


class EdgeTests(unittest.TestCase):
    def test_no_plan(self) -> None:
        data = sweep_tab.compose(plan=None, events=[], live=None, exclusions={}, live_fetched_at=None)
        self.assertIsNone(data["run"])
        self.assertEqual(data["batches"], [])
        self.assertIsNone(data["deferred"])

    def test_no_live_snapshot_keeps_plan_and_ledger_parts(self) -> None:
        data = sweep_tab.compose(plan=make_plan(), events=EVENTS, live=None, exclusions={}, live_fetched_at=None)
        self.assertEqual(data["run"]["batches"]["confirmed"], "2")
        self.assertIsNone(data["totals"])
        self.assertIsNone(data["ladder"])
        self.assertEqual(len(data["batches"]), 3)
        self.assertEqual(data["deferred"], {"addresses": "1", "usd": "24.05", "staked_strd": None, "staked_usd": None,
                                            "completes_at": DEFERRED_COMPLETES,
                                            "top": [{"address": A_DEFERRED, "usd": "24.05", "staked_strd": "7.50", "completes_at": DEFERRED_COMPLETES}]})


class SnapshotRuleTests(unittest.TestCase):
    def compose(self, at: str) -> dict:
        plan = vesting_plan(run_id=2, addresses=["stride1a"], locked={})
        events = confirmed_run(run_id=2, address="stride1a")[:1] + [
            ledger.confirmed(batch_id="002-001", tx_hash="V", at=at, height=5, gas_used=1, transfers=[t("stride1a", "stuatom", 1_000_000)], skipped=[])]
        live = live_with({"stride1a": {"stuatom": "1000000"}})  # the snapshot still shows the balance
        return sweep_tab.compose(plan=plan, events=events, live=live, exclusions={}, live_fetched_at=FETCHED_AT)

    def test_a_transfer_confirmed_after_the_snapshot_is_swept_not_refunded(self) -> None:
        data = self.compose(at=LATER)
        self.assertEqual(data["totals"]["swept"]["addresses"], "1")
        self.assertEqual(data["totals"]["refunded"]["addresses"], "0")
        self.assertEqual(data["refunded"], [])

    def test_a_transfer_confirmed_before_the_snapshot_with_the_balance_present_is_refunded(self) -> None:
        data = self.compose(at=EARLIER)
        self.assertEqual(data["totals"]["refunded"]["addresses"], "1")
        self.assertEqual(data["totals"]["swept"]["addresses"], "0")
        self.assertEqual([r["address"] for r in data["refunded"]], ["stride1a"])

    def test_a_transfer_between_balances_at_and_fetched_at_is_swept_not_refunded(self) -> None:
        # Balances were read at 13:00Z, the snapshot was stamped at 14:00Z; a batch confirmed at 13:30Z is not in the balances
        plan = vesting_plan(run_id=2, addresses=["stride1a"], locked={})
        events = confirmed_run(run_id=2, address="stride1a")[:1] + [
            ledger.confirmed(batch_id="002-001", tx_hash="V", at="2026-11-10T13:30:00+00:00", height=5, gas_used=1,
                             transfers=[t("stride1a", "stuatom", 1_000_000)], skipped=[])]
        live = {**live_with({"stride1a": {"stuatom": "1000000"}}), "balances_at": "2026-11-10T13:00:00+00:00"}

        data = sweep_tab.compose(plan=plan, events=events, live=live, exclusions={}, live_fetched_at=FETCHED_AT)

        self.assertEqual(data["totals"]["swept"]["addresses"], "1")
        self.assertEqual(data["totals"]["refunded"]["addresses"], "0")
        self.assertEqual(data["refunded"], [])

    def test_offsets_are_compared_as_datetimes_not_strings(self) -> None:
        # 15:00+02:00 is 13:00Z, before the 14:00Z snapshot, though it sorts after it as a string
        data = self.compose(at="2026-11-10T15:00:00+02:00")
        self.assertEqual(data["totals"]["refunded"]["addresses"], "1")


class CollectTests(unittest.TestCase):
    def test_balances_at_is_stamped_before_the_balances_are_read(self) -> None:
        order: list[str] = []
        before = datetime.datetime(2026, 11, 10, 13, 0, tzinfo=datetime.UTC)

        def read_balances(denoms: list, workers: int) -> dict:
            order.append("read_balances")
            self.assertEqual(workers, sweep_tab.config.SWEEP_BALANCE_WORKERS)
            return {}

        def now() -> datetime.datetime:
            order.append("now")
            return before

        with mock.patch.object(sweep_tab.holders, "resolve_denoms", return_value=[]), \
                mock.patch.object(sweep_tab.holders, "read_balances", side_effect=read_balances), \
                mock.patch.object(sweep_tab.holders, "read_staking_pool", return_value=0), \
                mock.patch.object(sweep_tab.sweep_chainio, "now", side_effect=now), \
                mock.patch.object(sweep_tab.sweep_chainio, "latest_height", return_value=9), \
                mock.patch.object(sweep_tab.sweep_chainio, "rest_get", return_value={"balance": {"amount": "1"}}):
            live = sweep_tab.collect()

        self.assertEqual(order, ["now", "read_balances"])
        self.assertEqual(live["balances_at"], before.isoformat())


class ProgressTests(unittest.TestCase):
    def test_counts_per_tier_and_overall_with_a_mixed_ledger(self) -> None:
        events = EVENTS[:2] + [ledger.submitted(run_id=2, batch_id="002-002", tx_hash="B", at=EARLIER, addresses=1, transfers_estimate=2, gas_wanted=1)]
        data = sweep_tab.compose(plan=make_plan(), events=events, live=LIVE, exclusions={}, live_fetched_at=FETCHED_AT)
        progress = data["progress"]
        by_tier = {tier["name"]: tier for tier in progress["tiers"]}

        self.assertEqual([tier["name"] for tier in progress["tiers"]], [tier.name.value for tier in make_plan().tiers])
        self.assertEqual(by_tier["main"]["batches"], "2")  # 002-001 confirmed, 002-002 only submitted
        self.assertEqual((by_tier["main"]["batches_confirmed"], by_tier["main"]["addresses_confirmed"], by_tier["main"]["usd_confirmed"]), ("1", "1", "60.05"))
        self.assertEqual((by_tier["keyless"]["batches"], by_tier["keyless"]["batches_confirmed"], by_tier["keyless"]["usd"]), ("1", "0", "12.05"))
        self.assertEqual((progress["run_id"], progress["floor_usd"], progress["test"]), ("2", "10", False))
        self.assertEqual((progress["batches"], progress["batches_confirmed"]), ("3", "1"))
        self.assertEqual((progress["addresses"], progress["addresses_confirmed"]), ("3", "1"))
        self.assertEqual((progress["usd"], progress["usd_confirmed"]), ("102.15", "60.05"))
        self.assertEqual(set(progress), {"run_id", "floor_usd", "test", "tiers", "batches", "batches_confirmed", "addresses",
                                         "addresses_confirmed", "usd", "usd_confirmed"})

    def test_none_without_a_plan(self) -> None:
        data = sweep_tab.compose(plan=None, events=[], live=None, exclusions={}, live_fetched_at=None)
        self.assertIsNone(data["progress"])


if __name__ == "__main__":
    unittest.main()
