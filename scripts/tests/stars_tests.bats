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

  assert_equal $(GET_VAL_ADDR STARS 1) 'starsvaloper12ffkl30v0ghtyaezvedazquhtsf4q5ng2c0xaf'
  assert_equal $(GET_ICA_ADDR STARS delegation) "stars1kl6wa99e6hf97xr90m2n04rl0smv842pj9utqyvgyrksrm9aacdqyfc3en"
  assert_equal $(GET_ICA_ADDR STARS redemption) "stars1x07hv0hxujj6l0mfyynwyuccf8fl27vjup0y8dmyajy9ugae22hqfvmv4e"
  assert_equal $(GET_ICA_ADDR STARS withdrawal) "stars1x5ndl5p9tjy376a9xmqhw79gz0s678480759cdgaretcgm36akvs0a78tj"
  assert_equal $(GET_ICA_ADDR STARS fee) "stars1v09y993sku5djvm0rffq0nfsk5rzke4d2vzvny5e6vmq7dz0dehqnwl4ay"
}

# # add test to register host zone
@test "[INTEGRATION-BASIC] host zones successfully registered" {
  run $STRIDE_MAIN_CMD q stakeibc show-host-zone STARS
  assert_line '  HostDenom: ustars'
  assert_line '  chainId: STARS'
  assert_line '  delegationAccount:'
  assert_line '    address: stars1kl6wa99e6hf97xr90m2n04rl0smv842pj9utqyvgyrksrm9aacdqyfc3en'
  assert_line '  feeAccount:'
  assert_line '    address: stars1v09y993sku5djvm0rffq0nfsk5rzke4d2vzvny5e6vmq7dz0dehqnwl4ay'
  assert_line '  redemptionAccount:'
  assert_line '    address: stars1x07hv0hxujj6l0mfyynwyuccf8fl27vjup0y8dmyajy9ugae22hqfvmv4e'
  assert_line '  withdrawalAccount:'
  assert_line '    address: stars1x5ndl5p9tjy376a9xmqhw79gz0s678480759cdgaretcgm36akvs0a78tj'
  assert_line '  unbondingFrequency: "1"'
}

##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################


@test "[INTEGRATION-BASIC-STARS] ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom ustrd | GETBAL)
  stars1_balance=$($STARS_MAIN_CMD q bank balances $(STARS_ADDRESS) --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_stars=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_STARS_DENOM | GETBAL)
  stars1_balance_stars=$($STARS_MAIN_CMD q bank balances $(STARS_ADDRESS) --denom ustars | GETBAL)
  # do IBC transfer
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-3 $(STARS_ADDRESS) 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test #&
  $STARS_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 3000ustars --from sgval1 --chain-id STARS -y --keyring-backend test #&
  WAIT_FOR_BLOCK $STRIDE_LOGS 8
  # get new balances
  str1_balance_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom ustrd | GETBAL)
  stars1_balance_new=$($STARS_MAIN_CMD q bank balances $(STARS_ADDRESS) --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_stars_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_STARS_DENOM | GETBAL)
  stars1_balance_stars_new=$($STARS_MAIN_CMD q bank balances $(STARS_ADDRESS) --denom ustars | GETBAL)
  # get all STRD balance diffs
  str1_diff=$(($str1_balance - $str1_balance_new))
  stars1_diff=$(($stars1_balance - $stars1_balance_new))
  assert_equal "$str1_diff" '3000'
  assert_equal "$stars1_diff" '-3000'
  # get all STARS_DENOM balance diffs
  str1_diff=$(($str1_balance_stars - $str1_balance_stars_new))
  stars1_diff=$(($stars1_balance_stars - $stars1_balance_stars_new))
  assert_equal "$str1_diff" '-3000'
  assert_equal "$stars1_diff" '3000'
}

@test "[INTEGRATION-BASIC-STARS] liquid stake mints stSTARS" {
  # get module address
  MODADDR=$($STRIDE_MAIN_CMD q stakeibc module-address stakeibc | awk '{print $NF}')
  # get initial balances
  mod_balance_stars=$($STRIDE_MAIN_CMD q bank balances $MODADDR --denom $IBC_STARS_DENOM | GETBAL)
  str1_balance_stars=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_STARS_DENOM | GETBAL)
  str1_balance_ststars=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STSTARS_DENOM | GETBAL)
  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 ustars --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID
  # sleep two block for the tx to settle on stride
  WAIT_FOR_STRING $STRIDE_LOGS '\[MINT ST ASSET\] success on STARS'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # make sure IBC_STARS_DENOM went down
  str1_balance_stars_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_STARS_DENOM | GETBAL)
  str1_stars_diff=$(($str1_balance_stars - $str1_balance_stars_new))
  assert_equal "$str1_stars_diff" '1000'
  # make sure STSTARS went up
  str1_balance_ststars_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STSTARS_DENOM | GETBAL)
  str1_ststars_diff=$(($str1_balance_ststars_new-$str1_balance_ststars))
  assert_equal "$str1_ststars_diff" "1000"
}

