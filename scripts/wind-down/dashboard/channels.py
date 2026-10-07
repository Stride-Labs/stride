"""Collector for the Channels tab: Stride <-> host channels, host -> Osmosis legs, and holder routes into Osmosis."""

import concurrent.futures
import dataclasses
import datetime
from dataclasses import dataclass
from enum import StrEnum
from typing import Any

import chain
import config

ICA_HOST_PORT = "icahost"
TRANSFER_NAME = "transfer"

STATE_OPEN = "OPEN"
STATE_CLOSED = "CLOSED"
STATE_UNKNOWN = "UNREACHABLE"  # the chain that owns this end could not be asked

STAKEIBC_HOST_ZONE_PATH = "/Stride-Labs/stride/stakeibc/host_zone"
MSG_TRANSFER = "/ibc.applications.transfer.v1.MsgTransfer"
MSG_SEND = "/cosmos.bank.v1beta1.MsgSend"
ALLOW_ALL = "*"  # an ICA host allow list that permits every message
BANK_SEND_NOTE = "bank send"

STUCK_AFTER_SECONDS = 30 * 60
ZONE_WORKERS = 16
CHANNEL_WORKERS = 6


class TimeSource(StrEnum):
    """Which tx gave a flow's age."""

    SENT = "sent"  # the send_packet tx on the sender
    RECEIVED = "received"  # the recv_packet tx on the receiver, for acks whose send is not indexed


class Status(StrEnum):
    CLOSED = "closed"
    HANDSHAKE_STUCK = "handshake stuck"
    STUCK = "stuck"
    PENDING = "pending"
    OK = "ok"


# Worst first, for rolling channel statuses up to a zone.
STATUS_SEVERITY = (
    Status.CLOSED,
    Status.HANDSHAKE_STUCK,
    Status.STUCK,
    Status.PENDING,
    Status.OK,
)


@dataclass(frozen=True)
class PacketFlow:
    """Packets one chain has sent on a channel and not yet seen acknowledged.

    Every commitment is either a pending packet (the receiver has not seen it) or a pending ack
    (the receiver has, the sender has not processed the ack). Counts are None when the receiving
    side could not be asked.

    The age comes from the send_packet tx. Epoch-hook sends are not in the tx index, so a pending ack
    with no send time ages from its recv_packet tx on the receiver instead (`oldest_time_source`); a
    pending packet with no send time has no known age.
    """

    pending_packets: int | None
    pending_acks: int | None
    capped: bool  # there are more commitments than chain.COMMITMENT_CAP
    oldest_sequence: (
        int | None
    )  # lowest commitment sequence; None when nothing is pending
    oldest_sent_at: (
        str | None
    )  # block time of the tx named by oldest_time_source
    oldest_time_source: TimeSource | None  # None when oldest_sent_at is None

    def age_seconds(self, now: datetime.datetime) -> float | None:
        if self.oldest_sent_at is None:
            return None
        return (
            now - chain.parse_timestamp(timestamp=self.oldest_sent_at)
        ).total_seconds()


NO_PENDING = PacketFlow(
    pending_packets=0,
    pending_acks=0,
    capped=False,
    oldest_sequence=None,
    oldest_sent_at=None,
    oldest_time_source=None,
)


@dataclass(frozen=True)
class PendingSplit:
    packets: list[int]
    acks: list[int]


@dataclass(frozen=True)
class ChannelRow:
    name: str  # "transfer" or the ICA type (DELEGATION, WITHDRAWAL, ...)
    port_id: str
    stride_channel: str
    host_channel: str | None
    stride_state: str | None
    host_state: str | None
    outbound: PacketFlow  # Stride -> host
    inbound: PacketFlow | None  # host -> Stride, transfer channels only
    last_sent: chain.TxRef | None
    last_received: chain.TxRef | None
    last_ack: chain.TxRef | None
    status: Status


@dataclass(frozen=True)
class Preflight:
    """The checks the ops plan runs before a zone's wind-down, each None (n/a) when its lookup failed.

    The supporting fields carry what the page shows beside each pill.
    """

    withdraw_address_ok: bool | None  # the delegation ICA's reward withdraw address is the withdrawal ICA
    withdraw_address: str | None  # as the host reports it
    withdrawal_ica_address: str | None  # as Stride records it
    delegation_ica_address: str | None
    allow_messages_ok: bool | None  # the host's ICA allow list covers what the wind-down sends
    allow_messages_wildcard: bool | None
    allow_messages_count: int | None
    allow_messages_missing: list[str] | None
    host_enabled: bool | None
    osmosis_leg_ok: bool | None  # the host's Osmosis channel is OPEN on a client of osmosis-1
    osmosis_leg_state: str | None
    osmosis_leg_chain_id: str | None  # the chain the leg's client tracks
    osmosis_leg_note: str | None  # why there is no leg ("bank send" for osmosis-1)
    host_client_of_stride_ok: bool | None  # the host's client of Stride is Active
    host_client_of_stride_status: str | None
    host_client_of_stride_id: str | None
    all_ok: bool  # every check that applies passed (a null check or any failure fails it)


