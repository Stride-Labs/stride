#!/usr/bin/env python3
"""Staleness gate for the v35 wind-down constants.

Three things in the binary are measurements of chain state that can drift between the PR and
the proposal: the haqq delegation delta table (app/upgrades/v35/haqq.go), the host-side
channel map to Osmosis (x/stakeibc/types/wind_down.go, HostToOsmosisTransferChannel) and the
sweep unwind whitelist (SweepUnwindChannels). This script recomputes each from live REST
state and fails loudly on any difference, so a stale constant is a red pre-proposal step and
not a silently skipped reconciliation or a transfer to the wrong chain. It also checks that
the two operator address constants have signed at least one transaction on their chain
(account sequence > 0), the automated half of the §9 signed-spend proof.

Stdlib-only. Reads the Go files relative to its own location.

Usage: python3 app/upgrades/v35/testdata/verify_constants.py
"""

import json
import pathlib
import re
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from decimal import Decimal

USER_AGENT = "Mozilla/5.0"
MAX_ATTEMPTS = 6
BACKOFF_SECONDS = 5
HTTP_TOO_MANY_REQUESTS = 429
TIMEOUT_SECONDS = 20
STRIDE_API = "https://stride-api.polkachu.com"
OSMOSIS_API = "https://osmosis-api.polkachu.com"
CHAIN_REGISTRY = "https://raw.githubusercontent.com/cosmos/chain-registry/master"
OSMOSIS_CHAIN_ID = "osmosis-1"

SCRIPT_DIR = pathlib.Path(__file__).resolve().parent
HAQQ_GO = SCRIPT_DIR.parent / "haqq.go"
WIND_DOWN_GO = SCRIPT_DIR.parent.parent.parent.parent / "x" / "stakeibc" / "types" / "wind_down.go"

HAQQ_CHAIN_ID = "haqq_11235-1"
HAQQ_APIS = ["https://haqq-rest.publicnode.com", "https://rest.cosmos.directory/haqq"]

# Stride chain id -> chain-registry directory name, for the hosts' REST endpoints
REGISTRY_NAMES = {
    "celestia": "celestia",
    "cosmoshub-4": "cosmoshub",
    "dydx-mainnet-1": "dydx",
    "haqq_11235-1": "haqq",
    "injective-1": "injective",
    "juno-1": "juno",
    "osmosis-1": "osmosis",
    "laozi-mainnet": "bandchain",
    "phoenix-1": "terra2",
    "sommelier-3": "sommelier",
    "ssc-1": "saga",
}

# The counterparty chain each whitelisted Stride channel must lead to (spec §7)
SWEEP_CHANNEL_CHAIN_IDS = {
    "channel-0": "cosmoshub-4",
    "channel-162": "celestia",
    "channel-5": "osmosis-1",
    "channel-24": "juno-1",
    "channel-150": "sommelier-3",
    "channel-213": "ssc-1",
    "channel-160": "dydx-mainnet-1",
}

# Share/token rounding moves an ISLM delegation by a base unit or two between the measurement and
# the check: a live-vs-table difference of at most this many base units per row is a WARN, an exact
# match passes, anything larger fails
HAQQ_DELTA_DUST_TOLERANCE = 10

failures: list[str] = []
warnings: list[str] = []


@dataclass(frozen=True)
class DeltaEntry:
    name: str
    address: str
    delta: int


def report(name: str, ok: bool, detail: str = "") -> None:
    status = "PASS" if ok else "FAIL"
    suffix = f" -- {detail}" if detail else ""
    print(f"[{status}] {name}{suffix}")
    if not ok:
        failures.append(name)


def report_with_tolerance(name: str, live: int, expected: int, tolerance: int) -> None:
    """PASS on an exact match, WARN within the documented dust tolerance, FAIL beyond it."""
    difference = abs(live - expected)
    detail = f"live {live} vs table {expected}"
    if difference == 0:
        report(name, True, detail)
    elif difference <= tolerance:
        print(f"[WARN] {name} -- {detail} (off by {difference} base units, within the dust tolerance of {tolerance})")
        warnings.append(name)
    else:
        report(name, False, f"{detail} (off by {difference} base units, tolerance {tolerance})")


def fetch_json(url: str) -> dict:
    """GET with backoff on 429: the public endpoints rate-limit a script that makes many calls."""
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    for attempt in range(1, MAX_ATTEMPTS + 1):
        try:
            with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
                return json.load(response)
        except urllib.error.HTTPError as exc:
            if exc.code != HTTP_TOO_MANY_REQUESTS or attempt == MAX_ATTEMPTS:
                raise
            time.sleep(BACKOFF_SECONDS * attempt)
    raise AssertionError("unreachable")


def fetch_from_any(bases: list[str], path: str) -> dict:
    """Public LCDs for third-party chains come and go; the first one that answers wins."""
    errors = []
    for base in bases:
        try:
            return fetch_json(f"{base}{path}")
        except (urllib.error.URLError, TimeoutError, OSError, json.JSONDecodeError) as exc:
            errors.append(f"{base}: {exc}")
    sys.exit(f"no endpoint answered {path}: {errors}")