# check that tokens were transferred to STARS
@test "[INTEGRATION-BASIC-STARS] tokens were transferred to STARS after liquid staking" {
  # initial balance of delegation ICA
  initial_delegation_ica_bal=$($STARS_MAIN_CMD q bank balances $(GET_ICA_ADDR STARS delegation) --denom ustars | GETBAL)
  WAIT_FOR_STRING $STRIDE_LOGS '\[IBC-TRANSFER\] success to STARS'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # get the new delegation ICA balance
  post_delegation_ica_bal=$($STARS_MAIN_CMD q bank balances $(GET_ICA_ADDR STARS delegation) --denom ustars | GETBAL)
  diff=$(($post_delegation_ica_bal - $initial_delegation_ica_bal))
  assert_equal "$diff" '1000'
}

# check that tokens on STARS are staked
@test "[INTEGRATION-BASIC-STARS] tokens on STARS were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS '\[DELEGATION\] success on STARS'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # check staked tokens
  NEW_STAKE=$($STARS_MAIN_CMD q staking delegation $(GET_ICA_ADDR STARS delegation) $(GET_VAL_ADDR STARS 1) | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

# # check that redemptions and claims work
@test "[INTEGRATION-BASIC-STARS] redemption works" {
  old_redemption_ica_bal=$($STARS_MAIN_CMD q bank balances $(GET_ICA_ADDR STARS redemption) --denom ustars | GETBAL)
  # call redeem-stake
  amt_to_redeem=100
  $STRIDE_MAIN_CMD tx stakeibc redeem-stake $amt_to_redeem STARS $STARS_RECEIVER_ACCT \
      --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[REDEMPTION] completed on STARS'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # check that the tokens were transferred to the redemption account
  new_redemption_ica_bal=$($STARS_MAIN_CMD q bank balances $(GET_ICA_ADDR STARS redemption) --denom ustars | GETBAL)
  diff_positive=$(($new_redemption_ica_bal > $old_redemption_ica_bal))
  assert_equal "$diff_positive" "1"
}

@test "[INTEGRATION-BASIC-STARS] claimed tokens are properly distributed" {
  # TODO(optimize tests) extra sleep just in case
  SENDER_ACCT=$(STRIDE_ADDRESS)
  old_sender_bal=$($STARS_MAIN_CMD q bank balances $STARS_RECEIVER_ACCT --denom ustars | GETBAL)
  # TODO check that the UserRedemptionRecord has isClaimable = true

  # grab the epoch number for the first deposit record in the list od DRs
  EPOCH=$(strided q records list-user-redemption-record  | grep -Fiw 'epochNumber' | head -n 1 | grep -o -E '[0-9]+')
  # claim the record
  $STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens STARS $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id STRIDE -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[CLAIM\] success on STARS'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check that the tokens were transferred to the sender account
  new_sender_bal=$($STARS_MAIN_CMD q bank balances $STARS_RECEIVER_ACCT --denom ustars | GETBAL)

  # check that the undelegated tokens were transfered to the sender account
  diff_positive=$(($new_sender_bal > $old_sender_bal))
  assert_equal "$diff_positive" "1"
}


# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC-STARS] rewards are being reinvested, exchange rate updating" {
  # read the exchange rate and current delegations
  RR1=$($STRIDE_MAIN_CMD q stakeibc show-host-zone STARS | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  OLD_STAKED_BAL=$($STARS_MAIN_CMD q staking delegation $(GET_ICA_ADDR STARS delegation) $(GET_VAL_ADDR STARS 1) | GETSTAKE)
  # liquid stake again to kickstart the reinvestment process
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 ustars --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # wait four days (transfers, stake, move rewards, reinvest rewards)
  epoch_duration=$($STRIDE_MAIN_CMD q epochs epoch-infos | grep -Fiw -B 2 'stride_epoch' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($epoch_duration * 4))
  # simple check that number of tokens staked increases
  NEW_STAKED_BAL=$($STARS_MAIN_CMD q staking delegation $(GET_ICA_ADDR STARS delegation) $(GET_VAL_ADDR STARS 1) | GETSTAKE)
  STAKED_BAL_INCREASED=$(($NEW_STAKED_BAL > $OLD_STAKED_BAL))
  assert_equal "$STAKED_BAL_INCREASED" "1"

  RR2=$($STRIDE_MAIN_CMD q stakeibc show-host-zone STARS | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  # check that the exchange rate has increased
  MULT=1000000
  RR_INCREASED=$(( $(FLOOR $(DECMUL $RR2 $MULT)) > $(FLOOR $(DECMUL $RR1 $MULT))))
  assert_equal "$RR_INCREASED" "1"
}

# # TODO check that the correct amount is being reinvested and the correct amount is flowing to the rev EOA
# @test "[NOT-IMPLEMENTED] reinvestment and revenue amounts are correct" {
# }
