#!/bin/bash
# Phase 4: admin drain with injected failures (slash, ICA timeout, offset, dead window). Runs on v35.
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 4: admin drain"

ADMIN_MEMBERS=m1,m2,m3
DUST=2000000

no_flags() { # chain-id
  strided_new q stakeibc show-validators "$1" -o json \
    | jq -e '[.validators[] | (.delegation_changes_in_progress // 0 | tonumber)] | all(. == 0)'
}

# Writes a one-validator drain file into the stride pod
write_drain_file() { # path valoper offset
  echo "[{\"address\":\"$2\",\"offset\":\"$3\"}]" | $KX exec -i stride-validator-0 -c validator -- sh -c "cat > $1"
}

# Drain txs are only safe in the first ~130s of the day epoch; the ICA timeout is tied to the epoch end
submit_window() {
  local next now
  next=$(day_epoch_next_start); now=$(date +%s)
  if (( next - now < 50 )); then sleep_until $((next + 5)); fi
}

drain() { # chain-id file-or-flag
  ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc undelegate-from-validators "$1" "$2" >/dev/null
}

validator_field_is() { # chain-id valoper jq-predicate
  strided_new q stakeibc show-validators "$1" -o json | jq -e --arg addr "$2" ".validators[] | select(.address == \$addr) | $3"
}

signal_gaiad() { # pod STOP|CONT
  $KX exec "$1" -c validator -- sh -c "pid=\$(pgrep -x gaiad || ps -o pid,comm | awk '\$2==\"gaiad\"{print \$1}' | head -1); kill -$2 \$pid"
}

cleanup() { signal_gaiad cosmoshub-validator-6 CONT || true; $KX scale deployment relayer-stride-osmosis --replicas=1 || true; }
trap cleanup EXIT

hub_val_jailed() { gaiad q staking validator "$1" -o json | jq -e '.validator.jailed == true'; }

# The ICA controller channel for an owner, in a given state (--limit: the default page hides later channels)
ica_channel_in_state() { # owner state
  strided_new q ibc channel channels --limit 1000 -o json \
    | jq -e --arg port "icacontroller-$1" --arg state "$2" '[.channels[] | select(.port_id | startswith($port)) | .state] | index($state) != null'
}

stuatom_supply_unchanged() {
  [[ "$(strided_new q bank total-supply-of stuatom -o json | jq -r '.amount.amount')" == "$STATOM_SUPPLY" ]]
}

# Slash was applied on host but not on Stride: the slashed validator keeps a delegation above 1,000,000, the rest are dust
val7_failed_others_drained() { # valoper
  strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -e --arg bad "$1" '
    .validators as $vals
    | ([$vals[] | select(.address != $bad) | (.delegation | tonumber)] | max) < 1000000
      and ([$vals[] | select(.address == $bad) | (.delegation | tonumber)] | .[0]) > 1000000'
}

dead_window_send_fails() {
  local out; out=$(ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc undelegate-from-validators cosmoshub-test-1 --all 2>&1) || true
  grep -qiE 'timeout|code=[1-9]' <<<"$out"
}

zone_total_at_dust() { # chain-id
  strided_new q stakeibc show-host-zone "$1" -o json | jq -e --argjson dust "$DUST" '(.host_zone.total_delegations | tonumber) < $dust'
}

wait_until 300 "hub flags clear"  no_flags cosmoshub-test-1
wait_until 300 "osmo flags clear" no_flags osmosis-test-1

# Baseline for the "nothing burned" checkpoint; replace any value left by an earlier run
STATOM_SUPPLY=$(strided_new q bank total-supply-of stuatom -o json | jq -r '.amount.amount')
{ grep -v '^STATOM_SUPPLY=' "$REHEARSAL_DIR/state.env" || true; echo "STATOM_SUPPLY=$STATOM_SUPPLY"; } > "$REHEARSAL_DIR/state.env.tmp"
mv "$REHEARSAL_DIR/state.env.tmp" "$REHEARSAL_DIR/state.env"
log "STATOM_SUPPLY=$STATOM_SUPPLY"

# Live test: one hub validator drained in full by file (the last one; val7 is used for the slash injection below)
HUB_VALS=$(strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -r '.validators[].address')
SMALL=$(tail -1 <<<"$HUB_VALS")
write_drain_file /tmp/one.json "$SMALL" 0
submit_window; drain cosmoshub-test-1 /tmp/one.json
wait_until 240 "live-test ack" validator_field_is cosmoshub-test-1 "$SMALL" '.delegation == "0" and (.delegation_changes_in_progress // 0 | tonumber) == 0'
checkpoint "nothing burned" stuatom_supply_unchanged

