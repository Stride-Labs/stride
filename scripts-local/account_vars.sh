#!/bin/bash

set -eu 
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

source ${SCRIPT_DIR}/vars.sh

STRIDE_ADDRESS=$($STRIDE_CMD keys show $STRIDE_VAL_ACCT --keyring-backend test -a)
GAIA_ADDRESS=$($GAIA_CMD keys show $GAIA_VAL_ACCT --keyring-backend test -a)

HERMES_STRIDE_ADDRESS=$($STRIDE_CMD keys show $HERMES_STRIDE_ACCT --keyring-backend test -a)
HERMES_GAIA_ADDRESS=$($GAIA_CMD keys show $HERMES_GAIA_ACCT --keyring-backend test -a)

GAIA_DELEGATE_VAL='cosmosvaloper1pcag0cj4ttxg8l7pcg0q4ksuglswuuedadj7ne'
GAIA_DELEGATE_VAL_2='cosmosvaloper133lfs9gcpxqj6er3kx605e3v9lqp2pg5syhvsz'