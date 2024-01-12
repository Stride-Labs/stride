#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# run test files
BATS=${SCRIPT_DIR}/bats/bats-core/bin/bats
INTEGRATION_TEST_FILE=${SCRIPT_DIR}/integration_tests.bats 

if [[ "$PART" == "1" ]]; then 
    CHAIN_NAME=GAIA TRANSFER_CHANNEL_NUMBER=0 $BATS $INTEGRATION_TEST_FILE
elif [[ "$PART" == "2" ]]; then
    CHAIN_NAME=EVMOS TRANSFER_CHANNEL_NUMBER=1 $BATS $INTEGRATION_TEST_FILE
    NEW_BINARY=true CHAIN_NAME=HOST TRANSFER_CHANNEL_NUMBER=2 $BATS $INTEGRATION_TEST_FILE
fi

