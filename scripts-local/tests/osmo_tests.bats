#!/usr/bin/env bats

setup_file() {
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

  # set allows us to export all variables in account_vars
  set -a
  source scripts-local/account_vars.sh

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
  assert_equal $STRIDE_VAL_ADDR "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7"

  assert_equal $GAIA_VAL_ADDR "cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2"
  assert_equal $GAIA_VAL_2_ADDR "cosmos133lfs9gcpxqj6er3kx605e3v9lqp2pg54sreu3"
  assert_equal $GAIA_VAL_3_ADDR "cosmos1fumal3j4lxzjp22fzffge8mw56qm33h9ez0xy2"

  assert_equal $HERMES_STRIDE_ADDR "stride1ft20pydau82pgesyl9huhhux307s9h3078692y"
  assert_equal $ICQ_STRIDE_ADDR "stride12vfkpj7lpqg0n4j68rr5kyffc6wu55dzqewda4"

  assert_equal $DELEGATION_ICA_ADDR "cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm"
  assert_equal $REDEMPTION_ICA_ADDR "cosmos1xmcwu75s8v7s54k79390wc5gwtgkeqhvzegpj0nm2tdwacv47tmqg9ut30"
  assert_equal $WITHDRAWAL_ICA_ADDR "cosmos1x5p8er7e2ne8l54tx33l560l8djuyapny55pksctuguzdc00dj7saqcw2l"
  assert_equal $REVENUE_EOA_ADDR "cosmos1wdplq6qjh2xruc7qqagma9ya665q6qhcwju3ng"
  assert_equal $FEE_ICA_ADDR "cosmos1lkgt5sfshn9shm7hd7chtytkq4mvwvswgmyl0hkacd4rmusu9wwq60cezx"
  assert_equal $GAIA_DELEGATE_VAL "cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne"
  assert_equal $GAIA_DELEGATE_VAL_2 "cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz"
  assert_equal $GAIA_RECEIVER_ACCT "cosmos1g6qdx6kdhpf000afvvpte7hp0vnpzapuyxp8uf"

  assert_equal $JUNO_DELEGATE_VAL "junovaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuued3knlr0"
  assert_equal $JUNO_DELEGATION_ICA_ADDR 'juno1xan7vt4nurz6c7x0lnqnvpmuc0lljz7rycqmuz2kk6wxv4k69d0sfats35'
  assert_equal $JUNO_REDEMPTION_ICA_ADDR 'juno1y6haxdt03cgkc7aedxrlaleeteel7fgc0nvtu2kggee3hnrlvnvs4kw2v9'
  assert_equal $JUNO_WITHDRAWAL_ICA_ADDR 'juno104n6h822n6n7psqjgjl7emd2uz67lptggp5cargh6mw0gxpch2gsk53qk5'
  assert_equal $JUNO_FEE_ICA_ADDR 'juno1rp8qgfq64wmjg7exyhjqrehnvww0t9ev3f3p2ls82umz2fxgylqsz3vl9h'
}

# # add test to register host zone
@test "[INTEGRATION-BASIC] host zones successfully registered" {

  run $STRIDE_CMD q stakeibc show-host-zone OSMO
  assert_line '  HostDenom: uosmo'
  assert_line '  chainId: OSMO'
  assert_line '  delegationAccount:'
  assert_line '    address: osmo1cx04p5974f8hzh2lqev48kjrjugdxsxy7mzrd0eyweycpr90vk8q8d6f3h'
  assert_line '  feeAccount:'
  assert_line '    address: osmo1n4r77qsmu9chvchtmuqy9cv3s539q87r398l6ugf7dd2q5wgyg9su3wd4g'
  assert_line '  redemptionAccount:'
  assert_line '    address: osmo1uy9p9g609676rflkjnnelaxatv8e4sd245snze7qsxzlk7dk7s8qrcjaez'
  assert_line '  withdrawalAccount:'
  assert_line '    address: osmo10arcf5r89cdmppntzkvulc7gfmw5lr66y2m25c937t6ccfzk0cqqz2l6xv'
  assert_line '  unbondingFrequency: "1"'
  assert_line '  RedemptionRate: "1.000000000000000000"'
}


