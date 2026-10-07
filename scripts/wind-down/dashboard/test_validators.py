import json
import unittest
import urllib.error
from decimal import Decimal
from unittest import mock

import chain
import config
import validators

OPERATOR_A = "valoperA"
OPERATOR_B = "valoperB"
OPERATOR_C = "valoperC"
OPERATOR_D = "valoperD"
STRIDE_ZONE_REST_ZONE = config.ZONES[0]


def stride_validator(address: str, delegation: int, weight: int = 10, **overrides: object) -> validators.StrideValidator:
    fields = {
        "address": address,
        "name": f"name-{address}",
        "delegation": delegation,
        "weight": weight,
        "rate": Decimal("1.05"),
        "slash_query_in_progress": False,
        "delegation_changes_in_progress": 0,
    }
    return validators.StrideValidator(**{**fields, **overrides})


def host_validator(moniker: str, rate: str | None = "1.04", status: str = "bonded", jailed: bool = False) -> validators.HostValidator:
    return validators.HostValidator(
        moniker=moniker, status=status, jailed=jailed, rate=None if rate is None else Decimal(rate)
    )


def rows_by_address(rows: list[validators.ValidatorRow]) -> dict[str, validators.ValidatorRow]:
    return {row.address: row for row in rows}


class SeverityTest(unittest.TestCase):
    def test_under_and_equal_are_neutral(self) -> None:
        self.assertEqual(validators.severity_of(recorded=100, actual=101), validators.Severity.NEUTRAL)
        self.assertEqual(validators.severity_of(recorded=100, actual=100), validators.Severity.NEUTRAL)
        self.assertEqual(validators.severity_of(recorded=0, actual=5), validators.Severity.NEUTRAL)

    def test_over_by_less_than_the_threshold_is_amber(self) -> None:
        # 0.0001% of 10,000,000 is exactly 10; 9 over is just below it.
        self.assertEqual(validators.severity_of(recorded=10_000_000, actual=9_999_991), validators.Severity.AMBER)

    def test_over_by_exactly_the_threshold_is_red(self) -> None:
        self.assertEqual(validators.severity_of(recorded=10_000_000, actual=9_999_990), validators.Severity.RED)

    def test_over_by_more_than_the_threshold_is_red(self) -> None:
        self.assertEqual(validators.severity_of(recorded=10_000_000, actual=0), validators.Severity.RED)

    def test_one_unit_over_a_tiny_recorded_amount_is_red(self) -> None:
        self.assertEqual(validators.severity_of(recorded=5, actual=4), validators.Severity.RED)

    def test_over_within_the_drain_buffer_is_buffered_only_below_rate_one(self) -> None:
        # 614 INJ in wei: the drain shaves recorded // 1e17 = 6,141 wei off a full drain of a slashed validator.
        recorded = 614_119_069_304_957_234_390
        buffer = recorded // validators.DRAIN_BUFFER_DIVISOR
        slashed = Decimal("0.999900000000000001")
        self.assertEqual(
            validators.severity_of(recorded=recorded, actual=recorded - buffer, stride_rate=slashed),
            validators.Severity.BUFFERED,
        )
        self.assertEqual(
            validators.severity_of(recorded=recorded, actual=recorded - buffer - 1, stride_rate=slashed),
            validators.Severity.AMBER,
        )
        # No buffer exists at a rate of exactly 1 (or without a rate), so the same overage is a real overage.
        self.assertEqual(
            validators.severity_of(recorded=recorded, actual=recorded - 2, stride_rate=Decimal(1)), validators.Severity.AMBER
        )
        self.assertEqual(validators.severity_of(recorded=recorded, actual=recorded - 2), validators.Severity.AMBER)

    def test_buffer_is_at_least_one_unit(self) -> None:
        self.assertEqual(
            validators.severity_of(recorded=5, actual=4, stride_rate=Decimal("0.9")), validators.Severity.BUFFERED
        )
        self.assertEqual(validators.severity_of(recorded=5, actual=3, stride_rate=Decimal("0.9")), validators.Severity.RED)

    def test_boundary_is_exact_at_eighteen_decimals(self) -> None:
        recorded = 10**24 + 7
        threshold = -(-recorded // validators.RED_FRACTION_DENOMINATOR)  # ceil(recorded / 1e6)
        self.assertEqual(validators.severity_of(recorded=recorded, actual=recorded - threshold), validators.Severity.RED)
        self.assertEqual(
            validators.severity_of(recorded=recorded, actual=recorded - threshold + 1), validators.Severity.AMBER
        )


class BuildRowsTest(unittest.TestCase):
    def build(self, **overrides: object) -> list[validators.ValidatorRow]:
        arguments = {
            "stride_validators": [
                stride_validator(address=OPERATOR_A, delegation=1_000_000, weight=30),
                stride_validator(address=OPERATOR_B, delegation=500_000, weight=10),
            ],
            "delegations": {OPERATOR_A: 900_000, OPERATOR_B: 600_000},
            "host_validators": {OPERATOR_A: host_validator(moniker="Alpha", rate="1.04"), OPERATOR_B: host_validator(moniker="Beta")},
            "unbonding_entries": {OPERATOR_A: 3},
        }
        return validators.build_rows(**{**arguments, **overrides})

    def test_row_fields_join_stride_and_host(self) -> None:
        row = rows_by_address(self.build())[OPERATOR_A]

        self.assertEqual(row.moniker, "Alpha")
        self.assertEqual((row.recorded, row.actual, row.diff), (1_000_000, 900_000, 100_000))
        self.assertEqual(row.diff_percent, Decimal("10"))
        self.assertEqual(row.weight_percent, Decimal("75.00"))
        self.assertEqual(row.stride_rate, Decimal("1.05"))
        self.assertEqual(row.chain_rate, Decimal("1.04"))
        self.assertEqual(row.rate_difference, Decimal("0.01"))
        self.assertEqual(row.unbonding_entries, 3)
        self.assertEqual((row.bond_status, row.jailed, row.registered), ("bonded", False, True))
        self.assertEqual(row.severity, validators.Severity.RED)

    def test_under_delegated_row_has_negative_diff_and_is_neutral(self) -> None:
        row = rows_by_address(self.build())[OPERATOR_B]

        self.assertEqual((row.diff, row.severity), (-100_000, validators.Severity.NEUTRAL))
        self.assertEqual(row.diff_percent, Decimal("-20"))
        self.assertEqual(row.unbonding_entries, 0)  # lookup succeeded and omitted the validator

    def test_validator_without_a_host_delegation_has_zero_actual(self) -> None:
        rows = self.build(delegations={OPERATOR_A: 900_000})

        row = rows_by_address(rows)[OPERATOR_B]
        self.assertEqual((row.actual, row.diff, row.severity), (0, 500_000, validators.Severity.RED))

    def test_unregistered_host_delegation_is_a_row_with_zero_recorded(self) -> None:
        rows = self.build(
            delegations={OPERATOR_A: 900_000, OPERATOR_B: 600_000, OPERATOR_C: 250},
            host_validators={OPERATOR_C: host_validator(moniker="Gamma", status="unbonding", jailed=True)},
        )

        row = rows_by_address(rows)[OPERATOR_C]
        self.assertFalse(row.registered)
        self.assertEqual((row.moniker, row.recorded, row.actual, row.diff), ("Gamma", 0, 250, -250))
        self.assertIsNone(row.diff_percent)
        self.assertIsNone(row.stride_rate)
        self.assertIsNone(row.rate_difference)
        self.assertEqual((row.weight_percent, row.bond_status, row.jailed), (Decimal(0), "unbonding", True))
        self.assertEqual(row.severity, validators.Severity.NEUTRAL)
        self.assertEqual((row.delegation_changes_in_progress, row.slash_query_in_progress), (None, None))
        self.assertFalse(row.in_progress)

    def test_zero_balance_host_delegation_is_not_an_unregistered_row(self) -> None:
        rows = self.build(delegations={OPERATOR_A: 900_000, OPERATOR_C: 0})

        self.assertNotIn(OPERATOR_C, rows_by_address(rows))

    def test_rows_sort_by_absolute_diff_descending(self) -> None:
        rows = self.build(
            stride_validators=[
                stride_validator(address=OPERATOR_A, delegation=1_000),
                stride_validator(address=OPERATOR_B, delegation=1_000),
            ],
            delegations={OPERATOR_A: 400, OPERATOR_B: 3_000, OPERATOR_C: 1_500},
            host_validators=None,
        )

        # diffs: A +600, B -2000, unregistered C -1500
        self.assertEqual([row.address for row in rows], [OPERATOR_B, OPERATOR_C, OPERATOR_A])

    def test_failed_optional_lookups_yield_nulls_and_fall_back_to_stride_name(self) -> None:
        row = rows_by_address(self.build(host_validators=None, unbonding_entries=None))[OPERATOR_A]

        self.assertEqual(row.moniker, f"name-{OPERATOR_A}")
        self.assertIsNone(row.unbonding_entries)
        self.assertIsNone(row.bond_status)
        self.assertIsNone(row.jailed)
        self.assertIsNone(row.chain_rate)
        self.assertIsNone(row.rate_difference)

    def test_in_progress_flags(self) -> None:
        rows = self.build(
            stride_validators=[
                stride_validator(address=OPERATOR_A, delegation=10, slash_query_in_progress=True),
                stride_validator(address=OPERATOR_B, delegation=10, delegation_changes_in_progress=2),
                stride_validator(address=OPERATOR_C, delegation=10),
            ],
        )
        by_address = rows_by_address(rows)

        self.assertEqual(
            (by_address[OPERATOR_A].slash_query_in_progress, by_address[OPERATOR_A].delegation_changes_in_progress),
            (True, 0),
        )
        self.assertEqual(
            (by_address[OPERATOR_B].slash_query_in_progress, by_address[OPERATOR_B].delegation_changes_in_progress),
            (False, 2),
        )
        self.assertTrue(by_address[OPERATOR_A].in_progress and by_address[OPERATOR_B].in_progress)
        self.assertFalse(by_address[OPERATOR_C].in_progress)

    def test_eighteen_decimal_amounts_stay_exact(self) -> None:
        recorded = 107_130_613_496_529_123_456_789
        actual = 107_129_220_024_855_000_000_001
        rows = self.build(
            stride_validators=[stride_validator(address=OPERATOR_A, delegation=recorded)],
            delegations={OPERATOR_A: actual},
        )

        self.assertEqual(rows[0].diff, recorded - actual)
        payload = rows[0].payload()
        self.assertEqual(payload["recorded"], str(recorded))
        self.assertEqual(payload["diff"], "1393471674123456788")
        self.assertEqual(payload["weight_percent"], "100.00")


WHOLE = 1_000_000  # one whole token at 6 decimals


class LiveTestPickTest(unittest.TestCase):
    def pick(
        self,
        stride_validators: list[validators.StrideValidator],
        unbonding_entries: dict[str, int] | None = None,
        delegations: dict[str, int] | None = None,
        decimals: int = 6,
    ) -> validators.LiveTestPicks:
        rows = validators.build_rows(
            stride_validators=stride_validators,
            delegations=delegations or {},
            host_validators=None,
            unbonding_entries=unbonding_entries,
        )
        return validators.pick_live_test(rows=rows, decimals=decimals)

    def test_smallest_recorded_wins_and_the_second_smallest_is_next(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_A, delegation=50 * WHOLE),
                stride_validator(address=OPERATOR_B, delegation=5 * WHOLE),
                stride_validator(address=OPERATOR_C, delegation=9 * WHOLE),
            ],
            unbonding_entries={},
        )

        self.assertEqual(picks.pick, validators.LiveTestCandidate(address=OPERATOR_B, moniker=f"name-{OPERATOR_B}", recorded=5 * WHOLE))
        self.assertEqual(picks.runner_up.address, OPERATOR_C)
        self.assertIsNone(picks.reason)

    def test_below_one_whole_token_is_skipped_and_exactly_one_qualifies(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_A, delegation=WHOLE - 1),
                stride_validator(address=OPERATOR_B, delegation=WHOLE),
            ],
            unbonding_entries={},
        )

        self.assertEqual(picks.pick.address, OPERATOR_B)
        self.assertIsNone(picks.runner_up)

    def test_the_floor_scales_with_the_zones_decimals(self) -> None:
        validator_set = [
            stride_validator(address=OPERATOR_A, delegation=10**18 - 1),
            stride_validator(address=OPERATOR_B, delegation=10**18),
        ]

        self.assertEqual(self.pick(validator_set, unbonding_entries={}, decimals=18).pick.address, OPERATOR_B)
        self.assertEqual(self.pick(validator_set, unbonding_entries={}, decimals=6).pick.address, OPERATOR_A)

    def test_a_validator_with_an_unbonding_entry_in_flight_is_skipped(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_A, delegation=2 * WHOLE),
                stride_validator(address=OPERATOR_B, delegation=3 * WHOLE),
                stride_validator(address=OPERATOR_C, delegation=4 * WHOLE),
            ],
            unbonding_entries={OPERATOR_A: 1},
        )

        self.assertEqual((picks.pick.address, picks.runner_up.address), (OPERATOR_B, OPERATOR_C))

    def test_changes_or_slash_query_in_progress_are_skipped(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_A, delegation=2 * WHOLE, delegation_changes_in_progress=1),
                stride_validator(address=OPERATOR_B, delegation=3 * WHOLE, slash_query_in_progress=True),
                stride_validator(address=OPERATOR_C, delegation=4 * WHOLE),
            ],
            unbonding_entries={},
        )

        self.assertEqual(picks.pick.address, OPERATOR_C)
        self.assertIsNone(picks.runner_up)

    def test_an_unregistered_host_delegation_is_never_picked(self) -> None:
        picks = self.pick(
            [stride_validator(address=OPERATOR_A, delegation=9 * WHOLE)],
            unbonding_entries={},
            delegations={OPERATOR_D: 2 * WHOLE},
        )

        self.assertEqual(picks.pick.address, OPERATOR_A)
        self.assertIsNone(picks.runner_up)

    def test_ties_break_on_address_so_the_pick_is_stable(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_B, delegation=2 * WHOLE),
                stride_validator(address=OPERATOR_A, delegation=2 * WHOLE),
            ],
            unbonding_entries={},
        )

        self.assertEqual((picks.pick.address, picks.runner_up.address), (OPERATOR_A, OPERATOR_B))

    def test_no_pick_with_a_reason_when_the_unbonding_lookup_failed(self) -> None:
        picks = self.pick([stride_validator(address=OPERATOR_A, delegation=2 * WHOLE)], unbonding_entries=None)

        self.assertEqual((picks.pick, picks.runner_up), (None, None))
        self.assertEqual(picks.reason, "unbonding entries could not be read")

    def test_falls_back_to_the_smallest_validator_with_room_under_the_entry_cap(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_A, delegation=2 * WHOLE),
                stride_validator(address=OPERATOR_B, delegation=3 * WHOLE),
                stride_validator(address=OPERATOR_C, delegation=1 * WHOLE),
            ],
            unbonding_entries={OPERATOR_A: 2, OPERATOR_B: 1, OPERATOR_C: validators.MAX_UNBONDING_ENTRIES},
        )

        # C is smallest but at the cap, so A (2 entries) is the pick and B the runner-up, with the reason attached.
        self.assertEqual((picks.pick.address, picks.runner_up.address), (OPERATOR_A, OPERATOR_B))
        self.assertEqual(picks.reason, "no entry-free validator; the pick has 2 of 7 entries in flight")

    def test_no_pick_with_a_reason_when_every_validator_is_at_the_cap_or_in_progress(self) -> None:
        picks = self.pick(
            [
                stride_validator(address=OPERATOR_A, delegation=2 * WHOLE),
                stride_validator(address=OPERATOR_B, delegation=2 * WHOLE, delegation_changes_in_progress=1),
            ],
            unbonding_entries={OPERATOR_A: validators.MAX_UNBONDING_ENTRIES, OPERATOR_B: 0},
        )

        self.assertEqual((picks.pick, picks.runner_up), (None, None))
        self.assertEqual(picks.reason, "all 2 funded validators are at the entry cap or have a change in progress")

    def test_no_pick_with_a_reason_when_no_validator_holds_a_whole_token(self) -> None:
        picks = self.pick([stride_validator(address=OPERATOR_A, delegation=WHOLE - 1)], unbonding_entries={})

        self.assertEqual((picks.pick, picks.runner_up), (None, None))
        self.assertEqual(picks.reason, "no registered validator holds a whole token")

    def test_collect_exposes_the_pick_as_json(self) -> None:
        zone = config.ZONES[0]
        stride_zone = {
            "chain_id": zone.chain_id,
            "delegation_ica_address": "ica",
            "validators": [
                {"address": address, "name": f"name-{address}", "delegation": str(delegation), "weight": "5",
                 "shares_to_tokens_rate": "1.0", "slash_query_in_progress": False, "delegation_changes_in_progress": "0"}
                for address, delegation in ((OPERATOR_A, 3 * WHOLE), (OPERATOR_B, 2 * WHOLE))
            ],
        }

        def fake_get_all_pages(chain: object, path: str, key: str, params: object = None, max_items: object = None) -> list[object]:
            if key == "delegation_responses":
                return []
            if key == "unbonding_responses":
                return []
            raise json.JSONDecodeError("bad", "", 0)

        with mock.patch.object(validators.chain, "rest_get_all_pages", fake_get_all_pages), mock.patch.object(
            validators.chain, "rest_get", side_effect=urllib.error.URLError("params down")
        ):
            entry = validators._collect_zone(zone=zone, stride_zone=stride_zone)

        json.dumps(entry)
        self.assertEqual(entry["live_test_pick"], {"address": OPERATOR_B, "moniker": f"name-{OPERATOR_B}", "recorded": str(2 * WHOLE)})
        self.assertEqual(entry["live_test_next"]["address"], OPERATOR_A)
        self.assertIsNone(entry["live_test_reason"])


