"""Every constant the dashboard uses: endpoints, in-scope zones, operator addresses, holder routes, intervals."""

from dataclasses import dataclass
from enum import StrEnum

HOST = "127.0.0.1"
PORT = 8787

HTTP_TIMEOUT_SECONDS = 20
USER_AGENT = "curl/8.0"  # Polkachu rejects the urllib default

REFRESH_INTERVAL_SECONDS = {"channels": 60, "funds": 120, "validators": 300}

CHANNELS_FEED_URL = "https://channels.main.stridenet.co/api/data"

STRIDE_CHAIN_ID = "stride-1"
OSMOSIS_CHAIN_ID = "osmosis-1"

# Stride <-> Osmosis transfer channel that the token sweep uses; it is osmosis-1's transfer channel in the feed.
SWEEP_CHANNEL = "channel-5"


def _polkachu_rest(name: str) -> str:
    return f"https://{name}-strd-api.polkachu.com"


def _polkachu_rpc(name: str) -> str:
    return f"https://{name}-strd-rpc.polkachu.com"


STRIDE_REST = _polkachu_rest(name="stride")
STRIDE_RPC = _polkachu_rpc(name="stride")


@dataclass(frozen=True)
class ZoneConfig:
    chain_id: str
    endpoint_name: str
    symbol: str
    decimals: int
    # Host-side transfer channel to Osmosis (x/stakeibc HostToOsmosisTransferChannel); None for osmosis-1 itself.
    osmosis_channel: str | None
    rest: str
    rpc: str


def _zone(
    chain_id: str,
    endpoint_name: str,
    symbol: str,
    decimals: int,
    osmosis_channel: str | None,
    rest_override: str | None = None,
) -> ZoneConfig:
    return ZoneConfig(
        chain_id=chain_id,
        endpoint_name=endpoint_name,
        symbol=symbol,
        decimals=decimals,
        osmosis_channel=osmosis_channel,
        rest=rest_override or _polkachu_rest(name=endpoint_name),
        rpc=_polkachu_rpc(name=endpoint_name),
    )


# Edit an entry here to point a zone at different endpoints.
ZONES: tuple[ZoneConfig, ...] = (
    _zone("celestia", "celestia", "TIA", 6, "channel-2"),
    _zone("cosmoshub-4", "cosmos", "ATOM", 6, "channel-141"),
    _zone("dydx-mainnet-1", "dydx", "DYDX", 18, "channel-3"),
    # haqq-strd-api.polkachu.com has returned 502 since 2026-09-30; the public Polkachu REST works.
    _zone(
        "haqq_11235-1",
        "haqq",
        "ISLM",
        18,
        "channel-2",
        rest_override="https://haqq-api.polkachu.com",
    ),
    _zone("injective-1", "injective", "INJ", 18, "channel-8"),
    _zone("juno-1", "juno", "JUNO", 6, "channel-0"),
    _zone("laozi-mainnet", "band", "BAND", 6, "channel-83"),
    _zone("osmosis-1", "osmosis", "OSMO", 6, None),
    _zone("phoenix-1", "terra", "LUNA", 6, "channel-1"),
    _zone("sommelier-3", "sommelier", "SOMM", 6, "channel-0"),
    _zone("ssc-1", "saga", "SAGA", 6, "channel-1"),
)

ZONES_BY_CHAIN_ID: dict[str, ZoneConfig] = {zone.chain_id: zone for zone in ZONES}


# Operator addresses
PROTOCOL_ADMIN = "stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh"
GOV = "stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl"
SWEEP_OPERATOR = "stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9"
OSMOSIS_VAULT = "osmo1k8c2m5cn322akk5wy8lpt87dd2f4yh9afcd7af"


class RelayedBy(StrEnum):
    FREE = "free"  # someone else already relays it
    OURS = "ours"  # we will relay it
    NOT_SERVED = "not served"


@dataclass(frozen=True)
class HolderRoute:
    """A chain whose holders reach the Osmosis pools, queried on the Osmosis side only."""

    chain: str
    osmosis_channel: str
    relayed_by: RelayedBy
    # cosmos.directory registry name, for chains without a Polkachu endpoint (used by later tabs)
    registry_name: str | None = None

    @property
    def rest(self) -> str | None:
        return (
            f"https://rest.cosmos.directory/{self.registry_name}"
            if self.registry_name
            else None
        )

    @property
    def rpc(self) -> str | None:
        return (
            f"https://rpc.cosmos.directory/{self.registry_name}"
            if self.registry_name
            else None
        )


# From the relayer scope table in docs/wind-down/sttoken-locations.md
HOLDER_ROUTES: tuple[HolderRoute, ...] = (
    HolderRoute("Cosmos Hub", "channel-0", RelayedBy.FREE),
    HolderRoute("HAQQ", "channel-1575", RelayedBy.FREE),
    HolderRoute("Injective", "channel-122", RelayedBy.NOT_SERVED),
    HolderRoute("Secret", "channel-88", RelayedBy.FREE, registry_name="secretnetwork"),
    HolderRoute("Penumbra", "channel-79703", RelayedBy.OURS, registry_name="penumbra"),
    HolderRoute("Agoric", "channel-320", RelayedBy.FREE, registry_name="agoric"),
    HolderRoute("Neutron", "channel-874", RelayedBy.FREE, registry_name="neutron"),
    HolderRoute("Carbon", "channel-188", RelayedBy.FREE, registry_name="carbon"),
    HolderRoute("Terra", "channel-251", RelayedBy.FREE),
    HolderRoute(
        "Dymension", "channel-19774", RelayedBy.FREE, registry_name="dymension"
    ),
    HolderRoute("Axelar", "channel-208", RelayedBy.FREE, registry_name="axelar"),
    HolderRoute("Celestia", "channel-6994", RelayedBy.FREE),
    HolderRoute("Saga", "channel-38946", RelayedBy.FREE),
    HolderRoute("Juno", "channel-42", RelayedBy.FREE),
    HolderRoute("dYdX", "channel-6787", RelayedBy.FREE),
    HolderRoute("Band", "channel-148", RelayedBy.FREE),
)

# Transmuter pool contracts on Osmosis the vault has not joined yet, so the Funds tab shows them before funding.
# Pools the vault holds an alloyed LP receipt of are discovered from its balances and need no entry here.
EXTRA_POOL_CONTRACTS: tuple[str, ...] = ()
