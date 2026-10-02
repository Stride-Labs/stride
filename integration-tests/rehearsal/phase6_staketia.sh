#!/bin/bash
# Phase 6: staketia wind-down on the Hub side (runs on v35)
source "$(dirname "$0")/lib.sh"
source "$REHEARSAL_DIR/state.env"
log "## Phase 6: staketia"

ADMIN_TX=(strided_new admin-ms "$ADMIN_MS_MEMBERS" --)
TRANSFER_JSON=$(mktemp)

hub_bal() { gaiad q bank balance "$1" uatom -o json | jq -r .balance.amount; }
stride_bal() { strided_new q bank balance "$1" "$ATOM_ON_STRIDE" -o json | jq -r .balance.amount; }

hub_unbonding_done() { gaiad q staking unbonding-delegations "$HUB_MS_COSMOS" -o json | jq -e '.unbonding_responses | length == 0'; }
hub_bal_at_least() { [[ $(hub_bal "$1") -ge $2 ]]; }
hub_bal_below() { [[ $(hub_bal "$1") -lt $2 ]]; }
claim_at_least() { [[ $(stride_bal "$STAKETIA_CLAIM") -ge $1 ]]; }
claim_empty() { [[ $(stride_bal "$STAKETIA_CLAIM") -eq 0 ]]; }
no_redemption_records() { strided_new q staketia redemption-records -o json | jq -e '.redemption_record_responses | length == 0'; }
all_unbondings_claimed() {
  strided_new q staketia unbonding-records -o json |
    jq -e '[.unbonding_records[] | select(.status != "CLAIMED" and (.native_amount|tonumber) > 0)] | length == 0'
}

# receiver memo amount -> broadcast JSON (pipe to tx_hash): the staketia operator executes the multisig's grant-limited transfer
exec_transfer() {
  # assumes the laptop and cluster clocks agree
  local timeout_ns=$(( ($(date +%s) + 900) * 1000000000 ))
  cat > "$TRANSFER_JSON" <<JSON
{"body":{"messages":[{"@type":"/ibc.applications.transfer.v1.MsgTransfer","source_port":"transfer","source_channel":"channel-0","token":{"denom":"uatom","amount":"$3"},
 "sender":"$HUB_MS_COSMOS","receiver":"$1","timeout_height":{"revision_number":"0","revision_height":"0"},"timeout_timestamp":"$timeout_ns","memo":"$2"}]}}
JSON
  $KX cp "$TRANSFER_JSON" cosmoshub-validator-0:/tmp/xfer.json -c validator >/dev/null
  gaiad tx authz exec /tmp/xfer.json --from st-operator $HUB_TX
}

# ibc-go v11 transfer_authorization.go ("not allowed receiver address for transfer", "not allowed memo",
# "memo must be empty ...") and authz ("authorization not found"), plus the SDK's "unauthorized"
GRANT_DENIAL='not allowed receiver|not allowed memo|memo must be empty|authorization not found|unauthorized'

# receiver memo amount: passes only when the grant blocks the transfer (CheckTx/simulation error or a failed tx
# whose log names the grant denial); a cp or key failure matches no denial text and so fails the checkpoint
transfer_rejected() {
  local out hash
  out=$(exec_transfer "$1" "$2" "$3" 2>&1) || { log "$out"; grep -qiE "$GRANT_DENIAL" <<<"$out"; return; }

  # CheckTx rejection still returns JSON with a non-zero code
  if [[ "$(jq -r '.code // 0' <<<"$out" 2>/dev/null)" != 0 ]]; then
    log "$out"; jq -r '.raw_log' <<<"$out" | grep -qiE "$GRANT_DENIAL"; return
  fi

  hash=$(tx_hash <<<"$out")
  [[ -n "$hash" && "$hash" != null ]] || return 1
  wait_tx gaiad "$hash" && return 1
  gaiad q tx "$hash" -o json | jq -r '.raw_log' | grep -qiE "$GRANT_DENIAL"
}

