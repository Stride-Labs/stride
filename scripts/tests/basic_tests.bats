#!/usr/bin/env bats

setup_file() {
  export STRIDE_ADDRESS_1="test"
  # get the containing directory of this file
  SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
  PATH="$SCRIPT_DIR/../../:$PATH"

  # set allows us to export all variables in account_vars
  set -a
  export STRIDE_ADDRESS_1="TEMP"
  source scripts/account_vars.sh
  export STRIDE_ADDRESS_1="HI"
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
  export SETUP_FILE_EXPORT_TEST=true
}

@test "variables exported in setup_file are visible in tests" {
  [[ $SETUP_FILE_EXPORT_TEST == "true" ]]
}

@test "proper initial address names" {
  [[ $STRIDE_ADDRESS_1 == "stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7" ]]
  assert_equal $STRIDE_ADDRESS_1 'stride1uk4ze0x4nvh4fk0xm4jdud58eqn4yxhrt52vv7'
  assert_equal $STRIDE_ADDRESS_2 'stride1ld5ewfgc3crml46n806km7djtr788vqdd5lnu5'
  assert_equal $STRIDE_ADDRESS_3 'stride16vlrvd7lsfqg8q7kyxcyar9v7nt0h99p5arglq'

  assert_equal $GAIA_ADDRESS_1 'cosmos1pcag0cj4ttxg8l7pcg0q4ksuglswuuedcextl2'
  assert_equal $GAIA_ADDRESS_2 'cosmos1t2aqq3c6mt8fa6l5ady44manvhqf77sywjcldv'
  assert_equal $GAIA_ADDRESS_3 'cosmos19e7sugzt8zaamk2wyydzgmg9n3ysylg6kfwrk2'

  assert_equal $RLY_ADDRESS_1 'stride1ft20pydau82pgesyl9huhhux307s9h3078692y'
  assert_equal $RLY_ADDRESS_2 'cosmos1uyrmx8zw0mxu7sdn58z29wnnqnxtqvvxqec074'
}