##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################

@test "[INTEGRATION-BASIC-OSMO] ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom ustrd | GETBAL)
  osmo1_balance=$($OSMO_CMD q bank balances $OSMO_ADDRESS --denom $IBC_STRD_DENOM_OSMO | GETBAL)
  str1_balance_osmo=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_OSMO_DENOM | GETBAL)
  osmo1_balance_osmo=$($OSMO_CMD q bank balances $OSMO_ADDRESS --denom uosmo | GETBAL)
  # do IBC transfer
  $STRIDE_CMD tx ibc-transfer transfer transfer channel-1 $OSMO_ADDRESS 3000000000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test &
  $OSMO_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 3000000000uosmo --from oval1 --chain-id OSMO -y --keyring-backend test &
  WAIT_FOR_BLOCK $STRIDE_LOGS 8
  # get new balances
  str1_balance_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom ustrd | GETBAL)
  osmo1_balance_new=$($OSMO_CMD q bank balances $OSMO_ADDRESS --denom $IBC_STRD_DENOM_OSMO | GETBAL)
  str1_balance_osmo_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_OSMO_DENOM | GETBAL)
  osmo1_balance_osmo_new=$($OSMO_CMD q bank balances $OSMO_ADDRESS --denom uosmo | GETBAL)
  # get all STRD balance diffs
  str1_diff=$(($str1_balance - $str1_balance_new))
  osmo1_diff=$(($osmo1_balance - $osmo1_balance_new))
  assert_equal "$str1_diff" '3000000000'
  assert_equal "$osmo1_diff" '-3000000000'
  # get all OSMO_DENOM balance diffs
  str1_diff=$(($str1_balance_osmo - $str1_balance_osmo_new))
  osmo1_diff=$(($osmo1_balance_osmo - $osmo1_balance_osmo_new))
  assert_equal "$str1_diff" '-3000000000'
  assert_equal "$osmo1_diff" '3000000000'
}

@test "[INTEGRATION-BASIC-OSMO] liquid stake mints stOSMO" {
  # get module address
  MODADDR=$($STRIDE_CMD q stakeibc module-address stakeibc | awk '{print $NF}')
  # get initial balances
  mod_balance_osmo=$($STRIDE_CMD q bank balances $MODADDR --denom $IBC_OSMO_DENOM | GETBAL)
  str1_balance_osmo=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_OSMO_DENOM | GETBAL)
  str1_balance_stosmo=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $STOSMO_DENOM | GETBAL)
  # liquid stake
  $STRIDE_CMD tx stakeibc liquid-stake 1000000000 uosmo --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
  # sleep two block for the tx to settle on stride
  WAIT_FOR_STRING $STRIDE_LOGS '\[MINT ST ASSET\] success on OSMO'
  # make sure IBC_OSMO_DENOM went down
  str1_balance_osmo_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_OSMO_DENOM | GETBAL)
  str1_osmo_diff=$(($str1_balance_osmo - $str1_balance_osmo_new))
  assert_equal "$str1_osmo_diff" '1000000000'
  # make sure STOSMO went up
  str1_balance_stosmo_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $STOSMO_DENOM | GETBAL)
  str1_stosmo_diff=$(($str1_balance_stosmo_new-$str1_balance_stosmo))
  assert_equal "$str1_stosmo_diff" "1000000000"
}

@test "[INTEGRATION-BASIC-OSMO] tokens were transferred to OSMO after liquid staking" {
  # initial balance of delegation ICA
  initial_delegation_ica_bal=$($OSMO_CMD q bank balances $OSMO_DELEGATION_ICA_ADDR --denom uosmo | GETBAL)
  WAIT_FOR_STRING $STRIDE_LOGS '\[IBC-TRANSFER\] success to OSMO'
  # get the new delegation ICA balance
  post_delegation_ica_bal=$($OSMO_CMD q bank balances $OSMO_DELEGATION_ICA_ADDR --denom uosmo | GETBAL)
  diff=$(($post_delegation_ica_bal - $initial_delegation_ica_bal))
  assert_equal "$diff" '1000000000'
}

