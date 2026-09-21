#!/usr/bin/env python3
"""
Measure drift between Stride's recorded per-validator delegations (stakeibc host_zone)
and actual on-chain delegations of Stride's delegation ICA on each host chain.

Read-only. Writes drift.json and report.md into this directory.
"""

import json
import time
import urllib.request
import urllib.error
from decimal import Decimal, getcontext

getcontext().prec = 60

SCRIPT_DIR = "/private/tmp/claude-501/-Users-sampocs-Documents-Projects-stride/4ac0fcca-c8bc-4981-b63e-b0f199cc9eaa/scratchpad/drift"

STRIDE_REST = "https://stride-api.polkachu.com"
UA_HEADER = {"User-Agent": "curl/8.0"}

# Stride chain_id -> (chain-registry directory name, decimals for human-readable column)
ZONES = {
    "celestia": {"registry": "celestia", "decimals": 6},
    "comdex-1": {"registry": "comdex", "decimals": 6},
    "dydx-mainnet-1": {"registry": "dydx", "decimals": 18},
    "haqq_11235-1": {"registry": "haqq", "decimals": 18},
    "juno-1": {"registry": "juno", "decimals": 6},
    "laozi-mainnet": {"registry": "bandchain", "decimals": 6},
    "osmosis-1": {"registry": "osmosis", "decimals": 6},
    "phoenix-1": {"registry": "terra2", "decimals": 6},
    "sommelier-3": {"registry": "sommelier", "decimals": 6},
    "ssc-1": {"registry": "saga", "decimals": 6},
}

EXTRA_ENDPOINTS = {
    "celestia": ["https://celestia.rpc.uquad.org:443"],
    "sommelier-3": ["https://rest.cosmos.directory/sommelier"],
}

HTTP_TIMEOUT = 20


def http_get_json(
    url: str, headers: dict | None = None, timeout: int = HTTP_TIMEOUT
) -> dict:
    # Some providers (e.g. polkachu, cosmos.directory) 403 the default
    # python-urllib user agent, so always send a browser/curl-like one
    # unless the caller overrides it.
    merged_headers = {**UA_HEADER, **(headers or {})}
    request = urllib.request.Request(url, headers=merged_headers)
    with urllib.request.urlopen(request, timeout=timeout) as response:
        return json.loads(response.read().decode("utf-8"))


def fetch_stride_host_zones() -> list[dict]:
    data = http_get_json(
        f"{STRIDE_REST}/Stride-Labs/stride/stakeibc/host_zone", headers=UA_HEADER
    )
    return data["host_zone"]


def fetch_registry_rest_endpoints(registry_name: str) -> list[str]:
    url = f"https://raw.githubusercontent.com/cosmos/chain-registry/master/{registry_name}/chain.json"
    data = http_get_json(url)
    endpoints = [
        api["address"].rstrip("/")
        for api in data.get("apis", {}).get("rest", [])
        if api.get("address")
    ]
    return endpoints


def try_endpoints_for(
    endpoints: list[str], path: str, params: str = ""
) -> tuple[str, dict] | None:
    """Try each endpoint until one returns valid JSON for path+params. Return (base_url, json)."""
    for base in endpoints:
        url = f"{base}{path}"
        if params:
            url = f"{url}?{params}"
        try:
            data = http_get_json(url, timeout=15)
            return base, data
        except (
            urllib.error.URLError,
            urllib.error.HTTPError,
            TimeoutError,
            ValueError,
            ConnectionError,
        ):
            continue
    return None


def fetch_all_delegations(base_url: str, delegator_address: str) -> list[dict]:
    """Paginate through /cosmos/staking/v1beta1/delegations/{addr}."""
    results = []
    next_key = None
    while True:
        params = "pagination.limit=500"
        if next_key:
            params += f"&pagination.key={urllib.parse.quote(next_key)}"
        url = f"{base_url}/cosmos/staking/v1beta1/delegations/{delegator_address}?{params}"
        data = http_get_json(url, timeout=20)
        results.extend(data.get("delegation_responses", []))
        next_key = data.get("pagination", {}).get("next_key")
        if not next_key:
            break
    return results


