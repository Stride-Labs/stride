"""From a HolderSet to a plan: tiers in sweep order, batches packed under the address and gas limits, and the plan
file plus one address file per batch (the CLI's input). The plan is overwritten by every `plan`; the ledger never is."""

import decimal
import hashlib
import json
import pathlib
from dataclasses import dataclass, field
from decimal import Decimal
from enum import StrEnum
from typing import Any

from sweep import config, holders, ledger

BATCH_FILE_PREFIX = "batch-"
BATCH_FILE_SUFFIX = ".txt"


class TierName(StrEnum):
    TEST = "test"
    CANARY = "canary"
    MAIN = "main"
    KEYLESS = "keyless"


class PlanError(Exception):
    """The plan file cannot be read."""


@dataclass(frozen=True)
class PlannedAddress:
    address: str
    balances: dict[str, int]
    usd: Decimal
    transfers: int
    locked: dict[str, int] = field(default_factory=dict)  # a vesting account's remainder: still held after a full sweep


@dataclass(frozen=True)
class Batch:
    id: str
    file: str
    sha256: str
    addresses: list[PlannedAddress]
    usd: Decimal
    transfers: int
    estimated_gas: int


@dataclass(frozen=True)
class Tier:
    name: TierName
    batches: list[Batch]


@dataclass(frozen=True)
class LadderRung:
    floor: Decimal
    holders: int
    usd: Decimal


@dataclass(frozen=True)
class Plan:
    run_id: int
    created_at: str
    height: int
    floor_usd: Decimal
    test: bool
    canary: int
    gas_per_transfer: int
    prices: dict[str, Decimal]
    denoms: list[holders.SweepDenom]
    tiers: list[Tier]
    excluded: list[holders.Excluded]
    skipped: list[holders.Excluded]
    below_floor_count: int
    below_floor_usd: Decimal
    ladder: list[LadderRung]
    operator_sequence: int  # the sweep operator's account sequence when planned; the runner's tx-count baseline

    def pending_batches(self) -> list[tuple[Tier, Batch]]:
        return [(tier, batch) for tier in self.tiers for batch in tier.batches]

    def batch(self, batch_id: str) -> Batch | None:
        return next((batch for _, batch in self.pending_batches() if batch.id == batch_id), None)


# ---- building


def build_plan(
    holder_set: holders.HolderSet,
    floor_usd: Decimal,
    run_id: int,
    canary: int,
    test: bool,
    gas_per_transfer: int,
    max_addresses: int,
    gas_budget: int,
    created_at: str,
    operator_sequence: int,
) -> Plan:
    members = _tier_members(holder_list=holder_set.holders, canary=canary, test=test)

    # Batch numbers run across tiers so an id names one batch of this run whatever its tier
    tiers: list[Tier] = []
    index = 1
    for name, tier_holders in members:
        batches = pack(holder_list=tier_holders, tier=name, run_id=run_id, first_index=index, gas_per_transfer=gas_per_transfer,
                       max_addresses=max_addresses, gas_budget=gas_budget)
        if not batches:
            continue
        tiers.append(Tier(name=name, batches=batches))
        index += len(batches)

    return Plan(
        run_id=run_id, created_at=created_at, height=holder_set.height, floor_usd=floor_usd, test=test, canary=canary,
        gas_per_transfer=gas_per_transfer, prices={d.denom: d.price_usd for d in holder_set.denoms}, denoms=list(holder_set.denoms),
        tiers=tiers, excluded=list(holder_set.excluded), skipped=list(holder_set.skipped),
        below_floor_count=len(holder_set.below_floor), below_floor_usd=sum((h.usd for h in holder_set.below_floor), Decimal(0)),
        ladder=ladder(holder_list=holder_set.holders + holder_set.below_floor, floor_usd=floor_usd),
        operator_sequence=operator_sequence,
    )


def _tier_members(holder_list: list[holders.Holder], canary: int, test: bool) -> list[tuple[TierName, list[holders.Holder]]]:
    """Who goes in which tier, in sweep order. `holder_list` arrives sorted by USD descending."""
    if canary < 0:
        raise ValueError(f"canary must be >= 0, got {canary}")
    if test:
        return [(TierName.TEST, list(holder_list))]
    keyed = [h for h in holder_list if not h.keyless]
    keyless = [h for h in holder_list if h.keyless]
    canaries = sorted(keyed, key=lambda h: h.usd)[:canary]
    canary_addresses = {h.address for h in canaries}
    main = [h for h in keyed if h.address not in canary_addresses]
    return [(TierName.CANARY, canaries), (TierName.MAIN, main), (TierName.KEYLESS, keyless)]


