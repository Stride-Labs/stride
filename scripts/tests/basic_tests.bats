#!/usr/bin/env bats

setup_file() {
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

  # set allows us to export all variables in account_vars
  #   alias GETBAL="head -n 1 | grep -o -E '[0-9]+'"
  set -a
  source scripts/account_vars.sh
  IBCSTRD='ibc/FF6C2E86490C1C4FBBD24F55032831D2415B9D7882F85C3CC9C2401D79362BEA'
  IBCATOM='ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9'
  STATOM="st${IBCATOM}"
  GETBAL() {
    head -n 1 | grep -o -E '[0-9]+'
  }
  set +a
}

setup() {
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

  load 'test_helper/bats-support/load'
  load 'test_helper/bats-assert/load'
}

@test "proper initial address names" {
  assert_equal $STRIDE_ADDRESS_1 'stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7'
  assert_equal $STRIDE_ADDRESS_2 'stride1ld5ewfgc3crml46n806km7djtr788vqdd5lnu5'
  assert_equal $STRIDE_ADDRESS_3 'stride16vlrvd7lsfqg8q7kyxcyar9v7nt0h99p5arglq'

  assert_equal $GAIA_ADDRESS_1 'cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2'
  assert_equal $GAIA_ADDRESS_2 'cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv'
  assert_equal $GAIA_ADDRESS_3 'cosmos19e7sugzt8zaamk2wyydzgmg9n3ysylg6kfwrk2'

  assert_equal $RLY_ADDRESS_1 'stride1ft20pydau82pgesyl9huhhux307s9h3078692y'
  assert_equal $RLY_ADDRESS_2 'cosmos1uyrmx8zw0mxu7sdn58z29wnnqnxtqvvxqec074'
}

@test "ibc transfer updates all balances" {
  # get initial balances
  str1_balance=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom ustrd | GETBAL)
  gaia1_balance=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom $IBCSTRD | GETBAL)
  str1_balance_atom=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  gaia1_balance_atom=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom uatom | GETBAL)
  # do IBC transfer
  $STR1_EXEC tx ibc-transfer transfer transfer channel-1 $GAIA_ADDRESS_1 1000ustrd --from val1 --chain-id STRIDE -y --keyring-backend test
  $GAIA1_EXEC tx ibc-transfer transfer transfer channel-0 $STRIDE_ADDRESS_1 1000uatom --from gval1 --chain-id GAIA -y --keyring-backend test
  sleep 20
  # get new balances
  str1_balance_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom ustrd | GETBAL)
  gaia1_balance_new=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom $IBCSTRD | GETBAL)
  str1_balance_atom_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  gaia1_balance_atom_new=$($GAIA1_EXEC q bank balances $GAIA_ADDRESS_1 --denom uatom | GETBAL)
  # get all STRD balance diffs
  str1_diff=$(($str1_balance - $str1_balance_new))
  gaia1_diff=$(($gaia1_balance - $gaia1_balance_new))
  assert_equal "$str1_diff" '1000'
  assert_equal "$gaia1_diff" '-1000'
  # get all ATOM balance diffs
  str1_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  gaia1_diff=$(($gaia1_balance_atom - $gaia1_balance_atom_new))
  assert_equal "$str1_diff" '-1000'
  assert_equal "$gaia1_diff" '1000'
}

@test "liquid stake mints stATOM" {
  str1_balance_atom=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  str1_balance_statom=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $STATOM | GETBAL)
  # liquid stake
  $STR1_EXEC tx stakeibc liquid-stake 1000 $IBCATOM --keyring-backend test --from val1 -y
  sleep 5
  # make sure IBCATOM went down 
  str1_balance_atom_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $IBCATOM | GETBAL)
  str1_atom_diff=$(($str1_balance_atom - $str1_balance_atom_new))
  assert_equal "$str1_atom_diff" '1000'
  # make sure STATOM went up
  str1_balance_statom_new=$($STR1_EXEC q bank balances $STRIDE_ADDRESS_1 --denom $STATOM | GETBAL)
  str1_statom_diff=$(($str1_balance_statom_new-$str1_balance_statom))
  assert_equal "$str1_statom_diff" '1000'
}

@test "liquid stake IBCs automatically" {
  # get module address 
  ibcaddr=$($STR1_EXEC q stakeibc module-address stakeibc | awk '{print $NF}') 
  module_atom=$($STR1_EXEC q bank balances $ibcaddr --denom $IBCATOM | GETBAL)
  assert_equal "$module_atom" '1000'
}

# add test to register host zone 
@test "host zone successfully registered" {
  run $STR1_EXEC q stakeibc show-host-zone GAIA
  host_zone_info=$($STR1_EXEC q stakeibc get-host-zone  | awk '{print $NF}')
  assert_line '  HostDenom: ibc/C4CFF46FD6DE35CA4CF4CE031E643C8FDC9BA4B99AE598E9B0ED98FE3A2319F9'
}

# TEST-74
# add test to see if assets are properly being staked on host zone
# add asset redemption test
