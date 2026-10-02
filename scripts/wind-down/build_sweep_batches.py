#!/usr/bin/env python3
"""Build MsgSweepTokensOffStride batches from a strided export.

Mirrors the on-chain skip rules of x/stakeibc/keeper/wind_down_sweep.go (20-byte address, protocol
address deny-list, blocked module address, transfer escrow exclusion, account type) and the
on-chain amount rule (a vesting account is swept for its spendable balance, i.e. minus what its
schedule still locks at --as-of), adds the
off-chain USD floor and excludes wasm contract addresses, so a batch this script emits should skip
nothing on chain. Holders are ordered by the USD value of the sweepable amounts of the listed denoms
and split into files of at most --batch-size addresses, one file per tx for
`strided tx stakeibc sweep-tokens-off-stride DENOMS FILE`.

    python3 scripts/wind-down/build_sweep_batches.py \
        --export export.json --prices prices.json --floor-usd 100 \
        --denoms stuatom,stuosmo,stutia,ustrd --out-dir sweep-batches --as-of 2026-09-29T00:00:00Z \
        [--sweep-operator stride1...] [--extra-denom ibc/27394F...=10.5:6] [--batch-size 100]

--as-of is the block time at the export height (ISO-8601 or unix seconds), NOT the export's
genesis_time (which is the chain's original genesis): get it with `strided q block <height>` and
read header.time. Vesting locks are measured at that moment.

prices.json: {"stuatom": {"usd_per_token": 4.2, "decimals": 6}, ...}. An --extra-denom that is an
ibc/ voucher must have its outermost hop in UNWIND_CHANNELS (the export's denom traces are read
from app_state.transfer.denoms when present).
"""

import argparse
import dataclasses
import datetime
import hashlib
import json
import pathlib
import re
from decimal import ROUND_HALF_EVEN, Decimal

import bech32_ref

# The chain has no cap on addresses per tx (a batch over the block gas limit fails atomically), so this is
# only a default; the real size comes from the localstride gas measurement
BATCH_SIZE_DEFAULT = 100
ADDRESS_LENGTH_BYTES = 20
TRANSFER_PORT = "transfer"
IBC_PREFIX = "ibc/"

# Mirror of types.SweepUnwindChannels; keep in sync with x/stakeibc/types/wind_down.go
UNWIND_CHANNELS = {
    "channel-0": "cosmos",
    "channel-1": "osmo",
}

# Mirror of types.SweepProtocolAddresses (constants part): staketia S0-S3 and stakedym S4-S7 plus
# the staketia operator
PROTOCOL_ADDRESSES = {
    "stride1ju3xt2f8xuhzxqg6590sazctlz6l4md0wc5w6c",  # staketia deposit
    "stride19ksqv50zmntzjfflfmnegj75tdfkk89vl2q5yu",  # staketia redemption
    "stride1pjw24gg0fm26758hxee3wta35kq9jpszcslm6z",  # staketia claim
    "stride1e7j8d6sdq272fqe2jfxjpgcagn04j75w9695fj",  # stakedym deposit
    "stride1jpsnc0ynufa2aheflj6mxzzzsu7nlwqk7ff69n",  # stakedym redemption
    "stride1q8juddwptg5yxyghh3n243pp4w8ctpvpmf6ras",  # stakedym claim
    "stride1tpzfseenwg4kq54sf9hdp3mkra652fvqtsuclq",  # staketia safe
    "stride1sj8gyqeqecqhqu7em67hn2tjzhpkdf8wz5plh7",  # stakedym safe
    "stride19xm04qaah8t2eupyeglz63vkaxzytpyc8m7kk4",  # staketia operator
}

# Mirror of the always-allowed native denoms in isSweepableNativeDenom; the rest are the stTokens of
# the export's non-deprecated host zones
ALWAYS_SWEEPABLE_NATIVE_DENOMS = {"ustrd", "stutia"}
# The default --sweep-operator (spec §4)
DEFAULT_SWEEP_OPERATOR = "stride1rjmd9gjxsexh0jg7n9wdvx9385hxxc8rjg9zzy"
INTERCHAIN_ACCOUNT = "/ibc.applications.interchain_accounts.v1.InterchainAccount"

