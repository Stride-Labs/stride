"""Command sets for the Multisig tab: the exact generate / sign / multisign+broadcast commands per zone for the
F5 protocol-admin multisig, built from the Validators snapshot. Pure: no chain calls.

Each tx is done end to end, one at a time, online (no pre-assigned sequences): every signer's `tx sign` looks the
multisig's account number and sequence up itself.
"""

import dataclasses
from dataclasses import dataclass
from decimal import Decimal
from typing import Any

import config

MULTISIG_KEY = "F5"  # the multisig's name in the keyring; `multisign` needs it
MULTISIG_ADDRESS = config.PROTOCOL_ADMIN
NODE = "https://stride-strd-rpc.polkachu.com:443"
CHAIN_ID = "stride-1"

WORKDIR = "/tmp/wind-down"
PLACEHOLDER_VALOPER = "<LIVE_TEST_VALOPER>"
ANYONE = "anyone"
NO_SNAPSHOT_REASON = "waiting for the Validators snapshot"
# `tx sign --multisig <address>` resolves the address through the signer's own keyring, so everyone needs the F5 key.
KEYRING_NOTE = f"needs the {MULTISIG_KEY} multisig key in your keyring: strided keys show {MULTISIG_ADDRESS}"

LIVE_TEST_GAS = 12_000_000
# Gas from the mainnet-export measurement (ops plan, drain-rest): the Hub is heaviest, osmosis-1 and ssc-1 next.
FULL_DRAIN_GAS_BY_CHAIN_ID = {"cosmoshub-4": 25_000_000, "osmosis-1": 15_000_000, "ssc-1": 15_000_000}
FULL_DRAIN_DEFAULT_GAS = 13_000_000
GAS_PRICE_USTRD = Decimal("0.005")


@dataclass(frozen=True)
class Signer:
    tag: str  # who they are in the UI
    key: str  # their member key in their own keyring


SAM = Signer(tag="Sam", key="FS5")
AIDAN = Signer(tag="Aidan", key="FA5")
RILEY = Signer(tag="Riley", key="FR5")
SIGNERS = (SAM, AIDAN, RILEY)
DEFAULT_SIGNERS = (SAM, AIDAN)  # any two of the three suffice
BROADCASTER = AIDAN


@dataclass(frozen=True)
class Command:
    tag: str  # who runs it: a signer's tag or "anyone"
    label: str
    text: str


@dataclass(frozen=True)
class MultisigTx:
    chain_id: str
    title: str
    ready: bool  # False with `reason` when the inputs are not known yet
    reason: str | None
    commands: list[Command]  # generate, sign for each signer (default ones, then the backup), multisign+broadcast
    files: list[str]  # the /tmp paths the commands share, for the "share these" note

    def payload(self) -> dict[str, Any]:
        return dataclasses.asdict(self)


@dataclass(frozen=True)
class TxSet:
    id: str
    step_id: str  # the ops step it belongs to
    title: str
    description: str
    txs: list[MultisigTx]  # one per zone, in config.ZONES order

    def payload(self) -> dict[str, Any]:
        return dataclasses.asdict(self)


# ---- public API


def tx_sets(validators_data: dict[str, Any] | None) -> list[TxSet]:
    """Every set, from the validators snapshot's `data` (None before the first snapshot)."""
    zones_by_chain_id = {zone["chain_id"]: zone for zone in (validators_data or {}).get("zones", [])}
    snapshot_reason = None if validators_data is not None else NO_SNAPSHOT_REASON

    live_test = TxSet(
        id="live-test-undelegate",
        step_id="drain-live-test",
        title="Live test: undelegate one validator per zone",
        description=(
            "Per zone, MsgUndelegateFromValidators for the single live-test validator, a full drain of it. Watch the "
            "ack and the callback: its recorded delegation drops to zero on the Validators tab and the Ops step's "
            "check turns green once the unbonding entry lands. Do one tx at a time, end to end: generate, two "
            "signatures, multisign and broadcast, then confirm it landed before starting the next zone."
        ),
        txs=[
            _live_test_tx(zone=zone, snapshot_zone=zones_by_chain_id.get(zone.chain_id), snapshot_reason=snapshot_reason)
            for zone in config.ZONES
        ],
    )
    full_drain = TxSet(
        id="full-drain",
        step_id="drain-rest",
        title="Full drain: undelegate every remaining validator",
        description=(
            "Per zone, MsgUndelegateFromValidators with --all for the validators left after the live test. Relay the "
            "acks before the timeouts and wait for every batch to ack. Do one tx at a time, end to end: generate, two "
            "signatures, multisign and broadcast, then confirm it landed before starting the next zone."
        ),
        txs=[
            _full_drain_tx(zone=zone, snapshot_zone=zones_by_chain_id.get(zone.chain_id), snapshot_reason=snapshot_reason)
            for zone in config.ZONES
        ],
    )
    return [live_test, full_drain]


# ---- txs


