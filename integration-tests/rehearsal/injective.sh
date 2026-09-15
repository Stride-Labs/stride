#!/usr/bin/env bash
#
# REHEARSAL ONLY — DO NOT MERGE
#
# Phase driver for the v34 Injective reconciliation upgrade rehearsal on the k8s integration
# network. Spec: docs/superpowers/specs/2026-09-15-v34-injective-rehearsal-design.md (§3–§6).
#
# Runs from the laptop against context/namespace `integration`, doing everything through
# `kubectl exec` with the strided / gaiad / rly CLIs inside the pods. Phases are run in order:
#
#   injective.sh setup            register the zone, add validators, liquid stake 3000 ATOM (v33)
#   injective.sh redeem           two redemptions (500 / 400 stATOM) on consecutive day epochs
#   injective.sh drift            reproduce the lost-ack theft (run right after `redeem`)
#   injective.sh measure          print the delta table to paste into app/upgrades/v34/injective.go
#   -- paste the table, commit --
#   injective.sh build-and-swap   build the v34 binary and drop it into cosmovisor on all pods
#   injective.sh upgrade          gov proposal + assertions on the upgrade handler logs
#   injective.sh verify           spec §6 steps 1–5
#   injective.sh status           snapshot of the zone / records / ICA balances / channels
#   injective.sh wait-day-epoch [even|odd]
#
# Works with bash 3.2 (macOS default): no associative arrays, no mapfile.
set -euo pipefail

# =============================================================================================
# Constants
# =============================================================================================
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INTEGRATION_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_ROOT="$(cd "$INTEGRATION_DIR/.." && pwd)"
KEYS_FILE="$INTEGRATION_DIR/network/configs/keys.json"
ADMINS_FILE="$REPO_ROOT/utils/admins.go"
MAINNET_ADMIN_ADDRESS=stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh # build.sh swaps this for the keys.json admin

NAMESPACE=integration
STRIDE_CHAIN_ID=stride-test-1
HOST_CHAIN_ID=cosmoshub-test-1
GAIA_CHAIN_ID=$HOST_CHAIN_ID
STRIDE_POD=stride-validator-0
STRIDE_PODS="stride-validator-0 stride-validator-1 stride-validator-2"
GAIA_POD=cosmoshub-validator-0
CHAIN_CONTAINER=validator
RELAYER_DEPLOYMENT=relayer-stride-cosmoshub
RELAYER_PATH=stride-cosmoshub

ADMIN_KEY="admin"
USER_KEY="user1"
STRIDE_GAS_PRICES=0.025ustrd # same prices the TS client uses (minimum-gas-prices is 0, gaia has feemarket)
GAIA_GAS_PRICES=1uatom

CONNECTION_ID=connection-0
TRANSFER_CHANNEL=channel-0
HOST_DENOM=uatom
ST_DENOM=stuatom
BECH32_PREFIX=cosmos
IBC_DENOM="ibc/$(printf 'transfer/%s/%s' "$TRANSFER_CHANNEL" "$HOST_DENOM" | shasum -a 256 | awk '{print toupper($1)}')"
UNBONDING_PERIOD_DAYS=7 # frequency 2: unbondings on even day epochs, pending undelegations on odd ones
LSM_ENABLED=true
MAX_MESSAGES_PER_ICA_TX=2
MIN_REDEMPTION_RATE=0.9
MAX_REDEMPTION_RATE=1.5
VALIDATOR_WEIGHT=10

DELEGATION_ICA_OWNER="${HOST_CHAIN_ID}.DELEGATION"
DELEGATION_ICA_PORT="icacontroller-${DELEGATION_ICA_OWNER}"

UPGRADE_NAME=v34
DAEMON_HOME=/home/validator/.stride
UPGRADE_BINARY_PATH="${DAEMON_HOME}/cosmovisor/upgrades/${UPGRADE_NAME}/bin/strided"
CURRENT_BINARY_PATH="${DAEMON_HOME}/cosmovisor/current/bin/strided"

LIQUID_STAKE_AMOUNT=3000000000 # 3000 ATOM
REDEEM_AMOUNT_1=500000000      # 500 stATOM
REDEEM_AMOUNT_2=400000000      # 400 stATOM
DRIFT_STAKE_AMOUNT=300000000   # 300 ATOM

STRIDE_EPOCH_SECONDS=45
DAY_EPOCH_SECONDS=180
HOST_UNBONDING_SECONDS=240

POLL_INTERVAL=${POLL_INTERVAL:-3}
TX_TIMEOUT=60
ICA_TIMEOUT=300       # ICA registration / restore handshakes and single ICA round trips
DEPOSIT_TIMEOUT=600   # transfer + delegate across two stride epochs, with a retry
UNBONDING_TIMEOUT=900 # up to two day epochs plus the undelegate ack / host unbonding
RELAYER_TIMEOUT=180
UPGRADE_TIMEOUT=600

SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/injective-rehearsal.XXXXXX")"

# =============================================================================================
# Generic helpers
# =============================================================================================
log() { printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"; }
die() { log "ERROR: $*" >&2; exit 1; }
banner() { printf '\n===== %s =====\n' "$*"; }

assert_eq() { # <description> <actual> <expected>
    local description=$1 actual=$2 expected=$3
    [[ "$actual" == "$expected" ]] || die "FAIL $description: got '$actual', expected '$expected'"
    log "PASS $description: $actual"
}

assert_ne() { # <description> <actual> <unexpected>
    local description=$1 actual=$2 unexpected=$3
    [[ "$actual" != "$unexpected" ]] || die "FAIL $description: got '$actual'"
    log "PASS $description: $actual"
}

assert_ge() { # <description> <actual> <minimum>   (integers)
    local description=$1 actual=$2 minimum=$3
    [[ $actual =~ ^-?[0-9]+$ ]] || die "FAIL $description: '$actual' is not an integer"
    (( actual >= minimum )) || die "FAIL $description: got $actual, expected >= $minimum"
    log "PASS $description: $actual >= $minimum"
}

# wait_for <timeout-seconds> <description> <predicate> [args...]
# Polls the predicate every POLL_INTERVAL seconds until it exits 0. Predicates run in a
# condition context, so a failing kubectl inside them just counts as "not yet".
try_wait_for() { # like wait_for but returns 1 on timeout instead of dying
    local timeout=$1 description=$2
    shift 2
    local deadline=$(( $(date +%s) + timeout ))
    log "waiting up to ${timeout}s for: $description"
    until "$@"; do
        (( $(date +%s) < deadline )) || { log "timed out after ${timeout}s waiting for: $description"; return 1; }
        sleep "$POLL_INTERVAL"
    done
    log "ok: $description"
}
wait_for() { try_wait_for "$@" || die "giving up: $2"; }

is_int() { [[ ${1:-} =~ ^[0-9]+$ ]]; }

# =============================================================================================
# Chain helpers
# =============================================================================================
kube() { kubectl -n "$NAMESPACE" "$@"; }

# `strided` on PATH is the genesis (v33) binary. Once cosmovisor has switched, its `current`
# symlink points at the v34 binary, so prefer that for CLI calls against the upgraded node.
# VERIFY: cosmovisor v1.1.0 maintains ${DAEMON_HOME}/cosmovisor/current (falls back to PATH otherwise)
STRIDE_BINARY_SHIM="bin=$CURRENT_BINARY_PATH; [ -x \"\$bin\" ] || bin=strided; exec \"\$bin\" \"\$@\""

chain_exec() { # <stride|gaia> <binary args...>
    local chain=$1
    shift
    case $chain in
        stride) kube exec "$STRIDE_POD" -c "$CHAIN_CONTAINER" -- sh -c "$STRIDE_BINARY_SHIM" strided "$@" ;;
        gaia) kube exec "$GAIA_POD" -c "$CHAIN_CONTAINER" -- gaiad "$@" ;;
        *) die "unknown chain '$chain'" ;;
    esac
}

