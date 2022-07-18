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

}

@test "[INTEGRATION-BASIC] ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom ustrd | GETBAL)
  gaia1_balance=$($GAIA_CMD q bank balances $GAIA_ADDRESS --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_atom=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_ATOM_DENOM | GETBAL)
  gaia1_balance_atom=$($GAIA_CMD q bank balances $GAIA_ADDRESS --denom uatom | GETBAL)
  # do IBC transfer
  $STRIDE_CMD tx ibc-transfer transfer transfer channel-0 $GAIA_ADDRESS 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
  $GAIA_CMD tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS 3000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
  sleep $IBC_TX_WAIT_SECONDS
  # get new balances
  str1_balance_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom ustrd | GETBAL)
  gaia1_balance_new=$($GAIA_CMD q bank balances $GAIA_ADDRESS --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_atom_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_ATOM_DENOM | GETBAL)
  gaia1_balance_atom_new=$($GAIA_CMD q bank balances $GAIA_ADDRESS --denom uatom | GETBAL)
  # get all STRD balance diffs
  str1_diff=$(($str1_balance - $str1_balance_new))
  gaia1_diff=$(($gaia1_balance - $gaia1_balance_new))
  assert_equal "$str1_diff" '3000'
  assert_equal "$gaia1_diff" '-3000'
  # get all ATOM_DENOM balance diffs
  str1_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  gaia1_diff=$(($gaia1_balance_atom - $gaia1_balance_atom_new))
  assert_equal "$str1_diff" '-3000'
  assert_equal "$gaia1_diff" '3000'
}

# # add test to register host zone 
@test "[INTEGRATION-BASIC] host zone successfully registered" {
  run $STRIDE_CMD q stakeibc show-host-zone GAIA
  assert_line '  HostDenom: uatom'
  assert_line '  chainId: GAIA'
  assert_line '  delegationAccount:'
  assert_line '    address: cosmos1sy63lffevueudvvlvh2lf6s387xh9xq72n3fsy6n2gr5hm6u2szs2v0ujm'
  assert_line '  feeAccount:'
  assert_line '    address: cosmos1lkgt5sfshn9shm7hd7chtytkq4mvwvswgmyl0hkacd4rmusu9wwq60cezx'
  assert_line '  redemptionAccount:'
  assert_line '    address: cosmos1xmcwu75s8v7s54k79390wc5gwtgkeqhvzegpj0nm2tdwacv47tmqg9ut30'
  assert_line '  withdrawalAccount:'
  assert_line '    address: cosmos1x5p8er7e2ne8l54tx33l560l8djuyapny55pksctuguzdc00dj7saqcw2l'
  assert_line '  unbondingFrequency: "1"'
  assert_line '  RedemptionRate: "1.000000000000000000"'
}


##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################

@test "[INTEGRATION-BASIC] liquid stake mints stATOM" {
  # get module address
  MODADDR=$($STRIDE_CMD q stakeibc module-address stakeibc | awk '{print $NF}') 
  # get initial balances
  mod_balance_atom=$($STRIDE_CMD q bank balances $MODADDR --denom $IBC_ATOM_DENOM | GETBAL)
  str1_balance_atom=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_ATOM_DENOM | GETBAL)
  str1_balance_statom=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $STATOM_DENOM | GETBAL)
  # liquid stake
  $STRIDE_CMD tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
  # sleep two block for the tx to settle on stride
  BLOCK_SLEEP 2
  # make sure IBC_ATOM_DENOM went down 
  str1_balance_atom_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $IBC_ATOM_DENOM | GETBAL)
  str1_atom_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  assert_equal "$str1_atom_diff" '1000'
  # make sure STATOM went up
  str1_balance_statom_new=$($STRIDE_CMD q bank balances $STRIDE_ADDRESS --denom $STATOM_DENOM | GETBAL)
  str1_statom_diff=$(($str1_balance_statom_new-$str1_balance_statom))
  assert_equal "$str1_statom_diff" "1000"
}

# check that tokens were transferred to GAIA
@test "[INTEGRATION-BASIC] tokens were transferred to GAIA after liquid staking" {
  # initial balance of delegation ICA
  initial_delegation_ica_bal=$($GAIA_CMD q bank balances $DELEGATION_ICA_ADDR --denom uatom | GETBAL)
  # wait for the epoch to pass (we liquid staked above)
  remaining_seconds=$($STRIDE_CMD q epochs seconds-remaining stride_epoch)
  sleep $remaining_seconds
  # sleep 30 seconds for the IBC calls to settle
  sleep $IBC_TX_WAIT_SECONDS
  # get the new delegation ICA balance
  post_delegation_ica_bal=$($GAIA_CMD q bank balances $DELEGATION_ICA_ADDR --denom uatom | GETBAL)
  diff=$(($post_delegation_ica_bal - $initial_delegation_ica_bal))
  assert_equal "$diff" '1000'
}

# check that tokens on GAIA are staked
@test "[INTEGRATION-BASIC] tokens on GAIA were staked" {
  # wait for another epoch to pass so that tokens are staked
  remaining_seconds=$($STRIDE_CMD q epochs seconds-remaining stride_epoch)
  sleep $remaining_seconds
  # sleep 30 seconds for the IBC calls to settle
  sleep $IBC_TX_WAIT_SECONDS
  # check staked tokens
  NEW_STAKE=$($GAIA_CMD q staking delegation $DELEGATION_ICA_ADDR $GAIA_DELEGATE_VAL | GETSTAKE)
  # note that old stake is 0, so we can safely check the new stake value rather than the diff
  assert_equal "$NEW_STAKE" "333"
}


