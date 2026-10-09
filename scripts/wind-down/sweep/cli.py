#!/usr/bin/env python3
"""The sweep runner.

    python3 scripts/wind-down/sweep/cli.py plan --floor-usd 100 [--canary 3] [--test] [--max-addresses 100] [--gas-budget 40000000]
    python3 scripts/wind-down/sweep/cli.py run [--tier test|canary|main|keyless] [--batches N] [--yes] [--dry-run] [--continue-on-skip]
    python3 scripts/wind-down/sweep/cli.py status
    python3 scripts/wind-down/sweep/cli.py resolve [--wait-seconds 600]

`plan` reads live holder state and writes state/plan.json plus one address file per batch. `run` walks the plan's
batches in tier order: preflight, then per batch a gas simulation, the signed broadcast from the stride-sweeper key,
and a poll for the tx result, each step appended to state/ledger.jsonl. It stops on any skip event, failed tx, or a
batch over the block gas limit. `status` reads the two files. `resolve` finishes a submission the poll gave up on.
"""

import argparse
import datetime
import json
import pathlib
import re
import sys
from dataclasses import dataclass
from decimal import Decimal

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # so `python3 sweep/cli.py` finds the package

from sweep import chainio, config, holders, ledger, planner  # noqa: E402

GAS_ESTIMATE_PATTERN = re.compile(r"gas estimate:\s*(\d+)")
EVENT_SWEEP_TRANSFER = "sweep_transfer"
EVENT_SWEEP_SKIPPED = "sweep_skipped"
STATE_OPEN = "STATE_OPEN"
ANSWER_YES = "y"
ANSWER_ALL = "a"


class RunStopped(Exception):
    """The run must not continue; the message says why."""


@dataclass(frozen=True)
class Check:
    name: str
    ok: bool
    detail: str


@dataclass(frozen=True)
class BroadcastResult:
    code: int
    tx_hash: str
    codespace: str
    raw_log: str


@dataclass(frozen=True)
class TxOutcome:
    code: int
    height: int
    gas_used: int
    codespace: str
    raw_log: str
    transfers: list[ledger.Transfer]
    skipped: list[ledger.Skip]


# ---- entry point


def _positive_int(text: str) -> int:
    value = int(text)
    if value < 1:
        raise argparse.ArgumentTypeError(f"must be at least 1, got {value}")
    return value


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv=argv)
    try:
        return args.command(args)
    except (RunStopped, holders.DenomError, holders.ExclusionsError, planner.PlanError, ledger.LedgerError, chainio.ChainError) as error:
        print(f"RESULT: FAIL — {error}")
        return 1