@dataclass(frozen=True)
class IcaHostParams:
    host_enabled: bool
    allow_messages: list[str]


@dataclass(frozen=True)
class OsmosisLeg:
    state: str
    chain_id: str  # the chain the leg's client tracks


@dataclass(frozen=True)
class ZoneChannels:
    chain_id: str
    symbol: str
    stride_client: chain.ClientHealth
    host_client: chain.ClientHealth | None
    stride_connection_state: str
    host_connection_state: str | None
    host_error: str | None  # why the host-side header lookups are null (host REST down), else None
    status: Status
    channels: list[ChannelRow]
    # Commitments on the zone's ICA channels (pending packets + pending acks, Stride side); None when one could not
    # be counted. The v35 flag reset only applies to a zone with an open delegation channel and 0 here.
    unacked_ica_packets: int | None = None
    preflight: Preflight | None = None  # None only when the zone's header could not be built
    # Every ICA row (all but the transfer one) is OPEN on both ends; None when any end's state is unknown.
    ica_channels_open: bool | None = None
    haqq_state: "HaqqState | None" = None  # the pre-upgrade haqq invariants (spec §9c); only on haqq_11235-1


# The haqq channels the pre-upgrade rules name (spec §9c). The delegation channel must stay as it is until the
# upgrade; nothing may be committed on the other ICA channels, or their ordered channels close when the relayer
# returns.
HAQQ_CHAIN_ID = "haqq_11235-1"
HAQQ_DELEGATION_CHANNEL = "channel-869"
HAQQ_DELEGATION_HOST_CHANNEL = "channel-29"
HAQQ_DELEGATION_SEQUENCES = list(range(85, 99))
HAQQ_QUIET_ICA_CHANNELS = {
    "FEE": "channel-614",
    "REDEMPTION": "channel-244",
    "COMMUNITY_POOL_DEPOSIT": "channel-245",
    "COMMUNITY_POOL_RETURN": "channel-246",
}


@dataclass(frozen=True)
class HaqqState:
    """Spec §9c, 'before the upgrade': the delegation channel is the closed channel-869 with packets 85-98 still
    committed, haqq's end (channel-29) is still OPEN so nobody can restore early, and the other ICA channels hold
    no packet commitment."""

    delegation_channel_id: str  # the feed's current DELEGATION channel; a different id means someone restored
    delegation_state_stride: str
    delegation_state_host: str | None
    delegation_commitments: list[int]
    quiet_channel_commitments: dict[str, int]  # ICA type -> commitments on its channel
    ok: bool


def haqq_state_passes(state: "HaqqState") -> bool:
    return (
        state.delegation_channel_id == HAQQ_DELEGATION_CHANNEL
        and state.delegation_state_stride == STATE_CLOSED
        and state.delegation_state_host == STATE_OPEN
        and state.delegation_commitments == HAQQ_DELEGATION_SEQUENCES
        and all(count == 0 for count in state.quiet_channel_commitments.values())
    )


@dataclass(frozen=True)
class HostHeader:
    """The host-side half of a zone header; null parts mean the host could not be asked."""

    client: chain.ClientHealth | None
    connection_state: str | None
    error: str | None


@dataclass(frozen=True)
class LegRow:
    chain_id: str
    symbol: str
    host_channel: str
    osmosis_channel: str | None
    host_state: str
    osmosis_state: str | None
    host_client: chain.ClientHealth
    osmosis_client: chain.ClientHealth | None
    last_received: chain.TxRef | None  # on Osmosis
    last_ack: chain.TxRef | None  # on the host
    status: Status


@dataclass(frozen=True)
class RouteRow:
    chain: str
    osmosis_channel: str
    relayed_by: str
    state: str
    counterparty_channel: str | None
    client: chain.ClientHealth
    last_received: chain.TxRef | None
    last_ack: chain.TxRef | None
    status: Status


@dataclass(frozen=True)
class Tiles:
    channels_open: int
    channels_total: int
    channels_closed: int
    channels_handshake_stuck: int
    pending_packets: int
    pending_packet_channels: int
    pending_acks: int
    pending_ack_channels: int
    pending_capped: bool
    oldest_pending_seconds: float | None
    oldest_pending_where: str | None
    clients_live: int
    clients_total: int
    soonest_expiry_seconds: float | None
    soonest_expiry_where: str | None


@dataclass(frozen=True)
class ChannelSpec:
    """One channel to inspect, before any packet lookups."""

    name: str
    port_id: str
    host_port_id: str
    stride_channel: str
    host_channel: str | None
    stride_state: str | None
    host_state: str | None


