#!/bin/bash
# Pre-upgrade state on v34 (spec docs/superpowers/specs/2026-10-02-wind-down-rehearsal-design.md).
# Every stride command goes through strided_old: the v35 binary has no liquid-stake / redeem-stake / staketia tx.
# Ends at about D4+55 (right after the hard status checkpoints) with state.env holding U, RATE_HUB, RATE_OSMO, HIST_TX.
source "$(dirname "$0")/lib.sh"
log "## Seed (v34)"
: > "$REHEARSAL_DIR/state.env"

# A tx whose CheckTx code is non-zero (or that errors) aborts the seed; the pause keeps one signer's
# sequence numbers from colliding between back-to-back txs (1s blocks, kubectl exec latency)
LAST_OUT=""
tx_step() { # label, then the command
  LAST_OUT=$(log_cmd "$@") || { log "FAILED: $1"; return 1; }
  printf '%s\n' "$LAST_OUT"
  if grep -Eq '"code": ?[1-9]' <<<"$LAST_OUT"; then log "REJECTED: $1"; return 1; fi
  sleep 3
}
put_file() { # pod path content: writes a file inside the pod
  printf '%s\n' "$3" | $KX exec -i "$1" -c validator -- sh -c "cat > $2"
}
vals_file_json() { # space-separated valopers -> add-validators JSON
  jq -n --arg list "$1" '{validators: ($list | split(" ") | to_entries
    | map({name: ("val" + ((.key + 1) | tostring)), address: .value, weight: 10}))}'
}
zone_field() { strided_old q stakeibc show-host-zone "$1" -o json | jq -r ".host_zone.$2"; }
zone_field_set() { local value; value=$(zone_field "$1" "$2"); [[ -n "$value" && "$value" != null ]]; }
zone_delegations_above() { (( $(zone_field "$1" total_delegations) > $2 )); }
balance_of() { # stride-address denom
  strided_old q bank balances "$1" -o json | jq -r --arg denom "$2" '.balances[] | select(.denom == $denom) | .amount'
}
hub_balance_has() { [[ -n "$(gaiad q bank balances "$1" -o json | jq -r --arg denom "$2" '.balances[] | select(.denom == $denom) | .amount')" ]]; }
stride_balance_has() { [[ -n "$(balance_of "$1" "$2")" ]]; }
osmo_val_jailed() { [[ "$(osmosisd q staking validator "$1" -o json | jq -r '.validator.jailed')" == true ]]; }
rate_limit_live() { strided_old q ratelimiting list-rate-limits -o json | jq -e '.rate_limits | length > 0'; }
hub_grant_present() {
  gaiad q authz grants "$HUB_MS_COSMOS" "$STAKETIA_OPERATOR_HUB" -o json \
    | jq -e '.grants[] | select(.authorization["@type"] == "/ibc.applications.transfer.v1.TransferAuthorization")'
}
pause_pod_process() { $KX exec "$1" -c validator -- sh -c "kill -STOP \$(pgrep -x $2)"; }
resume_pod_process() { $KX exec "$1" -c validator -- sh -c "kill -CONT \$(pgrep -x $2)"; }
staketia_queue_record_id() {
  strided_old q staketia unbonding-records -o json \
    | jq -r '[.unbonding_records[] | select(.status == "UNBONDING_QUEUE")][0].id // empty'
}
staketia_queue_record_amount() { # id
  strided_old q staketia unbonding-records -o json \
    | jq -r --arg id "$1" '.unbonding_records[] | select(.id == $id) | .native_amount'
}
hub_unbondings_in() { # status -> number of Hub host-zone unbondings in that status
  strided_old q records list-epoch-unbonding-record -o json | jq -r --arg status "$1" \
    '[.epoch_unbonding_record[].host_zone_unbondings[]? | select(.host_zone_id == "cosmoshub-test-1" and .status == $status)] | length'
}
hub_has_unbondings() { (( $(hub_unbondings_in "$1") >= $2 )); } # status, minimum count
hub_deposit_in_transfer_queue() {
  strided_old q records list-deposit-record -o json \
    | jq -e '[.deposit_record[]? | select(.host_zone_id == "cosmoshub-test-1" and .status == "TRANSFER_QUEUE")] | length > 0'
}
osmo_retry_record_present() {
  strided_old q records list-epoch-unbonding-record -o json \
    | jq -e '[.epoch_unbonding_record[].host_zone_unbondings[]? | select(.host_zone_id == "osmosis-test-1" and .status == "UNBONDING_RETRY_QUEUE")] | length > 0'
}
staketia_queue_ready() { [[ -n "$(staketia_queue_record_id)" ]]; }