# check that redemptions and claims work
@test "[INTEGRATION-BASIC] redemption works" {
  old_redemption_ica_bal=$($GAIA_CMD q bank balances $REDEMPTION_ICA_ADDR --denom uatom | GETBAL)
  # call redeem-stake
  amt_to_redeem=5
  $STRIDE_CMD tx stakeibc redeem-stake $amt_to_redeem GAIA $GAIA_RECEIVER_ACCT \
      --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN -y
  # wait for beginning of next day, then for ibc transaction time for the unbonding period to begin
  remaining_seconds=$($STRIDE_CMD q epochs seconds-remaining day)
  sleep $remaining_seconds
  sleep $IBC_TX_WAIT_SECONDS
  # TODO check for an unbonding record
  # TODO check that a UserRedemptionRecord was created with isClaimabled = false
  # wait for the unbonding period to pass
  UNBONDING_PERIOD=$($GAIA_CMD q staking params |  grep -o -E '[0-9]+' | tail -n 1)
  sleep $UNBONDING_PERIOD
  BLOCK_SLEEP 5 # for unbonded amount to land in delegation acct on host chain
  # wait for a day to pass (to transfer from delegation to redemption acct)
  remaining_seconds=$($STRIDE_CMD q epochs seconds-remaining day)
  sleep $remaining_seconds
  # TODO we're sleeping more than we should have to here, investigate why redemptions take so long!
  BLOCK_SLEEP 2
  remaining_seconds=$($STRIDE_CMD q epochs seconds-remaining day)
  sleep $remaining_seconds
  BLOCK_SLEEP 2
  remaining_seconds=$($STRIDE_CMD q epochs seconds-remaining day)
  sleep $remaining_seconds
  # wait for ica bank send to process on host chain (delegation => redemption acct)
  sleep $IBC_TX_WAIT_SECONDS
  # check that the tokens were transferred to the redemption account
  new_redemption_ica_bal=$($GAIA_CMD q bank balances $REDEMPTION_ICA_ADDR --denom uatom | GETBAL)
  diff=$(($new_redemption_ica_bal - $old_redemption_ica_bal))
  assert_equal "$diff" "$amt_to_redeem"
}

@test "[INTEGRATION-BASIC] claimed tokens are properly distributed" {

  # TODO(optimize tests) extra sleep just in case
  sleep 30
  SENDER_ACCT=$STRIDE_VAL_ADDR
  old_sender_bal=$($GAIA_CMD q bank balances $GAIA_RECEIVER_ACCT --denom uatom | GETBAL)
  # TODO check that the UserRedemptionRecord has isClaimable = true

  # grab the epoch number for the first deposit record in the list od DRs  
  EPOCH=$(strided q records list-user-redemption-record  | grep -Fiw 'epochNumber' | head -n 1 | grep -o -E '[0-9]+')
  # claim the record
  $STRIDE_CMD tx stakeibc claim-undelegated-tokens GAIA $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id STRIDE -y
  sleep $IBC_TX_WAIT_SECONDS
  # TODO check that UserRedemptionRecord has isClaimable = false
  
  # check that the tokens were transferred to the sender account
  new_sender_bal=$($GAIA_CMD q bank balances $GAIA_RECEIVER_ACCT --denom uatom | GETBAL)

  # check that the undelegated tokens were transfered to the sender account
  diff=$(($new_sender_bal - $old_sender_bal))
  assert_equal "$diff" "5"
}


# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC] rewards are being reinvested (delegated balance increasing)" {
  # liquid stake again to kickstart the reinvestment process
  $STRIDE_CMD tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN
  BLOCK_SLEEP 2  
  # wait four days (transfers, stake, move rewards, reinvest rewards)
  day_duration=$($STRIDE_CMD q epochs epoch-infos | grep -Fiw 'duration' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($day_duration * 4))

  # simple check that number of tokens staked increases
  NEW_STAKED_BAL=$($GAIA_CMD q staking delegation $DELEGATION_ICA_ADDR $GAIA_DELEGATE_VAL | GETSTAKE)
  EXPECTED_STAKED_BAL=680
  STAKED_BAL_INCREASED=$(($NEW_STAKED_BAL > $EXPECTED_STAKED_BAL))
  assert_equal "$STAKED_BAL_INCREASED" "1" 
}

# check that exchange rate is updating
@test "[INTEGRATION-BASIC] exchange rate is updating" {
  # read the exchange rate
  RR1=$($STRIDE_CMD q stakeibc list-host-zone | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')

  # wait for reinvestment to happen (4 days is enough)
  day_duration=$($STRIDE_CMD q epochs epoch-infos | grep -Fiw 'duration' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($day_duration * 4))

  RR2=$($STRIDE_CMD q stakeibc list-host-zone | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')

  # check that the exchange rate has increased
  MULT=1000000
  RR_INCREASED=$(( $(FLOOR $(DECMUL $RR2 $MULT)) > $(FLOOR $(DECMUL $RR1 $MULT))))
  assert_equal "$RR_INCREASED" "1"
}

# TODO check that the correct amount is being reinvested and the correct amount is flowing to the rev EOA
@test "[NOT-IMPLEMENTED] reinvestment and revenue amounts are correct" {
}
