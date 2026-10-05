# scripts/wind-down/test_gen_delta_table.py
import contextlib
import io
import json
import pathlib
import sys
import tempfile
import unittest
from unittest import mock

import gen_delta_table

CHAIN_ID = "haqq_11235-1"


def write_drift(rows: list[dict]) -> pathlib.Path:
    payload = {"generated_at": "2026-09-29T00:00:00Z", "zones": {CHAIN_ID: {"validators": rows}}}
    path = pathlib.Path(tempfile.mkdtemp()) / "drift.json"
    path.write_text(json.dumps(payload))
    return path


def write_host_zone(delegations: dict[str, int]) -> pathlib.Path:
    validators = [
        {"address": address, "name": f"{address}-stride-name", "delegation": str(delegation)}
        for address, delegation in delegations.items()
    ]
    path = pathlib.Path(tempfile.mkdtemp()) / "host_zone.json"
    path.write_text(json.dumps({"host_zone": {"validators": validators}}))
    return path


def entry(address: str, delta: int, recorded: int) -> gen_delta_table.DeltaEntry:
    return gen_delta_table.DeltaEntry(
        name=address.removeprefix("haqqvaloper1"), address=address, delta=delta, recorded=recorded,
    )


def run_main(args: list[str], stdout: io.StringIO) -> None:
    """Run the CLI, printing into stdout; a SystemExit propagates to the caller (also for a nonzero return)."""
    with mock.patch.object(sys, "argv", ["gen_delta_table.py", *args]), contextlib.redirect_stdout(stdout):
        status = gen_delta_table.main()
    if status:
        raise SystemExit(status)


