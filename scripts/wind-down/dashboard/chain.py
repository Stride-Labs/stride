"""HTTP and Cosmos query helpers shared by the collectors (REST, CometBFT RPC, pagination, IBC)."""

import datetime
import functools
import hashlib
import http.client
import json
import time
import urllib.parse
import urllib.request
from collections.abc import Callable
from dataclasses import dataclass
from typing import Any, TypeVar

import config

COMMITMENT_CAP = 1000  # packet commitments read per channel per direction
UNRECEIVED_BATCH_SIZE = 100  # sequences per unreceived_packets URL, to keep URLs short
PAGE_SIZE = 1000
# tx_search on a busy channel scans every match (40s and timeouts on Osmosis), so look at recent blocks first.
RECENT_BLOCK_WINDOW = 100_000
LATEST_HEIGHT_CACHE_SECONDS = 30

TRANSFER_PORT = "transfer"
STATE_PREFIX = "STATE_"
CLIENT_STATUS_ACTIVE = "Active"

# urllib wraps only the request send in URLError; connection resets, truncated bodies and TLS failures while
# reading the response surface as plain OSError / http.client errors (URLError and TimeoutError are OSError).
NETWORK_ERRORS = (OSError, http.client.HTTPException)
# What one zone's work may raise without blanking the tab; anything else aborts the refresh.
ZONE_ERRORS = (*NETWORK_ERRORS, json.JSONDecodeError, KeyError)

T = TypeVar("T")


@dataclass(frozen=True)
class Chain:
    """The handle passed around: a chain id with its REST and CometBFT RPC base URLs."""

    chain_id: str
    rest: str
    rpc: str


@dataclass(frozen=True)
class TxRef:
    height: int
    time: str  # block time, RFC 3339


@dataclass(frozen=True)
class ChannelEnd:
    state: str  # "OPEN", "CLOSED", "INIT", "TRYOPEN"
    counterparty_port: str
    counterparty_channel: (
        str | None
    )  # empty while the handshake has not reached the counterparty
    connection_id: str


@dataclass(frozen=True)
class ConnectionEnd:
    state: str
    client_id: str
    counterparty_client_id: str
    counterparty_connection_id: str


@dataclass(frozen=True)
class PacketCommitments:
    sequences: list[int]
    capped: bool  # more commitments exist than COMMITMENT_CAP


@dataclass(frozen=True)
class ZoneError:
    chain_id: str
    error: str


@dataclass(frozen=True)
class ClientHealth:
    chain_id: str  # the chain the client lives on
    client_id: str
    status: str  # "Active", "Expired", "Frozen"
    expires_at: str | None  # latest consensus state timestamp + trusting period
    seconds_remaining: float | None


# ---- chain handles


def stride_chain() -> Chain:
    return Chain(
        chain_id=config.STRIDE_CHAIN_ID, rest=config.STRIDE_REST, rpc=config.STRIDE_RPC
    )


def zone_chain(zone: config.ZoneConfig) -> Chain:
    return Chain(chain_id=zone.chain_id, rest=zone.rest, rpc=zone.rpc)


# ---- HTTP


def get_json(url: str) -> Any:
    """GET a URL and decode the JSON body, retrying once on a network error."""
    try:
        return _fetch_json(url=url)
    except NETWORK_ERRORS:
        return _fetch_json(url=url)


def rest_get(chain: Chain, path: str, params: dict[str, str] | None = None) -> Any:
    query = f"?{urllib.parse.urlencode(params)}" if params else ""
    return get_json(url=f"{chain.rest}{path}{query}")


def rest_get_all_pages(
    chain: Chain,
    path: str,
    key: str,
    params: dict[str, str] | None = None,
    max_items: int | None = None,
) -> list[Any]:
    """Collect `response[key]` across Cosmos `pagination.key` / `next_key` pages, stopping at max_items."""
    items: list[Any] = []
    page_params = {"pagination.limit": str(PAGE_SIZE), **(params or {})}

    while True:
        response = rest_get(chain=chain, path=path, params=page_params)
        items.extend(response.get(key) or [])

        next_key = (response.get("pagination") or {}).get("next_key")
        if not next_key or (max_items is not None and len(items) >= max_items):
            return items

        page_params = {**page_params, "pagination.key": next_key}