# Module accounts the bank keeper blocks (app.BlacklistedModuleAccountAddrs: the names in maccPerms in
# app/app.go except the stakeibc, reward collector, staketia and stakedym ones and
# cons_to_send_to_provider, which can send). Keep in sync with app/app.go
BLOCKED_MODULE_NAMES = [
    "fee_collector",
    "distribution",
    "cons_redistribute",
    "mint",
    "bonded_tokens_pool",
    "not_bonded_tokens_pool",
    "gov",
    "transfer",
    "claim",
    "interchainquery",
    "interchainaccounts",
    "wasm",
    "icqoracle",
    "auction",
    "strdburner",
    "poa",
]
MODULE_ACCOUNT = "/cosmos.auth.v1beta1.ModuleAccount"

CONTINUOUS_VESTING = "/cosmos.vesting.v1beta1.ContinuousVestingAccount"
DELAYED_VESTING = "/cosmos.vesting.v1beta1.DelayedVestingAccount"
PERIODIC_VESTING = "/cosmos.vesting.v1beta1.PeriodicVestingAccount"
STRIDE_PERIODIC_VESTING = "/stride.vesting.StridePeriodicVestingAccount"
VESTING_ACCOUNT_TYPES = {CONTINUOUS_VESTING, DELAYED_VESTING, PERIODIC_VESTING, STRIDE_PERIODIC_VESTING}
SWEEPABLE_ACCOUNT_TYPES = {"/cosmos.auth.v1beta1.BaseAccount"} | VESTING_ACCOUNT_TYPES


@dataclasses.dataclass
class VestingPeriod:
    start_time: int  # absolute; for SDK periodic accounts it is derived from the cumulative lengths
    length: int
    amount: dict[str, int]


@dataclasses.dataclass
class VestingSchedule:
    account_type: str
    original_vesting: dict[str, int]
    delegated_vesting: dict[str, int]
    start_time: int
    end_time: int
    periods: list[VestingPeriod]


@dataclasses.dataclass
class Export:
    balances: dict[str, dict[str, int]]  # address -> denom -> amount
    account_types: dict[str, str]  # address -> @type
    escrow_addresses: set[str]
    contract_addresses: set[str]  # app_state.wasm.contracts[].contract_address
    denom_traces: dict[str, list[str]]  # ibc/HASH -> ["transfer/channel-x", ...] outermost first
    vesting: dict[str, VestingSchedule]  # address -> schedule, vesting accounts only
    as_of: int  # --as-of as unix seconds: the block time at the export height, when locks are measured
    module_addresses: set[str]  # sha256(name)[:20] of every ModuleAccount in auth.accounts
    allowed_native_denoms: set[str]  # ustrd, stutia and st<host_denom> of every non-deprecated host zone
    keyless: set[str]  # accounts with no pubkey and sequence 0: possibly keyless hash addresses


@dataclasses.dataclass
class Holder:
    address: str
    usd: Decimal
    balances: dict[str, int]


@dataclasses.dataclass
class Skipped:
    address: str
    reason: str
    usd: Decimal


@dataclasses.dataclass
class HolderPlan:
    denoms: list[str]
    holders: list[Holder]
    skipped: list[Skipped]
    keyless_candidates: list[Holder] = dataclasses.field(default_factory=list)


@dataclasses.dataclass
class ExtraDenom:
    denom: str
    usd_per_token: float
    decimals: int


