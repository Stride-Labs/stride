"""The Sweep tab: `collect()` is the expensive live read (every holder of every sweep denom, in bulk) and is the tab's
collector, refreshed on demand; `body()` composes it with the plan, the ledger and the exclusions file read from disk
on every request, through the pure `compose()`. Every integer in the payload is a string."""

import pathlib
import sys
from decimal import ROUND_HALF_UP, Decimal
from typing import Any

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package lives beside the dashboard

import config  # noqa: E402
from sweep import chainio as sweep_chainio  # noqa: E402
from sweep import config as sweep_config  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402

CENTS = Decimal("0.01")
DEFERRED_TOP = 25
STATE_SWEPT = "swept"
STATE_REFUNDED = "refunded"
STATE_REMAINING = "remaining"


# ---- collector


def collect() -> dict[str, Any]:
    """The live read. `balances_at` is stamped before the balances are read: the snapshot's `fetched_at` is stamped
    after `collect` returns, so a batch confirmed during the read would look older than the snapshot though the
    balances miss it."""
    denoms = holders.resolve_denoms()
    balances_at = sweep_chainio.now().isoformat()
    balances = holders.read_balances(denoms=denoms, workers=config.SWEEP_BALANCE_WORKERS)
    operator = sweep_chainio.rest_get(
        path=f"/cosmos/bank/v1beta1/balances/{sweep_config.SWEEP_OPERATOR}/by_denom",
        params={"denom": sweep_config.FEE_DENOM},
    )
    return {
        "balances_at": balances_at,
        "height": str(sweep_chainio.latest_height()),
        "denoms": [
            {
                "denom": denom.denom,
                "symbol": denom.symbol,
                "decimals": str(denom.decimals),
                "price_usd": str(denom.price_usd),
                "destination": denom.destination,
                "channel": denom.channel,
            }
            for denom in denoms
        ],
        "balances": {
            address: {denom: str(amount) for denom, amount in coins.items()} for address, coins in balances.items()
        },
        "operator_strd": operator["balance"]["amount"],
        "staked_ustrd": str(holders.read_staking_pool()),
    }


# ---- route body


def body(view: dict[str, Any]) -> dict[str, Any]:
    """`GET /api/sweep`: plan, ledger and exclusions fresh from disk; the live part from the collector's snapshot."""
    live = None if view.get("loading") else view["data"]
    fetched_at = view.get("fetched_at")
    try:
        plan = planner.load_plan()
        events = ledger.read()
        exclusions = holders.load_exclusions()
    except (planner.PlanError, ledger.LedgerError, holders.ExclusionsError) as error:
        return {
            "fetched_at": fetched_at,
            "refreshing": view.get("refreshing", False),
            "error": str(error),
            "data": None,
        }
    return {
        "fetched_at": fetched_at,
        "refreshing": view.get("refreshing", False),
        "last_error": view.get("last_error"),
        "data": compose(plan=plan, events=events, live=live, exclusions=exclusions, live_fetched_at=fetched_at),
    }


# ---- composition


