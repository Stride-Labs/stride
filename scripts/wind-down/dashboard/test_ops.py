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

import config
import multisig
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
                        {
                            "id": "drift",
                            "text": "Measure drift",
                            "command": "python3 scripts/wind-down/measure_delegation_drift.py\npython3 scripts/wind-down/gen_delta_table.py",
                            "expect": "RESULT: PASS; on FAIL refresh the delegations",
                        },
                    ],
                }
            ],
        },
        {"date": "2026-10-14", "title": "Next", "windows": [{"steps": [{"id": "later", "text": "Later"}]}]},
    ],
}


# The uniform transfer-day groups (spec §4): one per zone set, ids `<prefix>-<suffix>` in this order, every step with
# the group's zones and the live check or Multisig link listed here.
TRANSFER_DAY_GROUPS = ("haqq", "osmo", "tia", "d21", "d28")
TRANSFER_DAY_CHECKS: dict[str, dict[str, Any]] = {
    "claimable": {"auto": {"tab": "funds", "path": "records.pending_before_claimable", "equals": 0}},
    "claims": {"auto": {"tab": "funds", "path": "records.user_redemption_records", "equals": 0}},
    "settled": {"auto": {"tab": "funds", "path": "funds_settled"}},
    "channels": {"auto": {"tab": "channels", "path": "ica_channels_open"}},
    "transfers": {"multisig": "ica-transfers"},
    "landed": {"auto": {"tab": "funds", "path": "transfers_landed"}},
    "join-pools": {"multisig": "pool-funding", "auto": {"tab": "pools", "path": "pools_funded"}},
}
# Celestia's extra steps, keyed by the uniform step they follow.
CELESTIA_EXTRAS = {"claims": ["tia-staketia-sweep", "tia-staketia-paid"], "channels": ["tia-claim-balance"]}
AUTO_TABS = {"channels", "validators", "funds", "pools", "multisig"}
BUNDLE_PATHS = {"seeded", "done"}  # what a live check on a Multisig set's bundles can read
ALL_ZONES = [zone.chain_id for zone in config.ZONES]  # the eleven zones in the order every eleven-zone step lists them


def _write_json(path: pathlib.Path, body: dict[str, Any]) -> None:
    path.write_text(json.dumps(body))


def _expected_group_ids(prefix: str) -> list[str]:
    ids: list[str] = []
    for suffix in TRANSFER_DAY_CHECKS:
        ids.append(f"{prefix}-{suffix}")
        if prefix == "tia":
            ids.extend(CELESTIA_EXTRAS.get(suffix, []))
    return ids


