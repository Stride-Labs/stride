#!/bin/bash
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
source ${SCRIPT_DIR}/../config.sh

# run test files
BATS=${SCRIPT_DIR}/bats/bats-core/bin/bats
INTEGRATION_TEST_FILE=${SCRIPT_DIR}/integration_tests.bats 

for i in ${!HOST_CHAINS[@]}; do
    CHAIN_NAME=${HOST_CHAINS[i]} TRANSFER_CHANNEL_NUMBER=$i $BATS $INTEGRATION_TEST_FILE
done