UNBONDING_SECONDS = 1_814_400.0  # 21 days
# UPGRADE_TIME 2026-10-12T12:00:00Z + 21 days
CUTOFF = "2026-11-02T12:00:00Z"


def unbonding(count: int, latest_completion: str) -> validators.UnbondingSummary:
    return validators.UnbondingSummary(count=count, latest_completion=latest_completion)


class DrainedCountTest(unittest.TestCase):
    def count(
        self,
        recorded: dict[str, int],
        actual: dict[str, int],
        entries: dict[str, validators.UnbondingSummary] | None,
        unbonding_seconds: float | None = UNBONDING_SECONDS,
    ) -> int | None:
        rows = validators.build_rows(
            stride_validators=[stride_validator(address=address, delegation=amount) for address, amount in recorded.items()],
            delegations=actual,
            host_validators=None,
            unbonding_entries=None,
        )
        return validators.drained_count(rows=rows, unbonding=entries, unbonding_seconds=unbonding_seconds)

    def test_counts_only_validators_at_zero_zero_with_a_post_upgrade_entry(self) -> None:
        count = self.count(
            recorded={OPERATOR_A: 0, OPERATOR_B: 0, OPERATOR_C: 0, OPERATOR_D: 5},
            actual={OPERATOR_C: 7},
            entries={
                OPERATOR_A: unbonding(count=1, latest_completion="2026-11-02T12:00:01Z"),  # one second after the cutoff
                OPERATOR_B: unbonding(count=2, latest_completion="2026-11-02T11:59:59Z"),  # pre-upgrade entries only
                OPERATOR_C: unbonding(count=1, latest_completion="2026-11-05T00:00:00Z"),  # still delegated on the host
                OPERATOR_D: unbonding(count=1, latest_completion="2026-11-05T00:00:00Z"),  # still recorded by Stride
            },
        )

        self.assertEqual(count, 1)

    def test_an_entry_completing_exactly_at_the_cutoff_predates_the_upgrade(self) -> None:
        count = self.count(
            recorded={OPERATOR_A: 0}, actual={}, entries={OPERATOR_A: unbonding(count=1, latest_completion=CUTOFF)}
        )

        self.assertEqual(count, 0)

    def test_the_cutoff_moves_with_the_hosts_unbonding_time(self) -> None:
        entries = {OPERATOR_A: unbonding(count=1, latest_completion="2026-10-20T00:00:00Z")}

        self.assertEqual(self.count(recorded={OPERATOR_A: 0}, actual={}, entries=entries, unbonding_seconds=7 * 86400), 1)
        self.assertEqual(self.count(recorded={OPERATOR_A: 0}, actual={}, entries=entries), 0)

    def test_a_validator_without_entries_is_not_drained(self) -> None:
        self.assertEqual(self.count(recorded={OPERATOR_A: 0}, actual={}, entries={}), 0)

    def test_nullable_when_the_entries_or_the_unbonding_time_lookup_failed(self) -> None:
        entries = {OPERATOR_A: unbonding(count=1, latest_completion="2027-01-01T00:00:00Z")}

        self.assertIsNone(self.count(recorded={OPERATOR_A: 0}, actual={}, entries=None))
        self.assertIsNone(self.count(recorded={OPERATOR_A: 0}, actual={}, entries=entries, unbonding_seconds=None))

    def test_collect_zone_reports_the_count_and_keeps_unbonding_entries_as_an_int(self) -> None:
        zone = config.ZONES[0]
        stride_zone = {
            "chain_id": zone.chain_id,
            "delegation_ica_address": "ica",
            "validators": [
                {"address": address, "name": address, "delegation": "0", "weight": "5", "shares_to_tokens_rate": "1.0",
                 "slash_query_in_progress": False, "delegation_changes_in_progress": "0"}
                for address in (OPERATOR_A, OPERATOR_B)
            ],
        }

        def fake_get_all_pages(chain: object, path: str, key: str, params: object = None, max_items: object = None) -> list[object]:
            if key == "unbonding_responses":
                return [
                    {"validator_address": OPERATOR_A, "entries": [
                        {"completion_time": "2026-10-30T00:00:00.123456789Z"}, {"completion_time": "2026-11-03T09:00:00.5Z"}]},
                    {"validator_address": OPERATOR_B, "entries": [{"completion_time": "2026-10-30T00:00:00Z"}]},
                ]
            if key == "delegation_responses":
                return []
            raise json.JSONDecodeError("bad", "", 0)

        def fake_rest_get(chain: object, path: str, params: object = None) -> dict[str, object]:
            assert path == validators.STAKING_PARAMS_PATH
            return {"params": {"unbonding_time": "1814400s"}}

        with mock.patch.object(validators.chain, "rest_get_all_pages", fake_get_all_pages), mock.patch.object(
            validators.chain, "rest_get", fake_rest_get
        ):
            entry = validators._collect_zone(zone=zone, stride_zone=stride_zone)

        self.assertEqual(entry["drained_count"], 1)
        self.assertEqual({row["address"]: row["unbonding_entries"] for row in entry["validators"]}, {OPERATOR_A: 2, OPERATOR_B: 1})

    def test_collect_zone_has_a_null_count_when_the_params_lookup_fails(self) -> None:
        zone = config.ZONES[0]
        stride_zone = {"chain_id": zone.chain_id, "delegation_ica_address": "ica", "validators": []}

        with mock.patch.object(validators.chain, "rest_get_all_pages", return_value=[]), mock.patch.object(
            validators.chain, "rest_get", side_effect=urllib.error.URLError("params down")
        ):
            entry = validators._collect_zone(zone=zone, stride_zone=stride_zone)

        self.assertIsNone(entry["drained_count"])


