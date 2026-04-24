"""
Queries the Stride host_zone endpoint and writes a CSV of every
validator's current on-chain weight across all host zones.

Output columns: chain_id, validator_name, validator_address, current_weight
"""

import csv
import sys
import urllib.request
import json
from pathlib import Path

HOST_ZONE_URL = "https://stride-api.polkachu.com/Stride-Labs/stride/stakeibc/host_zone"
OUTPUT_FILENAME = "current_weights.csv"
CSV_HEADERS = ["chain_id", "validator_name", "validator_address", "current_weight"]


def fetch_host_zones() -> list[dict]:
    request = urllib.request.Request(HOST_ZONE_URL, headers={"User-Agent": "stride-v32-weights/1.0"})
    with urllib.request.urlopen(request, timeout=30) as response:
        payload = json.load(response)
    return payload["host_zone"]


def build_rows(host_zones: list[dict]) -> list[dict]:
    rows = []
    for host_zone in host_zones:
        chain_id = host_zone["chain_id"]
        for validator in host_zone.get("validators", []):
            rows.append({
                "chain_id": chain_id,
                "validator_name": validator.get("name", ""),
                "validator_address": validator["address"],
                "current_weight": validator["weight"],
            })
    return rows


def write_csv(rows: list[dict], output_path: Path) -> None:
    with output_path.open("w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=CSV_HEADERS)
        writer.writeheader()
        writer.writerows(rows)


def main() -> int:
    output_path = Path(__file__).parent.parent / OUTPUT_FILENAME

    print(f"Fetching host zones from {HOST_ZONE_URL} ...")
    host_zones = fetch_host_zones()
    print(f"Found {len(host_zones)} host zones")

    rows = build_rows(host_zones)
    print(f"Collected {len(rows)} validator entries")

    write_csv(rows, output_path)
    print(f"Wrote {output_path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
