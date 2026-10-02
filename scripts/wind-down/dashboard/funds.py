"""Collector for the Funds flow tab: where each zone's backing is, from staked on the host to the Osmosis pools.

Every integer in the payload is serialised as a string: 18-decimal amounts exceed 2^53, and the page does its
arithmetic in BigInt. Counts are strings too, for uniformity.
"""

import base64
import concurrent.futures
import dataclasses
import datetime
import decimal
import functools
import json
from dataclasses import dataclass
from decimal import Decimal
from enum import StrEnum
from typing import Any

import chain
import config

CELESTIA_CHAIN_ID = "celestia"
STRD_DENOM = "ustrd"
STRD_SYMBOL = "STRD"
STRD_DECIMALS = 6
# Osmosis's end of Stride's channel-5: a stToken that arrives over it is the canonical stToken denom on Osmosis.
STRIDE_CHANNEL_ON_OSMOSIS = "channel-326"
ST_PREFIX = "st"
IBC_PREFIX = "ibc/"
FACTORY_PREFIX = "factory/"
ALLOYED_INFIX = "/alloyed/"

ZONE_WORKERS = 16
POOL_WORKERS = 8
TRANSFERS_PER_ICA = 100  # newest tx_search hits read per ICA; in-flight transfers time out after a day so they are recent
RATE_PLACES = Decimal("0.000001")
DECIMAL_CONTEXT = decimal.Context(
    prec=60
)  # an 18-decimal supply times the rate needs ~46 digits exactly


class IcaType(StrEnum):
    DELEGATION = "DELEGATION"
    WITHDRAWAL = "WITHDRAWAL"
    FEE = "FEE"
    REDEMPTION = "REDEMPTION"


class StakeSource(StrEnum):
    ICA = "ica"  # the zone's delegation ICA
    MULTISIG = "multisig"  # the staketia multisig on Celestia


class PoolKind(StrEnum):
    CANONICAL = "canonical"
    ROUTE = "route"


class TransferStatus(StrEnum):
    IN_FLIGHT = "in flight"
    SETTLED = "settled"


class ChainName(StrEnum):
    STRIDE = "stride"
    OSMOSIS = "osmosis"


# ---- payload


@dataclass(frozen=True)
class Stages:
    """The six stages of the bar, in native base units. None where the source could not be read."""

    staked: int
    unbonding: int
    liquid: int
    in_flight: int | None
    vault: int | None
    pools: int | None


@dataclass(frozen=True)
class Balance:
    denom: str
    amount: int


@dataclass(frozen=True)
class Account:
    """One row of the accounts table. Amounts are in `symbol` with `decimals`; None renders as a dash."""

    name: str
    chain: str
    address: str
    liquid: int | None
    staked: int | None
    unbonding: int | None
    symbol: str
    decimals: int
    note: str
    other_balances: list[
        Balance
    ]  # non-zero balances in other denoms, listed in the note


@dataclass(frozen=True)
class Transfer:
    """One ICA transfer to the Osmosis vault found in the host's tx index."""

    time: str
    height: int
    ica: IcaType
    sequence: int
    amount: int
    denom: (
        str  # as it left the host: the host denom, or a voucher such as the dYdX USDC
    )
    status: TransferStatus


@dataclass(frozen=True)
class Pool:
    """A transmuter pool assigned to a zone, read from its bank balances."""

    contract: str
    kind: PoolKind
    alloyed_denom: str
    native: int
    st_denom: str
    st_amount: int


@dataclass(frozen=True)
class UnbondingEntry:
    amount: int
    completion: str  # ISO timestamp


@dataclass(frozen=True)
class ValidatorPosition:
    """One validator's staked amount and unbonding entries for one delegator, for the breakdown under the diagram."""

    address: str
    moniker: str  # Stride's name for the validator, or the address when Stride does not track it
    source: StakeSource
    staked: int
    entries: list[UnbondingEntry]


@dataclass(frozen=True)
class Staketia:
    """The staketia lane of the celestia diagram: the multisig on Celestia and the claim address on Stride."""

    multisig_address: str
    staked: int
    unbonding: int
    unbonding_entries: int
    unbonding_earliest: str | None
    unbonding_latest: str | None
    liquid: int
    claim_address: str
    claim_balance: int
    unbonding_records: int | None
    per_validator: list[ValidatorPosition] = dataclasses.field(default_factory=list)


