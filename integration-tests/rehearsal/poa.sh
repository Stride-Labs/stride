#!/usr/bin/env bash
# v34 POA validator-swap rehearsal on the k8s integration network (REHEARSAL ONLY).
#
# The network runs five Stride nodes; only val1..val3 are in the genesis POA set (values.yaml:
# poaGenesisValidators). v34's SwapPoaValidators adds val4/val5 (constants.go IncomingValidators,
# pubkeys pasted in by `measure`) and removes val2/val3 (OutgoingMonikers). Phases:
#
#   measure          read val4/val5 consensus pubkeys from the pods and paste them into constants.go
#   build-and-swap   build the branch binary and copy it into cosmovisor/upgrades/v34 on every pod
#   upgrade          expedited gov proposal from the laptop's strided; assert the handler's log lines
#   verify           POA set, CometBFT validator set, voting powers, and who signs blocks after the swap
#   status           one-shot dump of the above
set -euo pipefail

INTEGRATION_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "$INTEGRATION_DIR/.." && pwd)"
KEYS_FILE="$INTEGRATION_DIR/network/configs/keys.json"
ADMINS_FILE="$REPO_ROOT/utils/admins.go"
CONSTANTS_FILE="$REPO_ROOT/app/upgrades/v34/constants.go"
MAINNET_ADMIN_ADDRESS=stride1k8c2m5cn322akk5wy8lpt87dd2f4yh9azg7jlh # build.sh swaps this for keys.json's admin

NAMESPACE=integration
STRIDE_CHAIN_ID=stride-test-1
STRIDE_GAS_PRICES=0.025ustrd
CHAIN_CONTAINER=validator
NUM_NODES=5
GENESIS_VALIDATORS="val1 val2 val3"
INCOMING="val4 val5"
OUTGOING="val2 val3"
EXPECTED_SET_AFTER="val1 val4 val5"
VALIDATOR_POWER=274523

UPGRADE_NAME=${UPGRADE_NAME:-v34}
DAEMON_HOME=/home/validator/.stride
UPGRADE_BINARY_PATH="${DAEMON_HOME}/cosmovisor/upgrades/${UPGRADE_NAME}/bin/strided"
CURRENT_BINARY_PATH="${DAEMON_HOME}/cosmovisor/current/bin/strided"
UPGRADE_TIMEOUT=600
SCRATCH="$(mktemp -d "${TMPDIR:-/tmp}/poa-rehearsal.XXXXXX")"

# =============================================================================================
# Logging / assertions
# =============================================================================================
log() { printf '[%s] %s\n' "$(date '+%H:%M:%S')" "$*"; }
die() { log "ERROR: $*" >&2; exit 1; }
banner() { printf '\n===== %s =====\n' "$*"; }
assert_eq() { [[ "$2" == "$3" ]] && log "PASS $1: $2" || die "FAIL $1: got '$2', expected '$3'"; }
is_int() { [[ ${1:-} =~ ^[0-9]+$ ]]; }

wait_for() { # <timeout-seconds> <description> <predicate...>
    local timeout=$1 desc=$2 elapsed=0
    shift 2
    log "waiting up to ${timeout}s for: $desc"
    until "$@"; do
        (( elapsed += 5 ))
        (( elapsed < timeout )) || die "timed out after ${timeout}s waiting for: $desc"
        sleep 5
    done
    log "ok: $desc"
}

# =============================================================================================
# Chain helpers
# =============================================================================================
kube() { kubectl -n "$NAMESPACE" "$@"; }
pod() { echo "stride-validator-$(( $1 - 1 ))"; } # val index (1-based) -> pod name
val_name() { echo "val$1"; }

# cosmovisor's `current` points at the upgrade binary after the switch; prefer it for CLI calls
STRIDE_BINARY_SHIM="bin=$CURRENT_BINARY_PATH; [ -x \"\$bin\" ] || bin=strided; exec \"\$bin\" \"\$@\""
strided_on() { local p=$1; shift; kube exec "$p" -c "$CHAIN_CONTAINER" -- sh -c "$STRIDE_BINARY_SHIM" strided "$@"; }
rpc_on() { kube exec "$1" -c "$CHAIN_CONTAINER" -- curl -s "localhost:26657/$2"; } # <pod> <path>
stride_q() { strided_on "$(pod 1)" q "$@" -o json 2>/dev/null; }