class SummarizeTest(unittest.TestCase):
    def test_totals_and_counts(self) -> None:
        rows = validators.build_rows(
            stride_validators=[
                stride_validator(address=OPERATOR_A, delegation=10_000_000, delegation_changes_in_progress=1),
                stride_validator(address=OPERATOR_B, delegation=1_000),
                stride_validator(address=OPERATOR_D, delegation=2_000, rate=Decimal("0.99")),
            ],
            delegations={OPERATOR_A: 9_999_991, OPERATOR_C: 40, OPERATOR_D: 1_999},
            host_validators=None,
            unbonding_entries=None,
        )

        totals = validators.summarize(rows=rows)

        # A is over by 9 (amber), B is over by 1000 with nothing on the host (red), C is unregistered,
        # D is over by 1 on a slashed validator (buffered, not counted as over).
        self.assertEqual(totals.validator_count, 4)
        self.assertEqual(totals.delegated_count, 3)
        self.assertEqual((totals.recorded, totals.actual, totals.diff), (10_003_000, 10_002_030, 970))
        self.assertEqual((totals.over_count, totals.red_count, totals.in_progress_count), (2, 1, 1))
        self.assertEqual(totals.payload()["diff"], "970")


class CollectTest(unittest.TestCase):
    def test_zone_error_is_isolated_to_that_zone(self) -> None:
        stride_zone = {"chain_id": "x", "delegation_ica_address": "ica", "validators": []}

        def fake_rest_get(chain: object, path: str, params: object = None) -> dict[str, object]:
            raise urllib.error.URLError("boom")

        with mock.patch.object(validators.chain, "rest_get", fake_rest_get):
            entry = validators._collect_zone_guarded(zone=config.ZONES[0], stride_zones={config.ZONES[0].chain_id: stride_zone})

        self.assertEqual(entry["chain_id"], config.ZONES[0].chain_id)
        self.assertIn("boom", entry["error"])

    def test_zone_missing_from_stride_is_an_error_entry(self) -> None:
        entry = validators._collect_zone_guarded(zone=config.ZONES[0], stride_zones={})

        self.assertIn("error", entry)

    def test_zone_payload_is_json_serialisable_with_string_amounts(self) -> None:
        zone = next(zone for zone in config.ZONES if zone.decimals == 18)
        stride_zone = {
            "chain_id": zone.chain_id,
            "delegation_ica_address": "ica",
            "validators": [
                {
                    "address": OPERATOR_A,
                    "name": "a",
                    "delegation": "2000000000000000000",
                    "weight": "5",
                    "shares_to_tokens_rate": "1.000000000000000000",
                    "slash_query_in_progress": False,
                    "delegation_changes_in_progress": "0",
                }
            ],
        }

        def fake_get_all_pages(chain: object, path: str, key: str, params: object = None, max_items: object = None) -> list[object]:
            if key == "delegation_responses":
                return [{"delegation": {"validator_address": OPERATOR_A}, "balance": {"amount": "1999999999999999999"}}]
            if key == "validators":
                return [{"operator_address": OPERATOR_A, "description": {"moniker": "Alpha"}, "status": "BOND_STATUS_BONDED",
                         "jailed": False, "tokens": "3", "delegator_shares": "4.000000000000000000"}]
            raise json.JSONDecodeError("bad", "", 0)  # unbonding lookup fails

        with mock.patch.object(validators.chain, "rest_get_all_pages", fake_get_all_pages), mock.patch.object(
            validators.chain, "rest_get", side_effect=urllib.error.URLError("params down")
        ):
            entry = validators._collect_zone(zone=zone, stride_zone=stride_zone)

        json.dumps(entry)
        self.assertEqual(entry["decimals"], 18)
        self.assertEqual(entry["totals"]["recorded"], "2000000000000000000")
        self.assertEqual(entry["totals"]["diff"], "1")
        row = entry["validators"][0]
        self.assertEqual((row["moniker"], row["diff"], row["severity"]), ("Alpha", "1", "amber"))
        self.assertEqual((row["chain_rate"], row["rate_difference"]), ("0.750000000000000000", "0.250000000000000000"))
        self.assertIsNone(row["unbonding_entries"])
        self.assertIsNone(entry["drained_count"])


