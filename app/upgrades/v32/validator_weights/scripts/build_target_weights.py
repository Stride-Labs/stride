"""
Parses the raw target CSVs and current_weights.csv to produce:
  - target_weights.csv:  only validators with a non-zero target weight (basis points)
  - all_validators.csv:  union of current and target validators, with target weight
                         (0 for any current validator not in the target)
  - new_validators.csv:  validators in the target that don't exist on-chain yet
                         (starting point — duplicate and edit names before using in Go)

Also validates:
  a) target weights sum to exactly 10000 per chain
  b) every non-zero entry in all_validators matches target_weights
"""

import csv
import sys
from decimal import Decimal
from pathlib import Path

WEIGHTS_DIR = Path(__file__).parent.parent
RAW_DIR = WEIGHTS_DIR / "target" / "raw"
CURRENT_WEIGHTS_PATH = WEIGHTS_DIR / "current_weights.csv"
TARGET_WEIGHTS_PATH = WEIGHTS_DIR / "target_weights.csv"
ALL_VALIDATORS_PATH = WEIGHTS_DIR / "all_validators.csv"
NEW_VALIDATORS_PATH = WEIGHTS_DIR / "new_validators_raw.csv"

TARGET_SUM = 10000


def parse_chain_id_from_filename(filename: str) -> str:
    return filename.replace("Stride Delegations Q1 2026 - ", "").replace(".csv", "")


def parse_raw_csvs() -> dict[str, list[dict]]:
    """Returns {chain_id: [{address, name, weight_bps}, ...]} for non-empty New Weight rows."""
    result: dict[str, list[dict]] = {}

    for csv_path in sorted(RAW_DIR.glob("*.csv")):
        chain_id = parse_chain_id_from_filename(csv_path.name)
        validators = []

        with csv_path.open() as f:
            reader = csv.DictReader(f)
            for row in reader:
                raw_weight = row.get("New Weight", "").strip().replace("%", "")
                if not raw_weight:
                    continue

                weight_bps = int(Decimal(raw_weight) * 100)
                if weight_bps > 0:
                    validators.append({
                        "address": row["Validator Address"].strip(),
                        "name": row["Validator Name"].strip(),
                        "weight_bps": weight_bps,
                    })

        if validators:
            result[chain_id] = validators

    return result


def normalize_weights(validators: list[dict]) -> list[dict]:
    """Adjusts weights so they sum to exactly TARGET_SUM using the largest-remainder method."""
    raw_sum = sum(v["weight_bps"] for v in validators)
    if raw_sum == TARGET_SUM:
        return validators

    diff = TARGET_SUM - raw_sum
    largest = max(validators, key=lambda v: v["weight_bps"])
    largest["weight_bps"] += diff
    return validators


def load_current_weights() -> dict[str, list[dict]]:
    """Returns {chain_id: [{address, name, weight}, ...]} from current_weights.csv."""
    result: dict[str, list[dict]] = {}

    with CURRENT_WEIGHTS_PATH.open() as f:
        reader = csv.DictReader(f)
        for row in reader:
            chain_id = row["chain_id"]
            result.setdefault(chain_id, []).append({
                "address": row["validator_address"],
                "name": row["validator_name"],
                "weight": int(row["current_weight"]),
            })

    return result


def build_all_validators(
    current_by_chain: dict[str, list[dict]],
    target_by_chain: dict[str, list[dict]],
) -> list[dict]:
    """Union of current and target validators with the target weight (0 if not in target)."""
    rows = []
    all_chain_ids = sorted(set(current_by_chain) | set(target_by_chain))

    for chain_id in all_chain_ids:
        target_map = {v["address"]: v for v in target_by_chain.get(chain_id, [])}
        current_map = {v["address"]: v for v in current_by_chain.get(chain_id, [])}
        all_addresses = sorted(set(target_map) | set(current_map))

        for address in all_addresses:
            target_entry = target_map.get(address)
            current_entry = current_map.get(address)
            name = (target_entry or current_entry)["name"]
            target_weight = target_entry["weight_bps"] if target_entry else 0

            rows.append({
                "chain_id": chain_id,
                "validator_name": name,
                "validator_address": address,
                "target_weight": target_weight,
            })

    return rows


