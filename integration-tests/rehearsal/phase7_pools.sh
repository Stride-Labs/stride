#!/bin/bash
# Phase 7: transmuter pools on Osmosis. Create three pools, run the coverage check before and after
# funding (test deposit, corrupted-asset marking, outsider join), exercise swaps, then the pool gate.
source "$(dirname "$0")/lib.sh"
source "$REHEARSAL_DIR/state.env"
log "## Phase 7: pools"

POOLS_FILE=$REHEARSAL_DIR/pools.json
GATE=$REPO/scripts/wind-down/check_transmuter_pool.py
UOSMO_GAS_RESERVE=5000000   # the vault pays gas in uosmo for every later multisig tx, so it is never swept into a pool

ibc_denom() { printf 'ibc/%s' "$(printf '%s' "$1" | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)"; }
STATOM_CANON=$(ibc_denom transfer/channel-0/stuatom)
STOSMO_CANON=$(ibc_denom transfer/channel-0/stuosmo)
STATOM_HUBROUTE=$(ibc_denom transfer/channel-1/transfer/channel-0/stuatom)
ATOM_ON_OSMO=$(ibc_denom transfer/channel-1/uatom)
HUBROUTE_TRACE=transfer/channel-1/transfer/channel-0/stuatom

# ---------------------------------------------------------------- helpers

# Run a single-key tx, log it, and require it to land with code 0 (broadcast mode sync only reports CheckTx)
user_tx() { # label cmd...
  local label=$1 out hash; shift
  out=$(log_cmd "$label" "$@")
  hash=$(printf '%s\n' "$out" | grep '^{' | tail -1 | tx_hash)
  [[ -n "$hash" && "$hash" != null ]] || { log "no txhash for: $label"; return 1; }
  wait_tx osmosisd "$hash"
}

# A refused tx must fail for the expected reason, not for an unrelated error (gas, sequence, typo)
REFUSAL_REASON_RE='Corrupted asset|must not increase'

assert_refusal_reason() { # captured-output
  if grep -qE "$REFUSAL_REASON_RE" <<<"$1"; then
    log "refused for the expected reason: $(grep -oE "$REFUSAL_REASON_RE" <<<"$1" | head -1)"
    return 0
  fi
  log "refused, but NOT for the expected reason ($REFUSAL_REASON_RE): $(printf '%s' "$1" | tail -3 | head -c 400)"
  return 1
}

# A tx that must be refused: either the CLI fails (simulation under --gas auto) or the chain returns code != 0
tx_refused() { # cmd...
  local out
  out=$("$@" 2>&1) && out=$(printf '%s\n' "$out" | grep '^{' | tail -1) || { assert_refusal_reason "$out"; return; }
  [[ "$(jq -r '.code' <<<"$out")" != 0 ]] || { log "tx was accepted, expected a refusal"; return 1; }
  assert_refusal_reason "$(jq -r '.raw_log' <<<"$out")"
}

# Vault multisig tx that must fail on chain (ms_tx logs the tx raw_log to stderr via wait_tx)
vault_tx_refused() {
  local out
  out=$(ms_tx osmosisd vault-ms m1,m2,m3 -- "$@" 2>&1) && { log "vault tx was accepted, expected a refusal"; return 1; }
  assert_refusal_reason "$out"
}

# Sets LAST_POOL_ID (a global: this must not run in a command substitution, ms_tx logs to stdout on failure)
create_pool() { # name sttoken native rate subdenom
  local inst
  inst=$(bash "$REHEARSAL_DIR/instantiate_pool.sh" "$2" "$3" "$4" "$5" "$VAULT_MS_OSMO")
  log "instantiate message for pool $1: $(jq -c . <<<"$inst")"

  # Instantiating the transmuter inside create-pool needs far more gas than the multisig default
  MS_GAS=2500000 ms_tx osmosisd vault-ms m1,m2,m3 -- cosmwasmpool create-pool 1 "$(jq -c . <<<"$inst")" >/dev/null

  # Pool ids are sequential across all pool types, so the newest pool's id is the pool count
  LAST_POOL_ID=$(osmosisd q poolmanager num-pools -o json | jq -r '.num_pools')
  [[ "$LAST_POOL_ID" =~ ^[1-9][0-9]*$ ]] || { log "bad pool count '$LAST_POOL_ID'"; return 1; }

  # Guard against a pool id that points at some other pool: the contract must list the expected stToken
  local configs
  configs=$(osmosisd q wasm contract-state smart "$(pool_contract "$LAST_POOL_ID")" '{"list_asset_configs":{}}' -o json)
  jq -e --arg denom "$2" '[.. | .denom? // empty] | index($denom)' <<<"$configs" >/dev/null \
    || { log "pool $LAST_POOL_ID asset configs lack $2: $(jq -c . <<<"$configs")"; return 1; }
}