stride_q() { chain_exec stride q "$@" -o json; }
gaia_q() { chain_exec gaia q "$@" -o json; }

stride_tx() { # <from-key> <tx args...>
    local from=$1
    shift
    broadcast_and_wait stride "$@" --from "$from" --chain-id "$STRIDE_CHAIN_ID" --gas-prices "$STRIDE_GAS_PRICES"
}

gaia_tx() { # <from-key> <tx args...>
    local from=$1
    shift
    broadcast_and_wait gaia "$@" --from "$from" --chain-id "$GAIA_CHAIN_ID" --gas-prices "$GAIA_GAS_PRICES"
}

# Broadcasts, waits for inclusion, and fails loudly (with raw_log) on a non-zero code
broadcast_and_wait() { # <chain> <tx args...>
    local chain=$1
    shift
    local response hash code
    response=$(chain_exec "$chain" tx "$@" --keyring-backend test --gas auto --gas-adjustment 1.5 -y -o json)
    code=$(jq -r '.code' <<<"$response")
    hash=$(jq -r '.txhash' <<<"$response")
    [[ $code == 0 ]] || die "broadcast rejected ($chain tx $*): $response"

    wait_for "$TX_TIMEOUT" "$chain tx $hash included" tx_included "$chain" "$hash"
    local result
    result=$(chain_exec "$chain" q tx "$hash" -o json)
    code=$(jq -r '.code' <<<"$result")
    [[ $code == 0 ]] || die "$chain tx $hash failed with code $code: $(jq -r '.raw_log' <<<"$result")"
    log "$chain tx $hash ok (height $(jq -r '.height' <<<"$result"))"
}

tx_included() { chain_exec "$1" q tx "$2" -o json >/dev/null 2>&1; }

key_mnemonic() { # <name>
    jq -er --arg name "$1" '[.admin, .faucet, .validators[], .relayers[], .users[]] | map(select(.name == $name)) | .[0].mnemonic' "$KEYS_FILE"
}

key_address() { chain_exec "$1" keys show "$2" -a --keyring-backend test; }

recover_key() { # <chain> <name>   (idempotent)
    local chain=$1 name=$2
    if chain_exec "$chain" keys show "$name" -a --keyring-backend test >/dev/null 2>&1; then
        log "$chain: key $name already in the keyring"
        return
    fi

    local pod=$STRIDE_POD binary=strided
    [[ $chain == gaia ]] && { pod=$GAIA_POD; binary=gaiad; }
    key_mnemonic "$name" | kube exec -i "$pod" -c "$CHAIN_CONTAINER" -- "$binary" keys add "$name" --recover --keyring-backend test >/dev/null
    log "$chain: recovered key $name ($(key_address "$chain" "$name"))"
}

# =============================================================================================
# Relayer helpers
# =============================================================================================
rly_manual() { kube exec "deploy/$RELAYER_DEPLOYMENT" -- rly "$@"; }

# The chart's rolling update keeps the old daemon alive until the new pod passes its readiness
# probe (30s initial delay). That is longer than the gap between a deposit reaching
# DELEGATION_QUEUE and the delegate ICA at the next stride epoch, and an old daemon still running
# at that moment would deliver the ack we need stranded. Recreate + a 1s grace period turns the
# switch into a few seconds. Idempotent.
prepare_relayer_deployment() {
    kube patch "deployment/$RELAYER_DEPLOYMENT" --type merge \
        -p '{"spec":{"strategy":{"type":"Recreate","rollingUpdate":null},"template":{"spec":{"terminationGracePeriodSeconds":1}}}}'
    kube rollout status "deployment/$RELAYER_DEPLOYMENT" --timeout="${RELAYER_TIMEOUT}s"
}

relayer_mode() { # manual|daemon
    local mode=$1
    case $mode in
        manual) kube set env "deployment/$RELAYER_DEPLOYMENT" RELAYER_MANUAL=true ;;
        daemon) kube set env "deployment/$RELAYER_DEPLOYMENT" RELAYER_MANUAL- ;;
        *) die "relayer_mode: expected manual|daemon, got '$mode'" ;;
    esac
    wait_for "$RELAYER_TIMEOUT" "a single relayer pod running in $mode mode" relayer_pod_is "$mode"

    # The readiness probe only starts 30s in; check the restored config directly instead
    wait_for "$RELAYER_TIMEOUT" "rly config restored (rly q channels shows STATE_OPEN)" rly_channels_open
    if [[ $mode == daemon ]]; then
        wait_for "$RELAYER_TIMEOUT" "relayer daemon started" relayer_daemon_started
    fi
}

relayer_pod_is() { # <manual|daemon>
    local want=0
    [[ $1 == manual ]] && want=1
    kube get pods -l "app=$RELAYER_DEPLOYMENT" -o json 2>/dev/null | jq -e --argjson want "$want" '
        (.items | length) == 1
        and .items[0].status.phase == "Running"
        and ([.items[0].spec.containers[0].env[]? | select(.name == "RELAYER_MANUAL" and .value == "true")] | length) == $want
    ' >/dev/null
}

rly_channels_open() { rly_manual q channels stride 2>/dev/null | grep -q STATE_OPEN; }
relayer_daemon_started() { kube logs "deploy/$RELAYER_DEPLOYMENT" --tail=200 2>/dev/null | grep -q "Starting relayer"; }

# =============================================================================================
# State readers
# =============================================================================================
host_zone() { stride_q stakeibc show-host-zone "$HOST_CHAIN_ID" | jq '.host_zone'; }
host_zone_field() { host_zone | jq -r "$1"; }
host_zone_exists() { stride_q stakeibc show-host-zone "$HOST_CHAIN_ID" >/dev/null 2>&1; }
tracked_total() { host_zone | jq -r '[.validators[].delegation | tonumber] | add // 0'; }
tracked_delegations() { host_zone | jq -r '.validators[] | "\(.address) \(.name) \(.delegation)"'; } # "valoper name amount" lines

gaia_delegations_json() { gaia_q staking delegations "$1"; }
gaia_delegated_total() { gaia_delegations_json "$1" | jq -r '[.delegation_responses[].balance.amount | tonumber] | add // 0'; }
gaia_delegation_to() { # <delegations-json> <valoper>
    jq -r --arg v "$2" '[.delegation_responses[] | select(.delegation.validator_address == $v) | .balance.amount | tonumber] | add // 0' <<<"$1"
}
gaia_balance() { gaia_q bank balances "$1" | jq -r --arg d "$HOST_DENOM" '[.balances[] | select(.denom == $d) | .amount | tonumber] | add // 0'; }
stride_balance() { stride_q bank balances "$1" | jq -r --arg d "$2" '[.balances[] | select(.denom == $d) | .amount | tonumber] | add // 0'; }

