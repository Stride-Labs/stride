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

# # add test to register host zone
@test "[INTEGRATION-BASIC] host zones successfully registered" {
  run $STRIDE_MAIN_CMD q stakeibc show-host-zone GAIA
  assert_line '  HostDenom: uatom'
  assert_line '  chainId: GAIA'
  refute_line '  delegationAccount: null'
  refute_line '  feeAccount: null'
  refute_line '  redemptionAccount: null'
  refute_line '  withdrawalAccount: null'
  assert_line '  unbondingFrequency: "1"'
}


##############################################################################################
######                TEST BASIC STRIDE FUNCTIONALITY                                   ######
##############################################################################################


@test "[INTEGRATION-BASIC-GAIA] ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom ustrd | GETBAL)
  gaia1_balance=$($GAIA_MAIN_CMD q bank balances $(GAIA_ADDRESS) --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_atom=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_ATOM_DENOM | GETBAL)
  gaia1_balance_atom=$($GAIA_MAIN_CMD q bank balances $(GAIA_ADDRESS) --denom uatom | GETBAL)
  # do IBC transfer
  $STRIDE_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(GAIA_ADDRESS) 3000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test &
  $GAIA_MAIN_CMD tx ibc-transfer transfer transfer channel-0 $(STRIDE_ADDRESS) 3000uatom --from gval1 --chain-id GAIA -y --keyring-backend test &
  WAIT_FOR_BLOCK $STRIDE_LOGS 8
  # get new balances
  str1_balance_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom ustrd | GETBAL)
  gaia1_balance_new=$($GAIA_MAIN_CMD q bank balances $(GAIA_ADDRESS) --denom $IBC_STRD_DENOM | GETBAL)
  str1_balance_atom_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_ATOM_DENOM | GETBAL)
  gaia1_balance_atom_new=$($GAIA_MAIN_CMD q bank balances $(GAIA_ADDRESS) --denom uatom | GETBAL)
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

@test "[INTEGRATION-BASIC-GAIA] liquid stake mints stATOM" {
  # get module address
  MODADDR=$($STRIDE_MAIN_CMD q stakeibc module-address stakeibc | awk '{print $NF}')
  # get initial balances
  mod_balance_atom=$($STRIDE_MAIN_CMD q bank balances $MODADDR --denom $IBC_ATOM_DENOM | GETBAL)
  str1_balance_atom=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_ATOM_DENOM | GETBAL)
  str1_balance_statom=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STATOM_DENOM | GETBAL)
  # liquid stake
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID
  # sleep two block for the tx to settle on stride
  WAIT_FOR_STRING $STRIDE_LOGS '\[MINT ST ASSET\] success on GAIA'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # make sure IBC_ATOM_DENOM went down
  str1_balance_atom_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $IBC_ATOM_DENOM | GETBAL)
  str1_atom_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  assert_equal "$str1_atom_diff" '1000'
  # make sure STATOM went up
  str1_balance_statom_new=$($STRIDE_MAIN_CMD q bank balances $(STRIDE_ADDRESS) --denom $STATOM_DENOM | GETBAL)
  str1_statom_diff=$(($str1_balance_statom_new-$str1_balance_statom))
  assert_equal "$str1_statom_diff" "1000"
}

# check that tokens were transferred to GAIA
@test "[INTEGRATION-BASIC-GAIA] tokens were transferred to GAIA after liquid staking" {
  # initial balance of delegation ICA
  initial_delegation_ica_bal=$($GAIA_MAIN_CMD q bank balances $(GET_ICA_ADDR GAIA delegation) --denom uatom | GETBAL)
  WAIT_FOR_STRING $STRIDE_LOGS '\[IBC-TRANSFER\] success to GAIA'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # get the new delegation ICA balance
  post_delegation_ica_bal=$($GAIA_MAIN_CMD q bank balances $(GET_ICA_ADDR GAIA delegation) --denom uatom | GETBAL)
  diff=$(($post_delegation_ica_bal - $initial_delegation_ica_bal))
  assert_equal "$diff" '1000'
}

# check that tokens on GAIA are staked
@test "[INTEGRATION-BASIC-GAIA] tokens on GAIA were staked" {
  # wait for another epoch to pass so that tokens are staked
  WAIT_FOR_STRING $STRIDE_LOGS '\[DELEGATION\] success on GAIA'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # check staked tokens
  NEW_STAKE=$($GAIA_MAIN_CMD q staking delegation $(GET_ICA_ADDR GAIA delegation) $GAIA_DELEGATE_VAL | GETSTAKE)
  stake_diff=$(($NEW_STAKE > 0))
  assert_equal "$stake_diff" "1"
}

