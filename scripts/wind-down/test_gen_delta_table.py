# scripts/wind-down/test_gen_delta_table.py
import json
import pathlib
import tempfile
import unittest

import gen_delta_table


def write_drift(rows: list[dict]) -> pathlib.Path:
    payload = {"generated_at": "2026-09-29T00:00:00Z", "zones": {"haqq_11235-1": {"validators": rows}}}
    path = pathlib.Path(tempfile.mkdtemp()) / "drift.json"
    path.write_text(json.dumps(payload))
    return path


class GenDeltaTableTest(unittest.TestCase):
    def test_emits_one_entry_per_nonzero_delta_sorted_by_magnitude(self) -> None:
        drift_path = write_drift(
            [
                {"validator_address": "haqqvaloper1small", "moniker": "Small Node", "diff": -5, "in_stride_list": True},
                {"validator_address": "haqqvaloper1big", "moniker": "Big [Node] ⚡", "diff": 853, "in_stride_list": True},
                {"validator_address": "haqqvaloper1zero", "moniker": "Zero", "diff": 0, "in_stride_list": True},
                {"validator_address": "haqqvaloper1foreign", "moniker": "Foreign", "diff": 7, "in_stride_list": False},
            ]
        )

        entries = gen_delta_table.load_deltas(drift_path=drift_path, chain_id="haqq_11235-1", names={})

        self.assertEqual(
            [("bignode", "haqqvaloper1big", -853), ("smallnode", "haqqvaloper1small", 5)],
            [(entry.name, entry.address, entry.delta) for entry in entries],
        )

    def test_names_come_from_the_host_zone_when_given(self) -> None:
        drift_path = write_drift([{"validator_address": "haqqvaloper1big", "moniker": "Big", "diff": 1, "in_stride_list": True}])

        entries = gen_delta_table.load_deltas(drift_path=drift_path, chain_id="haqq_11235-1", names={"haqqvaloper1big": "bigstride"})

        self.assertEqual("bigstride", entries[0].name)

    def test_render_go_literal(self) -> None:
        rendered = gen_delta_table.render_go(
            entries=[gen_delta_table.DeltaEntry(name="big", address="haqqvaloper1big", delta=-853)],
            var_name="HaqqDelegationDeltas",
        )

        self.assertEqual(
            'var HaqqDelegationDeltas = []DelegationDelta{\n'
            '\t{Name: "big", Address: "haqqvaloper1big", Delta: mustInt("-853")},\n'
            '}\n',
            rendered,
        )


if __name__ == "__main__":
    unittest.main()