def collect() -> dict[str, Any]:
    now = datetime.datetime.now(datetime.UTC)
    feed = chain.get_json(url=config.CHANNELS_FEED_URL)
    feed_by_chain = {entry["chain_id"]: entry for entry in feed}
    # Stride's host zones once for every zone's pre-flight checks; without them those checks are n/a.
    host_zones = chain.optional(_stride_host_zones)

    # Zones, legs and routes are independent, so all of them share one pool.
    with concurrent.futures.ThreadPoolExecutor(max_workers=ZONE_WORKERS) as pool:
        zone_futures = [
            pool.submit(_collect_zone, zone=zone, feed_by_chain=feed_by_chain, host_zones=host_zones, now=now)
            for zone in config.ZONES
        ]
        leg_futures = [
            pool.submit(_collect_leg, zone=zone, now=now)
            for zone in config.ZONES
            if zone.osmosis_channel
        ]
        route_futures = [
            pool.submit(_collect_route, route=route, now=now) for route in config.HOLDER_ROUTES
        ]
        zones = [future.result() for future in zone_futures]
        legs = [future.result() for future in leg_futures]
        routes = [future.result() for future in route_futures]

    tiles = build_tiles(
        zones=chain.succeeded(zones, ZoneChannels),
        legs=chain.succeeded(legs, LegRow),
        routes=chain.succeeded(routes, RouteRow),
        now=now,
    )
    return {
        "zones": [dataclasses.asdict(zone) for zone in zones],
        "legs": [dataclasses.asdict(leg) for leg in legs],
        "routes": [dataclasses.asdict(route) for route in routes],
        "tiles": dataclasses.asdict(tiles),
    }


# ---- pure logic


def split_pending(commitments: list[int], unreceived: list[int]) -> PendingSplit:
    """Split the sender's commitments into pending packets (not received) and pending acks (received)."""
    unreceived_set = set(unreceived)
    return PendingSplit(
        packets=[sequence for sequence in commitments if sequence in unreceived_set],
        acks=[sequence for sequence in commitments if sequence not in unreceived_set],
    )


def channel_status(
    end_states: list[str | None], flows: list[PacketFlow], now: datetime.datetime
) -> Status:
    """Status of a channel from the state of each known end and the pending packets it carries.

    An end whose chain could not be asked (STATE_UNKNOWN) says nothing either way.
    """
    known_states = [state for state in end_states if state != STATE_UNKNOWN]
    if STATE_CLOSED in known_states:
        return Status.CLOSED
    if any(state != STATE_OPEN for state in known_states):
        return Status.HANDSHAKE_STUCK

    pending = [flow for flow in flows if flow.oldest_sequence is not None]
    if not pending:
        return Status.OK

    ages = [
        age
        for age in (flow.age_seconds(now=now) for flow in pending)
        if age is not None
    ]
    if any(age > STUCK_AFTER_SECONDS for age in ages):
        return Status.STUCK
    return Status.PENDING


def worst_status(statuses: list[Status]) -> Status:
    return min(statuses, key=STATUS_SEVERITY.index, default=Status.OK)


def allow_messages_missing(allow_messages: list[str], chain_id: str) -> list[str]:
    """The messages the wind-down sends through the host's ICA that its allow list does not cover.

    Every zone sends MsgTransfer (the vault transfer); osmosis-1 has no IBC leg and pays the vault with a bank
    MsgSend. A `*` entry allows everything.
    """
    if ALLOW_ALL in allow_messages:
        return []
    required = [MSG_SEND] if chain_id == config.OSMOSIS_CHAIN_ID else []
    return [message for message in [MSG_TRANSFER, *required] if message not in allow_messages]


def leg_is_ok(state: str, counterparty_chain_id: str) -> bool:
    return state == STATE_OPEN and counterparty_chain_id == config.OSMOSIS_CHAIN_ID


def preflight_passes(
    chain_id: str,
    withdraw_address_ok: bool | None,
    allow_messages_ok: bool | None,
    osmosis_leg_ok: bool | None,
    host_client_of_stride_ok: bool | None,
) -> bool:
    """Every check passed; osmosis-1 has no leg (a bank send), so its leg being n/a is not a failure."""
    has_leg = chain_id != config.OSMOSIS_CHAIN_ID
    checks = [withdraw_address_ok, allow_messages_ok, host_client_of_stride_ok] + ([osmosis_leg_ok] if has_leg else [])
    return all(check is True for check in checks)


