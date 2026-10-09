"""The runner: preflight rows, the exact strided command line, gas-estimate parsing, the broadcast and poll loop over
fixture responses, and the stop conditions. chainio is patched throughout."""

import contextlib
import datetime
import io
import json
import pathlib
import tempfile
import unittest
from decimal import Decimal
from unittest import mock

from sweep import chainio, cli, config, holders, ledger, planner

DENOMS = [holders.SweepDenom(denom="stuatom", symbol="stATOM", decimals=6, price_usd=Decimal("6"), destination="osmosis-1", channel="channel-5"),
          holders.SweepDenom(denom="ustrd", symbol="STRD", decimals=6, price_usd=Decimal("0.05"), destination="osmosis-1", channel="channel-5")]
A1 = "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"
A2 = "stride1am99pcvynqqhyrwqfvfmnvxjk96rn46le9j65c"
NOW = datetime.datetime(2026, 11, 10, 9, 0, tzinfo=datetime.UTC)


def make_plan(created_at: str = NOW.isoformat(), addresses: tuple[str, ...] = (A1, A2)) -> planner.Plan:
    holder_list = [holders.Holder(address=a, balances={"stuatom": 1_000_000, "ustrd": 10}, usd=Decimal(6), keyless=False) for a in addresses]
    hs = holders.HolderSet(height=100, denoms=DENOMS, holders=holder_list, excluded=[], skipped=[], below_floor=[])
    return planner.build_plan(holder_set=hs, floor_usd=Decimal(1), run_id=1, canary=0, test=False, gas_per_transfer=100_000,
                              max_addresses=1, gas_budget=40_000_000, created_at=created_at)


TX_RESPONSE = {"tx_response": {
    "height": "4242", "txhash": "AB12", "code": 0, "codespace": "", "raw_log": "", "gas_wanted": "260000", "gas_used": "201000",
    "events": [
        {"type": "message", "attributes": [{"key": "action", "value": "/stride.stakeibc.MsgSweepTokensOffStride"}]},
        {"type": "sweep_transfer", "attributes": [
            {"key": "module", "value": "stakeibc"}, {"key": "address", "value": A1}, {"key": "denom", "value": "stuatom"},
            {"key": "amount", "value": "1000000"}, {"key": "channel", "value": "channel-5"}, {"key": "receiver", "value": "osmo1x"}]},
        {"type": "sweep_transfer", "attributes": [
            {"key": "module", "value": "stakeibc"}, {"key": "address", "value": A1}, {"key": "denom", "value": "ustrd"},
            {"key": "amount", "value": "10"}, {"key": "channel", "value": "channel-5"}, {"key": "receiver", "value": "osmo1x"}]},
        {"type": "sweep_skipped", "attributes": [
            {"key": "module", "value": "stakeibc"}, {"key": "address", "value": A2}, {"key": "reason", "value": "interchain account"}]},
    ],
}}


class ParsingTests(unittest.TestCase):
    def test_command_line_is_exactly_the_documented_shape(self) -> None:
        line = cli.command_line(denoms=["stuatom", "ustrd"], file=pathlib.Path("/s/batch-001-001.txt"), gas=260_000, dry_run=False)
        self.assertEqual(line, [
            "tx", "stakeibc", "sweep-tokens-off-stride", "stuatom,ustrd", "/s/batch-001-001.txt",
            "--from", "stride-sweeper", "--keyring-backend", "test", "--chain-id", "stride-1", "--node", config.RPC,
            "--gas", "260000", "--gas-prices", "0.001ustrd", "--broadcast-mode", "sync", "-y", "--output", "json",
        ])
        dry = cli.command_line(denoms=["stuatom"], file=pathlib.Path("/s/b.txt"), gas=None, dry_run=True)
        self.assertIn("--dry-run", dry)
        self.assertNotIn("--gas", dry)
        self.assertNotIn("-y", dry)

    def test_parse_gas_estimate(self) -> None:
        self.assertEqual(cli.parse_gas_estimate(text="some noise\ngas estimate: 201000\n"), 201_000)
        self.assertIsNone(cli.parse_gas_estimate(text="Error: rpc error"))

    def test_parse_broadcast(self) -> None:
        result = cli.parse_broadcast(stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}))
        self.assertEqual(result, cli.BroadcastResult(code=0, tx_hash="AB12", codespace="", raw_log=""))

    def test_parse_broadcast_stops_without_a_txhash(self) -> None:
        for stdout in ["not json", "[1]", json.dumps({"code": 0}), json.dumps({"code": 0, "txhash": ""})]:
            with self.subTest(stdout=stdout), self.assertRaises(cli.RunStopped):
                cli.parse_broadcast(stdout=stdout)

    def test_batches_must_be_positive(self) -> None:
        for value in ["0", "-1"]:
            with self.subTest(value=value), contextlib.redirect_stderr(io.StringIO()), self.assertRaises(SystemExit):
                cli.parse_args(argv=["run", "--batches", value])

    def test_parse_tx_response_reads_sweep_events(self) -> None:
        outcome = cli.parse_tx_response(body=TX_RESPONSE)
        self.assertEqual((outcome.code, outcome.height, outcome.gas_used), (0, 4242, 201_000))
        self.assertEqual(outcome.transfers, [
            ledger.Transfer(address=A1, denom="stuatom", amount=1_000_000, channel="channel-5", receiver="osmo1x"),
            ledger.Transfer(address=A1, denom="ustrd", amount=10, channel="channel-5", receiver="osmo1x"),
        ])
        self.assertEqual(outcome.skipped, [ledger.Skip(address=A2, reason="interchain account")])


class PreflightTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)
        self.plan = make_plan()
        planner.write_plan(plan=self.plan, state_dir=self.state)
        self.rest = {
            "/cosmos/base/tendermint/v1beta1/node_info": {"default_node_info": {"network": "stride-1"}},
            "/ibc/core/channel/v1/channels/channel-5/ports/transfer": {"channel": {"state": "STATE_OPEN"}},
            f"/cosmos/bank/v1beta1/balances/{config.SWEEP_OPERATOR}/by_denom": {"balance": {"denom": "ustrd", "amount": "5000000"}},
        }
        self.commands = {
            ("version",): chainio.CommandResult(returncode=0, stdout="v35.0.0\n", stderr=""),
            ("keys", "show", "stride-sweeper", "--keyring-backend", "test", "-a"): chainio.CommandResult(returncode=0, stdout=config.SWEEP_OPERATOR + "\n", stderr=""),
        }
        mock.patch.object(chainio, "rest_get", side_effect=lambda path, params=None: self.rest[path]).start()
        mock.patch.object(chainio, "strided", side_effect=lambda args: self.commands[tuple(args)]).start()
        mock.patch.object(config, "STATE_DIR", self.state).start()
        self.addCleanup(mock.patch.stopall)
        self.addCleanup(self.tmp.cleanup)

    def checks(self, events: list[ledger.Event] | None = None, exclusions: dict | None = None, now: datetime.datetime = NOW) -> dict[str, cli.Check]:
        return {c.name: c for c in cli.preflight(plan=self.plan, events=events or [], exclusions=exclusions or {}, now=now)}

    def test_everything_passes_on_a_clean_setup(self) -> None:
        checks = self.checks()
        self.assertTrue(all(c.ok for c in checks.values()), [c for c in checks.values() if not c.ok])
        self.assertEqual(set(checks), {"chain id", "binary version", "operator key", "plan age", "batch files", "unresolved submissions",
                                       "exclusions and protocol addresses", "channels open", "operator fee balance"})

    def test_each_failure_mode(self) -> None:
        self.assertFalse(self.checks(now=NOW + datetime.timedelta(hours=7))["plan age"].ok)
        self.assertFalse(self.checks(now=NOW - datetime.timedelta(hours=1))["plan age"].ok)
        (self.state / "batch-001-001.txt").write_text(f"{A2}\n")
        self.assertFalse(self.checks()["batch files"].ok)
        planner.write_plan(plan=self.plan, state_dir=self.state)
        submitted = ledger.submitted(run_id=1, batch_id="001-001", tx_hash="X", at="t", addresses=1, transfers_estimate=2, gas_wanted=1)
        self.assertFalse(self.checks(events=[submitted])["unresolved submissions"].ok)
        exclusion = holders.Exclusion(address=A1, section="team", label="F5", reason="r")
        self.assertFalse(self.checks(exclusions={A1: exclusion})["exclusions and protocol addresses"].ok)
        self.rest["/ibc/core/channel/v1/channels/channel-5/ports/transfer"] = {"channel": {"state": "STATE_CLOSED"}}
        self.assertFalse(self.checks()["channels open"].ok)
        self.rest[f"/cosmos/bank/v1beta1/balances/{config.SWEEP_OPERATOR}/by_denom"] = {"balance": {"denom": "ustrd", "amount": "1"}}
        self.assertFalse(self.checks()["operator fee balance"].ok)
        self.commands[("version",)] = chainio.CommandResult(returncode=0, stdout="v34.0.0\n", stderr="")
        self.assertFalse(self.checks()["binary version"].ok)
        self.commands[("keys", "show", "stride-sweeper", "--keyring-backend", "test", "-a")] = chainio.CommandResult(returncode=0, stdout="stride1other\n", stderr="")
        self.assertFalse(self.checks()["operator key"].ok)

    def test_already_confirmed_batches_are_not_rechecked_for_exclusions(self) -> None:
        confirmed = [ledger.submitted(run_id=1, batch_id="001-001", tx_hash="X", at="t", addresses=1, transfers_estimate=2, gas_wanted=1),
                     ledger.confirmed(batch_id="001-001", tx_hash="X", at="t", height=1, gas_used=1, transfers=[], skipped=[])]
        exclusion = holders.Exclusion(address=A1, section="team", label="F5", reason="r")
        self.assertTrue(self.checks(events=confirmed, exclusions={A1: exclusion})["exclusions and protocol addresses"].ok)


