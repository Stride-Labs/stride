#!/usr/bin/env bats

load "bats/bats-support/load.bash"
load "bats/bats-assert/load.bash"

setup_file() {
  _example_run_command="CHAIN_NAME=GAIA TRANSFER_CHANNEL_NUMBER=0 bats gaia_tests.bats"
  if [[ "$CHAIN_NAME" == "" ]]; then 
    echo "CHAIN_NAME variable must be set before running integration tests (e.g. $_example_run_command)" >&2
    return 1
  fi

  if [[ "$TRANSFER_CHANNEL_NUMBER" == "" ]]; then 
    echo "TRANSFER_CHANNEL_NUMBER variable must be set before running integration tests (e.g. $_example_run_command)" >&2
    return 1
  fi

  # set allows us to export all variables in account_vars
  set -a
  NEW_BINARY="${NEW_BINARY:-false}" source dockernet/config.sh

  HOST_CHAIN_ID=$(GET_VAR_VALUE  ${CHAIN_NAME}_CHAIN_ID)
  HOST_DENOM=$(GET_VAR_VALUE     ${CHAIN_NAME}_DENOM)
  HOST_IBC_DENOM=$(GET_VAR_VALUE IBC_${CHAIN_NAME}_CHANNEL_${TRANSFER_CHANNEL_NUMBER}_DENOM)
  HOST_MAIN_CMD=$(GET_VAR_VALUE  ${CHAIN_NAME}_MAIN_CMD)

  HOST_VAL_ADDRESS=$(${CHAIN_NAME}_ADDRESS)
  HOST_RECEIVER_ADDRESS=$(GET_VAR_VALUE ${CHAIN_NAME}_RECEIVER_ADDRESS)

  HOST_VAL="$(GET_VAR_VALUE ${CHAIN_NAME}_VAL_PREFIX)1"
  STRIDE_VAL=${STRIDE_VAL_PREFIX}1

  STRIDE_TRANSFER_CHANNEL="channel-${TRANSFER_CHANNEL_NUMBER}"
  HOST_TRANSFER_CHANNEL="channel-0"

  # IBC sttoken is only needed for autopilot tests which run on GAIA and HOST
  IBC_STTOKEN=$(GET_VAR_VALUE IBC_${CHAIN_NAME}_STDENOM)

  TRANSFER_AMOUNT=50000000
  STAKE_AMOUNT=10000000
  REDEEM_AMOUNT=10000
  PACKET_FORWARD_STAKE_AMOUNT=300000

  HOST_FEES="--fees 1000000ufee"

  # HELPER FUNCTIONS
  DECADD () {
    echo "scale=2; $1+$2" | bc
  }
  DECMUL () {
    echo "scale=2; $1*$2" | bc
  }
  FLOOR () {
    printf "%.0f\n" $1
  }
  CEIL () {
    printf "%.0f\n" $(ADD $1 1)
  }

  set +a
}

##############################################################################################
######                              HOW TO                                              ######
##############################################################################################
# Tests are written sequentially
# Each test depends on the previous tests, and examines the chain state at a point in time

# To add a new test, take an action then sleep for seconds / blocks / IBC_TX_WAIT_SECONDS
#     / epochs
# Reordering existing tests could break them

##############################################################################################
######                              SETUP TESTS                                         ######
##############################################################################################
# confirm host zone is registered
@test "[INTEGRATION-BASIC-$CHAIN_NAME] host zones successfully registered" {
  run $STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID
  assert_line "  host_denom: $HOST_DENOM"
  assert_line "  chain_id: $HOST_CHAIN_ID"
  assert_line "  transfer_channel_id: channel-$TRANSFER_CHANNEL_NUMBER"
  refute_line '  delegation_ica_address: ""'
  refute_line '  fee_ica_address: ""'
  refute_line '  redemption_ica_address: "'
  refute_line '  withdrawal_ica_address: ""'
  assert_line '  unbonding_period: "1"'
}

##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################