def compose(
    plan: planner.Plan | None,
    events: list[ledger.Event],
    live: dict[str, Any] | None,
    exclusions: dict[str, holders.Exclusion],
    live_fetched_at: str | None,
) -> dict[str, Any]:
    if plan is None:
        return {
            "run": None,
            "progress": None,
            "totals": None,
            "summary": _summary(plan=None, events=events, live=live, address_state=None),
            "by_denom": None,
            "batches": [],
            "ladder": None,
            "deferred": None,
            "refunded": [],
            "exclusions": [],
            "keyless": None,
            "live_fetched_at": live_fetched_at,
        }

    states = ledger.batch_states(events=events)
    transferred = ledger.confirmed_transfers(events=events)
    balances = _live_balances(live=live)
    # The balances were read at `balances_at`, before the snapshot was stamped; older snapshots only have the stamp
    snapshot_at = (live.get("balances_at") or live_fetched_at) if live is not None else live_fetched_at
    by_denom = {denom.denom: denom for denom in plan.denoms}
    address_state = (
        _address_states(
            plan=plan, events=events, transferred=transferred, balances=balances, snapshot_at=snapshot_at
        )
        if live is not None
        else None
    )
    refunds = (
        _refunds(plan=plan, events=events, balances=balances, snapshot_at=snapshot_at)
        if live is not None
        else []
    )

    return {
        "run": _run(plan=plan, states=states, live=live),
        "progress": _progress(plan=plan, states=states),
        "totals": _totals(
            plan=plan, address_state=address_state, transferred=transferred, balances=balances, by_denom=by_denom
        )
        if address_state is not None
        else None,
        "summary": _summary(plan=plan, events=events, live=live, address_state=address_state, refunds=refunds),
        "by_denom": _by_denom(plan=plan, address_state=address_state, transferred=transferred, balances=balances)
        if address_state is not None
        else None,
        "batches": _batches(plan=plan, events=events, states=states, refunds=refunds),
        "deferred": _deferred(plan=plan, live=live, balances=balances, by_denom=by_denom),
        "ladder": _ladder(plan=plan, address_state=address_state, balances=balances, by_denom=by_denom)
        if address_state is not None
        else None,
        "refunded": [
            {
                "address": transfer.address,
                "denom": transfer.denom,
                "amount": str(transfer.amount),
                "channel": transfer.channel,
                "batch_id": batch_id,
            }
            for transfer, batch_id in refunds
        ],
        "exclusions": [
            {
                "section": exclusion.section,
                "address": exclusion.address,
                "label": exclusion.label,
                "reason": exclusion.reason,
                "live_usd": _money(holders.usd_value(balances=balances.get(exclusion.address, {}), denoms=by_denom))
                if live is not None
                else None,
            }
            for exclusion in exclusions.values()
        ],
        "keyless": _keyless(plan=plan, address_state=address_state),
        "live_fetched_at": live_fetched_at,
    }


def _live_balances(live: dict[str, Any] | None) -> dict[str, dict[str, int]]:
    if live is None:
        return {}
    return {
        address: {denom: int(amount) for denom, amount in coins.items()} for address, coins in live["balances"].items()
    }


def _planned(plan: planner.Plan) -> dict[str, planner.PlannedAddress]:
    return {planned.address: planned for _, batch in plan.pending_batches() for planned in batch.addresses}


def _locked(plan: planner.Plan) -> dict[str, dict[str, int]]:
    """A vesting account keeps its locked part after a full sweep, so that much live balance is not a refund. The plan
    records it for every vesting holder, planned or below the floor, so an address swept in an earlier run and absent
    from this run's batches still has its locked part excluded from refunds."""
    return plan.locked


def _back(address: str, denom: str, balances: dict[str, dict[str, int]], locked: dict[str, dict[str, int]]) -> int:
    """What the address holds of `denom` beyond its locked remainder: the refunded amount, 0 if only the remainder."""
    return max(0, balances.get(address, {}).get(denom, 0) - locked.get(address, {}).get(denom, 0))


def _known_addresses(plan: planner.Plan, transferred: dict[str, list[ledger.Transfer]]) -> set[str]:
    """Addresses already accounted for: planned, deferred, swept in any run, excluded, or skipped by the chain-side
    filters."""
    return (
        set(_planned(plan=plan))
        | {deferred.address for deferred in plan.deferred}
        | set(transferred)
        | {excluded.address for excluded in plan.excluded}
        | {skipped.address for skipped in plan.skipped}
    )


def _judgeable_transfers(
    events: list[ledger.Event], snapshot_at: str | None
) -> list[tuple[ledger.Transfer, str]]:
    """Confirmed transfers (with their batch id) that the live snapshot can judge: those confirmed no later than the
    snapshot's balances were read (`snapshot_at`). A newer one is not in the snapshot yet, so a balance it shows says
    nothing about a refund."""
    snapshot_time = holders.parse_chain_time(text=snapshot_at) if snapshot_at is not None else None
    return [
        (transfer, event.batch_id)
        for event in events
        if event.kind == ledger.Kind.CONFIRMED
        and (snapshot_time is None or holders.parse_chain_time(text=event.at) <= snapshot_time)
        for transfer in event.transfers
    ]


