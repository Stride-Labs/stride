#!/bin/bash
# Phase 3: last redemption cycle, claims, and proof that nothing new accrues. Runs on v35.
source "$(dirname "$0")/lib.sh"; source "$REHEARSAL_DIR/state.env"; log "## Phase 3: last redemption cycle"

all_unbondings_claimable() {
  strided_new q records list-epoch-unbonding-record -o json | jq -e '[.epoch_unbonding_record[].host_zone_unbondings[]
    | select((.native_token_amount|tonumber) > 0 and .status != "CLAIMABLE")] | length == 0'
}

claim_all() {
  strided_new q records list-user-redemption-record -o json \
    | jq -r '.user_redemption_record[] | select(.claim_is_pending==false) | "\(.host_zone_id) \(.epoch_number) \(.receiver)"' \
    | while read -r zone epoch receiver; do
        log_cmd "claim $zone $epoch" strided_new tx stakeibc claim-undelegated-tokens "$zone" "$epoch" "$receiver" --from user1 $STRIDE_TX
      done
}

no_user_redemption_records() {
  strided_new q records list-user-redemption-record -o json | jq -e '.user_redemption_record | length == 0'
}

stranded_deposit_queued() {
  strided_new q records list-deposit-record -o json | jq -e '[.deposit_record[] | select(.status=="DELEGATION_QUEUE")] | length >= 1'
}

latest_unbonding_epoch() {
  strided_new q records list-epoch-unbonding-record -o json | jq -r '[.epoch_unbonding_record[].epoch_number | tonumber] | (max // 0)'
}

no_reinvest_or_reward_ica() {
  local logs; logs=$($KX logs stride-validator-0 -c validator --since=3m 2>&1)
  ! grep -Eqi 'reinvest|ClaimAccruedStakingRewards|withdrawal balance' <<<"$logs"
}

wait_until 900 "every hub/osmo unbonding CLAIMABLE" all_unbondings_claimable

# A second pass picks up records that only became claimable after the first claims landed
claim_all; sleep 60; claim_all
wait_until 300 "zero user redemption records" no_user_redemption_records

log_cmd "deposit records" strided_new q records list-deposit-record -o json
checkpoint "stranded deposit never staked" stranded_deposit_queued

# Wait past at least one full day epoch (180s) so a new unbonding record would have been created
EPOCH_BEFORE=$(latest_unbonding_epoch)
sleep 200
checkpoint "hub rate frozen"  assert_rate_unchanged cosmoshub-test-1 "$RATE_HUB"
checkpoint "osmo rate frozen" assert_rate_unchanged osmosis-test-1 "$RATE_OSMO"
checkpoint "no new epoch unbonding records" test "$(latest_unbonding_epoch)" = "$EPOCH_BEFORE"
checkpoint "no reinvest/claim-rewards ICA" no_reinvest_or_reward_ica