def pack(
    holder_list: list[holders.Holder],
    tier: TierName,
    run_id: int,
    first_index: int,
    gas_per_transfer: int,
    max_addresses: int,
    gas_budget: int,
) -> list[Batch]:
    """Walk the holders in order; open a new batch when the next one would break the address or gas limit. A single
    holder over the budget still gets its own batch (the runner's simulation is the real guard)."""
    batches: list[Batch] = []
    current: list[holders.Holder] = []
    current_gas = 0

    def close() -> None:
        if current:
            batches.append(_batch(members=current, run_id=run_id, index=first_index + len(batches), gas_per_transfer=gas_per_transfer))

    for holder in holder_list:
        gas = holder.transfers * gas_per_transfer
        if current and (len(current) >= max_addresses or current_gas + gas > gas_budget):
            close()
            current, current_gas = [], 0
        current.append(holder)
        current_gas += gas
    close()
    return batches


def _batch(members: list[holders.Holder], run_id: int, index: int, gas_per_transfer: int) -> Batch:
    batch_id = f"{run_id:03d}-{index:03d}"
    planned = [PlannedAddress(address=h.address, balances=dict(h.balances), usd=h.usd, transfers=h.transfers,
                              locked=dict(h.locked)) for h in members]
    content = _addresses_file_content(addresses=[p.address for p in planned])
    transfers = sum(p.transfers for p in planned)
    return Batch(id=batch_id, file=f"{BATCH_FILE_PREFIX}{batch_id}{BATCH_FILE_SUFFIX}", sha256=batch_sha256(content=content),
                 addresses=planned, usd=sum((p.usd for p in planned), Decimal(0)), transfers=transfers,
                 estimated_gas=transfers * gas_per_transfer)


def ladder(holder_list: list[holders.Holder], floor_usd: Decimal) -> list[LadderRung]:
    floors = [floor_usd] + [f for f in config.LADDER_FLOORS if f < floor_usd]
    return [LadderRung(floor=f, holders=sum(1 for h in holder_list if h.usd >= f), usd=sum((h.usd for h in holder_list if h.usd >= f), Decimal(0)))
            for f in floors]


def gas_per_transfer(events: list[ledger.Event]) -> int:
    return ledger.calibration(events=events) or config.GAS_PER_TRANSFER_DEFAULT


def next_run_id(events: list[ledger.Event]) -> int:
    return ledger.latest_run_id(events=events) + 1


# ---- files


def batch_file_content(batch: Batch) -> str:
    return _addresses_file_content(addresses=[p.address for p in batch.addresses])


def _addresses_file_content(addresses: list[str]) -> str:
    return "".join(f"{address}\n" for address in addresses)


def batch_sha256(content: str) -> str:
    return hashlib.sha256(content.encode()).hexdigest()


def write_plan(plan: Plan, state_dir: pathlib.Path = config.STATE_DIR) -> None:
    state_dir.mkdir(parents=True, exist_ok=True)
    new_files = {batch.file for _, batch in plan.pending_batches()}

    # new files and plan.json land first so a crash mid-write never leaves the old plan without its batch files
    for _, batch in plan.pending_batches():
        (state_dir / batch.file).write_text(batch_file_content(batch=batch))
    tmp = state_dir / "plan.json.tmp"
    tmp.write_text(json.dumps(plan_to_dict(plan=plan), indent=1))
    tmp.replace(state_dir / "plan.json")

    for stale in state_dir.glob(f"{BATCH_FILE_PREFIX}*{BATCH_FILE_SUFFIX}"):
        if stale.name not in new_files:
            stale.unlink()


def load_plan(path: pathlib.Path = config.PLAN_PATH) -> Plan | None:
    if not path.exists():
        return None
    try:
        return plan_from_dict(data=json.loads(path.read_text()))
    # a hand-edited or truncated plan must surface as PlanError, not a raw decode/shape/Decimal exception
    except (json.JSONDecodeError, KeyError, ValueError, TypeError, AttributeError, decimal.InvalidOperation) as error:
        raise PlanError(f"{path}: {error}") from None


