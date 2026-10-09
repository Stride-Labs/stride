"""Every constant of the sweep runner: the denoms and their rough prices, the operator key, endpoints, gas, limits
and state paths. Edit prices here; they are a proxy for the floor, not accounting."""

import pathlib
from dataclasses import dataclass
from decimal import Decimal

PACKAGE_DIR = pathlib.Path(__file__).resolve().parent
STATE_DIR = PACKAGE_DIR / "state"
EXCLUSIONS_PATH = PACKAGE_DIR / "exclusions.json"
PLAN_PATH = STATE_DIR / "plan.json"
LEDGER_PATH = STATE_DIR / "ledger.jsonl"
ACCOUNTS_CACHE_PATH = STATE_DIR / "accounts.json"

STRIDED_BINARY = "strided"
SWEEP_OPERATOR_KEY = "stride-sweeper"
KEYRING_BACKEND = "test"  # never `os`: it prompts for the keychain password on every call
SWEEP_OPERATOR = "stride1zvdp4efcjqs230kzuzd7qrexk4e40wutd3r8c9"
TEST_ADDRESS = "stride1nwyvkxm89yg8e3fyxgruyct4zp90mg4nlk87lg"  # ours; the mainnet rehearsal sweeps it alone
CHAIN_ID = "stride-1"
BINARY_VERSION_PREFIX = "v35"
REST = "https://stride-strd-api.polkachu.com"
RPC = "https://stride-strd-rpc.polkachu.com:443"
USER_AGENT = "curl/8.0"  # Polkachu rejects the urllib default
HTTP_TIMEOUT_SECONDS = 30
PAGE_SIZE = 1000

BECH32_PREFIX = "stride"
ADDRESS_LENGTH_BYTES = 20
TRANSFER_PORT = "transfer"
IBC_PREFIX = "ibc/"
STRIDE_TO_OSMOSIS_CHANNEL = "channel-5"
OSMOSIS_CHAIN_ID = "osmosis-1"
FEE_DENOM = "ustrd"

# Rough USD per whole native token (set 2026-10-09). The floor is a packet-count knob, so a factor of two is fine.
NATIVE_PRICES_USD: dict[str, Decimal] = {
    "ATOM": Decimal("4"), "OSMO": Decimal("0.3"), "TIA": Decimal("1.5"), "DYDX": Decimal("0.5"),
    "ISLM": Decimal("0.004"), "INJ": Decimal("8"), "JUNO": Decimal("0.1"), "BAND": Decimal("0.3"),
    "LUNA": Decimal("0.1"), "SOMM": Decimal("0.01"), "SAGA": Decimal("0.1"), "STRD": Decimal("0.05"),
}


@dataclass(frozen=True)
class NativeDenom:
    """A Stride-native sweep denom: an stToken (priced at its host's price times the zone's redemption rate) or ustrd."""

    denom: str
    symbol: str
    decimals: int
    price_symbol: str
    host_chain_id: str | None  # the stakeibc host zone whose redemption rate prices it; None for ustrd


@dataclass(frozen=True)
class VoucherDenom:
    """An IBC voucher that unwinds one hop: `ibc/sha256(transfer/<channel>/<base>)`, sent back over that channel."""

    base: str
    channel: str
    symbol: str
    decimals: int
    price_symbol: str
    chain_id: str


