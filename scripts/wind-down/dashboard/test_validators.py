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

        self.assertEqual(by_address[OPERATOR_A].in_progress_detail, "slash query in progress")
        self.assertEqual(by_address[OPERATOR_B].in_progress_detail, "2 delegation change(s) in progress")
        self.assertTrue(by_address[OPERATOR_A].in_progress and by_address[OPERATOR_B].in_progress)
        self.assertFalse(by_address[OPERATOR_C].in_progress)
        self.assertIsNone(by_address[OPERATOR_C].in_progress_detail)

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


class SummarizeTest(unittest.TestCase):
    def test_totals_and_counts(self) -> None:
        rows = validators.build_rows(
            stride_validators=[
                stride_validator(address=OPERATOR_A, delegation=10_000_000, delegation_changes_in_progress=1),
                stride_validator(address=OPERATOR_B, delegation=1_000),
            ],
            delegations={OPERATOR_A: 9_999_991, OPERATOR_C: 40},
            host_validators=None,
            unbonding_entries=None,
        )

        totals = validators.summarize(rows=rows)

        # A is over by 9 (amber), B is over by 1000 with nothing on the host (red), C is unregistered.
        self.assertEqual(totals.validator_count, 3)
        self.assertEqual(totals.delegated_count, 2)
        self.assertEqual((totals.recorded, totals.actual, totals.diff), (10_001_000, 10_000_031, 969))
        self.assertEqual((totals.over_count, totals.red_count, totals.in_progress_count), (2, 1, 1))
        self.assertEqual(totals.payload()["diff"], "969")


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

        with mock.patch.object(validators.chain, "rest_get_all_pages", fake_get_all_pages):
            entry = validators._collect_zone(zone=zone, stride_zone=stride_zone)

        json.dumps(entry)
        self.assertEqual(entry["decimals"], 18)
        self.assertEqual(entry["totals"]["recorded"], "2000000000000000000")
        self.assertEqual(entry["totals"]["diff"], "1")
        row = entry["validators"][0]
        self.assertEqual((row["moniker"], row["diff"], row["severity"]), ("Alpha", "1", "amber"))
        self.assertEqual((row["chain_rate"], row["rate_difference"]), ("0.750000000000000000", "0.250000000000000000"))
        self.assertIsNone(row["unbonding_entries"])


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
