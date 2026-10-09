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
USD_CENT = Decimal("0.01")


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
    # Every vesting holder's locked remainder, including those below the floor: a fully swept holder is not planned
    locked: dict[str, dict[str, int]] = field(default_factory=dict)

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
        batches = pack(
            holder_list=tier_holders,
            tier=name,
            run_id=run_id,
            first_index=index,
            gas_per_transfer=gas_per_transfer,
            max_addresses=max_addresses,
            gas_budget=gas_budget,
        )
        if not batches:
            continue
        tiers.append(Tier(name=name, batches=batches))
        index += len(batches)

    return Plan(
        run_id=run_id,
        created_at=created_at,
        height=holder_set.height,
        floor_usd=floor_usd,
        test=test,
        canary=canary,
        gas_per_transfer=gas_per_transfer,
        prices={entry.denom: entry.price_usd for entry in holder_set.denoms},
        denoms=list(holder_set.denoms),
        tiers=tiers,
        excluded=list(holder_set.excluded),
        skipped=list(holder_set.skipped),
        below_floor_count=len(holder_set.below_floor),
        below_floor_usd=sum((holder.usd for holder in holder_set.below_floor), Decimal(0)),
        ladder=ladder(holder_list=holder_set.holders + holder_set.below_floor, floor_usd=floor_usd),
        operator_sequence=operator_sequence,
        locked={
            holder.address: dict(holder.locked) for holder in holder_set.holders + holder_set.below_floor if holder.locked
        },
    )


def _tier_members(
    holder_list: list[holders.Holder], canary: int, test: bool
) -> list[tuple[TierName, list[holders.Holder]]]:
    """Who goes in which tier, in sweep order. `holder_list` arrives sorted by USD descending."""
    if canary < 0:
        raise ValueError(f"canary must be >= 0, got {canary}")
    if test:
        return [(TierName.TEST, list(holder_list))]
    keyed = [holder for holder in holder_list if not holder.keyless]
    keyless = [holder for holder in holder_list if holder.keyless]
    canaries = sorted(keyed, key=lambda holder: holder.usd)[:canary]
    canary_addresses = {holder.address for holder in canaries}
    main = [holder for holder in keyed if holder.address not in canary_addresses]
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
            batches.append(
                _batch(
                    members=current, run_id=run_id, index=first_index + len(batches), gas_per_transfer=gas_per_transfer
                )
            )

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
    planned = [
        PlannedAddress(
            address=member.address,
            balances=dict(member.balances),
            usd=member.usd,
            transfers=member.transfers,
            locked=dict(member.locked),
        )
        for member in members
    ]
    content = _addresses_file_content(addresses=[entry.address for entry in planned])
    transfers = sum(entry.transfers for entry in planned)
    return Batch(
        id=batch_id,
        file=f"{BATCH_FILE_PREFIX}{batch_id}{BATCH_FILE_SUFFIX}",
        sha256=batch_sha256(content=content),
        addresses=planned,
        usd=sum((entry.usd for entry in planned), Decimal(0)),
        transfers=transfers,
        estimated_gas=transfers * gas_per_transfer,
    )


def ladder(holder_list: list[holders.Holder], floor_usd: Decimal) -> list[LadderRung]:
    floors = [floor_usd] + [rung_floor for rung_floor in config.LADDER_FLOORS if rung_floor < floor_usd]
    return [
        LadderRung(
            floor=rung_floor,
            holders=sum(1 for holder in holder_list if holder.usd >= rung_floor),
            usd=sum((holder.usd for holder in holder_list if holder.usd >= rung_floor), Decimal(0)),
        )
        for rung_floor in floors
    ]


def gas_per_transfer(events: list[ledger.Event]) -> int:
    return ledger.calibration(events=events) or config.GAS_PER_TRANSFER_DEFAULT


def next_run_id(events: list[ledger.Event]) -> int:
    return ledger.latest_run_id(events=events) + 1


# ---- files


