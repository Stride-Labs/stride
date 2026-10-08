"""Every constant the dashboard uses: endpoints, in-scope zones, operator addresses, holder routes, intervals."""

import os
from dataclasses import dataclass
from enum import StrEnum

HOST = "127.0.0.1"
PORT = int(os.environ.get("WIND_DOWN_DASHBOARD_PORT", "8787"))  # a second instance can run beside yours

HTTP_TIMEOUT_SECONDS = 20
USER_AGENT = "curl/8.0"  # Polkachu rejects the urllib default

REFRESH_INTERVAL_SECONDS = {"channels": 60, "funds": 120, "validators": 300, "pools": 300}

CHANNELS_FEED_URL = "https://channels.main.stridenet.co/api/data"

# The Stride upgrade that starts the wind-down; must equal anchors.upgrade in ops/plan.json.
UPGRADE_TIME = "2026-10-12T12:00:00Z"

STRIDE_CHAIN_ID = "stride-1"
OSMOSIS_CHAIN_ID = "osmosis-1"


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
# The uosmo the plan tops the vault up with for the funding txs (about 10 OSMO): gas, not backing, so the osmosis-1
# canonical pool's allocation leaves it in the vault.
OSMO_FEE_RESERVE = 10_000_000


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


# From the relayer scope table in docs/wind-down/sttoken-locations.md
HOLDER_ROUTES: tuple[HolderRoute, ...] = (
    HolderRoute("Cosmos Hub", "channel-0", RelayedBy.FREE),
    HolderRoute("HAQQ", "channel-1575", RelayedBy.FREE),
    HolderRoute("Injective", "channel-122", RelayedBy.NOT_SERVED),
    HolderRoute("Secret", "channel-88", RelayedBy.FREE),
    HolderRoute("Penumbra", "channel-79703", RelayedBy.OURS),
    HolderRoute("Agoric", "channel-320", RelayedBy.FREE),
    HolderRoute("Neutron", "channel-874", RelayedBy.FREE),
    HolderRoute("Carbon", "channel-188", RelayedBy.FREE),
    HolderRoute("Terra", "channel-251", RelayedBy.FREE),
    HolderRoute("Dymension", "channel-19774", RelayedBy.FREE),
    HolderRoute("Axelar", "channel-208", RelayedBy.FREE),
    HolderRoute("Celestia", "channel-6994", RelayedBy.FREE),
    HolderRoute("Saga", "channel-38946", RelayedBy.FREE),
    HolderRoute("Juno", "channel-42", RelayedBy.FREE),
    HolderRoute("dYdX", "channel-6787", RelayedBy.FREE),
    HolderRoute("Band", "channel-148", RelayedBy.FREE),
)

@dataclass(frozen=True)
class HolderChain:
    """A chain whose stToken holders get a route pool: where the Pools tab's seed transfer is signed from.

    `name` is the chain registry's short name (it names the pool's alloyed subdenom, `stATOM.cosmoshub`); `node` is
    the RPC the seed command points at, `<RPC>` where Polkachu serves none.
    """

    name: str
    binary: str
    node: str


RPC_PLACEHOLDER = "<RPC>"


def _holder_node(chain_id: str, name: str) -> str:
    # The private endpoint where the dashboard already has one for the chain (its ZONES entry), else the public one.
    if chain_id in ZONES_BY_CHAIN_ID:
        return f"{ZONES_BY_CHAIN_ID[chain_id].rpc}:443"
    return f"https://{name}-rpc.polkachu.com:443"


def _holder(chain_id: str, name: str, binary: str, node_override: str | None = None) -> tuple[str, HolderChain]:
    node = node_override or _holder_node(chain_id=chain_id, name=name)
    return chain_id, HolderChain(name=name, binary=binary, node=node)


