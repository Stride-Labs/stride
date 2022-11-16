#!/usr/bin/env bats

setup_file() {
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

  # set allows us to export all variables in account_vars
  set -a
  source scripts/vars.sh

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

@test "[INTEGRATION-BASIC] address names are correct" {
  assert_equal $(STRIDE_ADDRESS) "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

  assert_equal $JUNO_DELEGATE_VAL "junovaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuued3knlr0"
  assert_equal $JUNO_DELEGATION_ICA_ADDR 'juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35'
  assert_equal $JUNO_REDEMPTION_ICA_ADDR 'juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9'
  assert_equal $JUNO_WITHDRAWAL_ICA_ADDR 'juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5'
  assert_equal $JUNO_FEE_ICA_ADDR 'juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h'
}

# # add test to register host zone
@test "[INTEGRATION-BASIC] host zones successfully registered" {
  run $STRIDE_MAIN_CMD q stakeibc show-host-zone JUNO
  assert_line '  HostDenom: ujuno'
  assert_line '  chainId: JUNO'
  assert_line '  delegationAccount:'
  assert_line '    address: juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35'
  assert_line '  feeAccount:'
  assert_line '    address: juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h'
  assert_line '  redemptionAccount:'
  assert_line '    address: juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9'
  assert_line '  withdrawalAccount:'
  assert_line '    address: juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5'
  assert_line '  unbondingFrequency: "1"'
  assert_line '  RedemptionRate: "1.000000000000000000"'
}


##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################

@test "[INTEGRATION-BASIC-JUNO] ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom ustrd | GETBAL)
  juno1_balance=$($JUNO_MAIN_CMD q bank balances $(JUNO_ADDRESS) --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_juno=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_JUNO_DENOM | GETBAL)
  juno1_balance_juno=$($JUNO_MAIN_CMD q bank balances $(JUNO_ADDRESS) --denom ujuno | GETBAL)
  # do IBC transfer
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-1 $(JUNO_ADDRESS) 100000000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
  $JUNO_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 100000000ujuno --from jval1 --chain-id JUNO -y --keyring-backend test
  WAIT_FOR_BLOCK $STRIDE_LOGS 8
  # get new balances
  str1_balance_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom ustrd | GETBAL)
  juno1_balance_new=$($JUNO_MAIN_CMD q bank balances $(JUNO_ADDRESS) --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_juno_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_JUNO_DENOM | GETBAL)
  juno1_balance_juno_new=$($JUNO_MAIN_CMD q bank balances $(JUNO_ADDRESS) --denom ujuno | GETBAL)
  # get all STRD balance diffs
  str1_diff=$(($str1_balance - $str1_balance_new))
  juno1_diff=$(($juno1_balance - $juno1_balance_new))
  assert_equal "$str1_diff" '100000000'
  assert_equal "$juno1_diff" '-100000000'
  # get all JUNO_DENOM balance diffs
  str1_diff=$(($str1_balance_juno - $str1_balance_juno_new))
  juno1_diff=$(($juno1_balance_juno - $juno1_balance_juno_new))
  assert_equal "$str1_diff" '-100000000'
  assert_equal "$juno1_diff" '100000000'
}

@test "[INTEGRATION-BASIC-JUNO] liquid stake mints stJUNO" {
  # get module address
  MODADDR=$($STRIDE_MAIN_CMD q stakeibc module-address stakeibc | awk '{print $NF}')
  # get initial balances
  mod_balance_juno=$($STRIDE_MAIN_CMD q bank balances $MODADDR --denom $IBC_JUNO_DENOM | GETBAL)
  str1_balance_juno=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_JUNO_DENOM | GETBAL)
  str1_balance_stjuno=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STJUNO_DENOM | GETBAL)
  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 10000000 ujuno --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID
  # sleep two block for the tx to settle on stride
  WAIT_FOR_STRING $STRIDE_LOGS '\[MINT ST ASSET\] success on JUNO'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # make sure IBC_JUNO_DENOM went down
  str1_balance_juno_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_JUNO_DENOM | GETBAL)
  str1_juno_diff=$(($str1_balance_juno - $str1_balance_juno_new))
  assert_equal "$str1_juno_diff" '10000000'
  # make sure STJUNO went up
  str1_balance_stjuno_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STJUNO_DENOM | GETBAL)
  str1_stjuno_diff=$(($str1_balance_stjuno_new-$str1_balance_stjuno))
  assert_equal "$str1_stjuno_diff" "10000000"
}