@test "[INTEGRATION-BASIC-$CHAIN_NAME] ibc transfer" {
  # get initial balances
  sval_strd_balance_start=$(GET_BALANCE  STRIDE      $(STRIDE_ADDRESS) $STRIDE_DENOM)
  hval_strd_balance_start=$(GET_BALANCE  $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STRD_DENOM)
  sval_token_balance_start=$(GET_BALANCE STRIDE      $(STRIDE_ADDRESS) $HOST_IBC_DENOM)
  hval_token_balance_start=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $HOST_DENOM)

  # do IBC transfer
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer $STRIDE_TRANSFER_CHANNEL $HOST_VAL_ADDRESS ${TRANSFER_AMOUNT}${STRIDE_DENOM} --from $STRIDE_VAL -y 
  $HOST_MAIN_CMD   tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL  $(STRIDE_ADDRESS) ${TRANSFER_AMOUNT}${HOST_DENOM} --from $HOST_VAL -y $HOST_FEES

  WAIT_FOR_BLOCK $STRIDE_LOGS 8

  # get new balances
  sval_strd_balance_end=$(GET_BALANCE  STRIDE      $(STRIDE_ADDRESS) $STRIDE_DENOM)
  hval_strd_balance_end=$(GET_BALANCE  $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STRD_DENOM)
  sval_token_balance_end=$(GET_BALANCE STRIDE      $(STRIDE_ADDRESS) $HOST_IBC_DENOM)
  hval_token_balance_end=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $HOST_DENOM)

  # get all STRD balance diffs
  sval_strd_balance_diff=$(($sval_strd_balance_start - $sval_strd_balance_end))
  hval_strd_balance_diff=$(($hval_strd_balance_start - $hval_strd_balance_end))
  assert_equal "$sval_strd_balance_diff" "$TRANSFER_AMOUNT"
  assert_equal "$hval_strd_balance_diff" "-$TRANSFER_AMOUNT"

  # get all host balance diffs
  sval_token_balance_diff=$(($sval_token_balance_start - $sval_token_balance_end))
  hval_token_balance_diff=$(($hval_token_balance_start - $hval_token_balance_end))
  assert_equal "$sval_token_balance_diff" "-$TRANSFER_AMOUNT"
  assert_equal "$hval_token_balance_diff" "$TRANSFER_AMOUNT"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] liquid stake mint and transfer" {
  # get initial balances on stride account
  token_balance_start=$(GET_BALANCE   STRIDE $(STRIDE_ADDRESS) $HOST_IBC_DENOM)
  sttoken_balance_start=$(GET_BALANCE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM)

  # get initial ICA accound balance
  delegation_address=$(GET_ICA_ADDR $HOST_CHAIN_ID delegation)
  delegation_ica_balance_start=$(GET_BALANCE $CHAIN_NAME $delegation_address $HOST_DENOM)

  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake $STAKE_AMOUNT $HOST_DENOM --from $STRIDE_VAL -y 

  # wait for the stTokens to get minted 
  WAIT_FOR_BALANCE_CHANGE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM 

  # make sure IBC_DENOM went down
  token_balance_end=$(GET_BALANCE STRIDE $(STRIDE_ADDRESS) $HOST_IBC_DENOM)
  token_balance_diff=$(($token_balance_start - $token_balance_end))
  assert_equal "$token_balance_diff" $STAKE_AMOUNT

  # make sure stToken went up
  sttoken_balance_end=$(GET_BALANCE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" $STAKE_AMOUNT

  # Wait for the transfer to complete
  WAIT_FOR_BALANCE_CHANGE $CHAIN_NAME $delegation_address $HOST_DENOM 

  # get the new delegation ICA balance
  delegation_ica_balance_end=$(GET_BALANCE $CHAIN_NAME $delegation_address $HOST_DENOM)
  diff=$(($delegation_ica_balance_end - $delegation_ica_balance_start))
  assert_equal "$diff" $STAKE_AMOUNT
}

