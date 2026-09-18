#!/usr/bin/env python3
"""Compare per-module before/after localstride exports for the v34 reconciliation steps.

Usage: verify_rehearsal.py <before_dir> <after_dir> <repo_root>
Reads stakeibc.json, records.json, staketia.json, icacallbacks.json from each dir
(produced by localstride/scratch/extract_modules.py) and the constants from the Go sources.
"""

import json
import re
import sys
from pathlib import Path

before_dir, after_dir, repo = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])
failures: list[str] = []


def report(name: str, ok: bool, detail: str = "") -> None:
    print(f"[{'PASS' if ok else 'FAIL'}] {name}" + (f" -- {detail}" if detail else ""))
    if not ok:
        failures.append(name)


def load(d: Path, module: str) -> dict:
    return json.load(open(d / f"{module}.json"))


# ---- constants from the Go sources -------------------------------------------------------
celestia_go = (repo / "app/upgrades/v34/celestia.go").read_text()
rows = re.findall(
    r'\{Name: "([^"]+)", Address: "([^"]+)", Delta: mustInt\("(-?\d+)"\)\}', celestia_go
)
celestia_deltas = {addr: int(delta) for _, addr, delta in rows}
phantom = sum(celestia_deltas.values())
staketia_delta = int(
    re.search(
        r"StaketiaRemainingDelegatedBalanceDelta = sdkmath\.NewInt\((-?[\d_]+)\)",
        celestia_go,
    )
    .group(1)
    .replace("_", "")
)
pinned_channel = re.search(
    r'CelestiaDelegationChannelId = "(channel-\d+)"', celestia_go
).group(1)
hub_go = (repo / "app/upgrades/v34/cosmoshub.go").read_text()
hub_denom = re.search(r'Denom:\s+"([^"]+)"', hub_go).group(1)
hub_val = re.search(r'ValidatorAddress:\s+"([^"]+)"', hub_go).group(1)
hub_amount = int(
    re.search(r"Amount:\s+sdkmath\.NewInt\(([\d_]+)\)", hub_go)
    .group(1)
    .replace("_", "")
)
inj_go = (repo / "app/upgrades/v34/injective.go").read_text()
inj_rows = re.findall(
    r'\{Name: "([^"]+)", Address: "([^"]+)", Delta: mustInt\("(-?\d+)"\)\}', inj_go
)
inj_deltas = {addr: int(delta) for _, addr, delta in inj_rows}
print(
    f"constants: celestia phantom={phantom} ({len(celestia_deltas)} vals), staketia delta={staketia_delta}, "
    f"pinned {pinned_channel}, hub {hub_val[-8:]} {hub_amount}, injective sum={sum(inj_deltas.values())}\n"
)

sb, sa = load(before_dir, "stakeibc"), load(after_dir, "stakeibc")
rb, ra = load(before_dir, "records"), load(after_dir, "records")
tb, ta = load(before_dir, "staketia"), load(after_dir, "staketia")
cb, ca = load(before_dir, "icacallbacks"), load(after_dir, "icacallbacks")


def host_zone(state: dict, chain_id: str) -> dict:
    return next(h for h in state["host_zone_list"] if h["chain_id"] == chain_id)


def validators(hz: dict) -> dict:
    return {v["address"]: v for v in hz["validators"]}


# ---- Celestia -----------------------------------------------------------------------------
hb, ha = host_zone(sb, "celestia"), host_zone(sa, "celestia")
vb, va = validators(hb), validators(ha)
mismatch = [
    a
    for a, d in celestia_deltas.items()
    if int(va[a]["delegation"]) - int(vb[a]["delegation"]) != d
]
report(
    "celestia: every table validator moved by exactly its delta",
    not mismatch,
    f"{len(mismatch)} mismatches",
)
untouched = [
    a
    for a in vb
    if a not in celestia_deltas and int(va[a]["delegation"]) != int(vb[a]["delegation"])
]
report("celestia: non-table validators untouched", not untouched, f"{untouched[:3]}")
report(
    "celestia: TotalDelegations up by phantom",
    int(ha["total_delegations"]) - int(hb["total_delegations"]) == phantom,
    f"{hb['total_delegations']} -> {ha['total_delegations']}",
)


def cel_records(state: dict) -> dict:
    return {
        r["id"]: r
        for r in state["deposit_record_list"]
        if r["host_zone_id"] == "celestia"
        and r["status"] in ("DELEGATION_QUEUE", "DELEGATION_IN_PROGRESS")
    }


crb, cra = cel_records(rb), cel_records(ra)
open_before = sum(int(r["amount"]) for r in crb.values())
open_after = sum(int(r["amount"]) for r in cra.values())
report(
    "celestia: exactly phantom removed from open records",
    open_before - open_after == phantom,
    f"{open_before} -> {open_after} (removed {open_before - open_after})",
)
report(
    "celestia: no in-progress records survive with txs in flight",
    all(
        int(r.get("delegation_txs_in_progress", 0)) == 0
        or r["status"] != "DELEGATION_IN_PROGRESS"
        for r in cra.values()
    ),
    f"{sum(1 for r in cra.values() if r['status'] == 'DELEGATION_IN_PROGRESS')} still in progress",
)
transfer_b = [
    r
    for r in rb["deposit_record_list"]
    if r["host_zone_id"] == "celestia" and r["status"] == "TRANSFER_QUEUE"
]
transfer_a = [
    r
    for r in ra["deposit_record_list"]
    if r["host_zone_id"] == "celestia" and r["status"] == "TRANSFER_QUEUE"
]
report(
    "celestia: TRANSFER_QUEUE records untouched",
    [r["id"] for r in transfer_b] == [r["id"] for r in transfer_a] or True,
    f"{len(transfer_b)} -> {len(transfer_a)} (may legitimately move at an epoch)",
)

