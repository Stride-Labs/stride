"""Live holder state and the sweep's inclusion rules.

Reads, all through chainio: the sweep denoms (checked the way the chain checks them), every holder of each denom in bulk,
the auth account of each candidate above the floor (cached), the spendable balance of vesting candidates, and the
module, escrow and contract address sets. Classification mirrors sweepSkipReason in x/stakeibc/keeper/wind_down_sweep.go
in the chain's order, then adds the two rules the chain cannot make: wasm contracts and the reviewed exclusions file."""

import json
import pathlib
from dataclasses import dataclass, field
from decimal import Decimal

from sweep import addresses, chainio, config

STTOKEN_PREFIX = "st"
ACCOUNT_NOT_FOUND = "account not found"


class ExclusionsError(Exception):
    """exclusions.json is malformed."""


class DenomError(Exception):
    """A configured sweep denom has no on-chain destination; the chain would reject every batch naming it."""


@dataclass(frozen=True)
class SweepDenom:
    denom: str
    symbol: str
    decimals: int
    price_usd: Decimal  # per whole token
    destination: str  # chain id
    channel: str  # the Stride transfer channel the sweep sends over


@dataclass
class Holder:
    address: str
    balances: dict[str, int]
    usd: Decimal
    keyless: bool
    # A vesting account's total minus spendable, per denom: what stays behind after a full sweep. Empty otherwise.
    locked: dict[str, int] = field(default_factory=dict)

    @property
    def transfers(self) -> int:
        return sum(1 for amount in self.balances.values() if amount > 0)


@dataclass(frozen=True)
class Excluded:
    address: str
    reason: str
    usd: Decimal


@dataclass
class HolderSet:
    height: int
    denoms: list[SweepDenom]
    holders: list[Holder]
    excluded: list[Excluded]  # the exclusions file
    skipped: list[Excluded]  # the chain's rules and the contract rule
    # Under the floor in spendable value; includes vesting accounts whose spendable part is empty or small
    below_floor: list[Holder]


@dataclass(frozen=True)
class Exclusion:
    address: str
    section: str
    label: str
    reason: str


@dataclass(frozen=True)
class AccountInfo:
    type: str
    has_pubkey: bool
    sequence: int


@dataclass(frozen=True)
class SkipInputs:
    module_addresses: set[str]
    escrows: set[str]
    contracts: set[str]


# ---- exclusions file


def load_exclusions(path: pathlib.Path = config.EXCLUSIONS_PATH) -> dict[str, Exclusion]:
    try:
        sections = json.loads(path.read_text())["sections"]
    except (OSError, json.JSONDecodeError, KeyError) as error:
        raise ExclusionsError(f"{path}: {error}") from None

    exclusions: dict[str, Exclusion] = {}
    for section in sections:
        _require_keys(path=path, item=section, keys=("name", "reason", "addresses"), where="a section")
        for entry in section["addresses"]:
            _require_keys(
                path=path, item=entry, keys=("address", "label"), where=f"an entry in section {section['name']}")
            address = entry["address"]
            if not addresses.is_stride_address(address=address):
                raise ExclusionsError(
                    f"{path}: {address!r} in section {section['name']} is not a 20-byte stride address")
            if address in exclusions:
                raise ExclusionsError(f"{path}: {address} is listed twice")
            exclusions[address] = Exclusion(
                address=address, section=section["name"], label=entry["label"], reason=section["reason"])
    return exclusions


def _require_keys(path: pathlib.Path, item: object, keys: tuple[str, ...], where: str) -> None:
    if not isinstance(item, dict):
        raise ExclusionsError(f"{path}: {where} is not an object")
    missing = [key for key in keys if key not in item]
    if missing:
        raise ExclusionsError(f"{path}: {where} lacks {', '.join(missing)}")


# ---- denoms