############################################
# SEED_RESUME=1 skips Part A and the already-done top of Part B (zones registered, liquid stakes sent)
# and continues at the "osmo delegated" wait. HIST_TX is then the first liquid-stake tx found on chain.
############################################
if [[ "${SEED_RESUME:-0}" != 1 ]]; then
############################################
# Part A: multisigs, zones, validators
############################################

# Multisigs go into the pod keyrings (members were restored by init-chain)
strided_old keys add admin-ms --multisig "$ADMIN_MS_MEMBERS" --multisig-threshold 2 --keyring-backend test || true
osmosisd    keys add vault-ms --multisig "$ADMIN_MS_MEMBERS" --multisig-threshold 2 --keyring-backend test || true
gaiad       keys add hub-ms   --multisig "$HUB_MS_MEMBERS"   --multisig-threshold 2 --keyring-backend test || true
checkpoint "admin-ms address"  test "$(strided_old keys show admin-ms -a --keyring-backend test)" = "$ADMIN_MS_STRIDE"
checkpoint "vault-ms address"  test "$(osmosisd keys show vault-ms -a --keyring-backend test)" = "$VAULT_MS_OSMO"
checkpoint "hub-ms address"    test "$(gaiad keys show hub-ms -a --keyring-backend test)" = "$HUB_MS_COSMOS"

# Fund the multisigs and the sweep operator from the faucet
tx_step "fund admin-ms"      strided_old tx bank send faucet "$ADMIN_MS_STRIDE" 1000000000ustrd --from faucet $STRIDE_TX
tx_step "fund sweep operator" strided_old tx bank send faucet "$SWEEP_OPERATOR" 100000000ustrd --from faucet $STRIDE_TX
tx_step "fund vault-ms"      osmosisd tx bank send faucet "$VAULT_MS_OSMO" 100000000uosmo --from faucet $OSMO_TX
tx_step "fund hub-ms"        gaiad tx bank send faucet "$HUB_MS_COSMOS" 100000000uatom --from faucet $HUB_TX

# Validators of each host, ordered val1..valN by moniker
HUB_VALS=$(gaiad q staking validators -o json | jq -r '.validators | sort_by(.description.moniker) | .[].operator_address' | tr '\n' ' ')
OSMO_VALS=$(osmosisd q staking validators -o json | jq -r '.validators | sort_by(.description.moniker) | .[].operator_address' | tr '\n' ' ')
HUB_VALS=${HUB_VALS% }
OSMO_VALS=${OSMO_VALS% }
checkpoint "8 hub validators"     test "$(wc -w <<<"$HUB_VALS")" -eq 8
checkpoint "3 osmosis validators" test "$(wc -w <<<"$OSMO_VALS")" -eq 3

# Register the two zones (1-day cadence; 3 messages per ICA on the Hub so 8 validators make 3 batches)
tx_step "register hub zone" strided_old tx stakeibc register-host-zone connection-0 uatom cosmos "$ATOM_ON_STRIDE" channel-0 1 false \
  --max-messages-per-ica-tx 3 --from admin $STRIDE_TX
tx_step "register osmo zone" strided_old tx stakeibc register-host-zone connection-1 uosmo osmo "$OSMO_ON_STRIDE" channel-1 1 false \
  --from admin $STRIDE_TX

put_file stride-validator-0 /tmp/hub_vals.json  "$(vals_file_json "$HUB_VALS")"
put_file stride-validator-0 /tmp/osmo_vals.json "$(vals_file_json "$OSMO_VALS")"
tx_step "add hub validators"  strided_old tx stakeibc add-validators cosmoshub-test-1 /tmp/hub_vals.json --from admin $STRIDE_TX
tx_step "add osmo validators" strided_old tx stakeibc add-validators osmosis-test-1 /tmp/osmo_vals.json --from admin $STRIDE_TX

# All four ICAs per zone are used later (fee and withdrawal get funded), so wait for each handshake
for chain in cosmoshub-test-1 osmosis-test-1; do
  for field in delegation_ica_address fee_ica_address withdrawal_ica_address redemption_ica_address; do
    wait_until 300 "$chain $field" zone_field_set "$chain" "$field"
  done
done

############################################
# Part B: tokens on Stride, liquid stakes, holders
############################################