def _multisig_reference_error(step: dict[str, Any], set_ids: set[str]) -> str | None:
    """Why a step's `multisig` link (`<set-id>` or `<set-id>/<zone>`) would not resolve on the page; None when it does."""
    set_id, _, zone = step["multisig"].partition("/")
    if set_id not in set_ids:
        return f"unknown tx set {set_id}"
    if zone and zone not in step.get("zones", []):
        return f"zone {zone} is not one of the step's zones"
    return None


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

    def test_command_and_expect_are_optional_strings(self) -> None:
        for day in self.plan["days"]:
            for window in day["windows"]:
                for step in window["steps"]:
                    for field in ("command", "expect"):
                        if field in step:
                            self.assertIsInstance(step[field], str, step["id"])
                            self.assertTrue(step[field].strip(), step["id"])

    def test_a_step_with_command_and_expect_is_tickable_like_any_other(self) -> None:
        step = SAMPLE_PLAN["days"][0]["windows"][0]["steps"][2]

        self.assertIn("\n", step["command"])
        self.assertIn(step["id"], ops.step_ids(plan=SAMPLE_PLAN))

    def _steps(self) -> list[dict[str, Any]]:
        return [step for day in self.plan["days"] for window in day["windows"] for step in window["steps"]]

    def _windows(self) -> list[dict[str, Any]]:
        return [window for day in self.plan["days"] for window in day["windows"]]

    def _step(self, step_id: str) -> dict[str, Any]:
        return next(step for step in self._steps() if step["id"] == step_id)

    def _window_of(self, step_id: str) -> dict[str, Any]:
        return next(window for window in self._windows() if any(step["id"] == step_id for step in window["steps"]))

    def _tx_sets_by_id(self) -> dict[str, multisig.TxSet]:
        # Three snapshots: the transfer and pool-funding sets read the Funds and Pools caches as well.
        tx_sets = multisig.tx_sets(validators_data=None, funds_data=None, pools_data=None)
        return {tx_set.id: tx_set for tx_set in tx_sets}

    def test_every_transfer_day_group_has_the_uniform_steps_in_order(self) -> None:
        for prefix in TRANSFER_DAY_GROUPS:
            with self.subTest(prefix=prefix):
                window = self._window_of(f"{prefix}-claimable")
                self.assertEqual([step["id"] for step in window["steps"]], _expected_group_ids(prefix=prefix))

    def test_transfer_day_steps_share_their_zones_and_carry_the_live_check_or_link(self) -> None:
        for prefix in TRANSFER_DAY_GROUPS:
            zones = self._step(f"{prefix}-claimable")["zones"]
            self.assertTrue(zones, prefix)
            for suffix, expected in TRANSFER_DAY_CHECKS.items():
                step = self._step(f"{prefix}-{suffix}")
                with self.subTest(step=step["id"]):
                    self.assertEqual(step["zones"], zones)
                    self.assertEqual({key: step[key] for key in ("auto", "multisig") if key in step}, expected)

    def test_celestia_extras_are_live_checked_and_linked(self) -> None:
        paid = self._step("tia-staketia-paid")
        self.assertEqual(paid["zones"], ["celestia"])
        self.assertEqual(paid["auto"], {"tab": "funds", "path": "records.staketia_claim_ready", "equals": True})
        self.assertEqual(self._step("tia-claim-balance")["multisig"], "staketia-claim-balance")

    def test_every_multisig_reference_resolves_to_a_set_and_one_of_the_steps_zones(self) -> None:
        set_ids = set(self._tx_sets_by_id())
        referencing = [step for step in self._steps() if "multisig" in step]

        self.assertTrue(referencing)
        for step in referencing:
            with self.subTest(step=step["id"]):
                self.assertIsNone(_multisig_reference_error(step=step, set_ids=set_ids))
                self.assertNotIn("command", step, "the Multisig tab is the source of its commands")

    def test_drain_sets_name_their_steps(self) -> None:
        sets_by_id = self._tx_sets_by_id()

        self.assertEqual(sets_by_id["live-test-undelegate"].step_id, "drain-live-test")
        self.assertEqual(sets_by_id["full-drain"].step_id, "drain-rest")

    # The voting week's pool prep (pool-prep spec §4): the two seeding steps are live-checked per zone from the Pools
    # tab, the creation step links the pool-creation set per zone and keeps its pools_ready check.
    def test_seeding_and_creation_are_live_checked_per_bundle_and_the_canonical_supply_per_zone(self) -> None:
        self.assertEqual(self._step("vote-seed-routes")["auto"], {"tab": "multisig", "set": "pool-creation", "path": "seeded"})
        self.assertNotIn("zones", self._step("vote-seed-routes"))
        canonical = self._step("vote-canonical-supply")
        self.assertEqual((canonical["zones"], canonical["auto"]), (ALL_ZONES, {"tab": "pools", "path": "canonical_seeded"}))
        create = self._step("vote-pools-create")
        self.assertEqual((create["multisig"], create["auto"]), ("pool-creation", {"tab": "multisig", "set": "pool-creation", "path": "done"}))
        self.assertIn("bundle 4", create["text"])
        self.assertEqual(self._step("vote-assetlist")["zones"], ALL_ZONES)

    def test_the_injective_block_and_the_single_assetlist_step(self) -> None:
        injective = next(day for day in self.plan["days"] if day["date"] == "2026-10-14")
        self.assertEqual([step["id"] for step in injective["windows"][0]["steps"]], ["inj-unblock", "inj-seed", "inj-create"])
        self.assertEqual(self._step("inj-create")["auto"], {"tab": "multisig", "set": "pool-creation", "path": "done", "members": ".injective"})
        self.assertEqual([step["id"] for step in self._steps() if step["id"].endswith("-assetlist")], ["vote-assetlist"])

    def test_pool_creation_set_is_first_and_names_its_step(self) -> None:
        # Fails until the pool-creation set lands (pool-prep spec §2, built separately): the plan references it already.
        sets_by_id = self._tx_sets_by_id()

        self.assertEqual(next(iter(sets_by_id)), "pool-creation")
        self.assertEqual(sets_by_id["pool-creation"].step_id, "vote-pools-create")

    def test_no_step_command_names_the_retired_pool_check(self) -> None:
        for step in self._steps():
            self.assertNotIn("check_transmuter_pool.py", step.get("command", ""), step["id"])

    def test_the_export_based_coverage_check_runs_only_at_the_halt(self) -> None:
        running = [step["id"] for step in self._steps() if "coverage_check.py" in step.get("command", "")]
        self.assertEqual(running, ["halt-pools"])

    def test_every_auto_check_reads_a_dashboard_collector_per_zone(self) -> None:
        autos = [step for step in self._steps() if "auto" in step]

        self.assertTrue(autos)
        set_ids = {tx_set.id for tx_set in multisig.tx_sets(validators_data=None, funds_data=None, pools_data=None)}
        for step in autos:
            self.assertIn(step["auto"]["tab"], AUTO_TABS, step["id"])
            if step["auto"]["tab"] == "multisig":  # read per bundle of a set, not per zone
                self.assertIn(step["auto"]["set"], set_ids, step["id"])
                self.assertIn(step["auto"]["path"], BUNDLE_PATHS, step["id"])
                continue
            self.assertTrue(step.get("zones"), f"{step['id']}: a live check is read per zone")

    def test_drain_rest_is_done_when_no_validator_holds_a_whole_token(self) -> None:
        step = next(step for step in self._steps() if step["id"] == "drain-rest")

        self.assertEqual(step["auto"], {"tab": "validators", "path": "funded_count", "equals": 0})

    def test_every_window_start_is_an_iso_utc_timestamp(self) -> None:
        starts = [window["start"] for window in self._windows() if "start" in window]

        self.assertTrue(starts)
        for start in starts:
            parsed = datetime.datetime.fromisoformat(start)
            self.assertEqual(parsed.utcoffset(), datetime.timedelta(0), start)

    def test_at_least_is_an_int_and_never_combined_with_equals(self) -> None:
        autos = [step["auto"] for step in self._steps() if "at_least" in step.get("auto", {})]

        self.assertTrue(autos)
        for auto in autos:
            self.assertIsInstance(auto["at_least"], int, auto)
            self.assertNotIsInstance(auto["at_least"], bool, auto)
            self.assertNotIn("equals", auto)

    def test_upgrade_time_matches_the_plan_anchor(self) -> None:
        self.assertEqual(config.UPGRADE_TIME, self.plan["anchors"]["upgrade"])