def _address_states(
    plan: planner.Plan,
    events: list[ledger.Event],
    transferred: dict[str, list[ledger.Transfer]],
    balances: dict[str, dict[str, int]],
    snapshot_at: str | None,
) -> dict[str, str]:
    """swept: transferred and holding none of it now; refunded: transferred and holding a transferred denom again;
    remaining: in the current plan with no confirmed transfer. Swept and refunded cover every address with a confirmed
    transfer in any run, so a re-plan at a lower floor does not drop the earlier runs' holders. A transfer confirmed
    after `snapshot_at` is swept whatever the snapshot shows (it predates the transfer, so it cannot show a refund
    of it); a refund needs balances read after the transfer."""
    locked = _locked(plan=plan)
    judged_denoms: dict[str, set[str]] = {}
    for transfer, _ in _judgeable_transfers(events=events, snapshot_at=snapshot_at):
        judged_denoms.setdefault(transfer.address, set()).add(transfer.denom)

    states = {address: STATE_REMAINING for address in _planned(plan=plan) if address not in transferred}
    for address in transferred:
        moved = judged_denoms.get(address, set())
        states[address] = (
            STATE_REFUNDED
            if any(_back(address=address, denom=denom, balances=balances, locked=locked) > 0 for denom in moved)
            else STATE_SWEPT
        )
    return states


def _refunds(
    plan: planner.Plan,
    events: list[ledger.Event],
    balances: dict[str, dict[str, int]],
    snapshot_at: str | None,
) -> list[tuple[ledger.Transfer, str]]:
    """Confirmed transfers whose denom is back beyond the locked remainder, judged only by a snapshot taken after
    them."""
    locked = _locked(plan=plan)
    return [
        (transfer, batch_id)
        for transfer, batch_id in _judgeable_transfers(events=events, snapshot_at=snapshot_at)
        if _back(address=transfer.address, denom=transfer.denom, balances=balances, locked=locked) > 0
    ]


def _progress(plan: planner.Plan, states: dict[str, ledger.BatchState]) -> dict[str, Any]:
    """How far the run is: batches, holders and USD done (a batch whose latest ledger event is `confirmed`) of the
    plan's total, overall and per tier in plan order."""
    tiers = [_tier_progress(tier=tier, states=states) for tier in plan.tiers]
    return {
        "run_id": str(plan.run_id),
        "floor_usd": str(plan.floor_usd),
        "test": plan.test,
        "tiers": [_progress_row(row=row, name=row["name"]) for row in tiers],
        **_progress_row(row=_sum_rows(rows=tiers)),
    }


def _tier_progress(tier: planner.Tier, states: dict[str, ledger.BatchState]) -> dict[str, Any]:
    done = [batch for batch in tier.batches if states.get(batch.id) == ledger.BatchState.CONFIRMED]
    return {
        "name": tier.name.value,
        "batches": len(tier.batches),
        "batches_confirmed": len(done),
        "addresses": sum(len(batch.addresses) for batch in tier.batches),
        "addresses_confirmed": sum(len(batch.addresses) for batch in done),
        "usd": sum((batch.usd for batch in tier.batches), Decimal(0)),
        "usd_confirmed": sum((batch.usd for batch in done), Decimal(0)),
    }


def _sum_rows(rows: list[dict[str, Any]]) -> dict[str, Any]:
    keys = ("batches", "batches_confirmed", "addresses", "addresses_confirmed", "usd", "usd_confirmed")
    return {key: sum((row[key] for row in rows), Decimal(0) if key.startswith("usd") else 0) for key in keys}


def _progress_row(row: dict[str, Any], name: str | None = None) -> dict[str, Any]:
    """The row as the payload carries it: counts as strings, USD as two-decimal strings."""
    out = {key: _money(value) if key.startswith("usd") else str(value) for key, value in row.items() if key != "name"}
    return {"name": name, **out} if name is not None else out


def _run(plan: planner.Plan, states: dict[str, ledger.BatchState], live: dict[str, Any] | None) -> dict[str, Any]:
    batches = [batch for _, batch in plan.pending_batches()]
    counts = {state.value: 0 for state in ledger.BatchState}
    for batch in batches:
        counts[states.get(batch.id, ledger.BatchState.PENDING).value] += 1
    pending_gas = sum(batch.estimated_gas for batch in batches if batch.id not in states)
    fee = Decimal(pending_gas) * sweep_config.GAS_ADJUSTMENT * sweep_config.GAS_PRICE_USTRD
    return {
        "run_id": str(plan.run_id),
        "floor_usd": str(plan.floor_usd),
        "created_at": plan.created_at,
        "height": str(plan.height),
        "test": plan.test,
        "skipped_denoms": list(plan.skipped_denoms),
        "batches": {
            "total": str(len(batches)),
            **{state: str(count) for state, count in counts.items() if state != "pending"},
            "pending": str(counts["pending"]),
        },
        "operator_strd": live["operator_strd"] if live is not None else None,
        "fee_estimate_ustrd": str(int(fee)),
    }