class GenDeltaTableTest(unittest.TestCase):
    def test_emits_one_entry_per_nonzero_delta_sorted_by_magnitude(self) -> None:
        drift_path = write_drift(
            [
                {
                    "validator_address": "haqqvaloper1small", "moniker": "Small Node", "diff": -5,
                    "in_stride_list": True, "recorded": 7_000_000,
                },
                {
                    "validator_address": "haqqvaloper1big", "moniker": "Big [Node] ⚡", "diff": 853,
                    "in_stride_list": True, "recorded": 8_550_624_701_423_372_833_667_126,
                },
                {"validator_address": "haqqvaloper1zero", "moniker": "Zero", "diff": 0, "in_stride_list": True},
                {"validator_address": "haqqvaloper1foreign", "moniker": "Foreign", "diff": 7, "in_stride_list": False},
            ]
        )

        entries = gen_delta_table.load_deltas(drift_path=drift_path, chain_id=CHAIN_ID, names={})

        self.assertEqual(
            [
                ("bignode", "haqqvaloper1big", -853, 8_550_624_701_423_372_833_667_126),
                ("smallnode", "haqqvaloper1small", 5, 7_000_000),
            ],
            [(row.name, row.address, row.delta, row.recorded) for row in entries],
        )

    def test_names_come_from_the_host_zone_when_given(self) -> None:
        drift_path = write_drift(
            [{"validator_address": "haqqvaloper1big", "moniker": "Big", "diff": 1, "in_stride_list": True, "recorded": 900}]
        )

        entries = gen_delta_table.load_deltas(drift_path=drift_path, chain_id=CHAIN_ID, names={"haqqvaloper1big": "bigstride"})

        self.assertEqual("bigstride", entries[0].name)

    def test_render_go_literal(self) -> None:
        rendered = gen_delta_table.render_go(
            entries=[gen_delta_table.DeltaEntry(name="big", address="haqqvaloper1big", delta=-853, recorded=900)],
            var_name="HaqqDelegationDeltas",
        )

        self.assertEqual(
            'var HaqqDelegationDeltas = []DelegationDelta{\n'
            '\t{Name: "big", Address: "haqqvaloper1big", Delta: mustInt("-853")},\n'
            '}\n',
            rendered,
        )

    def test_tracked_delegations_are_read_from_the_host_zone_file(self) -> None:
        host_zone_path = write_host_zone({"haqqvaloper1big": 8550624701423372833667126, "haqqvaloper1zero": 0})

        tracked = gen_delta_table.load_tracked_delegations(host_zone_path=host_zone_path)

        self.assertEqual({"haqqvaloper1big": 8550624701423372833667126, "haqqvaloper1zero": 0}, tracked)

    def test_render_go_tracked_map_pins_each_row_to_its_recorded_value_in_table_order(self) -> None:
        entries = [
            entry(address="haqqvaloper1big", delta=-853, recorded=900),
            entry(address="haqqvaloper1small", delta=5, recorded=7),
        ]

        rendered = gen_delta_table.render_go_tracked(entries=entries, var_name="HaqqExpectedTrackedDelegations")

        self.assertEqual(
            'var HaqqExpectedTrackedDelegations = map[string]sdkmath.Int{\n'
            '\t"haqqvaloper1big": mustInt("900"),\n'
            '\t"haqqvaloper1small": mustInt("7"),\n'
            '}\n',
            rendered,
        )

    def test_cross_check_passes_when_the_host_zone_matches_every_row(self) -> None:
        entries = [
            entry(address="haqqvaloper1big", delta=-853, recorded=900),
            entry(address="haqqvaloper1small", delta=5, recorded=7),
        ]
        # An unrelated validator may differ: only emitted rows are checked
        tracked = {"haqqvaloper1big": 900, "haqqvaloper1small": 7, "haqqvaloper1other": 123}

        self.assertEqual([], gen_delta_table.find_tracked_mismatches(entries=entries, tracked=tracked))
        gen_delta_table.require_matching_host_zone(entries=entries, tracked=tracked)

    def test_cross_check_fails_naming_the_validator_and_both_values(self) -> None:
        entries = [
            entry(address="haqqvaloper1big", delta=-853, recorded=900),
            entry(address="haqqvaloper1small", delta=5, recorded=7),
        ]
        # A slash was booked on `small` after the drift run: 7 became 6
        tracked = {"haqqvaloper1big": 900, "haqqvaloper1small": 6}

        with self.assertRaises(SystemExit) as raised:
            gen_delta_table.require_matching_host_zone(entries=entries, tracked=tracked)

        message = raised.exception.code
        self.assertIsInstance(message, str)  # a message, so a non-zero exit status
        self.assertIn("haqqvaloper1small: drift recorded 7, host zone tracked 6", message)
        self.assertNotIn("haqqvaloper1big", message)
        self.assertIn("measure_delegation_drift.py", message)

    def test_cross_check_fails_when_the_validator_is_missing_from_the_host_zone(self) -> None:
        entries = [entry(address="haqqvaloper1big", delta=-853, recorded=900)]

        with self.assertRaises(SystemExit) as raised:
            gen_delta_table.require_matching_host_zone(entries=entries, tracked={"haqqvaloper1other": 900})

        self.assertIn(
            "haqqvaloper1big: drift recorded 900, host zone tracked missing from the host zone file", raised.exception.code,
        )

    def test_cross_check_lists_every_mismatching_validator(self) -> None:
        entries = [
            entry(address="haqqvaloper1a", delta=1, recorded=10),
            entry(address="haqqvaloper1b", delta=1, recorded=20),
        ]

        with self.assertRaises(SystemExit) as raised:
            gen_delta_table.require_matching_host_zone(
                entries=entries, tracked={"haqqvaloper1a": 11, "haqqvaloper1b": 21},
            )

        self.assertIn("haqqvaloper1a: drift recorded 10, host zone tracked 11", raised.exception.code)
        self.assertIn("haqqvaloper1b: drift recorded 20, host zone tracked 21", raised.exception.code)