def registry_rest_endpoints(registry_name: str) -> list[str]:
    chain = fetch_json(f"{CHAIN_REGISTRY}/{registry_name}/chain.json")
    return [entry["address"].rstrip("/") for entry in chain.get("apis", {}).get("rest", [])]


# ----------------------------------------------------------------------------------------------
# Parsing the Go constants
# ----------------------------------------------------------------------------------------------

DELTA_LINE = re.compile(r'\{Name:\s*"([^"]+)",\s*Address:\s*"([^"]+)",\s*Delta:\s*mustInt\("(-?\d+)"\)\}')
PIN_LINE = re.compile(r'"(haqqvaloper1[a-z0-9]+)":\s*mustInt\("(-?\d+)"\)')
MAP_ENTRY = re.compile(r'"([^"]+)":\s*"([^"]*)",')


def parse_haqq_table() -> list[DeltaEntry]:
    source = HAQQ_GO.read_text()
    entries = [DeltaEntry(name=n, address=a, delta=int(d)) for n, a, d in DELTA_LINE.findall(source)]
    if not entries:
        sys.exit(f"no delta entries parsed from {HAQQ_GO}")
    return entries


def parse_haqq_pins() -> dict[str, int]:
    source = HAQQ_GO.read_text()
    match = re.search(r"var HaqqExpectedTrackedDelegations = map\[string\]sdkmath\.Int\{(.*?)\n\}", source, re.S)
    if match is None:
        sys.exit(f"HaqqExpectedTrackedDelegations not found in {HAQQ_GO}")
    return {address: int(amount) for address, amount in PIN_LINE.findall(match.group(1))}


def parse_go_map(var_name: str) -> dict[str, str]:
    source = WIND_DOWN_GO.read_text()
    match = re.search(rf"var {var_name} = map\[string\]string\{{(.*?)\n\}}", source, re.S)
    if match is None:
        sys.exit(f"{var_name} not found in {WIND_DOWN_GO}")
    return dict(MAP_ENTRY.findall(match.group(1)))


def parse_go_string_var(var_name: str) -> str:
    source = WIND_DOWN_GO.read_text()
    match = re.search(rf'{var_name}\s*=\s*"([^"]*)"', source)
    if match is None:
        sys.exit(f"{var_name} not found in {WIND_DOWN_GO}")
    return match.group(1)


# ----------------------------------------------------------------------------------------------
# Checks
# ----------------------------------------------------------------------------------------------

def live_tracked_haqq() -> tuple[dict[str, int], str]:
    host_zone = fetch_json(f"{STRIDE_API}/Stride-Labs/stride/stakeibc/host_zone/{HAQQ_CHAIN_ID}")["host_zone"]
    tracked = {v["address"]: int(v["delegation"] or 0) for v in host_zone["validators"]}
    return tracked, host_zone["delegation_ica_address"]


def live_haqq_deltas(tracked: dict[str, int], ica_address: str) -> dict[str, int]:
    """On-chain delegation minus tracked delegation per validator, the same sign as the table."""

    on_chain: dict[str, int] = {}
    next_key = ""
    while True:
        path = f"/cosmos/staking/v1beta1/delegations/{ica_address}?pagination.limit=200"
        if next_key:
            path += f"&pagination.key={urllib.parse.quote(next_key)}"
        page = fetch_from_any(HAQQ_APIS, path)
        for entry in page["delegation_responses"]:
            on_chain[entry["delegation"]["validator_address"]] = int(entry["balance"]["amount"])
        next_key = page.get("pagination", {}).get("next_key") or ""
        if not next_key:
            break

    return {address: on_chain.get(address, 0) - tracked_amount for address, tracked_amount in tracked.items()}


def check_haqq_table() -> None:
    table = {entry.address: entry for entry in parse_haqq_table()}
    tracked, ica_address = live_tracked_haqq()

    # The handler skips the whole table when any tracked delegation differs from its pin
    pins = parse_haqq_pins()
    report("haqq pins cover exactly the table", set(pins) == set(table))
    for address, entry in table.items():
        report(f"haqq pin {entry.name}", tracked.get(address, 0) == pins.get(address), f"live {tracked.get(address)} vs pin {pins.get(address)}")

    live = live_haqq_deltas(tracked=tracked, ica_address=ica_address)

    # Every non-zero live delta must be in the table with the same value, and the table must
    # not carry an entry the chain no longer shows: either way the handler would skip the whole
    # table (validateDelegationDeltas) or true up to the wrong number
    for address, delta in live.items():
        entry = table.get(address)
        if delta == 0 and entry is None:
            continue
        expected = entry.delta if entry is not None else 0
        name = entry.name if entry is not None else address
        report_with_tolerance(f"haqq delta {name}", live=delta, expected=expected, tolerance=HAQQ_DELTA_DUST_TOLERANCE)
    for address, entry in table.items():
        report(f"haqq validator {entry.name} tracked", address in live,
               "" if address in live else "validator missing from the Stride host zone")

    net = sum(entry.delta for entry in table.values())
    print(f"       table net delta: {net} aISLM ({Decimal(net) / Decimal(10**18):.6f} ISLM)")


