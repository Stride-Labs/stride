#!/usr/bin/env bats

setup_file() {
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

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
  source scripts/vars.sh

  HOST_CHAIN_ID=$(GET_VAR_VALUE  ${CHAIN_NAME}_CHAIN_ID)
  HOST_DENOM=$(GET_VAR_VALUE     ${CHAIN_NAME}_DENOM)
  HOST_IBC_DENOM=$(GET_VAR_VALUE IBC_${CHAIN_NAME}_CHANNEL_${TRANSFER_CHANNEL_NUMBER}_DENOM)
  HOST_MAIN_CMD=$(GET_VAR_VALUE  ${CHAIN_NAME}_MAIN_CMD)

  HOST_VAL_ADDRESS=$(${CHAIN_NAME}_ADDRESS)
  HOST_RECEIVER_ADDRESS=$(GET_VAR_VALUE  ${CHAIN_NAME}_RECEIVER_ADDRESS)

  HOST_VAL="$(GET_VAR_VALUE ${CHAIN_NAME}_VAL_PREFIX)1"
  STRIDE_VAL=${STRIDE_VAL_PREFIX}1

  STRIDE_TRANFER_CHANNEL="channel-${TRANSFER_CHANNEL_NUMBER}"
  HOST_TRANSFER_CHANNEL="channel-0"

  GETBAL() {
    head -n 1 | grep -o -E '[0-9]+'
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

setup() {
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

  # if these extensions don't load properly, adjust the paths accoring to these instructions
  TEST_BREW_PREFIX="$(brew --prefix)"
  load "${TEST_BREW_PREFIX}/lib/bats-support/load.bash"
  load "${TEST_BREW_PREFIX}/lib/bats-assert/load.bash"
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
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer $STRIDE_TRANFER_CHANNEL $HOST_VAL_ADDRESS 3000${STRIDE_DENOM} --from $STRIDE_VAL --chain-id $STRIDE_CHAIN_ID -y --keyring-backend test &
  $HOST_MAIN_CMD   tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL  $(STRIDE_ADDRESS) 3000${HOST_DENOM}   --from $HOST_VAL   --chain-id $HOST_CHAIN_ID   -y --keyring-backend test &
  WAIT_FOR_BLOCK $STRIDE_LOGS 8

  # get new balances
  sval_strd_balance_end=$($STRIDE_MAIN_CMD  q bank balances $(STRIDE_ADDRESS) --denom $STRIDE_DENOM   | GETBAL)
  hval_strd_balance_end=$($HOST_MAIN_CMD    q bank balances $HOST_VAL_ADDRESS --denom $IBC_STRD_DENOM | GETBAL)
  sval_token_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  hval_token_balance_end=$($HOST_MAIN_CMD   q bank balances $HOST_VAL_ADDRESS --denom $HOST_DENOM     | GETBAL)

  # get all STRD balance diffs
  sval_strd_balance_diff=$(($sval_strd_balance_start - $sval_strd_balance_end))
  hval_strd_balance_diff=$(($hval_strd_balance_start - $hval_strd_balance_end))
  assert_equal "$sval_strd_balance_diff" '3000'
  assert_equal "$hval_strd_balance_diff" '-3000'

  # get all host balance diffs
  sval_token_balance_diff=$(($sval_token_balance_start - $sval_token_balance_end))
  hval_token_balance_diff=$(($hval_token_balance_start - $hval_token_balance_end))
  assert_equal "$sval_token_balance_diff" '-3000'
  assert_equal "$hval_token_balance_diff" '3000'
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] liquid stake mints stToken" {
  # get initial balances
  token_balance_start=$($STRIDE_MAIN_CMD   q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  sttoken_balance_start=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM   | GETBAL)

  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 $HOST_DENOM --keyring-backend test --from $STRIDE_VAL -y --chain-id $STRIDE_CHAIN_ID

  # sleep two block for the tx to settle on stride
  WAIT_FOR_STRING $STRIDE_LOGS "\[MINT ST ASSET\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # make sure IBC_DENOM went down
  token_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  token_balance_diff=$(($token_balance_start - $token_balance_end))
  assert_equal "$token_balance_diff" '1000'

  # make sure stToken went up
  sttoken_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" "1000"
}

# check that tokens were transferred to host after liquid stake
@test "[INTEGRATION-BASIC-$CHAIN_NAME] tokens were transferred to $HOST_CHAIN_ID after liquid staking" {
  # initial balance of delegation ICA
  delegation_ica_balance_start=$($HOST_MAIN_CMD q bank balances $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) --denom $HOST_DENOM | GETBAL)
  WAIT_FOR_STRING $STRIDE_LOGS "\[IBC-TRANSFER\] success to $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # get the new delegation ICA balance
  delegation_ica_balance_end=$($HOST_MAIN_CMD q bank balances $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) --denom $HOST_DENOM | GETBAL)
  diff=$(($delegation_ica_balance_end - $delegation_ica_balance_start))
  assert_equal "$diff" '1000'
}

# check that tokens on the host are staked
@test "[INTEGRATION-BASIC-$CHAIN_NAME] tokens on $CHAIN_NAME were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS "\[DELEGATION\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check staked tokens
  NEW_STAKE=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $HOST_CHAIN_ID 1) | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

# check that redemptions and claims work
@test "[INTEGRATION-BASIC-$CHAIN_NAME] redemption works" {
  # get initial balance of redemption ICA
  redemption_ica_balance_start=$($HOST_MAIN_CMD q bank balances $(GET_ICA_ADDR $HOST_CHAIN_ID redemption) --denom $HOST_DENOM | GETBAL)

  # call redeem-stake
  amt_to_redeem=100
  $STRIDE_MAIN_CMD tx stakeibc redeem-stake $amt_to_redeem $HOST_CHAIN_ID $HOST_RECEIVER_ADDRESS \
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

  # grab the epoch number for the first deposit record in the list od DRs
  EPOCH=$($STRIDE_MAIN_CMD q records list-user-redemption-record  | grep -Fiw 'epoch_number' | head -n 1 | grep -o -E '[0-9]+')

  # claim the record (send to stride address)
  $STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens $HOST_CHAIN_ID $EPOCH $(STRIDE_ADDRESS) \
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
  # read the exchange rate and current delegations
  redemption_rate_start=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID | grep -Fiw 'redemption_rate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  staked_balance_start=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $HOST_CHAIN_ID 1) | GETSTAKE)

  # liquid stake again to kickstart the reinvestment process
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 $HOST_DENOM --keyring-backend test --from $STRIDE_VAL -y --chain-id $STRIDE_CHAIN_ID
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # wait four days (transfers, stake, move rewards, reinvest rewards)
  epoch_duration=$($STRIDE_MAIN_CMD q epochs epoch-infos | grep -Fiw -B 2 'stride_epoch' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($epoch_duration * 4))

  # simple check that number of tokens staked increases
  staked_balance_end=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $HOST_CHAIN_ID 1) | GETSTAKE)
  staked_balance_increased=$(($staked_balance_end > $staked_balance_start))
  assert_equal "$staked_balance_increased" "1"

  # check that the exchange rate has increased
  MULT=1000000
  redemption_rate_end=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID | grep -Fiw 'redemption_rate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  redemption_rate_increased=$(( $(FLOOR $(DECMUL $redemption_rate_end $MULT)) > $(FLOOR $(DECMUL $redemption_rate_start $MULT))))
  assert_equal "$redemption_rate_increased" "1"
}

# TODO check that the correct amount is being reinvested and the correct amount is flowing to the rev EOA
@test "[NOT-IMPLEMENTED] reinvestment and revenue amounts are correct" {
}