@dataclass(frozen=True)
class ZoneFunds:
    chain_id: str
    symbol: str
    decimals: int
    host_denom: str
    st_denom: str
    osmosis_denom: (
        str | None
    )  # the zone's native denom on Osmosis; None when the Osmosis channel could not be read
    host_channel: (
        str | None
    )  # host -> Osmosis transfer channel; None for osmosis-1 (bank send)
    osmosis_channel: str | None
    vault_address: str
    redemption_rate: str
    st_supply: int
    needed: int  # st supply x rate
    covered: int | None  # vault + pools' native + pools' stTokens x rate
    coverage: (
        str | None
    )  # covered / needed, six places; None when needed is zero or Osmosis could not be read
    stages: Stages
    unbonding_entries: int
    unbonding_earliest: str | None
    unbonding_latest: str | None
    validators: int  # validators the delegation ICA has a delegation with
    ica_liquid: dict[IcaType, int]  # each ICA's host-denom balance, for the diagram
    redemption_ica_balance: int
    open_redemption_records: int | None
    transfers: list[Transfer] | None  # None when the tx index lookup failed
    pools: list[Pool] | None  # None when Osmosis could not be read
    accounts: list[Account]
    staketia: Staketia | None
    validator_positions: list[ValidatorPosition]  # ICA then multisig, largest stake first, zero/zero omitted


# ---- internal structures


@dataclass(frozen=True)
class UnbondingSummary:
    amount: int
    entries: int
    earliest: str | None
    latest: str | None


@dataclass(frozen=True)
class StakingPosition:
    staked: int
    validators: int
    unbonding: UnbondingSummary
    per_validator: list[ValidatorPosition] = dataclasses.field(default_factory=list)


@dataclass(frozen=True)
class SentTransfer:
    """A transfer as the send_packet event records it, before its status is known."""

    height: int
    ica: IcaType
    sequence: int
    amount: int
    denom: str


@dataclass(frozen=True)
class AccountSpec:
    """An address to read for the accounts table, with the denom that is its liquid balance."""

    name: str
    chain_name: str
    chain_handle: chain.Chain
    address: str
    denom: str
    note: str


@dataclass(frozen=True)
class StaketiaRead:
    position: Staketia
    accounts: list[Account]


@dataclass(frozen=True)
class RawPool:
    """A transmuter contract as read from Osmosis, before it is assigned to a zone."""

    contract: str
    alloyed_denom: str
    assets: list[
        str
    ]  # the pool's asset denoms from its config, present even at zero balance
    balances: dict[str, int]
    base_denoms: dict[str, str]  # ibc asset denom -> base denom of its trace


@dataclass(frozen=True)
class OsmosisSnapshot:
    vault_balances: dict[str, int]
    pools: list[RawPool]


@dataclass(frozen=True)
class ZoneDenoms:
    chain_id: str
    osmosis_denom: str
    st_denom: str
    canonical_st_denom: str


@dataclass(frozen=True)
class HostSide:
    """Everything a zone's boundary reads before the Osmosis snapshot is merged in."""

    zone: config.ZoneConfig
    host_zone: dict[str, Any]
    st_supply: int
    osmosis_channel: str | None
    ica_balances: dict[IcaType, dict[str, int]]
    position: StakingPosition
    transfers: list[Transfer] | None
    deposit_balance: int
    staketia: Staketia | None
    staketia_accounts: list[Account]


def collect() -> dict[str, Any]:
    stride = chain.stride_chain()
    host_zones = {
        host_zone["chain_id"]: host_zone
        for host_zone in chain.rest_get(
            chain=stride, path="/Stride-Labs/stride/stakeibc/host_zone"
        )["host_zone"]
    }

    # The host sides, the Osmosis snapshot, the redemption records and the operators are independent.
    with concurrent.futures.ThreadPoolExecutor(max_workers=ZONE_WORKERS) as pool:
        host_futures = [
            pool.submit(
                _collect_host_side,
                zone=zone,
                host_zone=host_zones[zone.chain_id],
                stride=stride,
            )
            for zone in config.ZONES
        ]
        osmosis_future = pool.submit(lambda: chain.optional(_osmosis_snapshot))
        records_future = pool.submit(
            lambda: chain.optional(lambda: _open_redemption_counts(stride=stride))
        )
        operator_futures = [
            pool.submit(_operator_account, stride=stride, name=name, address=address)
            for name, address in (
                ("Protocol admin", config.PROTOCOL_ADMIN),
                ("Gov", config.GOV),
                ("Sweep operator", config.SWEEP_OPERATOR),
            )
        ]
        host_sides = [future.result() for future in host_futures]
        osmosis = osmosis_future.result()
        open_records = records_future.result()
        operators = [future.result() for future in operator_futures]

    # Pools can only be assigned once every zone's Osmosis denom is known.
    succeeded = chain.succeeded(host_sides, HostSide)
    zone_denoms = [denoms for denoms in map(_zone_denoms, succeeded) if denoms]
    pools_by_zone = (
        classify_pools(pools=osmosis.pools, zones=zone_denoms) if osmosis else None
    )
    zones = [
        side
        if isinstance(side, chain.ZoneError)
        else _zone_funds(
            side=side,
            osmosis=osmosis,
            pools_by_zone=pools_by_zone,
            open_records=open_records,
        )
        for side in host_sides
    ]
    return _stringify_ints(
        {
            "zones": [dataclasses.asdict(zone) for zone in zones],
            "operators": [dataclasses.asdict(operator) for operator in operators],
        }
    )