# This host zone's unbonding records, oldest first
unbonding_records() {
    stride_q records list-epoch-unbonding-record | jq --arg c "$HOST_CHAIN_ID" '
        [.epoch_unbonding_record[] | .epoch_number as $e | .host_zone_unbondings[] | select(.host_zone_id == $c)
            | {epoch_number: $e, status, native_token_amount, st_token_amount, unbonding_time, undelegation_txs_in_progress}]
        | sort_by(.epoch_number | tonumber)'
}
# Records are identified by their st amount (500 / 400 stATOM), which is unique per redemption here
unbonding_record() { unbonding_records | jq -e --arg st "$1" '[.[] | select(.st_token_amount == $st)][-1] // empty'; } # latest record with that st amount (an earlier same-size redeem may already be CLAIMABLE)
record_field() { unbonding_record "$1" | jq -r ".$2"; } # <st-amount> <field>
print_unbonding_records() { unbonding_records | jq -c '.[]' | sed 's/^/    /'; }

deposit_records() { stride_q records list-deposit-record | jq --arg c "$HOST_CHAIN_ID" '[.deposit_record[] | select(.host_zone_id == $c)]'; }
# The epoch's STRIDE-source record also absorbs a few thousand uatom of per-epoch dust, so match on >= amount
deposit_record_status() { deposit_records | jq -r --arg a "$1" '[.[] | select(.source == "STRIDE" and (.amount|tonumber) >= ($a|tonumber))] | if length == 0 then "none" else .[0].status end'; }
print_deposit_records() { deposit_records | jq -c '.[]' | sed 's/^/    /'; }

stride_channels() { stride_q ibc channel channels | jq '.channels'; }
open_delegation_channel() { stride_channels | jq -r --arg p "$DELEGATION_ICA_PORT" '[.[] | select(.port_id == $p and .state == "STATE_OPEN") | .channel_id] | join(" ")'; }
channel_state() { stride_q ibc channel end "$DELEGATION_ICA_PORT" "$1" | jq -r '.channel.state'; }
next_sequence_send() { stride_q ibc channel next-sequence-send "$DELEGATION_ICA_PORT" "$1" | jq -r '.next_sequence_send'; }
print_ica_channels() { stride_channels | jq -r '.[] | select(.port_id | startswith("icacontroller-")) | "    \(.channel_id) \(.state) \(.port_id)"'; }

epoch_number() { stride_q epochs epoch-infos | jq -r --arg id "$1" '.epochs[] | select(.identifier == $id) | .current_epoch'; }
day_epoch() { epoch_number day; }
parity_of() { (( $1 % 2 == 0 )) && echo even || echo odd; }