def build_tiles(
    zones: list[ZoneChannels],
    legs: list[LegRow],
    routes: list[RouteRow],
    now: datetime.datetime,
) -> Tiles:
    rows = [(zone.chain_id, row) for zone in zones for row in zone.channels]
    flows = [
        (f"{chain_id} {row.name}", flow)
        for chain_id, row in rows
        for flow in (row.outbound, row.inbound)
        if flow
    ]

    aged_flows = [
        (where, flow.age_seconds(now=now))
        for where, flow in flows
        if flow.oldest_sequence is not None
    ]
    known_ages = [(where, age) for where, age in aged_flows if age is not None]
    oldest_where, oldest_age = max(
        known_ages, key=lambda entry: entry[1], default=(None, None)
    )

    clients = _distinct_clients(zones=zones, legs=legs, routes=routes)
    live_clients = [
        client for client in clients if client.status == chain.CLIENT_STATUS_ACTIVE
    ]
    expiring = [
        client for client in live_clients if client.seconds_remaining is not None
    ]
    soonest = min(expiring, key=lambda client: client.seconds_remaining, default=None)

    return Tiles(
        channels_open=sum(
            1 for _, row in rows if row.stride_state == row.host_state == STATE_OPEN
        ),
        channels_total=len(rows),
        channels_closed=sum(1 for _, row in rows if row.status == Status.CLOSED),
        channels_handshake_stuck=sum(
            1 for _, row in rows if row.status == Status.HANDSHAKE_STUCK
        ),
        pending_packets=sum(flow.pending_packets or 0 for _, flow in flows),
        pending_packet_channels=sum(1 for _, flow in flows if flow.pending_packets),
        pending_acks=sum(flow.pending_acks or 0 for _, flow in flows),
        pending_ack_channels=sum(1 for _, flow in flows if flow.pending_acks),
        pending_capped=any(flow.capped for _, flow in flows),
        oldest_pending_seconds=oldest_age,
        oldest_pending_where=oldest_where,
        clients_live=len(live_clients),
        clients_total=len(clients),
        soonest_expiry_seconds=soonest.seconds_remaining if soonest else None,
        soonest_expiry_where=f"{soonest.chain_id} {soonest.client_id}"
        if soonest
        else None,
    )


# ---- Stride <-> host zones


def _collect_zone(
    zone: config.ZoneConfig,
    feed_by_chain: dict[str, Any],
    host_zones: dict[str, dict[str, Any]] | None,
    now: datetime.datetime,
) -> ZoneChannels | chain.ZoneError:
    return chain.within_error_boundary(
        chain_id=zone.chain_id,
        work=lambda: _zone_channels(zone=zone, feed_by_chain=feed_by_chain, host_zones=host_zones, now=now),
    )


def _zone_channels(
    zone: config.ZoneConfig,
    feed_by_chain: dict[str, Any],
    host_zones: dict[str, dict[str, Any]] | None,
    now: datetime.datetime,
) -> ZoneChannels:
    entry = feed_by_chain[zone.chain_id]
    stride = chain.stride_chain()
    host = chain.zone_chain(zone=zone)

    # Transfer channel first: its ends come from the chains, the ICA ends from the feed.
    ica_channels = entry["ica_channels"] or _ica_channels_from_stride(stride=stride, host=host, entry=entry)
    specs = [_transfer_spec(stride=stride, host=host, entry=entry)] + [_ica_spec(ica) for ica in ica_channels]
    with concurrent.futures.ThreadPoolExecutor(max_workers=CHANNEL_WORKERS) as pool:
        rows = list(pool.map(lambda spec: _channel_row(spec=spec, stride=stride, host=host, now=now), specs))

    host_header = _host_header(host=host, entry=entry)
    return ZoneChannels(
        chain_id=zone.chain_id,
        symbol=zone.symbol,
        stride_client=chain.ibc_client_health(chain=stride, client_id=entry["client_id"]),
        host_client=host_header.client,
        stride_connection_state=chain.ibc_connection_end(chain=stride, connection_id=entry["connection_id"]).state,
        host_connection_state=host_header.connection_state,
        host_error=host_header.error,
        status=worst_status(statuses=[row.status for row in rows]),
        channels=rows,
        unacked_ica_packets=unacked_ica_packets(rows=rows),
        ica_channels_open=ica_channels_open(rows=rows),
        preflight=_preflight(
            zone=zone, host=host, entry=entry, host_zone=(host_zones or {}).get(zone.chain_id)
        ),
        haqq_state=_haqq_state(stride=stride, ica_channels=ica_channels) if zone.chain_id == HAQQ_CHAIN_ID else None,
    )


