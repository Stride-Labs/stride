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

    # Zones, legs and routes are independent, so all of them share one pool.
    with concurrent.futures.ThreadPoolExecutor(max_workers=ZONE_WORKERS) as pool:
        zone_futures = [
            pool.submit(_collect_zone, zone=zone, feed_by_chain=feed_by_chain, now=now)
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
    zone: config.ZoneConfig, feed_by_chain: dict[str, Any], now: datetime.datetime
) -> ZoneChannels | chain.ZoneError:
    return chain.within_error_boundary(
        chain_id=zone.chain_id, work=lambda: _zone_channels(zone=zone, feed_by_chain=feed_by_chain, now=now)
    )


def _zone_channels(
    zone: config.ZoneConfig, feed_by_chain: dict[str, Any], now: datetime.datetime
) -> ZoneChannels:
    entry = feed_by_chain[zone.chain_id]
    stride = chain.stride_chain()
    host = chain.zone_chain(zone=zone)

    # Transfer channel first: its ends come from the chains, the ICA ends from the feed.
    specs = [_transfer_spec(stride=stride, host=host, entry=entry)] + [
        _ica_spec(ica) for ica in entry["ica_channels"]
    ]
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
    )


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
