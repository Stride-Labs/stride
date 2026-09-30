#!/usr/bin/env python3
"""Build MsgSweepTokensOffStride batches from a strided export.

Mirrors the on-chain skip rules of x/stakeibc/keeper/wind_down_sweep.go (20-byte address, transfer
escrow exclusion, account type) and the on-chain amount rule (a vesting account is swept for its
spendable balance, i.e. minus what its schedule still locks at the export's genesis_time), adds the
off-chain USD floor and excludes wasm contract addresses, so a batch this script emits should skip
nothing on chain. Holders are ordered by the USD value of the sweepable amounts of the listed denoms
and split into files of at most --batch-size addresses, one file per tx for
`strided tx stakeibc sweep-tokens-off-stride DENOMS FILE`.

    python3 scripts/wind-down/build_sweep_batches.py \
        --export export.json --prices prices.json --floor-usd 100 \
        --denoms stuatom,stuosmo,stutia,ustrd --out-dir sweep-batches \
        [--extra-denom ibc/27394F...=10.5:6] [--batch-size 100]

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
from decimal import ROUND_HALF_EVEN, Decimal

import bech32_ref

# Mirror of types.MaxSweepAddressesPerTx; a batch above it fails ValidateBasic on chain
MAX_SWEEP_ADDRESSES_PER_TX = 100
BATCH_SIZE_DEFAULT = MAX_SWEEP_ADDRESSES_PER_TX
ADDRESS_LENGTH_BYTES = 20
TRANSFER_PORT = "transfer"
IBC_PREFIX = "ibc/"

# Mirror of types.SweepUnwindChannels; keep in sync with x/stakeibc/types/wind_down.go
UNWIND_CHANNELS = {
    "channel-0": "cosmos",
    "channel-162": "celestia",
    "channel-5": "osmo",
    "channel-24": "juno",
    "channel-150": "somm",
    "channel-213": "saga",
    "channel-160": "dydx",
}

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
    as_of: int  # genesis_time as unix seconds: the moment the export's balances and locks describe


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


@dataclasses.dataclass
class ExtraDenom:
    denom: str
    usd_per_token: float
    decimals: int


def main() -> None:
    args = parse_args()
    export = load_export(args.export)
    prices = json.loads(args.prices.read_text())
    denoms = [denom for denom in args.denoms.split(",") if denom]

    for spec in args.extra_denom:
        extra = parse_extra_denom(spec)
        prices[extra.denom] = {"usd_per_token": extra.usd_per_token, "decimals": extra.decimals}
        denoms.append(extra.denom)

    # Every denom on the final list must have an on-chain destination, however it got there:
    # an ibc/ voucher passed through --denoms is checked exactly like an --extra-denom one
    for denom in denoms:
        check_denom_destination(denom=denom, traces=export.denom_traces)

    plan = classify_holders(export=export, denoms=denoms, prices=prices, floor_usd=args.floor_usd)
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
    parser.add_argument("--batch-size", type=batch_size_arg, default=BATCH_SIZE_DEFAULT, help=f"1..{MAX_SWEEP_ADDRESSES_PER_TX}")
    parser.add_argument("--extra-denom", action="append", default=[], help="DENOM=USD_PER_TOKEN:DECIMALS")
    return parser.parse_args()


def batch_size_arg(text: str) -> int:
    """argparse type for --batch-size: the chain rejects a tx with more than MAX_SWEEP_ADDRESSES_PER_TX addresses."""
    value = int(text)
    if value < 1 or value > MAX_SWEEP_ADDRESSES_PER_TX:
        raise argparse.ArgumentTypeError(f"batch size must be between 1 and {MAX_SWEEP_ADDRESSES_PER_TX}, got {value}")
    return value


def load_export(path: pathlib.Path) -> Export:
    genesis = json.loads(path.read_text())
    app_state = genesis["app_state"]
    as_of = int(datetime.datetime.fromisoformat(genesis["genesis_time"].replace("Z", "+00:00")).timestamp())

    balances: dict[str, dict[str, int]] = {}
    for entry in app_state["bank"]["balances"]:
        balances[entry["address"]] = {coin["denom"]: int(coin["amount"]) for coin in entry["coins"]}

    accounts = app_state["auth"]["accounts"]
    account_types = {account_address(account): account["@type"] for account in accounts}
    vesting = {account_address(account): vesting_schedule(account) for account in accounts if account["@type"] in VESTING_ACCOUNT_TYPES}

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
    )


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


def classify_holders(export: Export, denoms: list[str], prices: dict, floor_usd: float) -> HolderPlan:
    floor = Decimal(str(floor_usd))
    holders: list[Holder] = []
    skipped: list[Skipped] = []

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

        reason = skip_reason(address=address, export=export)
        if reason is not None:
            skipped.append(Skipped(address=address, reason=reason, usd=usd))
            continue
        if usd < floor:
            skipped.append(Skipped(address=address, reason=f"below floor (${usd:.2f} < ${floor:.2f})", usd=usd))
            continue
        holders.append(Holder(address=address, usd=usd, balances=listed))

    holders.sort(key=lambda holder: holder.usd, reverse=True)
    return HolderPlan(denoms=denoms, holders=holders, skipped=skipped)


def skip_reason(address: str, export: Export) -> str | None:
    """The on-chain rules, in the on-chain order (escrow before the account lookup), plus the
    contract exclusion the chain cannot make; None means sweepable."""
    if len(address_bytes(address)) != ADDRESS_LENGTH_BYTES:
        return "address is not 20 bytes"
    if address in export.escrow_addresses:
        return "transfer escrow address"
    account_type = export.account_types.get(address)
    if account_type is None:
        return "account not found"
    if account_type not in SWEEPABLE_ACCOUNT_TYPES:
        return f"account type {account_type} is not sweepable"
    if address in export.contract_addresses:
        return "wasm contract address"
    return None


def write_batches(plan: HolderPlan, out_dir: pathlib.Path, batch_size: int) -> list[pathlib.Path]:
    # Guarded here too so a caller that bypasses parse_args cannot emit a batch the chain rejects
    if batch_size < 1 or batch_size > MAX_SWEEP_ADDRESSES_PER_TX:
        raise ValueError(f"batch size must be between 1 and {MAX_SWEEP_ADDRESSES_PER_TX}, got {batch_size}")
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

    summary = {
        "denoms": plan.denoms,
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


def check_denom_destination(denom: str, traces: dict[str, list[str]]) -> None:
    """Mirror of resolveSweepDestination: a native denom always has one (channel-5), an ibc/ denom
    only if its outermost hop is a whitelisted unwind channel. Raises ValueError naming the denom."""
    if not denom.startswith(IBC_PREFIX):
        return
    hops = traces.get(denom)
    if not hops:
        raise ValueError(f"{denom} has no denom trace in the export")
    outer_channel = hops[0].split("/")[1]
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