pool_contract() { osmosisd q poolmanager pool "$1" -o json | jq -r '.pool.contract_address'; }

vault_balance() { osmosisd q bank balance "$VAULT_MS_OSMO" "$1" -o json | jq -r '.balance.amount'; }

# Execute a transmuter message as the vault
vault_exec() { # contract json [--amount coin]
  ms_tx osmosisd vault-ms m1,m2,m3 -- wasm execute "$@" >/dev/null
}

# Join with the full allocation, then mark the native asset corrupted so it can never be raised again
fund_pool() { # contract amount denom
  vault_exec "$1" '{"join_pool":{}}' --amount "${2}${3}"
  vault_exec "$1" "{\"mark_corrupted_assets\":{\"denoms\":[\"$3\"]}}"
}

# ---------------------------------------------------------------- code id 1

# Mainnet bytecode (code id 1 is whitelisted at genesis; uploading is only needed if the genesis lacks it)
if ! osmosisd q wasm code-info 1 >/dev/null 2>&1; then
  $KX cp "$REHEARSAL_DIR/transmuter_v3.2.0.wasm" osmosis-validator-0:/tmp/transmuter.wasm -c validator
  log_cmd "store code" osmosisd tx wasm store /tmp/transmuter.wasm --from val1 $OSMO_TX --gas 5000000
  sleep 6
fi
checkpoint "code 1 stored" osmosisd q wasm code-info 1

# ---------------------------------------------------------------- create the pools

if [[ -z "${POOL_ATOM:-}" ]]; then
create_pool atom "$STATOM_CANON" "$ATOM_ON_OSMO" "$RATE_HUB" stATOMr;        POOL_ATOM=$LAST_POOL_ID
create_pool osmo "$STOSMO_CANON" uosmo "$RATE_OSMO" stOSMOr;                 POOL_OSMO=$LAST_POOL_ID
create_pool atomhub "$STATOM_HUBROUTE" "$ATOM_ON_OSMO" "$RATE_HUB" stATOMhub; POOL_ATOM_HUB=$LAST_POOL_ID
printf 'POOL_ATOM=%s\nPOOL_OSMO=%s\nPOOL_ATOM_HUB=%s\nPOOLS_FILE=%s\n' "$POOL_ATOM" "$POOL_OSMO" "$POOL_ATOM_HUB" "$POOLS_FILE" >> "$REHEARSAL_DIR/state.env"
else
  log "pools already created (rerun): canonical stATOM=$POOL_ATOM, stOSMO=$POOL_OSMO, Hub-route stATOM=$POOL_ATOM_HUB"
fi
log "pools: canonical stATOM=$POOL_ATOM, stOSMO=$POOL_OSMO, Hub-route stATOM=$POOL_ATOM_HUB"

CONTRACT_ATOM=$(pool_contract "$POOL_ATOM")
CONTRACT_OSMO=$(pool_contract "$POOL_OSMO")
CONTRACT_ATOM_HUB=$(pool_contract "$POOL_ATOM_HUB")

# pools file in the shape coverage_check.py reads: keyed by stToken, one entry per in-scope token
jq -n --arg atom_native "$ATOM_ON_OSMO" --arg atom_pool "$POOL_ATOM" --arg atom_st "$STATOM_CANON" \
      --arg hub_pool "$POOL_ATOM_HUB" --arg hub_st "$STATOM_HUBROUTE" \
      --arg osmo_pool "$POOL_OSMO" --arg osmo_st "$STOSMO_CANON" '{
  stuatom: {chain_id: "cosmoshub-test-1", native_denom_on_osmosis: $atom_native,
            canonical_pool_id: $atom_pool, canonical_st_denom: $atom_st,
            route_pools: [{channel_id: "channel-0", pool_id: $hub_pool, st_denom: $hub_st}]},
  stuosmo: {chain_id: "osmosis-test-1", native_denom_on_osmosis: "uosmo",
            canonical_pool_id: $osmo_pool, canonical_st_denom: $osmo_st, route_pools: []}
}' > "$POOLS_FILE"
log "pools file $POOLS_FILE: $(jq -c . "$POOLS_FILE")"