stride_logs() { kube logs "$STRIDE_POD" -c "$CHAIN_CONTAINER" "$@"; }
log_contains() { stride_logs 2>/dev/null | grep -q -- "$1"; }
log_count() { stride_logs 2>/dev/null | grep -c -- "$1" || true; }
block_height() { kube exec "$1" -c "$CHAIN_CONTAINER" -- sh -c "$STRIDE_BINARY_SHIM" strided status 2>/dev/null | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height'; }

# Seconds until the record's unbonding matures on the host (unbonding_time is unix nanoseconds)
seconds_until_mature() { # <st-amount>
    local unbonding_time
    unbonding_time=$(record_field "$1" unbonding_time)
    is_int "$unbonding_time" || { echo "?"; return; }
    (( unbonding_time > 1000000000000 )) && unbonding_time=$(( unbonding_time / 1000000000 ))
    echo $(( unbonding_time - $(date +%s) ))
}

# =============================================================================================
# Predicates for wait_for
# =============================================================================================
ica_addresses_set() {
    host_zone 2>/dev/null | jq -e '[.delegation_ica_address, .withdrawal_ica_address, .redemption_ica_address, .fee_ica_address] | all(length > 0)' >/dev/null
}
host_zone_has_validators() { local n; n=$(host_zone 2>/dev/null | jq '.validators | length'); is_int "$n" && (( n >= $1 )); }
stride_balance_at_least() { local b; b=$(stride_balance "$1" "$2" 2>/dev/null); is_int "$b" && (( b >= $3 )); }
gaia_balance_is() { [[ $(gaia_balance "$1" 2>/dev/null) == "$2" ]]; }
tracked_total_is() { [[ $(tracked_total 2>/dev/null) == "$1" ]]; }
total_delegations_is() { [[ $(host_zone_field .total_delegations 2>/dev/null) == "$1" ]]; }
gaia_delegated_total_is() { [[ $(gaia_delegated_total "$1" 2>/dev/null) == "$2" ]]; }
gaia_delegated_total_at_least() { local t; t=$(gaia_delegated_total "$1" 2>/dev/null); is_int "$t" && (( t >= $2 )); }
day_epoch_reached() { local e; e=$(day_epoch 2>/dev/null); is_int "$e" && (( e >= $1 )); }
record_status_is() { [[ $(record_field "$1" status 2>/dev/null) == "$2" ]]; }
deposit_record_status_is() { [[ $(deposit_record_status "$1" 2>/dev/null) == "$2" ]]; }
sequence_advanced() { local s; s=$(next_sequence_send "$1" 2>/dev/null); is_int "$s" && (( s > $2 )); }
channel_state_is() { [[ $(channel_state "$1" 2>/dev/null) == "$2" ]]; }
new_delegation_channel_open() { local ch; ch=$(open_delegation_channel 2>/dev/null); [[ -n $ch && $ch != "$1" ]]; }
stuck_state_reached() { # <delegation-ica>: re-delegate acked (record gone) and on-chain >= tracked + 300
    [[ $(deposit_record_status "$DRIFT_STAKE_AMOUNT" 2>/dev/null) == none ]] || return 1
    local onchain tracked
    onchain=$(gaia_delegated_total "$1" 2>/dev/null)
    tracked=$(tracked_total 2>/dev/null)
    is_int "$onchain" && is_int "$tracked" && (( onchain >= tracked + DRIFT_STAKE_AMOUNT ))
}

# =============================================================================================
# Shared steps
# =============================================================================================
wait_day_epoch() { # [even|odd]   blocks until the next day epoch (with the given parity) starts
    local parity=${1:-} now target
    now=$(day_epoch)
    target=$(( now + 1 ))
    if [[ -n $parity && $(parity_of "$target") != "$parity" ]]; then
        target=$(( now + 2 ))
    fi
    wait_for $(( (target - now) * DAY_EPOCH_SECONDS + 60 )) "day epoch $target${parity:+ ($parity)} to start (now $now)" day_epoch_reached "$target"
}

ensure_day_epoch_parity() { # even|odd
    local now
    now=$(day_epoch)
    if [[ $(parity_of "$now") == "$1" ]]; then
        log "day epoch $now is $1"
        return
    fi
    wait_day_epoch "$1"
}

# Every host zone validator's tracked delegation must equal the delegation ICA's on-chain stake
assert_ledger_matches_chain() { # <delegation-ica>
    local onchain address name tracked
    onchain=$(gaia_delegations_json "$1")
    while read -r address name tracked; do
        assert_eq "$name tracked == on-chain" "$tracked" "$(gaia_delegation_to "$onchain" "$address")"
    done < <(tracked_delegations)
    assert_eq "TotalDelegations == Σ on-chain" "$(host_zone_field .total_delegations)" \
        "$(jq -r '[.delegation_responses[].balance.amount | tonumber] | add // 0' <<<"$onchain")"
}

assert_blocks_advancing() {
    local pod before after
    for pod in $STRIDE_PODS; do
        before=$(block_height "$pod")
        sleep 5
        after=$(block_height "$pod")
        if ! is_int "$before" || ! is_int "$after" || (( after <= before )); then
            die "$pod not producing blocks (height $before -> $after)"
        fi
        log "PASS $pod producing blocks: $before -> $after"
    done
}

# =============================================================================================
# Phases
# =============================================================================================
phase_setup() {
    banner "setup (v33): register $HOST_CHAIN_ID, add validators, liquid stake 3000 ATOM"
    prepare_relayer_deployment

    # 1. keys
    recover_key stride "$ADMIN_KEY"
    recover_key stride "$USER_KEY"
    recover_key gaia "$USER_KEY"
    local user_stride user_gaia
    user_stride=$(key_address stride "$USER_KEY")
    user_gaia=$(key_address gaia "$USER_KEY")
    log "user1 stride=$user_stride gaia=$user_gaia ibc denom=$IBC_DENOM"

    # 2. host zone (unbonding period 7 → frequency 2, see the constants)
    if host_zone_exists; then
        log "host zone $HOST_CHAIN_ID already registered"
    else
        stride_tx "$ADMIN_KEY" stakeibc register-host-zone "$CONNECTION_ID" "$HOST_DENOM" "$BECH32_PREFIX" "$IBC_DENOM" \
            "$TRANSFER_CHANNEL" "$UNBONDING_PERIOD_DAYS" "$LSM_ENABLED" \
            --min-redemption-rate "$MIN_REDEMPTION_RATE" --max-redemption-rate "$MAX_REDEMPTION_RATE" \
            --max-messages-per-ica-tx "$MAX_MESSAGES_PER_ICA_TX"
    fi
    wait_for "$ICA_TIMEOUT" "all four ICA addresses registered" ica_addresses_set
    assert_eq "unbonding period" "$(host_zone_field .unbonding_period)" "$UNBONDING_PERIOD_DAYS"
    host_zone | jq -c '{delegation_ica_address, withdrawal_ica_address, redemption_ica_address, fee_ica_address}'

    # 3. validators
    if host_zone_has_validators 1; then
        log "host zone already has validators"
    else
        add_gaia_validators
    fi
    wait_for 60 "3 validators on the host zone" host_zone_has_validators 3
    host_zone | jq -c '.validators[] | {name, address, weight}'

    # 4. fund user1 on Stride with ATOM
    if stride_balance_at_least "$user_stride" "$IBC_DENOM" "$LIQUID_STAKE_AMOUNT"; then
        log "user1 already holds >= $LIQUID_STAKE_AMOUNT $IBC_DENOM on Stride"
    else
        # the drift phase stakes another DRIFT_STAKE_AMOUNT later, so fund both up front
        gaia_tx "$USER_KEY" ibc-transfer transfer transfer "$TRANSFER_CHANNEL" "$user_stride" "$(( LIQUID_STAKE_AMOUNT + DRIFT_STAKE_AMOUNT ))${HOST_DENOM}"
        wait_for "$ICA_TIMEOUT" "user1 ibc/uatom balance on Stride >= $(( LIQUID_STAKE_AMOUNT + DRIFT_STAKE_AMOUNT ))" \
            stride_balance_at_least "$user_stride" "$IBC_DENOM" "$(( LIQUID_STAKE_AMOUNT + DRIFT_STAKE_AMOUNT ))"
    fi

    # 5. liquid stake and wait for the full delegation
    local ica
    ica=$(host_zone_field .delegation_ica_address)
    if (( $(tracked_total) >= LIQUID_STAKE_AMOUNT )); then
        log "tracked delegations already >= $LIQUID_STAKE_AMOUNT, skipping liquid stake"
    else
        stride_tx "$USER_KEY" stakeibc liquid-stake "$LIQUID_STAKE_AMOUNT" "$HOST_DENOM"
        log "user1 $ST_DENOM balance: $(stride_balance "$user_stride" "$ST_DENOM")"
    fi
    wait_for "$DEPOSIT_TIMEOUT" "tracked delegations == $LIQUID_STAKE_AMOUNT" tracked_total_is "$LIQUID_STAKE_AMOUNT"
    wait_for 120 "Gaia delegations of $ica == $LIQUID_STAKE_AMOUNT" gaia_delegated_total_is "$ica" "$LIQUID_STAKE_AMOUNT"
    assert_eq "liquid-stake deposit record removed" "$(deposit_record_status "$LIQUID_STAKE_AMOUNT")" none
    assert_ledger_matches_chain "$ica"
    log "setup complete; redemption rate $(host_zone_field .redemption_rate)"
}

add_gaia_validators() {
    local file=/home/validator/rehearsal-validators.json
    # Written to a local file first: piping straight into `kubectl exec -i` can SIGPIPE the producer
    gaia_q staking validators \
        | jq --argjson w "$VALIDATOR_WEIGHT" '{validators: [.validators[] | select(.status == "BOND_STATUS_BONDED")
            | {name: .description.moniker, address: .operator_address, weight: $w}]}' \
        > "$SCRATCH/validators.json"
    kube exec -i "$STRIDE_POD" -c "$CHAIN_CONTAINER" -- sh -c "cat > $file" < "$SCRATCH/validators.json"
    log "validator list: $(jq -c . "$SCRATCH/validators.json")"
    stride_tx "$ADMIN_KEY" stakeibc add-validators "$HOST_CHAIN_ID" "$file"
}

phase_redeem() {
    banner "redeem: $REDEEM_AMOUNT_1 then $REDEEM_AMOUNT_2 $ST_DENOM on consecutive day epochs"
    local user_stride user_gaia
    user_stride=$(key_address stride "$USER_KEY")
    user_gaia=$(key_address gaia "$USER_KEY")
    log "user1 $ST_DENOM balance $(stride_balance "$user_stride" "$ST_DENOM"), redemption rate $(host_zone_field .redemption_rate)"

    # Unbondings run at the start of even day epochs. Redeeming first in an even epoch puts the
    # second redemption in the following odd epoch, so both records unbond at the very next even
    # epoch (one day-epoch wait instead of two).
    ensure_day_epoch_parity even
    redeem_once "$REDEEM_AMOUNT_1" "$user_gaia"
    wait_day_epoch
    redeem_once "$REDEEM_AMOUNT_2" "$user_gaia"

    wait_for "$UNBONDING_TIMEOUT" "record 1 ($REDEEM_AMOUNT_1) EXIT_TRANSFER_QUEUE" record_status_is "$REDEEM_AMOUNT_1" EXIT_TRANSFER_QUEUE
    wait_for 120 "record 2 ($REDEEM_AMOUNT_2) EXIT_TRANSFER_QUEUE" record_status_is "$REDEEM_AMOUNT_2" EXIT_TRANSFER_QUEUE
    print_unbonding_records
    log "R1 = $(record_field "$REDEEM_AMOUNT_1" native_token_amount) uatom (epoch $(record_field "$REDEEM_AMOUNT_1" epoch_number))"
    log "R2 = $(record_field "$REDEEM_AMOUNT_2" native_token_amount) uatom (epoch $(record_field "$REDEEM_AMOUNT_2" epoch_number))"
    log "unbonding matures in ~$(seconds_until_mature "$REDEEM_AMOUNT_2")s — run 'drift' NOW, before the bundled sweep can succeed"
}