# check that tokens on the host are staked
@test "[INTEGRATION-BASIC-$CHAIN_NAME] delegation on $CHAIN_NAME" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS "\[DELEGATION\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 4

  # check staked tokens
  NEW_STAKE=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $CHAIN_NAME 1) | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] LSM liquid stake" {
  if [[ "$CHAIN_NAME" != "GAIA" ]]; then
    skip "Skipping LSM liquid stake for chains without LSM support" 
  fi

  staker_address_on_host=$(GET_ADDRESS $CHAIN_NAME $USER_ACCT)
  staker_address_on_stride=$(GET_ADDRESS STRIDE $USER_ACCT)
  validator_address=$(GET_VAL_ADDR $CHAIN_NAME 1)

  # delegate on the host chain
  $HOST_MAIN_CMD tx staking delegate $validator_address ${TRANSFER_AMOUNT}${HOST_DENOM} --from $USER_ACCT -y $HOST_FEES
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # tokenize shares
  $HOST_MAIN_CMD tx staking tokenize-share $validator_address ${TRANSFER_AMOUNT}${HOST_DENOM} $staker_address_on_host --from $USER_ACCT -y --gas 1000000 $HOST_FEES
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # get the record id from the tokenized share record
  record_id=$($HOST_MAIN_CMD q staking last-tokenize-share-record-id | awk '{print $2}' | tr -d '"')

  # transfer LSM tokens to stride
  lsm_token_denom=${validator_address}/${record_id}
  $HOST_MAIN_CMD tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL \
    $staker_address_on_stride ${TRANSFER_AMOUNT}${lsm_token_denom} --from $USER_ACCT -y $HOST_FEES
  
  WAIT_FOR_BLOCK $STRIDE_LOGS 8

  lsm_token_ibc_denom=$(GET_IBC_DENOM STRIDE $STRIDE_TRANSFER_CHANNEL ${validator_address}/${record_id})

  # get initial balances
  sttoken_balance_start=$(GET_BALANCE STRIDE $staker_address_on_stride st$HOST_DENOM)
  lsm_token_balance_start=$(GET_BALANCE STRIDE $staker_address_on_stride $lsm_token_ibc_denom)

  # LSM-liquid stake
  $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake $STAKE_AMOUNT $lsm_token_ibc_denom --from $USER_ACCT -y 
  WAIT_FOR_BALANCE_CHANGE STRIDE $staker_address_on_stride st$HOST_DENOM

  # make sure stToken went up
  sttoken_balance_end=$(GET_BALANCE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" "$STAKE_AMOUNT"

  # make sure LSM IBC Token balance went down
  lsm_token_balance_end=$(GET_BALANCE STRIDE $staker_address_on_stride $lsm_token_ibc_denom)
  lsm_token_balance_diff=$(($lsm_token_balance_start - $lsm_token_balance_end))
  assert_equal "$lsm_token_balance_diff" $STAKE_AMOUNT

  # wait for LSM token to get transferred and converted to native stake
  delegation_start=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID | grep "total_delegations" | NUMBERS_ONLY)
  WAIT_FOR_DELEGATION_CHANGE $HOST_CHAIN_ID $STAKE_AMOUNT

  # confirm delegation increased
  delegation_end=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID | grep "total_delegations" | NUMBERS_ONLY)
  delegation_diff=$(($delegation_end - $delegation_start))
  assert_equal $(echo "$delegation_diff >= $STAKE_AMOUNT" | bc -l) "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] LSM liquid stake with slash query" {
  if [[ "$CHAIN_NAME" != "GAIA" ]]; then
    skip "Skipping LSM liquid stake for chains without LSM support" 
  fi

  # get staker and validator addresses
  validator_address=$(GET_VAL_ADDR $CHAIN_NAME 1)
  staker_address_on_stride=$(GET_ADDRESS STRIDE $USER_ACCT)

  # get the LSM token denom
  record_id=$($HOST_MAIN_CMD q staking last-tokenize-share-record-id | awk '{print $2}' | tr -d '"')
  lsm_token_ibc_denom=$(GET_IBC_DENOM STRIDE $STRIDE_TRANSFER_CHANNEL ${validator_address}/${record_id})

  # get the stToken balance before the liquid stake
  sttoken_balance_start=$(GET_BALANCE STRIDE $staker_address_on_stride st$HOST_DENOM)

  # LSM-liquid stake again, this time the slash query should be invoked
  $STRIDE_MAIN_CMD tx stakeibc lsm-liquid-stake $STAKE_AMOUNT $lsm_token_ibc_denom --from $USER_ACCT -y 
  WAIT_FOR_BALANCE_CHANGE STRIDE $staker_address_on_stride st$HOST_DENOM

  # make sure stToken went up (after the slash query query callback)
  sttoken_balance_end=$(GET_BALANCE STRIDE $staker_address_on_stride st$HOST_DENOM)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" "$STAKE_AMOUNT"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] autopilot liquid stake" {
  memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "action": "LiquidStake" } } }'

  # get initial balances
  sttoken_balance_start=$(GET_BALANCE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM)

  # Send the IBC transfer with the JSON memo
  $HOST_MAIN_CMD tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL \
    $(STRIDE_ADDRESS) ${PACKET_FORWARD_STAKE_AMOUNT}${HOST_DENOM} --memo "$memo" --from $HOST_VAL -y $HOST_FEES

  # Wait for the transfer to complete
  WAIT_FOR_BALANCE_CHANGE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM

  # make sure stATOM balance increased
  sttoken_balance_end=$(GET_BALANCE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" "$PACKET_FORWARD_STAKE_AMOUNT"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] autopilot liquid stake and transfer" {
  memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "action": "LiquidStake", "ibc_receiver": "'$HOST_VAL_ADDRESS'" } } }'

  # get initial balances
  stibctoken_balance_start=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STTOKEN 2>/dev/null)

  # Send the IBC transfer with the JSON memo
  $HOST_MAIN_CMD tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL \
      $(STRIDE_ADDRESS) ${PACKET_FORWARD_STAKE_AMOUNT}${HOST_DENOM} --memo "$memo" --from $HOST_VAL -y $HOST_FEES

  # Wait for the transfer to complete
  WAIT_FOR_BALANCE_CHANGE $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STTOKEN

  # make sure stATOM balance increased
  stibctoken_balance_end=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STTOKEN 2>/dev/null)
  stibctoken_balance_diff=$(($stibctoken_balance_end-$stibctoken_balance_start))
  assert_equal "$stibctoken_balance_diff" "$PACKET_FORWARD_STAKE_AMOUNT"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] autopilot redeem stake" {
  # Over the next two tests, we will run two redemptions in a row and we want both to occur in the same epoch
  # To ensure we don't accidentally cross the epoch boundary, we'll make sure there's enough of a buffer here
  # between the two redemptions
  AVOID_EPOCH_BOUNDARY day 25

  # get initial balances
  stibctoken_balance_start=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STTOKEN)

  memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "action": "RedeemStake", "ibc_receiver": "'$HOST_RECEIVER_ADDRESS'" } } }'

  # do IBC transfer
  # For all other hosts (ibc-v5), pass an address for a receiver and the memo in the --memo field
  $HOST_MAIN_CMD tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL \
    $(STRIDE_ADDRESS) ${REDEEM_AMOUNT}${IBC_STTOKEN} --memo "$memo" --from $HOST_VAL -y $HOST_FEES

  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # make sure stATOM balance decreased
  stibctoken_balance_mid=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STTOKEN)
  stibctoken_balance_diff=$(($stibctoken_balance_start-$stibctoken_balance_mid))
  assert_equal "$stibctoken_balance_diff" "$REDEEM_AMOUNT"

  WAIT_FOR_BLOCK $STRIDE_LOGS 5

  # check that a user redemption record was created
  redemption_record_native_amount=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -Fiw 'native_token_amount' | head -n 1 | grep -o -E '[0-9]+')
  amount_positive=$(($redemption_record_native_amount > 0))
  assert_equal "$amount_positive" "1"

  # attempt to redeem with an invalid receiver address to invoke a failure
  invalid_memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "action": "RedeemStake", "ibc_receiver": "XXX" } } }'
  $HOST_MAIN_CMD tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL \
     $(STRIDE_ADDRESS) ${REDEEM_AMOUNT}${IBC_STTOKEN} --memo "$invalid_memo" --from $HOST_VAL -y $HOST_FEES
  
  WAIT_FOR_BLOCK $STRIDE_LOGS 10

  # Confirm the stATOM balance was refunded
  stibctoken_balance_end=$(GET_BALANCE $CHAIN_NAME $HOST_VAL_ADDRESS $IBC_STTOKEN)
  assert_equal "$stibctoken_balance_end" "$stibctoken_balance_mid"
}