# ---------------------------------------------------------------- coverage before funding

# Allocation: route pool = escrow(channel-0 on Stride) x rate; canonical = the rest
STRIDE_EXPORT=$REHEARSAL_DIR/export_pre_pools.json
bash "$REHEARSAL_DIR/export.sh" "$STRIDE_EXPORT"
CHECKPOINT_SOFT=1 checkpoint "coverage before funding (expected: shortfall)" \
  python3 "$REPO/scripts/wind-down/coverage_check.py" --export "$STRIDE_EXPORT" --pools "$POOLS_FILE" --vault "$VAULT_MS_OSMO"

ESCROW_HUB=$(strided q ibc-transfer escrow-address transfer channel-0)
ROUTE_ESCROW=$(strided q bank balance "$ESCROW_HUB" stuatom -o json | jq -r .balance.amount)
ROUTE_ALLOC=$(python3 - "$ROUTE_ESCROW" "$RATE_HUB" <<'PY'
import math, sys
from decimal import Decimal, getcontext

getcontext().prec = 80
print(math.ceil(Decimal(sys.argv[1]) * Decimal(sys.argv[2])))
PY
)
ATOM_TOTAL=$(vault_balance "$ATOM_ON_OSMO")
UOSMO_FUND=$(( $(vault_balance uosmo) - UOSMO_GAS_RESERVE ))
log "route escrow $ROUTE_ESCROW stuatom -> route allocation $ROUTE_ALLOC; vault ATOM $ATOM_TOTAL; uosmo to fund $UOSMO_FUND"
(( ROUTE_ALLOC > 1000000 && ATOM_TOTAL >= ROUTE_ALLOC && UOSMO_FUND > 0 )) || { log "allocation arithmetic is impossible"; exit 1; }

# ---------------------------------------------------------------- funding

# Test deposit, then mark, prove a join is refused while marked, then unmark and complete the allocation
vault_exec "$CONTRACT_ATOM_HUB" '{"join_pool":{}}' --amount "1000000$ATOM_ON_OSMO"
vault_exec "$CONTRACT_ATOM_HUB" "{\"mark_corrupted_assets\":{\"denoms\":[\"$ATOM_ON_OSMO\"]}}"
checkpoint "join blocked while marked" vault_tx_refused wasm execute "$CONTRACT_ATOM_HUB" '{"join_pool":{}}' --amount "1$ATOM_ON_OSMO"
vault_exec "$CONTRACT_ATOM_HUB" "{\"unmark_corrupted_assets\":{\"denoms\":[\"$ATOM_ON_OSMO\"]}}"

fund_pool "$CONTRACT_ATOM_HUB" $(( ROUTE_ALLOC - 1000000 )) "$ATOM_ON_OSMO"
fund_pool "$CONTRACT_ATOM" $(( ATOM_TOTAL - ROUTE_ALLOC )) "$ATOM_ON_OSMO"
fund_pool "$CONTRACT_OSMO" "$UOSMO_FUND" uosmo

# Outsider joins the canonical pool with stTokens: shows as a remaining claim in the coverage check
user_tx "outsider join" osmosisd tx wasm execute "$CONTRACT_ATOM" '{"join_pool":{}}' --amount "1000000$STATOM_CANON" --from user1 $OSMO_TX

# ---------------------------------------------------------------- swaps

# Exact rate, rounded down; native -> stToken refused
QUOTE=$(osmosisd q wasm contract-state smart "$CONTRACT_ATOM" \
  "{\"calc_out_amt_given_in\":{\"token_in\":{\"denom\":\"$STATOM_CANON\",\"amount\":\"1000000\"},\"token_out_denom\":\"$ATOM_ON_OSMO\",\"swap_fee\":\"0\"}}" \
  -o json | jq -r '.data.token_out.amount')
EXPECTED_QUOTE=$(python3 - "$RATE_HUB" <<'PY'
import sys
from decimal import Decimal, getcontext

getcontext().prec = 80
print(int(Decimal(1000000) * Decimal(sys.argv[1])))
PY
)
checkpoint "quote = floor(1e6 x rate)" test "$QUOTE" = "$EXPECTED_QUOTE"

