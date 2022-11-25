#!/bin/bash
BASE_SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# run test files
CHAIN_NAME=GAIA TRANSFER_CHANNEL_NUMBER=0 bats $BASE_SCRIPT_DIR/integration_tests.bats 
CHAIN_NAME=JUNO TRANSFER_CHANNEL_NUMBER=1 bats $BASE_SCRIPT_DIR/integration_tests.bats 
CHAIN_NAME=OSMO TRANSFER_CHANNEL_NUMBER=2 bats $BASE_SCRIPT_DIR/integration_tests.bats 
CHAIN_NAME=STARS TRANSFER_CHANNEL_NUMBER=3 bats $BASE_SCRIPT_DIR/integration_tests.bats 

