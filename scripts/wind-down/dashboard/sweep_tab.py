"""The Sweep tab: `collect()` is the expensive live read (every holder of every sweep denom, in bulk) and is the tab's
collector, refreshed on demand; `body()` composes it with the plan, the ledger and the exclusions file read from disk
on every request, through the pure `compose()`. Every integer in the payload is a string."""

import pathlib
import sys
from decimal import ROUND_HALF_UP, Decimal
from typing import Any

sys.path.insert(0, str(pathlib.Path(__file__).resolve().parent.parent))  # the sweep package lives beside the dashboard

from sweep import chainio as sweep_chainio  # noqa: E402
from sweep import config as sweep_config  # noqa: E402
from sweep import holders, ledger, planner  # noqa: E402

CENTS = Decimal("0.01")
STATE_SWEPT = "swept"
STATE_REFUNDED = "refunded"
STATE_REMAINING = "remaining"


# ---- collector


def collect() -> dict[str, Any]:
    denoms = holders.resolve_denoms()
    balances = holders.read_balances(denoms=denoms)
    operator = sweep_chainio.rest_get(
        path=f"/cosmos/bank/v1beta1/balances/{sweep_config.SWEEP_OPERATOR}/by_denom",
        params={"denom": sweep_config.FEE_DENOM},
    )
    return {
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
            "totals": None,
            "by_denom": None,
            "batches": [],
            "ladder": None,
            "refunded": [],
            "exclusions": [],
            "keyless": None,
            "live_fetched_at": live_fetched_at,
        }

    states = ledger.batch_states(events=events)
    transferred = ledger.confirmed_transfers(events=events)
    balances = _live_balances(live=live)
    by_denom = {denom.denom: denom for denom in plan.denoms}
    address_state = _address_states(plan=plan, transferred=transferred, balances=balances) if live is not None else None
    refunds = _refunds(plan=plan, events=events, balances=balances) if live is not None else []

    return {
        "run": _run(plan=plan, states=states, live=live),
        "totals": _totals(
            plan=plan, address_state=address_state, transferred=transferred, balances=balances, by_denom=by_denom
        )
        if address_state is not None
        else None,
        "by_denom": _by_denom(plan=plan, address_state=address_state, transferred=transferred, balances=balances)
        if address_state is not None
        else None,
        "batches": _batches(plan=plan, events=events, states=states, refunds=refunds),
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
    """A vesting account keeps its locked part after a full sweep, so that much live balance is not a refund. Only the
    current plan knows it: an address swept in an earlier run and absent from this plan counts as having none locked, so
    a swept vesting holder from a past run that still holds its locked part reads as refunded until it is re-planned."""
    return {address: planned.locked for address, planned in _planned(plan=plan).items() if planned.locked}


def _back(address: str, denom: str, balances: dict[str, dict[str, int]], locked: dict[str, dict[str, int]]) -> int:
    """What the address holds of `denom` beyond its locked remainder: the refunded amount, 0 if only the remainder."""
    return max(0, balances.get(address, {}).get(denom, 0) - locked.get(address, {}).get(denom, 0))


def _known_addresses(plan: planner.Plan, transferred: dict[str, list[ledger.Transfer]]) -> set[str]:
    """Addresses already accounted for: planned, swept in any run, excluded, or skipped by the chain-side filters."""
    return (
        set(_planned(plan=plan))
        | set(transferred)
        | {excluded.address for excluded in plan.excluded}
        | {skipped.address for skipped in plan.skipped}
    )


def _address_states(
    plan: planner.Plan, transferred: dict[str, list[ledger.Transfer]], balances: dict[str, dict[str, int]]
) -> dict[str, str]:
    """swept: transferred and holding none of it now; refunded: transferred and holding a transferred denom again;
    remaining: in the current plan with no confirmed transfer. Swept and refunded cover every address with a confirmed
    transfer in any run, so a re-plan at a lower floor does not drop the earlier runs' holders."""
    locked = _locked(plan=plan)
    states = {address: STATE_REMAINING for address in _planned(plan=plan) if address not in transferred}
    for address, transfers in transferred.items():
        moved = {transfer.denom for transfer in transfers}
        states[address] = (
            STATE_REFUNDED
            if any(_back(address=address, denom=denom, balances=balances, locked=locked) > 0 for denom in moved)
            else STATE_SWEPT
        )
    return states


def _refunds(
    plan: planner.Plan, events: list[ledger.Event], balances: dict[str, dict[str, int]]
) -> list[tuple[ledger.Transfer, str]]:
    locked = _locked(plan=plan)
    refunds: list[tuple[ledger.Transfer, str]] = []
    for event in events:
        if event.kind != ledger.Kind.CONFIRMED:
            continue
        for transfer in event.transfers:
            if _back(address=transfer.address, denom=transfer.denom, balances=balances, locked=locked) > 0:
                refunds.append((transfer, event.batch_id))
    return refunds


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
        "batches": {
            "total": str(len(batches)),
            **{state: str(count) for state, count in counts.items() if state != "pending"},
            "pending": str(counts["pending"]),
        },
        "operator_strd": live["operator_strd"] if live is not None else None,
        "fee_estimate_ustrd": str(int(fee)),
    }


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

    swept_usd = sum(
        (_transfer_usd(transfer=transfer, by_denom=by_denom) for address in swept for transfer in transferred[address]),
        Decimal(0),
    )
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
        "excluded": _bucket(count=len(plan.excluded), usd=excluded_usd),
        "below_floor": _bucket(count=len(below), usd=sum(below.values(), Decimal(0))),
        "unplanned": _bucket(count=len(unplanned), usd=sum(unplanned.values(), Decimal(0))),
    }


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


def _ladder(
    plan: planner.Plan,
    address_state: dict[str, str],
    balances: dict[str, dict[str, int]],
    by_denom: dict[str, holders.SweepDenom],
) -> list[dict[str, str]]:
    """Live holders still to sweep: everyone with a balance except the swept, the excluded and the chain-skipped."""
    out = {address for address, state in address_state.items() if state == STATE_SWEPT}
    out |= {excluded.address for excluded in plan.excluded} | {skipped.address for skipped in plan.skipped}
    values = [
        holders.usd_value(balances=coins, denoms=by_denom) for address, coins in balances.items() if address not in out
    ]
    floors = [plan.floor_usd] + [floor for floor in sweep_config.LADDER_FLOORS if floor < plan.floor_usd]
    return [
        {
            "floor": str(floor),
            "holders": str(sum(1 for value in values if value >= floor)),
            "usd": _money(sum((value for value in values if value >= floor), Decimal(0))),
        }
        for floor in floors
    ]


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
