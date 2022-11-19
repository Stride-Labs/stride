#!/bin/bash
BASE_SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# run test files
CHAIN_NAME=GAIA TRANSFER_CHANNEL_NUMBER=0 bats $BASE_SCRIPT_DIR/gaia_tests.bats 
# bats $BASE_SCRIPT_DIR/juno_tests.bats
# bats $BASE_SCRIPT_DIR/osmo_tests.bats
# bats $BASE_SCRIPT_DIR/stars_tests.bats

