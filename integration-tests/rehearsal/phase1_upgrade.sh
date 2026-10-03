#!/bin/bash
# Phase 1: schedule and apply the v35 upgrade, then verify the handler's effects (spec design doc, phase 1).
# Starts on v34 (strided_old); from the "v35 running" wait on, everything uses strided_new.
source "$(dirname "$0")/lib.sh"
source "$REHEARSAL_DIR/state.env"
log "## Phase 1: upgrade to v35"

MIN_UPGRADE_LEAD_BLOCKS=60
ADMIN_MEMBERS=m1,m2,m3
GOV_MODULE_ADDRESS=stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl
AUTOPILOT_LIQUID_STAKE_MEMO='{"autopilot":{"receiver":"'$USER1_STRIDE'","stakeibc":{"action":"LiquidStake"}}}'
LIQUID_STAKE_ICA_MSG_TYPE=/stride.stakeibc.MsgLiquidStake

v35_applied() { strided_new q upgrade applied v35 -o json | jq -e '(.height | tonumber) > 0'; }
handler_log_lines() { $KX logs stride-validator-0 -c validator --since=60m | grep -E 'v35|wind-down|Upgrade v35' || true; }
# CometBFT puts the level before the message ("ERR v35: ..."); expected skip-path lines are allow-listed. The v34 binary
# itself logs 'ERR UPGRADE "v35" NEEDED at height' and CONSENSUS FAILURE lines right before cosmovisor swaps binaries.
handler_log_errors() {
  local matched
  matched=$($KX logs stride-validator-0 -c validator --since=60m | grep -E 'ERR .*v35|v35.*panic' | grep -vE 'haqq|comdex|trade route|LSM|oracle|contract|UPGRADE "v35" NEEDED|CONSENSUS FAILURE' || true)
  log "handler error lines: ${matched:-none}"
  [[ -z "$matched" ]]
}
hub_deposit_in_transfer_queue() {
  strided_old q records list-deposit-record -o json \
    | jq -e '[.deposit_record[]? | select(.host_zone_id == "cosmoshub-test-1" and .status == "TRANSFER_QUEUE")] | length > 0'
}
# The drain guard (x/stakeibc/keeper/wind_down_undelegate.go checkNoQueuedUnbondings) must refuse while the Hub RD record is queued
hub_drain_refused() {
  local out
  # Only probe while RD is still queued and the day epoch (D5) is >15s away: once D5 submits RD and
  # its ack lands, the guard passes and this would be a REAL drain of the Hub zone
  (( $(date +%s) < $(day_epoch_next_start) - 15 )) || { log "hub drain-refusal probe skipped: D5 too close"; return 1; }
  strided_new q records list-epoch-unbonding-record -o json | jq -e '[.epoch_unbonding_record[].host_zone_unbondings[]? | select(.host_zone_id=="cosmoshub-test-1" and .status=="UNBONDING_QUEUE")] | length > 0' >/dev/null \
    || { log "hub drain-refusal probe skipped: RD no longer queued"; return 1; }
  out=$(ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc undelegate-from-validators cosmoshub-test-1 --all 2>&1) && return 1
  grep -qiE 'wait for the day epoch|let the day epoch|awaiting an ack|pending undelegation|UNBONDING_QUEUE|UNBONDING_RETRY_QUEUE' <<<"$out"
}
liquid_stake_unroutable() {
  # Capture first: `wrapper | grep -q` under pipefail fails on SIGPIPE when grep exits early
  local out; out=$(strided_new tx stakeibc liquid-stake 1000 uatom --from user1 $STRIDE_TX 2>&1 || true)
  grep -qiE "can't route|unknown command|not found" <<<"$out"
}
autopilot_stakeibc_off() { strided_new q autopilot params -o json | jq -e '(.params.stakeibc_active // false) == false'; }
rate_limits_removed() { strided_new q ratelimiting list-rate-limits -o json | jq -e '(.rate_limits // []) | length == 0'; }
wasm_upload_gov_only() {
  strided_new q wasm params -o json | jq -e --arg gov "$GOV_MODULE_ADDRESS" \
    '(.params // .) | .code_upload_access | .permission == "AnyOfAddresses" and .addresses == [$gov]'
}
ica_host_allow_list_trimmed() {
  strided_new q interchain-accounts host params -o json \
    | jq -e --arg msg "$LIQUID_STAKE_ICA_MSG_TYPE" '(.params.allow_messages // []) | index($msg) == null'
}
historical_tx_decodes() { strided_new q tx "$HIST_TX" -o json | jq -e '.txhash'; }
stride_balance() { # address denom -> amount ("0" when absent)
  strided_new q bank balances "$1" -o json | jq -r --arg denom "$2" '([.balances[] | select(.denom == $denom) | .amount][0]) // "0"'
}