@test "[INTEGRATION-BASIC-OSMO] tokens on OSMO were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS '\[DELEGATION\] success on OSMO'
  # check staked tokens
  NEW_STAKE=$($OSMO_CMD q staking delegation $OSMO_DELEGATION_ICA_ADDR $OSMO_DELEGATE_VAL | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

# check that redemptions and claims work
@test "[INTEGRATION-BASIC-OSMO] redemption works" {
  sleep 5
  old_redemption_ica_bal=$($OSMO_CMD q bank balances $OSMO_REDEMPTION_ICA_ADDR --denom uosmo | GETBAL)
  # call redeem-stake
  amt_to_redeem=5
  $STRIDE_CMD tx stakeibc redeem-stake $amt_to_redeem OSMO $OSMO_RECEIVER_ACCT \
      --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[REDEMPTION] completed on OSMO'
  # check that the tokens were transferred to the redemption account
  new_redemption_ica_bal=$($OSMO_CMD q bank balances $OSMO_REDEMPTION_ICA_ADDR --denom uosmo | GETBAL)
  diff_positive=$(($new_redemption_ica_bal > $old_redemption_ica_bal))
  assert_equal "$diff_positive" "1"
}

@test "[INTEGRATION-BASIC-OSMO] claimed tokens are properly distributed" {
  # TODO(optimize tests) extra sleep just in case
  SENDER_ACCT=$STRIDE_VAL_ADDR
  old_sender_bal=$($OSMO_CMD q bank balances $OSMO_RECEIVER_ACCT --denom uosmo | GETBAL)
  # TODO check that the UserRedemptionRecord has claimIsPending = false
  # grab the epoch number for the first deposit record in the list od DRs
  EPOCH=$(strided q records list-user-redemption-record  | grep -Fiw 'epochNumber' | head -n 1 | grep -o -E '[0-9]+')
  # claim the record
  $STRIDE_CMD tx stakeibc claim-undelegated-tokens OSMO $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id STRIDE -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[CLAIM\] success on OSMO'
  # TODO check that UserRedemptionRecord has claimIsPending = true

  # check that the tokens were transferred to the sender account
  new_sender_bal=$($OSMO_CMD q bank balances $OSMO_RECEIVER_ACCT --denom uosmo | GETBAL)
  
  # check that the undelegated tokens were transfered to the sender account
  diff_positive=$(($new_sender_bal > $old_sender_bal))
  assert_equal "$diff_positive" "1"
}

# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC-OSMO] rewards are being reinvested, exchange rate updating" {
  # read the exchange rate
  RR1=$($STRIDE_CMD q stakeibc show-host-zone OSMO | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  # liquid stake again to kickstart the reinvestment process
  $STRIDE_CMD tx stakeibc liquid-stake 1000 uosmo --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # wait four days (transfers, stake, move rewards, reinvest rewards)
  epoch_duration=$($STRIDE_CMD q epochs epoch-infos | grep -Fiw -B 2 'stride_epoch' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($epoch_duration * 4))
  # simple check that number of tokens staked increases
  NEW_STAKED_BAL=$($OSMO_CMD q staking delegation $OSMO_DELEGATION_ICA_ADDR $OSMO_DELEGATE_VAL | GETSTAKE)
  EXPECTED_STAKED_BAL=667
  STAKED_BAL_INCREASED=$(($NEW_STAKED_BAL > $EXPECTED_STAKED_BAL))
  assert_equal "$STAKED_BAL_INCREASED" "1"

  RR2=$($STRIDE_CMD q stakeibc show-host-zone OSMO | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  # check that the exchange rate has increased
  MULT=1000000
  RR_INCREASED=$(( $(FLOOR $(DECMUL $RR2 $MULT)) > $(FLOOR $(DECMUL $RR1 $MULT))))
  assert_equal "$RR_INCREASED" "1"
}

# TODO check that the correct amount is being reinvested and the correct amount is flowing to the rev EOA
@test "[NOT-IMPLEMENTED] reinvestment and revenue amounts are correct" {
}
