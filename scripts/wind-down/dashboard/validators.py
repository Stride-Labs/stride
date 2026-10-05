"""Collector for the Validators tab: Stride's recorded per-validator delegations against what the host chain holds.

Amounts are exact integers (strings in the payload, plus the zone's decimals); rates and percentages are
Decimal strings. Nothing here uses float.
"""

import concurrent.futures
import dataclasses
import decimal
from dataclasses import dataclass
from decimal import Decimal
from enum import StrEnum
from typing import Any

import chain
import config

STAKEIBC_HOST_ZONE_PATH = "/Stride-Labs/stride/stakeibc/host_zone"
STAKETIA_HOST_ZONE_PATH = "/Stride-Labs/stride/staketia/host_zone"
STAKETIA_CHAIN_ID = "staketia"
STAKETIA_HOST_CHAIN_ID = "celestia"  # the chain the staketia multisig delegates on

MAX_UNBONDING_ENTRIES = 7
# Over by at least this fraction of the recorded delegation is red: 0.0001% = 1 / 1,000,000.
RED_FRACTION_DENOMINATOR = 1_000_000
# x/stakeibc UndelegationSharesSafetyDivisor: a full drain of a validator whose stored rate is below 1 is
# shaved by recorded / this (at least 1 base unit), so an overage within that buffer cannot fail the drain.
DRAIN_BUFFER_DIVISOR = 100_000_000_000_000_000
BOND_STATUS_PREFIX = "BOND_STATUS_"

ZONE_WORKERS = 16
PERCENT_PLACES = Decimal("0.00000001")
RATE_PLACES = Decimal("0.000000000000000001")
WEIGHT_PLACES = Decimal("0.01")
# Amounts run to 1e30 (18-decimal zones); the thread-default 28 digits would round them.
CONTEXT = decimal.Context(prec=60)

class Severity(StrEnum):
    NEUTRAL = "neutral"  # recorded at or under the host amount
    BUFFERED = "buffered"  # over, but within the rounding buffer the drain leaves on a slashed validator
    AMBER = "amber"  # recorded over the host amount by any amount
    RED = "red"  # over by at least 0.0001% of the recorded delegation


@dataclass(frozen=True)
class StrideValidator:
    address: str
    name: str
    delegation: int  # recorded by Stride
    weight: int
    rate: Decimal  # shares_to_tokens_rate as Stride last recorded it
    slash_query_in_progress: bool
    delegation_changes_in_progress: int


@dataclass(frozen=True)
class HostValidator:
    moniker: str
    status: str  # "bonded", "unbonding", "unbonded"
    jailed: bool
    rate: (
        Decimal | None
    )  # tokens / delegator_shares; None for a validator with no shares


@dataclass(frozen=True)
class ValidatorRow:
    moniker: str
    address: str
    registered: (
        bool  # False when the host delegation has no entry in Stride's validator list
    )
    weight_percent: Decimal
    recorded: int
    actual: int
    diff: int  # recorded - actual; positive means Stride believes more than the host holds
    diff_percent: Decimal | None  # of recorded; None when nothing is recorded
    stride_rate: Decimal | None
    chain_rate: Decimal | None
    rate_difference: Decimal | None  # stride_rate - chain_rate
    unbonding_entries: int | None  # None when the lookup failed
    delegation_changes_in_progress: int | None  # None for a validator Stride does not track
    slash_query_in_progress: bool | None
    bond_status: (
        str | None
    )  # None when the host validator list could not be read or lacks the validator
    jailed: bool | None
    severity: Severity

    @property
    def in_progress(self) -> bool:
        return bool(self.delegation_changes_in_progress) or bool(self.slash_query_in_progress)

    def payload(self) -> dict[str, Any]:
        return {
            "moniker": self.moniker,
            "address": self.address,
            "registered": self.registered,
            "weight_percent": _decimal_text(value=self.weight_percent),
            "recorded": str(self.recorded),
            "actual": str(self.actual),
            "diff": str(self.diff),
            "diff_percent": _optional_decimal_text(value=self.diff_percent),
            "stride_rate": _optional_decimal_text(value=self.stride_rate),
            "chain_rate": _optional_decimal_text(value=self.chain_rate),
            "rate_difference": _optional_decimal_text(value=self.rate_difference),
            "unbonding_entries": self.unbonding_entries,
            "delegation_changes_in_progress": self.delegation_changes_in_progress,
            "slash_query_in_progress": self.slash_query_in_progress,
            "bond_status": self.bond_status,
            "jailed": self.jailed,
            "severity": self.severity,
        }