# The feed drops a zone's ICA channels when its own haqq lookups fail (seen 2026-10-06), so the current channel per
# ICA type is read from Stride's channel list on the zone's connection: the highest channel id per controller port,
# with the host's end asked for its state.
def _ica_channels_from_stride(stride: chain.Chain, host: chain.Chain, entry: dict[str, Any]) -> list[dict[str, Any]]:
    listed = chain.rest_get_all_pages(
        chain=stride, path=f"/ibc/core/channel/v1/connections/{entry['connection_id']}/channels", key="channels"
    )
    controller_prefix = f"icacontroller-{entry['chain_id']}."
    newest: dict[str, dict[str, Any]] = {}
    for channel in listed:
        if not channel["port_id"].startswith(controller_prefix):
            continue
        number = int(channel["channel_id"].rsplit("-", 1)[1])
        current = newest.get(channel["port_id"])
        if current is None or number > int(current["channel_id"].rsplit("-", 1)[1]):
            newest[channel["port_id"]] = channel
    return [
        {
            "type": port_id.removeprefix(controller_prefix),
            "port_id": port_id,
            "channel_id": channel["channel_id"],
            "counterparty_channel_id": channel["counterparty"]["channel_id"],
            "state": channel["state"],
            "counterparty_state": _host_channel_state(host=host, channel_id=channel["counterparty"]["channel_id"]),
        }
        for port_id, channel in sorted(newest.items())
    ]


def _host_channel_state(host: chain.Chain, channel_id: str) -> str:
    if not channel_id:
        return ""
    state = chain.optional(lambda: chain.ibc_channel_end(chain=host, channel_id=channel_id, port_id=ICA_HOST_PORT).state)
    return f"{chain.STATE_PREFIX}{state}" if state else ""


def _haqq_state(stride: chain.Chain, ica_channels: list[dict[str, Any]]) -> HaqqState:
    delegation = next(ica for ica in ica_channels if ica["type"] == "DELEGATION")
    commitments = chain.ibc_packet_commitments(
        chain=stride, channel_id=delegation["channel_id"], port_id=delegation["port_id"]
    ).sequences
    quiet = {
        ica_type: len(
            chain.ibc_packet_commitments(
                chain=stride, channel_id=channel_id, port_id=f"icacontroller-{HAQQ_CHAIN_ID}.{ica_type}"
            ).sequences
        )
        for ica_type, channel_id in HAQQ_QUIET_ICA_CHANNELS.items()
    }
    state = HaqqState(
        delegation_channel_id=delegation["channel_id"],
        delegation_state_stride=chain.strip_state_prefix(state=delegation["state"]),
        delegation_state_host=chain.strip_state_prefix(state=delegation["counterparty_state"]) if delegation.get("counterparty_state") else None,
        delegation_commitments=sorted(commitments),
        quiet_channel_commitments=quiet,
        ok=False,
    )
    return dataclasses.replace(state, ok=haqq_state_passes(state))


# ---- pre-flight checks


def _preflight(
    zone: config.ZoneConfig, host: chain.Chain, entry: dict[str, Any], host_zone: dict[str, Any] | None
) -> Preflight:
    """Each check is its own optional lookup: one dead query makes that pill n/a, not the zone."""
    delegation_ica = host_zone["delegation_ica_address"] if host_zone else None
    withdrawal_ica = host_zone["withdrawal_ica_address"] if host_zone else None
    withdraw_address = chain.optional(
        lambda: _withdraw_address(host=host, delegator=delegation_ica) if delegation_ica else None
    )
    withdraw_address_ok = None if withdraw_address is None or withdrawal_ica is None else withdraw_address == withdrawal_ica

    ica_params = chain.optional(lambda: _ica_host_params(host=host))
    missing = (
        None
        if ica_params is None
        else allow_messages_missing(allow_messages=ica_params.allow_messages, chain_id=zone.chain_id)
    )

    leg = chain.optional(lambda: _osmosis_leg(zone=zone, host=host))
    stride_client_id = entry["counterparty_client_id"]
    client_status = chain.optional(lambda: chain.ibc_client_status(chain=host, client_id=stride_client_id))
    client_ok = None if client_status is None else client_status == chain.CLIENT_STATUS_ACTIVE

    allow_messages_ok = None if missing is None else not missing
    leg_ok = None if leg is None else leg_is_ok(state=leg.state, counterparty_chain_id=leg.chain_id)
    return Preflight(
        withdraw_address_ok=withdraw_address_ok,
        withdraw_address=withdraw_address,
        withdrawal_ica_address=withdrawal_ica,
        delegation_ica_address=delegation_ica,
        allow_messages_ok=allow_messages_ok,
        allow_messages_wildcard=None if ica_params is None else ALLOW_ALL in ica_params.allow_messages,
        allow_messages_count=None if ica_params is None else len(ica_params.allow_messages),
        allow_messages_missing=missing,
        host_enabled=None if ica_params is None else ica_params.host_enabled,
        osmosis_leg_ok=leg_ok,
        osmosis_leg_state=leg.state if leg else None,
        osmosis_leg_chain_id=leg.chain_id if leg else None,
        osmosis_leg_note=None if zone.osmosis_channel else BANK_SEND_NOTE,
        host_client_of_stride_ok=client_ok,
        host_client_of_stride_status=client_status,
        host_client_of_stride_id=stride_client_id,
        all_ok=preflight_passes(
            chain_id=zone.chain_id,
            withdraw_address_ok=withdraw_address_ok,
            allow_messages_ok=allow_messages_ok,
            osmosis_leg_ok=leg_ok,
            host_client_of_stride_ok=client_ok,
        ),
    )