key_mnemonic() { jq -r --arg n "$1" '.validators[] | select(.name == $n) | .mnemonic' "$KEYS_FILE"; }

block_height() { rpc_on "$(pod 1)" status | jq -r '.result.sync_info.latest_block_height'; }
node_cons_address() { rpc_on "$(pod "$1")" status | jq -r '.result.validator_info.address'; } # hex
node_voting_power() { rpc_on "$(pod "$1")" status | jq -r '.result.validator_info.voting_power'; }
node_catching_up() { rpc_on "$(pod "$1")" status | jq -r '.result.sync_info.catching_up'; }
node_cons_pubkey() { strided_on "$(pod "$1")" tendermint show-validator 2>/dev/null | jq -r '.key'; }
comet_validator_count() { rpc_on "$(pod 1)" validators | jq -r '.result.total'; }
commit_signers() { rpc_on "$(pod 1)" "commit?height=$1" | jq -r '.result.signed_header.commit.signatures[] | select(.block_id_flag == 2) | .validator_address'; }

poa_validators() { stride_q poa validators | jq '.validators'; }
# removal is an update to power 0, so the entry may linger in the store: "active" = power > 0
# (a zero power is omitted from the JSON entirely)
poa_monikers() { poa_validators | jq -r '[.[] | select(((.power // "0") | tonumber) > 0) | .metadata.moniker] | sort | join(" ")'; }
poa_power_of() { poa_validators | jq -r --arg m "$1" '[.[] | select(.metadata.moniker == $m) | (.power // "0")][0] // "0"'; }

stride_logs() { kube logs "$(pod 1)" -c "$CHAIN_CONTAINER" "$@"; }
log_count() { stride_logs 2>/dev/null | grep -c -- "$1" || true; } # grep -c: no SIGPIPE under pipefail
log_contains() { [[ $(log_count "$1") -gt 0 ]]; }

blocks_advancing() {
    local before after
    before=$(block_height); sleep 10; after=$(block_height)
    is_int "$before" && is_int "$after" && (( after > before ))
}

# =============================================================================================
# Phases
# =============================================================================================
phase_status() {
    banner "status @ $(date '+%H:%M:%S') height $(block_height)"
    log "POA set: $(poa_monikers)"
    poa_validators | jq -c '.[] | {moniker: .metadata.moniker, power}' | sed 's/^/    /'
    log "CometBFT validator set size: $(comet_validator_count)"
    local i
    for i in $(seq 1 $NUM_NODES); do
        log "$(val_name "$i") ($(pod "$i")): cons=$(node_cons_address "$i") power=$(node_voting_power "$i") catching_up=$(node_catching_up "$i")"
    done
    local h; h=$(block_height)
    log "signers of block $h: $(commit_signers "$h" | tr '\n' ' ')"
}

