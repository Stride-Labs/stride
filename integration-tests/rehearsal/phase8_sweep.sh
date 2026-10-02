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
  strided_new tx stakeibc sweep-tokens-off-stride "$2" /tmp/sweep-batch.txt --from "$3" $STRIDE_TX | tx_hash
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

# --- Export, prices, batches ---
EXPORT=$REHEARSAL_DIR/export_pre_sweep.json
bash "$REHEARSAL_DIR/export.sh" "$EXPORT"

cat > "$PRICES_FILE" <<PRICES
{"stuatom":{"usd_per_token":4,"decimals":6},"stuosmo":{"usd_per_token":0.5,"decimals":6},"ustrd":{"usd_per_token":0.1,"decimals":6},"$ATOM_ON_STRIDE":{"usd_per_token":4,"decimals":6}}
PRICES

# --as-of is the block time at the export height, which the builder uses to measure vesting locks
EXPORT_HEIGHT=$(jq -r '.initial_height' "$EXPORT")
ASOF=$(strided_new q block --type=height "$EXPORT_HEIGHT" -o json | jq -r '.header.time // .block.header.time')
[[ -n "$ASOF" && "$ASOF" != null ]] || { log "could not read the block time at height $EXPORT_HEIGHT"; exit 1; }

rm -rf "$BATCH_DIR"
python3 "$REPO/scripts/wind-down/build_sweep_batches.py" --export "$EXPORT" --prices "$PRICES_FILE" --floor-usd 1 \
  --denoms stuatom,stuosmo,ustrd --extra-denom "$ATOM_ON_STRIDE=4:6" --out-dir "$BATCH_DIR" --as-of "$ASOF" \
  --sweep-operator "$SWEEP_OPERATOR" --batch-size 50 | tee -a "$LOG"

# The builder writes batch-NNN.txt, one stride address per line: exactly what the tx's addresses-file reads
BATCHES=("$BATCH_DIR"/batch-*.txt)
[[ -e "${BATCHES[0]}" ]] || { log "the builder wrote no batches"; exit 1; }

# The builder lists every funded genesis account at the $1 floor, which would drain the relayer keys and the faucet:
# filter every batch against the Stride addresses of the infrastructure keys and the admin multisig
EXCLUDE_FILE=$REHEARSAL_DIR/exclude.txt
: > "$EXCLUDE_FILE"
for name in faucet admin val1 val2 val3 val4 m1 m2 m3 d1 d2 d3 st-operator st-reward rly1 rly2 rly3 rly4 rly5 rly6 rly7 rly8; do
  strided_new keys show "$name" -a --keyring-backend test 2>/dev/null | tr -d '\r' >> "$EXCLUDE_FILE" || true
done
echo "$ADMIN_MS_STRIDE" >> "$EXCLUDE_FILE"
sed -i.bak '/^$/d' "$EXCLUDE_FILE" && rm -f "$EXCLUDE_FILE.bak"
log "excluding $(wc -l < "$EXCLUDE_FILE" | tr -d ' ') infrastructure addresses from the sweep batches"
for batch in "${BATCHES[@]}"; do
  grep -vxF -f "$EXCLUDE_FILE" "$batch" > "$batch.filtered" || true
  mv "$batch.filtered" "$batch"
done
# A batch emptied by the filter has nothing to sweep (an empty addresses-file is invalid)
NONEMPTY=()
for batch in "${BATCHES[@]}"; do
  if [[ -s "$batch" ]]; then NONEMPTY+=("$batch"); else rm -f "$batch"; fi
done
BATCHES=("${NONEMPTY[@]+"${NONEMPTY[@]}"}")  # bash 3.2: empty-array safe under set -u
[[ -e "${BATCHES[0]:-}" ]] || { log "every batch was empty after the exclusion filter"; exit 1; }

checkpoint "admin cannot sweep" admin_cannot_sweep "${BATCHES[0]}"

# --- Builder batches: each must skip nothing on chain ---
for batch in "${BATCHES[@]}"; do
  hash=$(send_sweep "$batch" "$SWEEP_DENOMS" sweep-operator)
  wait_tx strided_new "$hash"
  log "gas used: $(strided_new q tx "$hash" -o json | jq -r '.gas_used') for $(wc -l < "$batch" | tr -d ' ') addresses ($batch)"
  checkpoint "builder batch skipped nothing ($batch)" batch_skipped_nothing "$hash"
done

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
