#!/bin/bash
# Phase 8: sweep every holder's tokens off Stride (runs on v35, after gov is closed and everything is drained)
source "$(dirname "$0")/lib.sh"
source "$REHEARSAL_DIR/state.env"
log "## Phase 8: sweep off Stride"

BATCH_DIR=$REHEARSAL_DIR/sweep-batches
PRICES_FILE=$REHEARSAL_DIR/sweep-prices.json
SWEEP_DENOMS="stuatom,stuosmo,ustrd,$ATOM_ON_STRIDE"
STATOM_ON_OSMO=ibc/$(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
HOLDER_BASE_COSMOS=$(python3 -c "
import sys
sys.path.insert(0, '$REPO/scripts/wind-down')
import bech32_ref
print(bech32_ref.encode('cosmos', bech32_ref.decode('$HOLDER_BASE')[1]))")

# <local addresses-file> <denoms> <key>: copy the file into the pod and submit one sweep tx; prints the tx hash
send_sweep() {
  $KX cp "$1" stride-validator-0:/tmp/sweep-batch.txt -c validator >/dev/null
  # gas auto underestimates a sweep whose addresses are all skipped (reads only); 1.5M covers the measured 605k batch
  strided_new tx stakeibc sweep-tokens-off-stride "$2" /tmp/sweep-batch.txt --from "$3" --keyring-backend test --chain-id stride-test-1 --gas 1500000 --gas-prices 1ustrd -y -o json | tx_hash
}

# <hash>: how many addresses the sweep skipped (one sweep_skipped event each; the tx has no num_skipped attribute)
skipped_count() { strided_new q tx "$1" -o json | jq '[.events[] | select(.type=="sweep_skipped")] | length'; }
skipped_reasons() {
  strided_new q tx "$1" -o json | jq -r '[.events[] | select(.type=="sweep_skipped") | .attributes[] | select(.key=="reason") | .value] | sort | join("|")'
}
batch_skipped_nothing() { [[ $(skipped_count "$1") == 0 ]]; }
skipped_two_with_reasons() { [[ $(skipped_reasons "$1") == "blocked module address|transfer escrow address" ]]; }

# The admin key is not the sweep operator: the CLI's ValidateBasic (or the chain) must refuse it
admin_cannot_sweep() {
  local out
  out=$(send_sweep "$1" stuatom admin 2>&1 || true)
  grep -qi 'not the sweep operator\|unauthorized' <<<"$out"
}

holder_has_stride_denom() { [[ $(strided_new q bank balance "$1" "$2" -o json | jq -r '.balance.amount') -gt 0 ]]; }
osmo_balance_positive() { [[ $(osmosisd q bank balance "$1" "$2" -o json | jq -r '.balance.amount') -gt 0 ]]; }
hub_balance_positive() { [[ $(gaiad q bank balance "$1" uatom -o json | jq -r '.balance.amount') -gt 0 ]]; }
holder_base_has_none_left() {
  strided_new q bank balances "$HOLDER_BASE" -o json | jq -e '[.balances[] | select(.denom=="stuatom" or .denom=="ustrd")] | length == 0'
}
log "## Phase 8 (resume at the hand-built skip batch; the builder batch already swept 4 holders, 605223 gas, 0 skipped)"
# --- Hand-built batch: the distribution module and a transfer escrow are skipped, each with its reason ---
DIST=$(strided_new q auth module-account distribution -o json | jq -r '.account.value.address // .account.base_account.address')
ESCROW=$(strided_new q ibc-transfer escrow-address transfer channel-0)
printf '%s\n%s\n%s\n' "$DIST" "$ESCROW" "$USER1_STRIDE" > "$REHEARSAL_DIR/sweep-skip.txt"
hash=$(send_sweep "$REHEARSAL_DIR/sweep-skip.txt" stuatom sweep-operator)
wait_tx strided_new "$hash"
checkpoint "two skipped with reasons" skipped_two_with_reasons "$hash"

# --- Arrival: stATOM on Osmosis, ATOM unwound to the hub, holder-base drained on Stride ---
wait_until 300 "stATOM landed on osmosis for holder-base" osmo_balance_positive "$HOLDER_BASE_OSMO" "$STATOM_ON_OSMO"
wait_until 300 "ATOM unwound to the hub for holder-base" hub_balance_positive "$HOLDER_BASE_COSMOS"
checkpoint "holder-base empty on stride" holder_base_has_none_left

# --- Injection: a sweep transfer that times out, refunds, and is resent ---
# The sweep's transfer timeout is WindDownTransferTimeout (60s). Everything above is swept, so first give
# user1 fresh ustrd as the single address to sweep
INJECT_AMOUNT=1000000
strided_new tx bank send faucet "$USER1_STRIDE" "${INJECT_AMOUNT}ustrd" --from faucet $STRIDE_TX > /dev/null
user1_ustrd() { strided_new q bank balance "$USER1_STRIDE" ustrd -o json | jq -r '.balance.amount'; }
user1_ustrd_at_least() { [[ $(user1_ustrd) -ge $1 ]]; }
user1_ustrd_below() { [[ $(user1_ustrd) -lt $1 ]]; }
wait_until 60 "user1 funded with ustrd" user1_ustrd_at_least "$INJECT_AMOUNT"
BEFORE=$(user1_ustrd)
printf '%s\n' "$USER1_STRIDE" > "$REHEARSAL_DIR/sweep-inject.txt"

$KX scale deployment relayer-stride-osmosis --replicas=0
trap '$KX scale deployment relayer-stride-osmosis --replicas=1' EXIT
hash=$(send_sweep "$REHEARSAL_DIR/sweep-inject.txt" ustrd sweep-operator)
wait_tx strided_new "$hash"
wait_until 60 "user1's ustrd left Stride (escrowed)" user1_ustrd_below "$BEFORE"
log "sleeping 80s: past the 60s WindDownTransferTimeout (plus margin) with the relayer paused"
sleep 80
$KX scale deployment relayer-stride-osmosis --replicas=1
trap - EXIT
wait_until 600 "timed-out sweep refunded to user1" user1_ustrd_at_least "$BEFORE"

hash=$(send_sweep "$REHEARSAL_DIR/sweep-inject.txt" ustrd sweep-operator)
wait_tx strided_new "$hash"
wait_until 300 "resent sweep left Stride" user1_ustrd_below "$BEFORE"
log "Phase 8 complete"