# To Osmosis over channel-5: the eleven in-scope stTokens and ustrd (spec §7).
NATIVE_SWEEP_DENOMS: tuple[NativeDenom, ...] = (
    NativeDenom("stuatom", "stATOM", 6, "ATOM", "cosmoshub-4"),
    NativeDenom("stuosmo", "stOSMO", 6, "OSMO", "osmosis-1"),
    NativeDenom("stutia", "stTIA", 6, "TIA", "celestia"),
    NativeDenom("stinj", "stINJ", 18, "INJ", "injective-1"),
    NativeDenom("stadydx", "stDYDX", 18, "DYDX", "dydx-mainnet-1"),
    NativeDenom("staISLM", "stISLM", 18, "ISLM", "haqq_11235-1"),
    NativeDenom("stujuno", "stJUNO", 6, "JUNO", "juno-1"),
    NativeDenom("stuband", "stBAND", 6, "BAND", "laozi-mainnet"),
    NativeDenom("stuluna", "stLUNA", 6, "LUNA", "phoenix-1"),
    NativeDenom("stusomm", "stSOMM", 6, "SOMM", "sommelier-3"),
    NativeDenom("stusaga", "stSAGA", 6, "SAGA", "ssc-1"),
    NativeDenom("ustrd", "STRD", 6, "STRD", None),
)
# Back one hop over the channel each arrived on. Must stay a subset of types.SweepUnwindChannels
# (x/stakeibc/types/wind_down.go); the planner re-checks every one against the live denom trace.
VOUCHER_SWEEP_DENOMS: tuple[VoucherDenom, ...] = (
    VoucherDenom("uatom", "channel-0", "ATOM", 6, "ATOM", "cosmoshub-4"),
    VoucherDenom("utia", "channel-162", "TIA", 6, "TIA", "celestia"),
    VoucherDenom("uosmo", "channel-5", "OSMO", 6, "OSMO", "osmosis-1"),
    VoucherDenom("ujuno", "channel-24", "JUNO", 6, "JUNO", "juno-1"),
    VoucherDenom("usomm", "channel-150", "SOMM", 6, "SOMM", "sommelier-3"),
    VoucherDenom("usaga", "channel-213", "SAGA", 6, "SAGA", "ssc-1"),
    VoucherDenom("adydx", "channel-160", "DYDX", 18, "DYDX", "dydx-mainnet-1"),
)
# Mirror of types.SweepUnwindChannels: Stride channel -> counterparty bech32 prefix.
UNWIND_CHANNELS: dict[str, str] = {
    "channel-0": "cosmos", "channel-162": "celestia", "channel-5": "osmo", "channel-24": "juno",
    "channel-150": "somm", "channel-213": "saga", "channel-160": "dydx",
}
# Mirror of isSweepableNativeDenom's constants; the rest are st<host_denom> of non-deprecated host zones.
ALWAYS_SWEEPABLE_NATIVE_DENOMS = frozenset({"ustrd", "stutia"})

# Mirror of types.SweepProtocolAddresses: staketia S0-S3, stakedym S4-S7, the staketia operator.
PROTOCOL_ADDRESSES = frozenset({
    "stride1d6ntc7s8gs86tpdyn422vsqc6uaz9cejp8nc04",  # staketia deposit
    "stride15up3hegy8zuqhy0p9m8luh0c984ptu2gxqy20g",  # staketia redemption
    "stride13nw9fm4ua8pwzmsx9kdrhefl4puz0tp7ge3gxd",  # staketia claim
    "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj",  # stakedym deposit
    "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n",  # stakedym redemption
    "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras",  # stakedym claim
    "stride18p7xg4hj2u3zpk0v9gq68pjyuuua5wa387sjjc",  # staketia safe
    "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7",  # stakedym safe
    "stride1ghhu67ttgmxrsyxljfl2tysyayswklvxs7pepw",  # staketia operator
})
# Module accounts the bank keeper blocks (app.BlacklistedModuleAccountAddrs); keep in sync with app/app.go.
BLOCKED_MODULE_NAMES = (
    "fee_collector", "distribution", "cons_redistribute", "mint", "bonded_tokens_pool", "not_bonded_tokens_pool",
    "gov", "transfer", "claim", "interchainquery", "interchainaccounts", "wasm", "icqoracle", "auction",
    "strdburner", "poa",
)

BASE_ACCOUNT = "/cosmos.auth.v1beta1.BaseAccount"
MODULE_ACCOUNT = "/cosmos.auth.v1beta1.ModuleAccount"
INTERCHAIN_ACCOUNT = "/ibc.applications.interchain_accounts.v1.InterchainAccount"
VESTING_ACCOUNT_TYPES = frozenset({
    "/cosmos.vesting.v1beta1.ContinuousVestingAccount",
    "/cosmos.vesting.v1beta1.DelayedVestingAccount",
    "/cosmos.vesting.v1beta1.PeriodicVestingAccount",
    "/stride.vesting.StridePeriodicVestingAccount",
})
SWEEPABLE_ACCOUNT_TYPES = frozenset({BASE_ACCOUNT}) | VESTING_ACCOUNT_TYPES

# Batches: packed under a gas budget at plan time, simulated at run time, refused over the block limit.
MAX_ADDRESSES_PER_BATCH = 100
GAS_BUDGET_PER_BATCH = 40_000_000
BLOCK_GAS_LIMIT = 100_000_000  # stride-1's consensus max_gas
GAS_PER_TRANSFER_DEFAULT = 150_000  # until the ledger has a confirmed batch to calibrate from
CALIBRATION_MARGIN = Decimal("1.2")
GAS_ADJUSTMENT = Decimal("1.3")
GAS_PRICE_USTRD = Decimal("0.001")  # stride-1's minimum is 0.0005ustrd

MAX_PLAN_AGE_SECONDS = 6 * 3600
TX_POLL_SECONDS = 3
TX_WAIT_SECONDS = 180
RESOLVE_WAIT_SECONDS = 600
# The floor ladder printed by `plan` and shown on the dashboard: the plan's floor, then these below it.
LADDER_FLOORS: tuple[Decimal, ...] = (Decimal(10), Decimal(5), Decimal(1), Decimal(0))