if [[ "${PHASE1_RESUME:-0}" != 1 ]]; then
############################################
# Schedule and apply the upgrade
############################################

# U was fixed by the seed; aim the upgrade height at it (1s blocks). The proposal needs ~60s to pass, and the upgrade
# must land before D5 (U + 30s) so the queued Hub redemption is still pending: never push the height later than U.
height=$(strided_old status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height')
now=$(date +%s)
lead_blocks=$(( U - now ))
if (( lead_blocks < MIN_UPGRADE_LEAD_BLOCKS )); then
  log "ABORT: only ${lead_blocks}s to U (need >= ${MIN_UPGRADE_LEAD_BLOCKS}s for the proposal to pass); pushing the height out would cross D5. Re-seed."
  exit 1
fi
UPGRADE_HEIGHT=$(( height + lead_blocks ))
log "upgrade height $UPGRADE_HEIGHT (now $height, target time $U)"
UPGRADE_HEIGHT=$UPGRADE_HEIGHT bash "$REPO/integration-tests/network/scripts/upgrade.sh" | tee -a "$LOG"

# Pre-upgrade liquid stake so a deposit record is still TRANSFER_QUEUE at U (delegation happens at the next
# stride epoch, 45s, so it must land just before U). The proposal has passed; the chain is still on v34.
sleep_until $((U - 20))
log_cmd "deposit in flight" strided_old tx stakeibc liquid-stake 10000000 uatom --from user1 $STRIDE_TX
# Timing-sensitive (the deposit record only stays TRANSFER_QUEUE for one stride epoch)
CHECKPOINT_SOFT=1 checkpoint "hub deposit in TRANSFER_QUEUE" hub_deposit_in_transfer_queue

# The chain halts at the height, cosmovisor swaps the binary, then the node serves v35 queries
fi
wait_until 600 "v35 running" v35_applied
# Self-heal (rehearsal finding): validators whose daemon restarted in-process can come back with no peers and stall
# consensus at the upgrade height; if the height does not advance within 90s, restart the stuck processes
height_now() { strided_new status 2>/dev/null | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height'; }
h0=$(height_now); sleep 90
if [[ "$(height_now)" == "$h0" ]]; then
  log "FINDING (harness): chain stalled at $h0 after the upgrade; restarting stride-validator-1..3 processes"
  for i in 1 2 3; do $KX exec stride-validator-$i -c validator -- sh -c 'kill $(ps -o pid,args | grep "bin/strided start" | grep -v grep | awk "{print \$1}")' || true; done
  wait_until 300 "chain advancing after restart" bash -c "[[ \$(strided_new status 2>/dev/null | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height') -gt $h0 ]]"
fi
log_cmd "handler log lines" handler_log_lines
checkpoint "no handler error" handler_log_errors

# Only valid until D5 (the day epoch after D4 submits the queued RD record), so it runs first
CHECKPOINT_SOFT=1 checkpoint "drain refused while RD queued (hub)" hub_drain_refused

############################################
# Entry points and state the handler changed
############################################

checkpoint "liquid-stake cannot route"    liquid_stake_unroutable
checkpoint "autopilot stakeibc off"       autopilot_stakeibc_off
checkpoint "rate limits removed"          rate_limits_removed
checkpoint "wasm upload gov-only"         wasm_upload_gov_only
checkpoint "ica host allow-list trimmed"  ica_host_allow_list_trimmed
checkpoint "historical tx decodes"        historical_tx_decodes
checkpoint "hub rate frozen"              assert_rate_unchanged cosmoshub-test-1 "$RATE_HUB"
checkpoint "osmo rate frozen"             assert_rate_unchanged osmosis-test-1 "$RATE_OSMO"
log_cmd "records after upgrade" strided_new q records list-epoch-unbonding-record -o json

############################################
# Routes that bypass the msg router (soft: they depend on relayers and fresh channels)
############################################

# Autopilot: an ICS-20 transfer from the Hub with a liquid-stake memo must be refused and refunded. Memo shape is
# x/autopilot/types/autopilot.go (RawPacketMetadata) + parser.go: autopilot.receiver + autopilot.stakeibc.action.
statom_before=$(stride_balance "$USER1_STRIDE" stuatom)
atom_before=$(stride_balance "$USER1_STRIDE" "$ATOM_ON_STRIDE")
log_cmd "autopilot liquid-stake memo from Hub" gaiad tx ibc-transfer transfer transfer channel-0 "$USER1_STRIDE" 1000000uatom \
  --memo "$AUTOPILOT_LIQUID_STAKE_MEMO" --from user1 $HUB_TX
sleep 60
autopilot_route_refused() {
  [[ "$(stride_balance "$USER1_STRIDE" stuatom)" == "$statom_before" && "$(stride_balance "$USER1_STRIDE" "$ATOM_ON_STRIDE")" == "$atom_before" ]]
}
CHECKPOINT_SOFT=1 checkpoint "autopilot route refused (no stATOM minted, ATOM refunded)" autopilot_route_refused

# ICA host: an interchain account on Stride (controlled from the Hub) sends MsgLiquidStake; it must not execute.
# generate-packet-data runs on the v34 binary, which is still in the pod and still knows the message type.
ica_address_on_stride() {
  gaiad q interchain-accounts controller interchain-account "$USER1_COSMOS" connection-0 -o json | jq -er '.address'
}
ica_opened() { ica_address_on_stride >/dev/null; }
log_cmd "register ICA on Stride from Hub" gaiad tx interchain-accounts controller register connection-0 --from user1 $HUB_TX || true
if wait_until 120 "hub-controlled ICA open on Stride" ica_opened; then
  ica_address=$(ica_address_on_stride)
  log_cmd "fund ICA" strided_new tx bank send user1 "$ica_address" "1000000$ATOM_ON_STRIDE" --from user1 $STRIDE_TX || true
  sleep 6
  ica_message='{"@type":"'$LIQUID_STAKE_ICA_MSG_TYPE'","creator":"'$ica_address'","amount":"1000000","host_denom":"uatom"}'
  ica_packet=$(strided_old tx interchain-accounts host generate-packet-data "$ica_message" 2>/dev/null || true)
  printf '%s\n' "$ica_packet" | $KX exec -i cosmoshub-validator-0 -c validator -- sh -c 'cat > /tmp/ica_packet.json'
  log_cmd "send ICA liquid-stake packet" gaiad tx interchain-accounts controller send-tx connection-0 /tmp/ica_packet.json --from user1 $HUB_TX || true
  sleep 60
  ica_statom_before=0
  ica_route_refused() {
    [[ "$(stride_balance "$ica_address" stuatom)" == "$ica_statom_before" && "$(stride_balance "$ica_address" "$ATOM_ON_STRIDE")" == 1000000 ]]
  }
  CHECKPOINT_SOFT=1 checkpoint "ICA host route refused (no stATOM minted, ATOM untouched)" ica_route_refused
else
  CHECKPOINT_SOFT=1 checkpoint "ICA host route (channel did not open; not testable)" false
fi

log "phase 1 done"
