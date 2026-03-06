"""
Converts CSV files in /raw to JSON messages in /update-msgs and /add-msgs.

Each CSV has columns:
  Validator Name, Validator Address, ..., Current Weight, New Weight

All validators in the CSV are included in update-msgs.
Validators without a "New Weight" are explicitly set to weight 0.
Percentage values (e.g. "8.00%") are converted to integer basis points
so that relative ratios are preserved (e.g. 8.00% -> 800, 7.14% -> 714).

Validators with a "New Weight" that don't exist on-chain are written to add-msgs.

update-msgs format (strided tx stakeibc change-validator-weights):
  {
    "validator_weights": [
      {"address": "cosmosXXX", "weight": 800}
    ]
  }

add-msgs format (strided tx stakeibc add-validators):
  {
    "validators": [
      {"name": "Val1", "address": "cosmosXXX", "weight": 800}
    ]
  }
"""

import csv
import json
import re
import urllib.request
from pathlib import Path

RAW_DIR = Path(__file__).parent / "raw"
UPDATE_MSGS_DIR = Path(__file__).parent / "update-msgs"
ADD_MSGS_DIR = Path(__file__).parent / "add-msgs"
HOST_ZONE_API = "https://stride-api.polkachu.com/Stride-Labs/stride/stakeibc/host_zone"


def parse_weight_pct(value: str) -> int:
    """Convert a percentage string like '8.00%' to integer basis points (800)."""
    cleaned = value.strip().replace("%", "")
    return round(float(cleaned) * 100)


def extract_chain_id(filename: str) -> str:
    """Extract chain-id from filename like 'Stride Delegations Q1 2026 - cosmoshub-4.csv'."""
    match = re.search(r" - (.+)\.csv$", filename)
    if not match:
        raise ValueError(f"Could not extract chain-id from filename: {filename}")
    return match.group(1)


def parse_csv(csv_path: Path) -> list[dict]:
    """Read a CSV and return a list of {name, address, weight} dicts for all rows."""
    rows = []
    with open(csv_path, newline="", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            new_weight = row.get("New Weight", "").strip()
            rows.append({
                "name": row["Validator Name"].strip(),
                "address": row["Validator Address"].strip(),
                "weight": parse_weight_pct(new_weight) if new_weight else 0,
            })
    return rows


def fetch_onchain_validators() -> dict[str, dict[str, int]]:
    """Fetch current on-chain validators per host zone. Returns {chain_id: {address: weight}}."""
    print("Fetching on-chain validators...")
    req = urllib.request.Request(HOST_ZONE_API, headers={"User-Agent": "stride-validator-weights"})
    with urllib.request.urlopen(req) as resp:
        data = json.load(resp)

    return {
        hz["chain_id"]: {v["address"]: int(v["weight"]) for v in hz.get("validators", [])}
        for hz in data["host_zone"]
    }


def write_json(path: Path, data: dict) -> None:
    with open(path, "w") as f:
        json.dump(data, f, indent=2)
        f.write("\n")


def main() -> None:
    # Clean output dirs to avoid stale files from previous runs
    for d in (UPDATE_MSGS_DIR, ADD_MSGS_DIR):
        if d.exists():
            for f in d.glob("*.json"):
                f.unlink()
        d.mkdir(exist_ok=True)

    csv_files = sorted(RAW_DIR.glob("*.csv"))
    if not csv_files:
        print("No CSV files found in raw/")
        return

    onchain = fetch_onchain_validators()
    covered_chains: set[str] = set()

    for csv_path in csv_files:
        chain_id = extract_chain_id(csv_path.name)
        covered_chains.add(chain_id)
        rows = parse_csv(csv_path)

        if not rows:
            print(f"  {chain_id}: no validators, skipping")
            continue

        onchain_vals = onchain.get(chain_id, {})
        csv_addrs = {r["address"] for r in rows}

        # Include on-chain validators missing from CSV only if they have non-zero weight
        onchain_only = sorted(addr for addr in onchain_vals if addr not in csv_addrs and onchain_vals[addr] > 0)
        for addr in onchain_only:
            rows.append({"name": "", "address": addr, "weight": 0})

        # Add not-on-chain validators only if they have a non-zero weight
        add_entries = [r for r in rows if r["address"] not in onchain_vals and r["weight"] > 0]

        if add_entries:
            add_msg = {"validators": [
                {"name": r["name"], "address": r["address"], "weight": 0}
                for r in add_entries
            ]}
            write_json(ADD_MSGS_DIR / f"{chain_id}.json", add_msg)

        # Write update-msgs — exclude not-on-chain validators with 0 weight (no-ops)
        update_rows = [r for r in rows if r["address"] in onchain_vals or r["weight"] > 0]
        update_msg = {"validator_weights": [
            {"address": r["address"], "weight": r["weight"]} for r in update_rows
        ]}
        write_json(UPDATE_MSGS_DIR / f"{chain_id}.json", update_msg)

        msg = f"  {chain_id}: {len(update_rows)} update"
        if add_entries:
            msg += f", {len(add_entries)} add"
        if onchain_only:
            msg += f" ({len(onchain_only)} on-chain not in CSV)"
        print(msg)

    # Warn about on-chain zones with no CSV
    for chain_id in sorted(onchain.keys() - covered_chains):
        print(f"  {chain_id}: WARNING - on-chain host zone has no CSV")


if __name__ == "__main__":
    main()