# ---- pure logic


def needed_amount(st_supply: int, redemption_rate: str) -> int:
    """Native tokens the stToken supply redeems for at the rate, truncated like the chain does."""
    return int(DECIMAL_CONTEXT.multiply(Decimal(st_supply), Decimal(redemption_rate)))


def covered_amount(vault: int, pools: list[Pool], redemption_rate: str) -> int:
    """What already backs the stToken on Osmosis: the vault, the pools' native tokens, and the stTokens swapped
    into the pools at the rate, since a swapped-in stToken has already received its backing."""
    st_in_pools = sum(pool.st_amount for pool in pools)
    return (
        vault
        + sum(pool.native for pool in pools)
        + needed_amount(st_supply=st_in_pools, redemption_rate=redemption_rate)
    )


def coverage_ratio(covered: int, needed: int) -> str | None:
    if needed == 0:
        return None
    return str(
        DECIMAL_CONTEXT.quantize(
            DECIMAL_CONTEXT.divide(Decimal(covered), Decimal(needed)), RATE_PLACES
        )
    )


def build_stages(
    position: StakingPosition,
    ica_balances: dict[IcaType, dict[str, int]],
    host_denom: str,
    transfers: list[Transfer] | None,
    vault: int | None,
    pools: list[Pool] | None,
    staketia: Staketia | None,
) -> Stages:
    """The six stages: the delegation ICA's position and the three ICAs' liquid balances, plus the staketia
    multisig and claim address for celestia; in flight, vault and pools are None when their source failed."""
    liquid_icas = (IcaType.DELEGATION, IcaType.WITHDRAWAL, IcaType.FEE)
    liquid = sum(ica_balances[ica].get(host_denom, 0) for ica in liquid_icas)
    in_flight = (
        None
        if transfers is None
        else sum(
            transfer.amount
            for transfer in transfers
            if transfer.status == TransferStatus.IN_FLIGHT
            and transfer.denom == host_denom
        )
    )
    return Stages(
        staked=position.staked + (staketia.staked if staketia else 0),
        unbonding=position.unbonding.amount + (staketia.unbonding if staketia else 0),
        liquid=liquid + (staketia.liquid + staketia.claim_balance if staketia else 0),
        in_flight=in_flight,
        vault=vault,
        pools=None if pools is None else sum(pool.native for pool in pools),
    )


def validator_positions(
    delegations: list[dict[str, Any]],
    unbonding_responses: list[dict[str, Any]],
    names: dict[str, str],
    source: StakeSource,
) -> list[ValidatorPosition]:
    """One row per validator the delegator has stake or unbonding entries on, largest stake first."""
    staked = {
        delegation["delegation"]["validator_address"]: int(delegation["balance"]["amount"])
        for delegation in delegations
    }
    entries = {
        response["validator_address"]: [
            UnbondingEntry(
                amount=int(entry["balance"]),
                completion=chain.parse_timestamp(timestamp=entry["completion_time"]).isoformat(),
            )
            for entry in response["entries"]
        ]
        for response in unbonding_responses
    }
    rows = [
        ValidatorPosition(
            address=address,
            moniker=names.get(address, address),
            source=source,
            staked=staked.get(address, 0),
            entries=sorted(entries.get(address, []), key=lambda entry: entry.completion),
        )
        for address in staked.keys() | entries.keys()
    ]
    return sorted(
        (row for row in rows if row.staked or row.entries),
        key=lambda row: (-row.staked, row.address),
    )


def summarize_unbonding(unbonding_responses: list[dict[str, Any]]) -> UnbondingSummary:
    """Total, entry count, and the earliest and latest completion time across every validator's entries."""
    entries = [
        entry for response in unbonding_responses for entry in response["entries"]
    ]
    completions = sorted(
        chain.parse_timestamp(timestamp=entry["completion_time"]) for entry in entries
    )
    return UnbondingSummary(
        amount=sum(int(entry["balance"]) for entry in entries),
        entries=len(entries),
        earliest=completions[0].isoformat() if completions else None,
        latest=completions[-1].isoformat() if completions else None,
    )


def parse_sent_transfers(
    txs: list[dict[str, Any]], ica: IcaType, sender: str, receiver: str, channel: str
) -> list[SentTransfer]:
    """The ICS-20 sends in these txs from `sender` to `receiver` over `channel`, from their send_packet events.

    The packet data carries the sender, receiver, amount and denom, so one event is enough; a relayer tx may carry
    several ICA packets and so several sends.
    """
    return [
        SentTransfer(
            height=int(tx["height"]),
            ica=ica,
            sequence=int(attributes["packet_sequence"]),
            amount=int(packet["amount"]),
            denom=packet["denom"],
        )
        for tx in txs
        for attributes, packet in _send_packets(tx=tx)
        if attributes["packet_src_port"] == chain.TRANSFER_PORT
        and attributes["packet_src_channel"] == channel
        and packet["sender"] == sender
        and packet["receiver"] == receiver
    ]