phase_measure() {
    banner "measure: val4/val5 consensus pubkeys -> $CONSTANTS_FILE"
    assert_eq "genesis POA set" "$(poa_monikers)" "$GENESIS_VALIDATORS"
    local name key i
    for name in $INCOMING; do
        i=${name#val}
        key=$(node_cons_pubkey "$i")
        [[ $key =~ ^[A-Za-z0-9+/]{43}=$ ]] || die "$name: unexpected pubkey '$key'"
        assert_eq "$name voting power before the upgrade" "$(node_voting_power "$i")" 0
        # (macOS ships bash 3.2: no ${name^^})
        sed -i '' "s|MEASURE_ME_$(tr a-z A-Z <<<"$name")|$key|; s|\"$name\", ConsPubKeyBase64: \"[A-Za-z0-9+/=]*\"|\"$name\", ConsPubKeyBase64: \"$key\"|" "$CONSTANTS_FILE"
        log "$name pubkey $key"
    done
    grep -n "ConsPubKeyBase64" "$CONSTANTS_FILE" | sed 's/^/    /'
    grep -q MEASURE_ME "$CONSTANTS_FILE" && die "placeholders left in $CONSTANTS_FILE"
    (cd "$REPO_ROOT" && go build ./app/... ) && log "constants compile; commit them before build-and-swap"
}

phase_build_and_swap() {
    banner "build-and-swap: $UPGRADE_NAME binary into cosmovisor/upgrades/$UPGRADE_NAME on all Stride pods"
    grep -q MEASURE_ME "$CONSTANTS_FILE" && die "run measure first"
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

    local i p
    for i in $(seq 1 $NUM_NODES); do
        p=$(pod "$i")
        kube exec "$p" -c "$CHAIN_CONTAINER" -- mkdir -p "$(dirname "$UPGRADE_BINARY_PATH")"
        kube cp "$SCRATCH/strided" "$p:$UPGRADE_BINARY_PATH" -c "$CHAIN_CONTAINER"
        kube exec "$p" -c "$CHAIN_CONTAINER" -- chmod +x "$UPGRADE_BINARY_PATH"
        log "$p: $UPGRADE_BINARY_PATH version = $(kube exec "$p" -c "$CHAIN_CONTAINER" -- "$UPGRADE_BINARY_PATH" version 2>&1 | tail -1)"
    done
    log "swap complete; cosmovisor only reads the upgrade binary at the upgrade height"
}

# Submit and vote from the laptop's strided (0.2s per call against the RPC ingress) — in-pod CLI
# calls are too slow to land three votes inside the 29s expedited window
STRIDE_RPC=${STRIDE_RPC:-http://stride-rpc.internal.stridenet.co:80}
LOCAL_STRIDED_HOME="$SCRATCH/strided-home"
UPGRADE_BUFFER_BLOCKS=60
GOV_AUTHORITY=stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
lstrided() { strided --home "$LOCAL_STRIDED_HOME" --node "$STRIDE_RPC" "$@"; }
import_local_keys() {
    local name
    for name in $GENESIS_VALIDATORS; do
        strided --home "$LOCAL_STRIDED_HOME" keys show "$name" --keyring-backend test >/dev/null 2>&1 && continue
        key_mnemonic "$name" | strided --home "$LOCAL_STRIDED_HOME" keys add "$name" --recover --keyring-backend test >/dev/null
    done
}
proposal_decided() { [[ $(lstrided q gov proposal "$1" -o json 2>/dev/null | jq -r '.proposal.status') =~ PASSED|REJECTED|FAILED ]]; }
submit_upgrade_proposal_local() {
    import_local_keys
    local height id status name tx="--keyring-backend test --chain-id $STRIDE_CHAIN_ID --gas-prices $STRIDE_GAS_PRICES -y -o json"
    height=$(( $(lstrided status 2>/dev/null | jq -r '.sync_info.latest_block_height') + UPGRADE_BUFFER_BLOCKS ))
    printf '{"messages":[{"@type":"/cosmos.upgrade.v1beta1.MsgSoftwareUpgrade","authority":"%s","plan":{"name":"%s","height":"%s"}}],"deposit":"2000000000ustrd","title":"Upgrade %s","summary":"Upgrade %s","expedited":true}\n' \
        "$GOV_AUTHORITY" "$UPGRADE_NAME" "$height" "$UPGRADE_NAME" "$UPGRADE_NAME" > "$SCRATCH/upgrade-proposal.json"
    log "submitting $UPGRADE_NAME upgrade proposal at height $height"
    lstrided tx gov submit-proposal "$SCRATCH/upgrade-proposal.json" --from val1 --gas 400000 $tx 2>/dev/null | jq -r '"submit code=\(.code)"'
    sleep 3
    id=$(lstrided q gov proposals -o json 2>/dev/null | jq -r '[.proposals[].id | tonumber] | max')
    for name in $GENESIS_VALIDATORS; do
        lstrided tx gov vote "$id" yes --from "$name" --gas 300000 $tx 2>/dev/null | jq -r --arg n "$name" '"vote \($n) code=\(.code)"'
        sleep 1
    done
    wait_for 90 "proposal $id decided" proposal_decided "$id"
    status=$(lstrided q gov proposal "$id" -o json 2>/dev/null | jq -r '.proposal.status')
    [[ $status == PROPOSAL_STATUS_PASSED ]] || die "proposal $id ended $status"
    log "proposal $id PASSED; upgrade $UPGRADE_NAME scheduled at height $height"
}

phase_upgrade() {
    banner "upgrade: gov proposal for $UPGRADE_NAME"
    local i
    for i in $(seq 1 $NUM_NODES); do
        kube exec "$(pod "$i")" -c "$CHAIN_CONTAINER" -- test -x "$UPGRADE_BINARY_PATH" || die "$(pod "$i"): $UPGRADE_BINARY_PATH missing, run build-and-swap"
    done
    log "pre-upgrade POA set: $(poa_monikers); comet validators: $(comet_validator_count); height $(block_height)"
    submit_upgrade_proposal_local

    wait_for "$UPGRADE_TIMEOUT" "'Upgrade $UPGRADE_NAME complete' in $(pod 1) logs" log_contains "Upgrade $UPGRADE_NAME complete"
    stride_logs | grep -E "Starting upgrade $UPGRADE_NAME|v34: |Upgrade $UPGRADE_NAME complete" | sed 's/\x1b\[[0-9;]*m//g' | cut -c1-160 | sed 's/^/    /'
    local name
    for name in $INCOMING; do log_contains "v34: adding POA validator $name" || die "missing 'adding POA validator $name'"; done
    for name in $OUTGOING; do log_contains "v34: removing POA validator $name" || die "missing 'removing POA validator $name'"; done
    log "PASS handler logged add val4/val5 and remove val2/val3"
    wait_for 120 "blocks advancing after the upgrade" blocks_advancing
}

phase_verify() {
    banner "verify: validator set after the swap"
    wait_for 60 "blocks advancing" blocks_advancing
    assert_eq "POA set monikers" "$(poa_monikers)" "$EXPECTED_SET_AFTER"
    local name i
    for name in $EXPECTED_SET_AFTER; do assert_eq "$name POA power" "$(poa_power_of "$name")" "$VALIDATOR_POWER"; done
    for name in $OUTGOING; do assert_eq "$name POA power" "$(poa_power_of "$name")" 0; done

    # ABCI validator updates land one block after the upgrade block; give CometBFT a few blocks
    sleep 15
    assert_eq "CometBFT validator set size" "$(comet_validator_count)" 3
    for name in $EXPECTED_SET_AFTER; do i=${name#val}; assert_eq "$name comet voting power" "$(node_voting_power "$i")" "$VALIDATOR_POWER"; done
    for name in $OUTGOING; do i=${name#val}; assert_eq "$name comet voting power" "$(node_voting_power "$i")" 0; done
    for i in $(seq 1 $NUM_NODES); do assert_eq "$(val_name "$i") still synced" "$(node_catching_up "$i")" false; done

    # Who actually signs: every expected validator commits, neither outgoing one does, over 5 blocks
    local h signers
    h=$(block_height)
    for _ in 1 2 3 4 5; do
        signers=$(commit_signers "$h")
        for name in $EXPECTED_SET_AFTER; do i=${name#val}; grep -q "$(node_cons_address "$i")" <<<"$signers" || die "block $h not signed by $name"; done
        for name in $OUTGOING; do i=${name#val}; grep -q "$(node_cons_address "$i")" <<<"$signers" && die "block $h signed by outgoing $name"; done
        log "PASS block $h signed by exactly $(wc -l <<<"$signers" | tr -d ' ') validators: $EXPECTED_SET_AFTER"
        sleep 6; h=$(block_height)
    done
    log "gov voting period now: $(stride_q gov params | jq -r '.params.voting_period') (v34 UpdateGovParams)"
    log "verify complete"
}

usage() { echo "usage: $0 <measure|build-and-swap|upgrade|verify|status>"; exit 1; }
main() {
    local phase=${1:-}
    [[ -n $phase ]] || usage
    [[ $(kubectl config current-context) == integration ]] || die "kube context is $(kubectl config current-context), expected integration"
    case $phase in
        measure) phase_measure ;;
        build-and-swap) phase_build_and_swap ;;
        upgrade) phase_upgrade ;;
        verify) phase_verify ;;
        status) phase_status ;;
        *) usage ;;
    esac
}
main "$@"