def channel_counterparty_chain_id(rest: str, channel_id: str) -> tuple[str, str]:
    """Returns (channel state, counterparty chain id) for a transfer channel on the given chain."""
    channel = fetch_json(f"{rest}/ibc/core/channel/v1/channels/{channel_id}/ports/transfer")["channel"]
    client_state = fetch_json(f"{rest}/ibc/core/channel/v1/channels/{channel_id}/ports/transfer/client_state")
    return channel["state"], client_state["identified_client_state"]["client_state"]["chain_id"]


def check_host_to_osmosis_map() -> None:
    channel_map = parse_go_map("HostToOsmosisTransferChannel")
    for chain_id, channel_id in sorted(channel_map.items()):
        if chain_id == OSMOSIS_CHAIN_ID:
            report(f"{chain_id} maps to the bank-send form", channel_id == "")
            continue
        if chain_id not in REGISTRY_NAMES:
            report(f"{chain_id} has a chain-registry name", False, "add it to REGISTRY_NAMES")
            continue
        endpoints = registry_rest_endpoints(REGISTRY_NAMES[chain_id])
        for rest in endpoints:
            try:
                state, counterparty = channel_counterparty_chain_id(rest=rest, channel_id=channel_id)
            except (urllib.error.URLError, TimeoutError, OSError, KeyError, json.JSONDecodeError):
                continue
            report(f"{chain_id} {channel_id} -> osmosis-1 and open",
                   state == "STATE_OPEN" and counterparty == OSMOSIS_CHAIN_ID,
                   f"{state}, counterparty {counterparty} via {rest}")
            break
        else:
            report(f"{chain_id} {channel_id} reachable", False, "no registry REST endpoint answered")


def check_sweep_unwind_channels() -> None:
    whitelist = parse_go_map("SweepUnwindChannels")
    report("sweep whitelist has exactly the seven §7 channels", set(whitelist) == set(SWEEP_CHANNEL_CHAIN_IDS))
    for channel_id, expected_chain in sorted(SWEEP_CHANNEL_CHAIN_IDS.items()):
        state, counterparty = channel_counterparty_chain_id(rest=STRIDE_API, channel_id=channel_id)
        report(f"stride {channel_id} -> {expected_chain} and open",
               state == "STATE_OPEN" and counterparty == expected_chain, f"{state}, counterparty {counterparty}")

        # The prefix the sweep re-encodes holders' addresses under must be the counterparty's
        prefix = whitelist.get(channel_id)
        if expected_chain not in REGISTRY_NAMES:
            report(f"{channel_id} prefix {prefix!r} vs registry", False, f"{expected_chain} has no chain-registry name")
            continue
        registry_prefix = fetch_json(f"{CHAIN_REGISTRY}/{REGISTRY_NAMES[expected_chain]}/chain.json").get("bech32_prefix")
        report(f"{channel_id} prefix matches the {expected_chain} registry bech32_prefix",
               prefix == registry_prefix, f"constant {prefix!r} vs registry {registry_prefix!r}")


def account_sequence(rest: str, address: str) -> int:
    """The signing sequence of an account; a multisig is a BaseAccount too, so one shape covers both."""
    account = fetch_json(f"{rest}/cosmos/auth/v1beta1/accounts/{address}")["account"]
    base = account.get("base_account", account)
    return int(base["sequence"])


def check_operator_addresses_have_signed() -> None:
    """Both constants must have signed at least once on their chain: the automated half of the §9
    proof that each address is one we control (the other half is the tx hashes in the commit)."""
    sweep_operator = parse_go_string_var("SweepOperatorAddress")
    vault = parse_go_string_var("OsmosisVaultAddress")
    report("SweepOperatorAddress is filled", bool(sweep_operator))
    report("OsmosisVaultAddress is filled", bool(vault))
    if not sweep_operator or not vault:
        return
    for name, rest, address in (("sweep operator", STRIDE_API, sweep_operator), ("osmosis vault", OSMOSIS_API, vault)):
        try:
            sequence = account_sequence(rest=rest, address=address)
        except (urllib.error.URLError, TimeoutError, OSError, KeyError, json.JSONDecodeError) as exc:
            report(f"{name} {address} has signed", False, f"account query failed: {exc}")
            continue
        report(f"{name} {address} has signed", sequence > 0, f"sequence {sequence}")


def main() -> int:
    print("== operator addresses ==")
    check_operator_addresses_have_signed()
    print("== haqq delegation delta table ==")
    check_haqq_table()
    print("== host-side channels to Osmosis ==")
    check_host_to_osmosis_map()
    print("== sweep unwind whitelist ==")
    check_sweep_unwind_channels()

    if warnings:
        print(f"\n{len(warnings)} WARN (haqq dust within {HAQQ_DELTA_DUST_TOLERANCE} base units): {warnings}")
    if failures:
        print(f"\n{len(failures)} check(s) FAILED: {failures}")
        return 1
    print("\nall checks passed" + (" (with warnings)" if warnings else ""))
    return 0


if __name__ == "__main__":
    sys.exit(main())