def write_target_weights(target_by_chain: dict[str, list[dict]]) -> None:
    rows = []
    for chain_id in sorted(target_by_chain):
        for v in target_by_chain[chain_id]:
            rows.append({
                "chain_id": chain_id,
                "validator_name": v["name"],
                "validator_address": v["address"],
                "target_weight": v["weight_bps"],
            })

    headers = ["chain_id", "validator_name", "validator_address", "target_weight"]
    with TARGET_WEIGHTS_PATH.open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=headers)
        writer.writeheader()
        writer.writerows(rows)


def find_new_validators(
    current_by_chain: dict[str, list[dict]],
    target_by_chain: dict[str, list[dict]],
) -> list[dict]:
    """Validators in the target that don't exist on-chain yet."""
    rows = []
    for chain_id in sorted(target_by_chain):
        current_addresses = {v["address"] for v in current_by_chain.get(chain_id, [])}
        for v in target_by_chain[chain_id]:
            if v["address"] not in current_addresses:
                rows.append({
                    "chain_id": chain_id,
                    "validator_name": v["name"],
                    "validator_address": v["address"],
                })
    return rows


def write_new_validators(rows: list[dict]) -> None:
    headers = ["chain_id", "validator_name", "validator_address"]
    with NEW_VALIDATORS_PATH.open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=headers)
        writer.writeheader()
        writer.writerows(rows)


def write_all_validators(rows: list[dict]) -> None:
    headers = ["chain_id", "validator_name", "validator_address", "target_weight"]
    with ALL_VALIDATORS_PATH.open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=headers)
        writer.writeheader()
        writer.writerows(rows)


def validate(
    target_by_chain: dict[str, list[dict]],
    all_validator_rows: list[dict],
) -> bool:
    ok = True

    # (a) Target weights must sum to exactly 10000 per chain
    for chain_id in sorted(target_by_chain):
        chain_sum = sum(v["weight_bps"] for v in target_by_chain[chain_id])
        if chain_sum != TARGET_SUM:
            print(f"  FAIL: {chain_id} target weights sum to {chain_sum}, expected {TARGET_SUM}")
            ok = False
        else:
            print(f"  OK:   {chain_id} target weights sum to {TARGET_SUM}")

    # (b) Every non-zero entry in all_validators must match target_weights
    target_lookup: dict[tuple[str, str], int] = {}
    for chain_id, validators in target_by_chain.items():
        for v in validators:
            target_lookup[(chain_id, v["address"])] = v["weight_bps"]

    for row in all_validator_rows:
        weight = row["target_weight"]
        if weight == 0:
            continue

        key = (row["chain_id"], row["validator_address"])
        expected = target_lookup.get(key)
        if expected is None:
            print(f"  FAIL: {row['chain_id']} {row['validator_address']} has weight {weight} but is not in target")
            ok = False
        elif expected != weight:
            print(f"  FAIL: {row['chain_id']} {row['validator_address']} weight {weight} != target {expected}")
            ok = False

    return ok


def main() -> int:
    print("Parsing raw target CSVs...")
    target_by_chain = parse_raw_csvs()
    total_target = sum(len(vs) for vs in target_by_chain.values())
    print(f"Found {total_target} validators with target weights across {len(target_by_chain)} chains")

    print("\nNormalizing weights to sum to 10000 per chain...")
    for chain_id in target_by_chain:
        normalize_weights(target_by_chain[chain_id])

    print("Loading current weights...")
    current_by_chain = load_current_weights()

    print("Finding new validators (in target but not on-chain)...")
    new_validator_rows = find_new_validators(current_by_chain, target_by_chain)
    print(f"Found {len(new_validator_rows)} new validators to add")

    print("Building all-validators union...")
    all_validator_rows = build_all_validators(current_by_chain, target_by_chain)
    print(f"Total entries in all_validators: {len(all_validator_rows)}")

    write_target_weights(target_by_chain)
    print(f"\nWrote {TARGET_WEIGHTS_PATH}")

    write_new_validators(new_validator_rows)
    print(f"Wrote {NEW_VALIDATORS_PATH}")

    write_all_validators(all_validator_rows)
    print(f"Wrote {ALL_VALIDATORS_PATH}")

    print("\nValidation:")
    ok = validate(target_by_chain, all_validator_rows)

    if ok:
        print("\nAll validations passed!")
    else:
        print("\nValidation FAILED")
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