@dataclass(frozen=True)
class LiveTestCandidate:
    address: str
    moniker: str
    recorded: int

    def payload(self) -> dict[str, Any]:
        return {**dataclasses.asdict(self), "recorded": str(self.recorded)}


@dataclass(frozen=True)
class LiveTestPicks:
    """The validator to drain in full as the live test, and the runner-up; both None with a reason when none qualifies."""

    pick: LiveTestCandidate | None
    runner_up: LiveTestCandidate | None
    reason: str | None  # why there is no pick; None when there is one


@dataclass(frozen=True)
class ZoneTotals:
    validator_count: int
    delegated_count: int  # validators that hold a delegation on the host
    recorded: int
    actual: int
    diff: int
    in_progress_count: int
    over_count: int  # recorded over the host amount, at any severity
    red_count: int

    def payload(self) -> dict[str, Any]:
        return {
            **dataclasses.asdict(self),
            "recorded": str(self.recorded),
            "actual": str(self.actual),
            "diff": str(self.diff),
        }


@dataclass(frozen=True)
class MultisigRow:
    moniker: str
    address: str
    actual: int
    bond_status: str | None
    jailed: bool | None

    def payload(self) -> dict[str, Any]:
        return {**dataclasses.asdict(self), "actual": str(self.actual)}


# ---- public API


def collect() -> dict[str, Any]:
    stride_zones = _stride_host_zones()

    with concurrent.futures.ThreadPoolExecutor(max_workers=ZONE_WORKERS) as pool:
        zone_futures = [
            pool.submit(_collect_zone_guarded, zone=zone, stride_zones=stride_zones)
            for zone in config.ZONES
        ]
        staketia_future = pool.submit(_collect_staketia_guarded)

    zones = [future.result() for future in zone_futures] + [staketia_future.result()]
    return {"zones": zones}


def severity_of(recorded: int, actual: int, stride_rate: Decimal | None = None) -> Severity:
    """Over by any amount is amber; over by at least 0.0001% of the recorded delegation is red.

    An overage no larger than the drain's rounding buffer is `buffered` (harmless): that buffer only exists
    when Stride's stored rate is below 1, mirroring applySharesRoundingSafety.
    """
    diff = recorded - actual
    if diff <= 0:
        return Severity.NEUTRAL
    if stride_rate is not None and stride_rate < 1 and diff <= drain_buffer(recorded=recorded):
        return Severity.BUFFERED
    # Exact integer form of diff / recorded >= 0.0001%; recorded > 0 here because actual >= 0.
    return (
        Severity.RED if diff * RED_FRACTION_DENOMINATOR >= recorded else Severity.AMBER
    )


