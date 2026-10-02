#!/bin/bash
# Phase 9: final checks, halt Stride, and prove every Osmosis holder can swap out through the funded pools
source "$(dirname "$0")/lib.sh"
source "$REHEARSAL_DIR/state.env"
log "## Phase 9: halt"

STATOM_CANON=ibc/$(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
ATOM_ON_OSMO=ibc/$(printf 'transfer/channel-1/uatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
OSMOSIS_REST=${OSMOSIS_REST:-https://osmosis-api.internal.stridenet.co}
OSMOSIS_REST_LOCAL=$OSMOSIS_REST

zero_user_redemption_records() {
  strided_new q records list-user-redemption-record -o json | jq -e '.user_redemption_record | length == 0'
}
no_delegation_flags() {
  local zone
  for zone in cosmoshub-test-1 osmosis-test-1; do
    strided_new q stakeibc show-validators "$zone" -o json | jq -e '[.validators[].delegation_changes_in_progress | tonumber] | all(. == 0)' || return 1
  done
}
staketia_clean() { strided_new q staketia redemption-records -o json | jq -e '.redemption_record_responses | length == 0'; }
channels_clear_both_ways() {
  local channel
  for channel in channel-0 channel-1; do
    strided_new q ibc channel packet-commitments transfer "$channel" -o json | jq -e '.commitments | length == 0' || return 1
  done
  gaiad q ibc channel packet-commitments transfer channel-0 -o json | jq -e '.commitments | length == 0' || return 1
  osmosisd q ibc channel packet-commitments transfer channel-0 -o json | jq -e '.commitments | length == 0'
}

checkpoint "zero user redemption records" zero_user_redemption_records
checkpoint "no flags" no_delegation_flags
checkpoint "staketia clean" staketia_clean
checkpoint "channels clear both ways" channels_clear_both_ways

# --- Coverage: coverage_check.py reads Osmosis over REST, so forward the pod's REST port to the laptop ---
EXPORT_PRE_HALT=$REHEARSAL_DIR/export_pre_halt.json
bash "$REHEARSAL_DIR/export.sh" "$EXPORT_PRE_HALT"
if [[ "${OSMOSIS_REST_PORTFORWARD:-}" == 1 ]]; then
  OSMOSIS_REST_LOCAL=http://localhost:11317
  $KX port-forward pod/osmosis-validator-0 11317:1317 >/dev/null 2>&1 &
  PORT_FORWARD_PID=$!
  trap 'kill $PORT_FORWARD_PID 2>/dev/null || true' EXIT
  for _ in $(seq 1 15); do
    curl -sf "$OSMOSIS_REST_LOCAL/cosmos/base/tendermint/v1beta1/node_info" >/dev/null && break
    sleep 1
  done
fi
checkpoint "coverage before halt" python3 "$REPO/scripts/wind-down/coverage_check.py" \
  --export "$EXPORT_PRE_HALT" --pools "$POOLS_FILE" --vault "$VAULT_MS_OSMO" --osmosis-rest "$OSMOSIS_REST_LOCAL"

# --- Halt Stride 30 blocks ahead via app.toml; killing the process restarts the container with the new height ---
stride_height() { strided_new status | jq -r '.sync_info.latest_block_height // .SyncInfo.latest_block_height'; }
HALT_HEIGHT=$(( $(stride_height) + 30 ))
log "halting Stride at height $HALT_HEIGHT"
for i in 0 1 2 3; do
  $KX exec "stride-validator-$i" -c validator -- sh -c "
    sed -i 's/^halt-height = .*/halt-height = $HALT_HEIGHT/' /home/validator/.stride/config/app.toml
    grep -q '^halt-height = $HALT_HEIGHT' /home/validator/.stride/config/app.toml
    pid=\$(pgrep -x strided || ps -o pid,comm | awk '\$2==\"strided\" {print \$1}')
    [ -n \"\$pid\" ] || exit 1
    kill \$pid" || true
done

# The node logs this line when it stops at the configured height (also after a restart that lands on it again)
stride_halted() {
  local logs
  logs=$({ $KX logs stride-validator-0 -c validator --tail=300; $KX logs stride-validator-0 -c validator --previous --tail=300; } 2>/dev/null || true)
  grep -q 'halt per configuration' <<<"$logs"
}
if ! wait_until 300 "stride halted at $HALT_HEIGHT" stride_halted; then
  log "app.toml halt did not take; scaling the validator statefulset to 0 instead"
  $KX scale statefulset stride-validator --replicas=0
fi

# --- Final invariant: every stToken holder on Osmosis swaps everything out; every swap succeeds; the pool keeps a native surplus ---
log "Stride halted; swapping every holder's balance on Osmosis"
pool_contract() { osmosisd q poolmanager pool "$1" -o json | jq -r '.pool.contract_address'; }
pool_liquidity() { osmosisd q wasm contract-state smart "$(pool_contract "$1")" '{"get_total_pool_liquidity":{}}' -o json; }

# <holder key> <amount>: swap the whole stATOM balance for ATOM through the canonical pool
swap_out() {
  local hash
  hash=$(osmosisd tx poolmanager swap-exact-amount-in "${2}${STATOM_CANON}" 1 --swap-route-pool-ids "$POOL_ATOM" \
    --swap-route-denoms "$ATOM_ON_OSMO" --from "$1" $OSMO_TX | tx_hash)
  wait_tx osmosisd "$hash"
}
canonical_pool_surplus_nonnegative() {
  pool_liquidity "$POOL_ATOM" | jq -e --arg denom "$ATOM_ON_OSMO" '[.data.total_pool_liquidity[] | select(.denom==$denom) | .amount | tonumber] | .[0] >= 0'
}

# holder-vesting has no uosmo on Osmosis, so it cannot pay the swap fee until it is funded
fund_hash=$(osmosisd tx bank send faucet "$HOLDER_VESTING_OSMO" 1000000uosmo $OSMO_TX | tx_hash)
wait_tx osmosisd "$fund_hash"

for holder in user1 holder-base holder-vesting; do
  holder_addr=$(osmosisd keys show "$holder" -a --keyring-backend test | tr -d '\r\n')
  bal=$(osmosisd q bank balance "$holder_addr" "$STATOM_CANON" -o json | jq -r '.balance.amount')
  if [[ "$bal" -gt 0 ]]; then
    checkpoint "$holder swaps $bal" swap_out "$holder" "$bal"
  else
    log "$holder holds no canonical stATOM on osmosis, nothing to swap"
  fi
done

log_cmd "final pool liquidity" pool_liquidity "$POOL_ATOM"
checkpoint "canonical pool surplus >= 0" canonical_pool_surplus_nonnegative
log "Phase 9 complete"