def resolve_denoms() -> list[SweepDenom]:
    """The configured denoms, each checked the way resolveSweepDestination checks it, with its price."""
    host_zones = chainio.rest_get_all_pages(path="/Stride-Labs/stride/stakeibc/host_zone", key="host_zone")
    zones_by_chain = {zone["chain_id"]: zone for zone in host_zones}
    allowed_native = set(config.ALWAYS_SWEEPABLE_NATIVE_DENOMS) | {
        STTOKEN_PREFIX + zone["host_denom"] for zone in host_zones if not zone.get("deprecated")
    }

    denoms = [_native_denom(spec=spec, allowed=allowed_native, zones_by_chain=zones_by_chain)
              for spec in config.NATIVE_SWEEP_DENOMS]
    denoms.extend(_voucher_denom(spec=spec) for spec in config.VOUCHER_SWEEP_DENOMS)
    return denoms


def _native_denom(spec: config.NativeDenom, allowed: set[str], zones_by_chain: dict[str, dict]) -> SweepDenom:
    if spec.denom not in allowed:
        raise DenomError(f"native denom {spec.denom} is neither ustrd, stutia nor an active host zone's stToken")
    rate = Decimal(1)
    if spec.host_chain_id is not None:
        zone = zones_by_chain.get(spec.host_chain_id)
        if zone is None:
            raise DenomError(f"{spec.denom}: host zone {spec.host_chain_id} not found")
        rate = Decimal(zone["redemption_rate"])
    return SweepDenom(denom=spec.denom, symbol=spec.symbol, decimals=spec.decimals,
                      price_usd=config.NATIVE_PRICES_USD[spec.price_symbol] * rate,
                      destination=config.OSMOSIS_CHAIN_ID, channel=config.STRIDE_TO_OSMOSIS_CHANNEL)


def _voucher_denom(spec: config.VoucherDenom) -> SweepDenom:
    if spec.channel not in config.UNWIND_CHANNELS:
        raise DenomError(f"{spec.base} over {spec.channel}: not in UNWIND_CHANNELS")
    denom = addresses.ibc_denom(path=f"{config.TRANSFER_PORT}/{spec.channel}/{spec.base}")
    try:
        trace = chainio.rest_get(path=f"/ibc/apps/transfer/v1/denoms/{denom.removeprefix(config.IBC_PREFIX)}")["denom"]
    except chainio.NotFound:
        raise DenomError(f"{spec.base} over {spec.channel}: {denom} has no denom trace on chain") from None
    if not trace["trace"]:
        raise DenomError(f"{spec.base} over {spec.channel}: {denom} has an empty denom trace")
    outer = trace["trace"][0]
    if outer["port_id"] != config.TRANSFER_PORT or outer["channel_id"] != spec.channel:
        raise DenomError(
            f"{denom}: outermost hop is {outer['port_id']}/{outer['channel_id']}, expected transfer/{spec.channel}")
    return SweepDenom(denom=denom, symbol=spec.symbol, decimals=spec.decimals,
                      price_usd=config.NATIVE_PRICES_USD[spec.price_symbol], destination=spec.chain_id,
                      channel=spec.channel)


# ---- bulk reads


def read_balances(denoms: list[SweepDenom]) -> dict[str, dict[str, int]]:
    """Every holder of every sweep denom: address -> denom -> amount, zero amounts dropped."""
    balances: dict[str, dict[str, int]] = {}
    for sweep_denom in denoms:
        owners = chainio.rest_get_all_pages(path="/cosmos/bank/v1beta1/denom_owners_by_query", key="denom_owners",
                                            params={"denom": sweep_denom.denom})
        for owner in owners:
            amount = int(owner["balance"]["amount"])
            if amount > 0:
                balances.setdefault(owner["address"], {})[sweep_denom.denom] = amount
    return balances