@test "[INTEGRATION-BASIC-JUNO] tokens were transferred to JUNO after liquid staking" {
  # initial balance of delegation ICA
  initial_delegation_ica_bal=$($JUNO_MAIN_CMD q bank balances $JUNO_DELEGATION_ICA_ADDR --denom ujuno | GETBAL)
  WAIT_FOR_STRING $STRIDE_LOGS '\[IBC-TRANSFER\] success to JUNO'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # get the new delegation ICA balance
  post_delegation_ica_bal=$($JUNO_MAIN_CMD q bank balances $JUNO_DELEGATION_ICA_ADDR --denom ujuno | GETBAL)
  diff=$(($post_delegation_ica_bal - $initial_delegation_ica_bal))
  assert_equal "$diff" '10000000'
}

@test "[INTEGRATION-BASIC-JUNO] tokens on JUNO were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS '\[DELEGATION\] success on JUNO'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # check staked tokens
  NEW_STAKE=$($JUNO_MAIN_CMD q staking delegation $JUNO_DELEGATION_ICA_ADDR $JUNO_DELEGATE_VAL | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

# check that redemptions and claims work
@test "[INTEGRATION-BASIC-JUNO] redemption works" {
  sleep 5
  old_redemption_ica_bal=$($JUNO_MAIN_CMD q bank balances $JUNO_REDEMPTION_ICA_ADDR --denom ujuno | GETBAL)
  # call redeem-stake
  amt_to_redeem=100
  $STRIDE_MAIN_CMD tx stakeibc redeem-stake $amt_to_redeem JUNO $JUNO_RECEIVER_ACCT \
      --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[REDEMPTION] completed on JUNO'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # check that the tokens were transferred to the redemption account
  new_redemption_ica_bal=$($JUNO_MAIN_CMD q bank balances $JUNO_REDEMPTION_ICA_ADDR --denom ujuno | GETBAL)
  diff_positive=$(($new_redemption_ica_bal > $old_redemption_ica_bal))
  assert_equal "$diff_positive" "1"
}

@test "[INTEGRATION-BASIC-JUNO] claimed tokens are properly distributed" {
  # TODO(optimize tests) extra sleep just in case
  SENDER_ACCT=$(STRIDE_ADDRESS)
  old_sender_bal=$($JUNO_MAIN_CMD q bank balances $JUNO_RECEIVER_ACCT --denom ujuno | GETBAL)
  # TODO check that the UserRedemptionRecord has isClaimable = true
  # grab the epoch number for the first deposit record in the list od DRs
  EPOCH=$(strided q records list-user-redemption-record  | grep -Fiw 'epochNumber' | head -n 1 | grep -o -E '[0-9]+')
  # claim the record
  $STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens JUNO $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id STRIDE -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[CLAIM\] success on JUNO'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # TODO check that UserRedemptionRecord has isClaimable = false
  # check that the tokens were transferred to the sender account
  new_sender_bal=$($JUNO_MAIN_CMD q bank balances $JUNO_RECEIVER_ACCT --denom ujuno | GETBAL)
  # check that the undelegated tokens were transfered to the sender account
  diff_positive=$(($new_sender_bal > $old_sender_bal))
  assert_equal "$diff_positive" "1"
}

# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC-JUNO] rewards are being reinvested, exchange rate updating" {
  # read the exchange rate and current delegations
  RR1=$($STRIDE_MAIN_CMD q stakeibc show-host-zone JUNO | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  OLD_STAKED_BAL=$($JUNO_MAIN_CMD q staking delegation $JUNO_DELEGATION_ICA_ADDR $JUNO_DELEGATE_VAL | GETSTAKE)
  # liquid stake again to kickstart the reinvestment process
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 ujuno --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # wait four days (transfers, stake, move rewards, reinvest rewards)
  epoch_duration=$($STRIDE_MAIN_CMD q epochs epoch-infos | grep -Fiw -B 2 'stride_epoch' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($epoch_duration * 4))
  # simple check that number of tokens staked increases
  NEW_STAKED_BAL=$($JUNO_MAIN_CMD q staking delegation $JUNO_DELEGATION_ICA_ADDR $JUNO_DELEGATE_VAL | GETSTAKE)
  STAKED_BAL_INCREASED=$(($NEW_STAKED_BAL > $OLD_STAKED_BAL))
  assert_equal "$STAKED_BAL_INCREASED" "1"

  RR2=$($STRIDE_MAIN_CMD q stakeibc show-host-zone JUNO | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  # check that the exchange rate has increased
  MULT=1000000
  RR_INCREASED=$(( $(FLOOR $(DECMUL $RR2 $MULT)) > $(FLOOR $(DECMUL $RR1 $MULT))))
  assert_equal "$RR_INCREASED" "1"
}

# TODO check that the correct amount is being reinvested and the correct amount is flowing to the rev EOA
@test "[NOT-IMPLEMENTED] reinvestment and revenue amounts are correct" {
}