def mark_transfers(
    sent: list[SentTransfer], committed: set[int], times: dict[int, str]
) -> list[Transfer]:
    """Attach status and block time: a transfer is in flight while the host still holds its packet commitment."""
    return [
        Transfer(
            time=times[transfer.height],
            height=transfer.height,
            ica=transfer.ica,
            sequence=transfer.sequence,
            amount=transfer.amount,
            denom=transfer.denom,
            status=TransferStatus.IN_FLIGHT
            if transfer.sequence in committed
            else TransferStatus.SETTLED,
        )
        for transfer in sorted(
            sent,
            key=lambda transfer: (transfer.height, transfer.sequence),
            reverse=True,
        )
    ]


def discover_pool_contracts(
    vault_balances: dict[str, int], extra: tuple[str, ...]
) -> list[str]:
    """Transmuter contracts the vault holds an alloyed LP receipt of (`factory/<contract>/alloyed/...`), plus extras."""
    joined = [
        denom.split("/")[1]
        for denom in vault_balances
        if denom.startswith(FACTORY_PREFIX) and ALLOYED_INFIX in denom
    ]
    return list(dict.fromkeys(joined + list(extra)))


def classify_pools(
    pools: list[RawPool], zones: list[ZoneDenoms]
) -> dict[str, list[Pool]]:
    """Assign each pool to the zone whose native denom it holds, or failing that whose stToken it holds (a drained
    pool loses its native asset), and mark it canonical when its stToken came over Stride's own channel."""
    assigned: dict[str, list[Pool]] = {zone.chain_id: [] for zone in zones}
    for pool in pools:
        zone = _pool_zone(pool=pool, zones=zones)
        if zone is None:
            continue

        st_denoms = [
            denom
            for denom in pool.assets
            if pool.base_denoms.get(denom) == zone.st_denom
        ]
        assigned[zone.chain_id].append(
            Pool(
                contract=pool.contract,
                kind=PoolKind.CANONICAL
                if zone.canonical_st_denom in st_denoms
                else PoolKind.ROUTE,
                alloyed_denom=pool.alloyed_denom,
                native=pool.balances.get(zone.osmosis_denom, 0),
                st_denom=", ".join(st_denoms),
                st_amount=sum(pool.balances.get(denom, 0) for denom in st_denoms),
            )
        )
    return assigned


def canonical_st_denom(st_denom: str) -> str:
    return chain.ibc_denom(
        path=f"{chain.TRANSFER_PORT}/{STRIDE_CHANNEL_ON_OSMOSIS}/{st_denom}"
    )


def _pool_zone(pool: RawPool, zones: list[ZoneDenoms]) -> ZoneDenoms | None:
    by_native = [zone for zone in zones if zone.osmosis_denom in pool.assets]
    if by_native:
        return by_native[0]

    base_denoms = set(pool.base_denoms.values())
    return next((zone for zone in zones if zone.st_denom in base_denoms), None)


def _send_packets(tx: dict[str, Any]) -> list[tuple[dict[str, str], dict[str, Any]]]:
    """Each send_packet event's attributes with its decoded ICS-20 packet data."""
    events = [
        {attribute["key"]: attribute["value"] for attribute in event["attributes"]}
        for event in tx["tx_result"]["events"]
        if event["type"] == "send_packet"
    ]
    return [
        (attributes, json.loads(bytes.fromhex(attributes["packet_data_hex"])))
        for attributes in events
        if attributes.get("packet_src_port") == chain.TRANSFER_PORT
    ]


# ---- per zone: host side


def _collect_host_side(
    zone: config.ZoneConfig, host_zone: dict[str, Any], stride: chain.Chain
) -> HostSide | chain.ZoneError:
    return chain.within_error_boundary(
        chain_id=zone.chain_id,
        work=lambda: _host_side(zone=zone, host_zone=host_zone, stride=stride),
    )


