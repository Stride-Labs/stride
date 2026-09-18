#!/usr/bin/env python3
"""Mechanical staleness gate for the v34 Celestia/Cosmos Hub reconciliation constants.

The v34 upgrade handler cannot validate its numeric constants (CelestiaDelegationDeltas,
StaketiaRemainingDelegatedBalanceDelta, CosmosHubStrandedLsmDeposit) against on-chain state at
run time: tracked delegations legitimately move as daily delegations are acknowledged between
now and the upgrade, so an on-chain fingerprint baked into the handler would false-skip a
perfectly valid reconciliation. This script is the substitute: run it right before cutting the
release to recompute every constant straight from the live chain and from the checked-in
mainnet_export.json.gz fixture, and fail loudly if either has drifted.

Stdlib-only. Reads celestia.go and cosmoshub.go relative to its own location, so it has no
dependency on any path outside this repo.

Usage: python3 app/upgrades/v34/testdata/verify_constants.py
"""

import decimal
import gzip
import json
import re
import sys
import urllib.error
import urllib.request
from pathlib import Path

USER_AGENT = "curl/8.0"
TIMEOUT_SECONDS = 20

SCRIPT_DIR = Path(__file__).resolve().parent
V34_DIR = SCRIPT_DIR.parent
CELESTIA_GO = V34_DIR / "celestia.go"
COSMOSHUB_GO = V34_DIR / "cosmoshub.go"
FIXTURE_GZ = SCRIPT_DIR / "mainnet_export.json.gz"

STRIDE_API = "https://stride-api.polkachu.com"
CELESTIA_APIS = [
    "https://api.celestia.pops.one",
    "https://celestia-rest.publicnode.com",
    "https://public-celestia-lcd.numia.xyz",
]
COSMOSHUB_APIS = [
    "https://cosmos-rest.publicnode.com",
    "https://rest.cosmos.directory/cosmoshub",
    "https://cosmos.rpc.uquad.org:443",
]

CELESTIA_CHAIN_ID = "celestia"
COSMOSHUB_CHAIN_ID = "cosmoshub-4"
ACCUMULATING_UNBONDING_STATUSES = {"ACCUMULATING_REDEMPTIONS", "UNBONDING_QUEUE"}

failures = []


def report(name: str, ok: bool, detail: str = "") -> None:
    status = "PASS" if ok else "FAIL"
    suffix = f" -- {detail}" if detail else ""
    print(f"[{status}] {name}{suffix}")
    if not ok:
        failures.append(name)


def fetch_json(url: str) -> dict:
    request = urllib.request.Request(url, headers={"User-Agent": USER_AGENT})
    with urllib.request.urlopen(request, timeout=TIMEOUT_SECONDS) as response:
        return json.load(response)


def fetch_from_any(bases: list[str], path: str) -> tuple[dict, str]:
    """Tries each base URL in order and returns the first successful response plus which base
    answered, since public LCD endpoints for third-party chains come and go."""
    errors = []
    for base in bases:
        try:
            return fetch_json(f"{base}{path}"), base
        except (
            urllib.error.URLError,
            urllib.error.HTTPError,
            TimeoutError,
            OSError,
        ) as exc:
            errors.append(f"{base}: {exc}")
    raise RuntimeError(f"every endpoint failed for {path}: {'; '.join(errors)}")


def fetch_all_delegations(bases: list[str], delegator_address: str) -> dict[str, int]:
    """Pages through /cosmos/staking/v1beta1/delegations/<address> and returns validator address
    -> delegated amount, on whichever base URL answers first."""
    base = None
    delegations = {}
    pagination_key = ""
    while True:
        suffix = f"&pagination.key={pagination_key}" if pagination_key else ""
        path = f"/cosmos/staking/v1beta1/delegations/{delegator_address}?pagination.limit=500{suffix}"
        if base is None:
            page, base = fetch_from_any(bases, path)
        else:
            page = fetch_json(f"{base}{path}")
        for entry in page["delegation_responses"]:
            delegations[entry["delegation"]["validator_address"]] = int(
                entry["balance"]["amount"]
            )
        pagination_key = page.get("pagination", {}).get("next_key") or ""
        if not pagination_key:
            break
    return delegations


# ---------- Parse the Go constants ----------


def parse_celestia_deltas() -> list[tuple[str, str, int]]:
    source = CELESTIA_GO.read_text()
    block = re.search(
        r"var CelestiaDelegationDeltas = \[\]DelegationDelta\{(.*?)\n\}", source, re.S
    ).group(1)
    row_re = re.compile(
        r'\{Name:\s*"([^"]*)",\s*Address:\s*"([^"]*)",\s*Delta:\s*mustInt\("(-?\d+)"\)\}'
    )
    return [
        (name, address, int(delta)) for name, address, delta in row_re.findall(block)
    ]