# check that redemptions and claims work
@test "[INTEGRATION-BASIC-$CHAIN_NAME] redemption and undelegation on $CHAIN_NAME" {
  # get initial balance of redemption ICA
  redemption_ica_balance_start=$(GET_BALANCE $CHAIN_NAME $(GET_ICA_ADDR $HOST_CHAIN_ID redemption) $HOST_DENOM)

  # call redeem-stake
  $STRIDE_MAIN_CMD tx stakeibc redeem-stake $REDEEM_AMOUNT $HOST_CHAIN_ID $HOST_RECEIVER_ADDRESS \
      --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # Check that the redemption record created from the autopilot redeem above was incremented
  # and that there is still only one record
  num_records=$($STRIDE_MAIN_CMD q records list-user-redemption-record | grep -c "native_token_amount")
  assert_equal "$num_records" "1"

  redemption_record_st_amount=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -Fiw 'st_token_amount' | head -n 1 | grep -o -E '[0-9]+')
  expected_record_minimum=$(echo "$REDEEM_AMOUNT * 2" | bc)
  assert_equal "$redemption_record_st_amount" "$expected_record_minimum"

  WAIT_FOR_STRING $STRIDE_LOGS "\[REDEMPTION] completed on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check that the tokens were transferred to the redemption account
  redemption_ica_balance_end=$(GET_BALANCE $CHAIN_NAME $(GET_ICA_ADDR $HOST_CHAIN_ID redemption) $HOST_DENOM)
  diff_positive=$(($redemption_ica_balance_end > $redemption_ica_balance_start))
  assert_equal "$diff_positive" "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] claim redeemed tokens" {
  # get balance before claim
  start_balance=$(GET_BALANCE $CHAIN_NAME $HOST_RECEIVER_ADDRESS $HOST_DENOM)

  # grab the epoch number for the first deposit record in the list od DRs
  EPOCH=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -Fiw 'epoch_number' | head -n 1 | grep -o -E '[0-9]+')

  # claim the record (send to stride address)
  $STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens $HOST_CHAIN_ID $EPOCH $HOST_RECEIVER_ADDRESS \
    --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y

  WAIT_FOR_STRING $STRIDE_LOGS "\[CLAIM\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check that the tokens were transferred to the sender account
  end_balance=$(GET_BALANCE $CHAIN_NAME $HOST_RECEIVER_ADDRESS $HOST_DENOM)

  # check that the undelegated tokens were transferred to the sender account
  diff_positive=$(($end_balance > $start_balance))
  assert_equal "$diff_positive" "1"
}

# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC-$CHAIN_NAME] rewards are being reinvested, exchange rate updating" {
  # check that the exchange rate has increased (i.e. redemption rate is greater than 1)
  MULT=1000000
  redemption_rate=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID | grep -Fiw 'redemption_rate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  redemption_rate_increased=$(( $(FLOOR $(DECMUL $redemption_rate $MULT)) > $(FLOOR $(DECMUL 1.00000000000000000 $MULT))))
  assert_equal "$redemption_rate_increased" "1"
}