def _withdraw_address(host: chain.Chain, delegator: str) -> str:
    response = chain.rest_get(
        chain=host, path=f"/cosmos/distribution/v1beta1/delegators/{delegator}/withdraw_address"
    )
    return response["withdraw_address"]


def _ica_host_params(host: chain.Chain) -> IcaHostParams:
    params = chain.rest_get(chain=host, path="/ibc/apps/interchain_accounts/host/v1/params")["params"]
    return IcaHostParams(host_enabled=bool(params["host_enabled"]), allow_messages=params["allow_messages"] or [])


def _osmosis_leg(zone: config.ZoneConfig, host: chain.Chain) -> OsmosisLeg | None:
    """The host's channel to Osmosis (the one the leg table shows) and the chain its client tracks."""
    if not zone.osmosis_channel:
        return None

    end = chain.ibc_channel_end(chain=host, channel_id=zone.osmosis_channel, port_id=chain.TRANSFER_PORT)
    connection = chain.ibc_connection_end(chain=host, connection_id=end.connection_id)
    return OsmosisLeg(
        state=end.state, chain_id=chain.ibc_client_chain_id(chain=host, client_id=connection.client_id)
    )


def _stride_host_zones() -> dict[str, dict[str, Any]]:
    response = chain.rest_get(chain=chain.stride_chain(), path=STAKEIBC_HOST_ZONE_PATH)
    return {host_zone["chain_id"]: host_zone for host_zone in response["host_zone"]}


def _host_header(host: chain.Chain, entry: dict[str, Any]) -> HostHeader:
    """Host client and connection. A dead host REST must not hide what Stride and the feed know."""
    try:
        return HostHeader(
            client=chain.ibc_client_health(chain=host, client_id=entry["counterparty_client_id"]),
            connection_state=chain.ibc_connection_end(
                chain=host, connection_id=entry["counterparty_connection_id"]
            ).state,
            error=None,
        )
    except chain.ZONE_ERRORS as error:
        return HostHeader(client=None, connection_state=None, error=f"{type(error).__name__}: {error}")


def _transfer_spec(
    stride: chain.Chain, host: chain.Chain, entry: dict[str, Any]
) -> ChannelSpec:
    stride_channel = entry["transfer_channel_id"]
    stride_end = chain.ibc_channel_end(
        chain=stride, channel_id=stride_channel, port_id=chain.TRANSFER_PORT
    )
    host_channel = stride_end.counterparty_channel
    host_end = chain.optional(
        lambda: chain.ibc_channel_end(chain=host, channel_id=host_channel, port_id=chain.TRANSFER_PORT)
        if host_channel
        else None
    )
    # No counterparty channel yet is a real state (None); a host that did not answer is not.
    host_state = host_end.state if host_end else (STATE_UNKNOWN if host_channel else None)
    return ChannelSpec(
        name=TRANSFER_NAME,
        port_id=chain.TRANSFER_PORT,
        host_port_id=chain.TRANSFER_PORT,
        stride_channel=stride_channel,
        host_channel=host_channel,
        stride_state=stride_end.state,
        host_state=host_state,
    )


def ica_channels_open(rows: list[ChannelRow]) -> bool | None:
    """Every ICA row OPEN on both ends (a closed ordered channel rejects transfer-from-ica); None when any end's state
    is unknown or the zone has no ICA rows to judge."""
    states = [state for row in rows if row.name != TRANSFER_NAME for state in (row.stride_state, row.host_state)]
    if not states or None in states:
        return None
    return all(state == STATE_OPEN for state in states)


def unacked_ica_packets(rows: list[ChannelRow]) -> int | None:
    """Sum of pending packets and pending acks over the ICA rows; None when any count is unknown."""
    counts = [
        (row.outbound.pending_packets, row.outbound.pending_acks)
        for row in rows
        if row.name != TRANSFER_NAME
    ]
    if any(packets is None or acks is None for packets, acks in counts):
        return None
    return sum(packets + acks for packets, acks in counts)


def _ica_spec(ica: dict[str, Any]) -> ChannelSpec:
    return ChannelSpec(
        name=ica["type"],
        port_id=ica["port_id"],
        host_port_id=ICA_HOST_PORT,
        stride_channel=ica["channel_id"],
        host_channel=ica["counterparty_channel_id"] or None,
        stride_state=chain.strip_state_prefix(state=ica["state"]) or None,
        host_state=chain.strip_state_prefix(state=ica["counterparty_state"]) or None,
    )