def main() -> None:
    args = parse_args()
    export = load_export(args.export, as_of=args.as_of)
    prices = json.loads(args.prices.read_text())
    denoms = [denom for denom in args.denoms.split(",") if denom]

    for spec in args.extra_denom:
        extra = parse_extra_denom(spec)
        prices[extra.denom] = {"usd_per_token": extra.usd_per_token, "decimals": extra.decimals}
        denoms.append(extra.denom)

    # Every denom on the final list must have an on-chain destination, however it got there:
    # an ibc/ voucher passed through --denoms is checked exactly like an --extra-denom one
    for denom in denoms:
        check_denom_destination(denom=denom, traces=export.denom_traces, allowed_native=export.allowed_native_denoms)

    plan = classify_holders(
        export=export, denoms=denoms, prices=prices, floor_usd=args.floor_usd, sweep_operator=args.sweep_operator
    )
    files = write_batches(plan=plan, out_dir=args.out_dir, batch_size=args.batch_size)

    total_usd = sum((holder.usd for holder in plan.holders), Decimal(0))
    print(f"{len(plan.holders)} holders in {len(files)} batches, ${total_usd:.2f} swept, {len(plan.skipped)} skipped")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("--export", type=pathlib.Path, required=True)
    parser.add_argument("--prices", type=pathlib.Path, required=True)
    parser.add_argument("--denoms", required=True, help="comma-separated denoms to sweep")
    parser.add_argument("--floor-usd", type=float, required=True)
    parser.add_argument("--out-dir", type=pathlib.Path, required=True)
    parser.add_argument(
        "--as-of",
        type=as_of_arg,
        required=True,
        help="block time at the export height, ISO-8601 or unix seconds (`strided q block <height>`, header.time)",
    )
    parser.add_argument(
        "--sweep-operator",
        default=DEFAULT_SWEEP_OPERATOR,
        help="stride1... operator address, excluded like the protocol addresses (default: spec §4)",
    )
    parser.add_argument(
        "--batch-size",
        type=batch_size_arg,
        default=BATCH_SIZE_DEFAULT,
        help=f"addresses per tx, at least 1 (default {BATCH_SIZE_DEFAULT}); set it from the localstride gas "
        "measurement: the chain has no cap and a batch over the block gas limit fails atomically",
    )
    parser.add_argument("--extra-denom", action="append", default=[], help="DENOM=USD_PER_TOKEN:DECIMALS")
    return parser.parse_args()


def as_of_arg(text: str) -> int:
    """argparse type for --as-of: unix seconds, or ISO-8601 (a trailing Z and nanosecond fractions,
    as `strided q block` prints them, are accepted; the fraction is truncated)."""
    if text.isdigit():
        return int(text)
    normalized = re.sub(r"(\.\d{6})\d+", r"\1", text.replace("Z", "+00:00"))
    try:
        parsed = datetime.datetime.fromisoformat(normalized)
    except ValueError:
        raise argparse.ArgumentTypeError(f"--as-of must be ISO-8601 or unix seconds, got {text!r}") from None
    if parsed.tzinfo is None:
        raise argparse.ArgumentTypeError(f"--as-of needs a timezone (e.g. a trailing Z), got {text!r}")
    return int(parsed.timestamp())


def batch_size_arg(text: str) -> int:
    """argparse type for --batch-size: any positive size; the chain has no cap, only the block gas limit."""
    value = int(text)
    if value < 1:
        raise argparse.ArgumentTypeError(f"batch size must be at least 1, got {value}")
    return value


def load_export(path: pathlib.Path, as_of: int) -> Export:
    """The export's own genesis_time is deliberately unused: it is the chain's original genesis,
    not the export height, so vesting is measured at the caller's `as_of`."""
    genesis = json.loads(path.read_text())
    app_state = genesis["app_state"]

    balances: dict[str, dict[str, int]] = {}
    for entry in app_state["bank"]["balances"]:
        balances[entry["address"]] = {coin["denom"]: int(coin["amount"]) for coin in entry["coins"]}

    accounts = app_state["auth"]["accounts"]
    account_types = {account_address(account): account["@type"] for account in accounts}
    vesting = {account_address(account): vesting_schedule(account) for account in accounts if account["@type"] in VESTING_ACCOUNT_TYPES}

    keyless = {account_address(account) for account in accounts if is_keyless(account)}
    host_zones = app_state.get("stakeibc", {}).get("host_zone_list", [])
    allowed_native = ALWAYS_SWEEPABLE_NATIVE_DENOMS | {
        "st" + host_zone["host_denom"] for host_zone in host_zones if not host_zone.get("deprecated")
    }

    module_addresses = {module_address(account["name"]) for account in accounts if account["@type"] == MODULE_ACCOUNT}

    channels = app_state.get("ibc", {}).get("channel_genesis", {}).get("channels", [])
    escrows = {escrow_address(ch["port_id"], ch["channel_id"]) for ch in channels if ch["port_id"] == TRANSFER_PORT}
    contracts = {entry["contract_address"] for entry in app_state.get("wasm", {}).get("contracts", [])}

    traces: dict[str, list[str]] = {}
    for denom in app_state.get("transfer", {}).get("denoms", []):
        hops = [f"{hop['port_id']}/{hop['channel_id']}" for hop in denom.get("trace", [])]
        traces[ibc_denom(denom["base"], hops)] = hops

    return Export(
        balances=balances,
        account_types=account_types,
        escrow_addresses=escrows,
        contract_addresses=contracts,
        denom_traces=traces,
        vesting=vesting,
        as_of=as_of,
        module_addresses=module_addresses,
        allowed_native_denoms=allowed_native,
        keyless=keyless,
    )