user_tx "swap statom->atom" osmosisd tx poolmanager swap-exact-amount-in "1000000$STATOM_CANON" 1 \
  --swap-route-pool-ids "$POOL_ATOM" --swap-route-denoms "$ATOM_ON_OSMO" --from user1 $OSMO_TX
checkpoint "atom->statom refused" tx_refused osmosisd tx poolmanager swap-exact-amount-in "1000${ATOM_ON_OSMO}" 1 \
  --swap-route-pool-ids "$POOL_ATOM" --swap-route-denoms "$STATOM_CANON" --from user1 $OSMO_TX
user_tx "route-pool swap (two-hop statom)" osmosisd tx poolmanager swap-exact-amount-in "1000000$STATOM_HUBROUTE" 1 \
  --swap-route-pool-ids "$POOL_ATOM_HUB" --swap-route-denoms "$ATOM_ON_OSMO" --from user1 $OSMO_TX

# ---------------------------------------------------------------- coverage after funding

bash "$REHEARSAL_DIR/export.sh" "$REHEARSAL_DIR/export_post_funding.json"
CHECKPOINT_SOFT=1 checkpoint "coverage after funding" python3 "$REPO/scripts/wind-down/coverage_check.py" \
  --export "$REHEARSAL_DIR/export_post_funding.json" --pools "$POOLS_FILE" --vault "$VAULT_MS_OSMO"

# ---------------------------------------------------------------- pool gate

# Fill the gate's constants block from what exists now. The edits are idempotent: each replaces a whole line.
COSMWASMPOOL_MODULE=$(osmosisd q auth module-account cosmwasmpool -o json | jq -r '[.. | .address? // empty] | first')
[[ "$COSMWASMPOOL_MODULE" == osmo1* ]] || { log "could not read the cosmwasmpool module account"; exit 1; }
RATE_HUB_ARG=$RATE_HUB RATE_OSMO_ARG=$RATE_OSMO POOL_ATOM_ARG=$POOL_ATOM POOL_OSMO_ARG=$POOL_OSMO \
POOL_ATOM_HUB_ARG=$POOL_ATOM_HUB HUBROUTE_TRACE_ARG=$HUBROUTE_TRACE MODULE_ARG=$COSMWASMPOOL_MODULE GATE_ARG=$GATE \
python3 - <<'PY'
import os, pathlib, re

env = os.environ
path = pathlib.Path(env["GATE_ARG"])
source = path.read_text()
pools = (
    "POOLS: list[PoolSpec] = ["
    f'PoolSpec(chain_id="cosmoshub-test-1", pool_id="{env["POOL_ATOM_ARG"]}", rate_at_creation="{env["RATE_HUB_ARG"]}", route_trace=None), '
    f'PoolSpec(chain_id="osmosis-test-1", pool_id="{env["POOL_OSMO_ARG"]}", rate_at_creation="{env["RATE_OSMO_ARG"]}", route_trace=None), '
    f'PoolSpec(chain_id="cosmoshub-test-1", pool_id="{env["POOL_ATOM_HUB_ARG"]}", rate_at_creation="{env["RATE_HUB_ARG"]}", route_trace="{env["HUBROUTE_TRACE_ARG"]}")'
    "]  # rehearsal phase 7"
)
source = re.sub(r"^POOLS: list\[PoolSpec\] = .*$", lambda _: pools, source, flags=re.M)
source = re.sub(r'^COSMWASMPOOL_MODULE = .*$', lambda _: f'COSMWASMPOOL_MODULE = "{env["MODULE_ARG"]}"', source, flags=re.M)

# The gate only recognizes the mainnet Osmosis chain id when it resolves the stOSMO pool's native denom
source = source.replace('zone.chain_id == "osmosis-1"', 'zone.chain_id == "osmosis-test-1"')
path.write_text(source)
PY
grep -q 'osmosis-test-1' "$GATE" || { log "pool gate was not patched for osmosis-test-1"; exit 1; }
grep -q REHEARSAL_FILL_IN "$GATE" && { log "pool gate still has REHEARSAL_FILL_IN"; exit 1; }
checkpoint "pool gate" python3 "$GATE"