def plan_to_dict(plan: Plan) -> dict[str, Any]:
    return {
        "run_id": str(plan.run_id), "created_at": plan.created_at, "height": str(plan.height), "floor_usd": str(plan.floor_usd),
        "test": plan.test, "canary": str(plan.canary), "gas_per_transfer": str(plan.gas_per_transfer),
        "prices": {denom: str(price) for denom, price in plan.prices.items()},
        "denoms": [{"denom": d.denom, "symbol": d.symbol, "decimals": str(d.decimals), "price_usd": str(d.price_usd),
                    "destination": d.destination, "channel": d.channel} for d in plan.denoms],
        "tiers": [{"name": tier.name.value, "batches": [_batch_to_dict(batch=b) for b in tier.batches]} for tier in plan.tiers],
        "excluded": [_excluded_to_dict(entry=e) for e in plan.excluded],
        "skipped": [_excluded_to_dict(entry=e) for e in plan.skipped],
        "below_floor": {"count": str(plan.below_floor_count), "usd": str(plan.below_floor_usd)},
        "ladder": [{"floor": str(r.floor), "holders": str(r.holders), "usd": str(r.usd)} for r in plan.ladder],
        "operator_sequence": str(plan.operator_sequence),
    }


def _batch_to_dict(batch: Batch) -> dict[str, Any]:
    return {
        "id": batch.id, "file": batch.file, "sha256": batch.sha256, "usd": str(batch.usd), "transfers": str(batch.transfers),
        "estimated_gas": str(batch.estimated_gas),
        "addresses": [{"address": p.address, "balances": {d: str(a) for d, a in p.balances.items()}, "usd": str(p.usd),
                       "transfers": str(p.transfers), "locked": {d: str(a) for d, a in p.locked.items()}} for p in batch.addresses],
    }


def _excluded_to_dict(entry: holders.Excluded) -> dict[str, Any]:
    return {"address": entry.address, "reason": entry.reason, "usd": str(entry.usd)}


def plan_from_dict(data: dict) -> Plan:
    return Plan(
        run_id=int(data["run_id"]), created_at=data["created_at"], height=int(data["height"]), floor_usd=Decimal(data["floor_usd"]),
        test=bool(data["test"]), canary=int(data["canary"]), gas_per_transfer=int(data["gas_per_transfer"]),
        prices={denom: Decimal(price) for denom, price in data["prices"].items()},
        denoms=[holders.SweepDenom(denom=d["denom"], symbol=d["symbol"], decimals=int(d["decimals"]), price_usd=Decimal(d["price_usd"]),
                                   destination=d["destination"], channel=d["channel"]) for d in data["denoms"]],
        tiers=[Tier(name=TierName(t["name"]), batches=[_batch_from_dict(data=b) for b in t["batches"]]) for t in data["tiers"]],
        excluded=[_excluded_from_dict(data=e) for e in data["excluded"]],
        skipped=[_excluded_from_dict(data=e) for e in data["skipped"]],
        below_floor_count=int(data["below_floor"]["count"]), below_floor_usd=Decimal(data["below_floor"]["usd"]),
        ladder=[LadderRung(floor=Decimal(r["floor"]), holders=int(r["holders"]), usd=Decimal(r["usd"])) for r in data["ladder"]],
        operator_sequence=int(data["operator_sequence"]),
    )


def _batch_from_dict(data: dict) -> Batch:
    return Batch(
        id=data["id"], file=data["file"], sha256=data["sha256"], usd=Decimal(data["usd"]), transfers=int(data["transfers"]),
        estimated_gas=int(data["estimated_gas"]),
        addresses=[PlannedAddress(address=p["address"], balances={d: int(a) for d, a in p["balances"].items()}, usd=Decimal(p["usd"]),
                                  transfers=int(p["transfers"]), locked={d: int(a) for d, a in p.get("locked", {}).items()})
                   for p in data["addresses"]],
    )


def _excluded_from_dict(data: dict) -> holders.Excluded:
    return holders.Excluded(address=data["address"], reason=data["reason"], usd=Decimal(data["usd"]))