def _host_side(
    zone: config.ZoneConfig, host_zone: dict[str, Any], stride: chain.Chain
) -> HostSide:
    host = chain.zone_chain(zone=zone)
    icas = _ica_addresses(host_zone=host_zone)

    # The Osmosis-side channel id names the zone's native denom on Osmosis; osmosis-1 has no leg.
    osmosis_channel = (
        chain.ibc_channel_end(
            chain=host, channel_id=zone.osmosis_channel, port_id=chain.TRANSFER_PORT
        ).counterparty_channel
        if zone.osmosis_channel
        else None
    )

    # The tx index is optional: a host whose index does not hold our sends shows in flight as n/a.
    transfers = (
        chain.optional(
            lambda: _vault_transfers(host=host, channel=zone.osmosis_channel, icas=icas)
        )
        if zone.osmosis_channel
        else []
    )

    staketia = (
        _staketia(
            stride=stride, host=host, zone=zone, names=_validator_names(host_zone=host_zone)
        )
        if zone.chain_id == CELESTIA_CHAIN_ID
        else None
    )
    return HostSide(
        zone=zone,
        host_zone=host_zone,
        st_supply=_supply(chain_handle=stride, denom=_st_denom(host_zone=host_zone)),
        osmosis_channel=osmosis_channel,
        ica_balances={
            ica: _balances(chain_handle=host, address=address)
            for ica, address in icas.items()
        },
        position=_staking_position(
            host=host,
            delegator=icas[IcaType.DELEGATION],
            names=_validator_names(host_zone=host_zone),
            source=StakeSource.ICA,
        ),
        transfers=transfers,
        deposit_balance=_balances(
            chain_handle=stride, address=host_zone["deposit_address"]
        ).get(host_zone["ibc_denom"], 0),
        staketia=staketia.position if staketia else None,
        staketia_accounts=staketia.accounts if staketia else [],
    )


def _staking_position(
    host: chain.Chain, delegator: str, names: dict[str, str], source: StakeSource
) -> StakingPosition:
    delegations = chain.rest_get_all_pages(
        chain=host,
        path=f"/cosmos/staking/v1beta1/delegations/{delegator}",
        key="delegation_responses",
    )
    unbonding = chain.rest_get_all_pages(
        chain=host,
        path=f"/cosmos/staking/v1beta1/delegators/{delegator}/unbonding_delegations",
        key="unbonding_responses",
    )
    balances = [int(delegation["balance"]["amount"]) for delegation in delegations]
    return StakingPosition(
        staked=sum(balances),
        validators=sum(1 for balance in balances if balance),
        unbonding=summarize_unbonding(unbonding_responses=unbonding),
        per_validator=validator_positions(
            delegations=delegations, unbonding_responses=unbonding, names=names, source=source
        ),
    )


def _vault_transfers(
    host: chain.Chain, channel: str, icas: dict[IcaType, str]
) -> list[Transfer]:
    """Our ICA transfers to the Osmosis vault over the host -> Osmosis channel, newest first, with their status."""
    sent = [
        transfer
        for ica, address in icas.items()
        for transfer in parse_sent_transfers(
            txs=chain.rpc_tx_search(
                chain=host,
                query=f"ibc_transfer.sender='{address}' AND ibc_transfer.receiver='{config.OSMOSIS_VAULT}'",
                per_page=TRANSFERS_PER_ICA,
            ),
            ica=ica,
            sender=address,
            receiver=config.OSMOSIS_VAULT,
            channel=channel,
        )
    ]
    if not sent:
        return []

    committed = _unreceived_acks(
        host=host, channel=channel, sequences=[transfer.sequence for transfer in sent]
    )
    times = {
        transfer.height: chain.block_time(rpc=host.rpc, height=transfer.height)
        for transfer in sent
    }
    return mark_transfers(sent=sent, committed=committed, times=times)


def _unreceived_acks(host: chain.Chain, channel: str, sequences: list[int]) -> set[int]:
    """Of `sequences`, the ones the host still holds a packet commitment for: not yet acknowledged or timed out."""
    return set(
        chain.ibc_unreceived_acks(chain=host, channel_id=channel, port_id=chain.TRANSFER_PORT, sequences=sequences)
    )