def is_keyless(account: dict) -> bool:
    """No pubkey and sequence 0 in the export: the account has never signed, so it may be a hash
    address nobody holds a key for (autopilot's, for one). Vesting and ICA accounts nest the base."""
    base = account.get("base_account") or account.get("base_vesting_account", {}).get("base_account") or account
    return not base.get("pub_key") and int(base.get("sequence", 0)) == 0


def vesting_schedule(account: dict) -> VestingSchedule:
    """The fields the SDK's and Stride's vesting accounts use to compute what is still locked."""
    base = account["base_vesting_account"]
    account_type = account["@type"]
    start_time = int(account.get("start_time", 0))
    periods: list[VestingPeriod] = []
    if account_type == PERIODIC_VESTING:
        # SDK periodic accounts store lengths only; each period starts where the previous one ended
        cursor = start_time
        for period in account.get("vesting_periods", []):
            periods.append(VestingPeriod(start_time=cursor, length=int(period["length"]), amount=coins_by_denom(period["amount"])))
            cursor += int(period["length"])
    if account_type == STRIDE_PERIODIC_VESTING:
        # Stride periods carry their own absolute start time and vest linearly within the period
        periods = [
            VestingPeriod(start_time=int(period["start_time"]), length=int(period["length"]), amount=coins_by_denom(period["amount"]))
            for period in account.get("vesting_periods", [])
        ]
    return VestingSchedule(
        account_type=account_type,
        original_vesting=coins_by_denom(base.get("original_vesting", [])),
        delegated_vesting=coins_by_denom(base.get("delegated_vesting", [])),
        start_time=start_time,
        end_time=int(base.get("end_time", 0)),
        periods=periods,
    )


def locked_at(schedule: VestingSchedule, as_of: int) -> dict[str, int]:
    """What the bank's LockedCoins reports for this account at `as_of`: original vesting minus what
    has vested, minus the vesting coins that are delegated (min'd, the SDK's LockedCoinsFromVesting),
    per denom, zero entries dropped. Vested amounts follow each account type's GetVestedCoins:
    continuous is linear from start to end, delayed is all-or-nothing at end, SDK periodic vests a
    period whole once its cumulative end has passed, and Stride periodic vests each period linearly
    from its own start (utils.GetVestedCoinsAt). The SDK rounds its linear fractions with
    LegacyDec.RoundInt (banker's rounding), mirrored with ROUND_HALF_EVEN; a one-unit disagreement
    cannot matter because the chain moves its own SpendableCoin, not this number."""
    vested = vested_at(schedule=schedule, as_of=as_of)
    locked: dict[str, int] = {}
    for denom, original in schedule.original_vesting.items():
        still_vesting = max(original - vested.get(denom, 0), 0)
        amount = still_vesting - min(still_vesting, schedule.delegated_vesting.get(denom, 0))
        if amount > 0:
            locked[denom] = amount
    return locked


def vested_at(schedule: VestingSchedule, as_of: int) -> dict[str, int]:
    if schedule.account_type == DELAYED_VESTING:
        return dict(schedule.original_vesting) if as_of >= schedule.end_time else {}

    if schedule.account_type == CONTINUOUS_VESTING:
        if as_of <= schedule.start_time:
            return {}
        if as_of >= schedule.end_time:
            return dict(schedule.original_vesting)
        return linear_portion(amount=schedule.original_vesting, elapsed=as_of - schedule.start_time, length=schedule.end_time - schedule.start_time)

    if schedule.account_type == PERIODIC_VESTING:
        vested: dict[str, int] = {}
        for period in schedule.periods:
            if as_of < period.start_time + period.length:
                break
            vested = add_coins(vested, period.amount)
        return vested

    # Stride periodic: every period vests on its own linear schedule, all at once past end_time
    if as_of >= schedule.end_time and schedule.end_time > 0:
        return dict(schedule.original_vesting)
    vested = {}
    for period in schedule.periods:
        if as_of <= period.start_time:
            continue
        if as_of >= period.start_time + period.length:
            vested = add_coins(vested, period.amount)
            continue
        vested = add_coins(vested, linear_portion(amount=period.amount, elapsed=as_of - period.start_time, length=period.length))
    return vested


