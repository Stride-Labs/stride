import collections
import datetime
import http.server
import json
import pathlib
import tempfile
import threading
import unittest
import urllib.error
import urllib.request
from typing import Any
from unittest import mock

import ops
import server

SAMPLE_PLAN: dict[str, Any] = {
    "anchors": {"upgrade": "2026-10-12T12:00:00Z"},
    "days": [
        {
            "date": "2026-10-12",
            "until": "2026-10-13",
            "title": "Upgrade day",
            "avoid": ["do not restart the relayer"],
            "windows": [
                {
                    "label": "12:00 UTC",
                    "steps": [
                        {"id": "close-channel", "text": "Close the channel", "zones": ["celestia", "cosmoshub-4"]},
                        {"id": "verify", "text": "Verify", "ref": "§9", "conditional": "if slipped"},
                    ],
                }
            ],
        },
        {"date": "2026-10-14", "title": "Next", "windows": [{"steps": [{"id": "later", "text": "Later"}]}]},
    ],
}


def _write_json(path: pathlib.Path, body: dict[str, Any]) -> None:
    path.write_text(json.dumps(body))


class RealPlanTest(unittest.TestCase):
    """The committed plan (or the sample while ops/plan.json is not there yet) keeps the invariants the page relies on."""

    def setUp(self) -> None:
        self.plan = ops.load_plan() if ops.PLAN_PATH.exists() else SAMPLE_PLAN

    def test_plan_loads_with_days(self) -> None:
        self.assertTrue(self.plan["days"])

    def test_ids_are_unique_including_zone_ids(self) -> None:
        counts = collections.Counter(ops.step_ids(plan=self.plan))
        self.assertEqual([check_id for check_id, count in counts.items() if count > 1], [])

    def test_every_block_has_a_valid_date_and_until_not_before_it(self) -> None:
        for day in self.plan["days"]:
            date = datetime.date.fromisoformat(day["date"])
            if "until" in day:
                self.assertGreaterEqual(datetime.date.fromisoformat(day["until"]), date, day["title"])

    def test_every_step_has_id_and_text(self) -> None:
        for day in self.plan["days"]:
            for window in day["windows"]:
                for step in window["steps"]:
                    self.assertTrue(step["id"], step)
                    self.assertTrue(step["text"], step)


class StepIdsTest(unittest.TestCase):
    def test_zone_steps_expand_after_their_parent(self) -> None:
        self.assertEqual(
            ops.step_ids(plan=SAMPLE_PLAN),
            ["close-channel", "close-channel:celestia", "close-channel:cosmoshub-4", "verify", "later"],
        )

    def test_today_is_a_utc_iso_date(self) -> None:
        self.assertEqual(ops.today(), datetime.datetime.now(datetime.UTC).date().isoformat())


class ParseCheckRequestTest(unittest.TestCase):
    def test_valid_body(self) -> None:
        request = ops.parse_check_request(body={"id": "verify", "done": True, "by": " sam "})
        self.assertEqual(request, ops.CheckRequest(check_id="verify", done=True, by="sam"))

    def test_rejects_bad_bodies(self) -> None:
        bad_bodies = [
            None,
            [],
            {"done": True, "by": "sam"},
            {"id": "", "done": True, "by": "sam"},
            {"id": 3, "done": True, "by": "sam"},
            {"id": "verify", "by": "sam"},
            {"id": "verify", "done": "yes", "by": "sam"},
            {"id": "verify", "done": True},
            {"id": "verify", "done": True, "by": "  "},
            {"id": "verify", "done": True, "by": "x" * (ops.MAX_NAME_LENGTH + 1)},
        ]
        for body in bad_bodies:
            with self.subTest(body=body), self.assertRaises(ops.InvalidCheckError):
                ops.parse_check_request(body=body)


class RecordCheckTest(unittest.TestCase):
    def setUp(self) -> None:
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.plan_path = pathlib.Path(directory.name) / "plan.json"
        self.status_path = pathlib.Path(directory.name) / "status.json"
        _write_json(path=self.plan_path, body=SAMPLE_PLAN)
        _write_json(path=self.status_path, body={})

    def _record(self, check_id: str, done: bool, by: str = "sam") -> dict[str, Any]:
        return ops.record_check(
            check_id=check_id, done=done, by=by, plan_path=self.plan_path, status_path=self.status_path
        )

    def test_tick_round_trips_to_disk(self) -> None:
        returned = self._record(check_id="verify", done=True)

        on_disk = ops.load_status(status_path=self.status_path)
        self.assertEqual(returned, on_disk)
        self.assertEqual(on_disk["verify"]["by"], "sam")
        self.assertTrue(on_disk["verify"]["done"])
        self.assertEqual(datetime.datetime.fromisoformat(on_disk["verify"]["at"]).utcoffset(), datetime.timedelta(0))

    def test_zone_id_is_accepted(self) -> None:
        self.assertIn("close-channel:celestia", self._record(check_id="close-channel:celestia", done=True))

    def test_untick_deletes_the_key_and_keeps_others(self) -> None:
        self._record(check_id="verify", done=True)
        self._record(check_id="later", done=True)

        returned = self._record(check_id="verify", done=False)

        self.assertEqual(list(returned), ["later"])
        self.assertEqual(list(ops.load_status(status_path=self.status_path)), ["later"])

    def test_untick_of_an_unticked_id_is_a_no_op(self) -> None:
        self.assertEqual(self._record(check_id="verify", done=False), {})

    def test_file_is_sorted_two_space_indented_json(self) -> None:
        self._record(check_id="verify", done=True)
        self._record(check_id="close-channel:celestia", done=True)

        text = self.status_path.read_text()
        self.assertEqual(text, json.dumps(json.loads(text), indent=2, sort_keys=True) + "\n")
        self.assertLess(text.index("close-channel:celestia"), text.index("verify"))

    def test_unknown_id_is_rejected_and_nothing_written(self) -> None:
        with self.assertRaises(ops.UnknownStepError):
            self._record(check_id="nope", done=True)
        with self.assertRaises(ops.UnknownStepError):
            self._record(check_id="verify:celestia", done=True)  # zones only exist on steps that declare them

        self.assertEqual(ops.load_status(status_path=self.status_path), {})

    def test_no_temp_files_left_behind(self) -> None:
        self._record(check_id="verify", done=True)
        self.assertEqual(sorted(path.name for path in self.status_path.parent.iterdir()), ["plan.json", "status.json"])