def parse_args(argv: list[str] | None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    commands = parser.add_subparsers(dest="command_name", required=True)

    plan = commands.add_parser("plan", help="read live state and write the batches")
    plan.add_argument("--floor-usd", type=Decimal, default=None, help="sweep every holder at or above this value (required unless --test)")
    plan.add_argument("--canary", type=int, default=0, help="this many of the smallest holders above the floor go first")
    plan.add_argument("--test", action="store_true", help=f"plan only the test address {config.TEST_ADDRESS}")
    plan.add_argument("--max-addresses", type=int, default=config.MAX_ADDRESSES_PER_BATCH)
    plan.add_argument("--gas-budget", type=int, default=config.GAS_BUDGET_PER_BATCH)
    plan.set_defaults(command=cmd_plan)

    run = commands.add_parser("run", help="sign and submit the pending batches in tier order")
    run.add_argument("--tier", type=planner.TierName, choices=list(planner.TierName), default=None)
    run.add_argument("--batches", type=_positive_int, default=None, help="stop after this many batches")
    run.add_argument("--yes", action="store_true", help="do not prompt per batch")
    run.add_argument("--dry-run", action="store_true", help="simulate and print the commands; broadcast nothing")
    run.add_argument("--continue-on-skip", action="store_true", help="keep going after a batch with sweep_skipped events")
    run.set_defaults(command=cmd_run)

    status = commands.add_parser("status", help="what the plan and ledger say; no network")
    status.set_defaults(command=cmd_status)

    resolve = commands.add_parser("resolve", help="finish submissions the poll gave up on")
    resolve.add_argument("--wait-seconds", type=int, default=config.RESOLVE_WAIT_SECONDS)
    resolve.set_defaults(command=cmd_resolve)
    return parser.parse_args(argv)


# ---- plan


def cmd_plan(args: argparse.Namespace) -> int:
    if not args.test and args.floor_usd is None:
        raise RunStopped("--floor-usd is required unless --test")
    events = ledger.read(path=config.LEDGER_PATH)
    pending = ledger.unresolved(events=events)
    if pending:
        raise RunStopped(f"{len(pending)} submitted batch(es) unresolved ({', '.join(e.batch_id for e in pending)}); run `resolve` first")

    exclusions = holders.load_exclusions(path=config.EXCLUSIONS_PATH)
    accounts = holders.AccountCache.load(path=config.ACCOUNTS_CACHE_PATH)
    floor = Decimal(0) if args.test else args.floor_usd
    print(f"reading live holders at floor ${floor} ...")
    holder_set = holders.read_holder_set(floor_usd=floor, exclusions=exclusions, accounts=accounts,
                                         test_address=config.TEST_ADDRESS if args.test else None)
    accounts.save(path=config.ACCOUNTS_CACHE_PATH)

    plan = planner.build_plan(holder_set=holder_set, floor_usd=floor, run_id=planner.next_run_id(events=events), canary=args.canary,
                              test=args.test, gas_per_transfer=planner.gas_per_transfer(events=events), max_addresses=args.max_addresses,
                              gas_budget=args.gas_budget, created_at=chainio.now().isoformat())
    planner.write_plan(plan=plan, state_dir=config.STATE_DIR)
    print(render_plan(plan=plan))
    batches = len(plan.pending_batches())
    print(f"RESULT: PLANNED — run {plan.run_id}, {batches} batches under {config.STATE_DIR}; next: `cli.py run`")
    return 0


def render_plan(plan: planner.Plan) -> str:
    lines = [f"run {plan.run_id} · floor ${plan.floor_usd} · height {plan.height} · gas/transfer {plan.gas_per_transfer}"]
    for tier in plan.tiers:
        addresses = sum(len(b.addresses) for b in tier.batches)
        usd = sum((b.usd for b in tier.batches), Decimal(0))
        lines.append(f"  {tier.name:8} {len(tier.batches):4} batches {addresses:6} addresses ${usd:,.2f}")
    lines.append(f"  excluded {len(plan.excluded)} (${sum((e.usd for e in plan.excluded), Decimal(0)):,.2f}):")
    lines.extend(f"    {e.address}  {e.reason}  ${e.usd:,.2f}" for e in plan.excluded)
    lines.append(f"  skipped by chain rules {len(plan.skipped)} (${sum((e.usd for e in plan.skipped), Decimal(0)):,.2f}):")
    lines.extend(f"    {e.address}  {e.reason}  ${e.usd:,.2f}" for e in plan.skipped[:20])
    if len(plan.skipped) > 20:
        lines.append(f"    ... {len(plan.skipped) - 20} more in plan.json")
    lines.append(f"  below floor {plan.below_floor_count} addresses ${plan.below_floor_usd:,.2f}")
    lines.append("  ladder (holders not yet swept at each floor):")
    lines.extend(f"    ${r.floor:>6}  {r.holders:6} holders  ${r.usd:,.2f}" for r in plan.ladder)
    return "\n".join(lines)


# ---- run


def cmd_run(args: argparse.Namespace) -> int:
    global _answered_all
    _answered_all = False  # an `a` answer lasts one run, not the process
    plan = planner.load_plan(path=config.PLAN_PATH)
    if plan is None:
        raise RunStopped("no plan; run `plan` first")
    events = ledger.read(path=config.LEDGER_PATH)
    exclusions = holders.load_exclusions(path=config.EXCLUSIONS_PATH)

    checks = preflight(plan=plan, events=events, exclusions=exclusions, now=chainio.now())
    for check in checks:
        print(f"  [{'ok' if check.ok else 'FAIL'}] {check.name}: {check.detail}")
    if not all(check.ok for check in checks):
        raise RunStopped("preflight failed")

    states = ledger.batch_states(events=events)
    plan_tiers = [tier.name for tier in plan.tiers]
    if args.tier is not None and args.tier not in plan_tiers:
        raise RunStopped(f"no tier {args.tier} in the plan; the plan has: {', '.join(plan_tiers)}")
    todo = [(tier, batch) for tier, batch in plan.pending_batches()
            if batch.id not in states and (args.tier is None or tier.name == args.tier)]
    if args.batches is not None:
        todo = todo[: args.batches]
    print(f"{len(todo)} batch(es) to run")

    yes = args.yes
    done = 0
    for tier, batch in todo:
        event = run_batch(plan=plan, tier=tier, batch=batch, state_dir=config.STATE_DIR, ledger_path=config.LEDGER_PATH, yes=yes, dry_run=args.dry_run)
        if event is None and not args.dry_run:
            print("declined; stopping")
            break
        if event is not None or args.dry_run:
            done += 1
            yes = yes or _answered_all
            if event is not None and event.skipped and not args.continue_on_skip:
                raise RunStopped(f"batch {batch.id} had {len(event.skipped)} skip event(s): the planner and the chain disagree; "
                                 "inspect the ledger, fix the rule, re-plan (or pass --continue-on-skip)")
    print(f"RESULT: DONE — {done} batch(es) {'simulated' if args.dry_run else 'confirmed'}")
    return 0


_answered_all = False  # `a` at the prompt: yes to every later batch of this run


def ask(prompt: str) -> str:
    return input(prompt).strip().lower()


def run_batch(
    plan: planner.Plan,
    tier: planner.Tier,
    batch: planner.Batch,
    state_dir: pathlib.Path,
    ledger_path: pathlib.Path,
    yes: bool,
    dry_run: bool,
) -> ledger.Event | None:
    """One batch: describe, confirm, simulate, broadcast, poll, record. Returns the terminal event (or the submitted
    one when the poll gave up), None on a dry run or a declined prompt. Raises RunStopped when the run must not go on."""
    global _answered_all
    denoms = [d.denom for d in plan.denoms]
    file = state_dir / batch.file
    print(describe_batch(tier=tier, batch=batch))

    if not yes and not _answered_all and not dry_run:
        answer = ask(f"sweep batch {batch.id}? [y/N/a] ")
        if answer == ANSWER_ALL:
            _answered_all = True
        elif answer != ANSWER_YES:
            return None

    # Simulate first: a batch over the block limit must never be broadcast
    simulation = chainio.strided(args=command_line(denoms=denoms, file=file, gas=None, dry_run=True))
    estimate = parse_gas_estimate(text=simulation.stderr + simulation.stdout)
    if estimate is None:
        raise RunStopped(f"batch {batch.id}: simulation gave no gas estimate:\n{simulation.stderr}{simulation.stdout}")
    gas = int(Decimal(estimate) * config.GAS_ADJUSTMENT)
    print(f"  gas estimate {estimate:,} → wanted {gas:,} (planned {batch.estimated_gas:,})")
    if gas > config.BLOCK_GAS_LIMIT:
        raise RunStopped(f"batch {batch.id}: {gas:,} gas exceeds the block limit {config.BLOCK_GAS_LIMIT:,}; re-plan with a smaller --gas-budget")
    if dry_run:
        print("  dry run: " + " ".join([config.STRIDED_BINARY, *command_line(denoms=denoms, file=file, gas=gas, dry_run=False)]))
        return None

    broadcast = chainio.strided(args=command_line(denoms=denoms, file=file, gas=gas, dry_run=False))
    if broadcast.returncode != 0:
        raise RunStopped(f"batch {batch.id}: strided exited {broadcast.returncode}:\n{broadcast.stderr}")
    result = parse_broadcast(stdout=broadcast.stdout)
    if result.code != 0:
        event = ledger.failed(batch_id=batch.id, tx_hash=result.tx_hash, at=chainio.now().isoformat(), code=result.code,
                              codespace=result.codespace, raw_log=result.raw_log, run_id=plan.run_id)
        ledger.append(event=event, path=ledger_path)
        raise RunStopped(f"batch {batch.id}: broadcast rejected (code {result.code} {result.codespace}): {result.raw_log}")

    ledger.append(event=ledger.submitted(run_id=plan.run_id, batch_id=batch.id, tx_hash=result.tx_hash, at=chainio.now().isoformat(),
                                         addresses=len(batch.addresses), transfers_estimate=batch.transfers, gas_wanted=gas), path=ledger_path)
    print(f"  submitted {result.tx_hash}; waiting for inclusion")
    outcome = poll_tx(tx_hash=result.tx_hash, wait_seconds=config.TX_WAIT_SECONDS)
    if outcome is None:
        raise RunStopped(f"batch {batch.id}: {result.tx_hash} not found after {config.TX_WAIT_SECONDS}s; run `resolve`")
    event = record_outcome(batch_id=batch.id, tx_hash=result.tx_hash, outcome=outcome, ledger_path=ledger_path)
    if event.kind == ledger.Kind.FAILED:
        raise RunStopped(f"batch {batch.id}: tx failed (code {outcome.code} {outcome.codespace}): {outcome.raw_log}")
    return event


def record_outcome(batch_id: str, tx_hash: str, outcome: TxOutcome, ledger_path: pathlib.Path) -> ledger.Event:
    at = chainio.now().isoformat()
    if outcome.code != 0:
        event = ledger.failed(batch_id=batch_id, tx_hash=tx_hash, at=at, code=outcome.code, codespace=outcome.codespace, raw_log=outcome.raw_log)
    else:
        event = ledger.confirmed(batch_id=batch_id, tx_hash=tx_hash, at=at, height=outcome.height, gas_used=outcome.gas_used,
                                 transfers=outcome.transfers, skipped=outcome.skipped)
        print(f"  confirmed at {outcome.height}: {len(outcome.transfers)} transfers, {len(outcome.skipped)} skipped, gas used {outcome.gas_used:,}")
        for skip in outcome.skipped:
            print(f"    SKIPPED {skip.address}: {skip.reason}")
    ledger.append(event=event, path=ledger_path)
    return event


def describe_batch(tier: planner.Tier, batch: planner.Batch) -> str:
    top = ", ".join(f"{p.address[:14]}… ${p.usd:,.2f}" for p in batch.addresses[:3])
    return (f"batch {batch.id} · tier {tier.name} · {len(batch.addresses)} addresses · {batch.transfers} transfers · "
            f"${batch.usd:,.2f} · largest: {top}")


def command_line(denoms: list[str], file: pathlib.Path, gas: int | None, dry_run: bool) -> list[str]:
    """The strided arguments for one batch. The test suite pins this shape."""
    line = ["tx", "stakeibc", "sweep-tokens-off-stride", ",".join(denoms), str(file),
            "--from", config.SWEEP_OPERATOR_KEY, "--keyring-backend", config.KEYRING_BACKEND,
            "--chain-id", config.CHAIN_ID, "--node", config.RPC]
    if dry_run:
        return line + ["--gas-prices", f"{config.GAS_PRICE_USTRD}{config.FEE_DENOM}", "--dry-run"]
    return line + ["--gas", str(gas), "--gas-prices", f"{config.GAS_PRICE_USTRD}{config.FEE_DENOM}",
                   "--broadcast-mode", "sync", "-y", "--output", "json"]


def parse_gas_estimate(text: str) -> int | None:
    match = GAS_ESTIMATE_PATTERN.search(text)
    return int(match.group(1)) if match else None


def parse_broadcast(stdout: str) -> BroadcastResult:
    try:
        body = json.loads(stdout)
    except json.JSONDecodeError:
        raise RunStopped(f"broadcast output is not JSON:\n{stdout}") from None
    if not isinstance(body, dict) or not body.get("txhash"):
        raise RunStopped(f"broadcast output has no txhash:\n{stdout}")
    return BroadcastResult(code=int(body.get("code", 0)), tx_hash=body["txhash"], codespace=body.get("codespace", ""), raw_log=body.get("raw_log", ""))


def poll_tx(tx_hash: str, wait_seconds: int) -> TxOutcome | None:
    waited = 0
    while True:
        try:
            return parse_tx_response(body=chainio.rest_get(path=f"/cosmos/tx/v1beta1/txs/{tx_hash}"))
        except chainio.NotFound:
            pass
        if waited >= wait_seconds:
            return None
        chainio.sleep(config.TX_POLL_SECONDS)
        waited += config.TX_POLL_SECONDS


def parse_tx_response(body: dict) -> TxOutcome:
    response = body["tx_response"]
    transfers: list[ledger.Transfer] = []
    skipped: list[ledger.Skip] = []
    for event in response.get("events", []):
        attributes = {a["key"]: a["value"] for a in event.get("attributes", [])}
        if event["type"] == EVENT_SWEEP_TRANSFER:
            transfers.append(ledger.Transfer(address=attributes["address"], denom=attributes["denom"], amount=int(attributes["amount"]),
                                             channel=attributes["channel"], receiver=attributes["receiver"]))
        elif event["type"] == EVENT_SWEEP_SKIPPED:
            skipped.append(ledger.Skip(address=attributes["address"], reason=attributes["reason"]))
    return TxOutcome(code=int(response.get("code", 0)), height=int(response["height"]), gas_used=int(response.get("gas_used", 0)),
                     codespace=response.get("codespace", ""), raw_log=response.get("raw_log", ""), transfers=transfers, skipped=skipped)


# ---- preflight


def preflight(plan: planner.Plan, events: list[ledger.Event], exclusions: dict[str, holders.Exclusion], now: datetime.datetime) -> list[Check]:
    states = ledger.batch_states(events=events)
    pending = [batch for _, batch in plan.pending_batches() if batch.id not in states]
    return [
        _check_chain_id(),
        _check_binary_version(),
        _check_operator_key(),
        _check_plan_age(plan=plan, now=now),
        _check_batch_files(plan=plan),
        _check_unresolved(events=events),
        _check_exclusions(pending=pending, exclusions=exclusions),
        _check_channels(plan=plan),
        _check_fee_balance(pending=pending),
    ]


def _check_chain_id() -> Check:
    network = chainio.rest_get(path="/cosmos/base/tendermint/v1beta1/node_info")["default_node_info"]["network"]
    return Check(name="chain id", ok=network == config.CHAIN_ID, detail=f"node reports {network}")


def _check_binary_version() -> Check:
    output = chainio.strided(args=["version"]).stdout.strip()
    version = output.splitlines()[-1] if output else ""
    return Check(name="binary version", ok=version.startswith(config.BINARY_VERSION_PREFIX), detail=f"strided version {version or '?'}")


def _check_operator_key() -> Check:
    shown = chainio.strided(args=["keys", "show", config.SWEEP_OPERATOR_KEY, "--keyring-backend", config.KEYRING_BACKEND, "-a"]).stdout.strip()
    return Check(name="operator key", ok=shown == config.SWEEP_OPERATOR, detail=f"{config.SWEEP_OPERATOR_KEY} is {shown or 'missing'}")


def _check_plan_age(plan: planner.Plan, now: datetime.datetime) -> Check:
    age = (now - datetime.datetime.fromisoformat(plan.created_at)).total_seconds()
    return Check(name="plan age", ok=0 <= age <= config.MAX_PLAN_AGE_SECONDS, detail=f"run {plan.run_id} planned {age / 60:.0f} min ago at height {plan.height}")


def _check_batch_files(plan: planner.Plan) -> Check:
    bad = [batch.id for _, batch in plan.pending_batches()
           if not (config.STATE_DIR / batch.file).exists() or planner.batch_sha256(content=(config.STATE_DIR / batch.file).read_text()) != batch.sha256]
    return Check(name="batch files", ok=not bad, detail="every file matches the plan" if not bad else f"changed or missing: {', '.join(bad)}")


def _check_unresolved(events: list[ledger.Event]) -> Check:
    pending = ledger.unresolved(events=events)
    return Check(name="unresolved submissions", ok=not pending, detail="none" if not pending else f"run `resolve`: {', '.join(e.batch_id for e in pending)}")


def _check_exclusions(pending: list[planner.Batch], exclusions: dict[str, holders.Exclusion]) -> Check:
    forbidden = set(exclusions) | set(config.PROTOCOL_ADDRESSES) | {config.SWEEP_OPERATOR}
    hits = [f"{p.address} ({batch.id})" for batch in pending for p in batch.addresses if p.address in forbidden]
    return Check(name="exclusions and protocol addresses", ok=not hits, detail="no pending batch names one" if not hits else f"in a batch: {', '.join(hits)}")


def _check_channels(plan: planner.Plan) -> Check:
    closed = []
    for channel in sorted({d.channel for d in plan.denoms}):
        state = chainio.rest_get(path=f"/ibc/core/channel/v1/channels/{channel}/ports/{config.TRANSFER_PORT}")["channel"]["state"]
        if state != STATE_OPEN:
            closed.append(f"{channel} {state}")
    return Check(name="channels open", ok=not closed, detail="every destination channel is OPEN" if not closed else ", ".join(closed))


def _check_fee_balance(pending: list[planner.Batch]) -> Check:
    body = chainio.rest_get(path=f"/cosmos/bank/v1beta1/balances/{config.SWEEP_OPERATOR}/by_denom", params={"denom": config.FEE_DENOM})
    balance = int(body["balance"]["amount"])
    needed = int(sum(Decimal(b.estimated_gas) * config.GAS_ADJUSTMENT * config.GAS_PRICE_USTRD for b in pending))
    return Check(name="operator fee balance", ok=balance >= needed, detail=f"{balance / 1e6:.2f} STRD held, {needed / 1e6:.2f} STRD needed for {len(pending)} batches")


# ---- status / resolve


def cmd_status(args: argparse.Namespace) -> int:
    plan = planner.load_plan(path=config.PLAN_PATH)
    events = ledger.read(path=config.LEDGER_PATH)
    if plan is None:
        print("no plan")
        return 0
    states = ledger.batch_states(events=events)
    print(f"run {plan.run_id} · floor ${plan.floor_usd} · planned {plan.created_at} at height {plan.height}")
    for tier in plan.tiers:
        counts: dict[str, int] = {}
        for batch in tier.batches:
            state = states.get(batch.id, ledger.BatchState.PENDING)
            counts[state] = counts.get(state, 0) + 1
        print(f"  {tier.name:8} " + ", ".join(f"{n} {s}" for s, n in counts.items()))
    confirmed = [e for e in events if e.kind == ledger.Kind.CONFIRMED]
    transfers = sum(len(e.transfers) for e in confirmed)
    skipped = [(e.batch_id, s) for e in confirmed for s in e.skipped]
    print(f"  confirmed {len(confirmed)} batches, {transfers} transfers, {len(skipped)} skips")
    for batch_id, skip in skipped:
        print(f"    SKIP {batch_id} {skip.address}: {skip.reason}")
    for event in events:
        if event.kind == ledger.Kind.FAILED:
            print(f"    FAILED {event.batch_id} {event.tx_hash} code {event.code}: {event.raw_log}")
    for event in ledger.unresolved(events=events):
        print(f"    UNRESOLVED {event.batch_id} {event.tx_hash} submitted {event.at}")
    return 0


def cmd_resolve(args: argparse.Namespace) -> int:
    events = ledger.read(path=config.LEDGER_PATH)
    pending = ledger.unresolved(events=events)
    if not pending:
        print("nothing to resolve")
        return 0
    for event in pending:
        assert event.tx_hash is not None
        submitted_at = datetime.datetime.fromisoformat(event.at)
        remaining = max(0, args.wait_seconds - int((chainio.now() - submitted_at).total_seconds()))
        outcome = poll_tx(tx_hash=event.tx_hash, wait_seconds=remaining)
        if outcome is None:
            ledger.append(event=ledger.lost(batch_id=event.batch_id, tx_hash=event.tx_hash, at=chainio.now().isoformat()),
                          path=config.LEDGER_PATH)
            print(f"  {event.batch_id} {event.tx_hash}: not found after {args.wait_seconds}s, recorded as lost; the next `plan` reads the balances")
            continue
        recorded = record_outcome(batch_id=event.batch_id, tx_hash=event.tx_hash, outcome=outcome, ledger_path=config.LEDGER_PATH)
        print(f"  {event.batch_id} {event.tx_hash}: {recorded.kind}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
