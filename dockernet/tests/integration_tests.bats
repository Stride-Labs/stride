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

  STRIDE_TRANFER_CHANNEL="channel-${TRANSFER_CHANNEL_NUMBER}"
  HOST_TRANSFER_CHANNEL="channel-0"

  TRANSFER_AMOUNT=5000000
  STAKE_AMOUNT=1000000
  REDEEM_AMOUNT=10000
  PACKET_FORWARD_STAKE_AMOUNT=30000

  GETBAL() {
    head -n 1 | grep -o -E '[0-9]+' || "0"
  }
  GETSTAKE() {
    tail -n 2 | head -n 1 | grep -o -E '[0-9]+' | head -n 1
  }
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
  refute_line '  delegation_account: null'
  refute_line '  fee_account: null'
  refute_line '  redemption_account: null'
  refute_line '  withdrawal_account: null'
  assert_line '  unbonding_frequency: "1"'
}

##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################


@test "[INTEGRATION-BASIC-$CHAIN_NAME] ibc transfer updates all balances" {
  # get initial balances
  sval_strd_balance_start=$($STRIDE_MAIN_CMD  q bank balances $(STRIDE_ADDRESS) --denom $STRIDE_DENOM   | GETBAL)
  hval_strd_balance_start=$($HOST_MAIN_CMD    q bank balances $HOST_VAL_ADDRESS --denom $IBC_STRD_DENOM | GETBAL)
  sval_token_balance_start=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  hval_token_balance_start=$($HOST_MAIN_CMD   q bank balances $HOST_VAL_ADDRESS --denom $HOST_DENOM     | GETBAL)

  # do IBC transfer
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer $STRIDE_TRANFER_CHANNEL $HOST_VAL_ADDRESS ${TRANSFER_AMOUNT}${STRIDE_DENOM} --from $STRIDE_VAL -y 
  $HOST_MAIN_CMD   tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL  $(STRIDE_ADDRESS) ${TRANSFER_AMOUNT}${HOST_DENOM} --from $HOST_VAL -y 

  WAIT_FOR_BLOCK $STRIDE_LOGS 8

  # get new balances
  sval_strd_balance_end=$($STRIDE_MAIN_CMD  q bank balances $(STRIDE_ADDRESS) --denom $STRIDE_DENOM   | GETBAL)
  hval_strd_balance_end=$($HOST_MAIN_CMD    q bank balances $HOST_VAL_ADDRESS --denom $IBC_STRD_DENOM | GETBAL)
  sval_token_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  hval_token_balance_end=$($HOST_MAIN_CMD   q bank balances $HOST_VAL_ADDRESS --denom $HOST_DENOM     | GETBAL)

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
  token_balance_start=$($STRIDE_MAIN_CMD   q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  sttoken_balance_start=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM   | GETBAL)

  # get initial ICA accound balance
  delegation_address=$(GET_ICA_ADDR $HOST_CHAIN_ID delegation)
  delegation_ica_balance_start=$($HOST_MAIN_CMD q bank balances $delegation_address --denom $HOST_DENOM | GETBAL)

  # make sure we're not on the boundary of an epoch, otherwise the redemption rate could change mid-test
  WAIT_UNTIL_MIDDLE_OF_EPOCH

  # get redemption rate
  redemption_rate=$(GET_REDEMPTION_RATE $HOST_CHAIN_ID)

  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake $STAKE_AMOUNT $HOST_DENOM --from $STRIDE_VAL -y 

  # wait for the stTokens to get minted 
  WAIT_FOR_BALANCE_CHANGE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM 

  # make sure IBC_DENOM went down
  token_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  token_balance_diff=$(($token_balance_start - $token_balance_end))
  assert_equal "$token_balance_diff" $STAKE_AMOUNT

  # make sure stToken went up 
  expected_sttoken_amount=$(echo "scale=0; $STAKE_AMOUNT / $redemption_rate" | bc -l)
  sttoken_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" "$expected_sttoken_amount"

  # Wait for the transfer to complete 
  # The transfer must be at least as large as the stake
  WAIT_FOR_BALANCE_CHANGE $CHAIN_NAME $delegation_address $HOST_DENOM $STAKE_AMOUNT

  # Check that the delegation ICA balance went up by **at least** the stake amount
  # It could have increased by more than the stake amount due to reinvestment
  delegation_ica_balance_end=$($HOST_MAIN_CMD q bank balances $delegation_address --denom $HOST_DENOM | GETBAL)
  delegation_balance_diff=$(($delegation_ica_balance_end - $delegation_ica_balance_start))
  assert_equal "$(echo "$delegation_balance_diff - $STAKE_AMOUNT >= 0" | bc -l)" "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] packet forwarding automatically liquid stakes" {
  skip "DefaultActive set to false, skip test"
  memo='{ "autopilot": { "receiver": "'"$(STRIDE_ADDRESS)"'",  "stakeibc": { "stride_address": "'"$(STRIDE_ADDRESS)"'", "action": "LiquidStake" } } }'

  # get initial balances
  sttoken_balance_start=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)

  # make sure we're not on the boundary of an epoch, otherwise the redemption rate could change mid-test
  WAIT_UNTIL_MIDDLE_OF_EPOCH
  
  # get redemption rate
  redemption_rate=$(GET_REDEMPTION_RATE $HOST_CHAIN_ID)

  # Send the IBC transfer with the JSON memo
  transfer_msg_prefix="$HOST_MAIN_CMD tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL"
  if [[ "$CHAIN_NAME" == "GAIA" ]]; then
    # For GAIA (ibc-v3), pass the memo into the receiver field
    $transfer_msg_prefix "$memo" ${PACKET_FORWARD_STAKE_AMOUNT}${HOST_DENOM} --from $HOST_VAL -y 
  elif [[ "$CHAIN_NAME" == "HOST" ]]; then
    # For HOST (ibc-v5), pass an address for a receiver and the memo in the --memo field
    $transfer_msg_prefix $(STRIDE_ADDRESS) ${PACKET_FORWARD_STAKE_AMOUNT}${HOST_DENOM} --memo "$memo" --from $HOST_VAL -y 
  else
    # For all other hosts, skip this test
    skip "Packet forward liquid stake test is only run on GAIA and HOST"
  fi

  # Wait for the transfer to complete
  WAIT_FOR_BALANCE_CHANGE STRIDE $(STRIDE_ADDRESS) st$HOST_DENOM

  # make sure stToken went up (relax precision in case the redemption rate changed)
  expected_sttoken_amount=$(echo "$redemption_rate * $STAKE_AMOUNT / 1" | bc)
  sttoken_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal $(COMPARE_WITH_PRECISION $sttoken_balance_diff $expected_sttoken_amount) 
}