class RunBatchTests(unittest.TestCase):
    def setUp(self) -> None:
        self.enterContext(contextlib.redirect_stdout(io.StringIO()))
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)
        self.ledger_path = self.state / "ledger.jsonl"
        self.plan = make_plan()
        planner.write_plan(plan=self.plan, state_dir=self.state)
        self.tier, self.batch = self.plan.pending_batches()[0]
        self.strided_calls: list[list[str]] = []
        self.tx_lookups = 0
        mock.patch.object(chainio, "now", return_value=NOW).start()
        mock.patch.object(chainio, "sleep").start()
        mock.patch.object(cli, "ask", return_value="y").start()
        self.addCleanup(mock.patch.stopall)
        self.addCleanup(self.tmp.cleanup)

    def fake_strided(self, args: list[str]) -> chainio.CommandResult:
        self.strided_calls.append(args)
        if "--dry-run" in args:
            return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
        return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")

    def fake_rest(self, path: str, params: dict | None = None) -> dict:
        self.tx_lookups += 1
        if self.tx_lookups < 3:
            raise chainio.NotFound(path)
        return TX_RESPONSE

    def test_simulate_broadcast_poll_and_record(self) -> None:
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided), mock.patch.object(chainio, "rest_get", side_effect=self.fake_rest):
            event = cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=False, dry_run=False)

        self.assertEqual(event.kind, ledger.Kind.CONFIRMED)
        self.assertEqual(self.strided_calls[0][-1], "--dry-run")
        self.assertIn("--gas", self.strided_calls[1])
        self.assertEqual(self.strided_calls[1][self.strided_calls[1].index("--gas") + 1], "260000")  # 200,000 x 1.3
        events = ledger.read(path=self.ledger_path)
        self.assertEqual([e.kind for e in events], [ledger.Kind.SUBMITTED, ledger.Kind.CONFIRMED])
        self.assertEqual(events[0].gas_wanted, 260_000)
        self.assertEqual(events[1].height, 4242)
        self.assertEqual(len(events[1].transfers), 2)
        self.assertEqual(events[1].skipped[0].address, A2)

    def test_dry_run_broadcasts_nothing(self) -> None:
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided):
            self.assertIsNone(cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=True))
        self.assertEqual(len(self.strided_calls), 1)
        self.assertFalse(self.ledger_path.exists())

    def test_declined_prompt_signs_nothing(self) -> None:
        with mock.patch.object(cli, "ask", return_value="n"), mock.patch.object(chainio, "strided", side_effect=self.fake_strided):
            self.assertIsNone(cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=False, dry_run=False))
        self.assertEqual(self.strided_calls, [])

    def test_over_the_block_limit_refuses_the_batch(self) -> None:
        def huge(args: list[str]) -> chainio.CommandResult:
            return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 90000000\n")
        with mock.patch.object(chainio, "strided", side_effect=huge):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        self.assertFalse(self.ledger_path.exists())

    def test_rejected_broadcast_is_a_failed_event(self) -> None:
        def rejected(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "CD34", "codespace": "sdk", "code": 13, "raw_log": "insufficient fee"}), stderr="")
        with mock.patch.object(chainio, "strided", side_effect=rejected):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        events = ledger.read(path=self.ledger_path)
        self.assertEqual([e.kind for e in events], [ledger.Kind.FAILED])
        self.assertEqual(events[0].code, 13)
        self.assertEqual(events[0].run_id, self.plan.run_id)
        self.assertEqual(planner.next_run_id(events=events), self.plan.run_id + 1)  # a rejection-only run is not reused

    def test_simulation_without_a_gas_estimate_stops_before_broadcasting(self) -> None:
        def no_estimate(args: list[str]) -> chainio.CommandResult:
            self.strided_calls.append(args)
            return chainio.CommandResult(returncode=1, stdout="", stderr="Error: rpc error")
        with mock.patch.object(chainio, "strided", side_effect=no_estimate):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        self.assertEqual(len(self.strided_calls), 1)
        self.assertIn("--dry-run", self.strided_calls[0])
        self.assertFalse(self.ledger_path.exists())

    def test_delivered_tx_with_a_nonzero_code_records_submitted_then_failed_and_stops(self) -> None:
        failed_response = {"tx_response": {**TX_RESPONSE["tx_response"], "code": 5, "codespace": "sdk", "raw_log": "boom", "events": []}}
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided), mock.patch.object(chainio, "rest_get", return_value=failed_response):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        events = ledger.read(path=self.ledger_path)
        self.assertEqual([e.kind for e in events], [ledger.Kind.SUBMITTED, ledger.Kind.FAILED])
        self.assertEqual((events[1].code, events[1].codespace, events[1].raw_log), (5, "sdk", "boom"))

    def test_unfound_tx_leaves_the_submission_unresolved(self) -> None:
        with mock.patch.object(chainio, "strided", side_effect=self.fake_strided), \
             mock.patch.object(chainio, "rest_get", side_effect=chainio.NotFound("x")), \
             mock.patch.object(config, "TX_WAIT_SECONDS", 6):
            with self.assertRaises(cli.RunStopped):
                cli.run_batch(plan=self.plan, tier=self.tier, batch=self.batch, state_dir=self.state, ledger_path=self.ledger_path, yes=True, dry_run=False)
        self.assertEqual([e.kind for e in ledger.read(path=self.ledger_path)], [ledger.Kind.SUBMITTED])