class RoutesTest(unittest.TestCase):
    """The two ops routes over a real socket, with the plan and status paths pointed at a temp directory."""

    def setUp(self) -> None:
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.plan_path = pathlib.Path(directory.name) / "plan.json"
        self.status_path = pathlib.Path(directory.name) / "status.json"
        _write_json(path=self.plan_path, body=SAMPLE_PLAN)
        _write_json(path=self.status_path, body={})
        for name, value in (("PLAN_PATH", self.plan_path), ("STATUS_PATH", self.status_path)):
            patcher = mock.patch.object(ops, name, value)
            patcher.start()
            self.addCleanup(patcher.stop)

        self.httpd = http.server.ThreadingHTTPServer(("127.0.0.1", 0), server.Handler)
        threading.Thread(target=self.httpd.serve_forever, daemon=True).start()
        self.addCleanup(self.httpd.server_close)
        self.addCleanup(self.httpd.shutdown)
        self.base = f"http://127.0.0.1:{self.httpd.server_address[1]}"

    def _request(self, path: str, body: Any = None, raw: bytes | None = None) -> tuple[int, Any]:
        data = raw if raw is not None else (None if body is None else json.dumps(body).encode())
        request = urllib.request.Request(self.base + path, data=data, method="GET" if data is None else "POST")
        try:
            with urllib.request.urlopen(request) as response:
                return response.status, json.load(response)
        except urllib.error.HTTPError as error:
            return error.code, json.load(error)

    def test_get_returns_plan_status_and_today(self) -> None:
        code, body = self._request("/api/ops")

        self.assertEqual(code, 200)
        self.assertEqual(body, {"plan": SAMPLE_PLAN, "status": {}, "today": ops.today()})

    def test_get_reads_the_plan_from_disk_each_time(self) -> None:
        _write_json(path=self.plan_path, body={"days": []})
        self.assertEqual(self._request("/api/ops")[1]["plan"], {"days": []})

    def test_get_500_when_the_plan_is_missing_or_broken(self) -> None:
        self.plan_path.write_text("{not json")
        self.assertEqual(self._request("/api/ops")[0], 500)

        self.plan_path.unlink()
        self.assertEqual(self._request("/api/ops")[0], 500)

    def test_post_ticks_and_unticks(self) -> None:
        code, status = self._request("/api/ops/check", {"id": "verify", "done": True, "by": "sam"})
        self.assertEqual(code, 200)
        self.assertEqual(status["verify"]["by"], "sam")
        self.assertEqual(self._request("/api/ops")[1]["status"], status)

        code, status = self._request("/api/ops/check", {"id": "verify", "done": False, "by": "sam"})
        self.assertEqual((code, status), (200, {}))
        self.assertEqual(json.loads(self.status_path.read_text()), {})

    def test_post_400_on_bad_body_unknown_id_and_non_json(self) -> None:
        for body in ({"id": "verify", "done": "yes", "by": "sam"}, {"id": "nope", "done": True, "by": "sam"}):
            with self.subTest(body=body):
                self.assertEqual(self._request("/api/ops/check", body)[0], 400)
        self.assertEqual(self._request("/api/ops/check", raw=b"not json")[0], 400)
        self.assertEqual(self._request("/api/ops/check", raw=b"")[0], 400)

    def test_post_500_when_the_plan_is_missing(self) -> None:
        self.plan_path.unlink()
        self.assertEqual(self._request("/api/ops/check", {"id": "verify", "done": True, "by": "sam"})[0], 500)

    def test_ops_is_not_a_collector_and_other_routes_still_work(self) -> None:
        self.assertNotIn("ops", server.COLLECTORS)
        self.assertEqual(self._request("/api/config")[0], 200)
        self.assertEqual(self._request("/api/nope")[0], 404)
        self.assertEqual(self._request("/api/refresh/nope", {})[0], 404)


if __name__ == "__main__":
    unittest.main()
