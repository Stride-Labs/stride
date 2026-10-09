"""The append-only record of every batch submission and its outcome: state/ledger.jsonl, one JSON object per line.
Nothing here is ever rewritten; resume, status and the dashboard all derive from the lines."""

import dataclasses
import json
import os
import pathlib
from dataclasses import dataclass, field
from decimal import Decimal
from enum import StrEnum

from sweep import config


class LedgerError(Exception):
    """A line of the ledger cannot be read."""


class Kind(StrEnum):
    SUBMITTED = "submitted"
    CONFIRMED = "confirmed"
    FAILED = "failed"
    LOST = "lost"


class BatchState(StrEnum):
    PENDING = "pending"
    SUBMITTED = "submitted"
    CONFIRMED = "confirmed"
    FAILED = "failed"
    LOST = "lost"


TERMINAL_KINDS = {Kind.CONFIRMED, Kind.FAILED, Kind.LOST}
INT_FIELDS = ("run_id", "height", "gas_wanted", "gas_used", "addresses", "transfers_estimate", "code")


@dataclass(frozen=True)
class Transfer:
    address: str
    denom: str
    amount: int
    channel: str
    receiver: str


@dataclass(frozen=True)
class Skip:
    address: str
    reason: str


@dataclass(frozen=True)
class Event:
    kind: Kind
    batch_id: str
    at: str
    run_id: int | None = None
    tx_hash: str | None = None
    height: int | None = None
    gas_wanted: int | None = None
    gas_used: int | None = None
    addresses: int | None = None
    transfers_estimate: int | None = None
    transfers: list[Transfer] = field(default_factory=list)
    skipped: list[Skip] = field(default_factory=list)
    code: int | None = None
    codespace: str | None = None
    raw_log: str | None = None


# ---- constructors


def submitted(run_id: int, batch_id: str, tx_hash: str, at: str, addresses: int, transfers_estimate: int, gas_wanted: int) -> Event:
    return Event(kind=Kind.SUBMITTED, batch_id=batch_id, at=at, run_id=run_id, tx_hash=tx_hash, addresses=addresses,
                 transfers_estimate=transfers_estimate, gas_wanted=gas_wanted)


def confirmed(batch_id: str, tx_hash: str, at: str, height: int, gas_used: int, transfers: list[Transfer], skipped: list[Skip]) -> Event:
    return Event(kind=Kind.CONFIRMED, batch_id=batch_id, at=at, tx_hash=tx_hash, height=height, gas_used=gas_used,
                 transfers=list(transfers), skipped=list(skipped))


def failed(batch_id: str, tx_hash: str, at: str, code: int, codespace: str, raw_log: str) -> Event:
    return Event(kind=Kind.FAILED, batch_id=batch_id, at=at, tx_hash=tx_hash, code=code, codespace=codespace, raw_log=raw_log)


def lost(batch_id: str, tx_hash: str, at: str) -> Event:
    return Event(kind=Kind.LOST, batch_id=batch_id, at=at, tx_hash=tx_hash)


# ---- file


def append(event: Event, path: pathlib.Path = config.LEDGER_PATH) -> None:
    """One line, flushed and fsynced, so a crash leaves at most a torn last line (which `read` reports)."""
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(json.dumps(event_to_dict(event=event), separators=(",", ":")) + "\n")
        handle.flush()
        os.fsync(handle.fileno())


def read(path: pathlib.Path = config.LEDGER_PATH) -> list[Event]:
    if not path.exists():
        return []
    events: list[Event] = []
    for number, line in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
        if not line.strip():
            continue
        try:
            events.append(event_from_dict(data=json.loads(line)))
        except (json.JSONDecodeError, KeyError, TypeError, ValueError) as error:
            raise LedgerError(f"{path} line {number} is unreadable: {error}") from None
    return events


def event_to_dict(event: Event) -> dict:
    data = dataclasses.asdict(event)
    data["kind"] = event.kind.value
    for name in INT_FIELDS:
        if data[name] is not None:
            data[name] = str(data[name])
    for transfer in data["transfers"]:
        transfer["amount"] = str(transfer["amount"])
    return data


def event_from_dict(data: dict) -> Event:
    return Event(
        kind=Kind(data["kind"]),
        batch_id=data["batch_id"],
        at=data["at"],
        tx_hash=data.get("tx_hash"),
        codespace=data.get("codespace"),
        raw_log=data.get("raw_log"),
        transfers=[
            Transfer(
                address=transfer["address"],
                denom=transfer["denom"],
                amount=int(transfer["amount"]),
                channel=transfer["channel"],
                receiver=transfer["receiver"],
            )
            for transfer in data.get("transfers", [])
        ],
        skipped=[Skip(address=skip["address"], reason=skip["reason"]) for skip in data.get("skipped", [])],
        **{name: (int(data[name]) if data.get(name) is not None else None) for name in INT_FIELDS},
    )


# ---- derived state


def batch_states(events: list[Event]) -> dict[str, BatchState]:
    states: dict[str, BatchState] = {}
    for event in events:
        states[event.batch_id] = BatchState(event.kind.value)
    return states


def unresolved(events: list[Event]) -> list[Event]:
    """Submitted batches that have no confirmed, failed or lost line yet."""
    terminal = {event.batch_id for event in events if event.kind in TERMINAL_KINDS}
    return [event for event in events if event.kind == Kind.SUBMITTED and event.batch_id not in terminal]


def calibration(events: list[Event]) -> int | None:
    """Gas per transfer for the planner: the worst confirmed ratio with a margin, or None before any confirmation."""
    ratios = [Decimal(event.gas_used) / len(event.transfers) for event in events
              if event.kind == Kind.CONFIRMED and event.gas_used is not None and event.transfers]
    if not ratios:
        return None
    return int(max(ratios) * config.CALIBRATION_MARGIN)


def confirmed_transfers(events: list[Event]) -> dict[str, list[Transfer]]:
    by_address: dict[str, list[Transfer]] = {}
    for event in events:
        if event.kind != Kind.CONFIRMED:
            continue
        for transfer in event.transfers:
            by_address.setdefault(transfer.address, []).append(transfer)
    return by_address


def latest_run_id(events: list[Event]) -> int:
    return max((event.run_id for event in events if event.run_id is not None), default=0)