# check that tokens on the host are staked
@test "[INTEGRATION-BASIC-$CHAIN_NAME] tokens on $CHAIN_NAME were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS "\[DELEGATION\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 4

  # check staked tokens
  NEW_STAKE=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $CHAIN_NAME 1) | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

# check that redemptions and claims work
@test "[INTEGRATION-BASIC-$CHAIN_NAME] redemption works" {
  # get initial balance of redemption ICA
  redemption_ica_balance_start=$($HOST_MAIN_CMD q bank balances $(GET_ICA_ADDR $HOST_CHAIN_ID redemption) --denom $HOST_DENOM | GETBAL)

  # call redeem-stake
  $STRIDE_MAIN_CMD tx stakeibc redeem-stake $REDEEM_AMOUNT $HOST_CHAIN_ID $HOST_RECEIVER_ADDRESS \
      --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y

  WAIT_FOR_STRING $STRIDE_LOGS "\[REDEMPTION] completed on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check that the tokens were transferred to the redemption account
  redemption_ica_balance_end=$($HOST_MAIN_CMD q bank balances $(GET_ICA_ADDR $HOST_CHAIN_ID redemption) --denom $HOST_DENOM | GETBAL)
  diff_positive=$(($redemption_ica_balance_end > $redemption_ica_balance_start))
  assert_equal "$diff_positive" "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] claimed tokens are properly distributed" {
  # get balance before claim
  start_balance=$($HOST_MAIN_CMD q bank balances $HOST_RECEIVER_ADDRESS --denom $HOST_DENOM | GETBAL)

  # grab the epoch number from the latest epoch unbonding record (determined by sorting by epoch number)
  epoch=$($STRIDE_MAIN_CMD q records list-user-redemption-record \
    | grep epoch_number | awk '{print $2}' | tr -cd '[:digit:]\n' | sort -n | tail -n 1 | grep -o -E '[0-9]+')

  # claim the record (send to stride address)
  $STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens $HOST_CHAIN_ID $epoch $(STRIDE_ADDRESS) \
    --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y

  WAIT_FOR_STRING $STRIDE_LOGS "\[CLAIM\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check that the tokens were transferred to the sender account
  end_balance=$($HOST_MAIN_CMD q bank balances $HOST_RECEIVER_ADDRESS --denom $HOST_DENOM | GETBAL)

  # check that the undelegated tokens were transfered to the sender account
  diff_positive=$(($end_balance > $start_balance))
  assert_equal "$diff_positive" "1"
}

# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC-$CHAIN_NAME] rewards are being reinvested, exchange rate updating" {
  # TODO: Store the redemption rate from a prior test
  # store the initial redemption rate
  redemption_rate_start=$(GET_REDEMPTION_RATE $HOST_CHAIN_ID)

  # sleep 1 epoch
  sleep "${STRIDE_EPOCH_EPOCH_DURATION%s}" # remove "s" from end of timestamp

  # check that the exchange rate has increased (i.e. redemption rate is greater than 1)
  redemption_rate_end=$(GET_REDEMPTION_RATE $HOST_CHAIN_ID)
  assert_equal "$(echo "$redemption_rate_end > $redemption_rate_start" | bc -l)" "1"
}

# rewards have been collected and distributed to strd stakers
@test "[INTEGRATION-BASIC-$CHAIN_NAME] rewards are being distributed to stakers" {
  # get initial sttoken balance of the 2nd validator
  val_address=$($STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}2 --keyring-backend test -a)
  sttoken_balance_start=$($STRIDE_MAIN_CMD q bank balances $val_address --denom st$HOST_DENOM | GETBAL)

  # collect the validator's outstanding rewards
  $STRIDE_MAIN_CMD tx distribution withdraw-all-rewards --from ${STRIDE_VAL_PREFIX}2 -y 
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # confirm they've recieved stTokens
  sttoken_balance_end=$($STRIDE_MAIN_CMD q bank balances $val_address --denom st$HOST_DENOM | GETBAL)
  assert_equal "$(echo "$sttoken_balance_end > $sttoken_balance_start" | bc -l)" "1"
}