# Phase 2 confirmed only the records that existed then. The R2/R3 record leaves ACCUMULATING at the next %4 day epoch,
# which can be several day epochs away: wait for it to become UNBONDING_QUEUE (or for nothing to accumulate), confirm
# every queued record against phase 2's Hub undelegation, then wait for the hour epoch to mark them UNBONDED.
staketia_unbondings_in() { # status -> ids, space-separated
  strided_new q staketia unbonding-records -o json | jq -r --arg status "$1" '[.unbonding_records[] | select(.status == $status) | .id] | join(" ")'
}
staketia_nothing_accumulating_or_queued() {
  strided_new q staketia unbonding-records -o json | jq -e '
    ([.unbonding_records[] | select(.status == "ACCUMULATING" and (.native_amount | tonumber) > 0)] | length == 0)
    or ([.unbonding_records[] | select(.status == "UNBONDING_QUEUE")] | length > 0)'
}
confirmed_records_unbonded() { # ids
  local id
  for id in $1; do
    [[ "$(strided_new q staketia unbonding-records -o json | jq -r --arg id "$id" '.unbonding_records[] | select(.id == $id) | .status')" == UNBONDED ]] || return 1
  done
}

log "waiting (up to 800s) for the accumulating staketia record to be frozen into UNBONDING_QUEUE at the next %4 day epoch"
wait_until 800 "no ACCUMULATING record with amount, or an UNBONDING_QUEUE record exists" staketia_nothing_accumulating_or_queued
CONFIRM_IDS=$(staketia_unbondings_in UNBONDING_QUEUE)
log "confirming staketia unbonding records [${CONFIRM_IDS:-none}] against phase 2's Hub undelegation $STK_UNBOND_HASH"
for id in $CONFIRM_IDS; do
  out=$(log_cmd "confirm-undelegation $id" strided_new tx staketia confirm-undelegation "$id" "$STK_UNBOND_HASH" --from st-operator $STRIDE_TX)
  wait_tx strided_new "$(grep '^{' <<<"$out" | tx_hash)"
done
log "waiting (up to 300s) for confirmed records [${CONFIRM_IDS:-none}] to reach UNBONDED (hour epoch)"
wait_until 300 "confirmed staketia records UNBONDED" confirmed_records_unbonded "$CONFIRM_IDS"

wait_until 400 "hub multisig unbonding matured" hub_unbonding_done
BAL=$(hub_bal "$HUB_MS_COSMOS")

checkpoint "transfer with memo rejected by grant" transfer_rejected "$STAKETIA_CLAIM" hello 1000
checkpoint "transfer to other receiver rejected" transfer_rejected "$USER1_STRIDE" "" 1000

# Not rerun-safe once the balance is swept, hence the guard
(( BAL > 0 )) || { log "hub multisig balance is 0, sweep already done"; exit 1; }
H=$(exec_transfer "$STAKETIA_CLAIM" "" "$BAL" | tx_hash)
wait_tx gaiad "$H"
wait_until 120 "claim address funded" claim_at_least "$BAL"

# Confirm every matured unbonding record against the sweep transfer
for id in $(strided_new q staketia unbonding-records -o json | jq -r '.unbonding_records[] | select(.status=="UNBONDED") | .id'); do
  out=$(log_cmd "confirm-sweep $id" strided_new tx staketia confirm-sweep "$id" "$H" --from st-operator $STRIDE_TX)
  wait_tx strided_new "$(grep '^{' <<<"$out" | tx_hash)"
done
wait_until 120 "redeemers paid (hour epoch)" no_redemption_records
checkpoint "every unbonding record CLAIMED or empty" all_unbondings_claimed

# Live test with 1 ATOM, then the whole claim balance
DEL_ICA=$(strided_new q stakeibc show-host-zone cosmoshub-test-1 -o json | jq -r .host_zone.delegation_ica_address)
D0=$(hub_bal "$DEL_ICA")
ms_tx "${ADMIN_TX[@]}" stakeibc transfer-staketia-claim-balance 1000000 >/dev/null
wait_until 120 "1 ATOM landed on the delegation ICA" hub_bal_at_least "$DEL_ICA" $((D0 + 1000000))
ms_tx "${ADMIN_TX[@]}" stakeibc transfer-staketia-claim-balance 0 >/dev/null
wait_until 120 "claim address empty" claim_empty

# Send the delegation ICA's current balance on to the vault (rerun-safe: skips when empty)
DEL_BAL=$(hub_bal "$DEL_ICA")
if [[ "$DEL_BAL" -gt 0 ]]; then
  ms_tx "${ADMIN_TX[@]}" stakeibc transfer-from-ica cosmoshub-test-1 DELEGATION "${DEL_BAL}uatom" >/dev/null
fi
wait_until 300 "staketia ATOM in the vault" hub_bal_below "$DEL_ICA" 1000