def parse_staketia_delta() -> int:
    source = CELESTIA_GO.read_text()
    match = re.search(
        r"var StaketiaRemainingDelegatedBalanceDelta = sdkmath\.NewInt\((-?[\d_]+)\)",
        source,
    )
    return int(match.group(1).replace("_", ""))


def parse_hub_stranded_deposit() -> tuple[str, str, int]:
    source = COSMOSHUB_GO.read_text()
    match = re.search(
        r'Denom:\s*"([^"]*)",\s*\n\s*ValidatorAddress:\s*"([^"]*)",[^\n]*\n\s*Amount:\s*sdkmath\.NewInt\(([\d_]+)\)',
        source,
    )
    denom, validator_address, amount_raw = match.groups()
    return denom, validator_address, int(amount_raw.replace("_", ""))


# ---------- Checks against live chain state ----------


def check_celestia_deltas(table: list[tuple[str, str, int]]) -> None:
    host_zone = fetch_json(
        f"{STRIDE_API}/Stride-Labs/stride/stakeibc/host_zone/{CELESTIA_CHAIN_ID}"
    )["host_zone"]
    tracked = {v["address"]: int(v["delegation"]) for v in host_zone["validators"]}
    delegation_ica = host_zone["delegation_ica_address"]

    onchain = fetch_all_delegations(CELESTIA_APIS, delegation_ica)
    live_deltas = {}
    for address in set(tracked) | set(onchain):
        delta = onchain.get(address, 0) - tracked.get(address, 0)
        if delta != 0:
            live_deltas[address] = delta

    table_by_address = {address: (name, delta) for name, address, delta in table}
    mismatches = []
    for address, (name, delta) in table_by_address.items():
        live = live_deltas.get(address)
        if live != delta:
            mismatches.append(f"{name} ({address}): table={delta} live={live}")
    missing_from_table = set(live_deltas) - set(table_by_address)
    if missing_from_table:
        mismatches.append(
            f"live non-zero deltas missing from Go table: {sorted(missing_from_table)}"
        )

    report(
        "celestia per-validator deltas match live chain",
        not mismatches,
        "; ".join(mismatches)
        if mismatches
        else f"{len(table)} validators, all deltas match",
    )


def check_staketia_delta(constant_delta: int) -> None:
    staketia_host_zone = fetch_json(
        f"{STRIDE_API}/Stride-Labs/stride/staketia/host_zone"
    )["host_zone"]
    remaining_delegated_balance = int(staketia_host_zone["remaining_delegated_balance"])
    multisig_address = staketia_host_zone["delegation_address"]

    multisig_delegations = fetch_all_delegations(CELESTIA_APIS, multisig_address)
    multisig_total = sum(multisig_delegations.values())

    unbonding_records = fetch_json(
        f"{STRIDE_API}/Stride-Labs/stride/staketia/unbonding_records?pagination.limit=1000"
    )["unbonding_records"]
    owed = sum(
        int(record["native_amount"])
        for record in unbonding_records
        if record["status"] in ACCUMULATING_UNBONDING_STATUSES
    )

    live_delta = (multisig_total - owed) - remaining_delegated_balance
    report(
        "staketia remaining_delegated_balance delta matches formula",
        live_delta == constant_delta,
        f"live={live_delta} constant={constant_delta} "
        f"(multisig_total={multisig_total}, owed={owed}, remaining={remaining_delegated_balance})",
    )


def check_hub_lsm_deposit(denom: str, validator_address: str, amount: int) -> None:
    lsm_response = fetch_json(
        f"{STRIDE_API}/Stride-Labs/stride/stakeibc/lsm_deposits?chain_id={COSMOSHUB_CHAIN_ID}"
    )
    deposits = lsm_response.get("deposits", lsm_response.get("lsm_token_deposits", []))
    matches = [d for d in deposits if d["denom"] == denom]

    if len(matches) != 1:
        report(
            "hub LSM deposit found by denom",
            False,
            f"found {len(matches)} deposits for denom {denom}, expected 1",
        )
        return
    deposit = matches[0]
    ok = (
        deposit["status"] == "DETOKENIZATION_FAILED"
        and deposit["validator_address"] == validator_address
        and int(deposit["amount"]) == amount
    )
    report(
        "hub LSM deposit status/validator/amount match constants",
        ok,
        f"live status={deposit['status']} validator={deposit['validator_address']} amount={deposit['amount']}",
    )