# Injection 1: slash hub val7 after the refresh, then --all: val7's batch fails, the others succeed
HUB_VAL7=$(sed -n 7p <<<"$HUB_VALS")
# The pod index must map to HUB_VAL7, or the STOP would slash a different validator
HUB_VAL7_MONIKER=$(gaiad q staking validator "$HUB_VAL7" -o json | jq -r .validator.description.moniker)
if [[ "$HUB_VAL7_MONIKER" != "val7" ]]; then
  log "FAIL: $HUB_VAL7 has moniker '$HUB_VAL7_MONIKER', expected val7 (cosmoshub-validator-6)"
  exit 1
fi
signal_gaiad cosmoshub-validator-6 STOP
wait_until 240 "hub val7 jailed" hub_val_jailed "$HUB_VAL7"
signal_gaiad cosmoshub-validator-6 CONT
submit_window; drain cosmoshub-test-1 --all
wait_until 240 "drain acks" no_flags cosmoshub-test-1
log_cmd "hub validators after --all" strided_new q stakeibc show-validators cosmoshub-test-1 -o json
checkpoint "val7 batch failed, others drained" val7_failed_others_drained "$HUB_VAL7"

# Refresh applies the slash, then val7 drains by file
ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc update-delegation cosmoshub-test-1 "$HUB_VAL7" >/dev/null
wait_until 120 "val7 refreshed" validator_field_is cosmoshub-test-1 "$HUB_VAL7" '.slash_query_in_progress == false'
write_drain_file /tmp/v7.json "$HUB_VAL7" 0
submit_window; drain cosmoshub-test-1 /tmp/v7.json
wait_until 240 "val7 drained" validator_field_is cosmoshub-test-1 "$HUB_VAL7" '(.delegation | tonumber) < 1000000'

# Injection 2 (osmosis zone): ICA timeout -> channel closes -> restore -> resubmit. Pause the stride-osmosis relayer past the day-epoch timeout.
OSMO_VALS=$(strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -r '.validators[].address')
OV1=$(head -1 <<<"$OSMO_VALS")
write_drain_file /tmp/ov1.json "$OV1" 0
$KX scale deployment relayer-stride-osmosis --replicas=0
submit_window; drain osmosis-test-1 /tmp/ov1.json
sleep_until $(( $(day_epoch_next_start) + 20 ))
$KX scale deployment relayer-stride-osmosis --replicas=1
# Assumes no stale closed osmosis DELEGATION channel exists from an earlier run, else STATE_CLOSED matches immediately
wait_until 300 "osmo delegation channel closed" ica_channel_in_state osmosis-test-1.DELEGATION STATE_CLOSED
ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc restore-interchain-account osmosis-test-1 connection-1 osmosis-test-1.DELEGATION >/dev/null
wait_until 300 "osmo delegation channel reopened" ica_channel_in_state osmosis-test-1.DELEGATION STATE_OPEN
checkpoint "flags reset by restore" no_flags osmosis-test-1

# Lost-ack check (manual, orchestrator): did the closed channel leave an executed batch unacknowledged? Compare with Stride's record.
OSMO_DELEGATION_ICA=$(strided_new q stakeibc show-host-zone osmosis-test-1 -o json | jq -r '.host_zone.delegation_ica_address')
log_cmd "osmo host-side delegations of the delegation ICA after restore" osmosisd q staking delegations "$OSMO_DELEGATION_ICA" -o json
log_cmd "osmo validators after restore" strided_new q stakeibc show-validators osmosis-test-1 -o json

# Injection 3: offset drain of osmo val2, then --all for the rest; the offset stays recorded and is calibrated away
OV2=$(sed -n 2p <<<"$OSMO_VALS")
write_drain_file /tmp/ov2.json "$OV2" 1000000
submit_window; drain osmosis-test-1 /tmp/ov2.json
wait_until 240 "offset ack" validator_field_is osmosis-test-1 "$OV2" '.delegation == "1000000" and (.delegation_changes_in_progress // 0 | tonumber) == 0'
submit_window; drain osmosis-test-1 --all
wait_until 240 "osmo drained" no_flags osmosis-test-1

# Injection 4: the dead window: submit in the last fifth of the day epoch and expect a failed send, nothing flagged
sleep_until $(( $(day_epoch_next_start) - 25 ))
CHECKPOINT_SOFT=1 checkpoint "dead-window send fails" dead_window_send_fails
wait_until 300 "no flags after dead window" no_flags cosmoshub-test-1
checkpoint "hub total at dust"  zone_total_at_dust cosmoshub-test-1
checkpoint "osmo total at dust" zone_total_at_dust osmosis-test-1
checkpoint "hub rate frozen"  assert_rate_unchanged cosmoshub-test-1 "$RATE_HUB"
checkpoint "osmo rate frozen" assert_rate_unchanged osmosis-test-1 "$RATE_OSMO"