def fetch_validator(base_url: str, validator_address: str) -> dict | None:
    url = f"{base_url}/cosmos/staking/v1beta1/validators/{validator_address}"
    try:
        data = http_get_json(url, timeout=15)
        return data.get("validator")
    except (
        urllib.error.URLError,
        urllib.error.HTTPError,
        TimeoutError,
        ValueError,
        ConnectionError,
    ):
        return None


def find_working_endpoint(
    endpoints: list[str], delegation_ica_address: str
) -> str | None:
    """Find an endpoint that answers the delegations query for this ICA address (even if empty)."""
    for base in endpoints:
        url = f"{base}/cosmos/staking/v1beta1/delegations/{delegation_ica_address}?pagination.limit=1"
        try:
            data = http_get_json(url, timeout=15)
            if "delegation_responses" in data:
                return base
        except (
            urllib.error.URLError,
            urllib.error.HTTPError,
            TimeoutError,
            ValueError,
            ConnectionError,
        ):
            continue
    return None


import urllib.parse  # noqa: E402  (placed after usage above for clarity of grouping)


def build_zone_data(
    stride_chain_id: str, stride_zone: dict, registry_name: str
) -> dict:
    print(f"=== {stride_chain_id} ===")
    delegation_ica_address = stride_zone["delegation_ica_address"]
    host_denom = stride_zone["host_denom"]

    endpoints = list(EXTRA_ENDPOINTS.get(stride_chain_id, []))
    try:
        endpoints += fetch_registry_rest_endpoints(registry_name)
    except Exception as exc:
        print(f"  WARN: could not fetch chain-registry endpoints: {exc}")

    working_base = find_working_endpoint(endpoints, delegation_ica_address)
    if working_base is None:
        print(
            f"  FAILED: no working REST endpoint found among {len(endpoints)} candidates"
        )
        return {
            "chain_id": stride_chain_id,
            "host_denom": host_denom,
            "delegation_ica_address": delegation_ica_address,
            "rest_endpoint": None,
            "error": "no working host REST endpoint found",
            "stride_validators": stride_zone.get("validators", []),
            "host_delegations": [],
            "validators_info": {},
        }

    print(f"  using REST endpoint: {working_base}")

    try:
        host_delegations = fetch_all_delegations(working_base, delegation_ica_address)
    except Exception as exc:
        print(f"  FAILED fetching delegations from {working_base}: {exc}")
        return {
            "chain_id": stride_chain_id,
            "host_denom": host_denom,
            "delegation_ica_address": delegation_ica_address,
            "rest_endpoint": working_base,
            "error": f"delegations query failed: {exc}",
            "stride_validators": stride_zone.get("validators", []),
            "host_delegations": [],
            "validators_info": {},
        }

    print(f"  host delegations: {len(host_delegations)}")

    stride_validator_addrs = {v["address"] for v in stride_zone.get("validators", [])}
    host_validator_addrs = {
        d["delegation"]["validator_address"] for d in host_delegations
    }
    all_validator_addrs = stride_validator_addrs | host_validator_addrs

    validators_info = {}
    for i, val_addr in enumerate(sorted(all_validator_addrs)):
        info = fetch_validator(working_base, val_addr)
        if info:
            validators_info[val_addr] = {
                "tokens": info.get("tokens"),
                "delegator_shares": info.get("delegator_shares"),
                "moniker": info.get("description", {}).get("moniker"),
                "jailed": info.get("jailed"),
                "status": info.get("status"),
            }
        if (i + 1) % 20 == 0:
            print(f"    fetched {i + 1}/{len(all_validator_addrs)} validator infos")

    return {
        "chain_id": stride_chain_id,
        "host_denom": host_denom,
        "delegation_ica_address": delegation_ica_address,
        "rest_endpoint": working_base,
        "error": None,
        "stride_validators": stride_zone.get("validators", []),
        "host_delegations": host_delegations,
        "validators_info": validators_info,
    }