def batch_file_content(batch: Batch) -> str:
    return _addresses_file_content(addresses=[entry.address for entry in batch.addresses])


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
        "run_id": str(plan.run_id),
        "created_at": plan.created_at,
        "height": str(plan.height),
        "floor_usd": _usd_str(amount=plan.floor_usd),
        "test": plan.test,
        "canary": str(plan.canary),
        "gas_per_transfer": str(plan.gas_per_transfer),
        "prices": {denom: str(price) for denom, price in plan.prices.items()},  # USD per whole token, not rounded
        "denoms": [
            {
                "denom": entry.denom,
                "symbol": entry.symbol,
                "decimals": str(entry.decimals),
                "price_usd": str(entry.price_usd),
                "destination": entry.destination,
                "channel": entry.channel,
            }
            for entry in plan.denoms
        ],
        "tiers": [
            {"name": tier.name.value, "batches": [_batch_to_dict(batch=batch) for batch in tier.batches]}
            for tier in plan.tiers
        ],
        "excluded": [_excluded_to_dict(entry=entry) for entry in plan.excluded],
        "skipped": [_excluded_to_dict(entry=entry) for entry in plan.skipped],
        "below_floor": {"count": str(plan.below_floor_count), "usd": _usd_str(amount=plan.below_floor_usd)},
        "ladder": [
            {"floor": _usd_str(amount=rung.floor), "holders": str(rung.holders), "usd": _usd_str(amount=rung.usd)}
            for rung in plan.ladder
        ],
        "operator_sequence": str(plan.operator_sequence),
        "locked": {
            address: {denom: str(amount) for denom, amount in coins.items()} for address, coins in plan.locked.items()
        },
    }


def _usd_str(amount: Decimal) -> str:
    """USD is written to cents; the plan is a review artifact and full precision is noise. Reading accepts any."""
    return str(amount.quantize(USD_CENT, rounding=decimal.ROUND_HALF_UP))


def _batch_to_dict(batch: Batch) -> dict[str, Any]:
    return {
        "id": batch.id,
        "file": batch.file,
        "sha256": batch.sha256,
        "usd": _usd_str(amount=batch.usd),
        "transfers": str(batch.transfers),
        "estimated_gas": str(batch.estimated_gas),
        "addresses": [
            {
                "address": entry.address,
                "balances": {denom: str(amount) for denom, amount in entry.balances.items()},
                "usd": _usd_str(amount=entry.usd),
                "transfers": str(entry.transfers),
                "locked": {denom: str(amount) for denom, amount in entry.locked.items()},
            }
            for entry in batch.addresses
        ],
    }


def _excluded_to_dict(entry: holders.Excluded) -> dict[str, Any]:
    return {"address": entry.address, "reason": entry.reason, "usd": _usd_str(amount=entry.usd)}


def plan_from_dict(data: dict) -> Plan:
    return Plan(
        run_id=int(data["run_id"]),
        created_at=data["created_at"],
        height=int(data["height"]),
        floor_usd=Decimal(data["floor_usd"]),
        test=bool(data["test"]),
        canary=int(data["canary"]),
        gas_per_transfer=int(data["gas_per_transfer"]),
        prices={denom: Decimal(price) for denom, price in data["prices"].items()},
        denoms=[
            holders.SweepDenom(
                denom=denom_data["denom"],
                symbol=denom_data["symbol"],
                decimals=int(denom_data["decimals"]),
                price_usd=Decimal(denom_data["price_usd"]),
                destination=denom_data["destination"],
                channel=denom_data["channel"],
            )
            for denom_data in data["denoms"]
        ],
        tiers=[
            Tier(
                name=TierName(tier_data["name"]),
                batches=[_batch_from_dict(data=batch_data) for batch_data in tier_data["batches"]],
            )
            for tier_data in data["tiers"]
        ],
        excluded=[_excluded_from_dict(data=excluded_data) for excluded_data in data["excluded"]],
        skipped=[_excluded_from_dict(data=skipped_data) for skipped_data in data["skipped"]],
        below_floor_count=int(data["below_floor"]["count"]),
        below_floor_usd=Decimal(data["below_floor"]["usd"]),
        ladder=[
            LadderRung(
                floor=Decimal(rung_data["floor"]), holders=int(rung_data["holders"]), usd=Decimal(rung_data["usd"])
            )
            for rung_data in data["ladder"]
        ],
        operator_sequence=int(data["operator_sequence"]),
        locked={
            address: {denom: int(amount) for denom, amount in coins.items()}
            for address, coins in data.get("locked", {}).items()
        },
    )


def _batch_from_dict(data: dict) -> Batch:
    return Batch(
        id=data["id"],
        file=data["file"],
        sha256=data["sha256"],
        usd=Decimal(data["usd"]),
        transfers=int(data["transfers"]),
        estimated_gas=int(data["estimated_gas"]),
        addresses=[
            PlannedAddress(
                address=address_data["address"],
                balances={denom: int(amount) for denom, amount in address_data["balances"].items()},
                usd=Decimal(address_data["usd"]),
                transfers=int(address_data["transfers"]),
                locked={denom: int(amount) for denom, amount in address_data.get("locked", {}).items()},
            )
            for address_data in data["addresses"]
        ],
    )


def _excluded_from_dict(data: dict) -> holders.Excluded:
    return holders.Excluded(address=data["address"], reason=data["reason"], usd=Decimal(data["usd"]))