def _staketia(
    stride: chain.Chain, host: chain.Chain, zone: config.ZoneConfig, names: dict[str, str]
) -> StaketiaRead:
    """The staketia multisig on Celestia and its Stride-side addresses; only the multisig and the claim address
    feed the stages, the rest are shown in the accounts table."""
    staketia_zone = chain.rest_get(
        chain=stride, path="/Stride-Labs/stride/staketia/host_zone"
    )["host_zone"]
    multisig = staketia_zone["delegation_address"]
    voucher = staketia_zone["native_token_ibc_denom"]
    native = staketia_zone["native_token_denom"]

    position = _staking_position(
        host=host, delegator=multisig, names=names, source=StakeSource.MULTISIG
    )
    multisig_balances = _balances(chain_handle=host, address=multisig)
    claim_balance = _balances(
        chain_handle=stride, address=staketia_zone["claim_address"]
    ).get(voucher, 0)
    unbonding_records = chain.optional(
        lambda: len(
            chain.rest_get(
                chain=stride, path="/Stride-Labs/stride/staketia/unbonding_records"
            )["unbonding_records"]
        )
    )

    multisig_account = _account(
        name="Staketia multisig",
        chain_name=host.chain_id,
        address=multisig,
        balances=multisig_balances,
        denom=native,
        zone=zone,
        note=f"{position.validators} validators · {position.unbonding.entries} unbonding entries",
        position=position,
    )
    side_specs = [
        AccountSpec(
            name="Staketia reward address",
            chain_name=host.chain_id,
            chain_handle=host,
            address=staketia_zone["reward_address"],
            denom=native,
            note="rewards land here",
        ),
        AccountSpec(
            name="Staketia deposit address",
            chain_name=ChainName.STRIDE,
            chain_handle=stride,
            address=staketia_zone["deposit_address"],
            denom=voucher,
            note="TIA voucher",
        ),
        AccountSpec(
            name="Staketia redemption address",
            chain_name=ChainName.STRIDE,
            chain_handle=stride,
            address=staketia_zone["redemption_address"],
            denom=voucher,
            note="TIA voucher",
        ),
        AccountSpec(
            name="Staketia claim address",
            chain_name=ChainName.STRIDE,
            chain_handle=stride,
            address=staketia_zone["claim_address"],
            denom=voucher,
            note="TIA voucher · MsgTransferStaketiaClaimBalance sends it to the delegation ICA",
        ),
    ]
    side_accounts = [
        _account(
            name=spec.name,
            chain_name=spec.chain_name,
            address=spec.address,
            balances=_balances(chain_handle=spec.chain_handle, address=spec.address),
            denom=spec.denom,
            zone=zone,
            note=spec.note,
        )
        for spec in side_specs
    ]
    return StaketiaRead(
        position=Staketia(
            multisig_address=multisig,
            staked=position.staked,
            unbonding=position.unbonding.amount,
            unbonding_entries=position.unbonding.entries,
            unbonding_earliest=position.unbonding.earliest,
            unbonding_latest=position.unbonding.latest,
            liquid=multisig_balances.get(native, 0),
            claim_address=staketia_zone["claim_address"],
            claim_balance=claim_balance,
            unbonding_records=unbonding_records,
            per_validator=position.per_validator,
        ),
        accounts=[multisig_account] + side_accounts,
    )


# ---- once per collect: Osmosis, redemption records, operators


def _osmosis_snapshot() -> OsmosisSnapshot:
    osmosis = chain.zone_chain(zone=config.ZONES_BY_CHAIN_ID[config.OSMOSIS_CHAIN_ID])
    vault_balances = _balances(chain_handle=osmosis, address=config.OSMOSIS_VAULT)
    contracts = discover_pool_contracts(
        vault_balances=vault_balances, extra=config.EXTRA_POOL_CONTRACTS
    )
    with concurrent.futures.ThreadPoolExecutor(max_workers=POOL_WORKERS) as pool:
        pools = list(
            pool.map(
                lambda contract: _raw_pool(osmosis=osmosis, contract=contract),
                contracts,
            )
        )
    return OsmosisSnapshot(vault_balances=vault_balances, pools=pools)


def _raw_pool(osmosis: chain.Chain, contract: str) -> RawPool:
    """A transmuter's asset list (so its stToken is known before anything is swapped in) and bank balances."""
    configs = _smart_query(
        osmosis=osmosis, contract=contract, query={"list_asset_configs": {}}
    )["asset_configs"]
    assets = [entry["denom"] for entry in configs]
    return RawPool(
        contract=contract,
        alloyed_denom=_smart_query(
            osmosis=osmosis, contract=contract, query={"get_share_denom": {}}
        )["share_denom"],
        assets=assets,
        balances=_balances(chain_handle=osmosis, address=contract),
        base_denoms={
            denom: _trace_base_denom(rest=osmosis.rest, denom=denom)
            for denom in assets
            if denom.startswith(IBC_PREFIX)
        },
    )


def _smart_query(
    osmosis: chain.Chain, contract: str, query: dict[str, Any]
) -> dict[str, Any]:
    encoded = base64.b64encode(json.dumps(query).encode()).decode()
    return chain.rest_get(
        chain=osmosis, path=f"/cosmwasm/wasm/v1/contract/{contract}/smart/{encoded}"
    )["data"]


@functools.lru_cache(maxsize=1024)
def _trace_base_denom(rest: str, denom: str) -> str:
    """Base denom of an `ibc/` voucher; a trace never changes, so a hit is cached for the life of the process."""
    trace = chain.get_json(
        url=f"{rest}/ibc/apps/transfer/v1/denom_traces/{denom.removeprefix(IBC_PREFIX)}"
    )
    return trace["denom_trace"]["base_denom"]


def _open_redemption_counts(stride: chain.Chain) -> dict[str, int]:
    """Open user redemption records per zone; a record is deleted once claimed, so every record is open."""
    records = chain.rest_get_all_pages(
        chain=stride,
        path="/Stride-Labs/stride/records/user_redemption_record",
        key="user_redemption_record",
    )
    counts: dict[str, int] = {}
    for record in records:
        counts[record["host_zone_id"]] = counts.get(record["host_zone_id"], 0) + 1
    return counts