redeem_once() { # <st-amount> <receiver>   (idempotent per amount)
    if unbonding_record "$1" >/dev/null 2>&1; then
        log "redemption of $1 already recorded in epoch $(record_field "$1" epoch_number)"
        return
    fi
    stride_tx "$USER_KEY" stakeibc redeem-stake "$1" "$HOST_CHAIN_ID" "$2"
    log "redeemed $1 $ST_DENOM in day epoch $(day_epoch)"
}

# The rly daemon keeps running (so the light clients never expire and every other channel is
# served) but with the delegation ICA channel denylisted; that channel is relayed by hand with
# `hermes tx packet-recv` (receive-or-timeout, never acks) from the CLI-only hermes pod. The
# delegate's ack is therefore the only thing stranded.
HERMES_DEPLOYMENT=hermes-stride-cosmoshub
relayer_deny_channel() { # <channel>
    kube set env "deployment/$RELAYER_DEPLOYMENT" "RELAYER_DENY_CHANNELS=$1"
    kube rollout status "deployment/$RELAYER_DEPLOYMENT" --timeout=180s
    wait_for "$RELAYER_TIMEOUT" "relayer daemon restarted with $1 denied" relayer_daemon_started
}
relayer_allow_all() {
    kube set env "deployment/$RELAYER_DEPLOYMENT" RELAYER_DENY_CHANNELS-
    kube rollout status "deployment/$RELAYER_DEPLOYMENT" --timeout=180s
    wait_for "$RELAYER_TIMEOUT" "relayer daemon restarted without a denylist" relayer_daemon_started
}
relayer_denylist_current() {
    kube get deployment "$RELAYER_DEPLOYMENT" -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="RELAYER_DENY_CHANNELS")].value}' 2>/dev/null
}
hermes_cli() { kube exec "deploy/$HERMES_DEPLOYMENT" -- hermes "$@" 2>&1; }
hermes_recv_loop_start() { # <channel>: receive-or-timeout every packet on the channel, never ack
    kube exec "deploy/$HERMES_DEPLOYMENT" -- bash -c "while true; do
        hermes tx packet-recv --dst-chain $HOST_CHAIN_ID --src-chain $STRIDE_CHAIN_ID --src-port $DELEGATION_ICA_PORT --src-channel $1 >>/tmp/hermes-loop.log 2>&1 || true
        sleep 3
    done" >"$SCRATCH/hermes-loop.exec.log" 2>&1 &
    HERMES_LOOP_PID=$!
    log "hermes receive-or-timeout loop started on $1 (pid $HERMES_LOOP_PID)"
}
hermes_recv_loop_stop() {
    kill "${HERMES_LOOP_PID:-0}" 2>/dev/null || true
    kube exec "deploy/$HERMES_DEPLOYMENT" -- pkill -f "while true" 2>/dev/null || true
    log "hermes loop stopped; last lines:"
    kube exec "deploy/$HERMES_DEPLOYMENT" -- sh -c 'tail -5 /tmp/hermes-loop.log 2>/dev/null' | sed 's/^/    /' || true
}
drift_cleanup() { # never leave the delegation channel unserviced
    hermes_recv_loop_stop
    if [[ -n $(relayer_denylist_current) ]]; then
        log "drift failed: removing the relayer denylist"
        relayer_allow_all || true
    fi
}

phase_drift() {
    banner "drift: reproduce the lost-ack theft"
    # `die` exits, which never fires ERR, so the cleanup hangs off EXIT and is disarmed on success
    DRIFT_DONE=false
    trap '[[ $DRIFT_DONE == true ]] || drift_cleanup' EXIT
    prepare_relayer_deployment
    local ica chan seq_before onchain_before tracked_before
    ica=$(host_zone_field .delegation_ica_address)
    chan=$(open_delegation_channel)
    [[ $chan =~ ^channel-[0-9]+$ ]] || die "expected exactly one OPEN delegation channel, got '$chan'"
    log "on-chain=$(gaia_delegated_total "$ica") tracked=$(tracked_total) liquid=$(gaia_balance "$ica") delegation channel=$chan"

    # 1. hermes receive-or-timeout loop first, THEN the daemon off the channel: the rly restart takes
    #    ~30s and any ICA sent at a stride boundary in that gap would time out (36s) and close the channel
    hermes_recv_loop_start "$chan"
    relayer_deny_channel "$chan"
    seq_before=$(next_sequence_send "$chan")
    onchain_before=$(gaia_delegated_total "$ica")
    tracked_before=$(tracked_total)
    log "next-sequence-send=$seq_before on-chain=$onchain_before tracked=$tracked_before"

    # 2. stake 300: the daemon relays the transfer, hermes receives the delegate, nobody relays its ack
    if [[ $(deposit_record_status "$DRIFT_STAKE_AMOUNT") == none ]]; then
        stride_tx "$USER_KEY" stakeibc liquid-stake "$DRIFT_STAKE_AMOUNT" "$HOST_DENOM"
    else
        log "drift deposit record already exists ($(deposit_record_status "$DRIFT_STAKE_AMOUNT"))"
    fi
    # reinvest delegates land on Gaia too (acks withheld as well), so on-chain only ever grows: >=
    wait_for "$DEPOSIT_TIMEOUT" "Gaia delegations >= $(( onchain_before + DRIFT_STAKE_AMOUNT )) (delegate executed on Gaia)" \
        gaia_delegated_total_at_least "$ica" "$(( onchain_before + DRIFT_STAKE_AMOUNT ))"
    sleep 10
    assert_eq "tracked delegations unchanged (ack stranded)" "$(tracked_total)" "$tracked_before"
    assert_eq "drift deposit record status" "$(deposit_record_status "$DRIFT_STAKE_AMOUNT")" DELEGATION_IN_PROGRESS

    # 3. close the channel with a 1ns-timeout ICA; hermes relays it as MsgTimeout
    stride_tx "$ADMIN_KEY" stakeibc close-delegation-channel "$HOST_CHAIN_ID"
    wait_for 120 "delegation channel $chan STATE_CLOSED" channel_state_is "$chan" STATE_CLOSED
    hermes_recv_loop_stop
    # The denylisted rly never sees the close event, and the host refuses a new ICA channel while
    # its end of the old one is still OPEN, so confirm the closure on the host by hand
    local counterparty
    counterparty=$(stride_q ibc channel end "$DELEGATION_ICA_PORT" "$chan" | jq -r '.channel.counterparty.channel_id')
    hermes_cli tx chan-close-confirm --dst-chain "$HOST_CHAIN_ID" --src-chain "$STRIDE_CHAIN_ID" --dst-connection "$CONNECTION_ID" \
        --dst-port icahost --src-port "$DELEGATION_ICA_PORT" --src-channel "$chan" --dst-channel "$counterparty" | grep -E "SUCCESS|ERROR" | head -1
    log "host end $counterparty of $chan: $(gaia_q ibc channel end icahost "$counterparty" | jq -r '.channel.state')"
    print_ica_channels

    # 4. daemon back on every channel, restore the account on a fresh channel
    relayer_allow_all
    stride_tx "$ADMIN_KEY" stakeibc restore-interchain-account "$HOST_CHAIN_ID" "$CONNECTION_ID" "$DELEGATION_ICA_OWNER"
    assert_eq "drift deposit record reset by restore" "$(deposit_record_status "$DRIFT_STAKE_AMOUNT")" DELEGATION_QUEUE
    # rly usually completes the handshake; if the INIT channel sits for a minute, nudge it with hermes
    if ! try_wait_for 60 "new OPEN delegation channel (not $chan)" new_delegation_channel_open "$chan"; then
        local init_channel
        init_channel=$(stride_channels | jq -r --arg p "$DELEGATION_ICA_PORT" '[.[] | select(.port_id == $p and .state == "STATE_INIT") | .channel_id][0] // empty')
        [[ -n $init_channel ]] || die "no INIT delegation channel to nudge"
        log "nudging $init_channel with hermes chan-open-try"
        hermes_cli tx chan-open-try --dst-chain "$HOST_CHAIN_ID" --src-chain "$STRIDE_CHAIN_ID" --dst-connection "$CONNECTION_ID" \
            --dst-port icahost --src-port "$DELEGATION_ICA_PORT" --src-channel "$init_channel" | grep -E "SUCCESS|ERROR" | head -1
        wait_for "$ICA_TIMEOUT" "new OPEN delegation channel (not $chan)" new_delegation_channel_open "$chan"
    fi
    assert_ne "delegation ICA address" "$(host_zone_field .delegation_ica_address)" ""
    log "delegation channel restored: $chan -> $(open_delegation_channel)"
    print_ica_channels
    local drift=$(( $(gaia_delegated_total "$ica") - $(tracked_total) ))
    assert_ge "on-chain - tracked (the stake plus any reinvest whose ack was withheld)" "$drift" "$DRIFT_STAKE_AMOUNT"
    DRIFT_DONE=true
    log "drift choreography complete: the re-delegate will retry every stride epoch until the ICA has liquid funds"
}

