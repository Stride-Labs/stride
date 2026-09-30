#!/usr/bin/env python3
"""
Turn measure_delegation_drift.py's drift.json into the Go delta table an upgrade handler applies
through reconcileHostZoneDelegations (app/upgrades/v35/delegation_deltas.go).

drift.json stores diff = recorded - actual; the Go table stores Delta = actual - recorded, so
every sign flips here. Zero diffs and validators Stride does not track are left out.

Usage:
    python3 gen_delta_table.py DRIFT_JSON CHAIN_ID [--host-zone-json HOST_ZONE_JSON] [--var-name NAME]

HOST_ZONE_JSON is the REST response of /Stride-Labs/stride/stakeibc/host_zone/CHAIN_ID; it maps
validator addresses to Stride's validator names for the Name column. Without it the host
moniker (lower-cased, alphanumerics only) is used.
"""

import argparse
import json
import pathlib
import re
from dataclasses import dataclass


@dataclass(frozen=True)
class DeltaEntry:
    name: str
    address: str
    delta: int  # actual on-chain minus tracked, in the host base denom


def load_deltas(drift_path: pathlib.Path, chain_id: str, names: dict[str, str]) -> list[DeltaEntry]:
    drift = json.loads(drift_path.read_text())
    rows = drift["zones"][chain_id]["validators"]

    entries = [
        DeltaEntry(
            name=names.get(row["validator_address"], sanitize_name(row.get("moniker") or "")),
            address=row["validator_address"],
            delta=-int(row["diff"]),
        )
        for row in rows
        if row["in_stride_list"] and int(row["diff"]) != 0
    ]
    return sorted(entries, key=lambda entry: -abs(entry.delta))


def load_host_zone_names(host_zone_path: pathlib.Path | None) -> dict[str, str]:
    if host_zone_path is None:
        return {}
    host_zone = json.loads(host_zone_path.read_text())["host_zone"]
    return {validator["address"]: validator["name"] for validator in host_zone["validators"]}


def sanitize_name(moniker: str) -> str:
    return re.sub(r"[^a-z0-9]", "", moniker.lower())


def render_go(entries: list[DeltaEntry], var_name: str) -> str:
    lines = [f"var {var_name} = []DelegationDelta{{"]
    lines += [f'\t{{Name: "{entry.name}", Address: "{entry.address}", Delta: mustInt("{entry.delta}")}},' for entry in entries]
    lines.append("}")
    return "\n".join(lines) + "\n"


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    parser.add_argument("drift_json", type=pathlib.Path)
    parser.add_argument("chain_id")
    parser.add_argument("--host-zone-json", type=pathlib.Path, default=None)
    parser.add_argument("--var-name", default="HaqqDelegationDeltas")
    args = parser.parse_args()

    names = load_host_zone_names(host_zone_path=args.host_zone_json)
    entries = load_deltas(drift_path=args.drift_json, chain_id=args.chain_id, names=names)
    net = sum(entry.delta for entry in entries)

    print(f"// {len(entries)} deltas, net {net} (actual minus tracked)")
    print(render_go(entries=entries, var_name=args.var_name), end="")


if __name__ == "__main__":
    main()