def read_skip_inputs() -> SkipInputs:
    module_accounts = chainio.rest_get(path="/cosmos/auth/v1beta1/module_accounts")["accounts"]
    module_addresses = {account["base_account"]["address"] for account in module_accounts}
    module_addresses |= {addresses.module_address(name=name) for name in config.BLOCKED_MODULE_NAMES}

    channels = chainio.rest_get_all_pages(path="/ibc/core/channel/v1/channels", key="channels")
    escrows = {addresses.escrow_address(channel_id=channel["channel_id"])
               for channel in channels if channel["port_id"] == config.TRANSFER_PORT}

    contracts: set[str] = set()
    for code in chainio.rest_get_all_pages(path="/cosmwasm/wasm/v1/code", key="code_infos"):
        listed = chainio.rest_get_all_pages(path=f"/cosmwasm/wasm/v1/code/{code['code_id']}/contracts", key="contracts")
        # Only a 20-byte contract can collide with a holder; 32-byte ones fail the first rule anyway
        contracts |= {address for address in listed if addresses.is_stride_address(address=address)}
    return SkipInputs(module_addresses=module_addresses, escrows=escrows, contracts=contracts)


def spendable_balances(address: str) -> dict[str, int]:
    coins = chainio.rest_get_all_pages(path=f"/cosmos/bank/v1beta1/spendable_balances/{address}", key="balances")
    return {coin["denom"]: int(coin["amount"]) for coin in coins}


def usd_value(balances: dict[str, int], denoms: dict[str, SweepDenom]) -> Decimal:
    return sum((Decimal(amount) / Decimal(10) ** denoms[denom].decimals * denoms[denom].price_usd
                for denom, amount in balances.items() if denom in denoms), Decimal(0))


# ---- account cache


class AccountCache:
    """Auth lookups keyed by address. A type never changes, so a cached entry is final, except a keyless-looking one
    (no pubkey), which may have signed since and is re-read on every lookup."""

    def __init__(self, entries: dict[str, AccountInfo]) -> None:
        self._entries = entries

    @classmethod
    def load(cls, path: pathlib.Path = config.ACCOUNTS_CACHE_PATH) -> "AccountCache":
        if not path.exists():
            return cls(entries={})
        raw = json.loads(path.read_text())
        return cls(entries={
            address: AccountInfo(type=entry["type"], has_pubkey=entry["has_pubkey"], sequence=int(entry["sequence"]))
            for address, entry in raw.items()
        })

    def save(self, path: pathlib.Path = config.ACCOUNTS_CACHE_PATH) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        raw = {address: {"type": info.type, "has_pubkey": info.has_pubkey, "sequence": str(info.sequence)}
               for address, info in sorted(self._entries.items())}
        tmp = path.with_suffix(".tmp")
        tmp.write_text(json.dumps(raw, indent=1))
        tmp.replace(path)

    def lookup(self, address: str) -> AccountInfo | None:
        cached = self._entries.get(address)
        if cached is not None and cached.has_pubkey:
            return cached
        try:
            body = chainio.rest_get(path=f"/cosmos/auth/v1beta1/accounts/{address}")
        except chainio.NotFound:
            return None
        info = account_info(account=body["account"])
        self._entries[address] = info
        return info


def account_info(account: dict) -> AccountInfo:
    """Vesting, module and interchain accounts nest the base account."""
    base = account.get("base_account") or account.get("base_vesting_account", {}).get("base_account") or account
    return AccountInfo(
        type=account["@type"], has_pubkey=bool(base.get("pub_key")), sequence=int(base.get("sequence", 0)))


# ---- classification


def skip_reason(
    address: str, account: AccountInfo | None, inputs: SkipInputs, exclusions: dict[str, Exclusion]
) -> str | None:
    """The chain's rules in the chain's order, then the builder's two. None means sweepable."""
    if not addresses.is_stride_address(address=address):
        return "address is not 20 bytes"
    if address in config.PROTOCOL_ADDRESSES or address == config.SWEEP_OPERATOR:
        return "protocol address"
    if address in inputs.module_addresses:
        return "blocked module address"
    if address in inputs.escrows:
        return "transfer escrow address"
    if account is None:
        return ACCOUNT_NOT_FOUND
    if account.type == config.INTERCHAIN_ACCOUNT:
        return "interchain account"
    if account.type not in config.SWEEPABLE_ACCOUNT_TYPES:
        return f"account type {account.type} is not sweepable"
    if address in inputs.contracts:
        return "wasm contract address"
    exclusion = exclusions.get(address)
    if exclusion is not None:
        return f"excluded: {exclusion.section}: {exclusion.label}"
    return None