def compute_zone_rows(zone_data: dict) -> list[dict]:
    """Join Stride and host data per validator address."""
    stride_by_addr = {v["address"]: v for v in zone_data["stride_validators"]}
    host_by_addr = {}
    for entry in zone_data["host_delegations"]:
        deleg = entry["delegation"]
        host_by_addr[deleg["validator_address"]] = {
            "shares": deleg["shares"],
            "balance_amount": entry.get("balance", {}).get("amount"),
        }

    all_addrs = set(stride_by_addr) | set(host_by_addr)
    validators_info = zone_data["validators_info"]

    rows = []
    for addr in sorted(all_addrs):
        stride_v = stride_by_addr.get(addr)
        host_v = host_by_addr.get(addr)
        info = validators_info.get(addr, {})

        recorded = int(stride_v["delegation"]) if stride_v else 0
        actual = (
            int(host_v["balance_amount"])
            if host_v and host_v["balance_amount"] is not None
            else 0
        )
        diff = recorded - actual

        stride_rate = Decimal(stride_v["shares_to_tokens_rate"]) if stride_v else None

        chain_rate = None
        if info.get("tokens") is not None and info.get("delegator_shares"):
            shares_dec = Decimal(info["delegator_shares"])
            if shares_dec != 0:
                chain_rate = Decimal(info["tokens"]) / shares_dec

        rate_diff_pct = None
        if stride_rate is not None and chain_rate is not None and chain_rate != 0:
            rate_diff_pct = float((stride_rate - chain_rate) / chain_rate * 100)

        host_shares = Decimal(host_v["shares"]) if host_v else None

        calibrated = None
        if host_shares is not None and stride_rate is not None:
            calibrated = int(
                (host_shares * stride_rate).to_integral_value(rounding="ROUND_FLOOR")
            )

        calibrated_exact = None
        if host_shares is not None and chain_rate is not None:
            calibrated_exact = int(
                (host_shares * chain_rate).to_integral_value(rounding="ROUND_FLOOR")
            )

        rows.append(
            {
                "validator_address": addr,
                "moniker": info.get("moniker"),
                "in_stride_list": stride_v is not None,
                "in_host_delegations": host_v is not None,
                "recorded": recorded,
                "actual": actual,
                "diff": diff,
                "over_recorded": diff > 0,
                "under_recorded": diff < 0,
                "stride_rate": str(stride_rate) if stride_rate is not None else None,
                "chain_rate": str(chain_rate) if chain_rate is not None else None,
                "rate_diff_pct": rate_diff_pct,
                "host_shares": str(host_shares) if host_shares is not None else None,
                "calibrated": calibrated,
                "calibrated_exact": calibrated_exact,
                "calibrated_vs_actual_diff": (calibrated - actual)
                if calibrated is not None
                else None,
                "calibrated_exact_vs_actual_diff": (calibrated_exact - actual)
                if calibrated_exact is not None
                else None,
                "weight": stride_v.get("weight") if stride_v else None,
                "delegation_changes_in_progress": int(
                    stride_v.get("delegation_changes_in_progress", 0)
                )
                if stride_v
                else 0,
                "slash_query_in_progress": bool(stride_v.get("slash_query_in_progress"))
                if stride_v
                else False,
            }
        )
    return rows