class StaketiaTest(unittest.TestCase):
    def test_multisig_total_against_remaining_delegated_balance(self) -> None:
        def fake_rest_get(chain: object, path: str, params: object = None) -> dict[str, object]:
            return {"host_zone": {"delegation_address": "celestia1multisig", "remaining_delegated_balance": "1000"}}

        def fake_get_all_pages(chain: object, path: str, key: str, params: object = None, max_items: object = None) -> list[object]:
            if key == "delegation_responses":
                return [
                    {"delegation": {"validator_address": OPERATOR_A}, "balance": {"amount": "300"}},
                    {"delegation": {"validator_address": OPERATOR_B}, "balance": {"amount": "600"}},
                ]
            raise urllib.error.URLError("validators down")

        with mock.patch.object(validators.chain, "rest_get", fake_rest_get), mock.patch.object(
            validators.chain, "rest_get_all_pages", fake_get_all_pages
        ):
            entry = validators._collect_staketia_guarded()

        self.assertEqual(entry["chain_id"], "staketia")
        self.assertEqual(entry["delegation_address"], "celestia1multisig")
        self.assertEqual([row["actual"] for row in entry["validators"]], ["600", "300"])
        self.assertEqual((entry["actual_total"], entry["remaining_delegated_balance"], entry["diff"]), ("900", "1000", "100"))
        self.assertEqual(entry["severity"], "red")
        self.assertIsNone(entry["validators"][0]["bond_status"])
        json.dumps(entry)

    def test_staketia_failure_is_an_error_entry(self) -> None:
        with mock.patch.object(validators.chain, "rest_get", side_effect=TimeoutError("slow")):
            entry = validators._collect_staketia_guarded()

        self.assertEqual(entry["chain_id"], "staketia")
        self.assertIn("slow", entry["error"])


class ParseHostValidatorTest(unittest.TestCase):
    def test_rate_is_tokens_over_delegator_shares(self) -> None:
        parsed = validators._parse_host_validator(
            raw={"description": {"moniker": "m"}, "status": "BOND_STATUS_UNBONDED", "jailed": True,
                 "tokens": "1000", "delegator_shares": "1250.000000000000000000"}
        )

        self.assertEqual((parsed.status, parsed.jailed, parsed.rate), ("unbonded", True, Decimal("0.8")))

    def test_no_shares_means_no_rate(self) -> None:
        parsed = validators._parse_host_validator(
            raw={"description": {"moniker": "m"}, "status": "BOND_STATUS_BONDED", "tokens": "0", "delegator_shares": "0.000000000000000000"}
        )

        self.assertIsNone(parsed.rate)


if __name__ == "__main__":
    unittest.main()