def _swept_usd(
    addresses: list[str], transferred: dict[str, list[ledger.Transfer]], by_denom: dict[str, holders.SweepDenom]
) -> Decimal:
    """What the addresses swept, each (address, denom) valued once at its latest confirmed transfer: an address swept,
    refunded and swept again moved the same funds twice, so summing every transfer would count them double."""
    latest = {(address, transfer.denom): transfer for address in addresses for transfer in transferred[address]}
    return sum((_transfer_usd(transfer=transfer, by_denom=by_denom) for transfer in latest.values()), Decimal(0))


def _totals(
    plan: planner.Plan,
    address_state: dict[str, str],
    transferred: dict[str, list[ledger.Transfer]],
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> dict[str, dict[str, str]]:
    swept = [address for address, state in address_state.items() if state == STATE_SWEPT]
    refunded = [address for address, state in address_state.items() if state == STATE_REFUNDED]
    remaining = [address for address, state in address_state.items() if state == STATE_REMAINING]

    swept_usd = _swept_usd(addresses=swept, transferred=transferred, by_denom=by_denom)
    locked = _locked(plan=plan)
    refunded_usd = sum(
        (
            _refunded_usd(address=address, transferred=transferred, balances=balances, locked=locked, by_denom=by_denom)
            for address in refunded
        ),
        Decimal(0),
    )
    remaining_usd = sum(
        (holders.usd_value(balances=balances.get(address, {}), denoms=by_denom) for address in remaining), Decimal(0)
    )
    excluded_usd = sum(
        (holders.usd_value(balances=balances.get(excluded.address, {}), denoms=by_denom) for excluded in plan.excluded),
        Decimal(0),
    )
    deferred_usd = _live_deferred_usd(plan=plan, balances=balances, by_denom=by_denom)

    known = _known_addresses(plan=plan, transferred=transferred)
    others = {
        address: holders.usd_value(balances=coins, denoms=by_denom)
        for address, coins in balances.items()
        if address not in known
    }
    below = {address: usd for address, usd in others.items() if usd < plan.floor_usd}
    unplanned = {address: usd for address, usd in others.items() if usd >= plan.floor_usd}

    return {
        "swept": _bucket(count=len(swept), usd=swept_usd),
        "refunded": _bucket(count=len(refunded), usd=refunded_usd),
        "remaining": _bucket(count=len(remaining), usd=remaining_usd),
        "deferred": _bucket(count=len(plan.deferred), usd=deferred_usd),
        "excluded": _bucket(count=len(plan.excluded), usd=excluded_usd),
        "below_floor": _bucket(count=len(below), usd=sum(below.values(), Decimal(0))),
        "unplanned": _bucket(count=len(unplanned), usd=sum(unplanned.values(), Decimal(0))),
    }


def _summary(
    plan: planner.Plan | None,
    events: list[ledger.Event],
    live: dict[str, Any] | None,
    address_state: dict[str, str] | None,
    refunds: list[tuple[ledger.Transfer, str]] | None = None,
) -> dict[str, Any]:
    """The Sweep tab's summary area: `total` is cumulative and chain-wide, `plan` is the current plan.json alone and
    None without a plan. Each figure is {addresses, usd}."""
    transferred = ledger.confirmed_transfers(events=events)
    if plan is None:
        # Without a plan only the ledger is known (no exclusions or skips to judge remaining by), priced by the snapshot
        return {"total": _ledger_only_total(transferred=transferred, live=live), "plan": None}
    balances = _live_balances(live=live)
    by_denom = {denom.denom: denom for denom in plan.denoms}
    # Without a live snapshot nothing can be judged refunded, so every transferred address counts as swept
    states = (
        address_state
        if address_state is not None
        else _address_states(plan=plan, events=events, transferred=transferred, balances={}, snapshot_at=None)
    )
    return {
        "total": _total_summary(
            plan=plan, transferred=transferred, address_state=states, live=live, balances=balances, by_denom=by_denom
        ),
        "plan": _plan_summary(
            plan=plan, events=events, refunds=refunds or [], balances=balances, live=live, by_denom=by_denom
        ),
    }


def _total_summary(
    plan: planner.Plan,
    transferred: dict[str, list[ledger.Transfer]],
    address_state: dict[str, str],
    live: dict[str, Any] | None,
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> dict[str, Any]:
    """TOTAL, over every run in the ledger and the whole chain.
    swept: every address with a confirmed transfer in any run that is not refunded now (the tiles' `totals.swept`).
    remaining: every sweepable holder not yet swept, in the live snapshot at or above `config.SWEEP_TOTAL_FLOOR_USD`:
    not swept, not excluded, not skipped by the chain rules and not deferred; None without a live snapshot.
    deferred: the current plan's deferred list (the only source), valued live when a snapshot exists."""
    swept = [address for address, state in address_state.items() if state == STATE_SWEPT]
    swept_usd = _swept_usd(addresses=swept, transferred=transferred, by_denom=by_denom)
    deferred_addresses = {deferred.address for deferred in plan.deferred}
    remaining = (
        {
            address: usd
            for address, usd in _unswept_values(
                plan=plan, address_state=address_state, balances=balances, by_denom=by_denom
            ).items()
            if address not in deferred_addresses and usd >= config.SWEEP_TOTAL_FLOOR_USD
        }
        if live is not None
        else None
    )
    return {
        "swept": _bucket(count=len(swept), usd=swept_usd),
        "remaining": _bucket(count=len(remaining), usd=sum(remaining.values(), Decimal(0)))
        if remaining is not None
        else None,
        "deferred": _deferred_bucket(
            plan=plan, live=live, balances=balances, by_denom=by_denom, already_swept=set(swept)
        ),
    }


def _ledger_only_total(transferred: dict[str, list[ledger.Transfer]], live: dict[str, Any] | None) -> dict[str, Any]:
    """TOTAL without a plan: every address with a confirmed transfer, valued at the live snapshot's prices when there
    is one (None otherwise); remaining is None because the exclusions and chain skips live in the plan."""
    prices = {
        entry["denom"]: holders.SweepDenom(
            denom=entry["denom"],
            symbol=entry["symbol"],
            decimals=int(entry["decimals"]),
            price_usd=Decimal(entry["price_usd"]),
            destination=entry["destination"],
            channel=entry["channel"],
        )
        for entry in (live["denoms"] if live is not None else [])
    }
    # Without a snapshot there are no prices, so the USD is unknown rather than zero
    swept = (
        _bucket(count=len(transferred), usd=_swept_usd(addresses=list(transferred), transferred=transferred, by_denom=prices))
        if live is not None
        else {"addresses": str(len(transferred)), "usd": None}
    )
    return {
        "swept": swept,
        "remaining": None,
        "deferred": _bucket(count=0, usd=Decimal(0)),
    }


def _plan_summary(
    plan: planner.Plan,
    events: list[ledger.Event],
    refunds: list[tuple[ledger.Transfer, str]],
    balances: dict[str, dict[str, int]],
    live: dict[str, Any] | None,
    by_denom: dict[str, holders.SweepDenom],
) -> dict[str, dict[str, str]]:
    """CURRENT PLAN, this plan.json alone.
    swept: addresses in the plan's batches whose latest ledger event is `confirmed`, minus the refunded ones.
    remaining: addresses of the plan's batches that are not confirmed (the tiles' `totals.remaining`, valued live when
    a snapshot exists, else at the plan's USD).
    deferred / excluded: the plan's lists. refunded: addresses of the plan's batches with a refunded transfer."""
    states = ledger.batch_states(events=events)
    current = {batch.id for _, batch in plan.pending_batches()}
    confirmed = {batch.id for _, batch in plan.pending_batches() if states.get(batch.id) == ledger.BatchState.CONFIRMED}
    confirmed_transfers = [
        transfer
        for event in events
        if event.kind == ledger.Kind.CONFIRMED and event.batch_id in confirmed
        for transfer in event.transfers
    ]
    refunded_pairs = {(transfer.address, transfer.denom) for transfer, batch_id in refunds if batch_id in current}
    refunded_addresses = {address for address, _ in refunded_pairs}
    swept_addresses = {
        planned.address
        for _, batch in plan.pending_batches()
        if batch.id in confirmed
        for planned in batch.addresses
        if planned.address not in refunded_addresses
    }
    remaining = [
        planned
        for _, batch in plan.pending_batches()
        if batch.id not in confirmed
        for planned in batch.addresses
    ]
    locked = _locked(plan=plan)

    swept_usd = sum(
        (
            _transfer_usd(transfer=transfer, by_denom=by_denom)
            for transfer in confirmed_transfers
            if transfer.address in swept_addresses
        ),
        Decimal(0),
    )
    remaining_usd = sum(
        (
            holders.usd_value(balances=balances.get(planned.address, {}), denoms=by_denom)
            if live is not None
            else planned.usd
            for planned in remaining
        ),
        Decimal(0),
    )
    refunded_usd = holders.usd_value(
        balances=_sum_back(pairs=refunded_pairs, balances=balances, locked=locked), denoms=by_denom
    )
    excluded_usd = sum(
        (
            holders.usd_value(balances=balances.get(excluded.address, {}), denoms=by_denom)
            if live is not None
            else excluded.usd
            for excluded in plan.excluded
        ),
        Decimal(0),
    )
    return {
        "swept": _bucket(count=len(swept_addresses), usd=swept_usd),
        "remaining": _bucket(count=len(remaining), usd=remaining_usd),
        "deferred": _deferred_bucket(plan=plan, live=live, balances=balances, by_denom=by_denom, already_swept=set()),
        "excluded": _bucket(count=len(plan.excluded), usd=excluded_usd),
        "refunded": _bucket(count=len(refunded_addresses), usd=refunded_usd),
    }


def _sum_back(
    pairs: set[tuple[str, str]], balances: dict[str, dict[str, int]], locked: dict[str, dict[str, int]]
) -> dict[str, int]:
    """Per denom, what came back across the (address, denom) pairs, beyond each address's locked remainder."""
    back: dict[str, int] = {}
    for address, denom in pairs:
        back[denom] = back.get(denom, 0) + _back(address=address, denom=denom, balances=balances, locked=locked)
    return back


def _deferred_bucket(
    plan: planner.Plan,
    live: dict[str, Any] | None,
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
    already_swept: set[str],
) -> dict[str, str]:
    """The deferred addresses not already counted as swept (they would otherwise sit in two tiles at once)."""
    open_deferred = [deferred for deferred in plan.deferred if deferred.address not in already_swept]
    usd = sum(
        (
            holders.usd_value(balances=balances.get(deferred.address, {}), denoms=by_denom)
            if live is not None
            else deferred.usd
            for deferred in open_deferred
        ),
        Decimal(0),
    )
    return _bucket(count=len(open_deferred), usd=usd)


def _refunded_usd(
    address: str,
    transferred: dict[str, list[ledger.Transfer]],
    balances: dict[str, dict[str, int]],
    locked: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> Decimal:
    """What came back: the live balance of the transferred denoms beyond the locked remainder, at the plan's prices."""
    moved = {transfer.denom for transfer in transferred[address]}
    back = {denom: _back(address=address, denom=denom, balances=balances, locked=locked) for denom in moved}
    return holders.usd_value(balances=back, denoms=by_denom)


def _by_denom(
    plan: planner.Plan,
    address_state: dict[str, str],
    transferred: dict[str, list[ledger.Transfer]],
    balances: dict[str, dict[str, int]],
) -> list[dict[str, Any]]:
    by_denom = {denom.denom: denom for denom in plan.denoms}
    locked = _locked(plan=plan)
    known = _known_addresses(plan=plan, transferred=transferred)
    below_floor_holders = [
        coins
        for address, coins in balances.items()
        if address not in known and holders.usd_value(balances=coins, denoms=by_denom) < plan.floor_usd
    ]
    rows = []

    for denom in plan.denoms:
        swept_transfers = [
            transfer for transfers in transferred.values() for transfer in transfers if transfer.denom == denom.denom
        ]
        swept_amount = sum(transfer.amount for transfer in swept_transfers)
        remaining = {
            address: balances.get(address, {}).get(denom.denom, 0)
            for address, state in address_state.items()
            if state == STATE_REMAINING
        }
        remaining = {address: amount for address, amount in remaining.items() if amount > 0}
        remaining_amount = sum(remaining.values())
        refunded = [
            address
            for address, state in address_state.items()
            if state == STATE_REFUNDED
            and _back(address=address, denom=denom.denom, balances=balances, locked=locked) > 0
        ]
        excluded_usd = sum(
            (
                _amount_usd(amount=balances.get(excluded.address, {}).get(denom.denom, 0), denom=denom)
                for excluded in plan.excluded
            ),
            Decimal(0),
        )
        below_usd = sum(
            (_amount_usd(amount=coins.get(denom.denom, 0), denom=denom) for coins in below_floor_holders), Decimal(0)
        )
        deferred_usd = sum(
            (
                _amount_usd(amount=balances.get(deferred.address, {}).get(denom.denom, 0), denom=denom)
                for deferred in plan.deferred
            ),
            Decimal(0),
        )

        rows.append(
            {
                "denom": denom.denom,
                "symbol": denom.symbol,
                "decimals": str(denom.decimals),
                "destination": denom.destination,
                "channel": denom.channel,
                "swept": {
                    "addresses": str(len({transfer.address for transfer in swept_transfers})),
                    "amount": str(swept_amount),
                    "usd": _money(_amount_usd(amount=swept_amount, denom=denom)),
                },
                "remaining": {
                    "addresses": str(len(remaining)),
                    "amount": str(remaining_amount),
                    "usd": _money(_amount_usd(amount=remaining_amount, denom=denom)),
                },
                "refunded_addresses": str(len(refunded)),
                "excluded_usd": _money(excluded_usd),
                "below_floor_usd": _money(below_usd),
                "deferred_usd": _money(deferred_usd),
            }
        )
    return rows


def _batches(
    plan: planner.Plan,
    events: list[ledger.Event],
    states: dict[str, ledger.BatchState],
    refunds: list[tuple[ledger.Transfer, str]],
) -> list[dict[str, Any]]:
    last_event: dict[str, ledger.Event] = {}
    for event in events:
        last_event[event.batch_id] = event  # the last line for the batch is its current state
    submitted = {event.batch_id: event for event in events if event.kind == ledger.Kind.SUBMITTED}
    refunded_per_batch: dict[str, int] = {}
    for _, batch_id in refunds:
        refunded_per_batch[batch_id] = refunded_per_batch.get(batch_id, 0) + 1

    rows = []
    for tier, batch in plan.pending_batches():
        last = last_event.get(batch.id)
        rows.append(
            {
                "run_id": str(plan.run_id),
                "batch_id": batch.id,
                "tier": tier.name.value,
                "addresses": str(len(batch.addresses)),
                "usd": _money(batch.usd),
                "state": states.get(batch.id, ledger.BatchState.PENDING).value,
                "tx_hash": last.tx_hash if last else None,
                "at": last.at if last else None,
                "skipped": [
                    {"address": skip.address, "reason": skip.reason} for skip in (last.skipped if last else [])
                ],
                "refunded": str(refunded_per_batch.get(batch.id, 0)),
            }
        )

    # Batches of earlier runs, from the ledger alone (the plan file only holds the current run)
    current = {batch.id for _, batch in plan.pending_batches()}
    for batch_id, event in last_event.items():
        if batch_id in current:
            continue
        first = submitted.get(batch_id)
        rows.append(
            {
                "run_id": batch_id.split("-")[0].lstrip("0") or "0",
                "batch_id": batch_id,
                "tier": None,
                "addresses": str(first.addresses) if first and first.addresses is not None else None,
                "usd": None,
                "state": states[batch_id].value,
                "tx_hash": event.tx_hash,
                "at": event.at,
                "skipped": [{"address": skip.address, "reason": skip.reason} for skip in event.skipped],
                "refunded": str(refunded_per_batch.get(batch_id, 0)),
            }
        )
    rows.sort(key=lambda row: row["batch_id"])
    return rows


def _unswept_values(
    plan: planner.Plan,
    address_state: dict[str, str],
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> dict[str, Decimal]:
    """Live USD of every holder still to sweep: everyone with a balance except the swept, the excluded and the
    chain-skipped. Deferred holders are included."""
    out = {address for address, state in address_state.items() if state == STATE_SWEPT}
    out |= {excluded.address for excluded in plan.excluded} | {skipped.address for skipped in plan.skipped}
    return {
        address: holders.usd_value(balances=coins, denoms=by_denom)
        for address, coins in balances.items()
        if address not in out
    }


def _ladder(
    plan: planner.Plan,
    address_state: dict[str, str],
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> list[dict[str, str]]:
    """Live holders still to sweep: everyone with a balance except the swept, the excluded and the chain-skipped."""
    values = list(_unswept_values(plan=plan, address_state=address_state, balances=balances, by_denom=by_denom).values())
    floors = [plan.floor_usd] + [floor for floor in sweep_config.LADDER_FLOORS if floor < plan.floor_usd]
    return [
        {
            "floor": str(floor),
            "holders": str(sum(1 for value in values if value >= floor)),
            "usd": _money(sum((value for value in values if value >= floor), Decimal(0))),
        }
        for floor in floors
    ]


def _live_deferred_usd(
    plan: planner.Plan, balances: dict[str, dict[str, int]], by_denom: dict[str, holders.SweepDenom]
) -> Decimal:
    return sum(
        (holders.usd_value(balances=balances.get(deferred.address, {}), denoms=by_denom) for deferred in plan.deferred),
        Decimal(0),
    )


def _deferred(
    plan: planner.Plan,
    live: dict[str, Any] | None,
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> dict[str, Any]:
    """Holders with STRD still unbonding, left out of pass 1. `staked_strd` is the chain-wide pool (live), whereas the
    addresses and `top` are the plan's deferred list (above the floor), so the two are not expected to match."""
    ustrd_price = by_denom[sweep_config.FEE_DENOM].price_usd if sweep_config.FEE_DENOM in by_denom else None
    staked_ustrd = int(live["staked_ustrd"]) if live is not None and "staked_ustrd" in live else None
    completions = [deferred.completes_at for deferred in plan.deferred if deferred.completes_at is not None]
    largest = sorted(plan.deferred, key=lambda deferred: deferred.usd, reverse=True)[:DEFERRED_TOP]
    usd = _live_deferred_usd(plan=plan, balances=balances, by_denom=by_denom) if live is not None else None

    return {
        "addresses": str(len(plan.deferred)),
        "usd": _money(usd if usd is not None else sum((deferred.usd for deferred in plan.deferred), Decimal(0))),
        "staked_strd": _money(_whole_strd(ustrd=staked_ustrd)) if staked_ustrd is not None else None,
        "staked_usd": _money(_whole_strd(ustrd=staked_ustrd) * ustrd_price)
        if staked_ustrd is not None and ustrd_price is not None
        else None,
        "completes_at": max(completions, key=holders.parse_chain_time) if completions else None,
        "top": [
            {
                "address": deferred.address,
                "usd": _money(deferred.usd),
                "staked_strd": _money(_whole_strd(ustrd=deferred.staked_ustrd)),
                "completes_at": deferred.completes_at,
            }
            for deferred in largest
        ],
    }


def _whole_strd(ustrd: int) -> Decimal:
    return Decimal(ustrd) / Decimal(10) ** 6


def _keyless(plan: planner.Plan, address_state: dict[str, str] | None) -> dict[str, str] | None:
    tier = next((candidate for candidate in plan.tiers if candidate.name == planner.TierName.KEYLESS), None)
    if tier is None:
        return None
    addresses = [planned for batch in tier.batches for planned in batch.addresses]
    swept = sum(
        1 for planned in addresses if address_state is not None and address_state.get(planned.address) == STATE_SWEPT
    )
    return {
        "addresses": str(len(addresses)),
        "usd": _money(sum((planned.usd for planned in addresses), Decimal(0))),
        "swept": str(swept),
    }


def _transfer_usd(transfer: ledger.Transfer, by_denom: dict[str, holders.SweepDenom]) -> Decimal:
    return (
        _amount_usd(amount=transfer.amount, denom=by_denom[transfer.denom])
        if transfer.denom in by_denom
        else Decimal(0)
    )


def _amount_usd(amount: int, denom: holders.SweepDenom) -> Decimal:
    return Decimal(amount) / Decimal(10) ** denom.decimals * denom.price_usd


def _bucket(count: int, usd: Decimal) -> dict[str, str]:
    return {"addresses": str(count), "usd": _money(usd)}


def _money(value: Decimal) -> str:
    return str(value.quantize(CENTS, rounding=ROUND_HALF_UP))
