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
  STRIDE_VAL2=${STRIDE_VAL_PREFIX}2
  STRIDE_VAL3=${STRIDE_VAL_PREFIX}3

  STRIDE_TRANFER_CHANNEL="channel-${TRANSFER_CHANNEL_NUMBER}"
  HOST_TRANSFER_CHANNEL="channel-0"

  TRANSFER_AMOUNT=500000
  STAKE_AMOUNT=100000
  REDEEM_AMOUNT=1000

  ALLIANCE_STAKE_AMOUNT=10000
  REWARD_WEIGHT=1
  TAKE_RATE=0
  REWARD_CHANGE_RATE=1
  REWARD_CHANGE_INTERVAL=3600s
  PROPOSAL_ID=1

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
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer $STRIDE_TRANFER_CHANNEL $HOST_VAL_ADDRESS ${TRANSFER_AMOUNT}${STRIDE_DENOM} --from $STRIDE_VAL -y &
  $HOST_MAIN_CMD   tx ibc-transfer transfer transfer $HOST_TRANSFER_CHANNEL  $(STRIDE_ADDRESS) ${TRANSFER_AMOUNT}${HOST_DENOM} --from $HOST_VAL -y &

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

  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake $STAKE_AMOUNT $HOST_DENOM --from $STRIDE_VAL -y 

  # sleep two block for the tx to settle on stride
  WAIT_FOR_STRING $STRIDE_LOGS "\[MINT ST ASSET\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # make sure IBC_DENOM went down
  token_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $HOST_IBC_DENOM | GETBAL)
  token_balance_diff=$(($token_balance_start - $token_balance_end))
  assert_equal "$token_balance_diff" $STAKE_AMOUNT

  # make sure stToken went up
  sttoken_balance_end=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  sttoken_balance_diff=$(($sttoken_balance_end-$sttoken_balance_start))
  assert_equal "$sttoken_balance_diff" $STAKE_AMOUNT

  # Wait for the transfer to complete
  WAIT_FOR_BALANCE_CHANGE $CHAIN_NAME $delegation_address $HOST_DENOM 

  # get the new delegation ICA balance
  delegation_ica_balance_end=$($HOST_MAIN_CMD q bank balances $delegation_address --denom $HOST_DENOM | GETBAL)
  diff=$(($delegation_ica_balance_end - $delegation_ica_balance_start))
  assert_equal "$diff" $STAKE_AMOUNT
}