def _channel_row(
    spec: ChannelSpec, stride: chain.Chain, host: chain.Chain, now: datetime.datetime
) -> ChannelRow:
    outbound = _packet_flow(
        sender=stride,
        sender_channel=spec.stride_channel,
        sender_port=spec.port_id,
        receiver=host,
        receiver_channel=spec.host_channel,
        receiver_port=spec.host_port_id,
    )
    # Only transfer channels carry packets the host originates; ICA channels are controller -> host.
    # Reading them needs the host's REST, which may be down while Stride's own side is fine.
    inbound = (
        chain.optional(
            lambda: _packet_flow(
                sender=host,
                sender_channel=spec.host_channel,
                sender_port=spec.host_port_id,
                receiver=stride,
                receiver_channel=spec.stride_channel,
                receiver_port=spec.port_id,
            )
        )
        if spec.name == TRANSFER_NAME and spec.host_channel
        else None
    )

    # The tx index is optional: it may not reach back far enough, and epoch-hook ICA sends are not in it.
    last_sent = chain.optional(
        lambda: chain.rpc_tx_search_latest(
            chain=stride,
            query=f"send_packet.packet_src_channel='{spec.stride_channel}'",
        )
    )
    last_received = chain.optional(
        lambda: (
            chain.rpc_tx_search_latest(
                chain=host,
                query=f"recv_packet.packet_dst_channel='{spec.host_channel}'",
            )
            if spec.host_channel
            else None
        )
    )
    last_ack = chain.optional(
        lambda: chain.rpc_tx_search_latest(
            chain=stride,
            query=f"acknowledge_packet.packet_src_channel='{spec.stride_channel}'",
        )
    )

    flows = [flow for flow in (outbound, inbound) if flow]
    return ChannelRow(
        name=spec.name,
        port_id=spec.port_id,
        stride_channel=spec.stride_channel,
        host_channel=spec.host_channel,
        stride_state=spec.stride_state,
        host_state=spec.host_state,
        outbound=outbound,
        inbound=inbound,
        last_sent=last_sent,
        last_received=last_received,
        last_ack=last_ack,
        status=channel_status(
            end_states=[spec.stride_state, spec.host_state], flows=flows, now=now
        ),
    )


def _packet_flow(
    sender: chain.Chain,
    sender_channel: str,
    sender_port: str,
    receiver: chain.Chain,
    receiver_channel: str | None,
    receiver_port: str,
) -> PacketFlow:
    """Pending packets and acks for what `sender` has committed on a channel, judged by the receiver."""
    commitments = chain.ibc_packet_commitments(
        chain=sender, channel_id=sender_channel, port_id=sender_port
    )
    if not commitments.sequences:
        return NO_PENDING

    oldest = commitments.sequences[0]
    oldest_sent_at = chain.optional(
        lambda: _send_time(
            chain_handle=sender, channel_id=sender_channel, sequence=oldest
        )
    )

    # Without a counterparty channel (handshake never finished) nothing can be received, but we cannot ask.
    unreceived = chain.optional(
        lambda: (
            chain.ibc_unreceived_packets(
                chain=receiver,
                channel_id=receiver_channel,
                port_id=receiver_port,
                sequences=commitments.sequences,
            )
            if receiver_channel
            else None
        )
    )
    if unreceived is None:
        return PacketFlow(
            pending_packets=None,
            pending_acks=None,
            capped=commitments.capped,
            oldest_sequence=oldest,
            oldest_sent_at=oldest_sent_at,
            oldest_time_source=TimeSource.SENT if oldest_sent_at else None,
        )

    split = split_pending(commitments=commitments.sequences, unreceived=unreceived)

    # A pending ack the sender's index lacks (an epoch-hook send) still has a receive tx to age it from.
    if oldest_sent_at is None and oldest in split.acks:
        received_at = chain.optional(
            lambda: _receive_time(chain_handle=receiver, channel_id=receiver_channel, sequence=oldest)
        )
        return PacketFlow(
            pending_packets=len(split.packets),
            pending_acks=len(split.acks),
            capped=commitments.capped,
            oldest_sequence=oldest,
            oldest_sent_at=received_at,
            oldest_time_source=TimeSource.RECEIVED if received_at else None,
        )

    return PacketFlow(
        pending_packets=len(split.packets),
        pending_acks=len(split.acks),
        capped=commitments.capped,
        oldest_sequence=oldest,
        oldest_sent_at=oldest_sent_at,
        oldest_time_source=TimeSource.SENT if oldest_sent_at else None,
    )


def _send_time(chain_handle: chain.Chain, channel_id: str, sequence: int) -> str | None:
    query = f"send_packet.packet_src_channel='{channel_id}' AND send_packet.packet_sequence='{sequence}'"
    sent = chain.rpc_tx_search_latest(chain=chain_handle, query=query)
    return sent.time if sent else None


