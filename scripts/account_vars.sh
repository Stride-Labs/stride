#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
echo "Getting relevant addresses..."

# Stride
STRIDE_VAL_ADDRESS=$($STRIDE_CMD --home $SCRIPT_DIR/state/stride1 keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a)

# Gaia
GAIA_VAL_ADDRESS=$($GAIA_EXEC keys show ${GAIA_VAL_PREFIX}1 --keyring-backend test -a)

# Relayers
# NOTE: using $STRIDE_MAIN_CMD and $GAIA_MAIN_CMD here ONLY works because they rly1 and rly2
# keys are on stride1 and gaia1, respectively
RLY_ADDRESS_1=$($STRIDE_MAIN_CMD keys show rly1 --keyring-backend test -a)
RLY_ADDRESS_2=$($GAIA_MAIN_CMD keys show rly2 --keyring-backend test -a)

echo "Grabbed all data, running tests..."