phase_stuck() { # wait for the retrying re-delegate to consume 300 of R1+R2 and the bundled sweep to fail
    banner "stuck: wait for the theft and the failing bundled sweep"
    local ica r1 r2
    ica=$(host_zone_field .delegation_ica_address)
    r1=$(record_field "$REDEEM_AMOUNT_1" native_token_amount)
    r2=$(record_field "$REDEEM_AMOUNT_2" native_token_amount)
    log "R1=$r1 R2=$r2 (latest records); deposit record: $(deposit_record_status "$DRIFT_STAKE_AMOUNT")"
    wait_for "$UNBONDING_TIMEOUT" "re-delegate acked (deposit record gone, on-chain >= tracked + $DRIFT_STAKE_AMOUNT)" stuck_state_reached "$ica"
    local drift=$(( $(gaia_delegated_total "$ica") - $(tracked_total) ))
    assert_ge "on-chain - tracked" "$drift" "$DRIFT_STAKE_AMOUNT"
    # reinvest dust moves through the ICA every epoch, so allow 1 ATOM of slack
    local expected_liquid=$(( r1 + r2 - drift )) liquid
    liquid=$(gaia_balance "$ica")
    assert_ge "delegation ICA liquid >= R1+R2-drift-1ATOM ($expected_liquid)" "$liquid" "$(( expected_liquid - 1000000 ))"
    assert_ge "R1+R2-drift+1ATOM >= delegation ICA liquid ($liquid)" "$(( expected_liquid + 1000000 ))" "$liquid"
    assert_ne "record 1 not swept" "$(record_field "$REDEEM_AMOUNT_1" status)" CLAIMABLE
    assert_ne "record 2 not swept" "$(record_field "$REDEEM_AMOUNT_2" status)" CLAIMABLE

    log "letting two more stride epochs run so the bundled sweep fails twice more"
    sleep $(( STRIDE_EPOCH_SECONDS * 2 + 10 ))
    assert_ge "delegation ICA liquid still short (R1+R2 = $(( r1 + r2 )))" "$(( r1 + r2 - 1000000 ))" "$(gaia_balance "$ica")"
    assert_ne "record 1 still not swept" "$(record_field "$REDEEM_AMOUNT_1" status)" CLAIMABLE
    assert_ne "record 2 still not swept" "$(record_field "$REDEEM_AMOUNT_2" status)" CLAIMABLE
    print_unbonding_records
    log "sweep log lines:"
    stride_logs --tail=20000 | grep -iE "sweep|Transferring .* to host zone" | tail -6 | sed 's/^/    /' || true
    log "drift complete: the zone is in the mainnet-incident state"
}

close_delegation_channel_via_timeout() { # <channel>
    local attempt
    for attempt in 1 2 3 4 5; do
        rly_manual tx relay-packets "$RELAYER_PATH" "$1" || true
        if channel_state_is "$1" STATE_CLOSED; then
            log "PASS delegation channel $1 STATE_CLOSED (attempt $attempt)"
            return
        fi
        sleep 5
    done
    die "delegation channel $1 did not close (state $(channel_state "$1"))"
}

phase_measure() {
    banner "measure: on-chain minus tracked, per validator"
    local ica onchain address name tracked actual delta total=0 table=""
    ica=$(host_zone_field .delegation_ica_address)
    onchain=$(gaia_delegations_json "$ica")
    while read -r address name tracked; do
        actual=$(gaia_delegation_to "$onchain" "$address")
        delta=$(( actual - tracked ))
        log "$name $address tracked=$tracked on-chain=$actual delta=$delta"
        # Mirrors the mainnet table: zero deltas are left out
        if (( delta != 0 )); then
            table+=$(printf '\t{Name: "%s", Address: "%s", Delta: mustInt("%s")},' "$name" "$address" "$delta")$'\n'
        fi
        total=$(( total + delta ))
    done < <(tracked_delegations)

    printf '\n// paste into app/upgrades/v34/injective.go — InjectiveDelegationDeltas (%s)\n%s\n' "$HOST_CHAIN_ID" "$table"
    log "Σ delta = $total uatom (expected $DRIFT_STAKE_AMOUNT)"
    log "TotalDelegations = $(host_zone_field .total_delegations); Σ on-chain = $(gaia_delegated_total "$ica")"
    log "delegation ICA liquid = $(gaia_balance "$ica") uatom"
    log "R1 = $(record_field "$REDEEM_AMOUNT_1" native_token_amount), R2 = $(record_field "$REDEEM_AMOUNT_2" native_token_amount) (liquid should equal R1+R2-$DRIFT_STAKE_AMOUNT)"
    log "redemption rate = $(host_zone_field .redemption_rate)"
}