tx_step "atom to stride" gaiad tx ibc-transfer transfer transfer channel-0 "$USER1_STRIDE" 2000000000uatom --from user1 $HUB_TX
tx_step "osmo to stride" osmosisd tx ibc-transfer transfer transfer channel-0 "$USER1_STRIDE" 1000000000uosmo --from user1 $OSMO_TX
wait_until 120 "atom on stride" stride_balance_has "$USER1_STRIDE" "$ATOM_ON_STRIDE"
wait_until 120 "osmo on stride" stride_balance_has "$USER1_STRIDE" "$OSMO_ON_STRIDE"

tx_step "liquid stake 1000 ATOM" strided_old tx stakeibc liquid-stake 1000000000 uatom --from user1 $STRIDE_TX
# The earliest liquid stake is the one the post-upgrade "historical tx decodes" check looks up
HIST_TX=$(grep -Eo '"txhash": ?"[0-9A-Fa-f]{64}"' <<<"$LAST_OUT" | head -1 | grep -Eo '[0-9A-Fa-f]{64}')
tx_step "liquid stake 300 OSMO"  strided_old tx stakeibc liquid-stake 300000000 uosmo --from user1 $STRIDE_TX
wait_until 600 "hub delegated"  zone_delegations_above cosmoshub-test-1 900000000
else
  log "SEED_RESUME=1: skipping Part A and the first half of Part B"
  HUB_VALS=$(gaiad q staking validators -o json 2>/dev/null | jq -r '.validators | sort_by(.description.moniker) | .[].operator_address' | tr '\n' ' '); HUB_VALS=${HUB_VALS% }
  OSMO_VALS=$(osmosisd q staking validators -o json | jq -r '.validators | sort_by(.description.moniker) | .[].operator_address' | tr '\n' ' '); OSMO_VALS=${OSMO_VALS% }
  HIST_TX=$(strided_old q txs --query "message.action='/stride.stakeibc.MsgLiquidStake'" --limit 1 -o json 2>/dev/null | jq -r '.txs[0].txhash // empty')
  [[ -n "$HIST_TX" ]] || HIST_TX=$(strided_old q txs --query "message.module='stakeibc'" --limit 1 -o json 2>/dev/null | jq -r '.txs[0].txhash // empty')
  log "resume: HIST_TX=$HIST_TX, $(wc -w <<<"$HUB_VALS") hub validators, $(wc -w <<<"$OSMO_VALS") osmosis validators"
fi
wait_until 600 "osmo delegated" zone_delegations_above osmosis-test-1 250000000

# Holders: base, vesting (delayed, 1 year), distribution module (fund-community-pool), escrow (via IBC out)
tx_step "holder base" strided_old tx bank send user1 "$HOLDER_BASE" "50000000stuatom,10000000ustrd,20000000$ATOM_ON_STRIDE" $STRIDE_TX
tx_step "vesting acct" strided_old tx vesting create-vesting-account "$HOLDER_VESTING" 1000000ustrd "$(( $(date +%s) + 31536000 ))" \
  --delayed --from faucet $STRIDE_TX
tx_step "holder vesting" strided_old tx bank send user1 "$HOLDER_VESTING" 30000000stuatom,5000000ustrd $STRIDE_TX
tx_step "distribution holds stATOM" strided_old tx distribution fund-community-pool 5000000stuatom --from user1 $STRIDE_TX

# stATOM on Osmosis (canonical), on the Hub, and Hub -> Osmosis (two-hop)
tx_step "statom to osmosis" strided_old tx ibc-transfer transfer transfer channel-1 "$USER1_OSMO" 100000000stuatom --from user1 $STRIDE_TX
tx_step "statom to hub"     strided_old tx ibc-transfer transfer transfer channel-0 "$USER1_COSMOS" 60000000stuatom --from user1 $STRIDE_TX
tx_step "stosmo to osmosis" strided_old tx ibc-transfer transfer transfer channel-1 "$USER1_OSMO" 50000000stuosmo --from user1 $STRIDE_TX
STATOM_ON_HUB=ibc/$(printf 'transfer/channel-0/stuatom' | shasum -a 256 | cut -d' ' -f1 | tr a-f A-F)
wait_until 120 "statom on hub" hub_balance_has "$USER1_COSMOS" "$STATOM_ON_HUB"
tx_step "statom hub->osmosis (two-hop)" gaiad tx ibc-transfer transfer transfer channel-1 "$USER1_OSMO" "30000000$STATOM_ON_HUB" --from user1 $HUB_TX