def _live_test_tx(zone: config.ZoneConfig, snapshot_zone: dict[str, Any] | None, snapshot_reason: str | None) -> MultisigTx:
    pick = snapshot_zone.get("live_test_pick") if snapshot_zone else None
    reason = snapshot_reason or _zone_error(snapshot_zone=snapshot_zone) or (None if pick else snapshot_zone["live_test_reason"])

    valoper = pick["address"] if pick else PLACEHOLDER_VALOPER
    file_stem = f"{WORKDIR}/live-test-{zone.chain_id}"
    validators_file = f"{file_stem}.json"
    generate = (
        f"mkdir -p {WORKDIR}\n"
        f"echo '[{{\"address\": \"{valoper}\", \"offset\": \"0\"}}]' > {validators_file}\n"
        f"{_generate_line(arguments=f'{zone.chain_id} {validators_file}', gas=LIVE_TEST_GAS, file_stem=file_stem)}"
    )
    return MultisigTx(
        chain_id=zone.chain_id,
        title=f"{zone.chain_id} · live test: {_pick_description(pick=pick, zone=zone)}",
        ready=reason is None,
        reason=reason,
        commands=_commands(generate=generate, file_stem=file_stem),
        files=_shared_files(file_stem=file_stem),
    )


def _full_drain_tx(zone: config.ZoneConfig, snapshot_zone: dict[str, Any] | None, snapshot_reason: str | None) -> MultisigTx:
    reason = snapshot_reason or _zone_error(snapshot_zone=snapshot_zone)

    gas = FULL_DRAIN_GAS_BY_CHAIN_ID.get(zone.chain_id, FULL_DRAIN_DEFAULT_GAS)
    file_stem = f"{WORKDIR}/full-drain-{zone.chain_id}"
    generate = (
        f"mkdir -p {WORKDIR}\n"
        f"{_generate_line(arguments=f'{zone.chain_id} --all', gas=gas, file_stem=file_stem)}"
    )
    return MultisigTx(
        chain_id=zone.chain_id,
        title=f"{zone.chain_id} · full drain, every remaining validator",
        ready=reason is None,
        reason=reason,
        commands=_commands(generate=generate, file_stem=file_stem),
        files=_shared_files(file_stem=file_stem),
    )


def _commands(generate: str, file_stem: str) -> list[Command]:
    unsigned = f"{file_stem}.unsigned.json"
    sign_commands = [
        Command(
            tag=signer.tag,
            label=_sign_label(signer=signer),
            text=(
                f"strided tx sign {unsigned} --multisig {MULTISIG_ADDRESS} --from {signer.key} "
                f"--chain-id {CHAIN_ID} --node {NODE} \\\n  --output-document {_signature_file(file_stem=file_stem, signer=signer)}"
            ),
        )
        for signer in SIGNERS
    ]
    signature_files = " ".join(_signature_file(file_stem=file_stem, signer=signer) for signer in DEFAULT_SIGNERS)
    combine = Command(
        tag=BROADCASTER.tag,
        label="Combine and broadcast",
        text=(
            f"strided tx multisign {unsigned} {MULTISIG_KEY} {signature_files} --chain-id {CHAIN_ID} --node {NODE} "
            f"> {file_stem}.signed.json\n"
            f"strided tx broadcast {file_stem}.signed.json --node {NODE} --broadcast-mode sync"
        ),
    )
    return [Command(tag=ANYONE, label="Write the validators file and the unsigned tx", text=generate), *sign_commands, combine]


def _sign_label(signer: Signer) -> str:
    if signer == SAM:
        return f"Sign (online: the multisig's account number and sequence are looked up; {KEYRING_NOTE})"
    if signer in DEFAULT_SIGNERS:
        return f"Sign ({KEYRING_NOTE})"
    return f"Backup signer (any two signatures suffice; {KEYRING_NOTE})"


def _generate_line(arguments: str, gas: int, file_stem: str) -> str:
    fees = int(gas * GAS_PRICE_USTRD)
    return (
        f"strided tx stakeibc undelegate-from-validators {arguments} --from {MULTISIG_ADDRESS} --generate-only \\\n"
        f"  --chain-id {CHAIN_ID} --node {NODE} --gas {gas} --fees {fees}ustrd > {file_stem}.unsigned.json"
    )


def _signature_file(file_stem: str, signer: Signer) -> str:
    return f"{file_stem}.{signer.key}.json"


def _shared_files(file_stem: str) -> list[str]:
    return [f"{file_stem}.unsigned.json", *(_signature_file(file_stem=file_stem, signer=signer) for signer in SIGNERS)]


# ---- helpers


def _zone_error(snapshot_zone: dict[str, Any] | None) -> str | None:
    if snapshot_zone is None:
        return "zone missing from the Validators snapshot"
    return snapshot_zone.get("error")


def _pick_description(pick: dict[str, str] | None, zone: config.ZoneConfig) -> str:
    if pick is None:
        return PLACEHOLDER_VALOPER

    amount = int(pick["recorded"])
    shown = (
        f"{amount:,} u{zone.symbol.lower()}"
        if zone.decimals == 6
        else f"{Decimal(amount).scaleb(-zone.decimals).normalize():,f} {zone.symbol}"
    )
    return f"{pick['moniker']} ({_short_address(address=pick['address'])}), {shown}"


def _short_address(address: str) -> str:
    # bech32 data never contains "1", so the last one separates the prefix from the data.
    prefix_end = address.rfind("1") + 1
    return f"{address[:prefix_end + 3]}…"