def _operator_account(stride: chain.Chain, name: str, address: str) -> Account:
    balances = chain.optional(lambda: _balances(chain_handle=stride, address=address))
    return Account(
        name=name,
        chain=ChainName.STRIDE,
        address=address,
        liquid=None if balances is None else balances.get(STRD_DENOM, 0),
        staked=None,
        unbonding=None,
        symbol=STRD_SYMBOL,
        decimals=STRD_DECIMALS,
        note="fees",
        other_balances=[],
    )


# ---- assembly


def _zone_funds(
    side: HostSide,
    osmosis: OsmosisSnapshot | None,
    pools_by_zone: dict[str, list[Pool]] | None,
    open_records: dict[str, int] | None,
) -> ZoneFunds:
    zone = side.zone
    host_zone = side.host_zone
    host_denom = host_zone["host_denom"]
    st_denom = _st_denom(host_zone=host_zone)
    rate = host_zone["redemption_rate"]

    # Osmosis-side values are None when the snapshot failed or the zone's denom there is unknown.
    denoms = _zone_denoms(host_side=side)
    vault = (
        osmosis.vault_balances.get(denoms.osmosis_denom, 0)
        if osmosis and denoms
        else None
    )
    pools = (
        pools_by_zone.get(denoms.chain_id, [])
        if pools_by_zone is not None and denoms
        else None
    )
    covered = (
        covered_amount(vault=vault, pools=pools, redemption_rate=rate)
        if vault is not None and pools is not None
        else None
    )
    needed = needed_amount(st_supply=side.st_supply, redemption_rate=rate)

    stages = build_stages(
        position=side.position,
        ica_balances=side.ica_balances,
        host_denom=host_denom,
        transfers=side.transfers,
        vault=vault,
        pools=pools,
        staketia=side.staketia,
    )
    record_count = (
        open_records.get(zone.chain_id, 0) if open_records is not None else None
    )
    unbonding_times = [
        time
        for time in (
            side.position.unbonding.earliest,
            side.position.unbonding.latest,
            side.staketia.unbonding_earliest if side.staketia else None,
            side.staketia.unbonding_latest if side.staketia else None,
        )
        if time
    ]
    return ZoneFunds(
        chain_id=zone.chain_id,
        symbol=zone.symbol,
        decimals=zone.decimals,
        host_denom=host_denom,
        st_denom=st_denom,
        osmosis_denom=denoms.osmosis_denom if denoms else None,
        host_channel=zone.osmosis_channel,
        osmosis_channel=side.osmosis_channel,
        vault_address=config.OSMOSIS_VAULT,
        redemption_rate=rate,
        st_supply=side.st_supply,
        needed=needed,
        covered=covered,
        coverage=coverage_ratio(covered=covered, needed=needed)
        if covered is not None
        else None,
        stages=stages,
        unbonding_entries=side.position.unbonding.entries
        + (side.staketia.unbonding_entries if side.staketia else 0),
        unbonding_earliest=min(unbonding_times, key=_parse_iso, default=None),
        unbonding_latest=max(unbonding_times, key=_parse_iso, default=None),
        validators=side.position.validators,
        ica_liquid={
            ica: balances.get(host_denom, 0)
            for ica, balances in side.ica_balances.items()
        },
        redemption_ica_balance=side.ica_balances[IcaType.REDEMPTION].get(host_denom, 0),
        open_redemption_records=record_count,
        transfers=side.transfers,
        pools=pools,
        accounts=_zone_accounts(
            side=side, vault=vault, pools=pools, record_count=record_count
        ),
        staketia=side.staketia,
        validator_positions=side.position.per_validator
        + (side.staketia.per_validator if side.staketia else []),
    )


def _validator_names(host_zone: dict[str, Any]) -> dict[str, str]:
    return {validator["address"]: validator["name"] for validator in host_zone["validators"]}


def _zone_accounts(
    side: HostSide,
    vault: int | None,
    pools: list[Pool] | None,
    record_count: int | None,
) -> list[Account]:
    zone = side.zone
    host_zone = side.host_zone
    host_denom = host_zone["host_denom"]
    icas = _ica_addresses(host_zone=host_zone)
    position = side.position
    ica_notes = {
        IcaType.DELEGATION: f"{position.validators} validators · {position.unbonding.entries} unbonding entries",
        IcaType.WITHDRAWAL: "rewards land here",
        IcaType.FEE: "",
        IcaType.REDEMPTION: "n/a open user redemption records"
        if record_count is None
        else f"{record_count} open user redemption records",
    }
    ica_accounts = [
        _account(
            name=f"{ica.capitalize()} ICA",
            chain_name=zone.chain_id,
            address=address,
            balances=side.ica_balances[ica],
            denom=host_denom,
            zone=zone,
            note=ica_notes[ica],
            position=position if ica == IcaType.DELEGATION else None,
        )
        for ica, address in icas.items()
    ]
    deposit = _liquid_account(
        name="Deposit address",
        chain_name=ChainName.STRIDE,
        address=host_zone["deposit_address"],
        liquid=side.deposit_balance,
        zone=zone,
        note=f"{zone.symbol} voucher awaiting stake",
    )
    vault_account = _liquid_account(
        name="Osmosis vault",
        chain_name=ChainName.OSMOSIS,
        address=config.OSMOSIS_VAULT,
        liquid=vault,
        zone=zone,
        note="Osmosis unreachable"
        if pools is None
        else f"holds {len(pools)} alloyed LP receipt(s) for {zone.symbol} pools",
    )
    pool_accounts = [
        _liquid_account(
            name=f"st{zone.symbol} {pool.kind} pool",
            chain_name=ChainName.OSMOSIS,
            address=pool.contract,
            liquid=pool.native,
            zone=zone,
            note=f"st{zone.symbol} swapped in: {_format_amount(amount=pool.st_amount, decimals=zone.decimals)}",
        )
        for pool in pools or []
    ]
    return (
        ica_accounts
        + [deposit]
        + side.staketia_accounts
        + [vault_account]
        + pool_accounts
    )