def summarize_zone(chain_id: str, zone_data: dict, rows: list[dict]) -> dict:
    if zone_data.get("error"):
        return {
            "chain_id": chain_id,
            "error": zone_data["error"],
        }

    num_stride_validators = len(zone_data["stride_validators"])
    num_with_onchain = sum(1 for r in rows if r["actual"] > 0)
    sum_recorded = sum(r["recorded"] for r in rows)
    sum_actual = sum(r["actual"] for r in rows)
    sum_diff = sum_recorded - sum_actual

    over = [r for r in rows if r["over_recorded"]]
    under = [r for r in rows if r["under_recorded"]]

    max_over = max(over, key=lambda r: r["diff"], default=None)
    max_under = min(under, key=lambda r: r["diff"], default=None)

    missing_from_stride = [
        r["validator_address"]
        for r in rows
        if r["in_host_delegations"] and not r["in_stride_list"]
    ]
    zero_onchain_in_stride = [
        r["validator_address"]
        for r in rows
        if r["in_stride_list"] and not r["in_host_delegations"] and r["recorded"] != 0
    ]

    rate_diffs = [
        abs(r["rate_diff_pct"]) for r in rows if r["rate_diff_pct"] is not None
    ]
    max_rate_diff_pct = max(rate_diffs) if rate_diffs else None

    in_progress_count = sum(
        1
        for r in rows
        if r["delegation_changes_in_progress"] > 0 or r["slash_query_in_progress"]
    )

    calibrated_mismatches = [
        r for r in rows if r["calibrated_vs_actual_diff"] not in (None, 0)
    ]
    calibrated_exact_mismatches = [
        r for r in rows if r["calibrated_exact_vs_actual_diff"] not in (None, 0)
    ]
    max_calibrated_gap = max(
        (abs(r["calibrated_vs_actual_diff"]) for r in calibrated_mismatches), default=0
    )
    max_calibrated_exact_gap = max(
        (
            abs(r["calibrated_exact_vs_actual_diff"])
            for r in calibrated_exact_mismatches
        ),
        default=0,
    )

    return {
        "chain_id": chain_id,
        "error": None,
        "num_stride_validators": num_stride_validators,
        "num_with_onchain": num_with_onchain,
        "sum_recorded": sum_recorded,
        "sum_actual": sum_actual,
        "sum_diff": sum_diff,
        "count_over_recorded": len(over),
        "max_over_recorded_amount": max_over["diff"] if max_over else 0,
        "max_over_recorded_addr": max_over["validator_address"] if max_over else None,
        "max_over_recorded_pct": (max_over["diff"] / max_over["recorded"] * 100)
        if max_over and max_over["recorded"]
        else None,
        "count_under_recorded": len(under),
        "max_under_recorded_amount": max_under["diff"] if max_under else 0,
        "max_under_recorded_addr": max_under["validator_address"]
        if max_under
        else None,
        "missing_from_stride": missing_from_stride,
        "zero_onchain_in_stride": zero_onchain_in_stride,
        "max_rate_diff_pct": max_rate_diff_pct,
        "in_progress_count": in_progress_count,
        "calibrated_mismatch_count": len(calibrated_mismatches),
        "calibrated_exact_mismatch_count": len(calibrated_exact_mismatches),
        "max_calibrated_gap": max_calibrated_gap,
        "max_calibrated_exact_gap": max_calibrated_exact_gap,
    }


def human(amount: int, decimals: int) -> str:
    return f"{amount / (10**decimals):,.6f}"


def render_zone_table(
    chain_id: str, zone_data: dict, rows: list[dict], decimals: int
) -> str:
    if zone_data.get("error"):
        return f"## {chain_id}\n\n**FAILED**: {zone_data['error']}\n"

    nonzero_rows = [r for r in rows if r["diff"] != 0]
    nonzero_rows.sort(key=lambda r: -abs(r["diff"]))

    lines = [f"## {chain_id}", ""]
    lines.append(f"REST endpoint used: `{zone_data['rest_endpoint']}`")
    lines.append("")
    lines.append(
        "| Validator | Moniker | Recorded (raw) | Actual (raw) | Diff (raw) | Recorded | Actual | Diff | Stride Rate | Chain Rate | Rate Diff % | Over/Under | In-Progress |"
    )
    lines.append("|---|---|---|---|---|---|---|---|---|---|---|---|---|")

    sum_recorded = sum(r["recorded"] for r in rows)
    sum_actual = sum(r["actual"] for r in rows)

    for r in nonzero_rows:
        flag = "OVER" if r["over_recorded"] else "under"
        in_progress = (
            "yes"
            if (r["delegation_changes_in_progress"] > 0 or r["slash_query_in_progress"])
            else ""
        )
        rate_diff = (
            f"{r['rate_diff_pct']:.4f}%" if r["rate_diff_pct"] is not None else "n/a"
        )
        lines.append(
            f"| `{r['validator_address']}` | {r['moniker'] or ''} | {r['recorded']} | {r['actual']} | {r['diff']} | "
            f"{human(r['recorded'], decimals)} | {human(r['actual'], decimals)} | {human(r['diff'], decimals)} | "
            f"{r['stride_rate'] or ''} | {r['chain_rate'] or ''} | {rate_diff} | {flag} | {in_progress} |"
        )

    lines.append(
        f"| **TOTAL** | | {sum_recorded} | {sum_actual} | {sum_recorded - sum_actual} | "
        f"{human(sum_recorded, decimals)} | {human(sum_actual, decimals)} | {human(sum_recorded - sum_actual, decimals)} | | | | | |"
    )
    lines.append("")
    return "\n".join(lines)