class MainTest(unittest.TestCase):
    def drift_path(self) -> pathlib.Path:
        return write_drift(
            [
                {
                    "validator_address": "haqqvaloper1big", "moniker": "Big", "diff": 853,
                    "in_stride_list": True, "recorded": 900,
                },
                {
                    "validator_address": "haqqvaloper1small", "moniker": "Small", "diff": -5,
                    "in_stride_list": True, "recorded": 7,
                },
            ]
        )

    def test_tracked_table_is_emitted_without_a_host_zone_file(self) -> None:
        stdout = io.StringIO()

        run_main(args=[str(self.drift_path()), CHAIN_ID], stdout=stdout)

        output = stdout.getvalue()
        self.assertIn('\t{Name: "big", Address: "haqqvaloper1big", Delta: mustInt("-853")},', output)
        self.assertIn("var HaqqExpectedTrackedDelegations = map[string]sdkmath.Int{", output)
        self.assertIn('\t"haqqvaloper1big": mustInt("900"),', output)
        self.assertIn('\t"haqqvaloper1small": mustInt("7"),', output)

    def test_both_tables_are_emitted_when_the_host_zone_matches(self) -> None:
        host_zone_path = write_host_zone({"haqqvaloper1big": 900, "haqqvaloper1small": 7})

        stdout = io.StringIO()

        run_main(args=[str(self.drift_path()), CHAIN_ID, "--host-zone-json", str(host_zone_path)], stdout=stdout)

        output = stdout.getvalue()
        self.assertIn(
            '{Name: "haqqvaloper1big-stride-name", Address: "haqqvaloper1big", Delta: mustInt("-853")},', output,
        )
        self.assertIn('\t"haqqvaloper1big": mustInt("900"),', output)
        self.assertIn('\t"haqqvaloper1small": mustInt("7"),', output)

    def test_a_moved_host_zone_exits_without_printing_any_table(self) -> None:
        host_zone_path = write_host_zone({"haqqvaloper1big": 900, "haqqvaloper1small": 6})
        stdout = io.StringIO()

        with self.assertRaises(SystemExit) as raised:
            run_main(args=[str(self.drift_path()), CHAIN_ID, "--host-zone-json", str(host_zone_path)], stdout=stdout)

        self.assertIn("haqqvaloper1small: drift recorded 7, host zone tracked 6", raised.exception.code)
        self.assertEqual("", stdout.getvalue())

    def go_file(self, big_delta: int) -> pathlib.Path:
        text = (
            "// Generated long ago\n"
            "var HaqqDelegationDeltas = []DelegationDelta{\n"
            f'\t{{Name: "old", Address: "haqqvaloper1big", Delta: mustInt("{big_delta}")}},\n'
            '\t{Name: "x",   Address: "haqqvaloper1small", Delta: mustInt("5")},\n'
            "}\n\n"
            "var HaqqExpectedTrackedDelegations = map[string]sdkmath.Int{\n"
            '\t"haqqvaloper1big": mustInt("900"),\n'
            '\t"haqqvaloper1small":   mustInt("7"),\n'
            "}\n"
        )
        path = pathlib.Path(tempfile.mkdtemp()) / "haqq.go"
        path.write_text(text)
        return path

    def test_check_passes_when_the_file_already_holds_the_generated_tables(self) -> None:
        stdout = io.StringIO()
        check_path = self.go_file(big_delta=-853)

        run_main(args=[str(self.drift_path()), CHAIN_ID, "--check", str(check_path)], stdout=stdout)

        self.assertEqual(
            f"RESULT: PASS — {check_path} already matches the measured drift (2 deltas)",
            stdout.getvalue().splitlines()[-1],
        )

    def test_check_fails_when_the_file_is_stale(self) -> None:
        stdout = io.StringIO()
        check_path = self.go_file(big_delta=-800)

        with self.assertRaises(SystemExit) as raised:
            run_main(args=[str(self.drift_path()), CHAIN_ID, "--check", str(check_path)], stdout=stdout)

        self.assertEqual(1, raised.exception.code)
        last_line = stdout.getvalue().splitlines()[-1]
        self.assertTrue(last_line.startswith(f"RESULT: FAIL — the table in {check_path} is stale: replace it"))

    def test_without_check_the_last_line_is_done(self) -> None:
        stdout = io.StringIO()

        run_main(args=[str(self.drift_path()), CHAIN_ID], stdout=stdout)

        self.assertEqual(
            "RESULT: DONE — 2 deltas generated; paste into app/upgrades/v35/haqq.go", stdout.getvalue().splitlines()[-1],
        )


if __name__ == "__main__":
    unittest.main()