def check_hub_delegation_gap(
    validator_address: str, expected_gap: int
) -> decimal.Decimal | None:
    hub_host_zone = fetch_json(
        f"{STRIDE_API}/Stride-Labs/stride/stakeibc/host_zone/{COSMOSHUB_CHAIN_ID}"
    )["host_zone"]
    delegation_ica = hub_host_zone["delegation_ica_address"]
    tracked_validator = next(
        (v for v in hub_host_zone["validators"] if v["address"] == validator_address),
        None,
    )
    if tracked_validator is None:
        report(
            "hub validator present on host zone",
            False,
            f"{validator_address} not found",
        )
        return None
    tracked_delegation = int(tracked_validator["delegation"])
    shares_to_tokens_rate = decimal.Decimal(tracked_validator["shares_to_tokens_rate"])

    onchain = fetch_all_delegations(COSMOSHUB_APIS, delegation_ica)
    onchain_delegation = onchain.get(validator_address, 0)
    live_gap = onchain_delegation - tracked_delegation

    report(
        "hub stakewithus delegation gap matches CosmosHubStrandedLsmDeposit.Amount",
        live_gap == expected_gap,
        f"live_gap={live_gap} constant={expected_gap} (onchain={onchain_delegation}, tracked={tracked_delegation})",
    )
    return shares_to_tokens_rate


def check_hub_tokenized_rate_guard(
    amount: int, shares_to_tokens_rate: decimal.Decimal | None
) -> None:
    """Mirrors GetTotalTokenizedDelegations' per-deposit valuation (amount * rate, truncated) and
    checks it still equals the raw amount CloseCosmosHubLsmDeposit books as native delegation --
    the invariant the redemption-rate guard in CloseCosmosHubLsmDeposit depends on."""
    if shares_to_tokens_rate is None:
        report(
            "hub SharesToTokensRate is 1.0 (LSM-deposit bucket move stays rate-neutral)",
            False,
            "rate unavailable",
        )
        return
    tokenized_native_value = int(
        (decimal.Decimal(amount) * shares_to_tokens_rate).to_integral_value(
            rounding=decimal.ROUND_DOWN
        )
    )
    report(
        "hub SharesToTokensRate is 1.0 (LSM-deposit bucket move stays rate-neutral)",
        tokenized_native_value == amount,
        f"rate={shares_to_tokens_rate} amount={amount} tokenized_native_value={tokenized_native_value}",
    )


# ---------- Checks against the committed fixture ----------


def load_fixture() -> dict:
    with gzip.open(FIXTURE_GZ) as fixture_file:
        return json.load(fixture_file)


def check_fixture_celestia_records_cover_phantom(
    fixture: dict, phantom_amount: int
) -> None:
    coverable_statuses = {"DELEGATION_QUEUE", "DELEGATION_IN_PROGRESS"}
    coverable = sum(
        int(record["amount"])
        for record in fixture["app_state"]["records"]["deposit_record_list"]
        if record["host_zone_id"] == CELESTIA_CHAIN_ID
        and record["status"] in coverable_statuses
    )
    report(
        "fixture celestia deposit records cover the phantom amount",
        coverable >= phantom_amount,
        f"coverable={coverable} phantom_amount={phantom_amount}",
    )


def check_fixture_hub_lsm_deposit(
    fixture: dict, denom: str, validator_address: str, amount: int
) -> None:
    matches = [
        d
        for d in fixture["app_state"]["records"]["lsm_token_deposit_list"]
        if d["denom"] == denom
    ]
    if len(matches) != 1:
        report(
            "fixture carries exactly one hub LSM deposit for the stranded denom",
            False,
            f"found {len(matches)}",
        )
        return
    deposit = matches[0]
    ok = (
        deposit["status"] == "DETOKENIZATION_FAILED"
        and deposit["validator_address"] == validator_address
        and int(deposit["amount"]) == amount
    )
    report(
        "fixture hub LSM deposit status/validator/amount match constants",
        ok,
        f"fixture status={deposit['status']} validator={deposit['validator_address']} amount={deposit['amount']}",
    )


def main() -> int:
    celestia_deltas = parse_celestia_deltas()
    staketia_delta = parse_staketia_delta()
    hub_denom, hub_validator_address, hub_amount = parse_hub_stranded_deposit()
    phantom_amount = sum(delta for _, _, delta in celestia_deltas)

    print(
        f"parsed {len(celestia_deltas)} celestia delta rows (phantom amount {phantom_amount})"
    )
    print(f"parsed staketia delta {staketia_delta}")
    print(
        f"parsed hub stranded deposit denom={hub_denom} validator={hub_validator_address} amount={hub_amount}\n"
    )

    check_celestia_deltas(celestia_deltas)
    check_staketia_delta(staketia_delta)
    check_hub_lsm_deposit(hub_denom, hub_validator_address, hub_amount)
    shares_to_tokens_rate = check_hub_delegation_gap(hub_validator_address, hub_amount)
    check_hub_tokenized_rate_guard(hub_amount, shares_to_tokens_rate)

    fixture = load_fixture()
    check_fixture_celestia_records_cover_phantom(fixture, phantom_amount)
    check_fixture_hub_lsm_deposit(fixture, hub_denom, hub_validator_address, hub_amount)

    print()
    if failures:
        print(f"{len(failures)} check(s) FAILED: {failures}")
        return 1
    print("all checks PASSED")
    return 0


if __name__ == "__main__":
    sys.exit(main())
