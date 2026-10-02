#!/bin/bash
# Shared driver for the wind-down rehearsal (spec docs/superpowers/specs/2026-10-02-wind-down-rehearsal-design.md).
set -euo pipefail
REHEARSAL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO="$(cd "$REHEARSAL_DIR/../.." && pwd)"
LOG=$REPO/docs/wind-down/rehearsal-log.md
source "$REHEARSAL_DIR/addresses.env"

[[ "$(kubectl config current-context)" == "integration" ]] || { echo "kube context must be 'integration'"; exit 1; }
KX="kubectl --context integration -n integration"
NEW_BIN=/home/validator/.stride/cosmovisor/upgrades/v35/bin/strided

strided_old() { $KX exec stride-validator-0 -c validator -- strided "$@"; }
strided_new() { $KX exec stride-validator-0 -c validator -- $NEW_BIN "$@"; }
# The pod's cosmovisor `current` symlink points at upgrades/v35 once the upgrade has applied
stride_is_v35() {
  local target
  target=$($KX exec stride-validator-0 -c validator -- readlink /home/validator/.stride/cosmovisor/current)
  [[ $target == *v35* ]]
}
strided() { if stride_is_v35; then strided_new "$@"; else strided_old "$@"; fi; }
strided_pod() { local pod=$1; shift; $KX exec "$pod" -c validator -- strided "$@"; }
gaiad()    { $KX exec cosmoshub-validator-0 -c validator -- gaiad "$@"; }
osmosisd() { $KX exec osmosis-validator-0 -c validator -- osmosisd "$@"; }

# Intentionally unquoted at call sites: these are word-split into flags
STRIDE_TX="--keyring-backend test --chain-id stride-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1ustrd -y -o json"
HUB_TX="--keyring-backend test --chain-id cosmoshub-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 1uatom -y -o json"
OSMO_TX="--keyring-backend test --chain-id osmosis-test-1 --gas auto --gas-adjustment 1.5 --gas-prices 0.04uosmo -y -o json"

log() { printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$LOG"; }
log_cmd() { # label, then the command; returns the command's exit status
  local label=$1; shift
  log "### $label"; printf '```\n$ %s\n' "$*" >> "$LOG"
  local out status=0; out=$("$@" 2>&1) || status=$?
  printf '%s\n```\n' "$out" >> "$LOG"; printf '%s\n' "$out"
  return $status
}
tx_hash() { jq -r '.txhash'; }
wait_tx() { # chain hash
  local chain=$1 hash=$2 res=""
  for _ in $(seq 1 30); do
    res=$($chain q tx "$hash" -o json 2>/dev/null) && break; sleep 2
  done
  [[ -n "${res:-}" ]] || { log "tx $hash not found on $chain" >&2; return 1; }
  local code; code=$(jq -r '.code' <<<"$res")
  log "tx $hash code=$code $(jq -r '.raw_log' <<<"$res" | head -c 300)" >&2
  [[ "$code" == "0" ]]
}
wait_until() { # timeout-s description cmd...
  local timeout=$1 desc=$2; shift 2; local start=$SECONDS
  until "$@" >/dev/null 2>&1; do
    (( SECONDS - start < timeout )) || { log "TIMEOUT waiting for: $desc"; return 1; }; sleep 5
  done; log "ready: $desc"
}
# next_epoch_start_time is nanoseconds; cut keeps the seconds
day_epoch_next_start() { strided q stakeibc show-epoch-tracker day -o json | jq -r '.epoch_tracker.next_epoch_start_time' | cut -c1-10; }
sleep_until() { [[ -n "${1:-}" ]] || return 0; local now; now=$(date +%s); (( $1 > now )) && sleep $(( $1 - now )) || true; }
ms_tx() { # chain multisig-name members-csv -- tx args...
  local chain=$1 ms=$2 members=$3; shift 4
  local m1=${members%%,*} rest=${members#*,} m2=${rest%%,*} chainid gasprice bin pod
  case $chain in
    strided_new)         chainid=stride-test-1;    gasprice=1ustrd;    bin=$NEW_BIN; pod=stride-validator-0;;
    strided_old)         chainid=stride-test-1;    gasprice=1ustrd;    bin=strided;  pod=stride-validator-0;;
    strided)             chainid=stride-test-1;    gasprice=1ustrd;    pod=stride-validator-0
                         if stride_is_v35; then bin=$NEW_BIN; else bin=strided; fi;;
    gaiad)               chainid=cosmoshub-test-1; gasprice=1uatom;    bin=gaiad;    pod=cosmoshub-validator-0;;
    osmosisd)            chainid=osmosis-test-1;   gasprice=0.04uosmo; bin=osmosisd; pod=osmosis-validator-0;;
  esac
  # The pod runs the whole generate / sign / multisign / broadcast pipeline so no file leaves the container
  local args; args=$(printf ' %q' "$@")
  local raw hash
  raw=$($KX exec $pod -c validator -- sh -c "
    set -e
    d=\$(mktemp -d)
    $bin tx $args --from $ms --generate-only --keyring-backend test --chain-id $chainid --gas 600000 --gas-prices $gasprice > \$d/unsigned.json
    $bin tx sign \$d/unsigned.json --from $m1 --multisig $ms --sign-mode amino-json --keyring-backend test --chain-id $chainid --output-document \$d/s1.json
    $bin tx sign \$d/unsigned.json --from $m2 --multisig $ms --sign-mode amino-json --keyring-backend test --chain-id $chainid --output-document \$d/s2.json
    $bin tx multisign \$d/unsigned.json $ms \$d/s1.json \$d/s2.json --keyring-backend test --chain-id $chainid --output-document \$d/signed.json
    $bin tx broadcast \$d/signed.json --chain-id $chainid -o json
    rm -rf \$d")
  printf '%s\n' "broadcast output ($chain):" '```' "$raw" '```' >> "$LOG"
  hash=$(tx_hash <<<"$raw")
  [[ -n "$hash" && "$hash" != null ]] || { log "ms_tx: no txhash in broadcast output for $chain"; return 1; }
  echo "$hash"; wait_tx $chain "$hash"
}
checkpoint() { # name cmd...
  local name=$1; shift
  if "$@"; then log "CHECKPOINT PASS: $name"; else log "CHECKPOINT FAIL: $name"; [[ "${CHECKPOINT_SOFT:-0}" == 1 ]] || exit 1; fi
}
rate_of() { strided q stakeibc show-host-zone "$1" -o json | jq -r '.host_zone.redemption_rate'; }
assert_rate_unchanged() { # chain-id expected
  [[ "$(rate_of "$1")" == "$2" ]]
}

export -f strided_old strided_new strided strided_pod gaiad osmosisd tx_hash wait_tx log rate_of day_epoch_next_start stride_is_v35 checkpoint
export KX NEW_BIN LOG