def rpc_tx_search(
    chain: Chain, query: str, per_page: int = 100, newest_first: bool = True
) -> list[dict[str, Any]]:
    """The txs matching a CometBFT event query, newest first by default."""
    params = {
        "query": f'"{query}"',
        "per_page": str(per_page),
        "order_by": '"desc"' if newest_first else '"asc"',
    }
    response = get_json(url=f"{chain.rpc}/tx_search?{urllib.parse.urlencode(params)}")
    return response["result"]["txs"]


def rpc_tx_search_latest(chain: Chain, query: str) -> TxRef | None:
    """Height and block time of the newest tx matching the query, or None when the index holds none.

    Searches the most recent RECENT_BLOCK_WINDOW blocks first and only then the whole index, which is
    cheap because a channel with no recent matches has few matches overall.
    """
    window_start = rpc_latest_height(chain=chain) - RECENT_BLOCK_WINDOW
    recent = _newest_tx(chain=chain, query=f"{query} AND tx.height>={window_start}")
    return recent or _newest_tx(chain=chain, query=query)


def rpc_latest_height(chain: Chain) -> int:
    """The chain's latest block height, reused for LATEST_HEIGHT_CACHE_SECONDS."""
    return _latest_height(rpc=chain.rpc, time_bucket=int(time.time() // LATEST_HEIGHT_CACHE_SECONDS))


def ibc_denom(path: str) -> str:
    """The on-chain denom of an IBC voucher, from its trace path such as `transfer/channel-0/uatom`."""
    return "ibc/" + hashlib.sha256(path.encode()).hexdigest().upper()


# ---- IBC


def ibc_channel_end(chain: Chain, channel_id: str, port_id: str) -> ChannelEnd:
    response = rest_get(
        chain=chain, path=f"/ibc/core/channel/v1/channels/{channel_id}/ports/{port_id}"
    )
    channel = response["channel"]
    return ChannelEnd(
        state=strip_state_prefix(state=channel["state"]),
        counterparty_port=channel["counterparty"]["port_id"],
        counterparty_channel=channel["counterparty"]["channel_id"] or None,
        connection_id=channel["connection_hops"][0],
    )


def ibc_connection_end(chain: Chain, connection_id: str) -> ConnectionEnd:
    connection = rest_get(
        chain=chain, path=f"/ibc/core/connection/v1/connections/{connection_id}"
    )["connection"]
    return ConnectionEnd(
        state=strip_state_prefix(state=connection["state"]),
        client_id=connection["client_id"],
        counterparty_client_id=connection["counterparty"]["client_id"],
        counterparty_connection_id=connection["counterparty"]["connection_id"],
    )


def ibc_packet_commitments(
    chain: Chain, channel_id: str, port_id: str, cap: int = COMMITMENT_CAP
) -> PacketCommitments:
    """The sequences this chain still holds a commitment for, lowest first, read up to `cap`."""
    path = (
        f"/ibc/core/channel/v1/channels/{channel_id}/ports/{port_id}/packet_commitments"
    )
    # One past the cap, so "exactly cap" is distinguishable from "more than cap".
    commitments = rest_get_all_pages(
        chain=chain, path=path, key="commitments", max_items=cap + 1
    )
    sequences = sorted(int(commitment["sequence"]) for commitment in commitments)
    return PacketCommitments(sequences=sequences[:cap], capped=len(sequences) > cap)


def ibc_unreceived_packets(
    chain: Chain, channel_id: str, port_id: str, sequences: list[int]
) -> list[int]:
    """Of `sequences`, the ones the receiving chain has not received yet (asked of the receiving chain)."""
    unreceived: list[int] = []
    for start in range(0, len(sequences), UNRECEIVED_BATCH_SIZE):
        batch = ",".join(
            str(sequence)
            for sequence in sequences[start : start + UNRECEIVED_BATCH_SIZE]
        )
        path = f"/ibc/core/channel/v1/channels/{channel_id}/ports/{port_id}/packet_commitments/{batch}/unreceived_packets"
        unreceived.extend(
            int(sequence) for sequence in rest_get(chain=chain, path=path)["sequences"]
        )
    return unreceived


def ibc_unreceived_acks(
    chain: Chain, channel_id: str, port_id: str, sequences: list[int]
) -> list[int]:
    """Of `sequences`, the ones the sending chain still holds a commitment for: not yet acknowledged or timed out
    (asked of the sending chain)."""
    committed: list[int] = []
    for start in range(0, len(sequences), UNRECEIVED_BATCH_SIZE):
        batch = ",".join(
            str(sequence)
            for sequence in sequences[start : start + UNRECEIVED_BATCH_SIZE]
        )
        path = f"/ibc/core/channel/v1/channels/{channel_id}/ports/{port_id}/packet_commitments/{batch}/unreceived_acks"
        committed.extend(
            int(sequence) for sequence in rest_get(chain=chain, path=path)["sequences"]
        )
    return committed


@functools.lru_cache(maxsize=4096)
def block_time(rpc: str, height: int) -> str:
    """Block time at a height; heights never change, so a hit is cached for the life of the process."""
    return get_json(url=f"{rpc}/header?height={height}")["result"]["header"]["time"]


def ibc_client_status(chain: Chain, client_id: str) -> str:
    return rest_get(chain=chain, path=f"/ibc/core/client/v1/client_status/{client_id}")[
        "status"
    ]


def ibc_client_health(
    chain: Chain, client_id: str, now: datetime.datetime | None = None
) -> ClientHealth:
    """Client status plus its expiry: latest consensus state timestamp + trusting period."""
    status = ibc_client_status(chain=chain, client_id=client_id)

    client_state = rest_get(
        chain=chain, path=f"/ibc/core/client/v1/client_states/{client_id}"
    )["client_state"]
    latest = client_state["latest_height"]
    consensus_path = (
        f"/ibc/core/client/v1/consensus_states/{client_id}"
        f"/revision/{latest['revision_number']}/height/{latest['revision_height']}"
    )
    consensus = rest_get(chain=chain, path=consensus_path)["consensus_state"]

    expires_at = client_expiry(
        consensus_timestamp=consensus["timestamp"],
        trusting_period_seconds=parse_duration_seconds(
            duration=client_state["trusting_period"]
        ),
    )
    current = now or datetime.datetime.now(datetime.UTC)
    return ClientHealth(
        chain_id=chain.chain_id,
        client_id=client_id,
        status=status,
        expires_at=expires_at.isoformat(),
        seconds_remaining=(expires_at - current).total_seconds(),
    )


# ---- error boundaries


def within_error_boundary(chain_id: str, work: Callable[[], T]) -> T | ZoneError:
    """Run one zone's work; a network or decode failure becomes an error record instead of blanking the tab."""
    try:
        return work()
    except ZONE_ERRORS as error:
        return ZoneError(chain_id=chain_id, error=f"{type(error).__name__}: {error}")


def optional(lookup: Callable[[], T | None]) -> T | None:
    """An optional lookup: a failure means "not available", shown as n/a."""
    try:
        return lookup()
    except ZONE_ERRORS:
        return None


def succeeded(results: list[Any], result_type: type[T]) -> list[T]:
    return [result for result in results if isinstance(result, result_type)]


# ---- pure helpers


def strip_state_prefix(state: str) -> str:
    return state.removeprefix(STATE_PREFIX)


def parse_timestamp(timestamp: str) -> datetime.datetime:
    """Parse an RFC 3339 timestamp with up to nanosecond precision (datetime keeps microseconds)."""
    body = timestamp.removesuffix("Z")
    whole, _, fraction = body.partition(".")
    microseconds = fraction[:6].ljust(6, "0")
    return datetime.datetime.fromisoformat(f"{whole}.{microseconds}").replace(
        tzinfo=datetime.UTC
    )


def parse_duration_seconds(duration: str) -> float:
    """Parse the protobuf JSON duration form, e.g. `1036800s`."""
    return float(duration.removesuffix("s"))


def client_expiry(
    consensus_timestamp: str, trusting_period_seconds: float
) -> datetime.datetime:
    return parse_timestamp(timestamp=consensus_timestamp) + datetime.timedelta(
        seconds=trusting_period_seconds
    )


# ---- helpers


def _fetch_json(url: str) -> Any:
    request = urllib.request.Request(url, headers={"User-Agent": config.USER_AGENT})
    with urllib.request.urlopen(
        request, timeout=config.HTTP_TIMEOUT_SECONDS
    ) as response:
        return json.load(response)


def _newest_tx(chain: Chain, query: str) -> TxRef | None:
    txs = rpc_tx_search(chain=chain, query=query, per_page=1)
    if not txs:
        return None

    height = int(txs[0]["height"])
    return TxRef(height=height, time=block_time(rpc=chain.rpc, height=height))


@functools.lru_cache(maxsize=256)
def _latest_height(rpc: str, time_bucket: int) -> int:
    return int(get_json(url=f"{rpc}/status")["result"]["sync_info"]["latest_block_height"])
