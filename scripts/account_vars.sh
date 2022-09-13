#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
echo "Getting relevant addresses..."

STRIDE_ADDRESS=$($STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a)
GAIA_ADDRESS=$($GAIA_MAIN_CMD keys show ${GAIA_VAL_PREFIX}1 --keyring-backend test -a)
JUNO_ADDRESS=$($JUNO_MAIN_CMD keys show ${JUNO_VAL_PREFIX}1 --keyring-backend test -a)
OSMO_ADDRESS=$($OSMO_MAIN_CMD keys show ${OSMO_VAL_PREFIX}1 --keyring-backend test -a)

echo "Grabbed all data, running tests..."