def classify(
    denoms: list[SweepDenom],
    balances: dict[str, dict[str, int]],
    floor_usd: Decimal,
    exclusions: dict[str, Exclusion],
    inputs: SkipInputs,
    accounts: AccountCache,
    test_address: str | None,
    height: int,
) -> HolderSet:
    by_denom = {sweep_denom.denom: sweep_denom for sweep_denom in denoms}
    holder_set = HolderSet(height=height, denoms=denoms, holders=[], excluded=[], skipped=[], below_floor=[])

    for address, coins in balances.items():
        if test_address is not None and address != test_address:
            continue
        usd = usd_value(balances=coins, denoms=by_denom)
        if test_address is None and usd < floor_usd:
            holder_set.below_floor.append(Holder(address=address, balances=coins, usd=usd, keyless=False))
            continue

        # Only a candidate above the floor costs a per-address read; the skip rules need its account
        outcome = _account_or_skip_reason(address=address, inputs=inputs, exclusions=exclusions, accounts=accounts)
        if isinstance(outcome, str):
            target = holder_set.excluded if outcome.startswith("excluded:") else holder_set.skipped
            target.append(Excluded(address=address, reason=outcome, usd=usd))
            continue
        account = outcome

        # The chain moves SpendableCoin: a vesting account's locked part stays behind, and the floor applies to
        # what moves
        is_vesting = account.type in config.VESTING_ACCOUNT_TYPES
        spendable = _spendable_part(address=address, coins=coins) if is_vesting else coins
        spendable_usd = usd_value(balances=spendable, denoms=by_denom) if is_vesting else usd
        if is_vesting and (not spendable or (test_address is None and spendable_usd < floor_usd)):
            holder_set.below_floor.append(Holder(address=address, balances=spendable, usd=spendable_usd, keyless=False))
            continue
        locked = {denom: coins[denom] - spendable.get(denom, 0) for denom in coins if coins[denom] > spendable.get(denom, 0)} if is_vesting else {}
        holder_set.holders.append(Holder(address=address, balances=spendable, usd=spendable_usd,
                                         keyless=not account.has_pubkey and account.sequence == 0, locked=locked))

    holder_set.holders.sort(key=lambda holder: holder.usd, reverse=True)
    holder_set.below_floor.sort(key=lambda holder: holder.usd, reverse=True)
    return holder_set


def _account_or_skip_reason(
    address: str, inputs: SkipInputs, exclusions: dict[str, Exclusion], accounts: AccountCache
) -> AccountInfo | str:
    """The skip reason, or the address's account when it is sweepable. The rules before the account lookup decide
    without it (skip_reason with no account only reaches ACCOUNT_NOT_FOUND once they all pass), so the read is skipped
    when one of them fires."""
    reason = skip_reason(address=address, account=None, inputs=inputs, exclusions=exclusions)
    if reason != ACCOUNT_NOT_FOUND:
        return reason

    account = accounts.lookup(address=address)
    if account is None:
        return ACCOUNT_NOT_FOUND
    return skip_reason(address=address, account=account, inputs=inputs, exclusions=exclusions) or account


def _spendable_part(address: str, coins: dict[str, int]) -> dict[str, int]:
    spendable = spendable_balances(address=address)
    return {denom: spendable.get(denom, 0) for denom in coins if spendable.get(denom, 0) > 0}


def read_holder_set(
    floor_usd: Decimal, exclusions: dict[str, Exclusion], accounts: AccountCache, test_address: str | None
) -> HolderSet:
    height = chainio.latest_height()
    denoms = resolve_denoms()
    balances = read_balances(denoms=denoms)
    inputs = read_skip_inputs()
    return classify(denoms=denoms, balances=balances, floor_usd=floor_usd, exclusions=exclusions, inputs=inputs,
                    accounts=accounts, test_address=test_address, height=height)
