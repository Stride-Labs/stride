#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh
echo "Getting relevant addresses..."

STRIDE_ADDRESS=$($STRIDE_MAIN_CMD keys show ${STRIDE_VAL_PREFIX}1 --keyring-backend test -a)
GAIA_ADDRESS=$($GAIA_MAIN_CMD keys show ${GAIA_VAL_PREFIX}1 --keyring-backend test -a)
JUNO_ADDRESS=$($JUNO_MAIN_CMD keys show ${JUNO_VAL_PREFIX}1 --keyring-backend test -a)
OSMO_ADDRESS=$($OSMO_MAIN_CMD keys show ${OSMO_VAL_PREFIX}1 --keyring-backend test -a)

# Relayers
# NOTE: using $STRIDE_MAIN_CMD and $GAIA_MAIN_CMD here ONLY works because they rly1 and rly2
# keys are on stride1 and gaia1, respectively
HERMES_STRIDE_ADDRESS=$($STRIDE_MAIN_CMD keys show $HERMES_STRIDE_ACCT --keyring-backend test -a)
HERMES_GAIA_ADDRESS=$($GAIA_MAIN_CMD keys show $HERMES_GAIA_ACCT --keyring-backend test -a)
HERMES_JUNO_ADDRESS=$($JUNO_MAIN_CMD keys show $HERMES_JUNO_ACCT --keyring-backend test -a)
HERMES_OSMO_ADDRESS=$($OSMO_MAIN_CMD keys show $HERMES_OSMO_ACCT --keyring-backend test -a)

ICQ_STRIDE_ADDRESS=$($STRIDE_MAIN_CMD keys show $ICQ_STRIDE_ACCT --keyring-backend test -a)
ICQ_GAIA_ADDRESS=$($GAIA_MAIN_CMD keys show $ICQ_GAIA_ACCT --keyring-backend test -a)
ICQ_JUNO_ADDRESS=$($JUNO_MAIN_CMD keys show $ICQ_JUNO_ACCT --keyring-backend test -a)
ICQ_OSMO_ADDRESS=$($OSMO_MAIN_CMD keys show $ICQ_OSMO_ACCT --keyring-backend test -a)

echo "Grabbed all data, running tests..."