def _liquid_account(
    name: str,
    chain_name: str,
    address: str,
    liquid: int | None,
    zone: config.ZoneConfig,
    note: str,
) -> Account:
    """An account that only holds the zone's token liquid: no staking, nothing else worth listing."""
    return Account(
        name=name,
        chain=chain_name,
        address=address,
        liquid=liquid,
        staked=None,
        unbonding=None,
        symbol=zone.symbol,
        decimals=zone.decimals,
        note=note,
        other_balances=[],
    )


def _account(
    name: str,
    chain_name: str,
    address: str,
    balances: dict[str, int],
    denom: str,
    zone: config.ZoneConfig,
    note: str,
    position: StakingPosition | None = None,
) -> Account:
    return Account(
        name=name,
        chain=chain_name,
        address=address,
        liquid=balances.get(denom, 0),
        staked=position.staked if position else None,
        unbonding=position.unbonding.amount if position else None,
        symbol=zone.symbol,
        decimals=zone.decimals,
        note=note,
        other_balances=[
            Balance(denom=other, amount=amount)
            for other, amount in sorted(balances.items())
            if other != denom and amount
        ],
    )


def _zone_denoms(host_side: HostSide) -> ZoneDenoms | None:
    """The zone's native and stToken denoms on Osmosis; None when its Osmosis channel could not be read."""
    if (
        host_side.zone.chain_id != config.OSMOSIS_CHAIN_ID
        and not host_side.osmosis_channel
    ):
        return None

    host_denom = host_side.host_zone["host_denom"]
    st_denom = _st_denom(host_zone=host_side.host_zone)
    osmosis_denom = (
        host_denom
        if host_side.zone.chain_id == config.OSMOSIS_CHAIN_ID
        else chain.ibc_denom(
            path=f"{chain.TRANSFER_PORT}/{host_side.osmosis_channel}/{host_denom}"
        )
    )
    return ZoneDenoms(
        chain_id=host_side.zone.chain_id,
        osmosis_denom=osmosis_denom,
        st_denom=st_denom,
        canonical_st_denom=canonical_st_denom(st_denom=st_denom),
    )


# ---- helpers


def _ica_addresses(host_zone: dict[str, Any]) -> dict[IcaType, str]:
    return {ica: host_zone[f"{ica.lower()}_ica_address"] for ica in IcaType}


def _st_denom(host_zone: dict[str, Any]) -> str:
    return f"{ST_PREFIX}{host_zone['host_denom']}"


def _balances(chain_handle: chain.Chain, address: str) -> dict[str, int]:
    balances = chain.rest_get_all_pages(
        chain=chain_handle,
        path=f"/cosmos/bank/v1beta1/balances/{address}",
        key="balances",
    )
    return {balance["denom"]: int(balance["amount"]) for balance in balances}


def _supply(chain_handle: chain.Chain, denom: str) -> int:
    response = chain.rest_get(
        chain=chain_handle,
        path="/cosmos/bank/v1beta1/supply/by_denom",
        params={"denom": denom},
    )
    return int(response["amount"]["amount"])


def _parse_iso(timestamp: str) -> datetime.datetime:
    """Parse the timestamps this module itself emits (isoformat with an offset), for ordering."""
    return datetime.datetime.fromisoformat(timestamp)


def _format_amount(amount: int, decimals: int) -> str:
    """Whole tokens with two places, for notes; the page formats every tabulated amount itself."""
    scale = 10**decimals
    return f"{amount // scale:,}.{(amount % scale) * 100 // scale:02d}"


def _stringify_ints(value: Any) -> Any:
    """Every int (not bool) in a JSON-ready structure as a string, so 18-decimal amounts survive JSON.parse."""
    if isinstance(value, bool):
        return value
    if isinstance(value, int):
        return str(value)
    if isinstance(value, dict):
        return {key: _stringify_ints(item) for key, item in value.items()}
    if isinstance(value, list):
        return [_stringify_ints(item) for item in value]
    return value
