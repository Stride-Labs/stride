"""The ledger: append-only JSON lines, the batch states derived from them, and gas calibration."""

import pathlib
import tempfile
import unittest

from sweep import ledger

T1 = ledger.Transfer(address="stride1a", denom="stuatom", amount=1_000_000, channel="channel-5", receiver="osmo1a")
T2 = ledger.Transfer(address="stride1a", denom="ustrd", amount=5, channel="channel-5", receiver="osmo1a")


class LedgerFileTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.path = pathlib.Path(self.tmp.name) / "ledger.jsonl"

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_missing_file_reads_empty(self) -> None:
        self.assertEqual(ledger.read(path=self.path), [])

    def test_round_trip_keeps_every_field(self) -> None:
        submitted = ledger.submitted(run_id=1, batch_id="001-001", tx_hash="AB", at="t0", addresses=1, transfers_estimate=2, gas_wanted=300)
        confirmed = ledger.confirmed(batch_id="001-001", tx_hash="AB", at="t1", height=10, gas_used=240, transfers=[T1, T2], skipped=[ledger.Skip(address="stride1b", reason="interchain account")])
        ledger.append(event=submitted, path=self.path)
        ledger.append(event=confirmed, path=self.path)
        self.assertEqual(ledger.read(path=self.path), [submitted, confirmed])
        self.assertEqual(len(self.path.read_text().splitlines()), 2)

    def test_truncated_line_is_reported_with_its_number(self) -> None:
        ledger.append(event=ledger.lost(batch_id="001-001", tx_hash="AB", at="t"), path=self.path)
        with self.path.open("a") as handle:
            handle.write('{"kind": "confirmed", "batch_id": "001-0')
        with self.assertRaisesRegex(ledger.LedgerError, "line 2"):
            ledger.read(path=self.path)

    def test_valid_json_that_is_not_an_event_is_reported_with_its_number(self) -> None:
        ledger.append(event=ledger.lost(batch_id="001-001", tx_hash="AB", at="t"), path=self.path)
        with self.path.open("a") as handle:
            handle.write("5\n")
        with self.assertRaisesRegex(ledger.LedgerError, "line 2"):
            ledger.read(path=self.path)


class DerivedStateTests(unittest.TestCase):
    def test_batch_states_and_unresolved(self) -> None:
        events = [
            ledger.submitted(run_id=1, batch_id="001-001", tx_hash="A", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.confirmed(batch_id="001-001", tx_hash="A", at="t", height=1, gas_used=1, transfers=[T1], skipped=[]),
            ledger.submitted(run_id=1, batch_id="001-002", tx_hash="B", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.submitted(run_id=1, batch_id="001-003", tx_hash="C", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.failed(batch_id="001-003", tx_hash="C", at="t", code=11, codespace="sdk", raw_log="out of gas"),
            ledger.submitted(run_id=1, batch_id="001-004", tx_hash="D", at="t", addresses=1, transfers_estimate=1, gas_wanted=1),
            ledger.lost(batch_id="001-004", tx_hash="D", at="t"),
        ]
        states = ledger.batch_states(events=events)
        self.assertEqual(states, {
            "001-001": ledger.BatchState.CONFIRMED, "001-002": ledger.BatchState.SUBMITTED,
            "001-003": ledger.BatchState.FAILED, "001-004": ledger.BatchState.LOST,
        })
        self.assertEqual([event.batch_id for event in ledger.unresolved(events=events)], ["001-002"])
        self.assertEqual(ledger.latest_run_id(events=events), 1)
        self.assertEqual(ledger.confirmed_transfers(events=events), {"stride1a": [T1]})

    def test_calibration_uses_the_worst_confirmed_ratio_with_margin(self) -> None:
        self.assertIsNone(ledger.calibration(events=[]))
        events = [
            ledger.confirmed(batch_id="001-001", tx_hash="A", at="t", height=1, gas_used=200_000, transfers=[T1, T2], skipped=[]),
            ledger.confirmed(batch_id="001-002", tx_hash="B", at="t", height=2, gas_used=330_000, transfers=[T1, T2, T1], skipped=[]),
            ledger.confirmed(batch_id="001-003", tx_hash="C", at="t", height=3, gas_used=50_000, transfers=[], skipped=[]),
        ]
        self.assertEqual(ledger.calibration(events=events), 132_000)  # 110,000 per transfer x 1.2


if __name__ == "__main__":
    unittest.main()