def drain_buffer(recorded: int) -> int:
    return max(1, recorded // DRAIN_BUFFER_DIVISOR)


def build_rows(
    stride_validators: list[StrideValidator],
    delegations: dict[str, int],
    host_validators: dict[str, HostValidator] | None,
    unbonding_entries: dict[str, int] | None,
) -> list[ValidatorRow]:
    """One row per Stride validator plus one per host delegation Stride does not know, largest |diff| first."""
    total_weight = sum(validator.weight for validator in stride_validators)
    registered_rows = [
        _registered_row(
            validator=validator,
            total_weight=total_weight,
            actual=delegations.get(validator.address, 0),
            host_validator=(host_validators or {}).get(validator.address),
            unbonding_entries=_entries_for(
                unbonding=unbonding_entries, address=validator.address
            ),
        )
        for validator in stride_validators
    ]

    registered_addresses = {validator.address for validator in stride_validators}
    unregistered_rows = [
        _unregistered_row(
            address=address,
            actual=actual,
            host_validator=(host_validators or {}).get(address),
            unbonding_entries=_entries_for(
                unbonding=unbonding_entries, address=address
            ),
        )
        for address, actual in delegations.items()
        if address not in registered_addresses and actual > 0
    ]

    rows = registered_rows + unregistered_rows
    return sorted(
        rows, key=lambda row: (-abs(row.diff), row.moniker.lower(), row.address)
    )


def summarize(rows: list[ValidatorRow]) -> ZoneTotals:
    recorded = sum(row.recorded for row in rows)
    actual = sum(row.actual for row in rows)
    return ZoneTotals(
        validator_count=len(rows),
        delegated_count=sum(1 for row in rows if row.actual > 0),
        recorded=recorded,
        actual=actual,
        diff=recorded - actual,
        in_progress_count=sum(1 for row in rows if row.in_progress),
        over_count=sum(1 for row in rows if row.severity in (Severity.AMBER, Severity.RED)),
        red_count=sum(1 for row in rows if row.severity == Severity.RED),
    )


def pick_live_test(rows: list[ValidatorRow], decimals: int) -> LiveTestPicks:
    """The live-test validator: the smallest recorded delegation of at least one whole token with no unbonding entry
    in flight (mirrors scripts/wind-down/pick_live_test_validators.py), and the second smallest by the same rule.

    A validator Stride does not track, or one with a delegation change or slash query in progress, cannot be drained
    cleanly, so it is out. Without the unbonding entries nothing can be ruled in, so there is no pick.
    """
    if any(row.unbonding_entries is None for row in rows):
        return LiveTestPicks(pick=None, runner_up=None, reason="unbonding entries could not be read")

    whole_token = 10**decimals
    funded = [row for row in rows if row.registered and row.recorded >= whole_token]
    if not funded:
        return LiveTestPicks(pick=None, runner_up=None, reason="no registered validator holds a whole token")

    clean = sorted(
        (row for row in funded if not row.in_progress and row.unbonding_entries == 0),
        key=lambda row: (row.recorded, row.address),
    )
    if clean:
        candidates = [LiveTestCandidate(address=row.address, moniker=row.moniker, recorded=row.recorded) for row in clean[:2]]
        return LiveTestPicks(pick=candidates[0], runner_up=candidates[1] if len(candidates) > 1 else None, reason=None)

    # Once the day epoch has touched every validator, none is entry-free; a full drain still only adds one entry,
    # so fall back to the smallest validator with room under the host's cap (the script does the same).
    with_room = sorted(
        (row for row in funded if not row.in_progress and row.unbonding_entries < MAX_UNBONDING_ENTRIES),
        key=lambda row: (row.recorded, row.address),
    )
    if not with_room:
        return LiveTestPicks(
            pick=None,
            runner_up=None,
            reason=f"all {len(funded)} funded validators are at the entry cap or have a change in progress",
        )
    candidates = [LiveTestCandidate(address=row.address, moniker=row.moniker, recorded=row.recorded) for row in with_room[:2]]
    return LiveTestPicks(
        pick=candidates[0],
        runner_up=candidates[1] if len(candidates) > 1 else None,
        reason=f"no entry-free validator; the pick has {with_room[0].unbonding_entries} of {MAX_UNBONDING_ENTRIES} entries in flight",
    )


# ---- zones


def _collect_zone_guarded(
    zone: config.ZoneConfig, stride_zones: dict[str, dict[str, Any]]
) -> dict[str, Any]:
    # The one error boundary per zone: a dead host REST endpoint blanks that zone's chip, not the tab.
    try:
        return _collect_zone(zone=zone, stride_zone=stride_zones[zone.chain_id])
    except chain.ZONE_ERRORS as error:
        return _error_entry(chain_id=zone.chain_id, error=error)


def _collect_zone(
    zone: config.ZoneConfig, stride_zone: dict[str, Any]
) -> dict[str, Any]:
    host = chain.zone_chain(zone=zone)
    ica_address = stride_zone["delegation_ica_address"]
    stride_validators = [
        _parse_stride_validator(raw=raw) for raw in stride_zone["validators"]
    ]

    # Delegations are the one required lookup; the rest only decorate rows, so a failure becomes null.
    delegations = _host_delegations(host=host, delegator=ica_address)
    host_validators = chain.optional(lookup=lambda: _host_validators(host=host))
    unbonding_entries = chain.optional(
        lookup=lambda: _unbonding_entries(host=host, delegator=ica_address)
    )

    rows = build_rows(
        stride_validators=stride_validators,
        delegations=delegations,
        host_validators=host_validators,
        unbonding_entries=unbonding_entries,
    )
    totals = summarize(rows=rows)
    live_test = pick_live_test(rows=rows, decimals=zone.decimals)
    return {
        "kind": "zone",
        "chain_id": zone.chain_id,
        "symbol": zone.symbol,
        "decimals": zone.decimals,
        "delegation_address": ica_address,
        "max_unbonding_entries": MAX_UNBONDING_ENTRIES,
        "totals": totals.payload(),
        "live_test_pick": _optional_payload(candidate=live_test.pick),
        "live_test_next": _optional_payload(candidate=live_test.runner_up),
        "live_test_reason": live_test.reason,
        "validators": [row.payload() for row in rows],
    }


def _collect_staketia_guarded() -> dict[str, Any]:
    try:
        return _collect_staketia()
    except chain.ZONE_ERRORS as error:
        return _error_entry(chain_id=STAKETIA_CHAIN_ID, error=error)


def _collect_staketia() -> dict[str, Any]:
    zone = config.ZONES_BY_CHAIN_ID[STAKETIA_HOST_CHAIN_ID]
    host = chain.zone_chain(zone=zone)
    staketia_zone = chain.rest_get(
        chain=chain.stride_chain(), path=STAKETIA_HOST_ZONE_PATH
    )["host_zone"]
    multisig = staketia_zone["delegation_address"]
    remaining = int(staketia_zone["remaining_delegated_balance"])

    delegations = _host_delegations(host=host, delegator=multisig)
    host_validators = chain.optional(lookup=lambda: _host_validators(host=host))

    rows = sorted(
        (
            _multisig_row(
                address=address,
                actual=actual,
                host_validator=(host_validators or {}).get(address),
            )
            for address, actual in delegations.items()
            if actual > 0
        ),
        key=lambda row: (-row.actual, row.address),
    )
    actual_total = sum(row.actual for row in rows)
    return {
        "kind": "multisig",
        "chain_id": STAKETIA_CHAIN_ID,
        "symbol": zone.symbol,
        "decimals": zone.decimals,
        "delegation_address": multisig,
        "validators": [row.payload() for row in rows],
        "actual_total": str(actual_total),
        "remaining_delegated_balance": str(remaining),
        "diff": str(remaining - actual_total),
        "severity": severity_of(recorded=remaining, actual=actual_total),
    }


def _error_entry(chain_id: str, error: Exception) -> dict[str, Any]:
    zone_error = chain.ZoneError(chain_id=chain_id, error=f"{type(error).__name__}: {error}")
    return dataclasses.asdict(zone_error)


# ---- queries


def _stride_host_zones() -> dict[str, dict[str, Any]]:
    """Every stakeibc host zone by chain id, fetched once per collect."""
    response = chain.rest_get(chain=chain.stride_chain(), path=STAKEIBC_HOST_ZONE_PATH)
    return {zone["chain_id"]: zone for zone in response["host_zone"]}


def _host_delegations(host: chain.Chain, delegator: str) -> dict[str, int]:
    responses = chain.rest_get_all_pages(
        chain=host,
        path=f"/cosmos/staking/v1beta1/delegations/{delegator}",
        key="delegation_responses",
    )
    return {
        response["delegation"]["validator_address"]: int(response["balance"]["amount"])
        for response in responses
    }


def _host_validators(host: chain.Chain) -> dict[str, HostValidator]:
    # No status filter: the SDK then returns bonded, unbonding and unbonded validators alike.
    raw_validators = chain.rest_get_all_pages(
        chain=host, path="/cosmos/staking/v1beta1/validators", key="validators"
    )
    return {
        raw["operator_address"]: _parse_host_validator(raw=raw)
        for raw in raw_validators
    }


def _unbonding_entries(host: chain.Chain, delegator: str) -> dict[str, int]:
    responses = chain.rest_get_all_pages(
        chain=host,
        path=f"/cosmos/staking/v1beta1/delegators/{delegator}/unbonding_delegations",
        key="unbonding_responses",
    )
    return {
        response["validator_address"]: len(response["entries"])
        for response in responses
    }


# ---- parsing and rows


def _parse_stride_validator(raw: dict[str, Any]) -> StrideValidator:
    return StrideValidator(
        address=raw["address"],
        name=raw["name"],
        delegation=int(raw["delegation"]),
        weight=int(raw["weight"]),
        rate=Decimal(raw["shares_to_tokens_rate"]),
        slash_query_in_progress=bool(raw["slash_query_in_progress"]),
        delegation_changes_in_progress=int(raw["delegation_changes_in_progress"]),
    )


def _parse_host_validator(raw: dict[str, Any]) -> HostValidator:
    shares = Decimal(raw["delegator_shares"])
    return HostValidator(
        moniker=raw["description"]["moniker"],
        status=raw["status"].removeprefix(BOND_STATUS_PREFIX).lower(),
        jailed=bool(raw.get("jailed", False)),
        rate=_divide(numerator=Decimal(raw["tokens"]), denominator=shares, places=RATE_PLACES)
        if shares
        else None,
    )


def _registered_row(
    validator: StrideValidator,
    total_weight: int,
    actual: int,
    host_validator: HostValidator | None,
    unbonding_entries: int | None,
) -> ValidatorRow:
    weight_percent = (
        _percent(part=validator.weight, whole=total_weight, places=WEIGHT_PLACES)
        if total_weight
        else Decimal(0)
    )
    return _row(
        moniker=host_validator.moniker if host_validator else validator.name,
        address=validator.address,
        registered=True,
        weight_percent=weight_percent,
        recorded=validator.delegation,
        actual=actual,
        stride_rate=validator.rate,
        host_validator=host_validator,
        unbonding_entries=unbonding_entries,
        delegation_changes_in_progress=validator.delegation_changes_in_progress,
        slash_query_in_progress=validator.slash_query_in_progress,
    )


def _unregistered_row(
    address: str,
    actual: int,
    host_validator: HostValidator | None,
    unbonding_entries: int | None,
) -> ValidatorRow:
    return _row(
        moniker=host_validator.moniker if host_validator else address,
        address=address,
        registered=False,
        weight_percent=Decimal(0).quantize(WEIGHT_PLACES),
        recorded=0,
        actual=actual,
        stride_rate=None,
        host_validator=host_validator,
        unbonding_entries=unbonding_entries,
        delegation_changes_in_progress=None,
        slash_query_in_progress=None,
    )


def _row(
    moniker: str,
    address: str,
    registered: bool,
    weight_percent: Decimal,
    recorded: int,
    actual: int,
    stride_rate: Decimal | None,
    host_validator: HostValidator | None,
    unbonding_entries: int | None,
    delegation_changes_in_progress: int | None,
    slash_query_in_progress: bool | None,
) -> ValidatorRow:
    diff = recorded - actual
    chain_rate = host_validator.rate if host_validator else None
    rate_difference = (
        stride_rate - chain_rate
        if stride_rate is not None and chain_rate is not None
        else None
    )
    return ValidatorRow(
        moniker=moniker,
        address=address,
        registered=registered,
        weight_percent=weight_percent,
        recorded=recorded,
        actual=actual,
        diff=diff,
        diff_percent=_percent(part=diff, whole=recorded, places=PERCENT_PLACES)
        if recorded
        else None,
        stride_rate=stride_rate,
        chain_rate=chain_rate,
        rate_difference=rate_difference,
        unbonding_entries=unbonding_entries,
        delegation_changes_in_progress=delegation_changes_in_progress,
        slash_query_in_progress=slash_query_in_progress,
        bond_status=host_validator.status if host_validator else None,
        jailed=host_validator.jailed if host_validator else None,
        severity=severity_of(recorded=recorded, actual=actual, stride_rate=stride_rate),
    )


def _multisig_row(
    address: str, actual: int, host_validator: HostValidator | None
) -> MultisigRow:
    return MultisigRow(
        moniker=host_validator.moniker if host_validator else address,
        address=address,
        actual=actual,
        bond_status=host_validator.status if host_validator else None,
        jailed=host_validator.jailed if host_validator else None,
    )



def _entries_for(unbonding: dict[str, int] | None, address: str) -> int | None:
    # A failed lookup is null; a successful one that omits the validator means zero entries.
    return None if unbonding is None else unbonding.get(address, 0)


def _decimal_text(value: Decimal) -> str:
    """Plain (never exponent) notation; an exact zero is just "0"."""
    return "0" if value == 0 else format(value, "f")


def _optional_payload(candidate: LiveTestCandidate | None) -> dict[str, Any] | None:
    return None if candidate is None else candidate.payload()


def _optional_decimal_text(value: Decimal | None) -> str | None:
    return None if value is None else _decimal_text(value=value)


def _divide(numerator: Decimal, denominator: Decimal, places: Decimal) -> Decimal:
    return CONTEXT.divide(numerator, denominator).quantize(places, context=CONTEXT)


def _percent(part: int, whole: int, places: Decimal) -> Decimal:
    return _divide(numerator=CONTEXT.multiply(Decimal(part), Decimal(100)), denominator=Decimal(whole), places=places)