def linear_portion(amount: dict[str, int], elapsed: int, length: int) -> dict[str, int]:
    portion = Decimal(elapsed) / Decimal(length)
    return {denom: int((Decimal(value) * portion).quantize(Decimal(1), rounding=ROUND_HALF_EVEN)) for denom, value in amount.items()}


def add_coins(left: dict[str, int], right: dict[str, int]) -> dict[str, int]:
    total = dict(left)
    for denom, amount in right.items():
        total[denom] = total.get(denom, 0) + amount
    return total


def coins_by_denom(coins: list[dict]) -> dict[str, int]:
    return {coin["denom"]: int(coin["amount"]) for coin in coins}


def classify_holders(
    export: Export, denoms: list[str], prices: dict, floor_usd: float, sweep_operator: str | None = None
) -> HolderPlan:
    floor = Decimal(str(floor_usd))
    protocol = PROTOCOL_ADDRESSES | ({sweep_operator} if sweep_operator else set())
    blocked = export.module_addresses | {module_address(name) for name in BLOCKED_MODULE_NAMES}
    holders: list[Holder] = []
    skipped: list[Skipped] = []
    keyless_candidates: list[Holder] = []

    for address, coins in export.balances.items():
        # The chain sweeps SpendableCoin: a vesting account's balance minus what its schedule still
        # locks. Value and floor the same amount, and drop a denom whose spendable part is zero the
        # way the chain skips it silently
        locked = locked_at(schedule=export.vesting[address], as_of=export.as_of) if address in export.vesting else {}
        listed = {denom: amount - locked.get(denom, 0) for denom, amount in coins.items() if denom in denoms}
        listed = {denom: amount for denom, amount in listed.items() if amount > 0}
        if not listed:
            continue
        usd = sum((usd_value(denom, amount, prices) for denom, amount in listed.items()), Decimal(0))

        reason = skip_reason(address=address, export=export, protocol=protocol, blocked=blocked)
        if reason is not None:
            skipped.append(Skipped(address=address, reason=reason, usd=usd))
            continue
        if usd < floor:
            skipped.append(Skipped(address=address, reason=f"below floor (${usd:.2f} < ${floor:.2f})", usd=usd))
            continue
        holder = Holder(address=address, usd=usd, balances=listed)
        holders.append(holder)
        if address in export.keyless:
            keyless_candidates.append(holder)

    holders.sort(key=lambda holder: holder.usd, reverse=True)
    keyless_candidates.sort(key=lambda holder: holder.usd, reverse=True)
    return HolderPlan(denoms=denoms, holders=holders, skipped=skipped, keyless_candidates=keyless_candidates)


def skip_reason(address: str, export: Export, protocol: set[str], blocked: set[str]) -> str | None:
    """The on-chain rules, in the on-chain order (20 bytes, protocol, blocked, escrow, then the
    account lookup), plus the contract exclusion the chain cannot make; None means sweepable."""
    if len(address_bytes(address)) != ADDRESS_LENGTH_BYTES:
        return "address is not 20 bytes"
    if address in protocol:
        return "protocol address"
    if address in blocked:
        return "blocked module address"
    if address in export.escrow_addresses:
        return "transfer escrow address"
    account_type = export.account_types.get(address)
    if account_type is None:
        return "account not found"
    if account_type == INTERCHAIN_ACCOUNT:
        return "interchain account"
    if account_type not in SWEEPABLE_ACCOUNT_TYPES:
        return f"account type {account_type} is not sweepable"
    if address in export.contract_addresses:
        return "wasm contract address"
    return None


