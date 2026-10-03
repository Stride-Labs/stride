#!/bin/bash
# Phase 5: move every host-zone ICA balance to the Osmosis vault multisig (runs on v35)
source "$(dirname "$0")/lib.sh"
source "$REHEARSAL_DIR/state.env"
log "## Phase 5: transfers to Osmosis"

ATOM_ON_OSMO=ibc/$(printf 'transfer/channel-1/uatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
HUB_ZONE=cosmoshub-test-1
OSMO_ZONE=osmosis-test-1
ADMIN_TX=(strided_new admin-ms "$ADMIN_MS_MEMBERS" --)

vault_bal() { osmosisd q bank balance "$VAULT_MS_OSMO" "$1" -o json | jq -r '.balance.amount'; }

# <chain-wrapper> <zone> <ICA_TYPE> <denom>: the host denom's balance on that ICA (0 when absent)
ica_bal() {
  local addr field
  field="$(tr A-Z a-z <<<"$3")_ica_address"
  addr=$(strided_new q stakeibc show-host-zone "$2" -o json | jq -r --arg f "$field" '.host_zone[$f]')
  $1 q bank balances "$addr" -o json | jq -r --arg d "$4" '[.balances[] | select(.denom==$d) | .amount][0] // "0"'
}

# zone: every record CLAIMABLE or zero, no user records left
pre_checklist() {
  strided_new q records list-epoch-unbonding-record -o json | jq -e --arg z "$1" \
    '[.epoch_unbonding_record[].host_zone_unbondings[] | select(.host_zone_id==$z and (.native_token_amount|tonumber)>0 and .status!="CLAIMABLE")] | length == 0' &&
  strided_new q records list-user-redemption-record -o json | jq -e --arg z "$1" \
    '[.user_redemption_record[] | select(.host_zone_id==$z)] | length == 0'
}

# Predicates for wait_until (run in this shell, so functions are visible)
vault_above() { [[ $(vault_bal "$1") -gt $2 ]]; }
ica_at_least() { [[ $(ica_bal "$1" "$2" "$3" "$4") -ge $5 ]]; }
ica_below() { [[ $(ica_bal "$1" "$2" "$3" "$4") -lt $5 ]]; }

# chain zone type denom: send the ICA's whole current balance, skipping empty accounts
transfer_ica_balance() {
  local amt; amt=$(ica_bal "$1" "$2" "$3" "$4")
  if [[ "$amt" -le 0 ]]; then log "$2 $3 ICA is empty, nothing to transfer"; return 0; fi
  log "transfer-from-ica $2 $3 ${amt}$4"
  ms_tx "${ADMIN_TX[@]}" stakeibc transfer-from-ica "$2" "$3" "${amt}$4" >/dev/null
}

B0=$(vault_bal "$ATOM_ON_OSMO"); O0=$(vault_bal uosmo)
# Rerun-safe: when an earlier attempt already moved these balances the vault is funded and nothing new lands
(( B0 > 0 )) && B0=$(( B0 - 1 )); (( O0 > 100000000 )) && O0=$(( O0 - 1 ))

# Live test: WITHDRAWAL and FEE first, both zones
for ica_type in WITHDRAWAL FEE; do
  transfer_ica_balance gaiad "$HUB_ZONE" "$ica_type" uatom
  transfer_ica_balance osmosisd "$OSMO_ZONE" "$ica_type" uosmo
done
wait_until 300 "hub tokens landed in vault as ATOM-on-Osmosis" vault_above "$ATOM_ON_OSMO" "$B0"
wait_until 120 "osmo bank-send form landed" vault_above uosmo "$O0"

# Injection: a transfer that times out and refunds. WindDownTransferTimeout is 60s and the ICA-wrapped host->Osmosis
# transfer uses an inner timeout of 2 x that (120s), so the relayer stays paused ~140s before a refund is possible
$KX scale deployment relayer-cosmoshub-osmosis --replicas=0
trap '$KX scale deployment relayer-cosmoshub-osmosis --replicas=1' EXIT
W=$(ica_bal gaiad "$HUB_ZONE" WITHDRAWAL uatom)
if [[ "$W" == 0 ]]; then
  WITHDRAWAL_ICA=$(strided_new q stakeibc show-host-zone "$HUB_ZONE" -o json | jq -r .host_zone.withdrawal_ica_address)
  gaiad tx bank send user1 "$WITHDRAWAL_ICA" 1000000uatom $HUB_TX >/dev/null
  sleep 6
  W=1000000
fi
ms_tx "${ADMIN_TX[@]}" stakeibc transfer-from-ica "$HUB_ZONE" WITHDRAWAL "${W}uatom" >/dev/null
log "sleeping 140s: the ICA-wrapped transfer's inner timeout is 2 x the 60s WindDownTransferTimeout (120s) plus margin"
sleep 140
$KX scale deployment relayer-cosmoshub-osmosis --replicas=1
trap - EXIT
wait_until 300 "timed-out transfer refunded to the withdrawal ICA" ica_at_least gaiad "$HUB_ZONE" WITHDRAWAL uatom "$W"
ms_tx "${ADMIN_TX[@]}" stakeibc transfer-from-ica "$HUB_ZONE" WITHDRAWAL "${W}uatom" >/dev/null
wait_until 300 "resubmission landed" ica_below gaiad "$HUB_ZONE" WITHDRAWAL uatom "$W"

# Full balances once the pre-transfer checklist passes
checkpoint "hub pre-transfer checklist" pre_checklist "$HUB_ZONE"
checkpoint "osmo pre-transfer checklist" pre_checklist "$OSMO_ZONE"
for entry in "$HUB_ZONE:gaiad:uatom" "$OSMO_ZONE:osmosisd:uosmo"; do
  IFS=: read -r zone chain denom <<<"$entry"
  for ica_type in DELEGATION WITHDRAWAL REDEMPTION; do
    transfer_ica_balance "$chain" "$zone" "$ica_type" "$denom"
  done
done
for entry in "$HUB_ZONE:gaiad:uatom" "$OSMO_ZONE:osmosisd:uosmo"; do
  IFS=: read -r zone chain denom <<<"$entry"
  for ica_type in DELEGATION WITHDRAWAL FEE REDEMPTION; do
    wait_until 300 "$zone $ica_type at dust" ica_below "$chain" "$zone" "$ica_type" "$denom" 1000
  done
done
log_cmd "vault balances" osmosisd q bank balances "$VAULT_MS_OSMO" -o json