phase_build_and_swap() {
    banner "build-and-swap: v34 binary into cosmovisor/upgrades/$UPGRADE_NAME on all Stride pods"
    log "building from $REPO_ROOT @ $(git -C "$REPO_ROOT" rev-parse --short HEAD) ($(git -C "$REPO_ROOT" status --porcelain | wc -l | tr -d ' ') dirty files)"

    # build.sh bakes the keys.json admin into the image; the swapped binary must match it
    local admin_address
    admin_address=$(jq -r '.admin.address' "$KEYS_FILE")
    cp "$ADMINS_FILE" "$SCRATCH/admins.go.orig"
    trap 'cp "$SCRATCH/admins.go.orig" "$ADMINS_FILE"' EXIT
    sed -E "s|$MAINNET_ADMIN_ADDRESS|$admin_address|g" "$SCRATCH/admins.go.orig" > "$ADMINS_FILE"
    (cd "$REPO_ROOT" && docker build --platform linux/amd64 -f Dockerfile -t core:stride .)
    cp "$SCRATCH/admins.go.orig" "$ADMINS_FILE"
    trap - EXIT

    local container
    container=$(docker create --platform linux/amd64 core:stride)
    docker cp "$container:/usr/local/bin/strided" "$SCRATCH/strided"
    docker rm "$container" >/dev/null
    log "extracted $(du -h "$SCRATCH/strided" | cut -f1) binary"

    local pod
    for pod in $STRIDE_PODS; do
        assert_eq "$pod UPGRADE_NAME" "$(kube exec "$pod" -c "$CHAIN_CONTAINER" -- printenv UPGRADE_NAME)" "$UPGRADE_NAME"
        kube cp "$SCRATCH/strided" "$pod:$UPGRADE_BINARY_PATH" -c "$CHAIN_CONTAINER"
        kube exec "$pod" -c "$CHAIN_CONTAINER" -- chmod +x "$UPGRADE_BINARY_PATH"
        log "$pod: $UPGRADE_BINARY_PATH version = $(kube exec "$pod" -c "$CHAIN_CONTAINER" -- "$UPGRADE_BINARY_PATH" version 2>&1 | tail -1)"
    done
    log "swap complete; cosmovisor only reads the upgrade binary at the upgrade height"
}