def render_summary_table(summaries: list[dict], decimals_map: dict) -> str:
    lines = [
        "| Zone | #Val (Stride) | #Val on-chain | Sum Recorded | Sum Actual | Sum Diff | #Over | Max Over (amt / %) | #Under | Max Under | Missing from Stride | Zero on-chain in Stride | Max Rate Diff % | #In-Progress |",
        "|---|---|---|---|---|---|---|---|---|---|---|---|---|---|",
    ]
    for s in summaries:
        chain_id = s["chain_id"]
        decimals = decimals_map.get(chain_id, 6)
        if s.get("error"):
            lines.append(
                f"| {chain_id} | FAILED: {s['error']} | | | | | | | | | | | | |"
            )
            continue
        max_over_pct = (
            f"{s['max_over_recorded_pct']:.2f}%"
            if s["max_over_recorded_pct"] is not None
            else "n/a"
        )
        max_rate = (
            f"{s['max_rate_diff_pct']:.4f}%"
            if s["max_rate_diff_pct"] is not None
            else "n/a"
        )
        lines.append(
            f"| {chain_id} | {s['num_stride_validators']} | {s['num_with_onchain']} | "
            f"{human(s['sum_recorded'], decimals)} | {human(s['sum_actual'], decimals)} | {human(s['sum_diff'], decimals)} | "
            f"{s['count_over_recorded']} | {human(s['max_over_recorded_amount'], decimals)} / {max_over_pct} | "
            f"{s['count_under_recorded']} | {human(s['max_under_recorded_amount'], decimals)} | "
            f"{len(s['missing_from_stride'])} | {len(s['zero_onchain_in_stride'])} | {max_rate} | {s['in_progress_count']} |"
        )
    return "\n".join(lines)


def main():
    stride_zones = fetch_stride_host_zones()
    stride_by_chain_id = {z["chain_id"]: z for z in stride_zones}

    all_zone_data = {}
    all_rows = {}
    all_summaries = []

    for chain_id, cfg in ZONES.items():
        stride_zone = stride_by_chain_id.get(chain_id)
        if stride_zone is None:
            print(f"WARN: {chain_id} not found in Stride host_zone list")
            continue

        zone_data = build_zone_data(chain_id, stride_zone, cfg["registry"])
        all_zone_data[chain_id] = zone_data

        rows = compute_zone_rows(zone_data) if not zone_data.get("error") else []
        all_rows[chain_id] = rows

        summary = summarize_zone(chain_id, zone_data, rows)
        all_summaries.append(summary)

    # Write drift.json
    output = {
        "generated_at": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()),
        "zones": {},
    }
    for chain_id in ZONES:
        if chain_id not in all_zone_data:
            continue
        output["zones"][chain_id] = {
            "rest_endpoint": all_zone_data[chain_id].get("rest_endpoint"),
            "error": all_zone_data[chain_id].get("error"),
            "delegation_ica_address": all_zone_data[chain_id].get(
                "delegation_ica_address"
            ),
            "host_denom": all_zone_data[chain_id].get("host_denom"),
            "summary": next(s for s in all_summaries if s["chain_id"] == chain_id),
            "validators": all_rows.get(chain_id, []),
        }

    with open(f"{SCRIPT_DIR}/drift.json", "w") as f:
        json.dump(output, f, indent=2)

    # Write report.md
    decimals_map = {chain_id: cfg["decimals"] for chain_id, cfg in ZONES.items()}
    report_lines = ["# Stride Delegation Drift Report", ""]
    report_lines.append(f"Generated: {output['generated_at']}")
    report_lines.append("")
    report_lines.append("## Summary")
    report_lines.append("")
    report_lines.append(render_summary_table(all_summaries, decimals_map))
    report_lines.append("")
    report_lines.append("## Per-Zone Detail (validators with nonzero diff)")
    report_lines.append("")
    for chain_id in ZONES:
        if chain_id not in all_zone_data:
            continue
        report_lines.append(
            render_zone_table(
                chain_id,
                all_zone_data[chain_id],
                all_rows.get(chain_id, []),
                decimals_map[chain_id],
            )
        )

    with open(f"{SCRIPT_DIR}/report.md", "w") as f:
        f.write("\n".join(report_lines))

    print("\nDone. Wrote drift.json and report.md")


if __name__ == "__main__":
    main()
