#!/bin/bash
# Phase 2 (day 0): refresh every validator, prove the drain guards refuse, measure drift, staketia undelegation. Runs on v35.
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 2: day 0"

DRIFT_JSON=$REPO/scripts/wind-down/drift/drift.json
ADMIN_MEMBERS=m1,m2,m3

non_admin_refresh_rejected() { # valoper
  local out; out=$(strided_new tx stakeibc update-delegation cosmoshub-test-1 "$1" --from user1 $STRIDE_TX 2>&1) || true
  grep -qiE 'not an admin|unauthorized|invalid admin' <<<"$out"
}

no_slash_query_in_flight() {
  strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -e '[.validators[].slash_query_in_progress] | all(. != true)'
}

# The multisig admin is the only signer the msg accepts, so a refused drain shows up as a non-zero ms_tx
drain_refused() { # chain-id
  ! ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc undelegate-from-validators "$1" --all >/dev/null
}

# drift.json: .zones[<chain>].validators[] rows carry over_recorded, .zones[<chain>].summary.count_over_recorded
zero_over_recorded() {
  jq -e '(.zones | length == 2)
         and ([.zones[].error] | all(. == null))
         and ([.zones[].summary.count_over_recorded] | all(. == 0))
         and ([.zones[].validators[] | select(.over_recorded == true)] | length == 0)' "$DRIFT_JSON"
}

HUB_VALS=$(strided_new q stakeibc show-validators cosmoshub-test-1 -o json | jq -r '.validators[].address')
OSMO_VALS=$(strided_new q stakeibc show-validators osmosis-test-1 -o json | jq -r '.validators[].address')

# Only the admin multisig may refresh a validator
checkpoint "non-admin refresh rejected" non_admin_refresh_rejected "$(head -1 <<<"$HUB_VALS")"
for v in $HUB_VALS;  do ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc update-delegation cosmoshub-test-1 "$v" >/dev/null; done
for v in $OSMO_VALS; do ms_tx strided_new admin-ms $ADMIN_MEMBERS -- stakeibc update-delegation osmosis-test-1 "$v" >/dev/null; done

# Refresh queries answer asynchronously through the relayer
wait_until 300 "no slash query in flight" no_slash_query_in_flight
log_cmd "osmo validators after refresh" strided_new q stakeibc show-validators osmosis-test-1 -o json
checkpoint "osmo rate still frozen" assert_rate_unchanged osmosis-test-1 "$RATE_OSMO"

# Hub has queued records, osmo has a retry record: both must refuse a drain
checkpoint "drain refused while records queued (hub)" drain_refused cosmoshub-test-1
checkpoint "drain refused while retry record (osmo)"  drain_refused osmosis-test-1

log_cmd "drift" python3 "$REPO/scripts/wind-down/measure_delegation_drift.py" --chain-id cosmoshub-test-1 --chain-id osmosis-test-1
checkpoint "zero over-recorded" zero_over_recorded

# Staketia day 0: the operator undelegates the whole multisig delegation via authz and confirms the queued record(s)
HUB_DELEGATIONS=$(gaiad q staking delegations "$HUB_MS_COSMOS" -o json)
jq -e '.delegation_responses | length > 0' <<<"$HUB_DELEGATIONS" >/dev/null
jq -c --arg delegator "$HUB_MS_COSMOS" '{body: {messages: [.delegation_responses[] | {
    "@type": "/cosmos.staking.v1beta1.MsgUndelegate",
    delegator_address: $delegator,
    validator_address: .delegation.validator_address,
    amount: {denom: "uatom", amount: .balance.amount}}]}}' <<<"$HUB_DELEGATIONS" \
  | $KX exec -i cosmoshub-validator-0 -c validator -- sh -c 'cat > /tmp/unbond.json'

UNBOND_HASH=$(gaiad tx authz exec /tmp/unbond.json --from st-operator $HUB_TX | tx_hash)
wait_tx gaiad "$UNBOND_HASH"

for id in $(strided_new q staketia unbonding-records -o json | jq -r '.unbonding_records[] | select(.status=="UNBONDING_QUEUE") | .id'); do
  log_cmd "confirm-undelegation $id" strided_new tx staketia confirm-undelegation "$id" "$UNBOND_HASH" --from st-operator $STRIDE_TX
done