def write_batches(plan: HolderPlan, out_dir: pathlib.Path, batch_size: int) -> list[pathlib.Path]:
    out_dir.mkdir(parents=True, exist_ok=True)
    files: list[pathlib.Path] = []
    batches: list[dict] = []

    for index in range(0, len(plan.holders), batch_size):
        batch = plan.holders[index : index + batch_size]
        path = out_dir / f"batch-{len(files) + 1:03d}.txt"
        path.write_text("".join(f"{holder.address}\n" for holder in batch))
        files.append(path)
        batches.append({
            "file": path.name,
            "num_addresses": len(batch),
            "usd": f"{sum((holder.usd for holder in batch), Decimal(0)):.2f}",
        })

    # For ops to review by hand: holders that pass every rule but never signed, so they may be keyless
    # hash addresses. They stay in the batches; this only lists them
    (out_dir / "keyless_candidates.txt").write_text(
        "".join(f"{holder.address} ${holder.usd:.2f}\n" for holder in plan.keyless_candidates)
    )

    summary = {
        "denoms": plan.denoms,
        "num_keyless_candidates": len(plan.keyless_candidates),
        "batches": batches,
        "skipped": [{"address": entry.address, "reason": entry.reason, "usd": f"{entry.usd:.2f}"} for entry in plan.skipped],
    }
    (out_dir / "summary.json").write_text(json.dumps(summary, indent=2))
    return files


def parse_extra_denom(spec: str) -> ExtraDenom:
    if "=" not in spec or ":" not in spec.split("=", 1)[1]:
        raise ValueError(f"extra denom must be DENOM=USD_PER_TOKEN:DECIMALS, got {spec!r}")
    denom, rest = spec.split("=", 1)
    usd_per_token, decimals = rest.split(":", 1)
    return ExtraDenom(denom=denom, usd_per_token=float(usd_per_token), decimals=int(decimals))


def check_denom_destination(denom: str, traces: dict[str, list[str]], allowed_native: set[str]) -> None:
    """Mirror of resolveSweepDestination: a native denom has one (channel-5) only if it is in the
    allow-list, an ibc/ denom only if its outermost hop is a whitelisted unwind channel. Raises
    ValueError naming the denom."""
    if not denom.startswith(IBC_PREFIX):
        if denom not in allowed_native:
            raise ValueError(f"native denom {denom} is neither ustrd, stutia nor an active host zone's stToken")
        return
    hops = traces.get(denom)
    if not hops:
        raise ValueError(f"{denom} has no denom trace in the export")
    outer_port, outer_channel = hops[0].split("/")
    if outer_port != TRANSFER_PORT:
        raise ValueError(f"{denom} arrived over port {outer_port}, not {TRANSFER_PORT}")
    if outer_channel not in UNWIND_CHANNELS:
        raise ValueError(f"{denom} arrived over {outer_channel}, which is not in UNWIND_CHANNELS")


def usd_value(denom: str, amount: int, prices: dict) -> Decimal:
    price = prices.get(denom)
    if price is None:
        raise KeyError(f"no price for {denom}")
    tokens = Decimal(amount) / (Decimal(10) ** int(price["decimals"]))
    return tokens * Decimal(str(price["usd_per_token"]))


def account_address(account: dict) -> str:
    if "address" in account:
        return account["address"]
    if "base_account" in account:
        return account["base_account"]["address"]
    return account["base_vesting_account"]["base_account"]["address"]


def address_bytes(address: str) -> bytes:
    """The raw bytes behind a bech32 address, or b"" when the string is not valid bech32 (which
    the 20-byte rule then rejects, matching the on-chain decode failure)."""
    _, data = bech32_ref.bech32_decode(address)
    if data is None:
        return b""
    converted = bech32_ref.convertbits(data, 5, 8, False)
    return bytes(converted) if converted is not None else b""


def module_address(name: str) -> str:
    """SDK authtypes.NewModuleAddress: sha256(name)[:20], bech32 stride."""
    return bech32_ref.encode("stride", hashlib.sha256(name.encode()).digest()[:ADDRESS_LENGTH_BYTES])


def escrow_address(port_id: str, channel_id: str) -> str:
    """ibc-go transfertypes.GetEscrowAddress: sha256("ics20-1\\0" + port/channel)[:20], bech32 stride."""
    preimage = b"ics20-1\x00" + f"{port_id}/{channel_id}".encode()
    digest = hashlib.sha256(preimage).digest()[:ADDRESS_LENGTH_BYTES]
    return bech32_ref.encode("stride", digest)


def ibc_denom(base: str, hops: list[str]) -> str:
    full = "/".join(hops + [base]) if hops else base
    return IBC_PREFIX + hashlib.sha256(full.encode()).hexdigest().upper()


if __name__ == "__main__":
    main()