# Every chain a policy channel in REQUIRED_ROUTES resolves to (plus osmosis-1), keyed by chain id. A route whose chain
# is missing here is reported with an error on the Pools tab and gets no seed or creation command. Public hostnames
# were checked against /status on 2026-10-08; Polkachu serves no Secret or Carbon RPC, so those carry the placeholder.
HOLDER_CHAINS: dict[str, HolderChain] = dict(
    (
        _holder("cosmoshub-4", "cosmoshub", "gaiad"),
        _holder("injective-1", "injective", "injectived"),
        _holder("axelar-dojo-1", "axelar", "axelard"),
        _holder("phoenix-1", "terra", "terrad"),
        _holder("juno-1", "juno", "junod"),
        _holder("secret-4", "secret", "secretd", node_override=RPC_PLACEHOLDER),
        _holder("neutron-1", "neutron", "neutrond"),
        _holder("agoric-3", "agoric", "agd"),
        _holder("dydx-mainnet-1", "dydx", "dydxprotocold"),
        _holder("celestia", "celestia", "celestia-appd"),
        _holder("dymension_1100-1", "dymension", "dymd"),
        _holder("ssc-1", "saga", "sagad"),
        _holder("haqq_11235-1", "haqq", "haqqd"),
        _holder("laozi-mainnet", "band", "bandd"),
        _holder("osmosis-1", "osmosis", "osmosisd"),
        _holder("carbon-1", "carbon", "carbond", node_override=RPC_PLACEHOLDER),
    )
)

# Where a canonical stToken is seeded from: Stride's own transfer channel to Osmosis (Osmosis's end is channel-326).
STRIDE_CHANNEL_TO_OSMOSIS = "channel-5"
STRIDE_HOLDER = HolderChain(name="stride", binary="strided", node=f"{STRIDE_RPC}:443")

# Transmuter pool contracts on Osmosis the vault has not joined yet, so the Funds tab shows them before funding.
# Pools the vault holds an alloyed LP receipt of are discovered from its balances and need no entry here.
EXTRA_POOL_CONTRACTS: tuple[str, ...] = ()

# Osmosis's transfer channel to each in-scope host zone: where the native token's canonical Osmosis denom comes from
# (verified against the chain registry on 2026-09-23; sommelier-3 verified on chain on 2026-10-02). osmosis-1 has no
# entry: its native token is the bare denom.
OSMOSIS_CHANNEL_TO_HOST: dict[str, str] = {
    "cosmoshub-4": "channel-0",
    "celestia": "channel-6994",
    "dydx-mainnet-1": "channel-6787",
    "haqq_11235-1": "channel-1575",
    "injective-1": "channel-122",
    "juno-1": "channel-42",
    "laozi-mainnet": "channel-148",
    "phoenix-1": "channel-251",
    "sommelier-3": "channel-165",
    "ssc-1": "channel-38946",
}

# Route policy: stToken denom -> the Stride transfer channels whose holders get a route pool. A copy of
# scripts/wind-down/coverage_check.py's REQUIRED_ROUTES (the published-export audit at the halt keeps its own); a test
# asserts the two are equal, so a scope change updates both and docs/wind-down/sttoken-locations.md together.
REQUIRED_ROUTES: dict[str, frozenset[str]] = {
    "stuatom": frozenset({
        "channel-0", "channel-6", "channel-11", "channel-40", "channel-47", "channel-69", "channel-123", "channel-148",
    }),
    "staISLM": frozenset({"channel-240"}),
    "stutia": frozenset({"channel-148", "channel-123", "channel-47", "channel-0", "channel-197", "channel-162"}),
    "stinj": frozenset({"channel-6", "channel-40", "channel-0"}),
    "stuosmo": frozenset({"channel-0", "channel-40"}),
    "stuband": frozenset({"channel-258"}),
    "stadydx": frozenset({"channel-0", "channel-160"}),
    "stuluna": frozenset({"channel-13", "channel-52", "channel-47"}),
    "stusaga": frozenset({"channel-213"}),
    "stujuno": frozenset({"channel-24"}),
    "stusomm": frozenset(),
}
