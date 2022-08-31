
#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

STRIDE_ADDRESS=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a)
GAIA_ADDRESS=$($GAIA_CMD keys show $GAIA_VAL_ACCT --keyring-backend test -a)
JUNO_ADDRESS=$($JUNO_CMD keys show $JUNO_VAL_ACCT --keyring-backend test -a)
OSMO_ADDRESS=$($OSMO_CMD keys show $OSMO_VAL_ACCT --keyring-backend test -a)

HERMES_STRIDE_ADDRESS=$($STRIDE_CMD keys show $HERMES_STRIDE_ACCT --keyring-backend test -a)
HERMES_GAIA_ADDRESS=$($GAIA_CMD keys show $HERMES_GAIA_ACCT --keyring-backend test -a)
HERMES_JUNO_ADDRESS=$($JUNO_CMD keys show $HERMES_JUNO_ACCT --keyring-backend test -a)
HERMES_OSMO_ADDRESS=$($OSMO_CMD keys show $HERMES_OSMO_ACCT --keyring-backend test -a)