# check that tokens on the host are staked
@test "[INTEGRATION-BASIC-$CHAIN_NAME] tokens on $CHAIN_NAME were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS "\[DELEGATION\] success on $HOST_CHAIN_ID"
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check staked tokens
  NEW_STAKE=$($HOST_MAIN_CMD q staking delegation $(GET_ICA_ADDR $HOST_CHAIN_ID delegation) $(GET_VAL_ADDR $CHAIN_NAME 1) | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] submit a proposal for alliance asset registration" {
  # submit proposal for asset registration
  $STRIDE_MAIN_CMD tx gov submit-legacy-proposal create-alliance st$HOST_DENOM $REWARD_WEIGHT $TAKE_RATE $REWARD_CHANGE_RATE $REWARD_CHANGE_INTERVAL --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # query proposal confirmation
  $STRIDE_MAIN_CMD query gov proposals
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # deposit
  $STRIDE_MAIN_CMD tx gov deposit $PROPOSAL_ID 10000001ustrd --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # deposit confirmation
  $STRIDE_MAIN_CMD query gov proposals $PROPOSAL_ID
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # voting
  $STRIDE_MAIN_CMD tx gov vote $PROPOSAL_ID yes --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  $STRIDE_MAIN_CMD tx gov vote $PROPOSAL_ID yes --from $STRIDE_VAL2 --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  $STRIDE_MAIN_CMD tx gov vote $PROPOSAL_ID yes --from $STRIDE_VAL3 --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  
  sleep 60

  # vote confirmation
  vote_status=$($STRIDE_MAIN_CMD query gov proposal $PROPOSAL_ID | grep "status" | awk '{printf $2}')
  assert_equal "$vote_status" "PROPOSAL_STATUS_PASSED"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] delegate, redelegate alliance assets" {
  # delegate stTokens to val1
  $STRIDE_MAIN_CMD tx alliance delegate stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f "$ALLIANCE_STAKE_AMOUNT"st$HOST_DENOM --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_BLOCK $STRIDE_LOGS 4

  # check delegation amount
  delegated_amount=$($STRIDE_MAIN_CMD query alliance delegations | grep -Fiw 'amount' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  assert_equal "$delegated_amount" $ALLIANCE_STAKE_AMOUNT
  sleep 60

  # check reward amount
  reward_amount=$($STRIDE_MAIN_CMD query alliance rewards $(STRIDE_ADDRESS) stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f st$HOST_DENOM | grep -Fiw 'amount' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  diff_positive=$(($reward_amount > 0))
  assert_equal "$diff_positive" "1"

  # redelegate half staked stTokens from val1 to val2
  redelegation_amount=$(($ALLIANCE_STAKE_AMOUNT / 2))
  $STRIDE_MAIN_CMD tx alliance redelegate stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm "$redelegation_amount"st$HOST_DENOM --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y --gas auto
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check redelegated amount
  redelegated_amount=$($STRIDE_MAIN_CMD query alliance delegations --output json | jq '.delegations[1].balance.amount' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  assert_equal $redelegation_amount $redelegated_amount
  sleep 60

  # check reward amount
  reward_amount=$($STRIDE_MAIN_CMD query alliance rewards $(STRIDE_ADDRESS) stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm st$HOST_DENOM | grep -Fiw 'amount' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  diff_positive=$(($reward_amount > 0))
  assert_equal "$diff_positive" "1"
}


@test "[INTEGRATION-BASIC-$CHAIN_NAME] claim rewards, undelegate alliance assets" {
  # undelegate stTokens from val2 (undelegating also claims rewards)
  st_token_start_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  native_token_start_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STRD_DENOM | GETBAL)
  undelegation_amount=$($STRIDE_MAIN_CMD query alliance delegations --output json | jq '.delegations[1].balance.amount' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')

  $STRIDE_MAIN_CMD tx alliance undelegate stridevaloper1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrgpwsqm "$undelegation_amount"st$HOST_DENOM --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  sleep 130

  st_token_end_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  native_token_end_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STRD_DENOM | GETBAL)

  # get stToken and native token balance diffs
  st_token_balance_diff=$(($st_token_end_balance - $st_token_start_balance))
  native_token_balance_diff=$(($native_token_end_balance - $native_token_start_balance))
  diff_positive=$(($native_token_balance_diff > 0))
  assert_equal "$st_token_balance_diff" "$undelegation_amount"
  assert_equal "$diff_positive" "1"

  # claim rewards from val1
  st_token_start_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  native_token_start_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STRD_DENOM | GETBAL)

  $STRIDE_MAIN_CMD tx alliance claim-rewards stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f st$HOST_DENOM --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y --gas auto
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # undelegate stTokens from val1
  undelegation_amount=$($STRIDE_MAIN_CMD query alliance delegations --output json | jq '.delegations[0].balance.amount' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  $STRIDE_MAIN_CMD tx alliance undelegate stridevaloper1nnurja9zt97huqvsfuartetyjx63tc5zrj5x9f "$undelegation_amount"st$HOST_DENOM --from $STRIDE_VAL --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  sleep 130

  st_token_end_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom st$HOST_DENOM | GETBAL)
  native_token_end_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STRD_DENOM | GETBAL)

  # get stToken and native token balance diffs
  st_token_balance_diff=$(($st_token_end_balance - $st_token_start_balance))
  native_token_balance_diff=$(($native_token_end_balance - $native_token_start_balance))
  diff_positive=$(($native_token_balance_diff > 0))
  assert_equal "$st_token_balance_diff" "$undelegation_amount"
  assert_equal "$diff_positive" "1"
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
  # check that the exchange rate has increased (i.e. redemption rate is greater than 1)
  MULT=1000000
  redemption_rate=$($STRIDE_MAIN_CMD q stakeibc show-host-zone $HOST_CHAIN_ID | grep -Fiw 'redemption_rate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  redemption_rate_increased=$(( $(FLOOR $(DECMUL $redemption_rate $MULT)) > $(FLOOR $(DECMUL 1.00000000000000000 $MULT))))
  assert_equal "$redemption_rate_increased" "1"
}

@test "[INTEGRATION-BASIC-$CHAIN_NAME] revenue accrued, and clear-balance works" {
  # confirm the fee account has accrued revenue
  fee_ica_balance=$($HOST_MAIN_CMD q bank balances $(GET_ICA_ADDR $HOST_CHAIN_ID fee) --denom $HOST_DENOM | GETBAL)
  fee_ica_balance_positive=$(($fee_ica_balance > 0))
  assert_equal "$fee_ica_balance_positive" "1"

  # call clear balance (with amount = 1)
  $STRIDE_MAIN_CMD tx stakeibc clear-balance $HOST_CHAIN_ID 1 $HOST_TRANSFER_CHANNEL --from $STRIDE_ADMIN_ACCT -y
  WAIT_FOR_BLOCK $STRIDE_LOGS 8

  # check that balance went to revenue account
  fee_stride_balance=$($STRIDE_MAIN_CMD q bank balances $STRIDE_FEE_ADDRESS --denom $HOST_IBC_DENOM | GETBAL)
  assert_equal "$fee_stride_balance" "1"
}