class MultisigReferenceTest(unittest.TestCase):
    """Both forms of a step's `multisig` link, resolved the way the Ops page resolves them."""

    SET_IDS = {"ica-transfers"}

    def test_link_forms(self) -> None:
        cases = [
            ({"multisig": "ica-transfers", "zones": ["celestia"]}, None),
            ({"multisig": "ica-transfers"}, None),
            ({"multisig": "ica-transfers/celestia", "zones": ["celestia", "osmosis-1"]}, None),
            ({"multisig": "ica-transfers/juno-1", "zones": ["celestia"]}, "zone juno-1 is not one of the step's zones"),
            ({"multisig": "ica-transfers/celestia"}, "zone celestia is not one of the step's zones"),
            ({"multisig": "pool-funding", "zones": ["celestia"]}, "unknown tx set pool-funding"),
        ]
        for step, expected in cases:
            with self.subTest(step=step):
                self.assertEqual(_multisig_reference_error(step=step, set_ids=self.SET_IDS), expected)


class StepIdsTest(unittest.TestCase):
    def test_zone_steps_expand_after_their_parent(self) -> None:
        self.assertEqual(
            ops.step_ids(plan=SAMPLE_PLAN),
            ["close-channel", "close-channel:celestia", "close-channel:cosmoshub-4", "verify", "drift", "later"],
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