# Balances the transfer tx will move later
tx_step "fund hub fee ICA"         gaiad tx bank send user1 "$(zone_field cosmoshub-test-1 fee_ica_address)" 3000000uatom --from user1 $HUB_TX
tx_step "fund hub withdrawal ICA"  gaiad tx bank send user1 "$(zone_field cosmoshub-test-1 withdrawal_ica_address)" 4000000uatom --from user1 $HUB_TX
tx_step "fund osmo fee ICA"        osmosisd tx bank send user1 "$(zone_field osmosis-test-1 fee_ica_address)" 3000000uosmo --from user1 $OSMO_TX
tx_step "fund osmo withdrawal ICA" osmosisd tx bank send user1 "$(zone_field osmosis-test-1 withdrawal_ica_address)" 4000000uosmo --from user1 $OSMO_TX

############################################
# Part C: staketia authz set and the rate limit
############################################

# The Hub multisig delegates the 50 ATOM genesis says it holds, and grants the operator the mainnet authz set
HUB_VAL1=${HUB_VALS%% *}
HUB_VALS_CSV=${HUB_VALS// /,}
ms_tx gaiad hub-ms "$HUB_MS_MEMBERS" -- staking delegate "$HUB_VAL1" 50000000uatom
ms_tx gaiad hub-ms "$HUB_MS_MEMBERS" -- authz grant "$STAKETIA_OPERATOR_HUB" unbond --allowed-validators "$HUB_VALS_CSV"
ms_tx gaiad hub-ms "$HUB_MS_MEMBERS" -- authz grant "$STAKETIA_OPERATOR_HUB" delegate --allowed-validators "$HUB_VALS_CSV"
ms_tx gaiad hub-ms "$HUB_MS_MEMBERS" -- authz grant "$STAKETIA_OPERATOR_HUB" generic --msg-type /cosmos.distribution.v1beta1.MsgWithdrawDelegatorReward

# ibc transfer grants have no CLI shortcut: build the tx by hand and run the 2-of-3 signing in the Hub pod
TRANSFER_GRANT_JSON=$(cat <<EOF
{"body":{"messages":[{"@type":"/cosmos.authz.v1beta1.MsgGrant","granter":"$HUB_MS_COSMOS","grantee":"$STAKETIA_OPERATOR_HUB",
 "grant":{"authorization":{"@type":"/ibc.applications.transfer.v1.TransferAuthorization","allocations":[{"source_port":"transfer","source_channel":"channel-0",
 "spend_limit":[{"denom":"uatom","amount":"1000000000000"}],"allow_list":["$STAKETIA_CLAIM"],"allowed_packet_data":[]}]},"expiration":null}}],
 "memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},
 "auth_info":{"signer_infos":[],"fee":{"amount":[{"denom":"uatom","amount":"600000"}],"gas_limit":"600000","payer":"","granter":""}},"signatures":[]}
EOF
)
put_file cosmoshub-validator-0 /tmp/unsigned.json "$TRANSFER_GRANT_JSON"
log_cmd "multisign + broadcast transfer grant" $KX exec cosmoshub-validator-0 -c validator -- sh -c "set -e
  gaiad tx sign /tmp/unsigned.json --from d1 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s1.json
  gaiad tx sign /tmp/unsigned.json --from d2 --multisig hub-ms --sign-mode amino-json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/s2.json
  gaiad tx multisign /tmp/unsigned.json hub-ms /tmp/s1.json /tmp/s2.json --keyring-backend test --chain-id cosmoshub-test-1 --output-document /tmp/signed.json
  gaiad tx broadcast /tmp/signed.json --chain-id cosmoshub-test-1 -o json"
sleep 6
checkpoint "transfer grant present" hub_grant_present

# Rate limit on stATOM over the Osmosis channel: gov-only, so a proposal voted by the four validators.
# The msg type URL and fields come from ibc-go v11 modules/apps/rate-limiting/types/tx.pb.go; signer is the gov module account.
RATE_LIMIT_PROPOSAL=$(cat <<EOF
{"title":"stATOM rate limit","summary":"rehearsal","metadata":"","deposit":"2000000000ustrd","messages":[{"@type":"/ibc.applications.rate_limiting.v1.MsgAddRateLimit",
 "signer":"stride10d07y265gmmuvt4z0w9aw880jnsr700jefnezl","denom":"stuatom","channel_or_client_id":"channel-1","max_percent_send":"10","max_percent_recv":"10","duration_hours":"24"}]}
EOF
)
put_file stride-validator-0 /tmp/ratelimit.json "$RATE_LIMIT_PROPOSAL"
tx_step "rate limit proposal" strided_old tx gov submit-proposal /tmp/ratelimit.json --from val1 $STRIDE_TX
sleep 4
PROP=$(strided_old q gov proposals -o json | jq -r '.proposals | max_by(.id | tonumber).id')
for index in 0 1 2 3; do
  strided_pod "stride-validator-$index" tx gov vote "$PROP" yes --from "val$((index + 1))" $STRIDE_TX >/dev/null
done
wait_until 120 "rate limit live" rate_limit_live

############################################
# Part D: the redemption timeline
############################################

# Staketia prepares undelegations only on every 4th day epoch (hooks.go: CurrentEpoch % 4 == 0). Pick the first such
# boundary at least 4 epochs out as D4, so the Hub timeline below fits before it. Epoch E_k (k >= 1) starts at D0 + 180*(k-1).
EPOCH_JSON=$(strided_old q stakeibc show-epoch-tracker day -o json)
EPOCH_NOW=$(jq -r '.epoch_tracker.epoch_number' <<<"$EPOCH_JSON")
D0=$(jq -r '.epoch_tracker.next_epoch_start_time' <<<"$EPOCH_JSON" | cut -c1-10)
EPOCHS_AHEAD=5
while (( (EPOCH_NOW + EPOCHS_AHEAD) % 4 != 0 )); do EPOCHS_AHEAD=$((EPOCHS_AHEAD + 1)); done
D4=$((D0 + 180 * (EPOCHS_AHEAD - 1)))
D3=$((D4 - 180)); D2=$((D4 - 360)); D1=$((D4 - 540))
# Upgrade lands 150s after D4 so the proposal (voting 30s) has time to pass after the seed finishes (~D4+50).
# That is still before D5 (D4+180), so the queued Hub redemption is still pending at U (the in-flight deposit is seeded by phase 1 at U-20),
# and before the staketia record's 240s unbonding completes.
U=$((D4 + 150))
log "day epoch now=$EPOCH_NOW: D0=$D0 D1=$D1 D2=$D2 D3=$D3 D4=$D4 (staketia prepare epoch) upgrade target U=$U"

# Hub RA -> CLAIMABLE: submitted early so it unbonds and sweeps before U
tx_step "hub RA redeem" strided_old tx stakeibc redeem-stake 30000000 cosmoshub-test-1 "$USER1_COSMOS" --from user1 $STRIDE_TX

# Osmosis zone: val3 -> weight 0, then slashed by downtime, then a redemption that must land in UNBONDING_RETRY_QUEUE.
# (A clean osmosis CLAIMABLE record from before the slash is impossible now; the Hub's RA covers CLAIMABLE.)
OSMO_VAL3=$(cut -d' ' -f3 <<<"$OSMO_VALS")
tx_step "osmo val3 weight 0" strided_old tx stakeibc change-validator-weight osmosis-test-1 "$OSMO_VAL3" 0 --from admin $STRIDE_TX
log_cmd "stop signing osmosis-validator-2" pause_pod_process osmosis-validator-2 osmosisd
# A stopped validator must never outlive the script: resume on any exit
trap 'resume_pod_process osmosis-validator-2 osmosisd' EXIT
wait_until 240 "osmo val3 jailed" osmo_val_jailed "$OSMO_VAL3"
log_cmd "resume osmosis-validator-2" resume_pod_process osmosis-validator-2 osmosisd
trap - EXIT
# 120M exceeds val3's recorded delegation, so the zero-weight cascade attempts a full drain and the ICA fails
tx_step "osmo RE redeem (retry)" strided_old tx stakeibc redeem-stake 120000000 osmosis-test-1 "$USER1_OSMO" --from user1 $STRIDE_TX

# Hub: RB (D2) and RC (D3) both end in EXIT_TRANSFER_QUEUE (the undelegate ack skips IN_PROGRESS; RB matured, RC not), RD -> QUEUE (after D4)
sleep_until $((D2 + 10))
tx_step "hub RB redeem" strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 "$USER1_COSMOS" --from user1 $STRIDE_TX
sleep_until $((D3 + 10))
tx_step "hub RC redeem" strided_old tx stakeibc redeem-stake 20000000 cosmoshub-test-1 "$USER1_COSMOS" --from user1 $STRIDE_TX

# Staketia R1 lands in the accumulating record that the D4 prepare freezes into UNBONDING_QUEUE
tx_step "staketia R1" strided_old tx staketia redeem-stake 20000000 "$USER1_COSMOS" --from user1 $STRIDE_TX

# v34 operator flow: wait for the prepare, undelegate on the Hub through authz as the operator, then confirm on Stride
sleep_until $((D4 + 5))
wait_until 120 "staketia unbonding record in UNBONDING_QUEUE" staketia_queue_ready
STK_RECORD_ID=$(staketia_queue_record_id)
STK_NATIVE_AMOUNT=$(staketia_queue_record_amount "$STK_RECORD_ID")
log "staketia record $STK_RECORD_ID native_amount=$STK_NATIVE_AMOUNT"
UNDELEGATE_OUT=$(log_cmd "hub-ms undelegate via authz exec" $KX exec cosmoshub-validator-0 -c validator -- sh -c "set -e
  gaiad tx staking unbond $HUB_VAL1 ${STK_NATIVE_AMOUNT}uatom --from hub-ms --generate-only --keyring-backend test --chain-id cosmoshub-test-1 > /tmp/unbond.json
  gaiad tx authz exec /tmp/unbond.json --from st-operator $HUB_TX")
printf '%s\n' "$UNDELEGATE_OUT"
STK_UNDELEGATE_HASH=$(grep -Eo '"txhash": ?"[0-9A-Fa-f]{64}"' <<<"$UNDELEGATE_OUT" | head -1 | grep -Eo '[0-9A-Fa-f]{64}')
checkpoint "hub undelegate tx hash captured" test -n "$STK_UNDELEGATE_HASH"
sleep 6
tx_step "staketia confirm-undelegation" strided_old tx staketia confirm-undelegation "$STK_RECORD_ID" "$STK_UNDELEGATE_HASH" --from st-operator $STRIDE_TX

# After D4: RD queues behind the in-flight Hub unbonding; R2 and R3 open a new accumulating staketia record (spillover)
sleep_until $((D4 + 10))
tx_step "hub RD redeem (queue)" strided_old tx stakeibc redeem-stake 10000000 cosmoshub-test-1 "$USER1_COSMOS" --from user1 $STRIDE_TX
tx_step "staketia R2" strided_old tx staketia redeem-stake 20000000 "$USER1_COSMOS" --from user1 $STRIDE_TX
tx_step "staketia R3 spillover" strided_old tx staketia redeem-stake 40000000 "$USER1_COSMOS" --from user1 $STRIDE_TX

# The record mix the post-upgrade phases depend on. The seed ends right after the hard checkpoints.
# The retry needs the failed ICA ack (up to one more day epoch after RE); cap the wait so it ends by U-30.
RETRY_WAIT=$(( U - 30 - $(date +%s) ))
(( RETRY_WAIT < 10 )) && RETRY_WAIT=10
wait_until "$RETRY_WAIT" "osmo RE in UNBONDING_RETRY_QUEUE" osmo_retry_record_present
checkpoint "hub RA CLAIMABLE"              hub_has_unbondings CLAIMABLE 1
checkpoint "hub RB+RC EXIT_TRANSFER_QUEUE" hub_has_unbondings EXIT_TRANSFER_QUEUE 1
# RB is swept ~35s after this point, so the full count of 2 is only a soft check
CHECKPOINT_SOFT=1 checkpoint "hub RB+RC both EXIT_TRANSFER_QUEUE" hub_has_unbondings EXIT_TRANSFER_QUEUE 2
checkpoint "hub RD UNBONDING_QUEUE"        hub_has_unbondings UNBONDING_QUEUE 1

RATE_HUB=$(rate_of cosmoshub-test-1)
RATE_OSMO=$(rate_of osmosis-test-1)
log "pre-upgrade rates: hub=$RATE_HUB osmo=$RATE_OSMO"
printf 'RATE_HUB=%s\nRATE_OSMO=%s\nU=%s\nHIST_TX=%s\n' "$RATE_HUB" "$RATE_OSMO" "$U" "$HIST_TX" >> "$REHEARSAL_DIR/state.env"
checkpoint "state.env has HIST_TX" test -n "$HIST_TX"

log_cmd "records at upgrade" strided_old q records list-epoch-unbonding-record -o json
log_cmd "staketia records at upgrade" strided_old q staketia unbonding-records -o json
log "seed done at $(date +%s); upgrade target U=$U (now - U = $(( $(date +%s) - U ))s)"