def _receive_time(chain_handle: chain.Chain, channel_id: str, sequence: int) -> str | None:
    query = f"recv_packet.packet_dst_channel='{channel_id}' AND recv_packet.packet_sequence='{sequence}'"
    received = chain.rpc_tx_search_latest(chain=chain_handle, query=query)
    return received.time if received else None


# ---- host -> Osmosis legs and holder routes


def _collect_leg(zone: config.ZoneConfig, now: datetime.datetime) -> LegRow | chain.ZoneError:
    return chain.within_error_boundary(
        chain_id=zone.chain_id, work=lambda: _leg_row(zone=zone, now=now)
    )


def _leg_row(zone: config.ZoneConfig, now: datetime.datetime) -> LegRow:
    host = chain.zone_chain(zone=zone)
    osmosis = chain.zone_chain(zone=config.ZONES_BY_CHAIN_ID[config.OSMOSIS_CHAIN_ID])

    host_end = chain.ibc_channel_end(
        chain=host, channel_id=zone.osmosis_channel, port_id=chain.TRANSFER_PORT
    )
    host_client = _channel_client(chain_handle=host, end=host_end)

    # The Osmosis-side id is whatever the host channel names as its counterparty.
    osmosis_channel = host_end.counterparty_channel
    osmosis_end = (
        chain.ibc_channel_end(
            chain=osmosis, channel_id=osmosis_channel, port_id=chain.TRANSFER_PORT
        )
        if osmosis_channel
        else None
    )

    last_received = chain.optional(
        lambda: (
            chain.rpc_tx_search_latest(
                chain=osmosis,
                query=f"recv_packet.packet_dst_channel='{osmosis_channel}'",
            )
            if osmosis_channel
            else None
        )
    )
    last_ack = chain.optional(
        lambda: chain.rpc_tx_search_latest(
            chain=host,
            query=f"acknowledge_packet.packet_src_channel='{zone.osmosis_channel}'",
        )
    )
    osmosis_state = osmosis_end.state if osmosis_end else None
    return LegRow(
        chain_id=zone.chain_id,
        symbol=zone.symbol,
        host_channel=zone.osmosis_channel,
        osmosis_channel=osmosis_channel,
        host_state=host_end.state,
        osmosis_state=osmosis_state,
        host_client=host_client,
        osmosis_client=_channel_client(chain_handle=osmosis, end=osmosis_end)
        if osmosis_end
        else None,
        last_received=last_received,
        last_ack=last_ack,
        status=channel_status(
            end_states=[host_end.state, osmosis_state], flows=[], now=now
        ),
    )


def _collect_route(
    route: config.HolderRoute, now: datetime.datetime
) -> RouteRow | chain.ZoneError:
    return chain.within_error_boundary(
        chain_id=route.chain, work=lambda: _route_row(route=route, now=now)
    )


def _route_row(route: config.HolderRoute, now: datetime.datetime) -> RouteRow:
    osmosis = chain.zone_chain(zone=config.ZONES_BY_CHAIN_ID[config.OSMOSIS_CHAIN_ID])
    end = chain.ibc_channel_end(
        chain=osmosis, channel_id=route.osmosis_channel, port_id=chain.TRANSFER_PORT
    )

    # Osmosis side only: the holder chain's own RPC is not queried.
    last_received = chain.optional(
        lambda: chain.rpc_tx_search_latest(
            chain=osmosis,
            query=f"recv_packet.packet_dst_channel='{route.osmosis_channel}'",
        )
    )
    last_ack = chain.optional(
        lambda: chain.rpc_tx_search_latest(
            chain=osmosis,
            query=f"acknowledge_packet.packet_src_channel='{route.osmosis_channel}'",
        )
    )
    return RouteRow(
        chain=route.chain,
        osmosis_channel=route.osmosis_channel,
        relayed_by=route.relayed_by,
        state=end.state,
        counterparty_channel=end.counterparty_channel,
        client=_channel_client(chain_handle=osmosis, end=end),
        last_received=last_received,
        last_ack=last_ack,
        status=channel_status(end_states=[end.state], flows=[], now=now),
    )


def _channel_client(
    chain_handle: chain.Chain, end: chain.ChannelEnd
) -> chain.ClientHealth:
    connection = chain.ibc_connection_end(
        chain=chain_handle, connection_id=end.connection_id
    )
    return chain.ibc_client_health(chain=chain_handle, client_id=connection.client_id)


# ---- helpers


def _distinct_clients(
    zones: list[ZoneChannels], legs: list[LegRow], routes: list[RouteRow]
) -> list[chain.ClientHealth]:
    clients = [
        client
        for zone in zones
        for client in (zone.stride_client, zone.host_client)
        if client
    ]
    clients += [
        client
        for leg in legs
        for client in (leg.host_client, leg.osmosis_client)
        if client
    ]
    clients += [route.client for route in routes]
    return list(
        {(client.chain_id, client.client_id): client for client in clients}.values()
    )