class RunLoopTests(unittest.TestCase):
    """`main(["run", ...])` over a plan with two single-address batches."""

    def setUp(self) -> None:
        self.enterContext(contextlib.redirect_stdout(io.StringIO()))
        self.tmp = tempfile.TemporaryDirectory()
        self.state = pathlib.Path(self.tmp.name)
        planner.write_plan(plan=make_plan(), state_dir=self.state)
        mock.patch.object(config, "STATE_DIR", self.state).start()
        mock.patch.object(config, "PLAN_PATH", self.state / "plan.json").start()
        mock.patch.object(config, "LEDGER_PATH", self.state / "ledger.jsonl").start()
        mock.patch.object(config, "EXCLUSIONS_PATH", self.state / "exclusions.json").start()
        (self.state / "exclusions.json").write_text('{"sections": []}')
        mock.patch.object(cli, "preflight", return_value=[cli.Check(name="all", ok=True, detail="stubbed")]).start()
        mock.patch.object(chainio, "now", return_value=NOW).start()
        mock.patch.object(chainio, "sleep").start()
        mock.patch.object(cli, "ask", return_value="a").start()
        self.addCleanup(setattr, cli, "_answered_all", False)
        self.addCleanup(mock.patch.stopall)
        self.addCleanup(self.tmp.cleanup)

    def test_stops_after_a_batch_with_a_skip_event(self) -> None:
        def fake_strided(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")
        with mock.patch.object(chainio, "strided", side_effect=fake_strided), mock.patch.object(chainio, "rest_get", return_value=TX_RESPONSE):
            self.assertEqual(cli.main(argv=["run"]), 1)
        events = ledger.read(path=self.state / "ledger.jsonl")
        self.assertEqual([e.batch_id for e in events], ["001-001", "001-001"])  # the second batch was never started

    def test_continue_on_skip_runs_every_batch_and_batches_flag_limits(self) -> None:
        def fake_strided(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")
        with mock.patch.object(chainio, "strided", side_effect=fake_strided), mock.patch.object(chainio, "rest_get", return_value=TX_RESPONSE):
            self.assertEqual(cli.main(argv=["run", "--continue-on-skip", "--batches", "1"]), 0)
            self.assertEqual([e.batch_id for e in ledger.read(path=self.state / "ledger.jsonl")], ["001-001", "001-001"])
            self.assertEqual(cli.main(argv=["run", "--continue-on-skip"]), 0)
            self.assertEqual([e.batch_id for e in ledger.read(path=self.state / "ledger.jsonl")], ["001-001", "001-001", "001-002", "001-002"])

    def test_unknown_tier_stops_naming_the_plan_tiers(self) -> None:
        with mock.patch("builtins.print") as printed:
            self.assertEqual(cli.main(argv=["run", "--tier", "keyless"]), 1)
        self.assertIn("the plan has:", str(printed.call_args_list))

    def test_dry_run_reports_the_simulated_count(self) -> None:
        simulated = chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
        with mock.patch.object(chainio, "strided", return_value=simulated), mock.patch("builtins.print") as printed:
            self.assertEqual(cli.main(argv=["run", "--dry-run"]), 0)
        self.assertIn("2 batch(es) simulated", str(printed.call_args_list))

    def seed_submitted(self) -> pathlib.Path:
        path = self.state / "ledger.jsonl"
        ledger.append(event=ledger.submitted(run_id=1, batch_id="001-001", tx_hash="AB12", at=NOW.isoformat(), addresses=1, transfers_estimate=2, gas_wanted=1), path=path)
        return path

    def test_failed_preflight_stops_before_any_strided_call(self) -> None:
        failing = [cli.Check(name="chain id", ok=False, detail="node reports other-1")]
        with mock.patch.object(cli, "preflight", return_value=failing), mock.patch.object(chainio, "strided", side_effect=AssertionError("strided")) as strided:
            self.assertEqual(cli.main(argv=["run"]), 1)
        strided.assert_not_called()
        self.assertFalse((self.state / "ledger.jsonl").exists())

    def test_a_failed_tx_stops_the_run_before_the_second_batch(self) -> None:
        def fake_strided(args: list[str]) -> chainio.CommandResult:
            if "--dry-run" in args:
                return chainio.CommandResult(returncode=0, stdout="", stderr="gas estimate: 200000\n")
            return chainio.CommandResult(returncode=0, stdout=json.dumps({"height": "0", "txhash": "AB12", "codespace": "", "code": 0, "raw_log": ""}), stderr="")
        failed_response = {"tx_response": {**TX_RESPONSE["tx_response"], "code": 5, "codespace": "sdk", "raw_log": "boom", "events": []}}
        with mock.patch.object(chainio, "strided", side_effect=fake_strided), mock.patch.object(chainio, "rest_get", return_value=failed_response):
            self.assertEqual(cli.main(argv=["run"]), 1)
        events = ledger.read(path=self.state / "ledger.jsonl")
        self.assertEqual([(e.batch_id, e.kind) for e in events], [("001-001", ledger.Kind.SUBMITTED), ("001-001", ledger.Kind.FAILED)])

    def test_resolve_records_confirmed_when_the_poll_finds_the_tx(self) -> None:
        path = self.seed_submitted()
        with mock.patch.object(chainio, "rest_get", return_value=TX_RESPONSE):
            self.assertEqual(cli.main(argv=["resolve"]), 0)
        self.assertEqual([e.kind for e in ledger.read(path=path)], [ledger.Kind.SUBMITTED, ledger.Kind.CONFIRMED])

    def test_resolve_records_failed_when_the_poll_finds_a_nonzero_code(self) -> None:
        path = self.seed_submitted()
        failed_response = {"tx_response": {**TX_RESPONSE["tx_response"], "code": 5, "codespace": "sdk", "raw_log": "boom", "events": []}}
        with mock.patch.object(chainio, "rest_get", return_value=failed_response):
            self.assertEqual(cli.main(argv=["resolve"]), 0)
        events = ledger.read(path=path)
        self.assertEqual([e.kind for e in events], [ledger.Kind.SUBMITTED, ledger.Kind.FAILED])
        self.assertEqual(events[1].code, 5)

    def test_resolve_records_lost_when_the_tx_never_appears(self) -> None:
        path = self.seed_submitted()
        with mock.patch.object(chainio, "rest_get", side_effect=chainio.NotFound("x")):
            self.assertEqual(cli.main(argv=["resolve", "--wait-seconds", "6"]), 0)
        self.assertEqual([e.kind for e in ledger.read(path=path)], [ledger.Kind.SUBMITTED, ledger.Kind.LOST])
        chainio.sleep.assert_called()

    def test_plan_refuses_with_an_unresolved_submission(self) -> None:
        self.seed_submitted()
        with mock.patch.object(holders, "read_holder_set", side_effect=AssertionError("holders")) as read_holders:
            self.assertEqual(cli.main(argv=["plan", "--floor-usd", "1"]), 1)
        read_holders.assert_not_called()

    def test_plan_without_test_or_floor_refuses(self) -> None:
        with mock.patch.object(holders, "read_holder_set", side_effect=AssertionError("holders")) as read_holders:
            self.assertEqual(cli.main(argv=["plan"]), 1)
        read_holders.assert_not_called()

    def test_status_needs_no_network(self) -> None:
        with mock.patch.object(chainio, "rest_get", side_effect=AssertionError("network")), mock.patch.object(chainio, "strided", side_effect=AssertionError("subprocess")):
            self.assertEqual(cli.main(argv=["status"]), 0)


if __name__ == "__main__":
    unittest.main()