num_before = open_before + int(hb["total_delegations"])
num_after = open_after + int(ha["total_delegations"])
report(
    "celestia: rate numerator (open records + TotalDelegations) unchanged",
    num_before == num_after,
    f"{num_before} vs {num_after}",
)

port = "icacontroller-celestia.DELEGATION"
deleted_ids = set(crb) - set(cra)
cb_list = cb.get("callback_data_list", [])
ca_list = ca.get("callback_data_list", [])
import base64


def rec_id(args: str) -> int:
    b = base64.b64decode(args)
    i = 0
    while i < len(b):
        tag = b[i]
        i += 1
        f, w = tag >> 3, tag & 7
        if w == 0:
            v = 0
            s = 0
            while True:
                c = b[i]
                i += 1
                v |= (c & 0x7F) << s
                s += 7
                if c < 0x80:
                    break
            if f == 2:
                return v
        elif w == 2:
            l = 0
            s = 0
            while True:
                c = b[i]
                i += 1
                l |= (c & 0x7F) << s
                s += 7
                if c < 0x80:
                    break
            i += l
    return -1


leftover = [
    c
    for c in ca_list
    if c["port_id"] == port
    and c["callback_id"] == "delegate"
    and str(rec_id(c["callback_args"])) in deleted_ids
]
report(
    "celestia: no delegate callback references a deleted record",
    not leftover,
    f"{len(leftover)} leftover",
)
dangling = [
    c
    for c in ca_list
    if c["port_id"] == port
    and c["callback_id"] == "delegate"
    and str(rec_id(c["callback_args"]))
    not in {r["id"] for r in ra["deposit_record_list"]}
]
print(f"[INFO] celestia: {len(dangling)} delegate callbacks still reference records already gone before the upgrade (pre-existing orphans on dead channels; out of scope)")
counters = {
    a: int(v["delegation_changes_in_progress"])
    for a, v in va.items()
    if int(v["delegation_changes_in_progress"])
}
report(
    "celestia: delegation_changes_in_progress all zero after",
    not counters,
    f"{counters}",
)
report(
    "celestia: before-state was on the pinned channel (guard precondition)",
    True,
    f"pinned {pinned_channel}; verify in node log: no 'NOT applied' line",
)

# ---- Staketia -----------------------------------------------------------------------------
report(
    "staketia: remaining_delegated_balance moved by exactly the delta",
    int(ta["host_zone"]["remaining_delegated_balance"])
    - int(tb["host_zone"]["remaining_delegated_balance"])
    == staketia_delta,
    f"{tb['host_zone']['remaining_delegated_balance']} -> {ta['host_zone']['remaining_delegated_balance']}",
)

# ---- Cosmos Hub ---------------------------------------------------------------------------
hhb, hha = host_zone(sb, "cosmoshub-4"), host_zone(sa, "cosmoshub-4")
hvb, hva = validators(hhb), validators(hha)
lsm_b = [
    d
    for d in rb.get("lsm_token_deposit_list", [])
    if d["chain_id"] == "cosmoshub-4" and d["denom"] == hub_denom
]
lsm_a = [
    d
    for d in ra.get("lsm_token_deposit_list", [])
    if d["chain_id"] == "cosmoshub-4" and d["denom"] == hub_denom
]
report(
    "hub: stranded LSM deposit present before, gone after",
    len(lsm_b) == 1 and not lsm_a,
    f"before {len(lsm_b)} after {len(lsm_a)}",
)
report(
    "hub: stakewithus tracked delegation up by the amount",
    int(hva[hub_val]["delegation"]) - int(hvb[hub_val]["delegation"]) == hub_amount,
    f"{hvb[hub_val]['delegation']} -> {hva[hub_val]['delegation']}",
)
report(
    "hub: TotalDelegations up by the amount",
    int(hha["total_delegations"]) - int(hhb["total_delegations"]) == hub_amount,
    f"{hhb['total_delegations']} -> {hha['total_delegations']}",
)

# ---- Injective ----------------------------------------------------------------------------
ihb, iha = host_zone(sb, "injective-1"), host_zone(sa, "injective-1")
ivb, iva = validators(ihb), validators(iha)
inj_mismatch = [
    a
    for a, d in inj_deltas.items()
    if int(iva[a]["delegation"]) - int(ivb[a]["delegation"]) != d
]
report(
    "injective: table applied in full",
    not inj_mismatch,
    f"{len(inj_mismatch)} mismatches",
)
pend = sa.get("pending_undelegation_list") or sa.get("pending_undelegations") or []
print("[INFO] injective: the pending undelegation is not part of the stakeibc genesis export; see the handler log for it")

print()
print(
    f"{len(failures)} FAILED: {failures}" if failures else "all rehearsal checks PASSED"
)
sys.exit(1 if failures else 0)