# check that redemptions and claims work
@test "[INTEGRATION-BASIC-GAIA] redemption works" {
  old_redemption_ica_bal=$($GAIA_MAIN_CMD q bank balances $(GET_ICA_ADDR GAIA redemption) --denom uatom | GETBAL)
  # call redeem-stake
  amt_to_redeem=100
  $STRIDE_MAIN_CMD tx stakeibc redeem-stake $amt_to_redeem GAIA $GAIA_RECEIVER_ACCT \
      --from val1 --keyring-backend test --chain-id $STRIDE_CHAIN_ID -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[REDEMPTION] completed on GAIA'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # check that the tokens were transferred to the redemption account
  new_redemption_ica_bal=$($GAIA_MAIN_CMD q bank balances $(GET_ICA_ADDR GAIA redemption) --denom uatom | GETBAL)
  diff_positive=$(($new_redemption_ica_bal > $old_redemption_ica_bal))
  assert_equal "$diff_positive" "1"
}

@test "[INTEGRATION-BASIC-GAIA] claimed tokens are properly distributed" {
  # TODO(optimize tests) extra sleep just in case
  SENDER_ACCT=$(STRIDE_ADDRESS)
  old_sender_bal=$($GAIA_MAIN_CMD q bank balances $GAIA_RECEIVER_ACCT --denom uatom | GETBAL)
  # TODO check that the UserRedemptionRecord has isClaimable = true

  # grab the epoch number for the first deposit record in the list od DRs
  EPOCH=$(strided q records list-user-redemption-record  | grep -Fiw 'epochNumber' | head -n 1 | grep -o -E '[0-9]+')
  # claim the record
  $STRIDE_MAIN_CMD tx stakeibc claim-undelegated-tokens GAIA $EPOCH $SENDER_ACCT --from val1 --keyring-backend test --chain-id STRIDE -y
  WAIT_FOR_STRING $STRIDE_LOGS '\[CLAIM\] success on GAIA'
  WAIT_FOR_BLOCK $STRIDE_LOGS 2

  # check that the tokens were transferred to the sender account
  new_sender_bal=$($GAIA_MAIN_CMD q bank balances $GAIA_RECEIVER_ACCT --denom uatom | GETBAL)

  # check that the undelegated tokens were transfered to the sender account
  diff_positive=$(($new_sender_bal > $old_sender_bal))
  assert_equal "$diff_positive" "1"
}


# check that a second liquid staking call kicks off reinvestment
@test "[INTEGRATION-BASIC-GAIA] rewards are being reinvested, exchange rate updating" {
  # read the exchange rate and current delegations
  RR1=$($STRIDE_MAIN_CMD q stakeibc show-host-zone GAIA | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  OLD_STAKED_BAL=$($GAIA_MAIN_CMD q staking delegation $(GET_ICA_ADDR GAIA delegation) $(GET_VAL_ADDR GAIA 1) | GETSTAKE)
  # liquid stake again to kickstart the reinvestment process
  $STRIDE_MAIN_CMD tx stakeibc liquid-stake 1000 uatom --keyring-backend test --from val1 -y --chain-id $STRIDE_CHAIN_ID
  WAIT_FOR_BLOCK $STRIDE_LOGS 2
  # wait four days (transfers, stake, move rewards, reinvest rewards)
  epoch_duration=$($STRIDE_MAIN_CMD q epochs epoch-infos | grep -Fiw -B 2 'stride_epoch' | head -n 1 | grep -o -E '[0-9]+')
  sleep $(($epoch_duration * 4))
  # simple check that number of tokens staked increases
  NEW_STAKED_BAL=$($GAIA_MAIN_CMD q staking delegation $(GET_ICA_ADDR GAIA delegation) $(GET_VAL_ADDR GAIA 1) | GETSTAKE)
  STAKED_BAL_INCREASED=$(($NEW_STAKED_BAL > $OLD_STAKED_BAL))
  assert_equal "$STAKED_BAL_INCREASED" "1"

  RR2=$($STRIDE_MAIN_CMD q stakeibc show-host-zone GAIA | grep -Fiw 'RedemptionRate' | grep -Eo '[+-]?[0-9]+([.][0-9]+)?')
  # check that the exchange rate has increased
  MULT=1000000
  RR_INCREASED=$(( $(FLOOR $(DECMUL $RR2 $MULT)) > $(FLOOR $(DECMUL $RR1 $MULT))))
  assert_equal "$RR_INCREASED" "1"
}

# TODO check that the correct amount is being reinvested and the correct amount is flowing to the rev EOA
@test "[NOT-IMPLEMENTED] reinvestment and revenue amounts are correct" {
}