# The in-pod strided CLI takes ~8s per call on a 400m-CPU pod and each kubectl exec adds ~7s, so
# the harness's upgrade.sh cannot get three votes in before a 30s voting period closes. Submit and
# vote from the laptop instead: the local strided is ~0.2s per call against the RPC ingress.
STRIDE_RPC=${STRIDE_RPC:-http://stride-rpc.internal.stridenet.co:80}
LOCAL_STRIDED_HOME="$SCRATCH/strided-home"
UPGRADE_BUFFER_BLOCKS=60
GOV_AUTHORITY=stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
lstrided() { strided --home "$LOCAL_STRIDED_HOME" --node "$STRIDE_RPC" "$@"; }
import_local_keys() {
    local name
    for name in val1 val2 val3 admin; do
        lstrided keys show "$name" --keyring-backend test >/dev/null 2>&1 && continue
        key_mnemonic "$name" | lstrided keys add "$name" --recover --keyring-backend test >/dev/null
    done
}
submit_upgrade_proposal_local() {
    import_local_keys
    local height id status tx="--keyring-backend test --chain-id $STRIDE_CHAIN_ID --gas-prices $STRIDE_GAS_PRICES -y -o json"
    height=$(( $(lstrided status 2>/dev/null | jq -r '.sync_info.latest_block_height') + UPGRADE_BUFFER_BLOCKS ))
    printf '{"messages":[{"@type":"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade","authority":"%s","plan":{"name":"%s","height":"%s"}}],"deposit":"2000000000ustrd","title":"Upgrade %s","summary":"Upgrade %s"}\n' \
        "$GOV_AUTHORITY" "$UPGRADE_NAME" "$height" "$UPGRADE_NAME" "$UPGRADE_NAME" > "$SCRATCH/upgrade-proposal.json"
    log "submitting $UPGRADE_NAME upgrade proposal at height $height"
    lstrided tx gov submit-proposal "$SCRATCH/upgrade-proposal.json" --from val1 --gas 400000 $tx 2>/dev/null | jq -r '"submit code=\(.code)"'
    sleep 3
    id=$(lstrided q gov proposals -o json 2>/dev/null | jq -r '[.proposals[].id | tonumber] | max')
    local name
    for name in val1 val2 val3; do
        lstrided tx gov vote "$id" yes --from "$name" --gas 300000 $tx 2>/dev/null | jq -r --arg n "$name" '"vote \($n) code=\(.code)"'
        sleep 1
    done
    wait_for 90 "proposal $id decided" proposal_decided "$id"
    status=$(lstrided q gov proposal "$id" -o json 2>/dev/null | jq -r '.proposal.status')
    [[ $status == PROPOSAL_STATUS_PASSED ]] || die "proposal $id ended $status: $(lstrided q gov proposal "$id" -o json 2>/dev/null | jq -c '.proposal | {failed_reason, final_tally_result}')"
    log "proposal $id PASSED; upgrade $UPGRADE_NAME scheduled at height $height"
}
proposal_decided() { [[ $(lstrided q gov proposal "$1" -o json 2>/dev/null | jq -r '.proposal.status') =~ PASSED|REJECTED|FAILED ]]; }

phase_upgrade() {
    banner "upgrade: gov proposal for $UPGRADE_NAME"
    # upgrade.sh has one kubectl call without -n, so the context's default namespace must match
    local context_namespace
    context_namespace=$(kubectl config view --minify -o jsonpath='{..namespace}')
    [[ $context_namespace == "$NAMESPACE" ]] \
        || die "context namespace is '$context_namespace'; run: kubectl config set-context --current --namespace=$NAMESPACE"
    local pod
    for pod in $STRIDE_PODS; do
        kube exec "$pod" -c "$CHAIN_CONTAINER" -- test -x "$UPGRADE_BINARY_PATH" || die "$pod: $UPGRADE_BINARY_PATH missing, run build-and-swap"
    done

    log "pre-upgrade: TotalDelegations=$(host_zone_field .total_delegations) tracked=$(tracked_total) RR=$(host_zone_field .redemption_rate) day epoch=$(day_epoch)"
    print_unbonding_records
    submit_upgrade_proposal_local

    wait_for "$UPGRADE_TIMEOUT" "'Upgrade $UPGRADE_NAME complete' in $STRIDE_POD logs" log_contains "Upgrade $UPGRADE_NAME complete"
    local evidence
    evidence=$(stride_logs | grep -E "Starting upgrade $UPGRADE_NAME|v34: |Upgrade $UPGRADE_NAME complete" || true)
    printf '%s\n' "$evidence" | sed 's/^/    /'
    grep -q "Starting upgrade $UPGRADE_NAME" <<<"$evidence" || die "missing 'Starting upgrade $UPGRADE_NAME'"
    grep -q "TotalDelegations adjusted by" <<<"$evidence" || die "missing 'TotalDelegations adjusted by'"
    if grep -q "NOT applied" <<<"$evidence"; then die "the delta table was NOT applied"; fi
    log "PASS upgrade log lines present, no 'NOT applied'"
    assert_blocks_advancing
    log "post-upgrade: TotalDelegations=$(host_zone_field .total_delegations) tracked=$(tracked_total) RR=$(host_zone_field .redemption_rate)"
}

phase_verify() {
    banner "verify: spec §6 steps 1–5"
    local ica redemption user_gaia r1 r2 sum_delta
    ica=$(host_zone_field .delegation_ica_address)
    redemption=$(host_zone_field .redemption_ica_address)
    user_gaia=$(key_address gaia "$USER_KEY")
    r1=$(record_field "$REDEEM_AMOUNT_1" native_token_amount)
    r2=$(record_field "$REDEEM_AMOUNT_2" native_token_amount)
    sum_delta=$(stride_logs | grep -o "TotalDelegations adjusted by [0-9]*" | tail -1 | awk '{print $NF}')
    is_int "$sum_delta" || die "could not read the applied delta from the upgrade logs"
    log "R1=$r1 R2=$r2 Σdelta=$sum_delta delegation ICA=$ica redemption ICA=$redemption"

    # Baselines are taken before the undelegation can run (the sweep does not touch delegations)
    local total_before onchain_before rr_before submissions_before
    total_before=$(host_zone_field .total_delegations)
    onchain_before=$(gaia_delegated_total "$ica")
    rr_before=$(host_zone_field .redemption_rate)
    submissions_before=$(log_count "Submitting pending undelegation")
    (( submissions_before == 0 )) || log "WARN: pending undelegation already submitted ${submissions_before}x; before-values may be post-undelegation"
    log "before: TotalDelegations=$total_before on-chain=$onchain_before RR=$rr_before"
    log "records before the first sweep:"
    print_unbonding_records

    # 1. first stride epoch: per-record sweep pays record 1 only
    wait_for $(( STRIDE_EPOCH_SECONDS * 3 + 30 )) "record 1 CLAIMABLE" record_status_is "$REDEEM_AMOUNT_1" CLAIMABLE
    wait_for $(( STRIDE_EPOCH_SECONDS + 30 )) "record 2 back to EXIT_TRANSFER_QUEUE" record_status_is "$REDEEM_AMOUNT_2" EXIT_TRANSFER_QUEUE
    log "records after the first sweep:"
    print_unbonding_records
    wait_for 60 "redemption ICA balance == R1 ($r1)" gaia_balance_is "$redemption" "$r1"

    # 2. first odd day epoch: pending undelegation of Σdelta, acked → ledger == chain
    wait_for $(( DAY_EPOCH_SECONDS * 2 + 60 )) "'Submitting pending undelegation' in $STRIDE_POD logs" log_contains "Submitting pending undelegation"
    stride_logs | grep "pending undelegation" | tail -3 | sed 's/^/    /'
    wait_for 120 "Gaia delegations dropped by Σdelta to $(( onchain_before - sum_delta ))" gaia_delegated_total_is "$ica" "$(( onchain_before - sum_delta ))"
    wait_for 120 "TotalDelegations dropped by Σdelta to $(( total_before - sum_delta ))" total_delegations_is "$(( total_before - sum_delta ))"
    log "TotalDelegations $total_before -> $(host_zone_field .total_delegations); on-chain $onchain_before -> $(gaia_delegated_total "$ica")"
    assert_ledger_matches_chain "$ica"

    # 3. host unbonding period later: the returned Σdelta covers record 2
    wait_for $(( HOST_UNBONDING_SECONDS + STRIDE_EPOCH_SECONDS * 2 + 60 )) "record 2 CLAIMABLE" record_status_is "$REDEEM_AMOUNT_2" CLAIMABLE
    wait_for 60 "redemption ICA balance == R1+R2 ($(( r1 + r2 )))" gaia_balance_is "$redemption" "$(( r1 + r2 ))"
    print_unbonding_records

    # 4. claims
    local user_before epoch
    user_before=$(gaia_balance "$user_gaia")
    for epoch in $(claimable_epochs "$user_gaia"); do
        stride_tx "$USER_KEY" stakeibc claim-undelegated-tokens "$HOST_CHAIN_ID" "$epoch" "$user_gaia"
    done
    wait_for 180 "user1 Gaia balance == $user_before + R1 + R2" gaia_balance_is "$user_gaia" "$(( user_before + r1 + r2 ))"
    log "user1 Gaia balance $user_before -> $(gaia_balance "$user_gaia")"

    # 5. steady state over three more day epochs
    local submissions _round
    submissions=$(log_count "Submitting pending undelegation")
    for _round in 1 2 3; do
        wait_day_epoch
        assert_eq "pending undelegation submissions after day epoch $(day_epoch)" "$(log_count "Submitting pending undelegation")" "$submissions"
        assert_ledger_matches_chain "$ica"
    done
    log "RR before verify $rr_before, now $(host_zone_field .redemption_rate) (only normal accrual expected)"
    log "verify complete"
}

claimable_epochs() { # <receiver>
    stride_q records list-user-redemption-record | jq -r --arg c "$HOST_CHAIN_ID" --arg r "$1" \
        '.user_redemption_record[] | select(.host_zone_id == $c and .receiver == $r and (.claim_is_pending | not)) | .epoch_number'
}

phase_status() {
    banner "status @ $(date '+%H:%M:%S') day epoch $(day_epoch) stride epoch $(epoch_number stride_epoch)"
    local ica redemption
    ica=$(host_zone_field .delegation_ica_address)
    redemption=$(host_zone_field .redemption_ica_address)
    echo "host zone:"
    host_zone | jq -c '{total_delegations, redemption_rate, halted, delegation_ica_address, redemption_ica_address, unbonding_period}' | sed 's/^/    /'
    host_zone | jq -c '.validators[] | {name, address, weight, delegation, delegation_changes_in_progress}' | sed 's/^/    /'
    echo "unbonding records:"
    print_unbonding_records
    echo "deposit records:"
    print_deposit_records
    echo "ICA balances ($HOST_DENOM): delegation=$(gaia_balance "$ica") redemption=$(gaia_balance "$redemption")"
    echo "Gaia delegations of the delegation ICA (Σ $(gaia_delegated_total "$ica")):"
    gaia_delegations_json "$ica" | jq -r '.delegation_responses[] | "    \(.delegation.validator_address) \(.balance.amount)"'
    echo "ICA channels:"
    print_ica_channels
}

# =============================================================================================
# Main
# =============================================================================================
usage() {
    sed -n '3,20p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
    exit 1
}

preflight() {
    command -v jq >/dev/null || die "jq is required"
    command -v kubectl >/dev/null || die "kubectl is required"
    [[ -f $KEYS_FILE ]] || die "keys file not found: $KEYS_FILE"
    kube get pod "$STRIDE_POD" "$GAIA_POD" >/dev/null || die "pods not found in namespace $NAMESPACE (context: $(kubectl config current-context))"
}

main() {
    local phase=${1:-}
    [[ -n $phase ]] || usage
    preflight
    case $phase in
        setup) phase_setup ;;
        redeem) phase_redeem ;;
        drift) phase_drift ;;
        stuck) phase_stuck ;;
        measure) phase_measure ;;
        build-and-swap) phase_build_and_swap ;;
        upgrade) phase_upgrade ;;
        verify) phase_verify ;;
        status) phase_status ;;
        wait-day-epoch) wait_day_epoch "${2:-}" ;;
        *) usage ;;
    esac
}

main "$@"
