#!/bin/bash
# Prints the transmuter instantiate JSON for one pool (spec §3 "pool_asset_configs").
# usage: instantiate_pool.sh <sttoken-denom-on-osmosis> <native-denom> <rate> <subdenom> <admin>
# The stToken factor is 1e18 and the native factor is rate x 1e18, so out = in x f_native / f_st = in x rate.
# The alloyed (LP) factor equals the native factor: one share is one native base unit.
set -euo pipefail
(( $# == 5 )) || { echo "usage: $0 <sttoken-denom> <native-denom> <rate> <subdenom> <admin>" >&2; exit 2; }

# The rate goes in as an argument, never interpolated into the program text; refuse a rate finer than 18dp
rate_factor=$(python3 - "$3" <<'PY'
import sys
from decimal import Decimal, getcontext

getcontext().prec = 80
scaled = Decimal(sys.argv[1]) * 10**18
if scaled != scaled.to_integral_value() or scaled <= 0:
    sys.exit(f"rate {sys.argv[1]} is not a positive number with at most 18 decimals")
print(int(scaled))
PY
)

jq -n --arg st "$1" --arg nat "$2" --arg factor "$rate_factor" --arg sub "$4" --arg adm "$5" \
 '{pool_asset_configs: [{denom: $st, normalization_factor: "1000000000000000000"},
                        {denom: $nat, normalization_factor: $factor}],
   alloyed_asset_subdenom: $sub, alloyed_asset_normalization_factor: $factor, admin: $adm, moderator: $adm}'
